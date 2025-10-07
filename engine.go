package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// StrategyInstance represents a running strategy instance
type StrategyInstance struct {
	ID              string
	Strategy        Strategy
	Config          StrategyConfig
	Symbol          string
	Interval        string
	IsRunning       bool
	CancelFunc      context.CancelFunc
	CurrentPosition *Position
	Account         *Account
	LastCandleTime  int64
	mu              sync.RWMutex
}

// StrategyEngine manages all running strategy instances
type StrategyEngine struct {
	instances map[string]*StrategyInstance
	source    *Source
	account   *Account
	mu        sync.RWMutex
}

// NewStrategyEngine creates a new strategy engine
func NewStrategyEngine(source *Source, account *Account) *StrategyEngine {
	return &StrategyEngine{
		instances: make(map[string]*StrategyInstance),
		source:    source,
		account:   account,
	}
}

// StartStrategy starts a new strategy instance
func (e *StrategyEngine) StartStrategy(id string, strategy Strategy, config StrategyConfig, symbol string, interval string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if strategy is already running
	if _, exists := e.instances[id]; exists {
		return fmt.Errorf("strategy with id %s is already running", id)
	}

	// Create context for this strategy
	ctx, cancel := context.WithCancel(context.Background())

	instance := &StrategyInstance{
		ID:         id,
		Strategy:   strategy,
		Config:     config,
		Symbol:     symbol,
		Interval:   interval,
		IsRunning:  true,
		CancelFunc: cancel,
		Account:    e.account,
	}

	e.instances[id] = instance

	// Start strategy goroutine
	go e.runStrategy(ctx, instance)

	return nil
}

// StopStrategy stops a running strategy instance
func (e *StrategyEngine) StopStrategy(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	instance, exists := e.instances[id]
	if !exists {
		return fmt.Errorf("strategy with id %s not found", id)
	}

	// Cancel the context to stop the goroutine
	instance.CancelFunc()
	instance.IsRunning = false

	// Close any open position
	if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
		if err := e.closeStrategyPosition(instance, "Strategy Stopped"); err != nil {
			fmt.Printf("Error closing position for strategy %s: %v\n", id, err)
		}
	}

	delete(e.instances, id)

	return nil
}

// GetRunningStrategies returns all running strategy instances
func (e *StrategyEngine) GetRunningStrategies() []StrategyInstance {
	e.mu.RLock()
	defer e.mu.RUnlock()

	instances := make([]StrategyInstance, 0, len(e.instances))
	for _, instance := range e.instances {
		instances = append(instances, *instance)
	}

	return instances
}

// runStrategy is the main loop for a strategy instance
func (e *StrategyEngine) runStrategy(ctx context.Context, instance *StrategyInstance) {
	fmt.Printf("Starting strategy %s for %s on %s\n", instance.ID, instance.Symbol, instance.Interval)

	// Calculate interval duration
	intervalDuration := e.getIntervalDuration(instance.Interval)

	// Fetch initial historical data for strategy initialization
	candles, err := e.source.FetchHistoricalCandles(instance.Symbol, instance.Interval, 2)
	if err != nil {
		fmt.Printf("Error fetching initial candles for %s: %v\n", instance.ID, err)
		return
	}

	if len(candles) > 0 {
		instance.LastCandleTime = candles[len(candles)-1].Timestamp
	}

	// Initialize strategy with historical data
	if err := instance.Strategy.Initialize(candles, instance.Config); err != nil {
		fmt.Printf("Error initializing strategy %s: %v\n", instance.ID, err)
		return
	}

	// Create ticker for checking new candles
	ticker := time.NewTicker(intervalDuration / 5) // Check 5 times per interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Strategy %s stopped\n", instance.ID)
			return

		case <-ticker.C:
			// Check for new candle
			if err := e.checkAndProcessNewCandle(instance); err != nil {
				fmt.Printf("Error processing candle for %s: %v\n", instance.ID, err)
			}
		}
	}
}

// checkAndProcessNewCandle checks if a new candle is available and processes it
func (e *StrategyEngine) checkAndProcessNewCandle(instance *StrategyInstance) error {
	// Fetch latest candles
	candles, err := e.source.FetchHistoricalCandles(instance.Symbol, instance.Interval, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch candles: %w", err)
	}

	if len(candles) == 0 {
		return fmt.Errorf("no candles received")
	}

	latestCandle := candles[len(candles)-1]

	// Check if this is a new candle
	if latestCandle.Timestamp <= instance.LastCandleTime {
		return nil // No new candle yet
	}

	fmt.Printf("New candle for %s at %s: O=%s H=%s L=%s C=%s\n",
		instance.Symbol,
		time.Unix(latestCandle.Timestamp/1000, 0).Format("2006-01-02 15:04:05"),
		latestCandle.Open,
		latestCandle.High,
		latestCandle.Low,
		latestCandle.Close,
	)

	instance.LastCandleTime = latestCandle.Timestamp

	// Re-initialize strategy with updated data
	if err := instance.Strategy.Initialize(candles, instance.Config); err != nil {
		return fmt.Errorf("failed to reinitialize strategy: %w", err)
	}

	// Generate signals
	signals, err := instance.Strategy.GenerateSignals()
	if err != nil {
		return fmt.Errorf("failed to generate signals: %w", err)
	}

	// Process the latest signal
	if len(signals) > 0 {
		latestSignal := signals[len(signals)-1]

		// Only process signals from the latest candle
		if latestSignal.Index == len(candles)-1 {
			if err := e.processSignal(instance, latestSignal, latestCandle); err != nil {
				return fmt.Errorf("failed to process signal: %w", err)
			}
		}
	}

	// Check TP/SL for existing position
	if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
		if err := e.checkTPSL(instance, latestCandle); err != nil {
			return fmt.Errorf("failed to check TP/SL: %w", err)
		}
	}

	return nil
}

// processSignal handles a trading signal
func (e *StrategyEngine) processSignal(instance *StrategyInstance, signal Signal, candle hyperliquid.Candle) error {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	currentPrice := parseFloat(candle.Close)

	// Handle entry signals
	if signal.Type == SignalLong || signal.Type == SignalShort {
		side := "long"
		if signal.Type == SignalShort {
			side = "short"
		}

		// Check trade direction filter
		if instance.Config.TradeDirection == "long" && side == "short" {
			fmt.Printf("Skipping short signal for %s (long-only mode)\n", instance.ID)
			return nil
		}
		if instance.Config.TradeDirection == "short" && side == "long" {
			fmt.Printf("Skipping long signal for %s (short-only mode)\n", instance.ID)
			return nil
		}

		// Close existing position if any
		if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
			if err := e.closeStrategyPosition(instance, "New Signal"); err != nil {
				return fmt.Errorf("failed to close existing position: %w", err)
			}
		}

		// Open new position
		fmt.Printf("Opening %s position for %s at price %.2f (signal: %s)\n",
			side, instance.Symbol, currentPrice, signal.Reason)

		if err := e.openStrategyPosition(instance, side, currentPrice); err != nil {
			return fmt.Errorf("failed to open position: %w", err)
		}
	}

	// Handle exit signals
	if signal.Type == SignalCloseLong || signal.Type == SignalCloseShort {
		if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
			fmt.Printf("Closing position for %s at price %.2f (signal: %s)\n",
				instance.Symbol, currentPrice, signal.Reason)

			if err := e.closeStrategyPosition(instance, "Strategy Signal"); err != nil {
				return fmt.Errorf("failed to close position: %w", err)
			}
		}
	}

	return nil
}

// openStrategyPosition opens a new position via the account
func (e *StrategyEngine) openStrategyPosition(instance *StrategyInstance, side string, price float64) error {
	if !instance.Account.IsConnected() {
		return fmt.Errorf("account not connected")
	}

	var resp *OrderResponse
	var err error

	if side == "long" {
		resp, err = instance.Account.OpenLongPosition(
			instance.Symbol,
			instance.Config.PositionSize,
			price,
			"market",
		)
	} else {
		resp, err = instance.Account.OpenShortPosition(
			instance.Symbol,
			instance.Config.PositionSize,
			price,
			"market",
		)
	}

	if err != nil {
		return err
	}

	// Track position
	instance.CurrentPosition = &Position{
		EntryPrice: price,
		EntryTime:  time.Now().UnixMilli(),
		Side:       side,
		Size:       instance.Config.PositionSize,
		IsOpen:     true,
	}

	fmt.Printf("Position opened successfully: %+v\n", resp)
	return nil
}

// closeStrategyPosition closes the current position
func (e *StrategyEngine) closeStrategyPosition(instance *StrategyInstance, reason string) error {
	if instance.CurrentPosition == nil || !instance.CurrentPosition.IsOpen {
		return nil
	}

	if !instance.Account.IsConnected() {
		return fmt.Errorf("account not connected")
	}

	// Close position via account
	resp, err := instance.Account.ClosePosition(ClosePositionRequest{
		Coin: instance.Symbol,
		Size: instance.CurrentPosition.Size,
	})

	if err != nil {
		return err
	}

	// Update position
	instance.CurrentPosition.IsOpen = false
	instance.CurrentPosition.ExitReason = reason
	instance.CurrentPosition.ExitTime = time.Now().UnixMilli()

	fmt.Printf("Position closed successfully: %s - %+v\n", reason, resp)
	return nil
}

// checkTPSL checks if TP or SL is hit
func (e *StrategyEngine) checkTPSL(instance *StrategyInstance, candle hyperliquid.Candle) error {
	if instance.CurrentPosition == nil || !instance.CurrentPosition.IsOpen {
		return nil
	}

	high := parseFloat(candle.High)
	low := parseFloat(candle.Low)

	var tpPrice, slPrice float64

	if instance.CurrentPosition.Side == "long" {
		tpPrice = instance.CurrentPosition.EntryPrice * (1 + instance.Config.TakeProfitPercent/100)
		slPrice = instance.CurrentPosition.EntryPrice * (1 - instance.Config.StopLossPercent/100)

		// Check SL first
		if low <= slPrice && instance.Config.StopLossPercent > 0 {
			fmt.Printf("Stop Loss hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, slPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPosition(instance, "SL Hit")
		}

		// Check TP
		if high >= tpPrice && instance.Config.TakeProfitPercent > 0 {
			fmt.Printf("Take Profit hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, tpPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPosition(instance, "TP Hit")
		}
	} else { // short position
		tpPrice = instance.CurrentPosition.EntryPrice * (1 - instance.Config.TakeProfitPercent/100)
		slPrice = instance.CurrentPosition.EntryPrice * (1 + instance.Config.StopLossPercent/100)

		// Check SL first
		if high >= slPrice && instance.Config.StopLossPercent > 0 {
			fmt.Printf("Stop Loss hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, slPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPosition(instance, "SL Hit")
		}

		// Check TP
		if low <= tpPrice && instance.Config.TakeProfitPercent > 0 {
			fmt.Printf("Take Profit hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, tpPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPosition(instance, "TP Hit")
		}
	}

	return nil
}

// getIntervalDuration converts interval string to duration
func (e *StrategyEngine) getIntervalDuration(interval string) time.Duration {
	switch interval {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}

// StopAllStrategies stops all running strategies
func (e *StrategyEngine) StopAllStrategies() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for id := range e.instances {
		if err := e.StopStrategy(id); err != nil {
			fmt.Printf("Error stopping strategy %s: %v\n", id, err)
		}
	}
}

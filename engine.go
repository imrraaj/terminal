package main

import (
	"context"
	"fmt"
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

type StrategyInstance struct {
	ID              string         `json:"id"`
	Strategy        Strategy       `json:"-"`
	Config          StrategyConfig `json:"config"`
	Symbol          string         `json:"symbol"`
	Interval        string         `json:"interval"`
	IsRunning       bool           `json:"isRunning"`
	CurrentPosition *Position      `json:"currentPosition"`
	Account         *Account       `json:"-"`
	LastCandleTime  int64          `json:"lastCandleTime"`
}

type StrategyEngine struct {
	instances map[string]*StrategyInstance
	source    *Source
	account   *Account
}

func NewStrategyEngine(source *Source, account *Account) *StrategyEngine {
	return &StrategyEngine{
		instances: make(map[string]*StrategyInstance),
		source:    source,
		account:   account,
	}
}
func (e *StrategyEngine) StartStrategy(id string, strategy Strategy, config StrategyConfig, symbol string, interval string) error {
	if _, exists := e.instances[id]; exists {
		return fmt.Errorf("strategy with id %s is already running", id)
	}
	instance := &StrategyInstance{
		ID:        id,
		Strategy:  strategy,
		Config:    config,
		Symbol:    symbol,
		Interval:  interval,
		IsRunning: true,
		Account:   e.account,
	}
	e.instances[id] = instance
	return nil
}
func (e *StrategyEngine) StopStrategy(id string) error {
	instance, exists := e.instances[id]
	if !exists {
		return fmt.Errorf("strategy with id %s not found", id)
	}
	instance.IsRunning = false
	if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
		if err := e.closeStrategyPerpPosition(instance, "Strategy Stopped"); err != nil {
			fmt.Printf("Error closing position for strategy %s: %v\n", id, err)
		}
	}
	delete(e.instances, id)
	return nil
}
func (e *StrategyEngine) GetRunningStrategies() []StrategyInstance {
	instances := make([]StrategyInstance, 0, len(e.instances))
	for _, instance := range e.instances {
		instances = append(instances, *instance)
	}
	return instances
}
func (e *StrategyEngine) runStrategy(ctx context.Context, instance *StrategyInstance) {
	fmt.Printf("Starting strategy %s for %s on %s\n", instance.ID, instance.Symbol, instance.Interval)
	intervalDuration := e.getIntervalDuration(instance.Interval)
	candles, err := e.source.FetchHistoricalCandles(instance.Symbol, instance.Interval, 2)
	if err != nil {
		fmt.Printf("Error fetching initial candles for %s: %v\n", instance.ID, err)
		return
	}
	if len(candles) > 0 {
		instance.LastCandleTime = candles[len(candles)-1].Timestamp
	}
	if err := instance.Strategy.Initialize(candles, instance.Config); err != nil {
		fmt.Printf("Error initializing strategy %s: %v\n", instance.ID, err)
		return
	}
	ticker := time.NewTicker(intervalDuration / 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Strategy %s stopped\n", instance.ID)
			return
		case <-ticker.C:
			if err := e.checkAndProcessNewCandle(instance); err != nil {
				fmt.Printf("Error processing candle for %s: %v\n", instance.ID, err)
			}
		}
	}
}
func (e *StrategyEngine) checkAndProcessNewCandle(instance *StrategyInstance) error {
	candles, err := e.source.FetchHistoricalCandles(instance.Symbol, instance.Interval, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch candles: %w", err)
	}
	if len(candles) == 0 {
		return fmt.Errorf("no candles received")
	}
	latestCandle := candles[len(candles)-1]
	if latestCandle.Timestamp <= instance.LastCandleTime {
		return nil
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
	if err := instance.Strategy.Initialize(candles, instance.Config); err != nil {
		return fmt.Errorf("failed to reinitialize strategy: %w", err)
	}
	signals, err := instance.Strategy.GenerateSignals()
	if err != nil {
		return fmt.Errorf("failed to generate signals: %w", err)
	}
	if len(signals) > 0 {
		latestSignal := signals[len(signals)-1]
		if latestSignal.Index == len(candles)-1 {
			if err := e.processSignal(instance, latestSignal, latestCandle); err != nil {
				return fmt.Errorf("failed to process signal: %w", err)
			}
		}
	}
	if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
		if err := e.checkTPSL(instance, latestCandle); err != nil {
			return fmt.Errorf("failed to check TP/SL: %w", err)
		}
	}
	return nil
}
func (e *StrategyEngine) processSignal(instance *StrategyInstance, signal Signal, candle hyperliquid.Candle) error {
	currentPrice := parseFloat(candle.Close)
	if signal.Type == SignalLong || signal.Type == SignalShort {
		side := "long"
		if signal.Type == SignalShort {
			side = "short"
		}
		if instance.Config.TradeDirection == "long" && side == "short" {
			fmt.Printf("Skipping short signal for %s (long-only mode)\n", instance.ID)
			return nil
		}
		if instance.Config.TradeDirection == "short" && side == "long" {
			fmt.Printf("Skipping long signal for %s (short-only mode)\n", instance.ID)
			return nil
		}
		if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
			if err := e.closeStrategyPerpPosition(instance, "New Signal"); err != nil {
				return fmt.Errorf("failed to close existing position: %w", err)
			}
		}
		fmt.Printf("Opening %s position for %s at price %.2f (signal: %s)\n",
			side, instance.Symbol, currentPrice, signal.Reason)
		// TODO: factor out default 10x leverage
		if err := e.openStrategyPerpPosition(instance, side, currentPrice, 10); err != nil {
			return fmt.Errorf("failed to open position: %w", err)
		}
	}
	if signal.Type == SignalCloseLong || signal.Type == SignalCloseShort {
		if instance.CurrentPosition != nil && instance.CurrentPosition.IsOpen {
			fmt.Printf("Closing position for %s at price %.2f (signal: %s)\n",
				instance.Symbol, currentPrice, signal.Reason)
			if err := e.closeStrategyPerpPosition(instance, "Strategy Signal"); err != nil {
				return fmt.Errorf("failed to close position: %w", err)
			}
		}
	}
	return nil
}
func (e *StrategyEngine) openStrategyPerpPosition(instance *StrategyInstance, side string, price float64, leverage int) error {
	var resp *OrderResponse
	var err error
	if side == "long" {
		resp, err = instance.Account.OpenPerpOrder(
			instance.Symbol,
			true,
			instance.Config.PositionSize,
			leverage,
		)
	} else {
		resp, err = instance.Account.OpenPerpOrder(
			instance.Symbol,
			false,
			instance.Config.PositionSize,
			leverage,
		)
	}
	if err != nil {
		return err
	}
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
func (e *StrategyEngine) closeStrategyPerpPosition(instance *StrategyInstance, reason string) error {
	if instance.CurrentPosition == nil || !instance.CurrentPosition.IsOpen {
		return nil
	}
	resp, err := instance.Account.ClosePosition(instance.Symbol, instance.CurrentPosition.Size)
	if err != nil {
		return err
	}
	instance.CurrentPosition.IsOpen = false
	instance.CurrentPosition.ExitReason = reason
	instance.CurrentPosition.ExitTime = time.Now().UnixMilli()
	fmt.Printf("Position closed successfully: %s - %+v\n", reason, resp)
	return nil
}

func (e *StrategyEngine) openStrategySpotPosition(instance *StrategyInstance) error {
	candles, err := e.source.FetchHistoricalCandles(instance.Symbol, "1m", 1)
	if err != nil {
		return fmt.Errorf("failed to fetch current price: %w", err)
	}
	price := 0.0
	if len(candles) > 0 {
		price = parseFloat(candles[len(candles)-1].Close)
	}

	resp, err := instance.Account.SpotOrder(
		instance.Symbol,
		true,
		instance.Config.PositionSize,
	)
	if err != nil {
		return err
	}
	side := "buy"
	instance.CurrentPosition = &Position{
		EntryPrice: price,
		EntryTime:  time.Now().UnixMilli(),
		Side:       side,
		Size:       instance.Config.PositionSize,
		IsOpen:     true,
	}
	fmt.Printf("Spot %s order executed successfully at price %.2f: %+v\n", side, price, resp)
	return nil
}

func (e *StrategyEngine) closeStrategySpotPosition(instance *StrategyInstance, reason string) error {
	if instance.CurrentPosition == nil || !instance.CurrentPosition.IsOpen {
		return nil
	}

	resp, err := instance.Account.SpotOrder(
		instance.Symbol,
		false,
		instance.CurrentPosition.Size,
	)
	if err != nil {
		return err
	}
	instance.CurrentPosition.IsOpen = false
	instance.CurrentPosition.ExitReason = reason
	instance.CurrentPosition.ExitTime = time.Now().UnixMilli()
	fmt.Printf("Spot position closed successfully: %s - %+v\n", reason, resp)
	return nil
}
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
		if low <= slPrice && instance.Config.StopLossPercent > 0 {
			fmt.Printf("Stop Loss hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, slPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPerpPosition(instance, "SL Hit")
		}
		if high >= tpPrice && instance.Config.TakeProfitPercent > 0 {
			fmt.Printf("Take Profit hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, tpPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPerpPosition(instance, "TP Hit")
		}
	} else {
		tpPrice = instance.CurrentPosition.EntryPrice * (1 - instance.Config.TakeProfitPercent/100)
		slPrice = instance.CurrentPosition.EntryPrice * (1 + instance.Config.StopLossPercent/100)
		if high >= slPrice && instance.Config.StopLossPercent > 0 {
			fmt.Printf("Stop Loss hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, slPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPerpPosition(instance, "SL Hit")
		}
		if low <= tpPrice && instance.Config.TakeProfitPercent > 0 {
			fmt.Printf("Take Profit hit for %s at %.2f (entry: %.2f)\n",
				instance.Symbol, tpPrice, instance.CurrentPosition.EntryPrice)
			return e.closeStrategyPerpPosition(instance, "TP Hit")
		}
	}
	return nil
}
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
func (e *StrategyEngine) StopAllStrategies() {
	for id := range e.instances {
		if err := e.StopStrategy(id); err != nil {
			fmt.Printf("Error stopping strategy %s: %v\n", id, err)
		}
	}
}

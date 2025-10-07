package main

import (
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// SignalType represents the type of trading signal
type SignalType int

const (
	SignalNone SignalType = iota
	SignalLong
	SignalShort
	SignalCloseLong
	SignalCloseShort
)

// Signal represents a trading signal at a specific point
type Signal struct {
	Index      int        // Candle index where signal occurred
	Type       SignalType // Type of signal
	Price      float64    // Price at signal
	Time       int64      // Timestamp
	Confidence float64    // Optional: Signal confidence (0-1)
	Reason     string     // Optional: Reason for signal (e.g., "TP Hit", "SL Hit", "Strategy Exit")
}

// Position represents an open or closed position
type Position struct {
	EntryIndex    int     // Candle index of entry
	EntryPrice    float64 // Entry price
	EntryTime     int64   // Entry timestamp
	ExitIndex     int     // Candle index of exit (0 if still open)
	ExitPrice     float64 // Exit price (0 if still open)
	ExitTime      int64   // Exit timestamp (0 if still open)
	Side          string  // "long" or "short"
	Size          float64 // Position size
	PnL           float64 // Realized PnL (0 if still open)
	PnLPercentage float64 // PnL as percentage
	IsOpen        bool    // Whether position is still open
	ExitReason    string  // Reason for exit (e.g., "TP", "SL", "Strategy Signal")
	MaxDrawdown   float64 // Maximum drawdown during position
	MaxProfit     float64 // Maximum profit during position
}

// StrategyConfig represents the configuration for a strategy
type StrategyConfig struct {
	// Take Profit and Stop Loss (as percentage)
	TakeProfitPercent float64 // e.g., 10.0 for 10%
	StopLossPercent   float64 // e.g., 5.0 for 5%

	// Position sizing
	PositionSize  float64 // Position size (in base currency or percentage)
	UsePercentage bool    // If true, PositionSize is a percentage of capital

	// Risk management
	MaxPositions    int     // Maximum concurrent positions (0 = unlimited)
	MaxRiskPerTrade float64 // Maximum risk per trade as percentage of capital

	// Trade direction filter
	TradeDirection string // "both", "long", or "short"

	// Strategy-specific parameters (flexible map for any strategy)
	Parameters map[string]interface{}
}

// BacktestResult contains the results of backtesting a strategy
type BacktestResult struct {
	Positions []Position // All positions (open and closed)
	Signals   []Signal   // All generated signals

	// Performance metrics
	TotalPnL           float64
	TotalPnLPercent    float64
	WinRate            float64 // Percentage of winning trades
	TotalTrades        int
	WinningTrades      int
	LosingTrades       int
	AverageWin         float64
	AverageLoss        float64
	ProfitFactor       float64 // Gross profit / Gross loss
	MaxDrawdown        float64
	MaxDrawdownPercent float64

	// Additional metrics
	SharpeRatio       float64
	LongestWinStreak  int
	LongestLossStreak int
	AverageHoldTime   time.Duration
}

// StrategyOutput contains all visualization and backtest data
type StrategyOutputV2 struct {
	// Original trend data (for visualization)
	TrendLines  []float64
	TrendColors []string
	Directions  []int
	Labels      []Label

	// New: Trading signals and positions
	Signals        []Signal
	BacktestResult BacktestResult

	// Metadata
	StrategyName    string
	StrategyVersion string
}

// Strategy is the interface that all user strategies must implement
type Strategy interface {
	// GetName returns the strategy name
	GetName() string

	// GetVersion returns the strategy version
	GetVersion() string

	// GetDefaultConfig returns the default configuration for the strategy
	GetDefaultConfig() StrategyConfig

	// Initialize prepares the strategy with candle data and configuration
	Initialize(candles hyperliquid.Candles, config StrategyConfig) error

	// GenerateSignals analyzes the candles and generates trading signals
	// Returns entry/exit signals based on strategy logic
	GenerateSignals() ([]Signal, error)

	// GetVisualizationData returns data for chart visualization (optional)
	// For strategies that want to display indicators, trend lines, etc.
	GetVisualizationData() *StrategyOutput

	// Validate checks if the strategy configuration is valid
	Validate() error
}

// StrategyRunner executes a strategy and handles TP/SL logic
type StrategyRunner struct {
	strategy Strategy
	candles  hyperliquid.Candles
	config   StrategyConfig
}

// NewStrategyRunner creates a new strategy runner
func NewStrategyRunner(strategy Strategy, candles hyperliquid.Candles, config StrategyConfig) *StrategyRunner {
	return &StrategyRunner{
		strategy: strategy,
		candles:  candles,
		config:   config,
	}
}

// Run executes the strategy and returns complete output with backtest results
func (sr *StrategyRunner) Run() (*StrategyOutputV2, error) {
	// Initialize strategy
	if err := sr.strategy.Initialize(sr.candles, sr.config); err != nil {
		return nil, err
	}

	// Validate strategy
	if err := sr.strategy.Validate(); err != nil {
		return nil, err
	}

	// Generate signals from strategy
	signals, err := sr.strategy.GenerateSignals()
	if err != nil {
		return nil, err
	}

	// Process signals with TP/SL logic
	positions := sr.processSignalsWithTPSL(signals)

	// Calculate backtest metrics
	backtestResult := sr.calculateBacktestMetrics(positions)
	backtestResult.Signals = signals

	// Get visualization data (if strategy provides it)
	vizData := sr.strategy.GetVisualizationData()

	output := &StrategyOutputV2{
		Signals:         signals,
		BacktestResult:  backtestResult,
		StrategyName:    sr.strategy.GetName(),
		StrategyVersion: sr.strategy.GetVersion(),
	}

	// Include visualization data if available
	if vizData != nil {
		output.TrendLines = vizData.TrendLines
		output.TrendColors = vizData.TrendColors
		output.Directions = vizData.Directions
		output.Labels = vizData.Labels
	}

	return output, nil
}

// processSignalsWithTPSL processes trading signals and applies TP/SL logic
func (sr *StrategyRunner) processSignalsWithTPSL(signals []Signal) []Position {
	positions := []Position{}
	var currentPosition *Position
	var lastExitIndex int = -1 // Track last exit to prevent re-entry on same candle

	// Create a map of signals by candle index for fast lookup
	signalMap := make(map[int]Signal)
	for _, signal := range signals {
		signalMap[signal.Index] = signal
	}

	// Iterate through ALL candles, not just signals
	for candleIndex := 0; candleIndex < len(sr.candles); candleIndex++ {
		// Check if there's a signal at this candle
		signal, hasSignal := signalMap[candleIndex]

		if hasSignal {
			// Handle entry signals
			if signal.Type == SignalLong || signal.Type == SignalShort {
				side := "long"
				if signal.Type == SignalShort {
					side = "short"
				}

				// CRITICAL FIX: If we have an open position and signal is opposite direction, CLOSE IT
				if currentPosition != nil && currentPosition.IsOpen {
					// Check if signal is opposite to current position
					shouldClose := false
					if currentPosition.Side == "long" && signal.Type == SignalShort {
						shouldClose = true
					} else if currentPosition.Side == "short" && signal.Type == SignalLong {
						shouldClose = true
					}

					if shouldClose {
						sr.closePosition(currentPosition, signal.Index, signal.Price, "Trend Reversal")
						positions = append(positions, *currentPosition)
						currentPosition = nil
						lastExitIndex = candleIndex
					}
				}

				// Don't enter on same candle we just exited
				if candleIndex == lastExitIndex {
					continue
				}

				// Filter based on trade direction setting - only for NEW entries
				if sr.config.TradeDirection == "long" && side == "short" {
					continue // Skip short entry signals
				}
				if sr.config.TradeDirection == "short" && side == "long" {
					continue // Skip long entry signals
				}

				// Validate entry - check if stop loss would be hit on entry candle
				candle := sr.candles[signal.Index]
				high := parseFloat(candle.High)
				low := parseFloat(candle.Low)

				validEntry := true
				if side == "long" && sr.config.StopLossPercent > 0 {
					slPrice := signal.Price * (1 - sr.config.StopLossPercent/100)
					if low <= slPrice {
						validEntry = false // Entry candle already hit SL
					}
				} else if side == "short" && sr.config.StopLossPercent > 0 {
					slPrice := signal.Price * (1 + sr.config.StopLossPercent/100)
					if high >= slPrice {
						validEntry = false // Entry candle already hit SL
					}
				}

				// Only open position if entry is valid
				if validEntry {
					currentPosition = &Position{
						EntryIndex: signal.Index,
						EntryPrice: signal.Price,
						EntryTime:  signal.Time,
						Side:       side,
						Size:       sr.config.PositionSize,
						IsOpen:     true,
					}
				}
			}

			// Handle exit signals
			if signal.Type == SignalCloseLong || signal.Type == SignalCloseShort {
				if currentPosition != nil && currentPosition.IsOpen {
					sr.closePosition(currentPosition, signal.Index, signal.Price, "Strategy Signal")
					positions = append(positions, *currentPosition)
					currentPosition = nil
					lastExitIndex = candleIndex
				}
			}
		}

		// Check TP/SL on EVERY candle if we have an open position
		if currentPosition != nil && currentPosition.IsOpen {
			tpslExit := sr.checkTPSL(currentPosition, candleIndex)
			if tpslExit != nil {
				positions = append(positions, *tpslExit)
				currentPosition = nil
				lastExitIndex = candleIndex
			}
		}
	}

	// Close any remaining open position at the last candle
	if currentPosition != nil && currentPosition.IsOpen {
		lastCandle := sr.candles[len(sr.candles)-1]
		lastPrice := parseFloat(lastCandle.Close)
		sr.closePosition(currentPosition, len(sr.candles)-1, lastPrice, "End of Period")
		positions = append(positions, *currentPosition)
	}

	return positions
}

// checkTPSL checks if TP or SL is hit for the current position
func (sr *StrategyRunner) checkTPSL(position *Position, candleIndex int) *Position {
	if candleIndex >= len(sr.candles) {
		return nil
	}

	candle := sr.candles[candleIndex]
	high := parseFloat(candle.High)
	low := parseFloat(candle.Low)
	close := parseFloat(candle.Close)

	// Calculate TP and SL prices
	var tpPrice, slPrice float64
	if position.Side == "long" {
		tpPrice = position.EntryPrice * (1 + sr.config.TakeProfitPercent/100)
		slPrice = position.EntryPrice * (1 - sr.config.StopLossPercent/100)

		// Check SL first (more conservative)
		if low <= slPrice && sr.config.StopLossPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, slPrice, "SL Hit")
			return &closedPos
		}

		// Check TP
		if high >= tpPrice && sr.config.TakeProfitPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, tpPrice, "TP Hit")
			return &closedPos
		}
	} else { // short position
		tpPrice = position.EntryPrice * (1 - sr.config.TakeProfitPercent/100)
		slPrice = position.EntryPrice * (1 + sr.config.StopLossPercent/100)

		// Check SL first
		if high >= slPrice && sr.config.StopLossPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, slPrice, "SL Hit")
			return &closedPos
		}

		// Check TP
		if low <= tpPrice && sr.config.TakeProfitPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, tpPrice, "TP Hit")
			return &closedPos
		}
	}

	// Update max drawdown and profit
	currentPnL := sr.calculatePnL(position, close)
	if currentPnL > position.MaxProfit {
		position.MaxProfit = currentPnL
	}
	if currentPnL < position.MaxDrawdown {
		position.MaxDrawdown = currentPnL
	}

	return nil
}

// closePosition closes a position and calculates PnL
func (sr *StrategyRunner) closePosition(position *Position, exitIndex int, exitPrice float64, reason string) {
	position.ExitIndex = exitIndex
	position.ExitPrice = exitPrice
	position.ExitTime = sr.candles[exitIndex].Timestamp
	position.IsOpen = false
	position.ExitReason = reason

	position.PnL = sr.calculatePnL(position, exitPrice)
	position.PnLPercentage = (position.PnL / (position.EntryPrice * position.Size)) * 100
}

// calculatePnL calculates PnL for a position
func (sr *StrategyRunner) calculatePnL(position *Position, currentPrice float64) float64 {
	if position.Side == "long" {
		return (currentPrice - position.EntryPrice) * position.Size
	}
	// short
	return (position.EntryPrice - currentPrice) * position.Size
}

// calculateBacktestMetrics calculates all backtest performance metrics
func (sr *StrategyRunner) calculateBacktestMetrics(positions []Position) BacktestResult {
	result := BacktestResult{
		Positions: positions,
	}

	if len(positions) == 0 {
		return result
	}

	var totalWin, totalLoss float64
	var winStreak, lossStreak, currentWinStreak, currentLossStreak int
	var totalHoldTime time.Duration

	for _, pos := range positions {
		if pos.IsOpen {
			continue
		}

		result.TotalTrades++
		result.TotalPnL += pos.PnL

		if pos.PnL > 0 {
			result.WinningTrades++
			totalWin += pos.PnL
			currentWinStreak++
			currentLossStreak = 0
			if currentWinStreak > winStreak {
				winStreak = currentWinStreak
			}
		} else {
			result.LosingTrades++
			totalLoss += -pos.PnL
			currentLossStreak++
			currentWinStreak = 0
			if currentLossStreak > lossStreak {
				lossStreak = currentLossStreak
			}
		}

		// Track max drawdown
		if pos.MaxDrawdown < result.MaxDrawdown {
			result.MaxDrawdown = pos.MaxDrawdown
		}

		// Calculate hold time
		holdTime := time.Duration(pos.ExitTime-pos.EntryTime) * time.Millisecond
		totalHoldTime += holdTime
	}

	// Calculate metrics
	if result.TotalTrades > 0 {
		result.WinRate = (float64(result.WinningTrades) / float64(result.TotalTrades)) * 100
		result.AverageHoldTime = totalHoldTime / time.Duration(result.TotalTrades)
	}

	if result.WinningTrades > 0 {
		result.AverageWin = totalWin / float64(result.WinningTrades)
	}

	if result.LosingTrades > 0 {
		result.AverageLoss = totalLoss / float64(result.LosingTrades)
	}

	if totalLoss > 0 {
		result.ProfitFactor = totalWin / totalLoss
	}

	result.LongestWinStreak = winStreak
	result.LongestLossStreak = lossStreak

	return result
}

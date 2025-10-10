package main

import (
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

type SignalType int

const (
	SignalNone SignalType = iota
	SignalLong
	SignalShort
	SignalCloseLong
	SignalCloseShort
)

type Signal struct {
	Index      int
	Type       SignalType
	Price      float64
	Time       int64
	Confidence float64
	Reason     string
}
type Position struct {
	EntryIndex    int
	EntryPrice    float64
	EntryTime     int64
	ExitIndex     int
	ExitPrice     float64
	ExitTime      int64
	Side          string
	Size          float64
	PnL           float64
	PnLPercentage float64
	IsOpen        bool
	ExitReason    string
	MaxDrawdown   float64
	MaxProfit     float64
}
type StrategyConfig struct {
	TakeProfitPercent float64
	StopLossPercent   float64
	PositionSize      float64
	UsePercentage     bool
	MaxPositions      int
	MaxRiskPerTrade   float64
	TradeDirection    string
	Parameters        map[string]any
}
type BacktestResult struct {
	Positions          []Position
	Signals            []Signal
	TotalPnL           float64
	TotalPnLPercent    float64
	WinRate            float64
	TotalTrades        int
	WinningTrades      int
	LosingTrades       int
	AverageWin         float64
	AverageLoss        float64
	ProfitFactor       float64
	MaxDrawdown        float64
	MaxDrawdownPercent float64
	SharpeRatio        float64
	LongestWinStreak   int
	LongestLossStreak  int
	AverageHoldTime    time.Duration
}
type TrendPoint struct {
	Index     int
	Price     float64
	Direction int
	IsChange  bool
}
type TrendLine struct {
	StartIndex int
	StartPrice float64
	EndIndex   int
	EndPrice   float64
	Direction  int
}
type Label struct {
	Index      int
	Price      float64
	Text       string
	Direction  int
	Percentage float64
}
type StrategyOutput struct {
	TrendLines   []float64
	TrendColors  []string
	TrendChanges []TrendPoint
	Lines        []TrendLine
	FillColors   []string
	Directions   []int
	Labels       []Label
}
type StrategyOutputV2 struct {
	TrendLines      []float64
	TrendColors     []string
	Directions      []int
	Labels          []Label
	Signals         []Signal
	BacktestResult  BacktestResult
	StrategyName    string
	StrategyVersion string
}
type Strategy interface {
	GetName() string
	GetVersion() string
	GetDefaultConfig() StrategyConfig
	Initialize(candles hyperliquid.Candles, config StrategyConfig) error
	GenerateSignals() ([]Signal, error)
	GetVisualizationData() *StrategyOutput
	Validate() error
}
type StrategyRunner struct {
	strategy Strategy
	candles  hyperliquid.Candles
	config   StrategyConfig
}

func NewStrategyRunner(strategy Strategy, candles hyperliquid.Candles, config StrategyConfig) *StrategyRunner {
	return &StrategyRunner{
		strategy: strategy,
		candles:  candles,
		config:   config,
	}
}
func (sr *StrategyRunner) Run() (*StrategyOutputV2, error) {
	if err := sr.strategy.Initialize(sr.candles, sr.config); err != nil {
		return nil, err
	}
	if err := sr.strategy.Validate(); err != nil {
		return nil, err
	}
	signals, err := sr.strategy.GenerateSignals()
	if err != nil {
		return nil, err
	}
	positions := sr.processSignalsWithTPSL(signals)
	backtestResult := sr.calculateBacktestMetrics(positions)
	backtestResult.Signals = signals
	vizData := sr.strategy.GetVisualizationData()
	output := &StrategyOutputV2{
		Signals:         signals,
		BacktestResult:  backtestResult,
		StrategyName:    sr.strategy.GetName(),
		StrategyVersion: sr.strategy.GetVersion(),
	}
	if vizData != nil {
		output.TrendLines = vizData.TrendLines
		output.TrendColors = vizData.TrendColors
		output.Directions = vizData.Directions
		output.Labels = vizData.Labels
	}
	return output, nil
}
func (sr *StrategyRunner) processSignalsWithTPSL(signals []Signal) []Position {
	positions := []Position{}
	var currentPosition *Position
	var lastExitIndex int = -1
	signalMap := make(map[int]Signal)
	for _, signal := range signals {
		signalMap[signal.Index] = signal
	}
	for candleIndex := 0; candleIndex < len(sr.candles); candleIndex++ {
		signal, hasSignal := signalMap[candleIndex]
		if hasSignal {
			if signal.Type == SignalLong || signal.Type == SignalShort {
				side := "long"
				if signal.Type == SignalShort {
					side = "short"
				}
				if currentPosition != nil && currentPosition.IsOpen {
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
				if candleIndex == lastExitIndex {
					continue
				}
				if sr.config.TradeDirection == "long" && side == "short" {
					continue
				}
				if sr.config.TradeDirection == "short" && side == "long" {
					continue
				}
				candle := sr.candles[signal.Index]
				high := parseFloat(candle.High)
				low := parseFloat(candle.Low)
				validEntry := true
				if side == "long" && sr.config.StopLossPercent > 0 {
					slPrice := signal.Price * (1 - sr.config.StopLossPercent/100)
					if low <= slPrice {
						validEntry = false
					}
				} else if side == "short" && sr.config.StopLossPercent > 0 {
					slPrice := signal.Price * (1 + sr.config.StopLossPercent/100)
					if high >= slPrice {
						validEntry = false
					}
				}
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
			if signal.Type == SignalCloseLong || signal.Type == SignalCloseShort {
				if currentPosition != nil && currentPosition.IsOpen {
					sr.closePosition(currentPosition, signal.Index, signal.Price, "Strategy Signal")
					positions = append(positions, *currentPosition)
					currentPosition = nil
					lastExitIndex = candleIndex
				}
			}
		}
		if currentPosition != nil && currentPosition.IsOpen {
			tpslExit := sr.checkTPSL(currentPosition, candleIndex)
			if tpslExit != nil {
				positions = append(positions, *tpslExit)
				currentPosition = nil
				lastExitIndex = candleIndex
			}
		}
	}
	if currentPosition != nil && currentPosition.IsOpen {
		lastCandle := sr.candles[len(sr.candles)-1]
		lastPrice := parseFloat(lastCandle.Close)
		sr.closePosition(currentPosition, len(sr.candles)-1, lastPrice, "End of Period")
		positions = append(positions, *currentPosition)
	}
	return positions
}
func (sr *StrategyRunner) checkTPSL(position *Position, candleIndex int) *Position {
	if candleIndex >= len(sr.candles) {
		return nil
	}
	candle := sr.candles[candleIndex]
	high := parseFloat(candle.High)
	low := parseFloat(candle.Low)
	close := parseFloat(candle.Close)
	var tpPrice, slPrice float64
	if position.Side == "long" {
		tpPrice = position.EntryPrice * (1 + sr.config.TakeProfitPercent/100)
		slPrice = position.EntryPrice * (1 - sr.config.StopLossPercent/100)
		if low <= slPrice && sr.config.StopLossPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, slPrice, "SL Hit")
			return &closedPos
		}
		if high >= tpPrice && sr.config.TakeProfitPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, tpPrice, "TP Hit")
			return &closedPos
		}
	} else {
		tpPrice = position.EntryPrice * (1 - sr.config.TakeProfitPercent/100)
		slPrice = position.EntryPrice * (1 + sr.config.StopLossPercent/100)
		if high >= slPrice && sr.config.StopLossPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, slPrice, "SL Hit")
			return &closedPos
		}
		if low <= tpPrice && sr.config.TakeProfitPercent > 0 {
			closedPos := *position
			sr.closePosition(&closedPos, candleIndex, tpPrice, "TP Hit")
			return &closedPos
		}
	}
	currentPnL := sr.calculatePnL(position, close)
	if currentPnL > position.MaxProfit {
		position.MaxProfit = currentPnL
	}
	if currentPnL < position.MaxDrawdown {
		position.MaxDrawdown = currentPnL
	}
	return nil
}
func (sr *StrategyRunner) closePosition(position *Position, exitIndex int, exitPrice float64, reason string) {
	position.ExitIndex = exitIndex
	position.ExitPrice = exitPrice
	position.ExitTime = sr.candles[exitIndex].Timestamp
	position.IsOpen = false
	position.ExitReason = reason
	position.PnL = sr.calculatePnL(position, exitPrice)
	position.PnLPercentage = (position.PnL / (position.EntryPrice * position.Size)) * 100
}
func (sr *StrategyRunner) calculatePnL(position *Position, currentPrice float64) float64 {
	if position.Side == "long" {
		return (currentPrice - position.EntryPrice) * position.Size
	}
	return (position.EntryPrice - currentPrice) * position.Size
}
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
	var totalCapitalInvested float64
	for _, pos := range positions {
		if pos.IsOpen {
			continue
		}
		result.TotalTrades++
		result.TotalPnL += pos.PnL

		// Calculate capital invested (position size * entry price)
		capitalInvested := pos.Size * pos.EntryPrice
		totalCapitalInvested += capitalInvested

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
		if pos.MaxDrawdown < result.MaxDrawdown {
			result.MaxDrawdown = pos.MaxDrawdown
		}
		holdTime := time.Duration(pos.ExitTime-pos.EntryTime) * time.Millisecond
		totalHoldTime += holdTime
	}
	if result.TotalTrades > 0 {
		result.WinRate = (float64(result.WinningTrades) / float64(result.TotalTrades)) * 100
		result.AverageHoldTime = totalHoldTime / time.Duration(result.TotalTrades)

		// Calculate cumulative percentage gain based on average capital invested
		avgCapitalInvested := totalCapitalInvested / float64(result.TotalTrades)
		if avgCapitalInvested > 0 {
			result.TotalPnLPercent = (result.TotalPnL / avgCapitalInvested) * 100
		}
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

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
)

type Signal struct {
	Index  int
	Type   SignalType
	Price  float64
	Time   int64
	Reason string
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
	PositionSize        float64
	TradeDirection      string
	TakeProfitPercent   float64
	StopLossPercent     float64
	Interval            time.Duration
	Parameters          map[string]any
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
	TrendLines  []float64
	TrendColors []string
	Directions  []int
	Labels      []Label
	Lines       []TrendLine
	FillColors  []string
}

type BacktestOutput struct {
	TrendLines         []float64
	TrendColors        []string
	Directions         []int
	Labels             []Label
	Signals            []Signal
	Positions          []Position
	StrategyName       string
	StrategyVersion    string
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

type Strategy interface {
	GetName() string
	BuildConfig(params map[string]any) StrategyConfig

	// Used for charting: take candle data and give back the strategy output
	GenerateSignals(candles hyperliquid.Candles) ([]Signal, error)
	GetVisualizationData(candles hyperliquid.Candles) *StrategyOutput

	// Backtesting: take candle data and give back the backtest result
	Backtest(candles hyperliquid.Candles) (*BacktestOutput, error)

	// Live trading: take candle data and give back the strategy output
	Run(candles hyperliquid.Candles) (*StrategyOutput, error)
}

package main

// TrendPoint represents a trend change point
type TrendPoint struct {
	Index     int
	Price     float64
	Direction int // 1 for down, -1 for up
	IsChange  bool
}

// TrendLine represents a trend line
type TrendLine struct {
	StartIndex int
	StartPrice float64
	EndIndex   int
	EndPrice   float64
	Direction  int // 1 for down, -1 for up
}

// Label represents a label on the chart
type Label struct {
	Index      int
	Price      float64
	Text       string
	Direction  int // 1 for down, -1 for up
	Percentage float64
}

// StrategyOutput contains all the output data
type StrategyOutput struct {
	TrendLines   []float64    // Trend line values for each candle
	TrendColors  []string     // Color for each trend line point
	TrendChanges []TrendPoint // Points where trend changes
	Lines        []TrendLine  // Dashed lines
	Labels       []Label      // Labels with percentage
	FillColors   []string     // Fill colors between price and trend
	Directions   []int        // Direction for each candle
}

// Note: MaxTrendPointsStrategy implementation moved to max_trend_strategy.go
// This file kept for backward compatibility with old type definitions

package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// Strategy interface defines the contract for trading strategies
type Strategy interface {
	Execute() error
}

// MaxTrendPointsStrategy implements trend-following strategy
type MaxTrendPointsStrategy struct {
	ctx        context.Context
	exchange   *hyperliquid.Exchange
	info       *hyperliquid.Info
	coin       string
	factor     float64
	size       float64
	slippage   float64
	lastDir    int
	inPosition bool
}

// NewMaxTrendPointsStrategy creates a new strategy instance
func NewMaxTrendPointsStrategy(ctx context.Context, exchange *hyperliquid.Exchange, info *hyperliquid.Info, coin string, size float64) *MaxTrendPointsStrategy {
	return &MaxTrendPointsStrategy{
		ctx:        ctx,
		exchange:   exchange,
		info:       info,
		coin:       coin,
		factor:     8.0,
		size:       size,
		slippage:   0.05,
		lastDir:    0,
		inPosition: false,
	}
}

// Execute runs the strategy logic
func (s *MaxTrendPointsStrategy) Execute() error {
	// Fetch 5m candles (need at least 200 for HMA calculation)
	// Calculate time range: 250 candles * 5 minutes = 1250 minutes
	endTime := time.Now().UnixMilli()
	startTime := endTime - (250 * 5 * 60 * 1000) // 250 candles * 5 min * 60 sec * 1000 ms

	candles, err := s.info.CandlesSnapshot(s.ctx, s.coin, "5m", startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to fetch candles: %w", err)
	}

	if len(candles) < 200 {
		return fmt.Errorf("insufficient candles: need 200+, got %d", len(candles))
	}

	// Calculate signal
	direction := s.calculateTrend(candles)

	// Check for trend change
	if s.lastDir != 0 && direction != s.lastDir {
		fmt.Printf("Trend changed from %d to %d\n", s.lastDir, direction)

		// Close existing position if any
		if s.inPosition {
			fmt.Println("Closing existing position...")
			_, err := CloseLong(s.ctx, s.exchange, s.coin, s.size, s.slippage)
			if err != nil {
				return fmt.Errorf("failed to close position: %w", err)
			}
			s.inPosition = false
			fmt.Println("Position closed")
		}

		// If direction is uptrend (-1), open long
		if direction == -1 {
			fmt.Printf("Opening LONG on %s at size %.4f\n", s.coin, s.size)
			status, err := OpenLong(s.ctx, s.exchange, s.coin, s.size, s.slippage)
			if err != nil {
				return fmt.Errorf("failed to open long: %w", err)
			}
			s.inPosition = true
			fmt.Printf("Long opened: %s\n", status.String())
		}
	}

	s.lastDir = direction
	return nil
}

// calculateTrend computes the trend direction using the Max Trend Points algorithm
func (s *MaxTrendPointsStrategy) calculateTrend(candles []hyperliquid.Candle) int {
	n := len(candles)

	// Calculate HL2 and high-low differences
	hl2 := make([]float64, n)
	highLowDiff := make([]float64, n)

	for i := range candles {
		high := parseFloat(candles[i].High)
		low := parseFloat(candles[i].Low)
		hl2[i] = (high + low) / 2
		highLowDiff[i] = high - low
	}

	// Calculate HMA of high-low with period 200
	dist := hma(highLowDiff, 200)

	// Calculate bands
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)

	for i := range candles {
		upperBand[i] = hl2[i] + s.factor*dist[i]
		lowerBand[i] = hl2[i] - s.factor*dist[i]
	}

	// Calculate trend direction (only need last value)
	direction := 1
	trendLine := upperBand[0]

	for i := 1; i < n; i++ {
		close := parseFloat(candles[i-1].Close)

		// Update bands based on previous values
		if lowerBand[i] <= lowerBand[i-1] && close >= lowerBand[i-1] {
			lowerBand[i] = lowerBand[i-1]
		}
		if upperBand[i] >= upperBand[i-1] && close <= upperBand[i-1] {
			upperBand[i] = upperBand[i-1]
		}

		// Determine direction
		if dist[i-1] == 0 {
			direction = 1
		} else if trendLine == upperBand[i-1] {
			if parseFloat(candles[i].Close) > upperBand[i] {
				direction = -1
			} else {
				direction = 1
			}
		} else {
			if parseFloat(candles[i].Close) < lowerBand[i] {
				direction = 1
			} else {
				direction = -1
			}
		}

		// Set trend line based on direction
		if direction == -1 {
			trendLine = lowerBand[i]
		} else {
			trendLine = upperBand[i]
		}
	}

	return direction
}

// hma calculates Hull Moving Average
func hma(values []float64, period int) []float64 {
	if len(values) < period {
		return make([]float64, len(values))
	}

	halfPeriod := period / 2
	sqrtPeriod := int(math.Sqrt(float64(period)))

	wma1 := wma(values, halfPeriod)
	wma2 := wma(values, period)

	diff := make([]float64, len(values))
	for i := range diff {
		if i >= period-1 {
			diff[i] = 2*wma1[i] - wma2[i]
		}
	}

	return wma(diff, sqrtPeriod)
}

// wma calculates Weighted Moving Average
func wma(values []float64, period int) []float64 {
	result := make([]float64, len(values))
	if len(values) < period {
		return result
	}

	for i := period - 1; i < len(values); i++ {
		sum := 0.0
		weightSum := 0.0
		for j := 0; j < period; j++ {
			weight := float64(period - j)
			sum += values[i-j] * weight
			weightSum += weight
		}
		result[i] = sum / weightSum
	}

	return result
}

// parseFloat converts string to float64
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

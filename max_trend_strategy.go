package main

import (
	"fmt"
	"math"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// MaxTrendPointsStrategy implements the Strategy interface
type MaxTrendPointsStrategy struct {
	Factor  float64
	ColorUp string
	ColorDn string

	// Internal state
	candles hyperliquid.Candles
	config  StrategyConfig
	output  *StrategyOutput
}

// Ensure MaxTrendPointsStrategy implements Strategy interface
var _ Strategy = (*MaxTrendPointsStrategy)(nil)

func NewMaxTrendPointsStrategy(factor float64, colorUp, colorDn string) *MaxTrendPointsStrategy {
	if factor <= 0 {
		factor = 2.5
	}
	if colorUp == "" {
		colorUp = "rgb(28, 194, 216)"
	}
	if colorDn == "" {
		colorDn = "rgb(228, 144, 19)"
	}

	return &MaxTrendPointsStrategy{
		Factor:  factor,
		ColorUp: colorUp,
		ColorDn: colorDn,
	}
}

func (s *MaxTrendPointsStrategy) GetName() string {
	return "Max Trend Points"
}

func (s *MaxTrendPointsStrategy) GetVersion() string {
	return "1.0.0"
}

func (s *MaxTrendPointsStrategy) GetDefaultConfig() StrategyConfig {
	return StrategyConfig{
		TakeProfitPercent: 10.0,
		StopLossPercent:   5.0,
		PositionSize:      1.0,
		UsePercentage:     false,
		MaxPositions:      1,
		Parameters: map[string]interface{}{
			"factor":  2.5,
			"colorUp": "#1cc2d8",
			"colorDn": "#e49013",
		},
	}
}

func (s *MaxTrendPointsStrategy) Initialize(candles hyperliquid.Candles, config StrategyConfig) error {
	s.candles = candles
	s.config = config

	// Update parameters from config
	if factor, ok := config.Parameters["factor"].(float64); ok {
		s.Factor = factor
	}
	if colorUp, ok := config.Parameters["colorUp"].(string); ok {
		s.ColorUp = colorUp
	}
	if colorDn, ok := config.Parameters["colorDn"].(string); ok {
		s.ColorDn = colorDn
	}

	return nil
}

func (s *MaxTrendPointsStrategy) Validate() error {
	if len(s.candles) < 200 {
		return fmt.Errorf("insufficient data: need at least 200 candles, got %d", len(s.candles))
	}
	if s.Factor <= 0 {
		return fmt.Errorf("factor must be positive, got %f", s.Factor)
	}
	return nil
}

func (s *MaxTrendPointsStrategy) GenerateSignals() ([]Signal, error) {
	// First run the trend calculation to populate output
	if err := s.calculateTrends(); err != nil {
		return nil, err
	}

	signals := []Signal{}

	// Generate signals based on trend changes
	for i := 1; i < len(s.output.Directions); i++ {
		prevDirection := s.output.Directions[i-1]
		currDirection := s.output.Directions[i]

		// Trend change detected
		if prevDirection != currDirection {
			candle := s.candles[i]
			price := parseFloat(candle.Close)

			var signalType SignalType
			if currDirection == -1 {
				// Changed to uptrend - go long
				signalType = SignalLong
			} else {
				// Changed to downtrend - go short
				signalType = SignalShort
			}

			signals = append(signals, Signal{
				Index:      i,
				Type:       signalType,
				Price:      price,
				Time:       candle.Timestamp,
				Confidence: 0.8,
				Reason:     "Trend Reversal",
			})
		}
	}

	return signals, nil
}

func (s *MaxTrendPointsStrategy) GetVisualizationData() *StrategyOutput {
	return s.output
}

// calculateTrends runs the original trend calculation logic
func (s *MaxTrendPointsStrategy) calculateTrends() error {
	n := len(s.candles)

	s.output = &StrategyOutput{
		TrendLines:   make([]float64, n),
		TrendColors:  make([]string, n),
		TrendChanges: []TrendPoint{},
		Lines:        []TrendLine{},
		Labels:       []Label{},
		FillColors:   make([]string, n),
		Directions:   make([]int, n),
	}

	if n < 200 {
		return fmt.Errorf("insufficient candles")
	}

	// Calculate HL2 and high-low differences
	hl2 := make([]float64, n)
	highLowDiff := make([]float64, n)
	for i := range s.candles {
		high := parseFloat(s.candles[i].High)
		low := parseFloat(s.candles[i].Low)
		hl2[i] = (high + low) / 2
		highLowDiff[i] = high - low
	}

	// Calculate HMA of high-low with period 200
	dist := s.hma(highLowDiff, 200)

	// Calculate bands
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)

	for i := range s.candles {
		upperBand[i] = hl2[i] + s.Factor*dist[i]
		lowerBand[i] = hl2[i] - s.Factor*dist[i]
	}

	// Calculate trend lines and directions
	trendLine := make([]float64, n)
	direction := make([]int, n)

	for i := range s.candles {
		if i == 0 {
			direction[i] = 1
			trendLine[i] = upperBand[i]
		} else {
			// Update bands based on previous values
			close := parseFloat(s.candles[i-1].Close)
			if lowerBand[i] <= lowerBand[i-1] && close >= lowerBand[i-1] {
				lowerBand[i] = lowerBand[i-1]
			}
			if upperBand[i] >= upperBand[i-1] && close <= upperBand[i-1] {
				upperBand[i] = upperBand[i-1]
			}

			// Determine direction
			if dist[i-1] == 0 {
				direction[i] = 1
			} else if trendLine[i-1] == upperBand[i-1] {
				if parseFloat(s.candles[i].Close) > upperBand[i] {
					direction[i] = -1
				} else {
					direction[i] = 1
				}
			} else {
				if parseFloat(s.candles[i].Close) < lowerBand[i] {
					direction[i] = 1
				} else {
					direction[i] = -1
				}
			}

			// Set trend line based on direction
			if direction[i] == -1 {
				trendLine[i] = lowerBand[i]
			} else {
				trendLine[i] = upperBand[i]
			}
		}

		s.output.TrendLines[i] = trendLine[i]
		s.output.Directions[i] = direction[i]
	}

	// Track trend changes and create lines/labels
	var highest []float64
	var lowest []float64
	var start int
	var currentLineUp *TrendLine
	var currentLineDn *TrendLine

	for i := 1; i < n; i++ {
		tChange := direction[i] != direction[i-1]

		if tChange {
			highest = []float64{}
			lowest = []float64{}
			start = i

			s.output.TrendChanges = append(s.output.TrendChanges, TrendPoint{
				Index:     i,
				Price:     trendLine[i],
				Direction: direction[i],
				IsChange:  true,
			})

			if direction[i] == 1 {
				currentLineDn = &TrendLine{
					StartIndex: i,
					StartPrice: parseFloat(s.candles[i].Close),
					EndIndex:   i,
					EndPrice:   parseFloat(s.candles[i].Close),
					Direction:  1,
				}
				if currentLineUp != nil {
					s.output.Lines = append(s.output.Lines, *currentLineUp)
					currentLineUp = nil
				}
			} else {
				currentLineUp = &TrendLine{
					StartIndex: i,
					StartPrice: parseFloat(s.candles[i].Close),
					EndIndex:   i,
					EndPrice:   parseFloat(s.candles[i].Close),
					Direction:  -1,
				}
				if currentLineDn != nil {
					s.output.Lines = append(s.output.Lines, *currentLineDn)
					currentLineDn = nil
				}
			}
		} else {
			if direction[i] == -1 {
				highest = append(highest, parseFloat(s.candles[i].High))
				if currentLineUp != nil && len(highest) > 0 {
					maxIdx, maxVal := s.findMax(highest)
					currentLineUp.EndIndex = start + maxIdx + 1
					currentLineUp.EndPrice = maxVal
				}
			} else {
				lowest = append(lowest, parseFloat(s.candles[i].Low))
				if currentLineDn != nil && len(lowest) > 0 {
					minIdx, minVal := s.findMin(lowest)
					currentLineDn.EndIndex = start + minIdx + 1
					currentLineDn.EndPrice = minVal
				}
			}
		}

		// Set colors
		if direction[i] == -1 {
			s.output.TrendColors[i] = s.ColorUp
			s.output.FillColors[i] = s.ColorUp + "15"
		} else {
			s.output.TrendColors[i] = s.ColorDn
			s.output.FillColors[i] = s.ColorDn + "15"
		}
	}

	if currentLineUp != nil {
		s.output.Lines = append(s.output.Lines, *currentLineUp)
	}
	if currentLineDn != nil {
		s.output.Lines = append(s.output.Lines, *currentLineDn)
	}

	// Create labels with percentages
	for _, line := range s.output.Lines {
		var percentage float64
		var text string

		if line.Direction == -1 && line.EndPrice > line.StartPrice {
			percentage = ((line.EndPrice - line.StartPrice) / line.StartPrice) * 100
			text = s.formatPercent(percentage)
			s.output.Labels = append(s.output.Labels, Label{
				Index:      line.EndIndex,
				Price:      line.EndPrice,
				Text:       text,
				Direction:  -1,
				Percentage: percentage,
			})
		} else if line.Direction == 1 && line.EndPrice < line.StartPrice {
			percentage = ((line.EndPrice - line.StartPrice) / line.StartPrice) * 100
			text = s.formatPercent(percentage)
			s.output.Labels = append(s.output.Labels, Label{
				Index:      line.EndIndex,
				Price:      line.EndPrice,
				Text:       text,
				Direction:  1,
				Percentage: percentage,
			})
		}
	}

	return nil
}

// Calculate HMA (Hull Moving Average)
func (s *MaxTrendPointsStrategy) hma(values []float64, period int) []float64 {
	if len(values) < period {
		return make([]float64, len(values))
	}

	halfPeriod := period / 2
	sqrtPeriod := int(math.Sqrt(float64(period)))

	wma1 := s.wma(values, halfPeriod)
	wma2 := s.wma(values, period)

	diff := make([]float64, len(values))
	for i := range diff {
		if i >= period-1 {
			diff[i] = 2*wma1[i] - wma2[i]
		}
	}

	return s.wma(diff, sqrtPeriod)
}

// Calculate WMA (Weighted Moving Average)
func (s *MaxTrendPointsStrategy) wma(values []float64, period int) []float64 {
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

func (s *MaxTrendPointsStrategy) findMax(arr []float64) (int, float64) {
	if len(arr) == 0 {
		return 0, 0
	}
	maxIdx := 0
	maxVal := arr[0]
	for i, v := range arr {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx, maxVal
}

func (s *MaxTrendPointsStrategy) findMin(arr []float64) (int, float64) {
	if len(arr) == 0 {
		return 0, 0
	}
	minIdx := 0
	minVal := arr[0]
	for i, v := range arr {
		if v < minVal {
			minVal = v
			minIdx = i
		}
	}
	return minIdx, minVal
}

func (s *MaxTrendPointsStrategy) formatPercent(percentage float64) string {
	sign := ""
	if percentage > 0 {
		sign = "+"
	}
	return sign + fmt.Sprintf("%.2f%%", percentage)
}

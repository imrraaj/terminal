package main

import (
	"context"
	"fmt"
	"math"
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

type MaxTrendPointsStrategy struct {
	ID             string
	Symbol         string
	Interval       string
	LastCandleTime int64
	IsRunning      bool
	Position       *Position
	ctx            context.Context
	cancel         context.CancelFunc
	Factor         float64
	Config         StrategyConfig
	output         *StrategyOutput
	account        *Account
}

func NewMaxTrendPointsStrategy(params map[string]any) *MaxTrendPointsStrategy {
	strategy := &MaxTrendPointsStrategy{}
	strategy.Config = strategy.BuildConfig(params)
	if factor, ok := params["factor"].(float64); ok {
		strategy.Factor = factor
	}
	return strategy
}

func (s *MaxTrendPointsStrategy) GetName() string {
	return "Max Trend Points"
}

func (s *MaxTrendPointsStrategy) BuildConfig(params map[string]any) StrategyConfig {
	config := StrategyConfig{
		PositionSize:      0.01,
		TradeDirection:    "both",
		TakeProfitPercent: 5.0,
		StopLossPercent:   2.0,
		Parameters:        params,
	}

	if size, ok := params["positionSize"].(float64); ok {
		config.PositionSize = size
	}
	if direction, ok := params["tradeDirection"].(string); ok {
		config.TradeDirection = direction
	}
	if tp, ok := params["takeProfitPercent"].(float64); ok {
		config.TakeProfitPercent = tp
	}
	if sl, ok := params["stopLossPercent"].(float64); ok {
		config.StopLossPercent = sl
	}

	return config
}

func (s *MaxTrendPointsStrategy) GenerateSignals(candles hyperliquid.Candles) ([]Signal, error) {
	if err := s.calculateTrends(candles); err != nil {
		return nil, err
	}
	signals := []Signal{}
	for i := 1; i < len(s.output.Directions); i++ {
		prevDirection := s.output.Directions[i-1]
		currDirection := s.output.Directions[i]
		if prevDirection != currDirection {
			candle := candles[i]
			price := parseFloat(candle.Close)
			var signalType SignalType
			if prevDirection == 1 && currDirection == -1 {
				signalType = SignalLong
			} else if prevDirection == -1 && currDirection == 1 {
				signalType = SignalShort
			} else {
				continue
			}

			signals = append(signals, Signal{
				Index:  i,
				Type:   signalType,
				Price:  price,
				Time:   candle.Timestamp,
				Reason: "Trend Reversal",
			})
		}
	}
	return signals, nil
}

func (s *MaxTrendPointsStrategy) GetVisualizationData(candles hyperliquid.Candles) *StrategyOutput {
	if err := s.calculateTrends(candles); err != nil {
		return nil
	}
	return s.output
}

func (s *MaxTrendPointsStrategy) Run(candles hyperliquid.Candles) (*StrategyOutput, error) {
	if err := s.calculateTrends(candles); err != nil {
		return nil, err
	}
	return s.output, nil
}

func (s *MaxTrendPointsStrategy) Backtest(candles hyperliquid.Candles) (*BacktestOutput, error) {
	signals, err := s.GenerateSignals(candles)
	if err != nil {
		return nil, err
	}
	if err := s.calculateTrends(candles); err != nil {
		return nil, err
	}
	positions := s.calculateBacktestPositions(candles, signals)
	output := s.calculateBacktestOutput(positions)

	output.TrendLines = s.output.TrendLines
	output.TrendColors = s.output.TrendColors
	output.Directions = s.output.Directions
	output.Labels = s.output.Labels
	output.Signals = signals
	output.StrategyName = s.GetName()
	output.StrategyVersion = "1.0"

	return &output, nil
}

func (s *MaxTrendPointsStrategy) HandleSignal(signal Signal, candle hyperliquid.Candle) {
	price := parseFloat(candle.Close)

	fmt.Printf("[%s] 📊 Signal Received: Type=%d at %.2f - %s\n", s.ID, signal.Type, price, signal.Reason)

	if signal.Type != SignalLong && signal.Type != SignalShort {
		fmt.Printf("[%s] ⚠️  Invalid signal type: %d\n", s.ID, signal.Type)
		return
	}

	side := "long"
	isBuy := true
	if signal.Type == SignalShort {
		side = "short"
		isBuy = false
	}

	if s.Config.TradeDirection == "long" && side == "short" {
		fmt.Printf("[%s] ⚠️  Signal filtered: SHORT signal ignored (trade direction: long only)\n", s.ID)
		return
	}
	if s.Config.TradeDirection == "short" && side == "long" {
		fmt.Printf("[%s] ⚠️  Signal filtered: LONG signal ignored (trade direction: short only)\n", s.ID)
		return
	}

	if s.Position != nil && s.Position.IsOpen {
		if s.Position.Side == side {
			fmt.Printf("[%s] ℹ️  Already in %s position, ignoring signal\n", s.ID, side)
			return
		}
		fmt.Printf("[%s] 🔄 Closing existing %s position before opening new %s position\n", s.ID, s.Position.Side, side)
		s.ClosePosition("Trend Reversal")
	}

	fmt.Printf("[%s] 🚀 Opening %s position: size=%.4f, leverage=10x\n", s.ID, side, s.Config.PositionSize)
	resp, err := s.account.OpenPosition(s.Symbol, isBuy, s.Config.PositionSize, 20)
	if err != nil {
		fmt.Printf("[%s] ❌ Failed to open position: %v\n", s.ID, err)
		return
	}

	if resp.Success {
		s.Position = &Position{
			EntryPrice: price,
			EntryTime:  time.Now().UnixMilli(),
			Side:       side,
			Size:       s.Config.PositionSize,
			IsOpen:     true,
		}
		fmt.Printf("[%s] ✅ Position opened successfully: %s %.4f @ %.2f\n", s.ID, side, s.Config.PositionSize, price)
	} else {
		fmt.Printf("[%s] ❌ Position open failed: %s\n", s.ID, resp.Message)
	}
}

func (s *MaxTrendPointsStrategy) ClosePosition(reason string) {
	if s.Position == nil || !s.Position.IsOpen {
		return
	}

	fmt.Printf("[%s] Closing position: %s", s.ID, reason)
	resp, err := s.account.ClosePosition(s.Symbol, s.Position.Size)
	if err != nil {
		fmt.Printf("[%s] Failed to close position: %v", s.ID, err)
		return
	}

	s.Position.IsOpen = false
	s.Position.ExitReason = reason
	s.Position.ExitTime = time.Now().UnixMilli()
	fmt.Printf("[%s] Position closed: %s", s.ID, resp.Message)
}

func (s *MaxTrendPointsStrategy) calculateBacktestPositions(candles hyperliquid.Candles, signals []Signal) []Position {
	positions := []Position{}
	var currentPosition *Position

	for _, signal := range signals {
		if signal.Type != SignalLong && signal.Type != SignalShort {
			continue
		}

		side := "long"
		if signal.Type == SignalShort {
			side = "short"
		}

		if (s.Config.TradeDirection == "long" && side == "short") ||
			(s.Config.TradeDirection == "short" && side == "long") {
			continue
		}

		if currentPosition != nil && currentPosition.IsOpen {
			s.closeBacktestPosition(candles, currentPosition, signal.Index, signal.Price, "Trend Reversal")
			positions = append(positions, *currentPosition)
		}

		currentPosition = &Position{
			EntryIndex: signal.Index,
			EntryPrice: signal.Price,
			EntryTime:  signal.Time,
			Side:       side,
			Size:       s.Config.PositionSize,
			IsOpen:     true,
		}
	}

	if currentPosition != nil && currentPosition.IsOpen {
		lastCandle := candles[len(candles)-1]
		lastPrice := parseFloat(lastCandle.Close)
		s.closeBacktestPosition(candles, currentPosition, len(candles)-1, lastPrice, "End of Period")
		positions = append(positions, *currentPosition)
	}

	return positions
}

func (s *MaxTrendPointsStrategy) closeBacktestPosition(candles hyperliquid.Candles, position *Position, exitIndex int, exitPrice float64, reason string) {
	position.ExitIndex = exitIndex
	position.ExitPrice = exitPrice
	position.ExitTime = candles[exitIndex].Timestamp
	position.IsOpen = false
	position.ExitReason = reason

	priceDiff := 0.0
	if position.Side == "long" {
		priceDiff = exitPrice - position.EntryPrice
	} else {
		priceDiff = position.EntryPrice - exitPrice
	}

	position.PnLPercentage = (priceDiff / position.EntryPrice) * 100
	position.PnL = position.Size * position.EntryPrice * (position.PnLPercentage / 100)
}

func (s *MaxTrendPointsStrategy) calculateBacktestPnL(position *Position, currentPrice float64) float64 {
	priceDiff := 0.0
	if position.Side == "long" {
		priceDiff = currentPrice - position.EntryPrice
	} else {
		priceDiff = position.EntryPrice - currentPrice
	}
	percentageChange := (priceDiff / position.EntryPrice) * 100
	return position.Size * position.EntryPrice * (percentageChange / 100)
}

func (s *MaxTrendPointsStrategy) calculateTrends(candles hyperliquid.Candles) error {
	n := len(candles)
	s.output = &StrategyOutput{
		TrendLines:  make([]float64, n),
		TrendColors: make([]string, n),
		Lines:       []TrendLine{},
		Labels:      []Label{},
		FillColors:  make([]string, n),
		Directions:  make([]int, n),
	}
	if n < 200 {
		return fmt.Errorf("insufficient candles")
	}
	hl2 := make([]float64, n)
	highLowDiff := make([]float64, n)
	for i := range candles {
		high := parseFloat(candles[i].High)
		low := parseFloat(candles[i].Low)
		hl2[i] = (high + low) / 2
		highLowDiff[i] = high - low
	}
	dist := s.hma(highLowDiff, 200)
	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)
	for i := range candles {
		upperBand[i] = hl2[i] + s.Factor*dist[i]
		lowerBand[i] = hl2[i] - s.Factor*dist[i]
	}
	trendLine := make([]float64, n)
	direction := make([]int, n)
	for i := range candles {
		if i == 0 {
			direction[i] = 1
			trendLine[i] = upperBand[i]
		} else {
			close := parseFloat(candles[i-1].Close)
			if lowerBand[i] <= lowerBand[i-1] && close >= lowerBand[i-1] {
				lowerBand[i] = lowerBand[i-1]
			}
			if upperBand[i] >= upperBand[i-1] && close <= upperBand[i-1] {
				upperBand[i] = upperBand[i-1]
			}
			if dist[i-1] == 0 {
				direction[i] = 1
			} else if trendLine[i-1] == upperBand[i-1] {
				if parseFloat(candles[i].Close) > upperBand[i] {
					direction[i] = -1
				} else {
					direction[i] = 1
				}
			} else {
				if parseFloat(candles[i].Close) < lowerBand[i] {
					direction[i] = 1
				} else {
					direction[i] = -1
				}
			}
			if direction[i] == -1 {
				trendLine[i] = lowerBand[i]
			} else {
				trendLine[i] = upperBand[i]
			}
		}
		s.output.TrendLines[i] = trendLine[i]
		s.output.Directions[i] = direction[i]
		if direction[i] == 1 {
			s.output.TrendColors[i] = "#e49013"
		} else {
			s.output.TrendColors[i] = "#1cc2d8"
		}
	}
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
			if direction[i] == 1 {
				currentLineDn = &TrendLine{
					StartIndex: i,
					StartPrice: parseFloat(candles[i].Close),
					EndIndex:   i,
					EndPrice:   parseFloat(candles[i].Close),
					Direction:  1,
				}
				if currentLineUp != nil {
					s.output.Lines = append(s.output.Lines, *currentLineUp)
					currentLineUp = nil
				}
			} else {
				currentLineUp = &TrendLine{
					StartIndex: i,
					StartPrice: parseFloat(candles[i].Close),
					EndIndex:   i,
					EndPrice:   parseFloat(candles[i].Close),
					Direction:  -1,
				}
				if currentLineDn != nil {
					s.output.Lines = append(s.output.Lines, *currentLineDn)
					currentLineDn = nil
				}
			}
		} else {
			if direction[i] == -1 {
				highest = append(highest, parseFloat(candles[i].High))
				if currentLineUp != nil && len(highest) > 0 {
					maxIdx, maxVal := s.findMax(highest)
					currentLineUp.EndIndex = start + maxIdx + 1
					currentLineUp.EndPrice = maxVal
				}
			} else {
				lowest = append(lowest, parseFloat(candles[i].Low))
				if currentLineDn != nil && len(lowest) > 0 {
					minIdx, minVal := s.findMin(lowest)
					currentLineDn.EndIndex = start + minIdx + 1
					currentLineDn.EndPrice = minVal
				}
			}
		}
	}
	if currentLineUp != nil {
		s.output.Lines = append(s.output.Lines, *currentLineUp)
	}
	if currentLineDn != nil {
		s.output.Lines = append(s.output.Lines, *currentLineDn)
	}
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

func (s *MaxTrendPointsStrategy) calculateBacktestOutput(positions []Position) BacktestOutput {
	result := BacktestOutput{
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

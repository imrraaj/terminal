package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

type Source struct {
	client *hyperliquid.Client
	info   *hyperliquid.Info
	ctx    context.Context
}

func NewSource() *Source {
	info := hyperliquid.NewInfo(context.Background(), hyperliquid.MainnetAPIURL, true, nil, nil)
	client := hyperliquid.NewClient(hyperliquid.MainnetAPIURL)
	return &Source{info: info, client: client}
}

func (s *Source) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Source) FetchHistoricalCandles(symbol string, interval string, limit int) ([]hyperliquid.Candle, error) {
	return s.FetchCandlesBefore(symbol, interval, limit, 0)
}

func (s *Source) FetchCandlesBefore(symbol string, interval string, limit int, beforeTimestamp int64) ([]hyperliquid.Candle, error) {
	const maxCandlesPerRequest = 5000

	if limit <= maxCandlesPerRequest {
		return s.fetchSingleBatch(symbol, interval, limit, beforeTimestamp)
	}

	// For limits > 5000, fetch in batches
	var allCandles []hyperliquid.Candle
	remaining := limit
	currentEndTime := beforeTimestamp

	for remaining > 0 {
		batchSize := remaining
		if batchSize > maxCandlesPerRequest {
			batchSize = maxCandlesPerRequest
		}

		batch, err := s.fetchSingleBatch(symbol, interval, batchSize, currentEndTime)
		if err != nil {
			return nil, err
		}

		if len(batch) == 0 {
			break
		}

		allCandles = append(batch, allCandles...)
		currentEndTime = batch[0].Timestamp
		remaining -= len(batch)
		if len(batch) < batchSize {
			break
		}
	}

	if len(allCandles) == 0 {
		return nil, fmt.Errorf("Expected non-empty candles response")
	}

	return allCandles, nil
}

func (s *Source) fetchSingleBatch(symbol string, interval string, limit int, beforeTimestamp int64) ([]hyperliquid.Candle, error) {
	var endTime time.Time
	if beforeTimestamp > 0 {
		endTime = time.Unix(beforeTimestamp/1000, 0)
	} else {
		endTime = time.Now()
	}

	var intervalDuration time.Duration

	switch interval {
	case "1m":
		intervalDuration = time.Minute
	case "5m":
		intervalDuration = 5 * time.Minute
	case "15m":
		intervalDuration = 15 * time.Minute
	case "1h":
		intervalDuration = time.Hour
	case "4h":
		intervalDuration = 4 * time.Hour
	case "1d":
		intervalDuration = 24 * time.Hour
	default:
		intervalDuration = 5 * time.Minute
	}

	startTime := endTime.Add(-time.Duration(limit) * intervalDuration)
	Coin := symbol
	Interval := interval
	StartTime := startTime.Unix() * 1000
	EndTime := endTime.Unix() * 1000

	candles, err := s.info.CandlesSnapshot(context.TODO(), Coin, Interval, StartTime, EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles: %w", err)
	}

	return candles, nil
}

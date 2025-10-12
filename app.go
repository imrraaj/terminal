package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)

type App struct {
	ctx      context.Context
	rdb      *redis.Client
	cache    *CandleCache
	source   *Source
	strategy Strategy
	Account  *Account
	engine   *StrategyEngine
	config   *Config
}

func NewApp() *App {
	return &App{
		source:   NewSource(),
		strategy: NewMaxTrendPointsStrategy(8, "#1cc2d8", "#e49013"),
		rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		config: NewConfig(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.source.SetContext(ctx)
	a.cache = NewCandleCache(a.rdb, ctx)
	a.source.SetCache(a.cache)
	a.Account = NewAccount(ctx, a.config)
	a.engine = NewStrategyEngine(a.source, a.Account)
}

func (a *App) shutdown() {
	if a.engine != nil {
		a.engine.StopAllStrategies()
	}
}

func (a *App) FetchCandles(symbol string, interval string, limit int) (hyperliquid.Candles, error) {
	return a.cache.GetOrFetch(symbol, interval, limit, func() ([]hyperliquid.Candle, error) {
		return a.source.FetchHistoricalCandles(symbol, interval, limit)
	})
}

func (a *App) FetchCandlesBefore(symbol string, interval string, limit int, beforeTimestamp int64) (hyperliquid.Candles, error) {
	return a.source.FetchCandlesBefore(symbol, interval, limit, beforeTimestamp)
}

func (a *App) StrategyRun(candles hyperliquid.Candles) (*StrategyOutputV2, error) {
	config := a.strategy.GetDefaultConfig()
	runner := NewStrategyRunner(a.strategy, candles, config)
	return runner.Run()
}

func (a *App) ApplyStrategy(strategyId string, params map[string]any) error {
	if strategyId == "max-trend" {
		factor := 8.0
		colorUp := "#1cc2d8"
		colorDn := "#e49013"
		if f, ok := params["factor"].(float64); ok {
			factor = f
		}
		if c, ok := params["colorUp"].(string); ok {
			colorUp = c
		}
		if c, ok := params["colorDn"].(string); ok {
			colorDn = c
		}
		fmt.Println("Applying strategy with params:", factor, colorUp, colorDn)
		a.strategy = NewMaxTrendPointsStrategy(factor, colorUp, colorDn)
	}
	return nil
}

func (a *App) FetchAndApplyStrategy(symbol string, interval string, limit int, strategyId string, params map[string]any) (*StrategyOutputV2, error) {
	candles, err := a.cache.GetOrFetch(symbol, interval, limit, func() ([]hyperliquid.Candle, error) {
		return a.source.FetchHistoricalCandles(symbol, interval, limit)
	})
	if err != nil {
		return nil, err
	}
	config := StrategyConfig{
		TakeProfitPercent: 10.0,
		StopLossPercent:   5.0,
		PositionSize:      1.0,
		UsePercentage:     false,
		MaxPositions:      1,
		TradeDirection:    "both",
		Parameters:        params,
	}
	if tp, ok := params["takeProfitPercent"].(float64); ok {
		config.TakeProfitPercent = tp
	}
	if sl, ok := params["stopLossPercent"].(float64); ok {
		config.StopLossPercent = sl
	}
	if size, ok := params["positionSize"].(float64); ok {
		config.PositionSize = size
	}
	if maxPos, ok := params["maxPositions"].(float64); ok {
		config.MaxPositions = int(maxPos)
	}
	if direction, ok := params["tradeDirection"].(string); ok {
		config.TradeDirection = direction
	}
	if strategyId == "max-trend" {
		factor := 2.5
		colorUp := "#1cc2d8"
		colorDn := "#e49013"
		if f, ok := params["factor"].(float64); ok {
			factor = f
		}
		if c, ok := params["colorUp"].(string); ok {
			colorUp = c
		}
		if c, ok := params["colorDn"].(string); ok {
			colorDn = c
		}
		a.strategy = NewMaxTrendPointsStrategy(factor, colorUp, colorDn)
	}
	runner := NewStrategyRunner(a.strategy, candles, config)
	return runner.Run()
}

func (a *App) StartLiveStrategy(id string, strategyId string, params map[string]any, symbol string, interval string) error {
	config := StrategyConfig{
		TakeProfitPercent: 10.0,
		StopLossPercent:   5.0,
		PositionSize:      1.0,
		UsePercentage:     false,
		MaxPositions:      1,
		TradeDirection:    "both",
		Parameters:        params,
	}
	if tp, ok := params["takeProfitPercent"].(float64); ok {
		config.TakeProfitPercent = tp
	}
	if sl, ok := params["stopLossPercent"].(float64); ok {
		config.StopLossPercent = sl
	}
	if size, ok := params["positionSize"].(float64); ok {
		config.PositionSize = size
	}
	if maxPos, ok := params["maxPositions"].(float64); ok {
		config.MaxPositions = int(maxPos)
	}
	if direction, ok := params["tradeDirection"].(string); ok {
		config.TradeDirection = direction
	}
	var strategy Strategy
	if strategyId == "max-trend" {
		factor := 2.5
		colorUp := "#1cc2d8"
		colorDn := "#e49013"
		if f, ok := params["factor"].(float64); ok {
			factor = f
		}
		if c, ok := params["colorUp"].(string); ok {
			colorUp = c
		}
		if c, ok := params["colorDn"].(string); ok {
			colorDn = c
		}
		strategy = NewMaxTrendPointsStrategy(factor, colorUp, colorDn)
	} else {
		return fmt.Errorf("unknown strategy: %s", strategyId)
	}
	return a.engine.StartStrategy(id, strategy, config, symbol, interval)
}

func (a *App) StopLiveStrategy(id string) error {
	return a.engine.StopStrategy(id)
}

func (a *App) GetRunningStrategies() []StrategyInstance {
	return a.engine.GetRunningStrategies()
}

func (a *App) GetWalletAddress() string {
	return a.Account.address
}

func (a *App) GetConfig() *Config {
	return a.config
}

func (a *App) SetConfigURL(url string) {
	a.config.SetSourceURL(url)
	// Reinitialize account with new URL
	a.Account = NewAccount(a.ctx, a.config)
	a.engine = NewStrategyEngine(a.source, a.Account)
}

func (a *App) GetPortfolioSummary() (*PortfolioSummary, error) {
	return a.Account.GetPortfolioSummary()
}

func (a *App) GetActivePositions() ([]ActivePosition, error) {
	return a.Account.GetActivePositions()
}

func (a *App) IsRedisConnected() bool {
	if a.cache == nil {
		return false
	}
	return a.cache.IsConnected()
}

func (a *App) InvalidateCache() error {
	if a.cache == nil {
		return nil
	}
	return a.cache.Clear()
}

func (a *App) InvalidateCacheForSymbol(symbol string) error {
	if a.cache == nil {
		return nil
	}
	return a.cache.InvalidateSymbol(symbol)
}

func (a *App) SpotBuy(symbol string, size float64) (*OrderResponse, error) {
	fmt.Printf("Executing spot buy order for %s with size %.4f\n", symbol, size)

	resp, err := a.Account.SpotOrder(symbol, true, size)
	if err != nil {
		return nil, fmt.Errorf("error executing spot buy: %w", err)
	}

	fmt.Printf("Spot buy order executed successfully: %+v\n", resp)
	return resp, nil
}

func (a *App) SpotSell(symbol string, size float64) (*OrderResponse, error) {
	fmt.Printf("Executing spot sell order for %s with size %.4f\n", symbol, size)

	resp, err := a.Account.SpotOrder(symbol, false, size)
	if err != nil {
		return nil, fmt.Errorf("error executing spot sell: %w", err)
	}

	fmt.Printf("Spot sell order executed successfully: %+v\n", resp)
	return resp, nil
}

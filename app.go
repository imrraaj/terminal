package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)

type App struct {
	ctx     context.Context
	rdb     *redis.Client
	source  *Source
	account *Account
	engine  *StrategyEngine
	config  Config
}

func NewApp() *App {
	config := NewConfig()
	return &App{
		source: NewSource(config),
		rdb: redis.NewClient(&redis.Options{
			Addr: config.RedisURL,
		}),
		config: config,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.source.SetContext(ctx)
	a.source.SetRedis(a.rdb)
	a.account = NewAccount(ctx, a.config)
	a.engine = NewStrategyEngine(a.source)
}

func (a *App) shutdown(ctx context.Context) {
	a.engine.StopAllStrategies()
}

func (a *App) FetchCandles(symbol string, interval string, limit int) (hyperliquid.Candles, error) {
	return a.source.FetchHistoricalCandles(symbol, interval, limit)
}

func (a *App) FetchCandlesBefore(symbol string, interval string, limit int, beforeTimestamp int64) (hyperliquid.Candles, error) {
	return a.source.FetchCandlesBefore(symbol, interval, limit, beforeTimestamp)
}

func (a *App) StrategyRun(name, symbol string, interval string, params map[string]any) error {
	log.Printf("Strategy Run: %s %s %s %v\n", name, symbol, interval, params)
	strategy := NewMaxTrendPointsStrategy(params)
	strategy.Symbol = symbol
	strategy.Interval = interval
	strategy.account = a.account
	return a.engine.StartStrategy(name, *strategy)
}

func (a *App) StrategyBacktest(symbol string, interval string, limit int, params map[string]any) (*BacktestOutput, error) {
	candles, err := a.source.FetchHistoricalCandles(symbol, interval, limit)
	if err != nil {
		return nil, err
	}
	strategy := NewMaxTrendPointsStrategy(params)
	return strategy.Backtest(candles)
}

func (a *App) StopLiveStrategy(name string) error {
	return a.engine.StopStrategy(name)
}

func (a *App) GetRunningStrategies() []MaxTrendPointsStrategy {
	return a.engine.GetRunningStrategies()
}

func (a *App) GetWalletAddress() string {
	return a.account.address
}

func (a *App) GetPortfolioSummary() (PortfolioSummary, error) {
	return a.account.GetPortfolioSummary()
}

func (a *App) GetActivePositions() ([]ActivePosition, error) {
	return a.account.GetActivePositions()
}

func (a *App) InvalidateCache() error {
	return a.source.InvalidateCache()
}

func (a *App) InvalidateCacheForSymbol(symbol string) error {
	return a.source.InvalidateCacheForSymbol(symbol)
}

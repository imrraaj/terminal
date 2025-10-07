package main

import (
	"context"
	"fmt"

	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// App struct
type App struct {
	ctx      context.Context
	source   *Source
	strategy Strategy
	account  *Account
	engine   *StrategyEngine
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		source:   NewSource(),
		strategy: NewMaxTrendPointsStrategy(2.5, "#1cc2d8", "#e49013"),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.source.SetContext(ctx)
	a.account = NewAccount(ctx)
	a.engine = NewStrategyEngine(a.source, a.account)
}

func (a *App) shutdown(ctx context.Context) {
	// Stop all running strategies on shutdown
	if a.engine != nil {
		a.engine.StopAllStrategies()
	}
}

func (a *App) FetchCandles(symbol string, interval string, limit int) (hyperliquid.Candles, error) {
	return a.source.FetchHistoricalCandles(symbol, interval, limit)
}

func (a *App) FetchCandlesBefore(symbol string, interval string, limit int, beforeTimestamp int64) (hyperliquid.Candles, error) {
	return a.source.FetchCandlesBefore(symbol, interval, limit, beforeTimestamp)
}

// StrategyRun runs the current strategy and returns both visualization and backtest results
func (a *App) StrategyRun(candles hyperliquid.Candles) (*StrategyOutputV2, error) {
	config := a.strategy.GetDefaultConfig()
	runner := NewStrategyRunner(a.strategy, candles, config)
	return runner.Run()
}

// ApplyStrategy applies a strategy with given parameters
func (a *App) ApplyStrategy(strategyId string, params map[string]interface{}) error {
	if strategyId == "max-trend" {
		// Extract parameters with defaults
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

		fmt.Println("Applying strategy with params:", factor, colorUp, colorDn)
		a.strategy = NewMaxTrendPointsStrategy(factor, colorUp, colorDn)
	}

	return nil
}

// FetchAndApplyStrategy fetches candles and applies strategy in one call
func (a *App) FetchAndApplyStrategy(symbol string, interval string, limit int, strategyId string, params map[string]interface{}) (*StrategyOutputV2, error) {
	// Fetch candles
	candles, err := a.source.FetchHistoricalCandles(symbol, interval, limit)
	if err != nil {
		return nil, err
	}

	// Build config from params
	config := StrategyConfig{
		TakeProfitPercent: 10.0,
		StopLossPercent:   5.0,
		PositionSize:      1.0,
		UsePercentage:     false,
		MaxPositions:      1,
		TradeDirection:    "both", // Default to both
		Parameters:        params,
	}

	// Extract TP/SL and trade direction if provided
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

	// Apply strategy with parameters
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

	// Run strategy with backtest
	runner := NewStrategyRunner(a.strategy, candles, config)
	return runner.Run()
}

// Account-related methods

// ConnectWallet connects to a wallet using private key
func (a *App) ConnectWallet(privateKeyHex string) error {
	return a.account.ConnectWallet(privateKeyHex)
}

// GetWalletAddress returns the connected wallet address
func (a *App) GetWalletAddress() string {
	return a.account.GetAddress()
}

// IsWalletConnected returns whether a wallet is connected
func (a *App) IsWalletConnected() bool {
	return a.account.IsConnected()
}

// GetPortfolioSummary fetches complete portfolio information
func (a *App) GetPortfolioSummary() (*PortfolioSummary, error) {
	return a.account.GetPortfolioSummary()
}

// GetActivePositions returns only the active positions
func (a *App) GetActivePositions() ([]ActivePosition, error) {
	return a.account.GetActivePositions()
}

// OpenTrade places a new order to open a position
func (a *App) OpenTrade(req OrderRequest) (*OrderResponse, error) {
	return a.account.OpenTrade(req)
}

// ClosePosition closes an open position (full or partial)
func (a *App) ClosePosition(req ClosePositionRequest) (*OrderResponse, error) {
	return a.account.ClosePosition(req)
}

// OpenLongPosition is a convenience method to open a long position
func (a *App) OpenLongPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.account.OpenLongPosition(coin, size, price, orderType)
}

// OpenShortPosition is a convenience method to open a short position
func (a *App) OpenShortPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.account.OpenShortPosition(coin, size, price, orderType)
}

// Strategy Engine methods

// StartLiveStrategy starts a strategy in live trading mode
func (a *App) StartLiveStrategy(id string, strategyId string, params map[string]interface{}, symbol string, interval string) error {
	// Build config from params
	config := StrategyConfig{
		TakeProfitPercent: 10.0,
		StopLossPercent:   5.0,
		PositionSize:      1.0,
		UsePercentage:     false,
		MaxPositions:      1,
		TradeDirection:    "both",
		Parameters:        params,
	}

	// Extract TP/SL and trade direction if provided
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

	// Create strategy instance
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

	// Start strategy in engine
	return a.engine.StartStrategy(id, strategy, config, symbol, interval)
}

// StopLiveStrategy stops a running live strategy
func (a *App) StopLiveStrategy(id string) error {
	return a.engine.StopStrategy(id)
}

// GetRunningStrategies returns all running strategy instances
func (a *App) GetRunningStrategies() []StrategyInstance {
	return a.engine.GetRunningStrategies()
}

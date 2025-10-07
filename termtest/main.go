package main

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// OpenLong opens a long position on Hyperliquid
func OpenLong(ctx context.Context, exchange *hyperliquid.Exchange, coin string, size float64, slippage float64) (hyperliquid.OrderStatus, error) {
	return exchange.MarketOpen(ctx, coin, true, size, nil, slippage, nil, nil)
}

// CloseLong closes a long position on Hyperliquid by placing a reduce-only sell order
func CloseLong(ctx context.Context, exchange *hyperliquid.Exchange, coin string, size float64, slippage float64) (hyperliquid.OrderStatus, error) {
	// Get slippage price for selling
	slippagePrice, err := exchange.SlippagePrice(ctx, coin, false, slippage, nil)
	if err != nil {
		return hyperliquid.OrderStatus{}, err
	}

	// Place reduce-only order to close long
	return exchange.Order(ctx, hyperliquid.CreateOrderRequest{
		Coin:       coin,
		IsBuy:      false, // sell to close long
		Size:       size,
		Price:      slippagePrice,
		ReduceOnly: true,
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc},
		},
	}, nil)
}

func main() {
	ctx := context.Background()

	// Load private key
	privateKey, err := crypto.HexToECDSA("fcbd8c1e9300c87d420d2593100ca24e4b09a3d84fa6b1c7c600252a7f6d28be")
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Initialize exchange client
	exchange := hyperliquid.NewExchange(
		ctx,
		privateKey,
		hyperliquid.TestnetAPIURL,
		nil, // meta will be fetched
		"",  // vault address (empty if not using vault)
		"",  // account address (optional)
		nil, // spot meta
	)

	// Initialize info client for fetching market data
	info := hyperliquid.NewInfo(
		ctx,
		hyperliquid.TestnetAPIURL,
		true,
		nil, // meta will be fetched
		nil, // spot meta
	)

	// Strategy configuration
	coin := "BTC"
	positionSize := 0.01 // BTC

	// Create strategy
	strategy := NewMaxTrendPointsStrategy(ctx, exchange, info, coin, positionSize)

	// Create runner that executes strategy every 5 minutes
	runner := NewRunner(5*time.Minute, strategy)

	log.Printf("Starting strategy runner for %s with position size %.4f", coin, positionSize)
	log.Println("Strategy: Max Trend Points (5m chart, factor=8)")
	log.Println("Checking for signals every 2 minutes...")

	// Start the runner (blocks forever)
	runner.Start()
}

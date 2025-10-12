package main

import (
	// "embed"
	"context"
	"fmt"
	// "github.com/wailsapp/wails/v2"
	// "github.com/wailsapp/wails/v2/pkg/options"
	// "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

/// go:embed all:frontend/dist
// var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	ctx := context.Background()
	app.startup(ctx)
	// fmt.Println(app.Account.address)
	// fmt.Println(app.Account.privateKey)
	// resp,err := app.engine.account.SpotOrder("BTC", true, 0.001)
	// if err != nil {
	//     fmt.Printf(err.Error())
	// }
	// fmt.Printf(resp.Message)

	// Working

	// exchange := hyperliquid.NewExchange(ctx, app.Account.privateKey, hyperliquid.TestnetAPIURL, nil, "", "", nil, hyperliquid.ExchangeOptClientOptions())
	// exchange.UpdateLeverage(ctx, 2, "BTC", false)
	// resp, err := exchange.MarketOpen(ctx, "BTC", true, 0.005, nil, 0.05, nil, nil)
	// if err != nil {
	//     fmt.Printf(err.Error())
	// }
	// fmt.Printf(resp.String())

	// To BUY without leverage, set leverage to 1x and place a perp order
	// This behaves like a spot buy (no liquidation risk at 1x)
	// _, err := exchange.UpdateLeverage(ctx, 1, "ETH", false)
	// if err != nil {
	//     fmt.Printf("Leverage error: %s\n", err.Error())
	// }

    resp, err := app.engine.account.OpenPerpOrder("BTC", true, 1, 10)
    if err != nil {
        fmt.Printf("Order error: %s\n", err.Error())
    } else {
        fmt.Printf("Success: %s\n", resp.Message)
    }
	// resp, err := exchange.Order(ctx, hyperliquid.CreateOrderRequest{
	// 	Coin:  "HYPE/USDC",
	// 	IsBuy: true,
	// 	Size:  1,
	// }, nil)
	// if err != nil {
	// 	fmt.Printf("Order error: %s\n", err.Error())
	// } else {
	// 	fmt.Printf("Success: %s\n", resp.String())
	// }

	// Create application with options
	// err := wails.Run(&options.App{
	// 	Title:  "HyperTerminal",
	// 	Width:  1024,
	// 	Height: 768,
	// 	AssetServer: &assetserver.Options{
	// 		Assets: assets,
	// 	},
	// 	BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
	// 	OnStartup:        app.startup,
	// 	OnShutdown:       app.shutdown,
	// 	Fullscreen:       false,
	// 	Bind: []any{
	// 		app,
	// 	},
	// })

	// if err != nil {
	// 	println("Error:", err.Error())
	// }
}

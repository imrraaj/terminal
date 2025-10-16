package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/sonirico/go-hyperliquid"
)

type Account struct {
	ctx        context.Context
	info       *hyperliquid.Info
	exchange   *hyperliquid.Exchange
	address    string
	privateKey *ecdsa.PrivateKey
}

type AccountBalance struct {
	AccountValue    string
	TotalRawUsd     string
	Withdrawable    string
	TotalMargin     string
	AccountLeverage float64
}

type ActivePosition struct {
	Coin           string
	Side           string
	Size           string
	EntryPrice     string
	CurrentPrice   string
	LiquidationPx  string
	UnrealizedPnl  string
	PositionValue  string
	Leverage       int
	MarginUsed     string
	ReturnOnEquity string
}

type PortfolioSummary struct {
	Balance        AccountBalance
	Positions      []ActivePosition
	TotalPositions int
	TotalPnL       float64
	OpenOrders     []hyperliquid.OpenOrder
}

type OrderResponse struct {
	Success bool
	OrderID string
	Message string
	Status  string
}

func NewAccount(ctx context.Context, config Config) *Account {
	exchange := hyperliquid.NewExchange(ctx, config.PrivateKey, config.URL, nil, "", "", nil, hyperliquid.ExchangeOptClientOptions())
	info := hyperliquid.NewInfo(ctx, config.URL, true, nil, nil, hyperliquid.InfoOptClientOptions())

	return &Account{
		ctx:        ctx,
		exchange:   exchange,
		address:    config.Address,
		privateKey: config.PrivateKey,
		info:       info,
	}
}

func (a *Account) GetAddress() string {
	return a.address
}

func (a *Account) GetPortfolioSummary() (PortfolioSummary, error) {
	userState, err := a.info.UserState(a.ctx, a.address)
	if err != nil {
		return PortfolioSummary{}, fmt.Errorf("failed to fetch user state: %w", err)
	}

	bal := AccountBalance{
		AccountValue: userState.MarginSummary.AccountValue,
		TotalRawUsd:  userState.MarginSummary.TotalRawUsd,
		Withdrawable: userState.Withdrawable,
		TotalMargin:  userState.MarginSummary.TotalMarginUsed,
	}

	positions := make([]ActivePosition, 0, len(userState.AssetPositions))
	totalPnL := 0.0

	for _, assetPos := range userState.AssetPositions {
		pos := assetPos.Position
		sizeF := parseFloatSafe(pos.Szi)
		if sizeF == 0 {
			continue
		}

		side := "long"
		if sizeF < 0 {
			side = "short"
			sizeF = -sizeF
		}

		entry := ""
		if pos.EntryPx != nil {
			entry = *pos.EntryPx
		}

		liq := ""
		if pos.LiquidationPx != nil {
			liq = *pos.LiquidationPx
		}

		positions = append(positions, ActivePosition{
			Coin:           pos.Coin,
			Side:           side,
			Size:           fmt.Sprintf("%.8f", sizeF),
			EntryPrice:     entry,
			LiquidationPx:  liq,
			UnrealizedPnl:  pos.UnrealizedPnl,
			PositionValue:  pos.PositionValue,
			Leverage:       pos.Leverage.Value,
			MarginUsed:     pos.MarginUsed,
			ReturnOnEquity: pos.ReturnOnEquity,
		})

		totalPnL += parseFloatSafe(pos.UnrealizedPnl)
	}

	openOrders, err := a.info.OpenOrders(a.ctx, a.address)
	if err != nil {
		return PortfolioSummary{}, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	return PortfolioSummary{
		Balance:        bal,
		Positions:      positions,
		TotalPositions: len(positions),
		TotalPnL:       totalPnL,
		OpenOrders:     openOrders,
	}, nil
}

func (a *Account) GetActivePositions() ([]ActivePosition, error) {
	sum, err := a.GetPortfolioSummary()
	if err != nil {
		return nil, err
	}
	return sum.Positions, nil
}

func (a *Account) OpenPosition(coin string, isBuy bool, size float64, leverage int) (OrderResponse, error) {
	_, err := a.exchange.UpdateLeverage(a.ctx, leverage, coin, false)
	if err != nil {
		return OrderResponse{
			Success: false,
			Message: fmt.Sprintf("failed to set leverage: %v", err),
			Status:  "error",
		}, err
	}

	resp, err := a.exchange.MarketOpen(a.ctx, coin, isBuy, size, nil, 0.05, nil, nil)
	if err != nil {
		return OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	return parseOrderResponse(resp), nil
}

func (a *Account) ClosePosition(coin string, size float64) (OrderResponse, error) {
	userState, err := a.info.UserState(a.ctx, a.address)
	if err != nil {
		return OrderResponse{
			Success: false,
			Message: fmt.Sprintf("failed to fetch position: %v", err),
			Status:  "error",
		}, err
	}

	var positionSize float64
	var isBuy bool
	found := false

	for _, assetPos := range userState.AssetPositions {
		if assetPos.Position.Coin == coin {
			szi := parseFloatSafe(assetPos.Position.Szi)
			if szi == 0 {
				return OrderResponse{
					Success: false,
					Message: "no open position for " + coin,
					Status:  "error",
				}, fmt.Errorf("no open position for %s", coin)
			}

			isBuy = szi < 0
			if size > 0 {
				positionSize = size
			} else {
				positionSize = abs(szi)
			}
			found = true
			break
		}
	}

	if !found {
		return OrderResponse{
			Success: false,
			Message: "position not found for " + coin,
			Status:  "error",
		}, fmt.Errorf("position not found for %s", coin)
	}

	slippagePrice, err := a.exchange.SlippagePrice(a.ctx, coin, isBuy, 0.05, nil)
	if err != nil {
		return OrderResponse{
			Success: false,
			Message: fmt.Sprintf("failed to get slippage price: %v", err),
			Status:  "error",
		}, err
	}

	resp, err := a.exchange.Order(a.ctx, hyperliquid.CreateOrderRequest{
		Coin:       coin,
		IsBuy:      isBuy,
		Size:       positionSize,
		Price:      slippagePrice,
		OrderType:  hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}},
		ReduceOnly: true,
	}, nil)

	if err != nil {
		return OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	return parseOrderResponse(resp), nil
}

func parseFloatSafe(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func parseOrderResponse(resp hyperliquid.OrderStatus) OrderResponse {
	out := OrderResponse{Success: true}
	if resp.Resting != nil {
		out.Status = resp.Resting.Status
		out.OrderID = fmt.Sprintf("%d", resp.Resting.Oid)
	} else if resp.Filled != nil {
		out.Status = "filled"
		out.Message = fmt.Sprintf("filled avgPx=%s size=%s", resp.Filled.AvgPx, resp.Filled.TotalSz)
	} else if resp.Error != nil {
		out.Success = false
		out.Status = "error"
		out.Message = *resp.Error
	}
	return out
}

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

	OpenOrders []hyperliquid.OpenOrder
}

type OrderRequest struct {
	Coin       string
	IsBuy      bool
	Size       float64
	Price      float64
	OrderType  string
	ReduceOnly bool
}

type OrderResponse struct {
	Success bool
	OrderID string
	Message string
	Status  string
}

type ClosePositionRequest struct {
	Coin string
	Size float64
}

func NewAccount(ctx context.Context, walletAddress string) *Account {
	return &Account{
		ctx:      ctx,
		exchange: hyperliquid.NewExchange(ctx, nil, hyperliquid.MainnetAPIURL, nil, "", "", nil),
		address:  walletAddress,
		info:     hyperliquid.NewInfo(ctx, hyperliquid.MainnetAPIURL, true, nil, nil),
	}
}

func parseFloatSafe(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (a *Account) GetPortfolioSummary() (*PortfolioSummary, error) {
	userState, err := a.info.UserState(a.ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
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

		ap := ActivePosition{
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
		}

		positions = append(positions, ap)
		totalPnL += parseFloatSafe(pos.UnrealizedPnl)
	}

	openOrders, err := a.info.OpenOrders(a.ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	return &PortfolioSummary{
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

func (a *Account) OpenTrade(req OrderRequest) (*OrderResponse, error) {
	if a.exchange == nil {
		return nil, fmt.Errorf("exchange not initialized")
	}

	orderReq := hyperliquid.CreateOrderRequest{
		Coin:       req.Coin,
		IsBuy:      req.IsBuy,
		Size:       req.Size,
		Price:      req.Price,
		ReduceOnly: req.ReduceOnly,
	}

	if req.OrderType == "market" {
		orderReq.OrderType = hyperliquid.OrderType{
			Trigger: &hyperliquid.TriggerOrderType{
				TriggerPx: req.Price,
				IsMarket:  true,
			},
		}
	} else {
		orderReq.OrderType = hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifGtc,
			},
		}
	}

	resp, err := a.exchange.Order(a.ctx, orderReq, nil)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	out := &OrderResponse{Success: true}
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
	return out, nil
}

func (a *Account) ClosePosition(req ClosePositionRequest) (*OrderResponse, error) {
	positions, err := a.GetActivePositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	var pos *ActivePosition
	for i := range positions {
		if positions[i].Coin == req.Coin {
			pos = &positions[i]
			break
		}
	}
	if pos == nil {
		return nil, fmt.Errorf("no open position for %s", req.Coin)
	}

	size := req.Size
	if size == 0 {
		size = parseFloatSafe(pos.Size)
	}
	isBuy := pos.Side == "short"

	closeReq := OrderRequest{
		Coin:       req.Coin,
		IsBuy:      isBuy,
		Size:       size,
		Price:      0,
		OrderType:  "market",
		ReduceOnly: true,
	}
	return a.OpenTrade(closeReq)
}

func (a *Account) GetOpenOrders() ([]hyperliquid.OpenOrder, error) {
	return a.info.OpenOrders(a.ctx, a.address)
}

func (a *Account) CancelOrder(coin string, orderId int64) error {
	if a.exchange == nil {
		return fmt.Errorf("exchange not initialized")
	}
	_, err := a.exchange.Cancel(a.ctx, coin, orderId)
	if err != nil {
		return fmt.Errorf("cancel failed: %w", err)
	}
	return nil
}

func (a *Account) CancelAllOrders(coin string) error {
	orders, err := a.GetOpenOrders()
	if err != nil {
		return fmt.Errorf("failed fetch open orders: %w", err)
	}
	for _, o := range orders {
		if o.Coin == coin {
			if err := a.CancelOrder(coin, o.Oid); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Account) OpenLongPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.OpenTrade(OrderRequest{
		Coin:      coin,
		IsBuy:     true,
		Size:      size,
		Price:     price,
		OrderType: orderType,
	})
}

func (a *Account) OpenShortPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.OpenTrade(OrderRequest{
		Coin:      coin,
		IsBuy:     false,
		Size:      size,
		Price:     price,
		OrderType: orderType,
	})
}

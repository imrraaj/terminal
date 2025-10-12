package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonirico/go-hyperliquid"
)

type Account struct {
	ctx        context.Context
	info       hyperliquid.Info
	exchange   hyperliquid.Exchange
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

func NewAccount(ctx context.Context, config Config) Account {
	var privateKey *ecdsa.PrivateKey
	var address string

	if config.PrivateKey != "" {
		var err error
		privateKey, err = crypto.HexToECDSA(config.PrivateKey)
		if err != nil {
			panic(fmt.Errorf("invalid private key: %w", err))
		}
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			panic(fmt.Errorf("failed to cast public key to ECDSA"))
		}
		address = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	}

	// CRITICAL FIX: Use privateKey (not nil) when creating exchange
	exchange := hyperliquid.NewExchange(ctx, privateKey, config.URL, nil, "", "", nil, hyperliquid.ExchangeOptClientOptions())
	info := hyperliquid.NewInfo(ctx, config.URL, true, nil, nil, hyperliquid.InfoOptClientOptions())

	return Account{
		ctx:        ctx,
		exchange:   *exchange,
		address:    address,
		privateKey: privateKey,
		info:       *info,
	}
}

func parseFloatSafe(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (a *Account) GetAddress() string {
	return a.address
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

// SpotOrder places a spot market order (1x leverage perpetual on testnet)
func (a *Account) SpotOrder(coin string, isBuy bool, size float64) (*OrderResponse, error) {
	// On Hyperliquid testnet, true spot trading is limited
	// Using 1x leverage perpetual as spot-like behavior
	_, err := a.exchange.UpdateLeverage(a.ctx, 1, coin, false)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: fmt.Sprintf("failed to set 1x leverage: %v", err),
			Status:  "error",
		}, err
	}

	price := 10000.0 // High price for market buy, low for market sell
	if !isBuy {
		price = 0.01
	}

	orderReq := hyperliquid.CreateOrderRequest{
		Coin:  coin,
		IsBuy: isBuy,
		Size:  size,
		Price: price,
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifIoc, // Immediate or Cancel = market order
			},
		},
	}

	resp, err := a.exchange.Order(a.ctx, orderReq, nil)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	return parseOrderResponse(resp), nil
}

// OpenPerpOrder opens a leveraged perpetual position
func (a *Account) OpenPerpOrder(coin string, isBuy bool, size float64, leverage int) (*OrderResponse, error) {
	if leverage < 1 || leverage > 50 {
		return &OrderResponse{
			Success: false,
			Message: fmt.Sprintf("invalid leverage: %d (must be 1-50)", leverage),
			Status:  "error",
		}, fmt.Errorf("invalid leverage: %d", leverage)
	}

	_, err := a.exchange.UpdateLeverage(a.ctx, leverage, coin, false)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: fmt.Sprintf("failed to set leverage: %v", err),
			Status:  "error",
		}, err
	}

	price := 100000.0 // High price for market buy
	if !isBuy {
		price = 0.01 // Low price for market sell
	}

	orderReq := hyperliquid.CreateOrderRequest{
		Coin:  coin,
		IsBuy: isBuy,
		Size:  size,
		Price: price,
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifIoc, // Immediate or Cancel = market order
			},
		},
	}

	resp, err := a.exchange.Order(a.ctx, orderReq, nil)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	return parseOrderResponse(resp), nil
}

// ClosePosition closes an existing position
func (a *Account) ClosePosition(coin string, size float64) (*OrderResponse, error) {
	positions, err := a.GetActivePositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var pos *ActivePosition
	for i := range positions {
		if positions[i].Coin == coin {
			pos = &positions[i]
			break
		}
	}
	if pos == nil {
		return &OrderResponse{
			Success: false,
			Message: fmt.Sprintf("no open position for %s", coin),
			Status:  "error",
		}, fmt.Errorf("no open position for %s", coin)
	}

	closeSize := size
	if closeSize == 0 {
		closeSize = parseFloatSafe(pos.Size)
	}

	// To close: opposite direction
	isBuy := pos.Side == "short"

	price := 100000.0 // High price for market buy
	if !isBuy {
		price = 0.01 // Low price for market sell
	}

	orderReq := hyperliquid.CreateOrderRequest{
		Coin:  coin,
		IsBuy: isBuy,
		Size:  closeSize,
		Price: price,
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifIoc,
			},
		},
	}

	resp, err := a.exchange.Order(a.ctx, orderReq, nil)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "error",
		}, err
	}

	return parseOrderResponse(resp), nil
}

func parseOrderResponse(resp hyperliquid.OrderStatus) *OrderResponse {
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
	return out
}

func (a *Account) GetOpenOrders() ([]hyperliquid.OpenOrder, error) {
	return a.info.OpenOrders(a.ctx, a.address)
}

func (a *Account) CancelOrder(coin string, orderId int64) error {
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

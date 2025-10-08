package main
import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"github.com/ethereum/go-ethereum/crypto"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)
type Account struct {
	ctx         context.Context
	info        *hyperliquid.Info
	exchange    *hyperliquid.Exchange
	address     string
	privateKey  *ecdsa.PrivateKey
	isConnected bool
}
type AccountBalance struct {
	AccountValue    string  `json:"accountValue"`
	TotalRawUsd     string  `json:"totalRawUsd"`
	Withdrawable    string  `json:"withdrawable"`
	TotalMargin     string  `json:"totalMargin"`
	AccountLeverage float64 `json:"accountLeverage"`
}
type ActivePosition struct {
	Coin           string `json:"coin"`
	Side           string `json:"side"` 
	Size           string `json:"size"`
	EntryPrice     string `json:"entryPrice"`
	CurrentPrice   string `json:"currentPrice"`
	LiquidationPx  string `json:"liquidationPx"`
	UnrealizedPnl  string `json:"unrealizedPnl"`
	PositionValue  string `json:"positionValue"`
	Leverage       int    `json:"leverage"`
	MarginUsed     string `json:"marginUsed"`
	ReturnOnEquity string `json:"returnOnEquity"`
}
type PortfolioSummary struct {
	Balance        AccountBalance   `json:"balance"`
	Positions      []ActivePosition `json:"positions"`
	TotalPositions int              `json:"totalPositions"`
	TotalPnL       float64          `json:"totalPnL"`
}
type OrderRequest struct {
	Coin       string  `json:"coin"`
	IsBuy      bool    `json:"isBuy"`
	Size       float64 `json:"size"`
	Price      float64 `json:"price"`
	OrderType  string  `json:"orderType"` 
	ReduceOnly bool    `json:"reduceOnly"`
}
type OrderResponse struct {
	Success bool   `json:"success"`
	OrderID string `json:"orderId"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
type ClosePositionRequest struct {
	Coin string  `json:"coin"`
	Size float64 `json:"size"` 
}
func NewAccount(ctx context.Context) *Account {
	info := hyperliquid.NewInfo(ctx, hyperliquid.MainnetAPIURL, true, nil, nil)
	return &Account{
		ctx:         ctx,
		info:        info,
		isConnected: false,
	}
}
func (a *Account) ConnectWallet(privateKeyHex string) error {
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	exchange := hyperliquid.NewExchange(
		a.ctx,
		privateKey,
		hyperliquid.MainnetAPIURL,
		nil,
		"", 
		"", 
		nil,
	)
	a.privateKey = privateKey
	a.address = address
	a.exchange = exchange
	a.isConnected = true
	return nil
}
func (a *Account) GetAddress() string {
	return a.address
}
func (a *Account) IsConnected() bool {
	return a.isConnected
}
func (a *Account) GetPortfolioSummary() (*PortfolioSummary, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}
	userState, err := a.info.UserState(a.ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
	}
	balance := AccountBalance{
		AccountValue: userState.MarginSummary.AccountValue,
		TotalRawUsd:  userState.MarginSummary.TotalRawUsd,
		Withdrawable: userState.Withdrawable,
		TotalMargin:  userState.MarginSummary.TotalMarginUsed,
	}
	positions := []ActivePosition{}
	totalPnL := 0.0
	for _, assetPos := range userState.AssetPositions {
		pos := assetPos.Position
		size := parseFloat(pos.Szi)
		var side string
		if size > 0 {
			side = "long"
		} else if size < 0 {
			side = "short"
			size = -size 
		} else {
			continue 
		}
		entryPx := ""
		if pos.EntryPx != nil {
			entryPx = *pos.EntryPx
		}
		liquidationPx := ""
		if pos.LiquidationPx != nil {
			liquidationPx = *pos.LiquidationPx
		}
		activePos := ActivePosition{
			Coin:           pos.Coin,
			Side:           side,
			Size:           fmt.Sprintf("%.4f", size),
			EntryPrice:     entryPx,
			LiquidationPx:  liquidationPx,
			UnrealizedPnl:  pos.UnrealizedPnl,
			PositionValue:  pos.PositionValue,
			Leverage:       pos.Leverage.Value,
			MarginUsed:     pos.MarginUsed,
			ReturnOnEquity: pos.ReturnOnEquity,
		}
		positions = append(positions, activePos)
		totalPnL += parseFloat(pos.UnrealizedPnl)
	}
	return &PortfolioSummary{
		Balance:        balance,
		Positions:      positions,
		TotalPositions: len(positions),
		TotalPnL:       totalPnL,
	}, nil
}
func (a *Account) GetActivePositions() ([]ActivePosition, error) {
	summary, err := a.GetPortfolioSummary()
	if err != nil {
		return nil, err
	}
	return summary.Positions, nil
}
func (a *Account) OpenTrade(req OrderRequest) (*OrderResponse, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}
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
				Tpsl:      hyperliquid.Tpsl("tp"),
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
			Status:  "failed",
		}, err
	}
	orderResp := &OrderResponse{
		Success: true,
		Message: "Order placed successfully",
	}
	if resp.Resting != nil {
		orderResp.Status = resp.Resting.Status
		orderResp.OrderID = fmt.Sprintf("%d", resp.Resting.Oid)
	} else if resp.Filled != nil {
		orderResp.Status = "filled"
		orderResp.Message = fmt.Sprintf("Order filled - Avg Price: %s, Size: %s", resp.Filled.AvgPx, resp.Filled.TotalSz)
	} else if resp.Error != nil {
		orderResp.Success = false
		orderResp.Status = "error"
		orderResp.Message = *resp.Error
	}
	return orderResp, nil
}
func (a *Account) ClosePosition(req ClosePositionRequest) (*OrderResponse, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}
	positions, err := a.GetActivePositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}
	var position *ActivePosition
	for _, pos := range positions {
		if pos.Coin == req.Coin {
			position = &pos
			break
		}
	}
	if position == nil {
		return nil, fmt.Errorf("no open position found for %s", req.Coin)
	}
	size := req.Size
	if size == 0 {
		size = parseFloat(position.Size)
	}
	isBuy := position.Side == "short"
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
func (a *Account) GetOpenOrders() ([]hyperliquid.OpenOrder, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}
	orders, err := a.info.OpenOrders(a.ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}
	return orders, nil
}
func (a *Account) CancelOrder(coin string, orderId int64) error {
	if !a.isConnected {
		return fmt.Errorf("wallet not connected")
	}
	if a.exchange == nil {
		return fmt.Errorf("exchange not initialized")
	}
	_, err := a.exchange.Cancel(a.ctx, coin, orderId)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	return nil
}
func (a *Account) CancelAllOrders(coin string) error {
	if !a.isConnected {
		return fmt.Errorf("wallet not connected")
	}
	orders, err := a.GetOpenOrders()
	if err != nil {
		return fmt.Errorf("failed to get open orders: %w", err)
	}
	for _, order := range orders {
		if order.Coin == coin {
			if err := a.CancelOrder(coin, order.Oid); err != nil {
				return fmt.Errorf("failed to cancel order %d: %w", order.Oid, err)
			}
		}
	}
	return nil
}
func (a *Account) LoadPrivateKeyFromEnv() error {
	privateKeyHex := os.Getenv("HYPERLIQUID_PRIVATE_KEY")
	if privateKeyHex == "" {
		return fmt.Errorf("HYPERLIQUID_PRIVATE_KEY environment variable not set")
	}
	return a.ConnectWallet(privateKeyHex)
}

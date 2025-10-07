package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)

// Account handles user account information and trading operations
type Account struct {
	ctx         context.Context
	info        *hyperliquid.Info
	exchange    *hyperliquid.Exchange
	address     string
	privateKey  *ecdsa.PrivateKey
	isConnected bool
}

// AccountBalance represents user's account balance information
type AccountBalance struct {
	AccountValue    string  `json:"accountValue"`
	TotalRawUsd     string  `json:"totalRawUsd"`
	Withdrawable    string  `json:"withdrawable"`
	TotalMargin     string  `json:"totalMargin"`
	AccountLeverage float64 `json:"accountLeverage"`
}

// ActivePosition represents a currently open position
type ActivePosition struct {
	Coin           string `json:"coin"`
	Side           string `json:"side"` // "long" or "short"
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

// PortfolioSummary contains overall portfolio information
type PortfolioSummary struct {
	Balance        AccountBalance   `json:"balance"`
	Positions      []ActivePosition `json:"positions"`
	TotalPositions int              `json:"totalPositions"`
	TotalPnL       float64          `json:"totalPnL"`
}

// OrderRequest represents a request to open a trade
type OrderRequest struct {
	Coin       string  `json:"coin"`
	IsBuy      bool    `json:"isBuy"`
	Size       float64 `json:"size"`
	Price      float64 `json:"price"`
	OrderType  string  `json:"orderType"` // "limit" or "market"
	ReduceOnly bool    `json:"reduceOnly"`
}

// OrderResponse represents the response from placing an order
type OrderResponse struct {
	Success bool   `json:"success"`
	OrderID string `json:"orderId"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ClosePositionRequest represents a request to close a position
type ClosePositionRequest struct {
	Coin string  `json:"coin"`
	Size float64 `json:"size"` // 0 means close entire position
}

// NewAccount creates a new account instance
func NewAccount(ctx context.Context) *Account {
	info := hyperliquid.NewInfo(ctx, hyperliquid.MainnetAPIURL, true, nil, nil)
	return &Account{
		ctx:         ctx,
		info:        info,
		isConnected: false,
	}
}

// ConnectWallet connects to user's wallet using private key
func (a *Account) ConnectWallet(privateKeyHex string) error {
	// Remove '0x' prefix if present
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

	// Initialize exchange
	exchange := hyperliquid.NewExchange(
		a.ctx,
		privateKey,
		hyperliquid.MainnetAPIURL,
		nil,
		"", // vault address (empty for direct trading)
		"", // account address (empty to use private key address)
		nil,
	)

	a.privateKey = privateKey
	a.address = address
	a.exchange = exchange
	a.isConnected = true

	return nil
}

// GetAddress returns the connected wallet address
func (a *Account) GetAddress() string {
	return a.address
}

// IsConnected returns whether wallet is connected
func (a *Account) IsConnected() bool {
	return a.isConnected
}

// GetPortfolioSummary fetches complete portfolio information
func (a *Account) GetPortfolioSummary() (*PortfolioSummary, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}

	userState, err := a.info.UserState(a.ctx, a.address)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
	}

	// Parse account balance
	balance := AccountBalance{
		AccountValue: userState.MarginSummary.AccountValue,
		TotalRawUsd:  userState.MarginSummary.TotalRawUsd,
		Withdrawable: userState.Withdrawable,
		TotalMargin:  userState.MarginSummary.TotalMarginUsed,
	}

	// Parse positions
	positions := []ActivePosition{}
	totalPnL := 0.0

	for _, assetPos := range userState.AssetPositions {
		pos := assetPos.Position

		// Determine side based on position size
		size := parseFloat(pos.Szi)
		var side string
		if size > 0 {
			side = "long"
		} else if size < 0 {
			side = "short"
			size = -size // Make size positive for display
		} else {
			continue // Skip zero positions
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

// GetActivePositions returns only the active positions
func (a *Account) GetActivePositions() ([]ActivePosition, error) {
	summary, err := a.GetPortfolioSummary()
	if err != nil {
		return nil, err
	}
	return summary.Positions, nil
}

// OpenTrade places a new order to open a position
func (a *Account) OpenTrade(req OrderRequest) (*OrderResponse, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}

	if a.exchange == nil {
		return nil, fmt.Errorf("exchange not initialized")
	}

	// Prepare order request
	orderReq := hyperliquid.CreateOrderRequest{
		Coin:       req.Coin,
		IsBuy:      req.IsBuy,
		Size:       req.Size,
		Price:      req.Price,
		ReduceOnly: req.ReduceOnly,
	}

	// Set order type
	if req.OrderType == "market" {
		orderReq.OrderType = hyperliquid.OrderType{
			Trigger: &hyperliquid.TriggerOrderType{
				TriggerPx: req.Price,
				IsMarket:  true,
				Tpsl:      hyperliquid.Tpsl("tp"),
			},
		}
	} else {
		// Default to limit order
		orderReq.OrderType = hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifGtc, // Good till cancelled
			},
		}
	}

	// Place order (builder is nil for standard orders)
	resp, err := a.exchange.Order(a.ctx, orderReq, nil)
	if err != nil {
		return &OrderResponse{
			Success: false,
			Message: err.Error(),
			Status:  "failed",
		}, err
	}

	// Parse response
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

// ClosePosition closes an open position (full or partial)
func (a *Account) ClosePosition(req ClosePositionRequest) (*OrderResponse, error) {
	if !a.isConnected {
		return nil, fmt.Errorf("wallet not connected")
	}

	// Get current position to determine side
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

	// Determine size to close
	size := req.Size
	if size == 0 {
		// Close entire position
		size = parseFloat(position.Size)
	}

	// To close a long position, we sell. To close short, we buy.
	isBuy := position.Side == "short"

	// Use market order to close position immediately
	closeReq := OrderRequest{
		Coin:       req.Coin,
		IsBuy:      isBuy,
		Size:       size,
		Price:      0, // Market order
		OrderType:  "market",
		ReduceOnly: true, // Important: only reduce position, don't flip
	}

	return a.OpenTrade(closeReq)
}

// OpenLongPosition is a convenience method to open a long position
func (a *Account) OpenLongPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.OpenTrade(OrderRequest{
		Coin:      coin,
		IsBuy:     true,
		Size:      size,
		Price:     price,
		OrderType: orderType,
	})
}

// OpenShortPosition is a convenience method to open a short position
func (a *Account) OpenShortPosition(coin string, size float64, price float64, orderType string) (*OrderResponse, error) {
	return a.OpenTrade(OrderRequest{
		Coin:      coin,
		IsBuy:     false,
		Size:      size,
		Price:     price,
		OrderType: orderType,
	})
}

// GetOpenOrders returns all open orders
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

// CancelOrder cancels a specific order
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

// CancelAllOrders cancels all open orders for a coin
func (a *Account) CancelAllOrders(coin string) error {
	if !a.isConnected {
		return fmt.Errorf("wallet not connected")
	}

	// Get all open orders
	orders, err := a.GetOpenOrders()
	if err != nil {
		return fmt.Errorf("failed to get open orders: %w", err)
	}

	// Cancel each order for the specified coin
	for _, order := range orders {
		if order.Coin == coin {
			if err := a.CancelOrder(coin, order.Oid); err != nil {
				return fmt.Errorf("failed to cancel order %d: %w", order.Oid, err)
			}
		}
	}

	return nil
}

// LoadPrivateKeyFromEnv loads private key from environment variable
func (a *Account) LoadPrivateKeyFromEnv() error {
	privateKeyHex := os.Getenv("HYPERLIQUID_PRIVATE_KEY")
	if privateKeyHex == "" {
		return fmt.Errorf("HYPERLIQUID_PRIVATE_KEY environment variable not set")
	}
	return a.ConnectWallet(privateKeyHex)
}

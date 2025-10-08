package main
import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)
type CandleCache struct {
	client    *redis.Client
	connected bool
	ctx       context.Context
}
func NewCandleCache(client *redis.Client, ctx context.Context) *CandleCache {
	cache := &CandleCache{
		client:    client,
		ctx:       ctx,
		connected: false,
	}
	cache.checkConnection()
	return cache
}
func (c *CandleCache) checkConnection() {
	if c.client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()
	if err := c.client.Ping(ctx).Err(); err == nil {
		c.connected = true
	}
}
func (c *CandleCache) IsConnected() bool {
	return c.connected
}
func (c *CandleCache) buildKey(symbol, interval string, limit int) string {
	return fmt.Sprintf("candles:%s:%s:%d", symbol, interval, limit)
}
func (c *CandleCache) Get(symbol, interval string, limit int) ([]hyperliquid.Candle, bool) {
	if !c.connected {
		return nil, false
	}
	key := c.buildKey(symbol, interval, limit)
	ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
	defer cancel()
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	var candles []hyperliquid.Candle
	if err := json.Unmarshal(data, &candles); err != nil {
		return nil, false
	}
	return candles, true
}
func (c *CandleCache) Set(symbol, interval string, limit int, candles []hyperliquid.Candle) error {
	if !c.connected || len(candles) == 0 {
		return nil
	}
	key := c.buildKey(symbol, interval, limit)
	ttl := c.getTTL(interval)
	data, err := json.Marshal(candles)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
	defer cancel()
	return c.client.Set(ctx, key, data, ttl).Err()
}
func (c *CandleCache) getTTL(interval string) time.Duration {
	switch interval {
	case "1m":
		return 1 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 5 * time.Minute
	}
}
func (c *CandleCache) UpdateWithNewCandle(symbol, interval string, limit int, newCandles []hyperliquid.Candle) error {
	if !c.connected || len(newCandles) == 0 {
		return nil
	}
	cachedCandles, found := c.Get(symbol, interval, limit)
	if !found {
		return c.Set(symbol, interval, limit, newCandles)
	}
	lastCachedTime := cachedCandles[len(cachedCandles)-1].Timestamp
	var updatedCandles []hyperliquid.Candle
	for _, candle := range newCandles {
		if candle.Timestamp > lastCachedTime {
			updatedCandles = append(updatedCandles, candle)
		}
	}
	if len(updatedCandles) == 0 {
		key := c.buildKey(symbol, interval, limit)
		ttl := c.getTTL(interval)
		ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
		defer cancel()
		return c.client.Expire(ctx, key, ttl).Err()
	}
	merged := append(cachedCandles, updatedCandles...)
	if len(merged) > limit {
		merged = merged[len(merged)-limit:]
	}
	return c.Set(symbol, interval, limit, merged)
}
func (c *CandleCache) GetOrFetch(
	symbol, interval string,
	limit int,
	fetchFn func() ([]hyperliquid.Candle, error),
) ([]hyperliquid.Candle, error) {
	if c.connected {
		if candles, found := c.Get(symbol, interval, limit); found {
			return candles, nil
		}
	}
	candles, err := fetchFn()
	if err != nil {
		return nil, err
	}
	if c.connected {
		go c.Set(symbol, interval, limit, candles)
	}
	return candles, nil
}
func (c *CandleCache) InvalidatePattern(pattern string) error {
	if !c.connected {
		return nil
	}
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}
func (c *CandleCache) InvalidateSymbol(symbol string) error {
	return c.InvalidatePattern(fmt.Sprintf("candles:%s:*", symbol))
}
func (c *CandleCache) Clear() error {
	return c.InvalidatePattern("candles:*")
}

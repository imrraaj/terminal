package main
import (
	"context"
	"testing"
	"time"
	"github.com/redis/go-redis/v9"
	hyperliquid "github.com/sonirico/go-hyperliquid"
)
func TestCacheWithRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()
	ctx := context.Background()
	cache := NewCandleCache(client, ctx)
	if !cache.IsConnected() {
		t.Skip("Redis not connected, skipping test")
	}
	symbol := "BTC"
	interval := "5m"
	limit := 100
	testCandles := []hyperliquid.Candle{
		{Timestamp: 1000, Open: "100", High: "110", Low: "90", Close: "105", Volume: "1000"},
		{Timestamp: 2000, Open: "105", High: "115", Low: "95", Close: "110", Volume: "1100"},
	}
	cache.Clear()
	err := cache.Set(symbol, interval, limit, testCandles)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	retrieved, found := cache.Get(symbol, interval, limit)
	if !found {
		t.Fatal("Cache miss when should be hit")
	}
	if len(retrieved) != len(testCandles) {
		t.Fatalf("Expected %d candles, got %d", len(testCandles), len(retrieved))
	}
	time.Sleep(100 * time.Millisecond)
	retrieved, found = cache.Get(symbol, interval, limit)
	if !found {
		t.Fatal("Cache expired too quickly")
	}
	err = cache.InvalidateSymbol(symbol)
	if err != nil {
		t.Fatalf("Failed to invalidate: %v", err)
	}
	_, found = cache.Get(symbol, interval, limit)
	if found {
		t.Fatal("Cache should be invalidated")
	}
	t.Log("Cache test passed")
}
func TestCacheWithoutRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", 
	})
	defer client.Close()
	ctx := context.Background()
	cache := NewCandleCache(client, ctx)
	if cache.IsConnected() {
		t.Skip("Redis should not be connected on port 9999")
	}
	symbol := "BTC"
	interval := "5m"
	limit := 100
	testCandles := []hyperliquid.Candle{
		{Timestamp: 1000, Open: "100", High: "110", Low: "90", Close: "105", Volume: "1000"},
	}
	err := cache.Set(symbol, interval, limit, testCandles)
	if err != nil {
		t.Fatalf("Set should not error without Redis: %v", err)
	}
	_, found := cache.Get(symbol, interval, limit)
	if found {
		t.Fatal("Should not find cached data without Redis")
	}
	t.Log("No-Redis test passed")
}
func TestUpdateWithNewCandle(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()
	ctx := context.Background()
	cache := NewCandleCache(client, ctx)
	if !cache.IsConnected() {
		t.Skip("Redis not connected, skipping test")
	}
	symbol := "ETH"
	interval := "15m"
	limit := 100
	initialCandles := []hyperliquid.Candle{
		{Timestamp: 1000, Open: "100", High: "110", Low: "90", Close: "105", Volume: "1000"},
		{Timestamp: 2000, Open: "105", High: "115", Low: "95", Close: "110", Volume: "1100"},
	}
	cache.Set(symbol, interval, limit, initialCandles)
	newCandles := []hyperliquid.Candle{
		{Timestamp: 2000, Open: "105", High: "115", Low: "95", Close: "110", Volume: "1100"},  
		{Timestamp: 3000, Open: "110", High: "120", Low: "100", Close: "115", Volume: "1200"}, 
	}
	err := cache.UpdateWithNewCandle(symbol, interval, limit, newCandles)
	if err != nil {
		t.Fatalf("Failed to update cache: %v", err)
	}
	retrieved, found := cache.Get(symbol, interval, limit)
	if !found {
		t.Fatal("Cache should be found after update")
	}
	if len(retrieved) != 3 {
		t.Fatalf("Expected 3 candles after update, got %d", len(retrieved))
	}
	if retrieved[len(retrieved)-1].Timestamp != 3000 {
		t.Fatal("Last candle should be the new one with timestamp 3000")
	}
	t.Log("Update with new candle test passed")
}

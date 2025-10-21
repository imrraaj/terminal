package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	if err := database.init(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS strategies (
		id TEXT PRIMARY KEY,
		symbol TEXT NOT NULL,
		interval TEXT NOT NULL,
		factor REAL NOT NULL,
		position_size REAL NOT NULL,
		trade_direction TEXT NOT NULL,
		take_profit_percent REAL NOT NULL,
		stop_loss_percent REAL NOT NULL,
		is_running INTEGER NOT NULL,
		last_candle_time INTEGER NOT NULL,
		position_data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) SaveStrategy(strategy *MaxTrendPointsStrategy) error {
	var positionData []byte
	var err error
	if strategy.Position != nil {
		positionData, err = json.Marshal(strategy.Position)
		if err != nil {
			return fmt.Errorf("failed to marshal position: %w", err)
		}
	}

	query := `
	INSERT OR REPLACE INTO strategies (
		id, symbol, interval, factor, position_size, trade_direction,
		take_profit_percent, stop_loss_percent, is_running, last_candle_time, position_data, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err = d.db.Exec(query,
		strategy.ID,
		strategy.Symbol,
		strategy.Interval,
		strategy.Factor,
		strategy.Config.PositionSize,
		strategy.Config.TradeDirection,
		strategy.Config.TakeProfitPercent,
		strategy.Config.StopLossPercent,
		boolToInt(strategy.IsRunning),
		strategy.LastCandleTime,
		positionData,
	)

	return err
}

func (d *Database) LoadStrategies() ([]*MaxTrendPointsStrategy, error) {
	query := `
	SELECT id, symbol, interval, factor, position_size, trade_direction,
		   take_profit_percent, stop_loss_percent, is_running, last_candle_time, position_data
	FROM strategies
	WHERE is_running = 1
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []*MaxTrendPointsStrategy

	for rows.Next() {
		var (
			id               string
			symbol           string
			interval         string
			factor           float64
			positionSize     float64
			tradeDirection   string
			takeProfitPct    float64
			stopLossPct      float64
			isRunning        int
			lastCandleTime   int64
			positionDataJSON []byte
		)

		err := rows.Scan(
			&id, &symbol, &interval, &factor, &positionSize, &tradeDirection,
			&takeProfitPct, &stopLossPct, &isRunning, &lastCandleTime, &positionDataJSON,
		)
		if err != nil {
			return nil, err
		}

		strategy := &MaxTrendPointsStrategy{
			ID:             id,
			Symbol:         symbol,
			Interval:       interval,
			Factor:         factor,
			LastCandleTime: lastCandleTime,
			IsRunning:      isRunning == 1,
			Config: StrategyConfig{
				PositionSize:      positionSize,
				TradeDirection:    tradeDirection,
				TakeProfitPercent: takeProfitPct,
				StopLossPercent:   stopLossPct,
				Parameters: map[string]any{
					"factor":             factor,
					"positionSize":       positionSize,
					"tradeDirection":     tradeDirection,
					"takeProfitPercent":  takeProfitPct,
					"stopLossPercent":    stopLossPct,
				},
			},
		}

		if len(positionDataJSON) > 0 {
			var position Position
			if err := json.Unmarshal(positionDataJSON, &position); err == nil {
				strategy.Position = &position
			}
		}

		strategies = append(strategies, strategy)
	}

	return strategies, nil
}

func (d *Database) DeleteStrategy(id string) error {
	_, err := d.db.Exec("DELETE FROM strategies WHERE id = ?", id)
	return err
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

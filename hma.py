import pandas as pd
import numpy as np
import ccxt
import time

class StreamingHMATrendDetector:
    def __init__(self, buffer_size=1000, factor=2.5):
        self.buffer_size = buffer_size
        self.factor = factor
        self.candles = []
        self.last_direction = None

    def hull_moving_average(self, data, period):
        if len(data) < period:
            return np.full(len(data), np.nan)

        half_period = int(period / 2)
        sqrt_period = int(np.sqrt(period))

        wma_half = self.weighted_moving_average(data, half_period)
        wma_full = self.weighted_moving_average(data, period)
        raw_hma = 2 * wma_half - wma_full
        hma = self.weighted_moving_average(raw_hma, sqrt_period)

        return hma

    def weighted_moving_average(self, data, period):
        if len(data) < period:
            return np.full(len(data), np.nan)

        weights = np.arange(1, period + 1)
        wma = np.full(len(data), np.nan)

        for i in range(period - 1, len(data)):
            wma[i] = np.average(data[i - period + 1:i + 1], weights=weights)

        return wma

    def calculate_trend(self):
        if len(self.candles) < 201:
            return None

        df = pd.DataFrame(self.candles, columns=['open', 'high', 'low', 'close'])
        df['hl2'] = (df['high'] + df['low']) / 2
        df['range'] = df['high'] - df['low']
        df['hma_range'] = self.hull_moving_average(df['range'].values, 200)

        n = len(df)
        upper_band = np.full(n, np.nan)
        lower_band = np.full(n, np.nan)
        trend_line = np.full(n, np.nan)
        direction = np.full(n, np.nan)

        for i in range(200, n):
            if np.isnan(df['hma_range'].iloc[i]):
                continue

            src = df['hl2'].iloc[i]
            dist = df['hma_range'].iloc[i]

            current_upper = src + self.factor * dist
            current_lower = src - self.factor * dist

            prev_lower = lower_band[i-1] if not np.isnan(lower_band[i-1]) else current_lower
            prev_upper = upper_band[i-1] if not np.isnan(upper_band[i-1]) else current_upper
            prev_close = df['close'].iloc[i-1]

            if current_lower > prev_lower or prev_close < prev_lower:
                lower_band[i] = current_lower
            else:
                lower_band[i] = prev_lower

            if current_upper < prev_upper or prev_close > prev_upper:
                upper_band[i] = current_upper
            else:
                upper_band[i] = prev_upper

            if i == 200 or np.isnan(direction[i-1]):
                direction[i] = 1 if df['close'].iloc[i] <= upper_band[i] else -1
            else:
                prev_trend = trend_line[i-1]
                prev_upper = upper_band[i-1]
                if np.isclose(prev_trend, prev_upper, atol=1e-10):
                    direction[i] = -1 if df['close'].iloc[i] > upper_band[i] else 1
                else:
                    direction[i] = 1 if df['close'].iloc[i] < lower_band[i] else -1

            trend_line[i] = lower_band[i] if direction[i] == -1 else upper_band[i]

        return direction[-1] if not np.isnan(direction[-1]) else None

    def train(self, csv_file):
        df = pd.read_csv(csv_file)

        # Convert to list of OHLC tuples
        for _, row in df.iterrows():
            self.candles.append([row['open'], row['high'], row['low'], row['close']])

        # Keep only recent candles within buffer size
        if len(self.candles) > self.buffer_size:
            self.candles = self.candles[-self.buffer_size:]

        # Calculate initial trend
        initial_trend = self.calculate_trend()
        self.last_direction = initial_trend

        # Count trend changes in historical data
        df['hl2'] = (df['high'] + df['low']) / 2
        df['range'] = df['high'] - df['low']
        df['hma_range'] = self.hull_moving_average(df['range'].values, 200)

        n = len(df)
        direction = np.full(n, np.nan)
        upper_band = np.full(n, np.nan)
        lower_band = np.full(n, np.nan)
        trend_line = np.full(n, np.nan)

        for i in range(200, n):
            if np.isnan(df['hma_range'].iloc[i]):
                continue

            src = df['hl2'].iloc[i]
            dist = df['hma_range'].iloc[i]

            current_upper = src + self.factor * dist
            current_lower = src - self.factor * dist

            prev_lower = lower_band[i-1] if not np.isnan(lower_band[i-1]) else current_lower
            prev_upper = upper_band[i-1] if not np.isnan(upper_band[i-1]) else current_upper
            prev_close = df['close'].iloc[i-1]

            if current_lower > prev_lower or prev_close < prev_lower:
                lower_band[i] = current_lower
            else:
                lower_band[i] = prev_lower

            if current_upper < prev_upper or prev_close > prev_upper:
                upper_band[i] = current_upper
            else:
                upper_band[i] = prev_upper

            if i == 200 or np.isnan(direction[i-1]):
                direction[i] = 1 if df['close'].iloc[i] <= upper_band[i] else -1
            else:
                prev_trend = trend_line[i-1]
                prev_upper_val = upper_band[i-1]
                if np.isclose(prev_trend, prev_upper_val, atol=1e-10):
                    direction[i] = -1 if df['close'].iloc[i] > upper_band[i] else 1
                else:
                    direction[i] = 1 if df['close'].iloc[i] < lower_band[i] else -1

            trend_line[i] = lower_band[i] if direction[i] == -1 else upper_band[i]

        # Count trend changes
        trend_changes = 0
        for i in range(201, n):
            if not np.isnan(direction[i]) and not np.isnan(direction[i-1]):
                if direction[i] != direction[i-1]:
                    trend_changes += 1

        print(f"Training completed. Found {trend_changes} trend changes in historical data.")

    def update_candle(self, open_price, high_price, low_price, close_price):
        # Add new candle
        self.candles.append([open_price, high_price, low_price, close_price])

        # Keep buffer size
        if len(self.candles) > self.buffer_size:
            self.candles = self.candles[-self.buffer_size:]

        # Calculate current trend
        current_direction = self.calculate_trend()

        if current_direction is None:
            return

        # Check for trend change
        if self.last_direction is not None and self.last_direction != current_direction:
            trend_name = "DOWNTREND" if current_direction == 1 else "UPTREND"
            print(f"Trend changed at price ${close_price:.2f} - {trend_name}")
        else:
            trend_name = "DOWNTREND" if current_direction == 1 else "UPTREND"
            print(f"Trend continuing at price ${close_price:.2f} - {trend_name}")

        self.last_direction = current_direction

def main():
    detector = StreamingHMATrendDetector(buffer_size=1000, factor=4)
    detector.train("btc_15m.csv")

    print("\n" + "="*60)
    print("=== STARTING LIVE STREAMING ===")
    exchange = ccxt.binance()
    symbol = 'BTC/USDT'
    timeframe = '15m'
    print(f"Fetching live {timeframe} candles for {symbol}")
    print("Waiting for new candles...")
    last_candle_time = None
    try:
        while True:
            ohlcv = exchange.fetch_ohlcv(symbol, timeframe, limit=2)
            if len(ohlcv) >= 1:
                # Get the latest completed candle
                latest = ohlcv[-2] if len(ohlcv) > 1 else ohlcv[-1]
                timestamp = latest[0]
                if last_candle_time != timestamp:
                    last_candle_time = timestamp
                    open_price = latest[1]
                    high_price = latest[2]
                    low_price = latest[3]
                    close_price = latest[4]
                    print(f"\nNew candle: {pd.to_datetime(timestamp, unit='ms')}")
                    print(f"OHLC: ${open_price:.2f} ${high_price:.2f} ${low_price:.2f} ${close_price:.2f}")
                    detector.update_candle(open_price, high_price, low_price, close_price)
            time.sleep(60)
    except KeyboardInterrupt:
        print("\nStopped by user")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()












# import pandas as pd
# import numpy as np
# from collections import deque
# import ccxt
# import time

# class StreamingHMATrendDetector:
#     def __init__(self, buffer_size=1000, factor=2.5):
#         self.buffer_size = buffer_size
#         self.factor = factor

#         # Circular buffers for OHLC data
#         self.high = deque(maxlen=buffer_size)
#         self.low = deque(maxlen=buffer_size)
#         self.close = deque(maxlen=buffer_size)
#         self.open = deque(maxlen=buffer_size)

#         # Calculated values buffers
#         self.hl2 = deque(maxlen=buffer_size)
#         self.range_vals = deque(maxlen=buffer_size)
#         self.hma_range = deque(maxlen=buffer_size)
#         self.upper_band = deque(maxlen=buffer_size)
#         self.lower_band = deque(maxlen=buffer_size)
#         self.trend_line = deque(maxlen=buffer_size)
#         self.direction = deque(maxlen=buffer_size)

#         self.current_trend = None
#         self.candle_count = 0

#     def hull_moving_average(self, data, period):
#         """Calculate Hull Moving Average"""
#         if len(data) < period:
#             return np.full(len(data), np.nan)

#         half_period = int(period / 2)
#         sqrt_period = int(np.sqrt(period))
#         wma_half = self.weighted_moving_average(data, half_period)
#         wma_full = self.weighted_moving_average(data, period)
#         raw_hma = 2 * wma_half - wma_full
#         hma = self.weighted_moving_average(raw_hma, sqrt_period)
#         return hma

#     def weighted_moving_average(self, data, period):
#         """Calculate Weighted Moving Average"""
#         if len(data) < period:
#             return np.full(len(data), np.nan)
#         weights = np.arange(1, period + 1)
#         wma = np.full(len(data), np.nan)
#         for i in range(period - 1, len(data)):
#             wma[i] = np.average(data[i - period + 1:i + 1], weights=weights)
#         return wma

#     def train(self, csv_file):
#         """Train with historical data and print trend changes"""
#         print(f"=== TRAINING WITH {csv_file} ===")

#         try:
#             df = pd.read_csv(csv_file)
#             datetime_cols = ['datetime', 'timestamp', 'date', 'time']
#             for col in datetime_cols:
#                 if col in df.columns:
#                     try:
#                         df[col] = pd.to_datetime(df[col])
#                         df.set_index(col, inplace=True)
#                         break
#                     except:
#                         continue

#             print(f"Loaded {len(df)} candles for training")

#             # Process all historical data and track trend changes
#             trend_changes = 0
#             for i, (timestamp, row) in enumerate(df.iterrows()):
#                 prev_trend = self.get_current_trend()
#                 self._add_candle(row['open'], row['high'], row['low'], row['close'])
#                 new_trend = self.get_current_trend()

#                 # Check for trend change after sufficient data
#                 if i > 200 and prev_trend != new_trend and prev_trend is not None and new_trend is not None:
#                     trend_changes += 1
#                     print(f"{timestamp} | {new_trend} | ${row['close']:.2f}")

#             print(f"\nTraining complete. Total historical trend changes: {trend_changes}")
#             print(f"Current trend: {self.get_current_trend()}")

#         except Exception as e:
#             print(f"Training error: {e}")

#     def _add_candle(self, open_price, high_price, low_price, close_price):
#         """Internal method to add a candle and update all indicators"""
#         self.open.append(open_price)
#         self.high.append(high_price)
#         self.low.append(low_price)
#         self.close.append(close_price)
#         self.candle_count += 1

#         # Calculate HL2 and range
#         hl2_val = (high_price + low_price) / 2
#         range_val = high_price - low_price
#         self.hl2.append(hl2_val)
#         self.range_vals.append(range_val)

#         # Calculate HMA of range
#         hma_val = self._calculate_hma_incremental()
#         self.hma_range.append(hma_val)

#         # Calculate bands and direction if we have enough data
#         if len(self.range_vals) >= 200 and not np.isnan(hma_val):
#             self._update_trend_calculation()
#         else:
#             self.upper_band.append(np.nan)
#             self.lower_band.append(np.nan)
#             self.trend_line.append(np.nan)
#             self.direction.append(np.nan)

#     def _calculate_hma_incremental(self):
#         """Calculate HMA incrementally for the latest range value"""
#         if len(self.range_vals) < 200:
#             return np.nan

#         # Get last 200 range values
#         range_data = list(self.range_vals)[-200:]
#         return self.hull_moving_average(np.array(range_data), 200)[-1]

#     def _update_trend_calculation(self):
#         """Update trend calculation for the latest candle"""
#         src = self.hl2[-1]
#         dist = self.hma_range[-1]
#         close_price = self.close[-1]

#         # Calculate current bands
#         current_upper = src + self.factor * dist
#         current_lower = src - self.factor * dist

#         # Get previous values
#         prev_lower = self.lower_band[-1] if len(self.lower_band) > 0 and not np.isnan(self.lower_band[-1]) else current_lower
#         prev_upper = self.upper_band[-1] if len(self.upper_band) > 0 and not np.isnan(self.upper_band[-1]) else current_upper
#         prev_close = self.close[-2] if len(self.close) > 1 else close_price

#         # Update bands
#         if current_lower > prev_lower or prev_close < prev_lower:
#             lower_band_val = current_lower
#         else:
#             lower_band_val = prev_lower

#         if current_upper < prev_upper or prev_close > prev_upper:
#             upper_band_val = current_upper
#         else:
#             upper_band_val = prev_upper

#         self.lower_band.append(lower_band_val)
#         self.upper_band.append(upper_band_val)

#         # Calculate direction
#         if len(self.direction) == 0 or np.isnan(self.direction[-1]):
#             direction_val = 1 if close_price <= upper_band_val else -1
#         else:
#             prev_trend = self.trend_line[-1]
#             prev_upper = self.upper_band[-2] if len(self.upper_band) > 1 else upper_band_val

#             if np.isclose(prev_trend, prev_upper, atol=1e-10):
#                 direction_val = -1 if close_price > upper_band_val else 1
#             else:
#                 direction_val = 1 if close_price < lower_band_val else -1

#         self.direction.append(direction_val)

#         # Set trend line
#         trend_line_val = lower_band_val if direction_val == -1 else upper_band_val
#         self.trend_line.append(trend_line_val)

#     def update_candle(self, open_price, high_price, low_price, close_price):
#         """Add new candle and check for trend changes"""
#         prev_trend = self.get_current_trend()
#         self._add_candle(open_price, high_price, low_price, close_price)
#         new_trend = self.get_current_trend()

#         # Check for trend change
#         if prev_trend != new_trend and prev_trend is not None and new_trend is not None:
#             print(f"*** TREND CHANGE: {new_trend} at ${close_price:.2f} ***")
#             return True
#         else:
#             print(f"No change. Trend: {new_trend} @ ${close_price:.2f}")
#             return False

#     def get_current_trend(self):
#         """Get current trend direction"""
#         if len(self.direction) == 0 or np.isnan(self.direction[-1]):
#             return None
#         return 'UPTREND' if self.direction[-1] == -1 else 'DOWNTREND'

# def main():
#     detector = StreamingHMATrendDetector(buffer_size=1000, factor=4)
#     detector.train("btc_15m.csv")

#     print("\n" + "="*60)
#     print("=== STARTING LIVE STREAMING ===")
#     exchange = ccxt.binance()
#     symbol = 'BTC/USDT'
#     timeframe = '15m'
#     print(f"Fetching live {timeframe} candles for {symbol}")
#     print("Waiting for new candles...")
#     last_candle_time = None
#     try:
#         while True:
#             ohlcv = exchange.fetch_ohlcv(symbol, timeframe, limit=2)
#             if len(ohlcv) >= 1:
#                 # Get the latest completed candle
#                 latest = ohlcv[-2] if len(ohlcv) > 1 else ohlcv[-1]
#                 timestamp = latest[0]
#                 if last_candle_time != timestamp:
#                     last_candle_time = timestamp
#                     open_price = latest[1]
#                     high_price = latest[2]
#                     low_price = latest[3]
#                     close_price = latest[4]
#                     print(f"\nNew candle: {pd.to_datetime(timestamp, unit='ms')}")
#                     print(f"OHLC: ${open_price:.2f} ${high_price:.2f} ${low_price:.2f} ${close_price:.2f}")
#                     detector.update_candle(open_price, high_price, low_price, close_price)
#             time.sleep(60)
#     except KeyboardInterrupt:
#         print("\nStopped by user")
#     except Exception as e:
#         print(f"Error: {e}")

# if __name__ == "__main__":
#     main()

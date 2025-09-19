import { Time } from "lightweight-charts";

export interface CandleData {
    time: Time;
    open: number;
    high: number;
    low: number;
    close: number;
}

export interface StrategySignal {
    time: Time;
    price: number;
    direction: 'up' | 'down';
    type: 'trend_change' | 'trend_continue';
}

export interface TrendLinePoint {
    time: Time;
    value: number;
    color: string;
}

export class HmaStrategy {
    private bufferSize: number;
    private factor: number;
    private candles: CandleData[] = [];
    private lastDirection: number | null = null;
    private signals: StrategySignal[] = [];
    private trendLineData: TrendLinePoint[] = [];
    
    // Arrays to store calculated values - matching Pine Script exactly
    private upperBands: number[] = [];
    private lowerBands: number[] = [];
    private directions: number[] = [];
    private trendLines: number[] = [];

    constructor(bufferSize: number = 1000, factor: number = 4) {
        this.bufferSize = bufferSize;
        this.factor = factor;
    }

    // Weighted Moving Average - exact Pine Script implementation
    private wma(data: number[], length: number): number[] {
        const result = new Array(data.length).fill(NaN);
        
        for (let i = length - 1; i < data.length; i++) {
            let sum = 0;
            let weightSum = 0;
            
            for (let j = 0; j < length; j++) {
                const weight = j + 1;
                const index = i - length + 1 + j;
                if (index >= 0 && index < data.length && !isNaN(data[index])) {
                    sum += data[index] * weight;
                    weightSum += weight;
                }
            }
            
            if (weightSum > 0) {
                result[i] = sum / weightSum;
            }
        }
        
        return result;
    }

    // Hull Moving Average - exact Pine Script ta.hma implementation
    private hma(data: number[], length: number): number[] {
        if (length <= 0) return new Array(data.length).fill(NaN);
        
        const halfLength = Math.floor(length / 2);
        const sqrtLength = Math.floor(Math.sqrt(length));
        
        const wma1 = this.wma(data, halfLength);
        const wma2 = this.wma(data, length);
        
        // 2 * wma(source, length/2) - wma(source, length)
        const diff = wma1.map((val1, i) => {
            const val2 = wma2[i];
            if (isNaN(val1) || isNaN(val2)) return NaN;
            return 2 * val1 - val2;
        });
        
        return this.wma(diff, sqrtLength);
    }

    // Get previous value with default fallback (nz equivalent)
    private nz(arr: number[], index: number, defaultValue: number = 0): number {
        if (index < 0 || index >= arr.length || isNaN(arr[index])) {
            return defaultValue;
        }
        return arr[index];
    }

    // Main trend calculation - exact Pine Script logic
    private calculateTrendAtIndex(index: number): void {
        if (index < 0 || index >= this.candles.length) return;

        const candle = this.candles[index];
        const src = (candle.high + candle.low) / 2; // hl2
        
        // Calculate HMA of range (high-low) with period 200
        const ranges = this.candles.slice(0, index + 1).map(c => c.high - c.low);
        const hmaRanges = this.hma(ranges, 200);
        const dist = hmaRanges[index];

        if (isNaN(dist)) {
            this.upperBands[index] = NaN;
            this.lowerBands[index] = NaN;
            this.directions[index] = NaN;
            this.trendLines[index] = NaN;
            return;
        }

        let upperBand = src + this.factor * dist;
        let lowerBand = src - this.factor * dist;

        const prevLowerBand = this.nz(this.lowerBands, index - 1, lowerBand);
        const prevUpperBand = this.nz(this.upperBands, index - 1, upperBand);
        const prevClose = index > 0 ? this.candles[index - 1].close : candle.close;

        // Pine Script band logic
        lowerBand = (lowerBand > prevLowerBand || prevClose < prevLowerBand) ? lowerBand : prevLowerBand;
        upperBand = (upperBand < prevUpperBand || prevClose > prevUpperBand) ? upperBand : prevUpperBand;

        this.lowerBands[index] = lowerBand;
        this.upperBands[index] = upperBand;

        // Direction calculation - exact Pine Script logic
        let direction: number;
        const prevTrendLine = this.nz(this.trendLines, index - 1, NaN);
        const prevDist = index > 0 ? this.hma(this.candles.slice(0, index).map(c => c.high - c.low), 200)[index - 1] : NaN;
        
        if (isNaN(prevDist)) {
            // First calculation
            direction = 1;
        } else if (Math.abs(prevTrendLine - prevUpperBand) < 1e-10) {
            // Previous trend was at upper band
            direction = candle.close > upperBand ? -1 : 1;
        } else {
            // Previous trend was at lower band
            direction = candle.close < lowerBand ? 1 : -1;
        }

        this.directions[index] = direction;
        this.trendLines[index] = direction === -1 ? lowerBand : upperBand;
    }

    public updateCandle(candle: CandleData): StrategySignal | null {
        this.candles.push(candle);

        // Extend arrays
        this.upperBands.push(NaN);
        this.lowerBands.push(NaN);
        this.directions.push(NaN);
        this.trendLines.push(NaN);

        if (this.candles.length > this.bufferSize) {
            this.candles = this.candles.slice(-this.bufferSize);
            this.upperBands = this.upperBands.slice(-this.bufferSize);
            this.lowerBands = this.lowerBands.slice(-this.bufferSize);
            this.directions = this.directions.slice(-this.bufferSize);
            this.trendLines = this.trendLines.slice(-this.bufferSize);
        }

        const currentIndex = this.candles.length - 1;
        this.calculateTrendAtIndex(currentIndex);

        const direction = this.directions[currentIndex];
        const trendLine = this.trendLines[currentIndex];

        if (isNaN(direction) || isNaN(trendLine)) {
            return null;
        }

        // Color based on direction: -1 = uptrend (cyan), 1 = downtrend (orange)
        const color = direction === -1 ? '#1cc2d8' : '#e49013';
        this.trendLineData.push({
            time: candle.time,
            value: trendLine,
            color
        });

        if (this.trendLineData.length > this.bufferSize) {
            this.trendLineData = this.trendLineData.slice(-this.bufferSize);
        }

        let signal: StrategySignal | null = null;

        // Check for trend change (ta.cross equivalent)
        if (this.lastDirection !== null && this.lastDirection !== direction) {
            signal = {
                time: candle.time,
                price: candle.close,
                direction: direction === -1 ? 'up' : 'down',
                type: 'trend_change'
            };
            this.signals.push(signal);
        }

        this.lastDirection = direction;

        if (this.signals.length > 1000) {
            this.signals = this.signals.slice(-1000);
        }

        return signal;
    }

    public processHistoricalData(candles: CandleData[]): StrategySignal[] {
        this.candles = [];
        this.signals = [];
        this.trendLineData = [];
        this.upperBands = [];
        this.lowerBands = [];
        this.directions = [];
        this.trendLines = [];
        this.lastDirection = null;

        const allSignals: StrategySignal[] = [];

        for (const candle of candles) {
            const signal = this.updateCandle(candle);
            if (signal && signal.type === 'trend_change') {
                allSignals.push(signal);
            }
        }

        return allSignals;
    }

    public getTrendLineData(): TrendLinePoint[] {
        return [...this.trendLineData];
    }

    public getSignals(): StrategySignal[] {
        return [...this.signals];
    }

    public getCurrentTrend(): string | null {
        if (this.lastDirection === null) return null;
        return this.lastDirection === -1 ? 'UPTREND' : 'DOWNTREND';
    }

    public getCurrentDirection(): number | null {
        return this.lastDirection;
    }

    public reset(): void {
        this.candles = [];
        this.signals = [];
        this.trendLineData = [];
        this.upperBands = [];
        this.lowerBands = [];
        this.directions = [];
        this.trendLines = [];
        this.lastDirection = null;
    }
}
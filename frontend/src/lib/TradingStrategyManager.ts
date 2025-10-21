import { FetchCandles, StrategyBacktest, StrategyRun, StopLiveStrategy, GetRunningStrategies, CloseStrategyPosition } from '@/../wailsjs/go/main/App';
import { main } from '@/../wailsjs/go/models';
import { useChartStore } from '@/store/chartStore';

export class TradingStrategyManager {
    private static instance: TradingStrategyManager;
    private pendingRequests: Map<string, AbortController> = new Map();
    private debounceTimers: Map<string, NodeJS.Timeout> = new Map();

    private constructor() { }

    static getInstance(): TradingStrategyManager {
        if (!TradingStrategyManager.instance) {
            TradingStrategyManager.instance = new TradingStrategyManager();
        }
        return TradingStrategyManager.instance;
    }

    private cancelPendingRequest(key: string): void {
        const controller = this.pendingRequests.get(key);
        if (controller) {
            controller.abort();
            this.pendingRequests.delete(key);
        }
        const timer = this.debounceTimers.get(key);
        if (timer) {
            clearTimeout(timer);
            this.debounceTimers.delete(key);
        }
    }

    async loadData(symbol: string, interval: string, limit: number = 5000, initialViewport: number = 1000): Promise<void> {
        const key = `load-${symbol}-${interval}-${limit}`;
        this.cancelPendingRequest(key);

        const { setChartData, setLoading, setAllCandles } = useChartStore.getState();

        try {
            setLoading(true);
            const allCandles = await FetchCandles(symbol, interval, limit);

            const viewportCandles = allCandles.slice(-initialViewport);
            const viewportStart = Math.max(0, allCandles.length - initialViewport);

            setAllCandles(allCandles);
            setChartData({
                candles: viewportCandles,
                strategyOutput: null,
                fullStrategyOutput: null,
                symbol,
                interval,
                loadedRange: { start: viewportStart, end: allCandles.length },
                totalAvailable: allCandles.length,
            });
        } catch (error) {
            console.error('Failed to load data:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    private sliceStrategyOutput(output: main.BacktestOutput | null, start: number, end: number): main.BacktestOutput | null {
        if (!output) return null;

        return new main.BacktestOutput({
            TrendLines: output.TrendLines?.slice(start, end) || [],
            TrendColors: output.TrendColors?.slice(start, end) || [],
            Directions: output.Directions?.slice(start, end) || [],
            Labels: output.Labels || [],
            Signals: output.Signals || [],
            Positions: output.Positions || [],
            StrategyName: output.StrategyName,
            StrategyVersion: output.StrategyVersion,
            TotalPnL: output.TotalPnL,
            TotalPnLPercent: output.TotalPnLPercent,
            WinRate: output.WinRate,
            TotalTrades: output.TotalTrades,
            WinningTrades: output.WinningTrades,
            LosingTrades: output.LosingTrades,
            AverageWin: output.AverageWin,
            AverageLoss: output.AverageLoss,
            ProfitFactor: output.ProfitFactor,
            MaxDrawdown: output.MaxDrawdown,
            MaxDrawdownPercent: output.MaxDrawdownPercent,
            SharpeRatio: output.SharpeRatio,
            LongestWinStreak: output.LongestWinStreak,
            LongestLossStreak: output.LongestLossStreak,
            AverageHoldTime: output.AverageHoldTime,
        });
    }

    async applyStrategy(
        symbol: string,
        interval: string,
        limit: number,
        strategyId: string,
        params: Record<string, any>
    ): Promise<main.BacktestOutput> {
        const key = `apply-${symbol}-${interval}-${strategyId}`;
        this.cancelPendingRequest(key);

        const { setLoading, setChartData, chartData } = useChartStore.getState();

        try {
            setLoading(true);

            const fullStrategyOutput = await StrategyBacktest(
                symbol,
                interval,
                limit,
                params
            );

            const { loadedRange } = chartData;
            const viewportStrategyOutput = this.sliceStrategyOutput(
                fullStrategyOutput,
                loadedRange.start,
                loadedRange.end
            );

            setChartData({
                strategyOutput: viewportStrategyOutput,
                fullStrategyOutput: fullStrategyOutput,
            });

            return fullStrategyOutput;
        } catch (error) {
            console.error('Failed to apply strategy:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    async rerunStrategy(): Promise<void> {
        const { chartData, setLoading } = useChartStore.getState();

        if (!chartData.candles.length || !chartData.symbol || !chartData.interval) {
            console.warn('No candles loaded or missing symbol/interval');
            return;
        }

        try {
            setLoading(true);
        } catch (error) {
            console.error('Failed to rerun strategy:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    async startLiveStrategy(
        name: string,
        symbol: string,
        interval: string,
        params: Record<string, any>
    ): Promise<void> {
        return StrategyRun(name, symbol, interval, params);
    }

    async stopLiveStrategy(id: string): Promise<void> {
        return StopLiveStrategy(id);
    }

    async closePosition(id: string): Promise<void> {
        return CloseStrategyPosition(id);
    }

    async getRunningStrategies(): Promise<any[]> {
        return GetRunningStrategies();
    }
}

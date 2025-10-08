import { FetchCandles, StrategyRun, FetchAndApplyStrategy } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import { useChartStore } from '../store/chartStore';

const StartLiveStrategy = (window as any).go?.main?.App?.StartLiveStrategy;
const StopLiveStrategy = (window as any).go?.main?.App?.StopLiveStrategy;
const GetRunningStrategies = (window as any).go?.main?.App?.GetRunningStrategies;

export class TradingStrategyManager {
    private static instance: TradingStrategyManager;
    private pendingRequests: Map<string, AbortController> = new Map();
    private debounceTimers: Map<string, NodeJS.Timeout> = new Map();

    private constructor() {}

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
            const strategyOutput = await StrategyRun(allCandles);

            const viewportCandles = allCandles.slice(-initialViewport);

            setAllCandles(allCandles);
            setChartData({
                candles: viewportCandles,
                strategyOutput,
                symbol,
                interval,
                loadedRange: { start: Math.max(0, allCandles.length - initialViewport), end: allCandles.length },
                totalAvailable: allCandles.length,
            });
        } catch (error) {
            console.error('Failed to load data:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    loadMoreCandles(direction: 'left' | 'right', amount: number = 500): void {
        const { chartData, appendCandles, setChartData } = useChartStore.getState();
        const { allCandles, loadedRange } = chartData;

        if (direction === 'left') {
            const newStart = Math.max(0, loadedRange.start - amount);
            if (newStart < loadedRange.start) {
                const newCandles = allCandles.slice(newStart, loadedRange.start);
                appendCandles(newCandles, 'left');
                setChartData({
                    loadedRange: { start: newStart, end: loadedRange.end },
                });
            }
        } else {
            const newEnd = Math.min(allCandles.length, loadedRange.end + amount);
            if (newEnd > loadedRange.end) {
                const newCandles = allCandles.slice(loadedRange.end, newEnd);
                appendCandles(newCandles, 'right');
                setChartData({
                    loadedRange: { start: loadedRange.start, end: newEnd },
                });
            }
        }
    }

    async applyStrategy(
        symbol: string,
        interval: string,
        limit: number,
        strategyId: string,
        params: Record<string, any>
    ): Promise<main.StrategyOutputV2> {
        const key = `apply-${symbol}-${interval}-${strategyId}`;
        this.cancelPendingRequest(key);

        const { setLoading, updateStrategyOutput } = useChartStore.getState();

        try {
            setLoading(true);
            const strategyOutput = await FetchAndApplyStrategy(
                symbol,
                interval,
                limit,
                strategyId,
                params
            );

            updateStrategyOutput(strategyOutput);
            return strategyOutput;
        } catch (error) {
            console.error('Failed to apply strategy:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    async rerunStrategy(): Promise<void> {
        const { chartData, updateStrategyOutput, setLoading } = useChartStore.getState();

        if (!chartData.candles.length) {
            console.warn('No candles loaded');
            return;
        }

        try {
            setLoading(true);
            const strategyOutput = await StrategyRun(chartData.candles);
            updateStrategyOutput(strategyOutput);
        } catch (error) {
            console.error('Failed to rerun strategy:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    async startLiveStrategy(
        id: string,
        strategyId: string,
        params: Record<string, any>,
        symbol: string,
        interval: string
    ): Promise<void> {
        if (!StartLiveStrategy) {
            throw new Error('StartLiveStrategy function not available. Please restart the app.');
        }
        return StartLiveStrategy(id, strategyId, params, symbol, interval);
    }

    async stopLiveStrategy(id: string): Promise<void> {
        if (!StopLiveStrategy) {
            throw new Error('StopLiveStrategy function not available. Please restart the app.');
        }
        return StopLiveStrategy(id);
    }

    async getRunningStrategies(): Promise<any[]> {
        if (!GetRunningStrategies) {
            throw new Error('GetRunningStrategies function not available. Please restart the app.');
        }
        return GetRunningStrategies();
    }
}

import { FetchCandles, StrategyRun, FetchAndApplyStrategy } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';
import { useChartStore } from '../store/chartStore';

// Import dynamically to handle binding generation
const StartLiveStrategy = (window as any).go?.main?.App?.StartLiveStrategy;
const StopLiveStrategy = (window as any).go?.main?.App?.StopLiveStrategy;
const GetRunningStrategies = (window as any).go?.main?.App?.GetRunningStrategies;

export class TradingStrategyManager {
    private static instance: TradingStrategyManager;

    private constructor() {}

    static getInstance(): TradingStrategyManager {
        if (!TradingStrategyManager.instance) {
            TradingStrategyManager.instance = new TradingStrategyManager();
        }
        return TradingStrategyManager.instance;
    }

    async loadData(symbol: string, interval: string, limit: number = 5000): Promise<void> {
        const { setChartData, setLoading } = useChartStore.getState();

        try {
            setLoading(true);
            const candles = await FetchCandles(symbol, interval, limit);
            const strategyOutput = await StrategyRun(candles);

            setChartData({
                candles,
                strategyOutput,
                symbol,
                interval,
            });
        } catch (error) {
            console.error('Failed to load data:', error);
            throw error;
        } finally {
            setLoading(false);
        }
    }

    async applyStrategy(
        symbol: string,
        interval: string,
        limit: number,
        strategyId: string,
        params: Record<string, any>
    ): Promise<main.StrategyOutputV2> {
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

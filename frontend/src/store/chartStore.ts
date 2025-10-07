import { create } from 'zustand';
import { hyperliquid, main } from '../../wailsjs/go/models';

export interface ChartData {
    candles: hyperliquid.Candle[];
    strategyOutput: main.StrategyOutputV2 | null;
    symbol: string;
    interval: string;
    isLoading: boolean;
}

interface ChartStore {
    chartData: ChartData;
    setChartData: (data: Partial<ChartData>) => void;
    updateStrategyOutput: (output: main.StrategyOutputV2) => void;
    setLoading: (loading: boolean) => void;
}

export const useChartStore = create<ChartStore>((set) => ({
    chartData: {
        candles: [],
        strategyOutput: null,
        symbol: 'BTC',
        interval: '15m',
        isLoading: false,
    },
    setChartData: (data) =>
        set((state) => ({
            chartData: { ...state.chartData, ...data },
        })),
    updateStrategyOutput: (output) =>
        set((state) => ({
            chartData: { ...state.chartData, strategyOutput: output },
        })),
    setLoading: (loading) =>
        set((state) => ({
            chartData: { ...state.chartData, isLoading: loading },
        })),
}));

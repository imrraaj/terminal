export interface StrategyParameter {
    name: string;
    label: string;
    type: 'input' | 'select' | 'radio';
    inputType?: 'number' | 'text' | 'color';
    options?: { value: string | number; label: string }[];
    defaultValue: string | number;
    step?: number;
    min?: number;
    max?: number;
}

export interface Strategy {
    id: string;
    name: string;
    description: string;
    parameters: StrategyParameter[];
}

export const STRATEGIES: Strategy[] = [
    {
        id: 'max-trend',
        name: 'Max Trend Points',
        description: 'Trend-following strategy using Hull Moving Average',
        parameters: [
            {
                name: 'factor',
                label: 'Factor',
                type: 'input',
                inputType: 'number',
                defaultValue: 2.5,
                step: 0.1,
                min: 0.1,
                max: 10
            },
        ]
    },
    {
        id: 'rsi-strategy',
        name: 'RSI Strategy',
        description: 'Mean reversion strategy using RSI indicator',
        parameters: [
            {
                name: 'period',
                label: 'RSI Period',
                type: 'input',
                inputType: 'number',
                defaultValue: 14,
                step: 1,
                min: 2,
                max: 50
            },
            {
                name: 'overbought',
                label: 'Overbought Level',
                type: 'input',
                inputType: 'number',
                defaultValue: 70,
                step: 1,
                min: 50,
                max: 100
            },
            {
                name: 'oversold',
                label: 'Oversold Level',
                type: 'input',
                inputType: 'number',
                defaultValue: 30,
                step: 1,
                min: 0,
                max: 50
            },
            {
                name: 'mode',
                label: 'Trading Mode',
                type: 'select',
                defaultValue: 'both',
                options: [
                    { value: 'both', label: 'Long & Short' },
                    { value: 'long', label: 'Long Only' },
                    { value: 'short', label: 'Short Only' }
                ]
            }
        ]
    }
];

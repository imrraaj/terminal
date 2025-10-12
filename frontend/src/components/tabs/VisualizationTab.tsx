import { useState, useEffect } from "react";
import { TradingChart } from "@/components/TradingChart";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { STRATEGIES, Strategy, StrategyParameter } from "@/types/strategy";
import { TradingStrategyManager } from "@/lib/TradingStrategyManager";
import { useChartStore } from "@/store/chartStore";
import { TIMEFRAMES, SYMBOLS } from "@/config/trading";

const strategyManager = TradingStrategyManager.getInstance();

export function VisualizationTab() {
    const { chartData } = useChartStore();
    const [symbol, setSymbol] = useState("BTC");
    const [timeframe, setTimeframe] = useState("5m");
    const [selectedStrategy, setSelectedStrategy] = useState<Strategy>(
        STRATEGIES[0]
    );
    const [strategyParams, setStrategyParams] = useState<Record<string, any>>(
        selectedStrategy.parameters.reduce(
            (acc, param) => ({
                ...acc,
                [param.name]: param.defaultValue,
            }),
            {}
        )
    );
    const [takeProfitPercent, setTakeProfitPercent] = useState(2.0);
    const [stopLossPercent, setStopLossPercent] = useState(2.0);
    const [tradeDirection, setTradeDirection] = useState<
        "both" | "long" | "short"
    >("long");
    const [strategyApplied, setStrategyApplied] = useState(false);

    const LIMIT = 7000;
    const INITIAL_VIEWPORT = 1000;

    useEffect(() => {
        strategyManager.loadData(symbol, timeframe, LIMIT, INITIAL_VIEWPORT);
    }, [symbol, timeframe]);

    const currentTimeframe =
        TIMEFRAMES.find((tf) => tf.value === timeframe) || TIMEFRAMES[3];

    const handleStrategyChange = (strategyId: string) => {
        const strategy = STRATEGIES.find((s) => s.id === strategyId);
        if (strategy) {
            setSelectedStrategy(strategy);
            setStrategyParams(
                strategy.parameters.reduce(
                    (acc, param) => ({
                        ...acc,
                        [param.name]: param.defaultValue,
                    }),
                    {}
                )
            );
            setStrategyApplied(false);
        }
    };

    const handleApplyStrategy = async () => {
        try {
            const paramsWithTPSL = {
                ...strategyParams,
                takeProfitPercent,
                stopLossPercent,
                tradeDirection,
            };

            await strategyManager.applyStrategy(
                symbol,
                currentTimeframe.value,
                LIMIT,
                selectedStrategy.id,
                paramsWithTPSL
            );

            setStrategyApplied(true);
        } catch (error) {
            console.error("Failed to apply strategy:", error);
        }
    };

    const handleStartStrategy = async () => {
        try {
            if (!strategyApplied) {
                alert("Please apply strategy first");
                return;
            }

            const strategyId = `${selectedStrategy.id}-${symbol}-${timeframe}-${Date.now()}`;

            const params = {
                ...strategyParams,
                takeProfitPercent,
                stopLossPercent,
                tradeDirection,
            };

            await strategyManager.startLiveStrategy(
                strategyId,
                selectedStrategy.id,
                params,
                symbol,
                timeframe
            );

            alert(`Strategy ${selectedStrategy.name} started successfully!`);
        } catch (error) {
            console.error("Failed to start strategy:", error);
            alert(`Failed to start strategy: ${error}`);
        }
    };

    const renderParameter = (param: StrategyParameter) => {
        if (param.type === "input") {
            if (param.inputType === "color") {
                return (
                    <div key={param.name} className="space-y-2">
                        <Label htmlFor={param.name}>{param.label}</Label>
                        <div className="flex gap-2">
                            <input
                                type="color"
                                value={strategyParams[param.name]}
                                onChange={(e) =>
                                    setStrategyParams({
                                        ...strategyParams,
                                        [param.name]: e.target.value,
                                    })
                                }
                                className="w-12 h-10 border rounded cursor-pointer"
                            />
                            <Input
                                id={param.name}
                                type="text"
                                value={strategyParams[param.name]}
                                onChange={(e) =>
                                    setStrategyParams({
                                        ...strategyParams,
                                        [param.name]: e.target.value,
                                    })
                                }
                                className="flex-1"
                            />
                        </div>
                    </div>
                );
            }
            return (
                <div key={param.name} className="space-y-2">
                    <Label htmlFor={param.name}>{param.label}</Label>
                    <Input
                        id={param.name}
                        type={param.inputType || "text"}
                        step={param.step}
                        min={param.min}
                        max={param.max}
                        value={strategyParams[param.name]}
                        onChange={(e) =>
                            setStrategyParams({
                                ...strategyParams,
                                [param.name]:
                                    param.inputType === "number"
                                        ? parseFloat(e.target.value)
                                        : e.target.value,
                            })
                        }
                    />
                </div>
            );
        }

        if (param.type === "select") {
            return (
                <div key={param.name} className="space-y-2">
                    <Label htmlFor={param.name}>{param.label}</Label>
                    <Select
                        value={String(strategyParams[param.name])}
                        onValueChange={(value) =>
                            setStrategyParams({
                                ...strategyParams,
                                [param.name]: value,
                            })
                        }
                    >
                        <SelectTrigger id={param.name}>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {param.options?.map((opt) => (
                                <SelectItem
                                    key={String(opt.value)}
                                    value={String(opt.value)}
                                >
                                    {opt.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            );
        }

        return null;
    };

    // Format timestamp to readable date
    const formatDate = (timestamp: number) => {
        if (!timestamp) return "N/A";
        return new Date(timestamp).toLocaleString();
    };

    return (
        <div className="h-full flex flex-col p-4 gap-4 overflow-y-auto">
            <div className="flex gap-4 flex-shrink-0">
                <div className="flex-1 min-w-0">
                    <Card className="h-full flex flex-col">
                        <CardHeader className="pb-3 flex-shrink-0">
                            <div className="flex items-center gap-4">
                                <div className="relative">
                                    <Select
                                        value={symbol}
                                        onValueChange={setSymbol}
                                    >
                                        <SelectTrigger className="w-[140px]">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {SYMBOLS.map((sym) => (
                                                <SelectItem
                                                    key={sym.value}
                                                    value={sym.value}
                                                >
                                                    {sym.label}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>

                                <Select
                                    value={timeframe}
                                    onValueChange={setTimeframe}
                                >
                                    <SelectTrigger className="w-[140px]">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {TIMEFRAMES.map((tf) => (
                                            <SelectItem
                                                key={tf.value}
                                                value={tf.value}
                                            >
                                                {tf.label}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>

                                <Badge variant="secondary">Backtest Mode</Badge>
                            </div>
                        </CardHeader>
                        <CardContent className="w-full h-full overflow-hidden p-4 relative z-0">
                            <TradingChart
                                intervalSeconds={currentTimeframe.seconds}
                            />
                        </CardContent>
                    </Card>
                </div>

                <div className="w-80 flex-shrink-0 flex flex-col gap-4 overflow-y-auto">
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-lg">
                                Strategy Configuration
                            </CardTitle>
                            <CardDescription>
                                Configure backtest parameters
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="viz-strategy">Strategy</Label>
                                <Select
                                    value={selectedStrategy.id}
                                    onValueChange={handleStrategyChange}
                                >
                                    <SelectTrigger id="viz-strategy">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {STRATEGIES.map((strategy) => (
                                            <SelectItem
                                                key={strategy.id}
                                                value={strategy.id}
                                            >
                                                {strategy.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <p className="text-xs text-muted-foreground">
                                    {selectedStrategy.description}
                                </p>
                            </div>

                            <Separator />

                            {selectedStrategy.parameters.map(renderParameter)}

                            <Separator />

                            <div className="space-y-4">
                                <div className="space-y-2">
                                    <Label htmlFor="tradeDirection">
                                        Trade Direction
                                    </Label>
                                    <Select
                                        value={tradeDirection}
                                        onValueChange={(value) =>
                                            setTradeDirection(
                                                value as
                                                    | "both"
                                                    | "long"
                                                    | "short"
                                            )
                                        }
                                    >
                                        <SelectTrigger id="tradeDirection">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="both">
                                                Both (Long & Short)
                                            </SelectItem>
                                            <SelectItem value="long">
                                                Long Only
                                            </SelectItem>
                                            <SelectItem value="short">
                                                Short Only
                                            </SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <p className="text-xs text-muted-foreground">
                                        Filter which trade directions to execute
                                    </p>
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="takeProfit">
                                        Take Profit (%)
                                    </Label>
                                    <Input
                                        id="takeProfit"
                                        type="number"
                                        step="0.1"
                                        min="0.1"
                                        max="100"
                                        value={takeProfitPercent}
                                        onChange={(e) =>
                                            setTakeProfitPercent(
                                                parseFloat(e.target.value)
                                            )
                                        }
                                    />
                                    <p className="text-xs text-muted-foreground">
                                        Close position when profit reaches this
                                        percentage
                                    </p>
                                </div>

                                <div className="space-y-2">
                                    <Label htmlFor="stopLoss">
                                        Stop Loss (%)
                                    </Label>
                                    <Input
                                        id="stopLoss"
                                        type="number"
                                        step="0.1"
                                        min="0.1"
                                        max="100"
                                        value={stopLossPercent}
                                        onChange={(e) =>
                                            setStopLossPercent(
                                                parseFloat(e.target.value)
                                            )
                                        }
                                    />
                                    <p className="text-xs text-muted-foreground">
                                        Close position when loss reaches this
                                        percentage
                                    </p>
                                </div>
                            </div>

                            <div className="flex gap-2">
                                <Button
                                    className="flex-1"
                                    variant="outline"
                                    onClick={handleApplyStrategy}
                                >
                                    Apply Strategy
                                </Button>
                                <Button
                                    className="flex-1"
                                    onClick={handleStartStrategy}
                                    disabled={!strategyApplied}
                                >
                                    Start
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Backtest Results and Trades Table - Always visible */}
            <Card className="flex-shrink-0">
                <CardHeader>
                    <CardTitle className="text-lg">Backtest Results</CardTitle>
                    <CardDescription>
                        {strategyApplied &&
                        chartData.strategyOutput?.BacktestResult
                            ? `${chartData.strategyOutput.StrategyName} v${
                                  chartData.strategyOutput.StrategyVersion
                              } - ${
                                  chartData.strategyOutput.BacktestResult
                                      .Positions?.length || 0
                              } trades executed`
                            : "No backtest results - apply a strategy to see results"}
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    {strategyApplied &&
                    chartData.strategyOutput?.BacktestResult ? (
                        <>
                            {/* Strategy Parameters */}
                            <div className="bg-muted/50 rounded-lg p-4">
                                <h3 className="font-semibold mb-3 text-sm">Strategy Parameters</h3>
                                <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                                    {selectedStrategy.parameters.map((param) => (
                                        <div key={param.name} className="flex justify-between">
                                            <span className="text-muted-foreground">{param.label}:</span>
                                            <span className="font-medium">
                                                {typeof strategyParams[param.name] === 'number'
                                                    ? strategyParams[param.name].toFixed(param.inputType === 'number' ? 2 : 0)
                                                    : strategyParams[param.name]}
                                            </span>
                                        </div>
                                    ))}
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">Trade Direction:</span>
                                        <span className="font-medium capitalize">{tradeDirection}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">Take Profit:</span>
                                        <span className="font-medium">{takeProfitPercent}%</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-muted-foreground">Stop Loss:</span>
                                        <span className="font-medium">{stopLossPercent}%</span>
                                    </div>
                                </div>
                            </div>

                            <Separator />

                            {/* Performance Metrics Grid */}
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                <div className="space-y-1">
                                    <p className="text-xs text-muted-foreground">
                                        Total PnL
                                    </p>
                                    <p
                                        className={`text-2xl font-bold ${
                                            chartData.strategyOutput
                                                .BacktestResult.TotalPnL >= 0
                                                ? "text-green-500"
                                                : "text-red-500"
                                        }`}
                                    >
                                        $
                                        {chartData.strategyOutput.BacktestResult.TotalPnL.toFixed(
                                            2
                                        )}
                                        <span className="text-sm ml-2">
                                            ({chartData.strategyOutput.BacktestResult.TotalPnLPercent >= 0 ? '+' : ''}
                                            {chartData.strategyOutput.BacktestResult.TotalPnLPercent.toFixed(2)}%)
                                        </span>
                                    </p>
                                </div>
                                <div className="space-y-1">
                                    <p className="text-xs text-muted-foreground">
                                        Win Rate
                                    </p>
                                    <p className="text-2xl font-bold">
                                        {chartData.strategyOutput.BacktestResult.WinRate.toFixed(
                                            1
                                        )}
                                        %
                                    </p>
                                </div>
                                <div className="space-y-1">
                                    <p className="text-xs text-muted-foreground">
                                        Profit Factor
                                    </p>
                                    <p
                                        className={`text-2xl font-bold ${
                                            chartData.strategyOutput
                                                .BacktestResult.ProfitFactor >=
                                            1
                                                ? "text-green-500"
                                                : "text-red-500"
                                        }`}
                                    >
                                        {chartData.strategyOutput.BacktestResult.ProfitFactor.toFixed(
                                            2
                                        )}
                                    </p>
                                </div>
                                <div className="space-y-1">
                                    <p className="text-xs text-muted-foreground">
                                        Total Trades
                                    </p>
                                    <p className="text-2xl font-bold">
                                        {
                                            chartData.strategyOutput
                                                .BacktestResult.TotalTrades
                                        }
                                    </p>
                                </div>
                            </div>

                            <Separator />

                            {/* Additional Metrics */}
                            <div className="grid grid-cols-2 md:grid-cols-3 gap-3 text-sm">
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Winning Trades
                                    </span>
                                    <span className="font-medium text-green-500">
                                        {
                                            chartData.strategyOutput
                                                .BacktestResult.WinningTrades
                                        }
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Losing Trades
                                    </span>
                                    <span className="font-medium text-red-500">
                                        {
                                            chartData.strategyOutput
                                                .BacktestResult.LosingTrades
                                        }
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Avg Win
                                    </span>
                                    <span className="font-medium text-green-500">
                                        $
                                        {chartData.strategyOutput.BacktestResult.AverageWin.toFixed(
                                            2
                                        )}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Avg Loss
                                    </span>
                                    <span className="font-medium text-red-500">
                                        $
                                        {Math.abs(
                                            chartData.strategyOutput
                                                .BacktestResult.AverageLoss
                                        ).toFixed(2)}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Max Drawdown
                                    </span>
                                    <span className="font-medium text-red-500">
                                        $
                                        {chartData.strategyOutput.BacktestResult.MaxDrawdown.toFixed(
                                            2
                                        )}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Sharpe Ratio
                                    </span>
                                    <span className="font-medium">
                                        {chartData.strategyOutput.BacktestResult.SharpeRatio.toFixed(
                                            2
                                        )}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Win Streak
                                    </span>
                                    <span className="font-medium text-green-500">
                                        {
                                            chartData.strategyOutput
                                                .BacktestResult.LongestWinStreak
                                        }
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">
                                        Loss Streak
                                    </span>
                                    <span className="font-medium text-red-500">
                                        {
                                            chartData.strategyOutput
                                                .BacktestResult
                                                .LongestLossStreak
                                        }
                                    </span>
                                </div>
                            </div>

                            <Separator />

                            {/* Trades Table */}
                            {chartData.strategyOutput.BacktestResult
                                .Positions &&
                            chartData.strategyOutput.BacktestResult.Positions
                                .length > 0 ? (
                                <div>
                                    <h3 className="font-semibold mb-3">
                                        Trade History
                                    </h3>
                                    <div className="max-h-80 overflow-y-auto">
                                        <Table>
                                            <TableHeader>
                                                <TableRow>
                                                    <TableHead className="w-12">
                                                        #
                                                    </TableHead>
                                                    <TableHead>Side</TableHead>
                                                    <TableHead>
                                                        Entry Price
                                                    </TableHead>
                                                    <TableHead>
                                                        Exit Price
                                                    </TableHead>
                                                    <TableHead>Size</TableHead>
                                                    <TableHead>PnL</TableHead>
                                                    <TableHead>PnL %</TableHead>
                                                    <TableHead>
                                                        Exit Reason
                                                    </TableHead>
                                                    <TableHead>
                                                        Entry Time
                                                    </TableHead>
                                                    <TableHead>
                                                        Exit Time
                                                    </TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {chartData.strategyOutput.BacktestResult.Positions.map(
                                                    (position, index) => (
                                                        <TableRow
                                                            key={index}
                                                            className={
                                                                position.PnL >=
                                                                0
                                                                    ? "bg-green-500/5"
                                                                    : "bg-red-500/5"
                                                            }
                                                        >
                                                            <TableCell className="font-medium">
                                                                {index + 1}
                                                            </TableCell>
                                                            <TableCell>
                                                                <Badge
                                                                    variant={
                                                                        position.Side ===
                                                                        "long"
                                                                            ? "default"
                                                                            : "secondary"
                                                                    }
                                                                >
                                                                    {position.Side.toUpperCase()}
                                                                </Badge>
                                                            </TableCell>
                                                            <TableCell className="font-mono">
                                                                $
                                                                {position.EntryPrice.toFixed(
                                                                    2
                                                                )}
                                                            </TableCell>
                                                            <TableCell className="font-mono">
                                                                $
                                                                {position.ExitPrice.toFixed(
                                                                    2
                                                                )}
                                                            </TableCell>
                                                            <TableCell className="font-mono">
                                                                {position.Size.toFixed(
                                                                    4
                                                                )}
                                                            </TableCell>
                                                            <TableCell
                                                                className={`font-semibold ${
                                                                    position.PnL >=
                                                                    0
                                                                        ? "text-green-500"
                                                                        : "text-red-500"
                                                                }`}
                                                            >
                                                                $
                                                                {position.PnL.toFixed(
                                                                    2
                                                                )}
                                                            </TableCell>
                                                            <TableCell
                                                                className={`font-semibold ${
                                                                    position.PnLPercentage >=
                                                                    0
                                                                        ? "text-green-500"
                                                                        : "text-red-500"
                                                                }`}
                                                            >
                                                                {position.PnLPercentage >=
                                                                0
                                                                    ? "+"
                                                                    : ""}
                                                                {position.PnLPercentage.toFixed(
                                                                    2
                                                                )}
                                                                %
                                                            </TableCell>
                                                            <TableCell>
                                                                <Badge
                                                                    variant="outline"
                                                                    className="text-xs"
                                                                >
                                                                    {
                                                                        position.ExitReason
                                                                    }
                                                                </Badge>
                                                            </TableCell>
                                                            <TableCell className="text-xs text-muted-foreground">
                                                                {formatDate(
                                                                    position.EntryTime
                                                                )}
                                                            </TableCell>
                                                            <TableCell className="text-xs text-muted-foreground">
                                                                {formatDate(
                                                                    position.ExitTime
                                                                )}
                                                            </TableCell>
                                                        </TableRow>
                                                    )
                                                )}
                                            </TableBody>
                                        </Table>
                                    </div>
                                </div>
                            ) : (
                                <div className="text-center py-8 text-muted-foreground">
                                    No trades executed
                                </div>
                            )}
                        </>
                    ) : (
                        <div className="text-center py-8 text-muted-foreground">
                            No backtest results to display
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

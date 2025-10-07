import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { X, TrendingUp, TrendingDown } from "lucide-react";
import { TradingStrategyManager } from "@/lib/TradingStrategyManager";

interface ActiveStrategy {
    id: string;
    name: string;
    symbol: string;
    timeframe: string;
    params: Record<string, any>;
    status: 'running' | 'paused';
    positions: Position[];
    realizedPnL: number;
}

interface Position {
    id: string;
    side: 'long' | 'short';
    entryPrice: number;
    currentPrice: number;
    size: number;
    leverage: number;
    margin: number;
    unrealizedPnL: number;
    pnlPercentage: number;
    liquidationPrice: number;
}

const mockStrategies: ActiveStrategy[] = [
    {
        id: '1',
        name: 'Max Trend Points',
        symbol: 'BTC',
        timeframe: '1h',
        params: { factor: 2.5 },
        status: 'running',
        positions: [
            {
                id: 'p1',
                side: 'long',
                entryPrice: 43250.00,
                currentPrice: 43580.00,
                size: 0.5,
                leverage: 5,
                margin: 4325.00,
                unrealizedPnL: 165.00,
                pnlPercentage: 3.81,
                liquidationPrice: 34600.00
            }
        ],
        realizedPnL: 1245.50
    },
    {
        id: '2',
        name: 'Max Trend Points',
        symbol: 'ETH',
        timeframe: '4h',
        params: { factor: 3.0 },
        status: 'running',
        positions: [
            {
                id: 'p2',
                side: 'short',
                entryPrice: 2580.00,
                currentPrice: 2545.00,
                size: 2.0,
                leverage: 3,
                margin: 1720.00,
                unrealizedPnL: 70.00,
                pnlPercentage: 4.07,
                liquidationPrice: 3096.00
            }
        ],
        realizedPnL: 850.25
    },
    {
        id: '3',
        name: 'Max Trend Points',
        symbol: 'SOL',
        timeframe: '1h',
        params: { factor: 2.0 },
        status: 'paused',
        positions: [],
        realizedPnL: -125.00
    }
];

const strategyManager = TradingStrategyManager.getInstance();

export function ActiveStrategiesTab() {
    const [strategies, setStrategies] = useState<any[]>([]);
    const [filterStatus, setFilterStatus] = useState<string>('all');
    const [sortBy, setSortBy] = useState<string>('pnl');
    const [isLoading, setIsLoading] = useState(false);

    // Fetch running strategies
    useEffect(() => {
        const fetchStrategies = async () => {
            try {
                setIsLoading(true);
                const running = await strategyManager.getRunningStrategies();
                setStrategies(running || []);
            } catch (error) {
                console.error('Failed to fetch strategies:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchStrategies();

        // Poll every 5 seconds for updates
        const interval = setInterval(fetchStrategies, 5000);
        return () => clearInterval(interval);
    }, []);

    const handleClosePosition = async (strategyId: string, positionId: string) => {
        console.log('Close position:', strategyId, positionId);
        // TODO: Implement position closing
    };

    const handlePauseStrategy = async (id: string) => {
        console.log('Pause strategy:', id);
        // TODO: Implement pause
    };

    const handleStopStrategy = async (id: string) => {
        try {
            await strategyManager.stopLiveStrategy(id);
            // Refresh strategies list
            const running = await strategyManager.getRunningStrategies();
            setStrategies(running || []);
        } catch (error) {
            console.error('Failed to stop strategy:', error);
            alert(`Failed to stop strategy: ${error}`);
        }
    };

    const filteredStrategies = strategies.filter(s => {
        if (filterStatus === 'all') return true;
        return s.status === filterStatus;
    });

    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Active Strategies</h2>
                    <p className="text-muted-foreground">
                        Monitor and manage your running strategies {isLoading && <span className="text-xs">(Loading...)</span>}
                    </p>
                </div>
                <div className="flex gap-3">
                    <Select value={filterStatus} onValueChange={setFilterStatus}>
                        <SelectTrigger className="w-[140px]">
                            <SelectValue placeholder="Filter by status" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All</SelectItem>
                            <SelectItem value="running">Running</SelectItem>
                            <SelectItem value="paused">Paused</SelectItem>
                        </SelectContent>
                    </Select>
                    <Select value={sortBy} onValueChange={setSortBy}>
                        <SelectTrigger className="w-[140px]">
                            <SelectValue placeholder="Sort by" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="pnl">PnL</SelectItem>
                            <SelectItem value="symbol">Symbol</SelectItem>
                            <SelectItem value="status">Status</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <div className="grid gap-4">
                {filteredStrategies.map((strategy) => (
                    <Card key={strategy.id}>
                        <CardHeader className="pb-3">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                    <CardTitle className="text-lg">{strategy.name}</CardTitle>
                                    <Badge variant={strategy.status === 'running' ? 'default' : 'secondary'}>
                                        {strategy.status}
                                    </Badge>
                                    <span className="text-sm text-muted-foreground">
                                        {strategy.symbol}/USD · {strategy.timeframe} · Factor: {strategy.params.factor}
                                    </span>
                                </div>
                                <div className="flex gap-2">
                                    {strategy.status === 'running' && strategy.positions.length === 0 && (
                                        <>
                                            <Button variant="outline" size="sm" onClick={() => handlePauseStrategy(strategy.id)}>
                                                Pause
                                            </Button>
                                            <Button variant="destructive" size="sm" onClick={() => handleStopStrategy(strategy.id)}>
                                                Stop
                                            </Button>
                                        </>
                                    )}
                                    <div className={`text-sm font-semibold ${strategy.realizedPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                        Realized: ${strategy.realizedPnL.toFixed(2)}
                                    </div>
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            {strategy.positions.length > 0 ? (
                                <div className="space-y-3">
                                    {strategy.positions.map((position) => (
                                        <div key={position.id} className="border rounded-lg p-4 space-y-3">
                                            <div className="flex items-center justify-between">
                                                <div className="flex items-center gap-3">
                                                    <Badge variant={position.side === 'long' ? 'default' : 'destructive'} className="text-xs">
                                                        {position.leverage}x {position.side.toUpperCase()}
                                                    </Badge>
                                                    <span className="font-semibold">{strategy.symbol}/USD</span>
                                                </div>
                                                <Button 
                                                    variant="ghost" 
                                                    size="sm"
                                                    onClick={() => handleClosePosition(strategy.id, position.id)}
                                                >
                                                    <X className="h-4 w-4" />
                                                </Button>
                                            </div>

                                            <div className="grid grid-cols-6 gap-4 text-sm">
                                                <div>
                                                    <p className="text-muted-foreground">Entry Price</p>
                                                    <p className="font-medium">${position.entryPrice.toFixed(2)}</p>
                                                </div>
                                                <div>
                                                    <p className="text-muted-foreground">Mark Price</p>
                                                    <p className="font-medium">${position.currentPrice.toFixed(2)}</p>
                                                </div>
                                                <div>
                                                    <p className="text-muted-foreground">Position</p>
                                                    <p className="font-medium">{position.size} {strategy.symbol}</p>
                                                </div>
                                                <div>
                                                    <p className="text-muted-foreground">Margin</p>
                                                    <p className="font-medium">${position.margin.toFixed(2)}</p>
                                                </div>
                                                <div>
                                                    <p className="text-muted-foreground">Liq. Price</p>
                                                    <p className="font-medium text-red-400">${position.liquidationPrice.toFixed(2)}</p>
                                                </div>
                                                <div>
                                                    <p className="text-muted-foreground">Unrealized PnL</p>
                                                    <div className="flex items-center gap-1">
                                                        {position.unrealizedPnL >= 0 ? (
                                                            <TrendingUp className="h-3 w-3 text-green-500" />
                                                        ) : (
                                                            <TrendingDown className="h-3 w-3 text-red-500" />
                                                        )}
                                                        <p className={`font-semibold ${position.unrealizedPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                                            ${position.unrealizedPnL.toFixed(2)}
                                                        </p>
                                                        <span className={`text-xs ${position.unrealizedPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                                            ({position.pnlPercentage >= 0 ? '+' : ''}{position.pnlPercentage.toFixed(2)}%)
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-sm text-muted-foreground text-center py-4">No active positions</p>
                            )}
                        </CardContent>
                    </Card>
                ))}
            </div>

            {filteredStrategies.length === 0 && (
                <Card>
                    <CardContent className="flex flex-col items-center justify-center py-12">
                        <p className="text-muted-foreground">No strategies found</p>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface PastTrade {
    id: string;
    strategy: string;
    symbol: string;
    side: 'long' | 'short';
    entryPrice: number;
    exitPrice: number;
    size: number;
    pnl: number;
    entryTime: string;
    exitTime: string;
}

// Mock data
const accountData = {
    balance: 50000.00,
    equity: 51245.50,
    unrealizedPnL: 165.00,
    realizedPnL: 2420.50,
    totalPnL: 2585.50,
    marginUsed: 8650.00,
    marginAvailable: 42595.50,
    leverage: 1.2
};

const pastTrades: PastTrade[] = [
    {
        id: '1',
        strategy: 'Max Trend Points',
        symbol: 'BTC',
        side: 'long',
        entryPrice: 42800.00,
        exitPrice: 43250.00,
        size: 0.5,
        pnl: 225.00,
        entryTime: '2024-01-15 10:30',
        exitTime: '2024-01-15 14:45'
    },
    {
        id: '2',
        strategy: 'Max Trend Points',
        symbol: 'ETH',
        side: 'short',
        entryPrice: 2580.00,
        exitPrice: 2545.00,
        size: 2.0,
        pnl: 70.00,
        entryTime: '2024-01-14 16:20',
        exitTime: '2024-01-14 18:30'
    },
    {
        id: '3',
        strategy: 'Max Trend Points',
        symbol: 'BTC',
        side: 'long',
        entryPrice: 43100.00,
        exitPrice: 42950.00,
        size: 0.5,
        pnl: -75.00,
        entryTime: '2024-01-13 09:15',
        exitTime: '2024-01-13 11:00'
    },
    {
        id: '4',
        strategy: 'Max Trend Points',
        symbol: 'SOL',
        side: 'long',
        entryPrice: 98.50,
        exitPrice: 102.30,
        size: 15.0,
        pnl: 57.00,
        entryTime: '2024-01-12 14:00',
        exitTime: '2024-01-12 16:45'
    },
    {
        id: '5',
        strategy: 'Max Trend Points',
        symbol: 'ETH',
        side: 'short',
        entryPrice: 2620.00,
        exitPrice: 2590.00,
        size: 2.0,
        pnl: 60.00,
        entryTime: '2024-01-11 12:30',
        exitTime: '2024-01-11 15:00'
    }
];

export function PortfolioTab() {
    return (
        <div className="p-6 space-y-6">
            <div>
                <h2 className="text-2xl font-bold tracking-tight">Portfolio</h2>
                <p className="text-muted-foreground">
                    Account overview and trading history
                </p>
            </div>

            {/* Account Summary */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Total Balance</CardDescription>
                        <CardTitle className="text-3xl">${accountData.balance.toFixed(2)}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Equity</CardDescription>
                        <CardTitle className="text-3xl">${accountData.equity.toFixed(2)}</CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Total PnL</CardDescription>
                        <CardTitle className={`text-3xl ${accountData.totalPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                            ${accountData.totalPnL.toFixed(2)}
                        </CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Margin Used</CardDescription>
                        <CardTitle className="text-3xl">${accountData.marginUsed.toFixed(0)}</CardTitle>
                    </CardHeader>
                </Card>
            </div>

            {/* Detailed Metrics */}
            <Card>
                <CardHeader>
                    <CardTitle>Account Details</CardTitle>
                    <CardDescription>Detailed account metrics</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-2 gap-6">
                        <div className="space-y-4">
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">Unrealized PnL</span>
                                <span className={`font-semibold ${accountData.unrealizedPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                    ${accountData.unrealizedPnL.toFixed(2)}
                                </span>
                            </div>
                            <Separator />
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">Realized PnL</span>
                                <span className={`font-semibold ${accountData.realizedPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                    ${accountData.realizedPnL.toFixed(2)}
                                </span>
                            </div>
                        </div>
                        <div className="space-y-4">
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">Margin Available</span>
                                <span className="font-semibold">${accountData.marginAvailable.toFixed(2)}</span>
                            </div>
                            <Separator />
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">Leverage</span>
                                <span className="font-semibold">{accountData.leverage.toFixed(1)}x</span>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Trading History */}
            <Card>
                <CardHeader>
                    <CardTitle>Trading History</CardTitle>
                    <CardDescription>Past trades across all strategies</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Strategy</TableHead>
                                <TableHead>Symbol</TableHead>
                                <TableHead>Side</TableHead>
                                <TableHead className="text-right">Entry</TableHead>
                                <TableHead className="text-right">Exit</TableHead>
                                <TableHead className="text-right">Size</TableHead>
                                <TableHead className="text-right">PnL</TableHead>
                                <TableHead>Exit Time</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {pastTrades.map((trade) => (
                                <TableRow key={trade.id}>
                                    <TableCell className="font-medium">{trade.strategy}</TableCell>
                                    <TableCell>{trade.symbol}/USD</TableCell>
                                    <TableCell>
                                        <Badge variant={trade.side === 'long' ? 'default' : 'destructive'}>
                                            {trade.side.toUpperCase()}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="text-right">${trade.entryPrice.toFixed(2)}</TableCell>
                                    <TableCell className="text-right">${trade.exitPrice.toFixed(2)}</TableCell>
                                    <TableCell className="text-right">{trade.size}</TableCell>
                                    <TableCell className={`text-right font-semibold ${trade.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                        ${trade.pnl.toFixed(2)}
                                    </TableCell>
                                    <TableCell className="text-muted-foreground text-sm">{trade.exitTime}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}

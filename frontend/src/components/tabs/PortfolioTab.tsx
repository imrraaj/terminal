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
import { Button } from "@/components/ui/button";
import { usePortfolio } from "@/hooks/usePortfolio";
import { Loader2, RefreshCw, AlertCircle } from "lucide-react";

export function PortfolioTab() {
    const { portfolio, address, loading, error, refresh } = usePortfolio();
    if (loading && !portfolio) {
        return (
            <div className="flex items-center justify-center h-full">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center h-full space-y-4">
                <AlertCircle className="h-12 w-12 text-destructive" />
                <div className="text-center">
                    <h3 className="text-lg font-semibold">Error Loading Portfolio</h3>
                    <p className="text-sm text-muted-foreground">{error}</p>
                    <Button onClick={refresh} variant="outline" className="mt-4">
                        Try Again
                    </Button>
                </div>
            </div>
        );
    }

    if (!portfolio) {
        return null;
    }

    const accountValue = parseFloat(portfolio.Balance.AccountValue);
    const totalMargin = parseFloat(portfolio.Balance.TotalMargin);
    const marginAvailable = accountValue - totalMargin;
    const unrealizedPnL = portfolio.Positions.reduce(
        (sum, pos) => sum + parseFloat(pos.UnrealizedPnl),
        0
    );

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Portfolio</h2>
                    <p className="text-muted-foreground">
                        {address
                            ? `${address.slice(0, 6)}...${address.slice(-4)}`
                            : "Account overview"}
                    </p>
                </div>
                <Button
                    onClick={refresh}
                    variant="outline"
                    size="sm"
                    disabled={loading}
                >
                    {loading ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                        <RefreshCw className="h-4 w-4" />
                    )}
                    <span className="ml-2">Refresh</span>
                </Button>
            </div>

            {/* Account Summary */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Account Value</CardDescription>
                        <CardTitle className="text-3xl">
                            ${accountValue.toFixed(2)}
                        </CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Withdrawable</CardDescription>
                        <CardTitle className="text-3xl">
                            ${parseFloat(portfolio.Balance.Withdrawable).toFixed(2)}
                        </CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Unrealized PnL</CardDescription>
                        <CardTitle
                            className={`text-3xl ${unrealizedPnL >= 0 ? "text-green-500" : "text-red-500"
                                }`}
                        >
                            ${unrealizedPnL.toFixed(2)}
                        </CardTitle>
                    </CardHeader>
                </Card>
                <Card>
                    <CardHeader className="pb-3">
                        <CardDescription>Margin Used</CardDescription>
                        <CardTitle className="text-3xl">
                            ${totalMargin.toFixed(0)}
                        </CardTitle>
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
                                <span className="text-sm text-muted-foreground">
                                    Total Raw USD
                                </span>
                                <span className="font-semibold">
                                    ${parseFloat(portfolio.Balance.TotalRawUsd).toFixed(2)}
                                </span>
                            </div>
                            <Separator />
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">
                                    Open Positions
                                </span>
                                <span className="font-semibold">
                                    {portfolio.TotalPositions}
                                </span>
                            </div>
                        </div>
                        <div className="space-y-4">
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">
                                    Margin Available
                                </span>
                                <span className="font-semibold">
                                    ${marginAvailable.toFixed(2)}
                                </span>
                            </div>
                            <Separator />
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">
                                    Account Leverage
                                </span>
                                <span className="font-semibold">
                                    {portfolio.Balance.AccountLeverage.toFixed(1)}x
                                </span>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Open Positions */}
            <Card>
                <CardHeader>
                    <CardTitle>Open Positions</CardTitle>
                    <CardDescription>Current active positions</CardDescription>
                </CardHeader>
                <CardContent>
                    {portfolio.Positions.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            No open positions
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Symbol</TableHead>
                                    <TableHead>Side</TableHead>
                                    <TableHead className="text-right">Entry</TableHead>
                                    <TableHead className="text-right">Current</TableHead>
                                    <TableHead className="text-right">Size</TableHead>
                                    <TableHead className="text-right">Value</TableHead>
                                    <TableHead className="text-right">PnL</TableHead>
                                    <TableHead className="text-right">ROE</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {portfolio.Positions.map((position, idx) => {
                                    const pnl = parseFloat(position.UnrealizedPnl);
                                    return (
                                        <TableRow key={idx}>
                                            <TableCell className="font-medium">
                                                {position.Coin}/USD
                                            </TableCell>
                                            <TableCell>
                                                <Badge
                                                    variant={
                                                        position.Side === "long" ? "default" : "destructive"
                                                    }
                                                >
                                                    {position.Side.toUpperCase()}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="text-right">
                                                ${parseFloat(position.EntryPrice).toFixed(2)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                ${parseFloat(position.CurrentPrice).toFixed(2)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {position.Size}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                ${parseFloat(position.PositionValue).toFixed(2)}
                                            </TableCell>
                                            <TableCell
                                                className={`text-right font-semibold ${pnl >= 0 ? "text-green-500" : "text-red-500"
                                                    }`}
                                            >
                                                ${pnl.toFixed(2)}
                                            </TableCell>
                                            <TableCell
                                                className={`text-right ${parseFloat(position.ReturnOnEquity) >= 0
                                                        ? "text-green-500"
                                                        : "text-red-500"
                                                    }`}
                                            >
                                                {position.ReturnOnEquity}%
                                            </TableCell>
                                        </TableRow>
                                    );
                                })}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

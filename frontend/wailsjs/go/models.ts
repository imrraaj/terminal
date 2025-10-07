export namespace hyperliquid {
	
	export class Candle {
	    T: number;
	    c: string;
	    h: string;
	    i: string;
	    l: string;
	    n: number;
	    o: string;
	    s: string;
	    t: number;
	    v: string;
	
	    static createFrom(source: any = {}) {
	        return new Candle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.T = source["T"];
	        this.c = source["c"];
	        this.h = source["h"];
	        this.i = source["i"];
	        this.l = source["l"];
	        this.n = source["n"];
	        this.o = source["o"];
	        this.s = source["s"];
	        this.t = source["t"];
	        this.v = source["v"];
	    }
	}

}

export namespace main {
	
	export class AccountBalance {
	    accountValue: string;
	    totalRawUsd: string;
	    withdrawable: string;
	    totalMargin: string;
	    accountLeverage: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountBalance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountValue = source["accountValue"];
	        this.totalRawUsd = source["totalRawUsd"];
	        this.withdrawable = source["withdrawable"];
	        this.totalMargin = source["totalMargin"];
	        this.accountLeverage = source["accountLeverage"];
	    }
	}
	export class ActivePosition {
	    coin: string;
	    side: string;
	    size: string;
	    entryPrice: string;
	    currentPrice: string;
	    liquidationPx: string;
	    unrealizedPnl: string;
	    positionValue: string;
	    leverage: number;
	    marginUsed: string;
	    returnOnEquity: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivePosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coin = source["coin"];
	        this.side = source["side"];
	        this.size = source["size"];
	        this.entryPrice = source["entryPrice"];
	        this.currentPrice = source["currentPrice"];
	        this.liquidationPx = source["liquidationPx"];
	        this.unrealizedPnl = source["unrealizedPnl"];
	        this.positionValue = source["positionValue"];
	        this.leverage = source["leverage"];
	        this.marginUsed = source["marginUsed"];
	        this.returnOnEquity = source["returnOnEquity"];
	    }
	}
	export class Signal {
	    Index: number;
	    Type: number;
	    Price: number;
	    Time: number;
	    Confidence: number;
	    Reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Signal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Type = source["Type"];
	        this.Price = source["Price"];
	        this.Time = source["Time"];
	        this.Confidence = source["Confidence"];
	        this.Reason = source["Reason"];
	    }
	}
	export class Position {
	    EntryIndex: number;
	    EntryPrice: number;
	    EntryTime: number;
	    ExitIndex: number;
	    ExitPrice: number;
	    ExitTime: number;
	    Side: string;
	    Size: number;
	    PnL: number;
	    PnLPercentage: number;
	    IsOpen: boolean;
	    ExitReason: string;
	    MaxDrawdown: number;
	    MaxProfit: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.EntryIndex = source["EntryIndex"];
	        this.EntryPrice = source["EntryPrice"];
	        this.EntryTime = source["EntryTime"];
	        this.ExitIndex = source["ExitIndex"];
	        this.ExitPrice = source["ExitPrice"];
	        this.ExitTime = source["ExitTime"];
	        this.Side = source["Side"];
	        this.Size = source["Size"];
	        this.PnL = source["PnL"];
	        this.PnLPercentage = source["PnLPercentage"];
	        this.IsOpen = source["IsOpen"];
	        this.ExitReason = source["ExitReason"];
	        this.MaxDrawdown = source["MaxDrawdown"];
	        this.MaxProfit = source["MaxProfit"];
	    }
	}
	export class BacktestResult {
	    Positions: Position[];
	    Signals: Signal[];
	    TotalPnL: number;
	    TotalPnLPercent: number;
	    WinRate: number;
	    TotalTrades: number;
	    WinningTrades: number;
	    LosingTrades: number;
	    AverageWin: number;
	    AverageLoss: number;
	    ProfitFactor: number;
	    MaxDrawdown: number;
	    MaxDrawdownPercent: number;
	    SharpeRatio: number;
	    LongestWinStreak: number;
	    LongestLossStreak: number;
	    AverageHoldTime: number;
	
	    static createFrom(source: any = {}) {
	        return new BacktestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Positions = this.convertValues(source["Positions"], Position);
	        this.Signals = this.convertValues(source["Signals"], Signal);
	        this.TotalPnL = source["TotalPnL"];
	        this.TotalPnLPercent = source["TotalPnLPercent"];
	        this.WinRate = source["WinRate"];
	        this.TotalTrades = source["TotalTrades"];
	        this.WinningTrades = source["WinningTrades"];
	        this.LosingTrades = source["LosingTrades"];
	        this.AverageWin = source["AverageWin"];
	        this.AverageLoss = source["AverageLoss"];
	        this.ProfitFactor = source["ProfitFactor"];
	        this.MaxDrawdown = source["MaxDrawdown"];
	        this.MaxDrawdownPercent = source["MaxDrawdownPercent"];
	        this.SharpeRatio = source["SharpeRatio"];
	        this.LongestWinStreak = source["LongestWinStreak"];
	        this.LongestLossStreak = source["LongestLossStreak"];
	        this.AverageHoldTime = source["AverageHoldTime"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ClosePositionRequest {
	    coin: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new ClosePositionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coin = source["coin"];
	        this.size = source["size"];
	    }
	}
	export class Label {
	    Index: number;
	    Price: number;
	    Text: string;
	    Direction: number;
	    Percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new Label(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Index = source["Index"];
	        this.Price = source["Price"];
	        this.Text = source["Text"];
	        this.Direction = source["Direction"];
	        this.Percentage = source["Percentage"];
	    }
	}
	export class OrderRequest {
	    coin: string;
	    isBuy: boolean;
	    size: number;
	    price: number;
	    orderType: string;
	    reduceOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OrderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coin = source["coin"];
	        this.isBuy = source["isBuy"];
	        this.size = source["size"];
	        this.price = source["price"];
	        this.orderType = source["orderType"];
	        this.reduceOnly = source["reduceOnly"];
	    }
	}
	export class OrderResponse {
	    success: boolean;
	    orderId: string;
	    message: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new OrderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.orderId = source["orderId"];
	        this.message = source["message"];
	        this.status = source["status"];
	    }
	}
	export class PortfolioSummary {
	    balance: AccountBalance;
	    positions: ActivePosition[];
	    totalPositions: number;
	    totalPnL: number;
	
	    static createFrom(source: any = {}) {
	        return new PortfolioSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.balance = this.convertValues(source["balance"], AccountBalance);
	        this.positions = this.convertValues(source["positions"], ActivePosition);
	        this.totalPositions = source["totalPositions"];
	        this.totalPnL = source["totalPnL"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class StrategyConfig {
	    TakeProfitPercent: number;
	    StopLossPercent: number;
	    PositionSize: number;
	    UsePercentage: boolean;
	    MaxPositions: number;
	    MaxRiskPerTrade: number;
	    TradeDirection: string;
	    Parameters: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new StrategyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TakeProfitPercent = source["TakeProfitPercent"];
	        this.StopLossPercent = source["StopLossPercent"];
	        this.PositionSize = source["PositionSize"];
	        this.UsePercentage = source["UsePercentage"];
	        this.MaxPositions = source["MaxPositions"];
	        this.MaxRiskPerTrade = source["MaxRiskPerTrade"];
	        this.TradeDirection = source["TradeDirection"];
	        this.Parameters = source["Parameters"];
	    }
	}
	export class Account {
	
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class StrategyInstance {
	    ID: string;
	    Strategy: any;
	    Config: StrategyConfig;
	    Symbol: string;
	    Interval: string;
	    IsRunning: boolean;
	    CurrentPosition?: Position;
	    // Go type: Account
	    Account?: any;
	    LastCandleTime: number;
	
	    static createFrom(source: any = {}) {
	        return new StrategyInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Strategy = source["Strategy"];
	        this.Config = this.convertValues(source["Config"], StrategyConfig);
	        this.Symbol = source["Symbol"];
	        this.Interval = source["Interval"];
	        this.IsRunning = source["IsRunning"];
	        this.CurrentPosition = this.convertValues(source["CurrentPosition"], Position);
	        this.Account = this.convertValues(source["Account"], null);
	        this.LastCandleTime = source["LastCandleTime"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StrategyOutputV2 {
	    TrendLines: number[];
	    TrendColors: string[];
	    Directions: number[];
	    Labels: Label[];
	    Signals: Signal[];
	    BacktestResult: BacktestResult;
	    StrategyName: string;
	    StrategyVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new StrategyOutputV2(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TrendLines = source["TrendLines"];
	        this.TrendColors = source["TrendColors"];
	        this.Directions = source["Directions"];
	        this.Labels = this.convertValues(source["Labels"], Label);
	        this.Signals = this.convertValues(source["Signals"], Signal);
	        this.BacktestResult = this.convertValues(source["BacktestResult"], BacktestResult);
	        this.StrategyName = source["StrategyName"];
	        this.StrategyVersion = source["StrategyVersion"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}


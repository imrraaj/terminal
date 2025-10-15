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
	export class OpenOrder {
	    coin: string;
	    limitPx: number;
	    oid: number;
	    side: string;
	    sz: number;
	    timestamp: number;
	
	    static createFrom(source: any = {}) {
	        return new OpenOrder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.coin = source["coin"];
	        this.limitPx = source["limitPx"];
	        this.oid = source["oid"];
	        this.side = source["side"];
	        this.sz = source["sz"];
	        this.timestamp = source["timestamp"];
	    }
	}

}

export namespace main {
	
	export class AccountBalance {
	    AccountValue: string;
	    TotalRawUsd: string;
	    Withdrawable: string;
	    TotalMargin: string;
	    AccountLeverage: number;
	
	    static createFrom(source: any = {}) {
	        return new AccountBalance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AccountValue = source["AccountValue"];
	        this.TotalRawUsd = source["TotalRawUsd"];
	        this.Withdrawable = source["Withdrawable"];
	        this.TotalMargin = source["TotalMargin"];
	        this.AccountLeverage = source["AccountLeverage"];
	    }
	}
	export class ActivePosition {
	    Coin: string;
	    Side: string;
	    Size: string;
	    EntryPrice: string;
	    CurrentPrice: string;
	    LiquidationPx: string;
	    UnrealizedPnl: string;
	    PositionValue: string;
	    Leverage: number;
	    MarginUsed: string;
	    ReturnOnEquity: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivePosition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Coin = source["Coin"];
	        this.Side = source["Side"];
	        this.Size = source["Size"];
	        this.EntryPrice = source["EntryPrice"];
	        this.CurrentPrice = source["CurrentPrice"];
	        this.LiquidationPx = source["LiquidationPx"];
	        this.UnrealizedPnl = source["UnrealizedPnl"];
	        this.PositionValue = source["PositionValue"];
	        this.Leverage = source["Leverage"];
	        this.MarginUsed = source["MarginUsed"];
	        this.ReturnOnEquity = source["ReturnOnEquity"];
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
	export class Signal {
	    Index: number;
	    Type: number;
	    Price: number;
	    Time: number;
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
	        this.Reason = source["Reason"];
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
	export class BacktestOutput {
	    TrendLines: number[];
	    TrendColors: string[];
	    Directions: number[];
	    Labels: Label[];
	    Signals: Signal[];
	    Positions: Position[];
	    StrategyName: string;
	    StrategyVersion: string;
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
	        return new BacktestOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TrendLines = source["TrendLines"];
	        this.TrendColors = source["TrendColors"];
	        this.Directions = source["Directions"];
	        this.Labels = this.convertValues(source["Labels"], Label);
	        this.Signals = this.convertValues(source["Signals"], Signal);
	        this.Positions = this.convertValues(source["Positions"], Position);
	        this.StrategyName = source["StrategyName"];
	        this.StrategyVersion = source["StrategyVersion"];
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
	
	export class StrategyConfig {
	    PositionSize: number;
	    TradeDirection: string;
	    TakeProfitPercent: number;
	    StopLossPercent: number;
	    Interval: number;
	    Parameters: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new StrategyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PositionSize = source["PositionSize"];
	        this.TradeDirection = source["TradeDirection"];
	        this.TakeProfitPercent = source["TakeProfitPercent"];
	        this.StopLossPercent = source["StopLossPercent"];
	        this.Interval = source["Interval"];
	        this.Parameters = source["Parameters"];
	    }
	}
	export class MaxTrendPointsStrategy {
	    ID: string;
	    Symbol: string;
	    Interval: string;
	    LastCandleTime: number;
	    IsRunning: boolean;
	    Position?: Position;
	    Factor: number;
	    Config: StrategyConfig;
	
	    static createFrom(source: any = {}) {
	        return new MaxTrendPointsStrategy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Symbol = source["Symbol"];
	        this.Interval = source["Interval"];
	        this.LastCandleTime = source["LastCandleTime"];
	        this.IsRunning = source["IsRunning"];
	        this.Position = this.convertValues(source["Position"], Position);
	        this.Factor = source["Factor"];
	        this.Config = this.convertValues(source["Config"], StrategyConfig);
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
	export class PortfolioSummary {
	    Balance: AccountBalance;
	    Positions: ActivePosition[];
	    TotalPositions: number;
	    TotalPnL: number;
	    OpenOrders: hyperliquid.OpenOrder[];
	
	    static createFrom(source: any = {}) {
	        return new PortfolioSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Balance = this.convertValues(source["Balance"], AccountBalance);
	        this.Positions = this.convertValues(source["Positions"], ActivePosition);
	        this.TotalPositions = source["TotalPositions"];
	        this.TotalPnL = source["TotalPnL"];
	        this.OpenOrders = this.convertValues(source["OpenOrders"], hyperliquid.OpenOrder);
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


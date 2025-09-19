import { BarSeries, CandlestickData, CandlestickSeries, ColorType, createChart, IChartApi, LineSeries, OhlcData, Time } from "lightweight-charts";
import { useEffect, useRef, useState } from "react"

type candleSnapshot = {
    T: number,
    t: number,
    o: string,
    h: string,
    l: string,
    c: string,
    v: string,
}
const fetchCandleSnapshotData = async (coin: string, interval: string, startTime: number, endTime: number): Promise<candleSnapshot[]> => {
    const response = await fetch('https://api.hyperliquid.xyz/info', {
        method: "POST",
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            type: "candleSnapshot",
            req: {
                coin,
                interval,
                startTime,
                endTime
            }
        })
    });
    const candles = await response.json();
    return candles;
}

function App() {
    const chartRef = useRef<HTMLDivElement>(null);
    const chartInstanceRef = useRef<any>(null);
    const candleSeriesRef = useRef<any>(null);

    const INTERVAL = 15 * 60; // 15m
    const INTERVAL_STR = '15m';
    const SYMBOL = 'BTC';



    // Initialize chart only once
    useEffect(() => {
        if (!chartRef.current || chartInstanceRef.current) return;

        const handleResize = () => {
            if (chartInstanceRef.current) {
                chartInstanceRef.current.applyOptions({ width: chartRef.current!.clientWidth });
            }
        };

        const chart = createChart(chartRef.current, {
            layout: {
                background: { type: ColorType.Solid, color: '#0f0f0f' },
                textColor: 'white',
                fontFamily: 'Nunito, sans-serif',
                fontSize: 16,
            },
            width: chartRef.current.clientWidth,
            height: window.innerHeight,
            autoSize: true,
            rightPriceScale: {
                borderColor: '#ffffff22',
            },
            timeScale: {
                borderColor: '#ffffff22',
                timeVisible: true,
                secondsVisible: true,
                uniformDistribution: true,
                tickMarkMaxCharacterLength: 5,
                tickMarkFormatter: (time: Time) => {
                    const date = new Date(time as number);
                    if (INTERVAL < 60 * 60) {
                        return date.toLocaleString('en-IN', { hour: '2-digit', minute: '2-digit', hour12: false });
                    } else if (INTERVAL < 60 * 60 * 24) {
                        return date.toLocaleString('en-IN', { day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false });
                    } else {
                        return date.toLocaleString('en-IN', { day: '2-digit', month: 'short' });
                    }
                }
            },
            localization: {
                timeFormatter: (time: Time) => {
                    const date = new Date(time as number);
                    return date.toLocaleString('en-IN', { year: '2-digit', month: 'short', day: '2-digit', hour12: false, hour: '2-digit', minute: '2-digit' });
                }
            },
            grid: {
                vertLines: {
                    color: '#ffffff11',
                    visible: false,
                },
                horzLines: {
                    color: '#ffffff11',
                    visible: false,
                },
            },
        });

        const candleSeries = chart.addSeries(CandlestickSeries, {
            upColor: '#089981',
            downColor: '#ffffff99',
            wickUpColor: '#089981',
            wickDownColor: '#ffffff99',
            borderUpColor: '#089981',
            borderDownColor: '#ffffff99',
            baseLineVisible: false,
            borderVisible: false,
            priceLineVisible: true,
            wickVisible: true,
            title: SYMBOL,
        });

        chartInstanceRef.current = chart;
        candleSeriesRef.current = candleSeries;
        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            if (chartInstanceRef.current) {
                chartInstanceRef.current.remove();
                chartInstanceRef.current = null;
                candleSeriesRef.current = null;
            }
        };
    }, []);

    // Load initial data
    useEffect(() => {
        const loadInitialData = async () => {
            try {
                const endTime = Math.floor(Date.now());
                const startTime = endTime - (INTERVAL * 1000 * 1000); // 1000 candles
                const candles = await fetchCandleSnapshotData(SYMBOL, INTERVAL_STR, startTime, endTime);
                const formattedData = candles.map(c => ({
                    time: c.t as Time,
                    open: parseFloat(c.o),
                    high: parseFloat(c.h),
                    low: parseFloat(c.l),
                    close: parseFloat(c.c),
                }));
                if (candleSeriesRef.current) {
                    candleSeriesRef.current.setData(formattedData);
                    chartInstanceRef.current.timeScale().fitContent();
                }
            } catch (error) {
                console.error('Failed to load initial data:', error);
            }
        };
        loadInitialData();
    }, []);

    // WebSocket connection
    useEffect(() => {
        const ws = new WebSocket('wss://api.hyperliquid.xyz/ws');
        ws.onopen = () => {
            console.log('WebSocket connected');
            ws.send(JSON.stringify({
                "method": "subscribe",
                "subscription": { "type": "candle", "coin": SYMBOL, "interval": INTERVAL_STR }
            }));
        };

        ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                console.log('WebSocket message:', message);
                if (message.channel === 'candle' && message.data) {
                    const candle = message.data as candleSnapshot;
                    const newCandle: CandlestickData<Time> = {
                        time: (candle.T) as Time,
                        open: parseFloat(candle.o),
                        high: parseFloat(candle.h),
                        low: parseFloat(candle.l),
                        close: parseFloat(candle.c),
                    };
                    candleSeriesRef.current.update(newCandle);
                }
            } catch (error) {
                console.error('Error processing WebSocket message:', error);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        ws.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
        };

        return () => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.close();
            }
        };
    }, []); // Empty dependency array - only run once

    return (
        <div ref={chartRef} style={{ width: '100%', height: '100vh' }} />
    );
}

export default App;

import {
    CandlestickSeries,
    ColorType,
    createChart,
    LineSeries,
    Time,
    LineData,
    IChartApi,
    ISeriesApi,
    createSeriesMarkers,
} from "lightweight-charts";
import { useEffect, useRef, useState } from "react";
import { useChartStore } from "../store/chartStore";
import { hyperliquid } from "@/../wailsjs/go/models";
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
    ContextMenuSeparator,
} from "@/components/ui/context-menu";

interface TradingChartProps {
    intervalSeconds: number;
}

export function TradingChart({ intervalSeconds }: TradingChartProps) {
    const chartRef = useRef<HTMLDivElement>(null);
    const chartInstanceRef = useRef<IChartApi | null>(null);
    const candleSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
    const trendLineSeriesRef = useRef<ISeriesApi<"Line">[]>([]);
    const [clickedPrice, setClickedPrice] = useState<number | null>(null);

    const { chartData } = useChartStore();
    const { candles, strategyOutput, symbol } = chartData;

    // Initialize chart once
    useEffect(() => {
        if (!chartRef.current || chartInstanceRef.current) return;

        const handleResize = () => {
            if (chartInstanceRef.current && chartRef.current) {
                chartInstanceRef.current.applyOptions({
                    width: chartRef.current.clientWidth,
                    height: chartRef.current.clientHeight,
                });
            }
        };

        const chart = createChart(chartRef.current, {
            layout: {
                background: { type: ColorType.Solid, color: "#0f0f0f" },
                textColor: "white",
            },
            width: chartRef.current.clientWidth,
            height: chartRef.current.clientHeight,
            autoSize: false,
            rightPriceScale: {
                borderColor: "#ffffff22",
            },
            timeScale: {
                borderColor: "#ffffff22",
                timeVisible: true,
                secondsVisible: true,
                uniformDistribution: true,
                tickMarkMaxCharacterLength: 5,
                tickMarkFormatter: (time: Time) => {
                    const date = new Date((time as number) * 1000);
                    if (intervalSeconds < 60 * 60) {
                        const hours = date
                            .getHours()
                            .toString()
                            .padStart(2, "0");
                        const minutes = date
                            .getMinutes()
                            .toString()
                            .padStart(2, "0");
                        return `${hours}:${minutes}`;
                    } else if (intervalSeconds < 60 * 24) {
                        const month = (date.getMonth() + 1)
                            .toString()
                            .padStart(2, "0");
                        const day = date.getDate().toString().padStart(2, "0");
                        const hours = date
                            .getHours()
                            .toString()
                            .padStart(2, "0");
                        return `${month}/${day} ${hours}:00`;
                    } else {
                        const month = date.toLocaleDateString("en-US", {
                            month: "short",
                        });
                        const day = date.getDate();
                        return `${month} ${day}`;
                    }
                },
            },
            localization: {
                timeFormatter: (time: Time) => {
                    const date = new Date((time as number) * 1000);
                    return date.toLocaleString("en-US", {
                        year: "numeric",
                        month: "short",
                        day: "2-digit",
                        hour12: false,
                        hour: "2-digit",
                        minute: "2-digit",
                    });
                },
            },
            grid: {
                vertLines: {
                    color: "#ffffff11",
                    // visible: false,
                },
                horzLines: {
                    color: "#ffffff11",
                    // visible: false,
                },
            },
        });

        const candleSeries = chart.addSeries(CandlestickSeries, {
            upColor: "#089981cc",
            downColor: "#ffffffcc",
            wickUpColor: "#089981",
            wickDownColor: "#ffffff99",
            borderUpColor: "#089981",
            borderDownColor: "#ffffff99",
            baseLineVisible: false,
            borderVisible: false,
            priceLineVisible: true,
            wickVisible: true,
            title: symbol,
        });
        chartInstanceRef.current = chart;
        candleSeriesRef.current = candleSeries;
        window.addEventListener("resize", handleResize);

        return () => {
            window.removeEventListener("resize", handleResize);
            if (chartInstanceRef.current) {
                chartInstanceRef.current.remove();
                chartInstanceRef.current = null;
                candleSeriesRef.current = null;
                trendLineSeriesRef.current = [];
            }
        };
    }, []);

    // Update chart data when candles or strategy changes
    useEffect(() => {
        if (
            !candleSeriesRef.current ||
            !chartInstanceRef.current ||
            !candles.length
        )
            return;

        const formattedData = candles.map((c: hyperliquid.Candle) => ({
            time: (c.t / 1000) as Time,
            open: parseFloat(c.o),
            high: parseFloat(c.h),
            low: parseFloat(c.l),
            close: parseFloat(c.c),
        }));

        // Set candlestick data
        candleSeriesRef.current.setData(formattedData);

        // Remove old trend line series
        trendLineSeriesRef.current.forEach((series) => {
            chartInstanceRef.current?.removeSeries(series);
        });
        trendLineSeriesRef.current = [];

        if (strategyOutput && strategyOutput.Directions.length > 0) {
            // Create separate line series for each trend segment
            let segmentStart = 0;
            let currentDirection = strategyOutput.Directions[0];

            for (let i = 1; i <= strategyOutput.Directions.length; i++) {
                if (
                    i === strategyOutput.Directions.length ||
                    strategyOutput.Directions[i] !== currentDirection
                ) {
                    const color =
                        currentDirection === -1 ? "#1cc2d8" : "#e49013";
                    const lineSeries = chartInstanceRef.current.addSeries(
                        LineSeries,
                        {
                            color: color,
                            lineWidth: 2,
                            priceLineVisible: false,
                            lastValueVisible: false,
                        }
                    );

                    const segmentData: LineData[] = [];
                    let j;

                    // Find max and min prices in the segment
                    let maxPrice = -Infinity;
                    let minPrice = Infinity;
                    let maxIndex = segmentStart;
                    let minIndex = segmentStart;

                    for (j = segmentStart; j < i; j++) {
                        if (strategyOutput.TrendLines[j] > 0) {
                            segmentData.push({
                                time: (candles[j].t / 1000) as Time,
                                value: strategyOutput.TrendLines[j],
                            });
                        }

                        const high = parseFloat(candles[j].h);
                        const low = parseFloat(candles[j].l);

                        if (high > maxPrice) {
                            maxPrice = high;
                            maxIndex = j;
                        }
                        if (low < minPrice) {
                            minPrice = low;
                            minIndex = j;
                        }
                    }

                    // Calculate exit index (last candle in segment)
                    const exitIndex = i - 1;

                    // Build markers array
                    const markers: any[] = [
                        {
                            time: (candles[segmentStart].t / 1000) as Time,
                            position: "aboveBar",
                            color:
                                currentDirection === -1 ? "#1cc2d8" : "#e49013",
                            shape:
                                currentDirection === 1
                                    ? "arrowDown"
                                    : "arrowUp",
                            text: `Entry: ${candles[segmentStart].c}`,
                        },
                        {
                            time: segmentData[segmentData.length - 1]
                                .time as Time,
                            position: "aboveBar",
                            color:
                                currentDirection === -1 ? "#1cc2d8" : "#e49013",
                            shape:
                                currentDirection === 1
                                    ? "arrowDown"
                                    : "arrowUp",
                            text: `Exit: ${candles[exitIndex].c}`,
                        },
                    ];

                    // Add max/min price markers based on trend direction
                    // if (currentDirection === -1) {
                    //     // Uptrend - show max price reached
                    //     markers.push({
                    //         time: (candles[maxIndex].t / 1000) as Time,
                    //         position: "aboveBar",
                    //         color: "#00ff00",
                    //         shape: "circle",
                    //         text: `High: ${maxPrice.toFixed(2)}`,
                    //     });
                    // } else {
                    //     // Downtrend - show min price reached
                    //     markers.push({
                    //         time: (candles[minIndex].t / 1000) as Time,
                    //         position: "belowBar",
                    //         color: "#ff0000",
                    //         shape: "circle",
                    //         text: `Low: ${minPrice.toFixed(2)}`,
                    //     });
                    // }

                    createSeriesMarkers(lineSeries, markers);
                    lineSeries.setData(segmentData);
                    trendLineSeriesRef.current.push(lineSeries);

                    if (i < strategyOutput.Directions.length) {
                        segmentStart = i;
                        currentDirection = strategyOutput.Directions[i];
                    }
                }
            }
        }

        chartInstanceRef.current.timeScale().fitContent();
    }, [candles, strategyOutput]);

    const handleChartClick = (e: React.MouseEvent) => {
        if (!chartInstanceRef.current || !candleSeriesRef.current) return;
        const rect = chartRef.current?.getBoundingClientRect();
        if (!rect) return;
        const y = e.clientY - rect.top;
        const price = candleSeriesRef.current.coordinateToPrice(y);
        if (price) setClickedPrice(price);
    };

    const resetChart = () => {
        if (chartInstanceRef.current) {
            chartInstanceRef.current.timeScale().fitContent();
        }
    };

    const copyPrice = () => {
        if (clickedPrice) {
            navigator.clipboard.writeText(clickedPrice.toFixed(2));
        }
    };

    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <div
                    ref={chartRef}
                    style={{ width: "100%", height: "100%" }}
                    onClick={handleChartClick}
                />
            </ContextMenuTrigger>
            <ContextMenuContent className="w-48">
                <ContextMenuItem onClick={resetChart}>
                    Reset Chart
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem onClick={copyPrice} disabled={!clickedPrice}>
                    Copy Price{" "}
                    {clickedPrice ? `(${clickedPrice.toFixed(2)})` : ""}
                </ContextMenuItem>
            </ContextMenuContent>
        </ContextMenu>
    );
}

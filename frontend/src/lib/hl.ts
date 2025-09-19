import { Hyperliquid } from 'hyperliquid';
const sdk = new Hyperliquid({ enableWs: true });
await sdk.connect();


// const endTime = Date.now();
// const startTime = endTime - 60 * 60 * 1000;
// const candle = await sdk.info.getCandleSnapshot("BTC", "1m", startTime, endTime);
// console.log(candle);

sdk.subscriptions.subscribeToCandle("BTC", "1m", (candle) => {
    console.log(candle);
});

sdk.disconnect();
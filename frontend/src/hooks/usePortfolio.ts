import { useState, useEffect } from 'react';
import { GetPortfolioSummary, GetWalletAddress } from '@/../wailsjs/go/main/App';
import { hyperliquid, main } from "@/../wailsjs/go/models";

export function usePortfolio() {
    const [portfolio, setPortfolio] = useState<main.PortfolioSummary | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [address, setAddress] = useState<string>('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchPortfolio = async () => {
        if (!GetPortfolioSummary) {
            setError('Wallet functions not available');
            return;
        }

        try {
            setLoading(true);
            setError(null);

            const addr = await GetWalletAddress();
            setAddress(addr);

            const data = await GetPortfolioSummary();
            setPortfolio(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch portfolio');
            console.error('Portfolio fetch error:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchPortfolio();
        // Refresh every 10 seconds
        const interval = setInterval(fetchPortfolio, 10000);
        return () => clearInterval(interval);
    }, []);

    return {
        portfolio,
        isConnected,
        address,
        loading,
        error,
        refresh: fetchPortfolio,
    };
}

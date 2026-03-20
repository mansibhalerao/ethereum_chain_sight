import { useEffect, useState } from "react";
import { Routes, Route, useNavigate } from "react-router-dom";
import { ethers } from "ethers";

import BlockInfo from "./components/BlockInfo";
import TransactionInfo from "./components/TransactionInfo";

import { getLatestBlock, getBlockTransactions } from "./services/blockchainService";
import {
  getAddressSummary,
  getMostActive,
  getNetworkMetrics,
  getTopReceivers,
  getTopSenders,
} from "./services/insightsService";
import { isValidAddress } from "./services/walletService";

import WalletTest from "./pages/WalletTest";
import BlockPage from "./pages/BlockPage";
import "./App.css";

function shortAddress(address) {
  if (!address || address.length < 10) return address;
  return `${address.slice(0, 8)}...${address.slice(-6)}`;
}

function formatWeiToEth(weiValue) {
  try {
    const formatted = ethers.formatUnits(String(weiValue || "0"), 18);
    const asNumber = Number(formatted);
    if (!Number.isFinite(asNumber)) {
      return formatted;
    }
    return asNumber.toFixed(4).replace(/\.?0+$/, "");
  } catch {
    return "0";
  }
}

function formatWeiToGwei(weiValue) {
  try {
    const formatted = ethers.formatUnits(String(weiValue || "0"), 9);
    const asNumber = Number(formatted);
    if (!Number.isFinite(asNumber)) {
      return formatted;
    }
    return asNumber.toFixed(2).replace(/\.?0+$/, "");
  } catch {
    return "0";
  }
}

function Navigation() {
  return (
    <nav className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 shadow-lg text-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <span className="text-2xl font-bold text-white">ChainSight</span>
          </div>
        </div>
      </div>
    </nav>
  );
}

function Home() {
  const navigate = useNavigate();

  const [block, setBlock] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [error, setError] = useState(null);

  const [addressInput, setAddressInput] = useState("");
  const [addressSummary, setAddressSummary] = useState(null);
  const [addressError, setAddressError] = useState(null);
  const [addressLoading, setAddressLoading] = useState(false);

  const [topSenders, setTopSenders] = useState([]);
  const [topReceivers, setTopReceivers] = useState([]);
  const [mostActive, setMostActive] = useState([]);
  const [metrics, setMetrics] = useState([]);
  const [insightsError, setInsightsError] = useState(null);
  const [insightsLoading, setInsightsLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        setError(null);

        const latestBlock = await getLatestBlock();

        if (!latestBlock) {
          setError("Failed to fetch latest block.");
          return;
        }

        const transactionObjects = await getBlockTransactions(latestBlock, 5);

        setBlock(latestBlock);
        setTransactions(transactionObjects);
      } catch (err) {
        console.error("Failed to load block:", err);
        setError("Failed to load homepage data.");
      }
    };

    loadData();
  }, []);

  useEffect(() => {
    const loadInsights = async () => {
      try {
        setInsightsError(null);
        setInsightsLoading(true);

        const [senders, receivers, active, metricRows] = await Promise.all([
          getTopSenders(5),
          getTopReceivers(5),
          getMostActive(24, 5),
          getNetworkMetrics("minute", 24, 40),
        ]);

        setTopSenders(senders);
        setTopReceivers(receivers);
        setMostActive(active);
        setMetrics(metricRows);
      } catch (err) {
        console.error("Failed to load dashboard insights:", err);
        setInsightsError("Insights unavailable until Postgres indexer has data.");
      } finally {
        setInsightsLoading(false);
      }
    };

    loadInsights();
  }, []);

  const handleAddressLookup = async () => {
    const normalized = addressInput.trim();
    setAddressSummary(null);
    setAddressError(null);

    if (!isValidAddress(normalized)) {
      setAddressError("Enter a valid Ethereum address");
      return;
    }

    try {
      setAddressLoading(true);
      const summary = await getAddressSummary(normalized);
      setAddressSummary(summary);
    } catch (err) {
      console.error("Failed to load address summary:", err);
      setAddressError("Address not found in indexed DB yet.");
    } finally {
      setAddressLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50">
      <Navigation />
      
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Ethereum Blockchain Explorer</h1>
          <p className="text-lg text-gray-600">Real-time insights into the Ethereum network</p>
        </div>

        {error && (
          <div className="bg-red-50 border-l-4 border-red-500 p-4 mb-6 rounded-r-lg">
            <div className="flex items-center">
              <svg className="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
              <p className="text-red-700 font-medium">{error}</p>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-2xl font-bold text-gray-900">Latest Block</h2>
              <div className="flex items-center space-x-1 text-green-500">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span className="text-sm font-medium">Live</span>
              </div>
            </div>

            {block ? (
              <div>
                <button
                  onClick={() => navigate(`/block/${block.number}`)}
                  className="w-full mb-4 bg-gradient-to-r from-indigo-500 to-purple-600 text-white py-3 px-4 rounded-lg font-semibold hover:from-indigo-600 hover:to-purple-700 transition-all shadow-md hover:shadow-lg transform hover:-translate-y-0.5"
                >
                  View Block #{block.number}
                </button>
                <BlockInfo block={block} />
              </div>
            ) : (
              <div className="flex items-center justify-center py-12">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600"></div>
              </div>
            )}
          </div>

          <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Quick Actions</h2>
            <div className="space-y-3">
              <button
                onClick={() => navigate("/wallet")}
                className="w-full bg-gradient-to-r from-blue-500 to-cyan-600 text-white py-4 px-6 rounded-lg font-semibold hover:from-blue-600 hover:to-cyan-700 transition-all shadow-md hover:shadow-lg transform hover:-translate-y-0.5 flex items-center justify-between"
              >
                <span>Check Wallet Balance</span>
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
              <div className="bg-gradient-to-r from-emerald-50 to-teal-50 p-4 rounded-lg border border-emerald-200">
                <div className="flex items-start space-x-3">
                  <svg className="w-6 h-6 text-emerald-600 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                  </svg>
                  <div>
                    <h3 className="font-semibold text-emerald-900 mb-1">About ChainSight</h3>
                    <p className="text-sm text-emerald-700">Explore Ethereum blocks, transactions, and wallet balances in real-time with our intuitive blockchain explorer.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Address Profile Lookup</h2>
            <div className="flex gap-2 mb-4">
              <input
                type="text"
                value={addressInput}
                onChange={(e) => setAddressInput(e.target.value)}
                placeholder="0x..."
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg font-mono text-sm focus:ring-2 focus:ring-purple-600 focus:border-transparent"
              />
              <button
                onClick={handleAddressLookup}
                disabled={addressLoading || !addressInput.trim()}
                className="px-4 py-2 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-lg font-semibold disabled:opacity-50"
              >
                {addressLoading ? "Loading..." : "Lookup"}
              </button>
            </div>

            {addressError && <p className="text-sm text-red-600 mb-2">{addressError}</p>}

            {addressSummary && (
              <div className="space-y-2 text-sm">
                <div className="flex justify-between"><span className="text-gray-600">Address</span><span className="font-mono text-xs">{shortAddress(addressSummary.Address || addressSummary.address)}</span></div>
                <div className="flex justify-between"><span className="text-gray-600">Transactions</span><span className="font-semibold">{addressSummary.TxCount ?? addressSummary.txCount ?? 0}</span></div>
                <div className="flex justify-between"><span className="text-gray-600">Total Sent</span><span className="font-semibold">{formatWeiToEth(addressSummary.TotalSentWei || addressSummary.totalSentWei || addressSummary.total_sent_wei)} ETH</span></div>
                <div className="flex justify-between"><span className="text-gray-600">Total Received</span><span className="font-semibold">{formatWeiToEth(addressSummary.TotalReceivedWei || addressSummary.totalReceivedWei || addressSummary.total_received_wei)} ETH</span></div>
                <div className="flex justify-between"><span className="text-gray-600">First Seen Block</span><span className="font-semibold">{addressSummary.FirstSeenBlock ?? addressSummary.firstSeenBlock ?? 0}</span></div>
                <div className="flex justify-between"><span className="text-gray-600">Last Seen Block</span><span className="font-semibold">{addressSummary.LastSeenBlock ?? addressSummary.lastSeenBlock ?? 0}</span></div>
              </div>
            )}
          </div>

          <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Leaderboards</h2>
            {insightsError && <p className="text-sm text-amber-700 mb-3">{insightsError}</p>}
            {insightsLoading ? (
              <div className="flex items-center justify-center py-10">
                <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-purple-600"></div>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-sm">
                <div>
                  <h3 className="font-semibold text-gray-800 mb-3">Top Senders</h3>
                  <div className="space-y-2">
                    {topSenders.map((row, idx) => (
                      <div
                        key={`s-${row.address}-${idx}`}
                        className="p-2 rounded-lg bg-gray-50 border border-gray-100 flex flex-col"
                      >
                        <span className="font-mono text-xs break-all mb-1">{row.address}</span>
                        <span className="font-semibold text-xs text-gray-900">{formatWeiToEth(row.valueWei)} ETH</span>
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-800 mb-3">Top Receivers</h3>
                  <div className="space-y-2">
                    {topReceivers.map((row, idx) => (
                      <div
                        key={`r-${row.address}-${idx}`}
                        className="p-2 rounded-lg bg-gray-50 border border-gray-100 flex flex-col"
                      >
                        <span className="font-mono text-xs break-all mb-1">{row.address}</span>
                        <span className="font-semibold text-xs text-gray-900">{formatWeiToEth(row.valueWei)} ETH</span>
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-800 mb-3">Most Active in Last 24h</h3>
                  <div className="space-y-2">
                    {mostActive.map((row, idx) => (
                      <div
                        key={`a-${row.address}-${idx}`}
                        className="p-2 rounded-lg bg-gray-50 border border-gray-100 flex flex-col"
                      >
                        <span className="font-mono text-xs break-all mb-1">{row.address}</span>
                        <span className="font-semibold text-xs text-gray-900">{row.txCount} tx</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100 mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Network Metrics (Time-series)</h2>
          {metrics.length === 0 ? (
            <p className="text-sm text-gray-600">No metric buckets yet. Let indexer run for a few minutes.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-600 border-b">
                    <th className="py-2">Bucket</th>
                    <th className="py-2">Avg Gas Price</th>
                    <th className="py-2">Fees</th>
                    <th className="py-2">Tx Count</th>
                    <th className="py-2">Avg Block Time</th>
                  </tr>
                </thead>
                <tbody>
                  {metrics.slice(-12).reverse().map((m, idx) => (
                    <tr
                      key={`${m.BucketStart || m.bucketStart || m.bucket_start}-${idx}`}
                      className="border-b last:border-b-0"
                    >
                      <td className="py-2 text-gray-700">
                        {new Date(m.BucketStart || m.bucketStart || m.bucket_start).toLocaleTimeString()}
                      </td>
                      <td className="py-2 font-semibold">
                        {formatWeiToGwei(m.AvgGasPriceWei || m.avgGasPriceWei || m.avg_gas_price_wei)} Gwei
                      </td>
                      <td className="py-2 font-semibold">
                        {formatWeiToEth(m.TotalFeesWei || m.totalFeesWei || m.total_fees_wei)} ETH
                      </td>
                      <td className="py-2">{m.TxCount || m.txCount || m.tx_count}</td>
                      <td className="py-2">
                        {Number(m.AvgBlockTimeSec || m.avgBlockTimeSec || m.avg_block_time_sec || 0).toFixed(2)}s
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Latest Transactions</h2>
            <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">
              {transactions.length} transactions
            </span>
          </div>

          {transactions.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600"></div>
            </div>
          ) : (
            <div className="space-y-4">
              {transactions.map((tx) => (
                <TransactionInfo key={tx.hash} tx={tx} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/wallet" element={<WalletTest />} />
      <Route path="/block/:blockNumber" element={<BlockPage />} />
    </Routes>
  );
}

export default App;

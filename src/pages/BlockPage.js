import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getBlockByNumber, getBlockTransactions } from "../services/blockchainService";
import TransactionInfo from "../components/TransactionInfo";
import BlockInfo from "../components/BlockInfo";

function BlockPage() {
  const { blockNumber } = useParams();
  const navigate = useNavigate();

  const [block, setBlock] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadBlockData = async () => {
      try {
        setError(null);
        setBlock(null);
        setTransactions([]);

        const blockData = await getBlockByNumber(blockNumber);

        if (!blockData) {
          setError("Block not found or failed to fetch.");
          return;
        }

        setBlock(blockData);

        // Fetch full tx objects (limit 10 for safety)
        const txObjects = await getBlockTransactions(blockData, 10);
        setTransactions(txObjects);
      } catch (err) {
        console.error("Failed to load block page:", err);
        setError("Failed to load block data.");
      }
    };

    loadBlockData();
  }, [blockNumber]);

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-xl p-8 max-w-md border border-gray-100">
          <div className="flex flex-col items-center text-center">
            <svg className="w-16 h-16 text-red-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h2 className="text-2xl font-bold text-gray-900 mb-2">Error Loading Block</h2>
            <p className="text-gray-600 mb-6">{error}</p>
            <button
              onClick={() => navigate("/")}
              className="bg-gradient-to-r from-indigo-500 to-purple-600 text-white py-2 px-6 rounded-lg font-semibold hover:from-indigo-600 hover:to-purple-700 transition-all shadow-md hover:shadow-lg"
            >
              Go Home
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!block) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-16 w-16 border-b-4 border-purple-600 mx-auto mb-4"></div>
          <p className="text-xl text-gray-700 font-medium">Loading block #{blockNumber}...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-purple-50">
      <nav className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <button 
              onClick={() => navigate("/")}
              className="flex items-center space-x-2 text-white hover:text-gray-200 transition-colors"
            >
              <svg className="w-8 h-8" fill="currentColor" viewBox="0 0 20 20">
                <path d="M10 2a8 8 0 100 16 8 8 0 000-16zM9 9V5a1 1 0 012 0v4a1 1 0 01-2 0zm0 4a1 1 0 112 0 1 1 0 01-2 0z"/>
              </svg>
              <span className="text-2xl font-bold">ChainSight</span>
            </button>
            <button
              onClick={() => navigate("/")}
              className="flex items-center space-x-2 text-white hover:bg-white/20 px-4 py-2 rounded-lg transition-all"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              <span className="font-medium">Back to Home</span>
            </button>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Block Details</h1>
          <p className="text-gray-600">Detailed information for block #{blockNumber}</p>
        </div>

        <div className="grid grid-cols-1 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Block Information</h2>
            <BlockInfo block={block} />
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-xl p-6 border border-gray-100">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Transactions</h2>
            <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">
              {transactions.length} of {block.transactions.length} shown
            </span>
          </div>

          {transactions.length === 0 ? (
            <div className="text-center py-12">
              <svg className="w-16 h-16 text-gray-400 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p className="text-gray-600">No transactions to display</p>
            </div>
          ) : (
            <div className="space-y-4">
              {transactions.map((tx, index) => (
                <TransactionInfo key={index} tx={tx} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default BlockPage;

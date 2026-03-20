import { ethers } from "ethers";

function TransactionInfo({ tx }) {
  const value = parseFloat(ethers.formatEther(tx.value));
  const isContractCreation = !tx.to;

  return (
    <div className="bg-gradient-to-r from-slate-50 to-gray-50 p-4 rounded-lg border border-gray-200 hover:shadow-md transition-all hover:border-indigo-300">
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <div className="flex items-center space-x-2 mb-2">
            <svg className="w-4 h-4 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
              <path d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z"/>
              <path fillRule="evenodd" d="M4 5a2 2 0 012-2 3 3 0 003 3h2a3 3 0 003-3 2 2 0 012 2v11a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm3 4a1 1 0 000 2h.01a1 1 0 100-2H7zm3 0a1 1 0 000 2h3a1 1 0 100-2h-3zm-3 4a1 1 0 100 2h.01a1 1 0 100-2H7zm3 0a1 1 0 100 2h3a1 1 0 100-2h-3z" clipRule="evenodd"/>
            </svg>
            <span className="text-xs font-mono text-gray-600 bg-white px-2 py-0.5 rounded border border-gray-200 break-all">
              {tx.hash}
            </span>
          </div>
        </div>
        <div className={`px-3 py-1 rounded-full text-xs font-semibold ${
          value > 0 
            ? 'bg-green-100 text-green-700' 
            : 'bg-gray-100 text-gray-600'
        }`}>
          {value > 0 ? `${value.toFixed(4)} ETH` : '0 ETH'}
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex items-center space-x-2 text-sm">
          <span className="text-gray-500 w-12">From:</span>
          <span className="font-mono text-gray-900 bg-white px-2 py-0.5 rounded text-xs border border-gray-200 break-all">
            {tx.from}
          </span>
        </div>
        <div className="flex items-center space-x-2 text-sm">
          <span className="text-gray-500 w-12">To:</span>
          {isContractCreation ? (
            <span className="px-2 py-0.5 bg-blue-100 text-blue-700 rounded text-xs font-medium">
              Contract Creation
            </span>
          ) : (
            <span className="font-mono text-gray-900 bg-white px-2 py-0.5 rounded text-xs border border-gray-200 break-all">
              {tx.to}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

export default TransactionInfo;

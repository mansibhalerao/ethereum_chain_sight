import { Link } from "react-router-dom";

function BlockInfo({ block }) {
  return (
    <div className="space-y-3">
      <div className="bg-gradient-to-r from-slate-50 to-gray-50 p-4 rounded-lg border border-gray-200">
        <div className="flex justify-between items-center">
          <span className="text-sm font-medium text-gray-600">Block Number</span>
          <Link 
            to={`/block/${block.number}`}
            className="text-lg font-bold text-indigo-600 hover:text-indigo-800 transition-colors"
          >
            #{block.number}
          </Link>
        </div>
      </div>

      <div className="bg-gradient-to-r from-slate-50 to-gray-50 p-4 rounded-lg border border-gray-200">
        <div className="space-y-2">
          <div className="flex justify-between items-start">
            <span className="text-sm font-medium text-gray-600">Block Hash</span>
            <span className="text-sm font-mono text-gray-900 bg-white px-2 py-1 rounded border border-gray-200 break-all">
              {block.hash}
            </span>
          </div>
          <div className="flex justify-between items-start">
            <span className="text-sm font-medium text-gray-600">Parent Hash</span>
            <span className="text-sm font-mono text-gray-900 bg-white px-2 py-1 rounded border border-gray-200 break-all">
              {block.parentHash}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="bg-gradient-to-br from-blue-50 to-cyan-50 p-4 rounded-lg border border-blue-200">
          <div className="text-sm font-medium text-blue-600 mb-1">Timestamp</div>
          <div className="text-xs text-blue-900">
            {new Date(block.timestamp * 1000).toLocaleString()}
          </div>
        </div>
        <div className="bg-gradient-to-br from-purple-50 to-pink-50 p-4 rounded-lg border border-purple-200">
          <div className="text-sm font-medium text-purple-600 mb-1">Transactions</div>
          <div className="text-2xl font-bold text-purple-900">
            {block.transactions.length}
          </div>
        </div>
      </div>
    </div>
  );
}

export default BlockInfo;
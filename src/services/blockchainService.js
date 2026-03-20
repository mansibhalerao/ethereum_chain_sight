const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080";

async function apiGet(path) {
  const response = await fetch(`${API_BASE_URL}${path}`);

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`API error (${response.status}): ${errorBody}`);
  }

  return response.json();
}

/**
 * Get latest block
 */
export async function getLatestBlock() {
  try {
    console.log("Fetching latest block");
    return await apiGet("/api/blocks/latest");
  } catch (error) {
    console.error("Error fetching latest block:", error);
    throw error;
  }
}

/**
 * Get block by number
 */
export async function getBlockByNumber(blockNumber) {
  try {
    const blockNum = Number(blockNumber);

    if (!Number.isInteger(blockNum) || blockNum < 0) {
      throw new Error(`Invalid block number: ${blockNumber}`);
    }
    
    console.log("Fetching block:", blockNum);
    return await apiGet(`/api/blocks/${blockNum}`);
  } catch (error) {
    console.error("Error fetching block by number:", error);
    throw error;
  }
}

/**
 * Get transactions for a block (full tx objects)
 */
export async function getBlockTransactions(block, limit = 5) {
  try {
    console.log("Fetching transactions for block:", block.number);
    const result = await apiGet(`/api/blocks/${block.number}/transactions?limit=${limit}`);
    return result.transactions || [];
  } catch (error) {
    console.error("Error fetching block transactions:", error);
    throw error;  
  }
}
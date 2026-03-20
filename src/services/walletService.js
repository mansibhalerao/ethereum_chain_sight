import { ethers } from "ethers";

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080";

/**
 * Validate Ethereum address
 */
export function isValidAddress(address) {
  return ethers.isAddress(address);
}

/**
 * Get ETH balance for address
 */
export async function getWalletBalance(address) {
  try {
    const response = await fetch(`${API_BASE_URL}/api/wallets/${address}/balance`);

    if (!response.ok) {
      return null;
    }

    const data = await response.json();
    return ethers.formatEther(data.balanceWei);
  } catch (err) {
    console.error("Failed to fetch wallet balance", err);
    return null;
  }
}

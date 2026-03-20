const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || "http://localhost:8080";

async function apiGet(path) {
  const response = await fetch(`${API_BASE_URL}${path}`);

  if (!response.ok) {
    const errorBody = await response.text();
    throw new Error(`API error (${response.status}): ${errorBody}`);
  }

  return response.json();
}

export async function getAddressSummary(address) {
  return apiGet(`/api/addresses/${address}/summary`);
}

export async function getTopSenders(limit = 5) {
  const result = await apiGet(`/api/leaderboards/top-senders?limit=${limit}`);
  return result.results || [];
}

export async function getTopReceivers(limit = 5) {
  const result = await apiGet(`/api/leaderboards/top-receivers?limit=${limit}`);
  return result.results || [];
}

export async function getMostActive(hours = 24, limit = 5) {
  const result = await apiGet(`/api/leaderboards/most-active?hours=${hours}&limit=${limit}`);
  return result.results || [];
}

export async function getNetworkMetrics(granularity = "minute", hours = 24, limit = 60) {
  const result = await apiGet(
    `/api/analytics/metrics?granularity=${granularity}&hours=${hours}&limit=${limit}`
  );
  return result.results || [];
}

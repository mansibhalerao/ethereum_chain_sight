package httpapi

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mansi/chainsight/server/internal/eth"
	"github.com/mansi/chainsight/server/internal/store"
)

type Server struct {
	ethClient *eth.Client
	repo      *store.Repository
	dbEnabled bool
}

func NewServer(c *eth.Client, repo *store.Repository, dbEnabled bool) *Server {
	return &Server{ethClient: c, repo: repo, dbEnabled: dbEnabled}
}

// returns 503 when analytics is disabled
func (s *Server) requireAnalytics(w http.ResponseWriter) bool {
	if !s.dbEnabled || s.repo == nil {
		writeError(w, http.StatusServiceUnavailable,
			"analytics disabled: set DB_ENABLED=true (and INDEXER_ENABLED=true for fresh analytics data)", nil)
		return false
	}
	return true
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)

	// live RPC endpoints (always enabled)
	mux.HandleFunc("GET /api/blocks/latest", s.handleLatest)
	mux.HandleFunc("GET /api/blocks/", s.handleBlocks)
	mux.HandleFunc("GET /api/wallets/", s.handleWallets)

	// analytics endpoints (feature-flagged, require DB)
	mux.HandleFunc("GET /api/addresses/", s.handleAddresses)
	mux.HandleFunc("GET /api/leaderboards/top-senders", s.handleTopSenders)
	mux.HandleFunc("GET /api/leaderboards/top-receivers", s.handleTopReceivers)
	mux.HandleFunc("GET /api/leaderboards/most-active", s.handleMostActive)
	mux.HandleFunc("GET /api/analytics/metrics", s.handleMetrics)
	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleLatest(w http.ResponseWriter, r *http.Request) {
	block, err := s.ethClient.GetLatestBlock(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch latest block", err)
		return
	}
	writeJSON(w, http.StatusOK, block)
}

func (s *Server) handleBlocks(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/blocks/"), "/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "block number is required", nil)
		return
	}

	if strings.HasSuffix(path, "/transactions") {
		s.handleBlockTransactions(w, r, path)
		return
	}

	number, err := strconv.ParseUint(path, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block number", err)
		return
	}

	block, err := s.ethClient.GetBlockByNumber(r.Context(), number)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch block", err)
		return
	}

	writeJSON(w, http.StatusOK, block)
}

func (s *Server) handleBlockTransactions(w http.ResponseWriter, r *http.Request, path string) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "transactions" {
		writeError(w, http.StatusNotFound, "route not found", nil)
		return
	}

	number, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block number", err)
		return
	}

	limit := 5
	if qLimit := r.URL.Query().Get("limit"); qLimit != "" {
		parsed, parseErr := strconv.Atoi(qLimit)
		if parseErr != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "invalid limit query param", parseErr)
			return
		}
		limit = parsed
	}

	txs, err := s.ethClient.GetBlockTransactions(r.Context(), number, limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch block transactions", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"blockNumber":    number,
		"limit":          limit,
		"transactions":   txs,
		"transactionLen": len(txs),
	})
}

func (s *Server) handleWallets(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/wallets/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "balance" {
		writeError(w, http.StatusNotFound, "route not found", nil)
		return
	}

	address := parts[0]
	if !isValidAddress(address) {
		writeError(w, http.StatusBadRequest, "invalid ethereum address", nil)
		return
	}

	balanceWei, err := s.ethClient.GetWalletBalance(r.Context(), address)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch wallet balance", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"address":    strings.ToLower(address),
		"balanceWei": balanceWei,
	})
}

func (s *Server) handleAddresses(w http.ResponseWriter, r *http.Request) {
	if !s.requireAnalytics(w) {
		return
	}
	if s.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured", nil)
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/addresses/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "summary" {
		writeError(w, http.StatusNotFound, "route not found", nil)
		return
	}

	address := parts[0]
	if !isValidAddress(address) {
		writeError(w, http.StatusBadRequest, "invalid ethereum address", nil)
		return
	}

	summary, err := s.repo.GetAddressSummary(address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch address summary", err)
		return
	}
	if summary == nil {
		writeError(w, http.StatusNotFound, "address summary not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTopSenders(w http.ResponseWriter, r *http.Request) {
	if !s.requireAnalytics(w) {
		return
	}
	if s.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured", nil)
		return
	}

	limit := queryInt(r, "limit", 10)
	rows, err := s.repo.GetTopSenders(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch top senders", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"limit":   limit,
		"results": rows,
	})
}

func (s *Server) handleTopReceivers(w http.ResponseWriter, r *http.Request) {
	if !s.requireAnalytics(w) {
		return
	}
	if s.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured", nil)
		return
	}

	limit := queryInt(r, "limit", 10)
	rows, err := s.repo.GetTopReceivers(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch top receivers", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"limit":   limit,
		"results": rows,
	})
}

func (s *Server) handleMostActive(w http.ResponseWriter, r *http.Request) {
	if !s.requireAnalytics(w) {
		return
	}
	if s.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured", nil)
		return
	}

	limit := queryInt(r, "limit", 10)
	hours := queryInt(r, "hours", 24)
	if hours <= 0 {
		hours = 24
	}

	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	rows, err := s.repo.GetMostActiveAddressesSince(since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch most active addresses", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"limit":   limit,
		"hours":   hours,
		"results": rows,
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if !s.requireAnalytics(w) {
		return
	}
	if s.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database is not configured", nil)
		return
	}

	granularity := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("granularity")))
	if granularity == "" {
		granularity = "minute"
	}

	hours := queryInt(r, "hours", 24)
	if hours <= 0 {
		hours = 24
	}

	limit := queryInt(r, "limit", 200)
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	rows, err := s.repo.GetNetworkMetrics(granularity, since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch network metrics", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"granularity": granularity,
		"hours":       hours,
		"limit":       limit,
		"results":     rows,
	})
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return parsed
}

func isValidAddress(address string) bool {
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return false
	}
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]string{"error": message}
	if err != nil {
		response["details"] = err.Error()
	}
	writeJSON(w, status, response)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

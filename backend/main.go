package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// WorkloadStatus represents the attestation status of a CoCo workload
type WorkloadStatus struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Attested          bool      `json:"attested"`
	AttestationStatus string    `json:"attestation_status"`
	Timestamp         string    `json:"timestamp"`
	Details           string    `json:"details"`
	GateOneStatus     string    `json:"gate_one_status"`  // Code Integrity
	GateTwoStatus     string    `json:"gate_two_status"`  // TEE Attestation
	LastChecked       time.Time `json:"last_checked"`
	TEEType           string    `json:"tee_type,omitempty"`
}

// DashboardResponse is the API response for the dashboard
type DashboardResponse struct {
	OverallStatus string           `json:"overall_status"` // "compliant" or "violation"
	Workloads     []WorkloadStatus `json:"workloads"`
	LastUpdated   time.Time        `json:"last_updated"`
}

// TrustVector represents EAR trust tier values from Collector
type TrustVector struct {
	InstanceIdentity int `json:"instance_identity"`
	Configuration    int `json:"configuration"`
	Executables      int `json:"executables"`
	FileSystem       int `json:"file_system"`
	Hardware         int `json:"hardware"`
	RuntimeOpaque    int `json:"runtime_opaque"`
	StorageOpaque    int `json:"storage_opaque"`
	SourcedData      int `json:"sourced_data"`
}

// CollectorReport matches the Attestation Collector's report format
type CollectorReport struct {
	PodName     string       `json:"pod_name"`
	Namespace   string       `json:"namespace"`
	TEEType     string       `json:"tee_type,omitempty"`
	Attested    bool         `json:"attested"`
	TrustVector *TrustVector `json:"trust_vector,omitempty"`
	EARToken    string       `json:"ear_token,omitempty"`
	Timestamp   time.Time    `json:"timestamp"`
	Error       string       `json:"error,omitempty"`
}

// Server holds the dashboard backend state
type Server struct {
	collectorURL string
	statusCache  map[string]*WorkloadStatus
	cacheMutex   sync.RWMutex
	httpClient   *http.Client
	pollInterval time.Duration
}

func main() {
	log.Println("Starting Hospital Dashboard Backend...")

	// Load configuration - get Collector URL from environment
	collectorURL := getEnv("COLLECTOR_URL", "http://attestation-collector:8080")

	server := &Server{
		collectorURL: collectorURL,
		statusCache:  make(map[string]*WorkloadStatus),
		pollInterval: 30 * time.Second,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	log.Printf("Configured to fetch from Attestation Collector: %s", collectorURL)

	// Start background polling from Collector
	go server.pollCollector()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/status", server.handleStatus)
	mux.HandleFunc("/api/workloads", server.handleWorkloads)
	mux.HandleFunc("/api/workload/", server.handleWorkloadDetail)

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Serve static files (frontend)
	fs := http.FileServer(http.Dir("/app/static"))
	mux.Handle("/", fs)

	port := getEnv("PORT", "8080")
	log.Printf("Dashboard backend listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, loggingMiddleware(corsMiddleware(mux))))
}

// handleStatus returns the overall dashboard status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	response := DashboardResponse{
		OverallStatus: "compliant",
		Workloads:     make([]WorkloadStatus, 0, len(s.statusCache)),
		LastUpdated:   time.Now(),
	}

	for _, status := range s.statusCache {
		response.Workloads = append(response.Workloads, *status)
		if !status.Attested || status.GateTwoStatus == "failed" {
			response.OverallStatus = "violation"
		}
	}

	// If no workloads configured, return demo data
	if len(response.Workloads) == 0 {
		response = getDemoResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleWorkloads returns all workload statuses
func (s *Server) handleWorkloads(w http.ResponseWriter, r *http.Request) {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	workloads := make([]WorkloadStatus, 0, len(s.statusCache))
	for _, status := range s.statusCache {
		workloads = append(workloads, *status)
	}

	// If no workloads configured, return demo data
	if len(workloads) == 0 {
		workloads = getDemoResponse().Workloads
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workloads)
}

// handleWorkloadDetail returns details for a specific workload
func (s *Server) handleWorkloadDetail(w http.ResponseWriter, r *http.Request) {
	// Extract workload name from path: /api/workload/{name}
	name := r.URL.Path[len("/api/workload/"):]
	if name == "" {
		http.Error(w, "workload name required", http.StatusBadRequest)
		return
	}

	s.cacheMutex.RLock()
	status, exists := s.statusCache[name]
	s.cacheMutex.RUnlock()

	if !exists {
		http.Error(w, "workload not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// pollCollector periodically fetches attestation reports from the Collector
func (s *Server) pollCollector() {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	// Initial fetch
	s.fetchFromCollector()

	for range ticker.C {
		s.fetchFromCollector()
	}
}

// fetchFromCollector fetches all attestation reports from the Collector API
func (s *Server) fetchFromCollector() {
	url := fmt.Sprintf("%s/api/v1/reports", s.collectorURL)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		log.Printf("Failed to fetch from Collector: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Collector returned status %d", resp.StatusCode)
		return
	}

	var reports []CollectorReport
	if err := json.NewDecoder(resp.Body).Decode(&reports); err != nil {
		log.Printf("Failed to decode Collector response: %v", err)
		return
	}

	log.Printf("Fetched %d reports from Collector", len(reports))

	// Convert Collector reports to WorkloadStatus and update cache
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Clear old cache and repopulate
	s.statusCache = make(map[string]*WorkloadStatus)

	for _, report := range reports {
		status := s.convertCollectorReport(report)
		key := report.Namespace + "/" + report.PodName
		s.statusCache[key] = status
	}
}

// convertCollectorReport converts a Collector report to WorkloadStatus
func (s *Server) convertCollectorReport(report CollectorReport) *WorkloadStatus {
	status := &WorkloadStatus{
		Name:        report.PodName,
		Namespace:   report.Namespace,
		Attested:    report.Attested,
		Timestamp:   report.Timestamp.Format(time.RFC3339),
		LastChecked: time.Now(),
		TEEType:     report.TEEType,
	}

	// Determine attestation status and details
	if report.Attested {
		status.AttestationStatus = "verified"
		status.GateOneStatus = "passing"
		status.GateTwoStatus = "passing"

		// Build details from trust vector
		if report.TrustVector != nil {
			status.Details = fmt.Sprintf("TEE attestation successful (%s) - Hardware: %s, Config: %s, Executables: %s",
				report.TEEType,
				trustTierToString(report.TrustVector.Hardware),
				trustTierToString(report.TrustVector.Configuration),
				trustTierToString(report.TrustVector.Executables))
		} else {
			status.Details = fmt.Sprintf("TEE attestation successful (%s)", report.TEEType)
		}
	} else {
		status.AttestationStatus = "failed"
		status.GateOneStatus = "passing" // Assume code integrity passes if pod exists
		status.GateTwoStatus = "failed"

		if report.Error != "" {
			status.Details = report.Error
		} else {
			status.Details = "TEE attestation failed - not running in genuine confidential environment"
		}
	}

	return status
}

// trustTierToString converts EAR trust tier value to human-readable string
func trustTierToString(tier int) string {
	switch tier {
	case 0:
		return "None"
	case 2:
		return "Affirming"
	case 32:
		return "Warning"
	case 96:
		return "Contraindicated"
	default:
		return fmt.Sprintf("Unknown(%d)", tier)
	}
}

// getDemoResponse returns demo data when no real workloads are configured
func getDemoResponse() DashboardResponse {
	return DashboardResponse{
		OverallStatus: "compliant",
		Workloads: []WorkloadStatus{
			{
				Name:              "janine-ai-model-v1.3",
				Namespace:         "janine-dev",
				Attested:          true,
				AttestationStatus: "verified",
				Timestamp:         time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
				Details:           "TEE attestation successful",
				GateOneStatus:     "passing",
				GateTwoStatus:     "passing",
				LastChecked:       time.Now(),
			},
			{
				Name:              "database-backup-service",
				Namespace:         "janine-dev",
				Attested:          true,
				AttestationStatus: "verified",
				Timestamp:         time.Now().Add(-45 * time.Minute).Format(time.RFC3339),
				Details:           "Container signature verified, TEE attestation passed",
				GateOneStatus:     "passing",
				GateTwoStatus:     "passing",
				LastChecked:       time.Now(),
			},
		},
		LastUpdated: time.Now(),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

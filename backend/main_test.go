package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFetchFromCollector tests that the backend fetches data from the Collector API
func TestFetchFromCollector(t *testing.T) {
	// Create a mock Collector server
	mockCollector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/reports" {
			t.Errorf("Expected path /api/v1/reports, got %s", r.URL.Path)
		}

		// Return mock attestation reports
		reports := []CollectorReport{
			{
				PodName:   "janine-hospital-coco-abc123",
				Namespace: "janine-app",
				TEEType:   "tdx",
				Attested:  true,
				TrustVector: &TrustVector{
					Hardware:      2, // AFFIRMING
					Configuration: 2,
					Executables:   2,
				},
				Timestamp: time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reports)
	}))
	defer mockCollector.Close()

	// Create server with mock Collector URL
	server := &Server{
		collectorURL: mockCollector.URL,
		statusCache:  make(map[string]*WorkloadStatus),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	// Fetch from Collector
	server.fetchFromCollector()

	// Verify cache was populated
	if len(server.statusCache) != 1 {
		t.Errorf("Expected 1 workload in cache, got %d", len(server.statusCache))
	}

	// Verify the workload status
	status, exists := server.statusCache["janine-app/janine-hospital-coco-abc123"]
	if !exists {
		t.Fatal("Expected workload janine-app/janine-hospital-coco-abc123 in cache")
	}

	if !status.Attested {
		t.Error("Expected Attested to be true")
	}

	if status.AttestationStatus != "verified" {
		t.Errorf("Expected AttestationStatus 'verified', got '%s'", status.AttestationStatus)
	}

	if status.GateTwoStatus != "passing" {
		t.Errorf("Expected GateTwoStatus 'passing', got '%s'", status.GateTwoStatus)
	}

	if status.TEEType != "tdx" {
		t.Errorf("Expected TEEType 'tdx', got '%s'", status.TEEType)
	}
}

// TestConvertCollectorReportAttested tests conversion of an attested report
func TestConvertCollectorReportAttested(t *testing.T) {
	server := &Server{}

	report := CollectorReport{
		PodName:   "test-pod",
		Namespace: "test-ns",
		TEEType:   "tdx",
		Attested:  true,
		TrustVector: &TrustVector{
			Hardware:      2,
			Configuration: 2,
			Executables:   2,
		},
		Timestamp: time.Now(),
	}

	status := server.convertCollectorReport(report)

	if status.Name != "test-pod" {
		t.Errorf("Expected Name 'test-pod', got '%s'", status.Name)
	}

	if status.Namespace != "test-ns" {
		t.Errorf("Expected Namespace 'test-ns', got '%s'", status.Namespace)
	}

	if !status.Attested {
		t.Error("Expected Attested to be true")
	}

	if status.AttestationStatus != "verified" {
		t.Errorf("Expected AttestationStatus 'verified', got '%s'", status.AttestationStatus)
	}

	if status.GateOneStatus != "passing" {
		t.Errorf("Expected GateOneStatus 'passing', got '%s'", status.GateOneStatus)
	}

	if status.GateTwoStatus != "passing" {
		t.Errorf("Expected GateTwoStatus 'passing', got '%s'", status.GateTwoStatus)
	}
}

// TestConvertCollectorReportFailed tests conversion of a failed attestation report
func TestConvertCollectorReportFailed(t *testing.T) {
	server := &Server{}

	report := CollectorReport{
		PodName:   "tampered-pod",
		Namespace: "test-ns",
		TEEType:   "tdx",
		Attested:  false,
		Error:     "CDH unreachable: connection refused",
		Timestamp: time.Now(),
	}

	status := server.convertCollectorReport(report)

	if status.Attested {
		t.Error("Expected Attested to be false")
	}

	if status.AttestationStatus != "failed" {
		t.Errorf("Expected AttestationStatus 'failed', got '%s'", status.AttestationStatus)
	}

	if status.GateTwoStatus != "failed" {
		t.Errorf("Expected GateTwoStatus 'failed', got '%s'", status.GateTwoStatus)
	}

	if status.Details != "CDH unreachable: connection refused" {
		t.Errorf("Expected error in Details, got '%s'", status.Details)
	}
}

// TestTrustTierToString tests trust tier value conversion
func TestTrustTierToString(t *testing.T) {
	tests := []struct {
		tier     int
		expected string
	}{
		{0, "None"},
		{2, "Affirming"},
		{32, "Warning"},
		{96, "Contraindicated"},
		{99, "Unknown(99)"},
	}

	for _, test := range tests {
		result := trustTierToString(test.tier)
		if result != test.expected {
			t.Errorf("trustTierToString(%d) = '%s', expected '%s'", test.tier, result, test.expected)
		}
	}
}

// TestHandleStatusReturnsLiveData tests that /api/status returns live data from Collector
func TestHandleStatusReturnsLiveData(t *testing.T) {
	// Create a mock Collector server
	mockCollector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reports := []CollectorReport{
			{
				PodName:   "janine-hospital-coco-xyz",
				Namespace: "janine-app",
				TEEType:   "tdx",
				Attested:  true,
				TrustVector: &TrustVector{
					Hardware:      2,
					Configuration: 2,
					Executables:   2,
				},
				Timestamp: time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reports)
	}))
	defer mockCollector.Close()

	// Create server
	server := &Server{
		collectorURL: mockCollector.URL,
		statusCache:  make(map[string]*WorkloadStatus),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}

	// Fetch data first
	server.fetchFromCollector()

	// Create request
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.handleStatus(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response DashboardResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.OverallStatus != "compliant" {
		t.Errorf("Expected OverallStatus 'compliant', got '%s'", response.OverallStatus)
	}

	if len(response.Workloads) != 1 {
		t.Errorf("Expected 1 workload, got %d", len(response.Workloads))
	}

	if response.Workloads[0].Name != "janine-hospital-coco-xyz" {
		t.Errorf("Expected workload name 'janine-hospital-coco-xyz', got '%s'", response.Workloads[0].Name)
	}
}

// TestServerHasCollectorURL tests that Server struct has collectorURL field
func TestServerHasCollectorURL(t *testing.T) {
	server := &Server{
		collectorURL: "http://attestation-collector:8080",
	}

	if server.collectorURL != "http://attestation-collector:8080" {
		t.Errorf("Expected collectorURL, got '%s'", server.collectorURL)
	}
}

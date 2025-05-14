package acceptance

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benjamin/file-storage-go/cmd/server"
	"github.com/benjamin/file-storage-go/pkg/adapters/metrics"
	"github.com/benjamin/file-storage-go/pkg/adapters/repository"
	"github.com/benjamin/file-storage-go/pkg/adapters/storage"
	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/benjamin/file-storage-go/pkg/domain"
	"github.com/gin-gonic/gin"
)

var (
	testServerListener net.Listener
)

// TestMain will setup the server and run tests
func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Attempt to load config
	cfg, configErr := config.LoadConfig()
	// We will proceed even if configErr is not nil, and decide on storage type later.

	// Determine if we should use mock storage
	useMockStorage := false
	if os.Getenv("USE_MOCK_STORAGE") == "true" {
		useMockStorage = true
		fmt.Println("INFO: USE_MOCK_STORAGE environment variable is set to true. Using MockStorage.")
	} else if configErr != nil {
		useMockStorage = true
		fmt.Printf("INFO: Failed to load full config (%v). Defaulting to MockStorage.\n", configErr)
		// If configErr is because some essential non-storage config is missing, tests might still fail.
		// For now, we assume the primary reason for configErr in test envs is missing storage creds.
		// If cfg is nil due to error, create a minimal one for NewPrometheusMetrics if it needs any cfg.
		if cfg == nil {
			cfg = &config.Config{} // Provide a default empty config
		}
	}

	// If API_BASE_URL is set, try to parse port from it for the test server.
	// Otherwise, use a dynamic port.
	var explicitPort string
	envApiBaseURL := getEnv("API_BASE_URL", "") // Use getEnv from file_lifecycle_test.go
	if envApiBaseURL != "" {
		parsedURL, err := url.Parse(envApiBaseURL)
		if err == nil && parsedURL.Port() != "" {
			explicitPort = parsedURL.Port()
			fmt.Printf("INFO: Will attempt to use explicit port %s from API_BASE_URL for test server listener.\n", explicitPort)
		}
	}

	// Initialize dependencies
	metricsCollector := metrics.NewPrometheusMetrics()

	var fileStorageService domain.FileStorage
	if useMockStorage {
		fmt.Println("INFO: Initializing MockStorage.")
		fileStorageService = storage.NewMockStorage()
	} else {
		fmt.Println("INFO: Initializing AzureBlobStorage. Ensure Azure environment variables are set.")
		var err error
		if cfg.BlobStorageURL == "" { // Add a check for safety, though LoadConfig should handle it
			log.Fatalf("AzureBlobStorage requires BLOB_STORAGE_URL. It's missing and not using mock.")
		}
		fileStorageService, err = storage.NewAzureBlobStorage(
			cfg.BlobStorageURL,
			cfg.StorageKey,
			cfg.ContainerName,
			metricsCollector,
		)
		if err != nil {
			log.Fatalf("Failed to initialize AzureBlobStorage: %v. If you intended to use mock storage, set USE_MOCK_STORAGE=true or ensure Azure config env vars are missing.", err)
		}
	}

	jobRepo := repository.NewInMemoryRepository()

	// Setup router using the new centralized function
	r := server.SetupRouter(fileStorageService, jobRepo)

	// Start the server on a dynamic port or explicit port
	var serverAddr string
	if explicitPort != "" {
		serverAddr = ":" + explicitPort
	} else {
		serverAddr = "localhost:0" // OS picks a free port
	}

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", serverAddr, err)
	}
	testServerListener = listener // Store listener to get the address

	// Update global apiBaseURL from file_lifecycle_test.go
	// Ensure this variable is accessible or re-declared in this package if needed
	// For now, assuming apiBaseURL (from file_lifecycle_test.go) will be updated by the test runner environment
	// or we update it directly here.
	chosenPort := testServerListener.Addr().(*net.TCPAddr).Port
	apiBaseURL = fmt.Sprintf("http://localhost:%d", chosenPort)
	fmt.Printf("INFO: Test server starting on %s\n", apiBaseURL)

	defaultTestToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpc3MiOiJ0ZXN0LWlzc3VlciIsImV4cCI6MTc0NTA4MTYwMCwiaWF0IjoxNzE3NzQ0MDAwfQ.placeholder_sig_for_testing_only"
	authToken = getEnv("API_AUTH_TOKEN", defaultTestToken)
	if authToken == defaultTestToken && os.Getenv("API_AUTH_TOKEN") == "" { // Check if it's the default because env var was empty
		fmt.Printf("INFO: API_AUTH_TOKEN environment variable not set. Using a default placeholder test token: %s\n", authToken)
		fmt.Println("INFO: If your API validates tokens, ensure this default is recognized by your test auth setup, or set a valid API_AUTH_TOKEN via environment variable.")
	} else {
		fmt.Println("INFO: Using API_AUTH_TOKEN from environment variable.")
	}

	srv := &http.Server{Handler: r}

	go func() {
		if err := srv.Serve(testServerListener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Test server ListenAndServe error: %v", err)
		}
	}()

	// Wait a moment for the server to start
	// A more robust way would be to ping a health check endpoint.
	time.Sleep(100 * time.Millisecond)

	// Create dummy test file (logic moved from file_lifecycle_test.go's TestMain)
	// Ensure testfileName and testfileContent are defined (they are const in file_lifecycle_test.go)
	// For simplicity, assume they are accessible or redeclare/pass them.
	// We'll create the test file in the current working directory of the test.
	wd, _ := os.Getwd()
	testFilePath := filepath.Join(wd, testfileName)

	err = os.WriteFile(testFilePath, []byte(testfileContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file '%s': %v", testFilePath, err)
	}

	// Run the tests
	exitCode := m.Run()

	// Cleanup: Stop the server
	fmt.Println("INFO: Shutting down test server...")
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Test server graceful shutdown failed: %v", err)
	} else {
		fmt.Println("INFO: Test server shut down gracefully.")
	}

	// Cleanup dummy test file
	os.Remove(testFilePath)

	os.Exit(exitCode)
}

// getEnv helper from file_lifecycle_test.go - ensure it's defined if not already.
// func getEnv(key, fallback string) string { ... }
// This will be defined in file_lifecycle_test.go and accessible within the package.

// url.Parse import
// import "net/url" // Removed from here

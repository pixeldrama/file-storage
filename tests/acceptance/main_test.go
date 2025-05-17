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
	"github.com/benjamin/file-storage-go/pkg/adapters/repository"
	"github.com/benjamin/file-storage-go/pkg/adapters/storage"
	"github.com/benjamin/file-storage-go/pkg/config"
	"github.com/gin-gonic/gin"
)

var (
	testServerListener net.Listener
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	if cfg == nil {
		cfg = &config.Config{}
	}

	var explicitPort string
	envApiBaseURL := getEnv("API_BASE_URL", "")
	if envApiBaseURL != "" {
		parsedURL, err := url.Parse(envApiBaseURL)
		if err == nil && parsedURL.Port() != "" {
			explicitPort = parsedURL.Port()
			fmt.Printf("INFO: Will attempt to use explicit port %s from API_BASE_URL for test server listener.\n", explicitPort)
		}
	}

	fileStorageService := storage.NewMockStorage()
	jobRepo := repository.NewInMemoryRepository()

	r := server.SetupRouter(fileStorageService, jobRepo)

	var serverAddr string
	if explicitPort != "" {
		serverAddr = ":" + explicitPort
	} else {
		serverAddr = "localhost:0"
	}

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", serverAddr, err)
	}
	testServerListener = listener

	chosenPort := testServerListener.Addr().(*net.TCPAddr).Port
	apiBaseURL = fmt.Sprintf("http://localhost:%d", chosenPort)
	fmt.Printf("INFO: Test server starting on %s\n", apiBaseURL)

	defaultTestToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJpc3MiOiJ0ZXN0LWlzc3VlciIsImV4cCI6MTc0NTA4MTYwMCwiaWF0IjoxNzE3NzQ0MDAwfQ.placeholder_sig_for_testing_only"
	authToken = getEnv("API_AUTH_TOKEN", defaultTestToken)
	if authToken == defaultTestToken && os.Getenv("API_AUTH_TOKEN") == "" {
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

	time.Sleep(100 * time.Millisecond)

	wd, _ := os.Getwd()
	testFilePath := filepath.Join(wd, testfileName)

	err = os.WriteFile(testFilePath, []byte(testfileContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file '%s': %v", testFilePath, err)
	}

	exitCode := m.Run()

	fmt.Println("INFO: Shutting down test server...")
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("Test server graceful shutdown failed: %v", err)
	} else {
		fmt.Println("INFO: Test server shut down gracefully.")
	}

	os.Remove(testFilePath)

	os.Exit(exitCode)
}

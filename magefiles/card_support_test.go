package magefiles

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestEnsureOracleCardsDownloadsAndReusesCache(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		switch request.URL.Path {
		case "/metadata":
			_ = json.NewEncoder(writer).Encode(bulkDataMetadata{DownloadURI: serverURL(request) + "/cards"})
		case "/cards":
			_, _ = writer.Write([]byte(`[{"name":"Bear"}]`))
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	path := filepath.Join(t.TempDir(), "cache", "oracle-cards.json")
	if err := ensureOracleCards(context.Background(), server.Client(), server.URL+"/metadata", path); err != nil {
		t.Fatal(err)
	}
	if err := ensureOracleCards(context.Background(), server.Client(), server.URL+"/metadata", path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `[{"name":"Bear"}]`; got != want {
		t.Fatalf("cache = %q, want %q", got, want)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("requests = %d, want metadata and download only", got)
	}
}

func serverURL(request *http.Request) string {
	return "http://" + request.Host
}

func TestEnsureOracleCardsRejectsMetadataWithoutDownload(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	err := ensureOracleCards(
		context.Background(),
		server.Client(),
		server.URL,
		filepath.Join(t.TempDir(), "oracle-cards.json"),
	)
	if err == nil {
		t.Fatal("ensureOracleCards() succeeded without a download URI")
	}
}

func TestEnsureOracleCardsFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		metadataStatus  int
		downloadStatus  int
		downloadContent string
	}{
		{name: "metadata status", metadataStatus: http.StatusServiceUnavailable},
		{name: "download status", metadataStatus: http.StatusOK, downloadStatus: http.StatusServiceUnavailable},
		{name: "empty download", metadataStatus: http.StatusOK, downloadStatus: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/metadata":
					writer.WriteHeader(test.metadataStatus)
					if test.metadataStatus == http.StatusOK {
						_ = json.NewEncoder(writer).Encode(bulkDataMetadata{DownloadURI: serverURL(request) + "/cards"})
					}
				case "/cards":
					writer.WriteHeader(test.downloadStatus)
					_, _ = writer.Write([]byte(test.downloadContent))
				default:
					http.NotFound(writer, request)
				}
			}))
			t.Cleanup(server.Close)
			path := filepath.Join(t.TempDir(), "oracle-cards.json")
			if err := ensureOracleCards(context.Background(), server.Client(), server.URL+"/metadata", path); err == nil {
				t.Fatal("ensureOracleCards() succeeded")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("cache file exists or stat failed: %v", err)
			}
		})
	}
}

func TestEnsureOracleCardsRejectsInvalidExistingCache(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "oracle-cards.json")
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := ensureOracleCards(context.Background(), http.DefaultClient, "", path); err == nil {
		t.Fatal("ensureOracleCards() accepted a directory as cache")
	}
}

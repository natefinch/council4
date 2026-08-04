package magefiles

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"
)

func TestCardSupportSettings(t *testing.T) {
	t.Parallel()
	externalOutput := filepath.Join("other", "cards")
	tests := []struct {
		name         string
		output       string
		wantOutput   string
		wantCompiler string
		wantDocs     bool
	}{
		{
			name:         "repository generation",
			wantOutput:   filepath.FromSlash(defaultCardSupportOutput),
			wantCompiler: "./cardgen/oracle/cmd/compilecards",
			wantDocs:     true,
		},
		{
			name:         "external generation",
			output:       externalOutput,
			wantOutput:   externalOutput,
			wantCompiler: "github.com/natefinch/council4/cardgen/oracle/cmd/compilecards",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			output, compiler, docs := cardSupportSettings(test.output)
			if output != test.wantOutput {
				t.Errorf("output = %q, want %q", output, test.wantOutput)
			}
			if compiler != test.wantCompiler {
				t.Errorf("compiler = %q, want %q", compiler, test.wantCompiler)
			}
			if docs != test.wantDocs {
				t.Errorf("documentation = %v, want %v", docs, test.wantDocs)
			}
		})
	}
}

func TestCardSupportArgs(t *testing.T) {
	t.Parallel()
	args := cardSupportArgs("compiler", "oracle.json", "generated", true)
	assertArgPair(t, args, "-in", "oracle.json")
	assertArgPair(t, args, "-out", "generated")
	assertArgPair(t, args, "-report", filepath.FromSlash(CardSupportReportPath))
	if !containsArg(args, "-readme") {
		t.Fatalf("documentation args absent: %v", args)
	}
}

func assertArgPair(t *testing.T, args []string, name, want string) {
	t.Helper()
	for index := range len(args) - 1 {
		if args[index] == name {
			if got := args[index+1]; got != want {
				t.Fatalf("%s value = %q, want %q", name, got, want)
			}
			return
		}
	}
	t.Fatalf("%s not found in %v", name, args)
}

func containsArg(args []string, want string) bool {
	return slices.Contains(args, want)
}

func TestEnsureOracleCardsDownloadsAndReusesCache(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	download := gzipJSONL(t, `{"name":"Bear"}`, `{"name":"Wolf"}`)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		switch request.URL.Path {
		case "/metadata":
			_ = json.NewEncoder(writer).Encode(map[string]string{
				"jsonl_download_uri": serverURL(request) + "/cards",
			})
		case "/cards":
			_, _ = writer.Write(download)
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
	if got, want := string(data), `[{"name":"Bear"},{"name":"Wolf"}]`; got != want {
		t.Fatalf("cache = %q, want %q", got, want)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("requests = %d, want metadata and download only", got)
	}
}

func gzipJSONL(t *testing.T, records ...string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	for _, record := range records {
		if _, err := writer.Write([]byte(record + "\n")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
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
		downloadContent []byte
	}{
		{name: "metadata status", metadataStatus: http.StatusServiceUnavailable},
		{name: "download status", metadataStatus: http.StatusOK, downloadStatus: http.StatusServiceUnavailable},
		{
			name:            "invalid gzip",
			metadataStatus:  http.StatusOK,
			downloadStatus:  http.StatusOK,
			downloadContent: []byte("not gzip"),
		},
		{
			name:            "empty JSONL",
			metadataStatus:  http.StatusOK,
			downloadStatus:  http.StatusOK,
			downloadContent: gzipJSONL(t),
		},
		{
			name:            "invalid JSONL",
			metadataStatus:  http.StatusOK,
			downloadStatus:  http.StatusOK,
			downloadContent: gzipJSONL(t, `{"name":`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/metadata":
					writer.WriteHeader(test.metadataStatus)
					if test.metadataStatus == http.StatusOK {
						_ = json.NewEncoder(writer).Encode(map[string]string{
							"jsonl_download_uri": serverURL(request) + "/cards",
						})
					}
				case "/cards":
					writer.WriteHeader(test.downloadStatus)
					_, _ = writer.Write(test.downloadContent)
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

package magefiles

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/internal/preservation"
)

func TestPinnedCorpusVerifiesDownloadAndConvertedBytes(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name          string
		wrongDownload bool
		wrongJSON     bool
	}{
		{name: "verified"},
		{name: "changed archive", wrongDownload: true},
		{name: "changed converted bytes", wrongJSON: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			download := gzipJSONL(t, `{"name":"Bear"}`, `{"name":"Wolf"}`)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(download)
			}))
			t.Cleanup(server.Close)
			archiveHash, err := preservation.Digest(bytes.NewReader(download))
			if err != nil {
				t.Fatal(err)
			}
			jsonHash, err := preservation.Digest(strings.NewReader(`[{"name":"Bear"},{"name":"Wolf"}]`))
			if err != nil {
				t.Fatal(err)
			}
			pin := preservation.Pin{
				DownloadURI: server.URL, DownloadSHA256: archiveHash, JSONSHA256: jsonHash,
				UpdatedAt: "2026-10-13", RulesEdition: "2026-09-25", BaselineRevision: "baseline",
			}
			if scenario.wrongDownload {
				pin.DownloadSHA256 = strings.Repeat("a", 64)
			}
			if scenario.wrongJSON {
				pin.JSONSHA256 = strings.Repeat("b", 64)
			}
			path := filepath.Join(t.TempDir(), "pinned.json")
			err = installPinnedCorpus(context.Background(), server.Client(), pin, path)
			if scenario.wrongDownload || scenario.wrongJSON {
				if err == nil {
					t.Fatal("accepted mismatched bytes")
				}
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatal("installed an unverified corpus")
				}
			} else if err != nil {
				t.Fatal(err)
			} else if err := pin.Verify(path); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLegacyParserSubsetExcludesRefusedAndAuthoredOnlyEntries(t *testing.T) {
	t.Parallel()
	got := legacySupportedMarkdown([]preservation.Record{
		{Name: "Legacy", DefinitionSHA256: "fingerprint"},
		{Name: "Refused", Refusals: []string{"unsupported"}},
	})
	if got != "- Legacy\n" {
		t.Fatalf("legacy parser subset = %q", got)
	}
}

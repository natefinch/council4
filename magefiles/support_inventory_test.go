package magefiles

import (
	"slices"
	"testing"
)

func TestSupportInventoryReusesExistingEvidence(t *testing.T) {
	t.Parallel()
	got := supportInventoryArgs("oracle.json", "existing-report.json")
	want := []string{
		"run", "./cmd/supportinventory",
		"-corpus", "oracle.json", "-compiler-report", "existing-report.json", "-docs",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("report command = %v, want %v", got, want)
	}
}

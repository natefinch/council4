package cardgen

import (
	"strings"
	"testing"
)

func TestMissingWorklistIncludesMissingAndInvalidCards(t *testing.T) {
	manifest := Manifest{Version: ManifestVersion, Cards: []ManifestCard{
		{CanonicalName: "Ready Card", FileStatus: BatchFileStatusExisting, Validation: BatchValidationStatusValid},
		{CanonicalName: "Missing Card", FileStatus: BatchFileStatusMissing},
		{CanonicalName: "Invalid Card", FileStatus: BatchFileStatusExisting, Validation: BatchValidationStatusInvalid},
		{CanonicalName: "Fetch Error", Status: BatchStatusFetchError, FileStatus: BatchFileStatusMissing},
	}}

	names := MissingWorklist(manifest, 0)

	if strings.Join(names, ",") != "Missing Card,Invalid Card" {
		t.Fatalf("names = %v, want missing and invalid cards", names)
	}
}

func TestMissingWorklistHonorsLimit(t *testing.T) {
	manifest := Manifest{Version: ManifestVersion, Cards: []ManifestCard{
		{CanonicalName: "First", FileStatus: BatchFileStatusMissing},
		{CanonicalName: "Second", FileStatus: BatchFileStatusMissing},
	}}

	names := MissingWorklist(manifest, 1)

	if len(names) != 1 || names[0] != "First" {
		t.Fatalf("names = %v, want only First", names)
	}
}

func TestValidationProgramIncludesWantedLettersAndNames(t *testing.T) {
	program := validationProgram(map[string]bool{
		"Lightning Bolt": true,
		"Sol Ring":       true,
	})

	for _, want := range []string{
		`"github.com/natefinch/council4/mtg/cards/l"`,
		`"github.com/natefinch/council4/mtg/cards/s"`,
		`"Lightning Bolt": true`,
		`"Sol Ring": true`,
	} {
		if !strings.Contains(program, want) {
			t.Fatalf("program missing %q:\n%s", want, program)
		}
	}
}

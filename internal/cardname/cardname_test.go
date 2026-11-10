package cardname

import (
	"encoding/json"
	"go/token"
	"os"
	"regexp"
	"slices"
	"testing"
	"unicode"
)

var fileNamePattern = regexp.MustCompile(`^[_A-Za-z0-9][_A-Za-z0-9-]*\.go$`)

func TestDiacriticNamesDeriveASCIISourceNames(t *testing.T) {
	for _, tc := range []struct{ name, stem, symbol string }{
		{"El-Hajjâj", "el_hajjaj", "ElHajjaj"},
		{"Lim-Dûl's Vault", "lim_dul_s_vault", "LimDulSVault"},
		{"Éowyn, Lady of Rohan", "eowyn_lady_of_rohan", "EowynLadyOfRohan"},
		{"Mjölnir, Hammer of Thor", "mjolnir_hammer_of_thor", "MjolnirHammerOfThor"},
		{"Robo-Piñata", "robo_pinata", "RoboPinata"},
		{"Altaïr Ibn-La'Ahad", "altair_ibn_la_ahad", "AltairIbnLaAhad"},
		{"Æther Vial", "aether_vial", "AEtherVial"},
		{"Ōkubo, Šárka", "okubo_sarka", "OkuboSarka"},
		{"+2 Mace", "2_mace", "Card2Mace"},
		{"Strong Arm", "strong_arm_card", "StrongArm"},
		{"Cards", "cards_card", "Cards"},
		{"Field Test", "field_test_card", "FieldTest"},
		{"Rhino®—Charge", "rhino_charge", "RhinoCharge"},
	} {
		stem, err := FileStem(tc.name)
		if err != nil || stem != tc.stem {
			t.Errorf("FileStem(%q) = %q, %v; want %q", tc.name, stem, err, tc.stem)
		}
		symbol, err := Symbol(tc.name)
		if err != nil || symbol != tc.symbol {
			t.Errorf("Symbol(%q) = %q, %v; want %q", tc.name, symbol, err, tc.symbol)
		}
	}
}

func TestUntransliterableNamesFailClosed(t *testing.T) {
	for _, name := range []string{"Ψυχή", "東京", "!!!", ""} {
		if stem, err := FileStem(name); err == nil {
			t.Errorf("FileStem(%q) = %q, want error", name, stem)
		}
		if symbol, err := Symbol(name); err == nil {
			t.Errorf("Symbol(%q) = %q, want error", name, symbol)
		}
	}
}

// TestSavedQueueDiacriticNamesDeriveDistinctSourceNames covers every
// non-ASCII printed name in the saved production population, including all the
// cards parked behind generated-filename support: each derives a valid ASCII
// file name and symbol that no other printed name in the population shares.
func TestSavedQueueDiacriticNamesDeriveDistinctSourceNames(t *testing.T) {
	data, err := os.ReadFile("../../testdata/card-authoring/production.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Records []struct {
			Name string `json:"name"`
		} `json:"records"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	stems := map[string][]string{}
	symbols := map[string][]string{}
	var diacritics []string
	for _, record := range manifest.Records {
		if !isASCII(record.Name) {
			diacritics = append(diacritics, record.Name)
		}
		if stem, err := FileStem(record.Name); err == nil && !slices.Contains(stems[stem], record.Name) {
			stems[stem] = append(stems[stem], record.Name)
		}
		if symbol, err := Symbol(record.Name); err == nil && !slices.Contains(symbols[symbol], record.Name) {
			symbols[symbol] = append(symbols[symbol], record.Name)
		}
	}
	if len(diacritics) < 100 {
		t.Fatalf("saved population has only %d non-ASCII card names", len(diacritics))
	}
	for _, name := range diacritics {
		stem, err := FileStem(name)
		if err != nil {
			t.Error(err)
			continue
		}
		symbol, err := Symbol(name)
		if err != nil {
			t.Error(err)
			continue
		}
		if !fileNamePattern.MatchString(stem+".go") || !fileNamePattern.MatchString(stem+"_test.go") {
			t.Errorf("%q derives invalid file name %q", name, stem)
		}
		if !token.IsIdentifier(symbol) || !token.IsExported(symbol) || !isASCII(symbol) {
			t.Errorf("%q derives invalid symbol %q", name, symbol)
		}
		if len(stems[stem]) != 1 || len(symbols[symbol]) != 1 {
			t.Errorf("%q shares derived names: %v %v", name, stems[stem], symbols[symbol])
		}
	}
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

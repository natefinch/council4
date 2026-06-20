package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseAmongControlledManaTypedSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name          string
		text          string
		wantTypes     []CardType
		wantSupertype []Supertype
		wantKind      SelectionKind
	}{
		{
			"legendary creatures and planeswalkers",
			"{T}: Add one mana of any color among legendary creatures and planeswalkers you control.",
			[]CardType{CardTypeCreature, CardTypePlaneswalker},
			[]Supertype{SupertypeLegendary},
			SelectionCreature,
		},
		{
			"legendary permanents",
			"{T}: Add one mana of any color among legendary permanents you control.",
			nil,
			[]Supertype{SupertypeLegendary},
			SelectionPermanent,
		},
		{
			"creatures",
			"{T}: Add one mana of any color among creatures you control.",
			[]CardType{CardTypeCreature},
			nil,
			SelectionCreature,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			mana := effects[0].Mana
			if !mana.ColorsAmongControlled || !mana.LegacyBodyExact || mana.ColorsAmongSelection == nil {
				t.Fatalf("mana = %#v", mana)
			}
			selection := *mana.ColorsAmongSelection
			if selection.Controller != SelectionControllerYou || selection.Zone != zone.None {
				t.Fatalf("selection controller/zone = %#v", selection)
			}
			if selection.Kind != tc.wantKind {
				t.Fatalf("selection kind = %q, want %q", selection.Kind, tc.wantKind)
			}
			if !slices.Equal(selection.RequiredTypesAny, tc.wantTypes) {
				t.Fatalf("required types = %#v, want %#v", selection.RequiredTypesAny, tc.wantTypes)
			}
			if !slices.Equal(selection.Supertypes, tc.wantSupertype) {
				t.Fatalf("supertypes = %#v, want %#v", selection.Supertypes, tc.wantSupertype)
			}
		})
	}
}

func TestParseAmongControlledManaFailsClosed(t *testing.T) {
	t.Parallel()
	// A bare wildcard permanent group, a foreign controller, and a non-"one
	// mana" quantity are not modeled by the among-controlled recognizer, so the
	// among-controlled flag must stay unset.
	variants := []string{
		"{T}: Add one mana of any color among permanents you control.",
		"{T}: Add one mana of any color among creatures an opponent controls.",
		"{T}: Add two mana of any color among creatures you control.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			continue
		}
		if effects[0].Mana.ColorsAmongControlled {
			t.Fatalf("variant unexpectedly recognized among-controlled mana:\n%s", source)
		}
	}
}

// TestParseEachColorAmongControlledMana verifies Bloom Tender's "For each color
// among <permanents> you control, add one mana of that color" body sets the
// each-color flag and carries the permanent filter, accepting a bare
// "permanents you control" group and a narrowed "creatures you control" group.
func TestParseEachColorAmongControlledMana(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		text      string
		wantTypes []CardType
		wantKind  SelectionKind
	}{
		{
			"permanents",
			"{T}: For each color among permanents you control, add one mana of that color.",
			nil,
			SelectionPermanent,
		},
		{
			"creatures",
			"{T}: For each color among creatures you control, add one mana of that color.",
			[]CardType{CardTypeCreature},
			SelectionCreature,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			mana := effects[0].Mana
			if !mana.EachColorAmongControlled || mana.ColorsAmongSelection == nil {
				t.Fatalf("mana = %#v", mana)
			}
			selection := *mana.ColorsAmongSelection
			if selection.Controller != SelectionControllerYou || selection.Zone != zone.None {
				t.Fatalf("selection controller/zone = %#v", selection)
			}
			if selection.Kind != tc.wantKind {
				t.Fatalf("selection kind = %q, want %q", selection.Kind, tc.wantKind)
			}
			if !slices.Equal(selection.RequiredTypesAny, tc.wantTypes) {
				t.Fatalf("required types = %#v, want %#v", selection.RequiredTypesAny, tc.wantTypes)
			}
		})
	}
}

// TestParseEachColorAmongControlledManaFailsClosed verifies a foreign controller
// and a non-"one mana" quantity are not recognized as each-color mana.
func TestParseEachColorAmongControlledManaFailsClosed(t *testing.T) {
	t.Parallel()
	variants := []string{
		"{T}: For each color among creatures an opponent controls, add one mana of that color.",
		"{T}: For each color among permanents you control, add two mana of that color.",
		"{T}: For each color among monocolored permanents you control, add one mana of that color.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{})
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
			continue
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 1 {
			continue
		}
		if effects[0].Mana.EachColorAmongControlled {
			t.Fatalf("variant unexpectedly recognized each-color mana:\n%s", source)
		}
	}
}

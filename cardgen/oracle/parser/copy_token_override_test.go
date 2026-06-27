package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestParseCopyTokenOverrideReplace verifies the replacement form of the
// copy-token characteristic-overriding exception ("except it's a 1/1 green Frog"
// — Croaking Counterpart) records a non-additive power/toughness, color, and
// subtype override on the copy create.
func TestParseCopyTokenOverrideReplace(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature, except it's a 1/1 green Frog.",
		Context{InstantOrSorcery: true, CardName: "Frog Maker"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.TokenCopyOverride {
		t.Fatalf("TokenCopyOverride = false, want true: %#v", effect)
	}
	if !effect.TokenCopyOverridePTKnown ||
		effect.TokenCopyOverridePower != 1 || effect.TokenCopyOverrideToughness != 1 {
		t.Errorf("P/T = %d/%d known=%v, want 1/1 known",
			effect.TokenCopyOverridePower, effect.TokenCopyOverrideToughness, effect.TokenCopyOverridePTKnown)
	}
	if len(effect.TokenCopyOverrideColors) != 1 || effect.TokenCopyOverrideColors[0] != ColorGreen {
		t.Errorf("colors = %v, want [green]", effect.TokenCopyOverrideColors)
	}
	if len(effect.TokenCopyOverrideSubtypes) != 1 || effect.TokenCopyOverrideSubtypes[0] != types.Frog {
		t.Errorf("subtypes = %v, want [Frog]", effect.TokenCopyOverrideSubtypes)
	}
	if effect.TokenCopyOverrideAdditiveColors || effect.TokenCopyOverrideAdditiveTypes {
		t.Error("replacement form must not mark colors or types additive")
	}
}

// TestParseCopyTokenOverrideAdditiveColorsAndTypes verifies the additive form
// whose "in addition to its other colors and types" suffix marks both the color
// and the subtype additive ("except it's not legendary and it's a 2/2 black
// Zombie in addition to its other colors and types" — Ratadrabik of Urborg),
// alongside the legendary drop.
func TestParseCopyTokenOverrideAdditiveColorsAndTypes(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature, except it's not legendary "+
			"and it's a 2/2 black Zombie in addition to its other colors and types.",
		Context{InstantOrSorcery: true, CardName: "Zombie Maker"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.TokenCopyOverride || !effect.TokenCopyDropLegendary {
		t.Fatalf("override=%v dropLegendary=%v, want both true",
			effect.TokenCopyOverride, effect.TokenCopyDropLegendary)
	}
	if !effect.TokenCopyOverrideAdditiveColors || !effect.TokenCopyOverrideAdditiveTypes {
		t.Errorf("additiveColors=%v additiveTypes=%v, want both true",
			effect.TokenCopyOverrideAdditiveColors, effect.TokenCopyOverrideAdditiveTypes)
	}
	if len(effect.TokenCopyOverrideColors) != 1 || effect.TokenCopyOverrideColors[0] != ColorBlack {
		t.Errorf("colors = %v, want [black]", effect.TokenCopyOverrideColors)
	}
	if len(effect.TokenCopyOverrideSubtypes) != 1 || effect.TokenCopyOverrideSubtypes[0] != types.Zombie {
		t.Errorf("subtypes = %v, want [Zombie]", effect.TokenCopyOverrideSubtypes)
	}
}

// TestParseCopyTokenOverrideNamedFailsClosed verifies a copy-token exception
// that renames the token ("except it's a legendary Alien named Prisoner Zero" —
// The Eleventh Hour) is not recognized as an exact copy create: a printed token
// name is outside the supported characteristic-override grammar, so the copy
// fails closed rather than silently dropping the name.
func TestParseCopyTokenOverrideNamedFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Create a token that's a copy of target creature, except it's a legendary Alien named Prisoner Zero.",
		Context{InstantOrSorcery: true, CardName: "Named Maker"})
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.TokenCopyOverride {
				t.Fatalf("named-token exception must fail closed, got override with subtypes %v",
					effect.TokenCopyOverrideSubtypes)
			}
		}
	}
}

// TestParseCopyTokenOverrideMixedAdditivityFailsClosed verifies that a
// multi-clause exception combining a replace-mode characteristic with a
// separate additive clause that turns on the shared additive flag fails closed
// rather than silently lowering the replace-mode subtype as additive. Here the
// first clause replaces the subtype ("a 1/1 green Frog") while the second turns
// on the additive-types flag ("an artifact in addition to its other types")
// without carrying a subtype of its own, so the committed replace mode for the
// Frog subtype disagrees with the global additive flag the lowering consults.
func TestParseCopyTokenOverrideMixedAdditivityFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Create a token that's a copy of target creature, except it's a 1/1 green Frog "+
			"and it's an artifact in addition to its other types.",
		Context{InstantOrSorcery: true, CardName: "Mixed Maker"})
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.TokenCopyOverride {
				t.Fatalf("mixed-additivity exception must fail closed, got override subtypes=%v additiveTypes=%v",
					effect.TokenCopyOverrideSubtypes, effect.TokenCopyOverrideAdditiveTypes)
			}
		}
	}
}

package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParseCopyTokenExceptKeywordRider verifies the inline copy-token rider
// "except <the token> has <keyword> and it isn't legendary" (Irenicus's Vile
// Duplication) folds into a single copy create that records the granted keyword
// and the drop-legendary flag rather than stranding a separate keyword-grant
// effect.
func TestParseCopyTokenExceptKeywordRider(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature you control, except the token has flying and it isn't legendary.",
		Context{InstantOrSorcery: true, CardName: "Irenicus's Vile Duplication"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want a single folded copy create", effects)
	}
	effect := effects[0]
	if !effect.TokenCopyOfTarget {
		t.Fatalf("TokenCopyOfTarget = false, want true: %#v", effect)
	}
	if !effect.TokenCopyDropLegendary {
		t.Error("TokenCopyDropLegendary = false, want true")
	}
	if len(effect.TokenCopyGrantKeywords) != 1 || effect.TokenCopyGrantKeywords[0] != KeywordFlying {
		t.Errorf("TokenCopyGrantKeywords = %v, want [flying]", effect.TokenCopyGrantKeywords)
	}
}

// TestParseCopyTokenExceptQuotedAbilityFailsClosed verifies a copy-token rider
// carrying a quoted granted ability ("except it has haste and \"...sacrifice
// this token.\"" — Electroduplicate, Heat Shimmer) fails closed: the quoted
// ability cannot be represented, so the copy is not recognized as exact rather
// than silently dropping the ability and keeping only the keyword.
func TestParseCopyTokenExceptQuotedAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Create a token that's a copy of target creature you control, except it has haste and \"At the beginning of the end step, sacrifice this token.\"",
		Context{InstantOrSorcery: true, CardName: "Electroduplicate"})
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.TokenCopyOfTarget {
				t.Fatalf("quoted-ability rider must fail closed, got recognized copy create with keywords %v", effect.TokenCopyGrantKeywords)
			}
		}
	}
}

// TestParseCopyTokenGraveyardCardTarget verifies that a copy-token whose
// blueprint is a card chosen in a graveyard ("Create a token that's a copy of
// target creature card in your graveyard, except it's an artifact in addition to
// its other types." — Feldon of the Third Path) is recognized as a target copy
// even though a graveyard-card target does not round-trip through the
// permanent-target reconstruction. The graveyard target keeps its Graveyard zone
// and the "except it's an artifact" additive-type rider folds into the copy.
func TestParseCopyTokenGraveyardCardTarget(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types.",
		Context{InstantOrSorcery: true, CardName: "Feldon of the Third Path"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want a single folded copy create", effects)
	}
	effect := effects[0]
	if !effect.TokenCopyOfTarget {
		t.Fatalf("TokenCopyOfTarget = false, want true: %#v", effect)
	}
	if len(effect.Targets) != 1 || effect.Targets[0].Selection.Zone != zone.Graveyard {
		t.Fatalf("target = %#v, want a single graveyard-card target", effect.Targets)
	}
	if effect.Targets[0].Exact {
		t.Error("graveyard-card target Exact = true, want false")
	}
}

// TestParseCopyTokenBattlefieldTargetStillExact guards that the battlefield
// permanent copy target keeps its exact round-tripping recognition unchanged by
// the graveyard-card relaxation ("Create a token that's a copy of target
// creature you control." — Spitting Image family).
func TestParseCopyTokenBattlefieldTargetStillExact(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Create a token that's a copy of target creature you control.",
		Context{InstantOrSorcery: true, CardName: "Spitting Image"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if !effect.TokenCopyOfTarget {
		t.Fatalf("TokenCopyOfTarget = false, want true: %#v", effect)
	}
	if len(effect.Targets) != 1 || effect.Targets[0].Selection.Zone != zone.None {
		t.Fatalf("target = %#v, want a single battlefield-permanent target", effect.Targets)
	}
	if !effect.Targets[0].Exact {
		t.Error("battlefield-permanent target Exact = false, want true")
	}
}

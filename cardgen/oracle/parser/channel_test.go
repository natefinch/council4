package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseChannelDiscardSelfCost(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"Channel — {1}{G}, Discard this card: Destroy target artifact.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Label != "Channel" {
		t.Fatalf("ability word = %#v, want Channel", ability.AbilityWord)
	}
	if ability.CostSyntax == nil || len(ability.CostSyntax.Components) != 2 {
		t.Fatalf("cost = %#v, want mana and discard-self components", ability.CostSyntax)
	}
	discard := ability.CostSyntax.Components[1]
	if discard.Kind != CostComponentDiscard || !discard.SourceSelf {
		t.Fatalf("discard = %#v, want discard-self", discard)
	}
	if discard.SourceZone != zone.Hand {
		t.Fatalf("discard source zone = %v, want hand", discard.SourceZone)
	}
}

func TestParseOpponentControlledArtifactEnchantmentOrNonbasicLandTarget(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"Destroy target artifact, enchantment, or nonbasic land an opponent controls.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	targets := document.Abilities[0].Sentences[0].Targets
	if len(targets) != 1 {
		t.Fatalf("targets = %#v, want one", targets)
	}
	target := targets[0]
	if !target.Exact || target.Selection.Controller != SelectionControllerOpponent ||
		len(target.Selection.Alternatives) != 3 {
		t.Fatalf("target = %#v, want exact opponent-controlled three-way union", target)
	}
	land := target.Selection.Alternatives[2]
	if land.Kind != SelectionLand ||
		len(land.ExcludedSupertypes) != 1 ||
		land.ExcludedSupertypes[0] != SupertypeBasic {
		t.Fatalf("land alternative = %#v, want nonbasic land", land)
	}
}

func TestParseAffectedPlayerBasicLandTypeSearch(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"Destroy target land. That player may search their library for a land card with a basic land type, put it onto the battlefield, then shuffle.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[1].Effects
	if len(effects) == 0 {
		t.Fatal("search sentence has no effects")
	}
	search := effects[0]
	if search.Kind != EffectSearch || search.UnsupportedDetail != "" ||
		!search.Selection.BasicLandType ||
		len(search.Selection.Supertypes) != 0 {
		t.Fatalf("search = %#v, want exact land-with-basic-land-type search", search)
	}
}

package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCourtOfLocthwain covers Court of Locthwain's
// upkeep trigger: "At the beginning of your upkeep, exile the top card of target
// opponent's library. You may play that card for as long as it remains exiled,
// and mana of any type can be spent to cast it. If you're the monarch, until end
// of turn, you may cast a spell from among cards exiled with this enchantment
// without paying its mana cost." The exile and its any-type-mana play permission
// lower to one ImpulseExile of the top card of the single target opponent's
// library, remembered under a source-keyed linked set, and the monarch-gated
// free cast lowers to a ControllerIsMonarch-gated ApplyRule installing an
// until-end-of-turn RuleEffectCastLinkedExileForFree over that same pool. The
// whole ability lowers without diagnostics.
func TestGenerateExecutableCardSourceCourtOfLocthwain(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Court of Locthwain",
		Layout:   "normal",
		ManaCost: "{2}{B}{B}",
		TypeLine: "Enchantment",
		OracleText: "When this enchantment enters, you become the monarch.\n" +
			"At the beginning of your upkeep, exile the top card of target opponent's library. You may play that card for as long as it remains exiled, and mana of any type can be spent to cast it. If you're the monarch, until end of turn, you may cast a spell from among cards exiled with this enchantment without paying its mana cost.",
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		`Constraint: "target opponent",`,
		"game.Selection{Player: game.PlayerOpponent}",
		"Primitive: game.ImpulseExile{",
		"Player:        game.TargetPlayerReference(0),",
		"Duration:      game.DurationPermanent,",
		"SpendAnyMana:  true,",
		`PublishLinked: game.LinkedKey("court-of-locthwain-exile"),`,
		"Primitive: game.ApplyRule{",
		"Kind:           game.RuleEffectCastLinkedExileForFree,",
		"AffectedPlayer: game.PlayerYou,",
		`ExiledLinkKey:  game.LinkedKey("court-of-locthwain-exile"),`,
		"Duration: game.DurationUntilEndOfTurn,",
		"ControllerIsMonarch: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCourtOfLocthwainMonarchGateFailsClosed confirms
// the upkeep recognizer reads the typed monarch condition, not merely the
// presence of a resolving condition: an "If an opponent is the monarch" gate has
// a different condition predicate, so the ability is not recognized, does not
// emit the monarch-gated free cast, and is reported unsupported.
func TestGenerateExecutableCardSourceCourtOfLocthwainMonarchGateFailsClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, _ := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Locthwain Opponent Gate",
		Layout:     "normal",
		ManaCost:   "{2}{B}{B}",
		TypeLine:   "Enchantment",
		OracleText: "At the beginning of your upkeep, exile the top card of target opponent's library. You may play that card for as long as it remains exiled, and mana of any type can be spent to cast it. If an opponent is the monarch, until end of turn, you may cast a spell from among cards exiled with this enchantment without paying its mana cost.",
	}, "t")
	if len(diagnostics) == 0 {
		t.Fatalf("expected diagnostics for unrecognized monarch gate, got source:\n%s", source)
	}
	if strings.Contains(source, "game.RuleEffectCastLinkedExileForFree") {
		t.Fatalf("opponent-monarch gate unexpectedly lowered to the linked free cast:\n%s", source)
	}
}

package cardgen

import (
	"strings"
	"testing"
)

// generateCurseSource renders the executable Go source for a single-face Aura
// Curse from its enchant line plus ability text, failing on any error or
// diagnostic, and returns the source so a test can assert the rendered constructs.
func generateCurseSource(t *testing.T, name, manaCost, ability string) string {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       name,
		Layout:     "normal",
		ManaCost:   manaCost,
		TypeLine:   "Enchantment — Aura Curse",
		OracleText: "Enchant player\n" + ability,
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	return source
}

// TestGenerateExecutableCardSourceCurseOfVitality proves the gain-life reflexive
// rider renders the controller gain plus a group gain whose recipients are the
// opponents-attacking-trigger-player group, exercising the player-group reference
// renderer for that kind.
func TestGenerateExecutableCardSourceCurseOfVitality(t *testing.T) {
	t.Parallel()
	source := generateCurseSource(t, "Curse of Vitality", "{2}{W}",
		"Whenever enchanted player is attacked, you gain 2 life. Each opponent attacking that player does the same.")
	for _, want := range []string{
		"Primitive: game.GainLife{",
		"Player: game.ControllerReference()",
		"PlayerGroup: game.OpponentsAttackingTriggerPlayerReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCurseOfVerbosity proves the draw reflexive rider
// renders the controller draw plus a group draw over the opponents-attacking
// group.
func TestGenerateExecutableCardSourceCurseOfVerbosity(t *testing.T) {
	t.Parallel()
	source := generateCurseSource(t, "Curse of Verbosity", "{2}{U}",
		"Whenever enchanted player is attacked, you draw a card. Each opponent attacking that player does the same.")
	for _, want := range []string{
		"Primitive: game.Draw{",
		"Player: game.ControllerReference()",
		"PlayerGroup: game.OpponentsAttackingTriggerPlayerReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCurseOfBounty proves the explicit untap rider
// renders the controller's own nonland untap plus a per-attacker untap that
// iterates the opponents-attacking group with the group-offer member reference,
// exercising the ForEachPlayerGroup instruction renderer.
func TestGenerateExecutableCardSourceCurseOfBounty(t *testing.T) {
	t.Parallel()
	source := generateCurseSource(t, "Curse of Bounty", "{1}{G}",
		"Whenever enchanted player is attacked, untap all nonland permanents you control. Each opponent attacking that player untaps all nonland permanents they control.")
	for _, want := range []string{
		"Primitive: game.Untap{",
		"Controller: game.ControllerYou",
		"game.PlayerControlledGroup(game.GroupOfferMemberReference()",
		"ForEachPlayerGroup: opt.Val(game.OpponentsAttackingTriggerPlayerReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

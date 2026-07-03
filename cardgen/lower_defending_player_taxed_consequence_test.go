package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDefendingPlayerTaxedCantBeBlocked covers
// Shrouded Serpent: "Whenever this creature attacks, defending player may pay
// {4}. If that player doesn't, this creature can't be blocked this turn." The
// offer lowers to a Pay charged to the defending player that publishes an unpaid
// result, and the source can't-be-blocked consequence is gated on the payment
// having been declined.
func TestGenerateExecutableCardSourceDefendingPlayerTaxedCantBeBlocked(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Shrouded Serpent",
		Layout:     "normal",
		TypeLine:   "Creature — Snake",
		OracleText: "Whenever this creature attacks, defending player may pay {4}. If that player doesn't, this creature can't be blocked this turn.",
		Power:      new("5"),
		Toughness:  new("5"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Pay",
		"Payer:  opt.Val(game.DefendingPlayerReference())",
		"PublishResult: game.ResultKey(\"defending-player-unpaid\")",
		"Primitive: game.ApplyRule",
		"Kind: game.RuleEffectCantBeBlocked,",
		"Succeeded: game.TriFalse,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateCastFromGraveyardSource covers the targeted graveyard
// cast-permission primitive "you may cast target <card> from your graveyard this
// turn" (Norika Yamazaki, the Poet), which grants the controller permission to
// cast the chosen graveyard card normally until end of turn.
func TestGenerateCastFromGraveyardSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Norika Yamazaki, the Poet",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Human Samurai",
		ManaCost: "{2}{W}",
		OracleText: "Vigilance\n" +
			"Whenever a Samurai or Warrior you control attacks alone, you may cast target enchantment card from your graveyard this turn.",
	}, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.VigilanceStaticBody",
		"Constraint: \"target enchantment card from your graveyard\",",
		"Allow:      game.TargetAllowCard,",
		"TargetZone: zone.Graveyard,",
		"Primitive: game.GrantCastPermission{",
		"Card:     game.CardReference{Kind: game.CardReferenceTarget},",
		"FromZone: zone.Graveyard,",
		"Face:     game.FaceFront,",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDauthiVoidwalker proves Dauthi Voidwalker
// generates cleanly: the Shadow keyword becomes a static ability, the
// "If a card would be put into an opponent's graveyard from anywhere, instead
// exile it with a void counter on it." replacement lowers to the
// graveyard-redirect-exile-with-counter constructor, and the "{T}, Sacrifice
// this creature: Choose an exiled card an opponent owns with a void counter on
// it. You may play it this turn without paying its mana cost." activated ability
// lowers to a single PlayChosenExiledCard primitive carrying the opponent owner
// scope, the void counter filter, the this-turn window, and the free-cast rider.
func TestGenerateExecutableCardSourceDauthiVoidwalker(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Dauthi Voidwalker",
		Layout:    "normal",
		ManaCost:  "{B}{B}",
		TypeLine:  "Creature — Dauthi Rogue",
		Power:     new("3"),
		Toughness: new("2"),
		OracleText: strings.Join([]string{
			"Shadow (This creature can block or be blocked by only creatures with shadow.)",
			"If a card would be put into an opponent's graveyard from anywhere, instead exile it with a void counter on it.",
			"{T}, Sacrifice this creature: Choose an exiled card an opponent owns with a void counter on it. You may play it this turn without paying its mana cost.",
		}, "\n"),
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ShadowStaticBody",
		"game.GraveyardRedirectExileWithCounterReplacement(",
		"Primitive: game.PlayChosenExiledCard{",
		"Player:                game.ControllerReference(),",
		"Zone:                  zone.Exile,",
		"OwnerScope:            game.PlayerOpponent,",
		"Counter:               opt.Val(counter.Void),",
		"Duration:              game.DurationThisTurn,",
		"WithoutPayingManaCost: true,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

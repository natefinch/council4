package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules"
)

func cyclingHandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name,
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		},
	}}
}

// TestGenericDoesNotCompulsivelyCycle checks that cycling is scored on its merits
// — it draws a card but discards the card being cycled, so it is card-neutral, not
// free card advantage. Scored above passing (as every hand-activated ability was,
// at the flat activate score) the agent cycles its whole hand away instead of
// keeping cards to develop a game plan. Cycling must therefore score at or below
// passing so the agent only cycles when it has nothing better to do.
func TestGenericDoesNotCompulsivelyCycle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addObservedHandCard(g, game.Player1, cyclingHandCard("Cycler"))
	act := action.ActivateAbility(cardID, 0, nil, 0)
	strategy := GenericStrategy{}

	score := strategy.ScoreAction(rules.NewObservation(g, game.Player1), act)
	if score > scorePass {
		t.Fatalf("cycling scored %v, want at or below pass %v (cycling is card-neutral, not free draw)", score, scorePass)
	}
}

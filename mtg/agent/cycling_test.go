package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func cyclingHandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name,
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		},
	}}
}

// discardPayoffCreature is a permanent whose ability triggers when its controller
// discards a card, standing in for Captain Howler, Brallin, or Glint-Horn
// Buccaneer.
func discardPayoffCreature(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventCardDiscarded,
						Player: game.TriggerPlayerYou,
					},
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
					},
				}.Ability(),
			},
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

// TestCyclingScoredAsRealPlayWithDiscardPayoff checks that cycling is scored as a
// real play — above passing — when the player controls a "whenever you discard"
// payoff, because then discarding advances a plan rather than idly churning cards.
// The search must keep it as a candidate so its position evaluation can weigh the
// triggered board.
func TestCyclingScoredAsRealPlayWithDiscardPayoff(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addObservedHandCard(g, game.Player1, cyclingHandCard("Cycler"))
	addObservedPermanent(g, game.Player1, discardPayoffCreature("Discard Payoff"))
	act := action.ActivateAbility(cardID, 0, nil, 0)
	strategy := GenericStrategy{}

	score := strategy.ScoreAction(rules.NewObservation(g, game.Player1), act)
	if score <= scorePass {
		t.Fatalf("cycling with a discard payoff scored %v, want above pass %v (it advances the payoff)", score, scorePass)
	}
}

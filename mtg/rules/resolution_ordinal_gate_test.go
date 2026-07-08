package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestResolutionOrdinalGateFiresOnNthResolution proves the runtime tally behind
// "if this is the second time this ability has resolved this turn" (Prowl,
// Pursuit Vehicle). A triggered ability flagged CountsResolutionsThisTurn whose
// draw is gated on the second resolution must not draw on its first resolution
// and must draw on its second resolution in the same turn.
func TestResolutionOrdinalGateFiresOnNthResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Ordinal Source",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			}},
			CountsResolutionsThisTurn: true,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				Condition: opt.Val(game.EffectCondition{
					Condition: opt.Val(game.Condition{SourceAbilityResolutionOrdinalThisTurn: 2}),
				}),
			}}}.Ability(),
		}},
	}}
	source := addCombatPermanent(g, game.Player1, def)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw B"}})

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		AbilityIndex: 0,
	}

	if got := engine.resolveTriggeredAbility(g, obj, &TurnLog{}); got != "resolved" {
		t.Fatalf("first resolution = %q, want resolved", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size after first resolution = %d, want 0 (gate requires second resolution)", got)
	}

	if got := engine.resolveTriggeredAbility(g, obj, &TurnLog{}); got != "resolved" {
		t.Fatalf("second resolution = %q, want resolved", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after second resolution = %d, want 1 (gate satisfied)", got)
	}
}

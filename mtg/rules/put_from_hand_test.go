package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// defaultChoiceAgent always picks the offered default selection, modeling a
// player who accepts the first available option.
type defaultChoiceAgent struct{}

func (defaultChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (defaultChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

func putLandFromHandInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.PutFromHandChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Land}},
			game.Fixed(1),
			false,
			false,
			false,
		),
	}
}

// TestPutFromHandPutsChosenLandOntoBattlefield verifies the controller's chosen
// matching land moves from hand to the battlefield while a non-matching card is
// left untouched.
func TestPutFromHandPutsChosenLandOntoBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putLandFromHandInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("chosen land still in hand")
	}
	if !g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("non-matching card was removed from hand")
	}
	if _, ok := reanimatedPermanent(g, land); !ok {
		t.Fatal("land was not put onto the battlefield")
	}
}

// TestPutFromHandWithNoMatchingCardDoesNothing verifies that with no matching
// card in hand, the effect leaves the hand intact and creates no permanent.
func TestPutFromHandWithNoMatchingCardDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putLandFromHandInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("non-matching card was removed from hand")
	}
	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield permanents = %d, want 1 (only the source)", len(g.Battlefield))
	}
}

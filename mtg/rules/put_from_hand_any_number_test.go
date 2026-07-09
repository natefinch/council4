package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// selectAllChoiceAgent selects every offered option, modeling a player who puts
// as many cards as possible for an "any number" choice.
type selectAllChoiceAgent struct{}

func (selectAllChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (selectAllChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	indices := make([]int, len(request.Options))
	for i := range indices {
		indices[i] = i
	}
	return indices
}

func putAnyNumberCreaturesFromHandInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.PutFromHandChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.Fixed(1),
			false,
			false,
			true, // any number
		),
	}
}

// TestPutAnyNumberCreaturesFromHand verifies the "put any number of creature
// cards from your hand onto the battlefield" form (Ghalta, Stampede Tyrant; Last
// March of the Ents) moves every chosen creature from hand to the battlefield
// while leaving a non-creature card in hand.
func TestPutAnyNumberCreaturesFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	bear := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	elf := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Elf",
		Types: []types.Card{types.Creature},
	}})
	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putAnyNumberCreaturesFromHandInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(bear) || g.Players[game.Player1].Hand.Contains(elf) {
		t.Fatal("a chosen creature is still in hand")
	}
	if !g.Players[game.Player1].Hand.Contains(spell) {
		t.Fatal("the non-creature card was removed from hand")
	}
	if _, ok := reanimatedPermanent(g, bear); !ok {
		t.Fatal("the first creature was not put onto the battlefield")
	}
	if _, ok := reanimatedPermanent(g, elf); !ok {
		t.Fatal("the second creature was not put onto the battlefield")
	}
}

// TestPutAnyNumberFromHandChoosingNoneIsLegal verifies the empty choice is legal
// for the any-number form: a player who declines leaves every card in hand and
// puts nothing onto the battlefield.
func TestPutAnyNumberFromHandChoosingNoneIsLegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	bear := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})

	// defaultChoiceAgent accepts the default selection, which for an any-number
	// choice is the empty set.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putAnyNumberCreaturesFromHandInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(bear) {
		t.Fatal("creature left hand despite the empty choice")
	}
	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield permanents = %d, want 1 (only the source)", len(g.Battlefield))
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// filteredDiscardOptionalInstructions builds the "You may discard a creature
// card. If you do, draw a card." optional flow: the controller chooses a
// creature card to discard (ChooseDiscardFromHand gated by a creature-only
// Selection), the choice publishes its result, and a trailing draw is gated on
// it having succeeded.
func filteredDiscardOptionalInstructions() []game.Instruction {
	key := game.ResultKey("if-you-do")
	return []game.Instruction{
		{
			Optional:      true,
			PublishResult: key,
			Primitive: game.ChooseDiscardFromHand{
				Player:    game.ControllerReference(),
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: key, Succeeded: game.TriTrue}),
		},
	}
}

// TestChooseDiscardFromHandSelectionFiltersByType verifies that a
// ChooseDiscardFromHand carrying a creature-only Selection discards a creature
// card from hand and leaves a noncreature card untouched: the filter excludes
// the noncreature card from the candidate pool entirely.
func TestChooseDiscardFromHandSelectionFiltersByType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	instantID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})
	addInstructionSpellToStack(g, []game.Instruction{{
		Primitive: game.ChooseDiscardFromHand{
			Player:    game.ControllerReference(),
			Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(creatureID) {
		t.Fatal("creature card must be discarded by the creature-only filtered discard")
	}
	if !g.Players[game.Player1].Hand.Contains(instantID) {
		t.Fatal("noncreature card must remain in hand (filtered out of the discard choice)")
	}
}

// TestOptionalFilteredDiscardAcceptDiscardsAndDraws verifies that accepting the
// optional filtered self-discard discards the creature card and performs the
// gated draw.
func TestOptionalFilteredDiscardAcceptDiscardsAndDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	drawID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, filteredDiscardOptionalInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(creatureID) {
		t.Fatal("accepting must discard the creature card")
	}
	if !g.Players[game.Player1].Hand.Contains(drawID) {
		t.Fatal("accepting must perform the gated draw")
	}
}

// TestOptionalFilteredDiscardDeclineSkips verifies that declining the optional
// filtered self-discard discards nothing and skips the gated draw.
func TestOptionalFilteredDiscardDeclineSkips(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	drawID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gated Draw"}})
	addInstructionSpellToStack(g, filteredDiscardOptionalInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(creatureID) {
		t.Fatal("declining must leave the creature card in hand")
	}
	if g.Players[game.Player1].Hand.Contains(drawID) {
		t.Fatal("declining must skip the gated draw")
	}
}

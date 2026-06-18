package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// optionalSearchAgent answers the optional "Apply optional effect?" may-choice
// according to accept, and answers the subsequent search choice by selecting the
// option whose label matches wanted.
type optionalSearchAgent struct {
	accept bool
	wanted string
}

func (optionalSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a optionalSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.accept {
			return []int{1}
		}
		return []int{0}
	}
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return []int{}
}

// TestOptionalSearchDeclineLeavesLibrary verifies that declining an Optional
// Search instruction skips the search entirely, leaving the library untouched.
func TestOptionalSearchDeclineLeavesLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional: true,
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
			},
		},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: false, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(bear) || g.Players[game.Player1].Hand.Contains(bear) {
		t.Fatal("declining the optional search still moved a card out of the library")
	}
}

// TestOptionalSearchAcceptFindsCard verifies that accepting an Optional Search
// instruction performs the search and moves the chosen card.
func TestOptionalSearchAcceptFindsCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	addInstructionSpellToStack(g, []game.Instruction{{
		Optional: true,
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
			},
		},
	}})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(bear) || g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("accepting the optional search did not move the chosen card to hand")
	}
}

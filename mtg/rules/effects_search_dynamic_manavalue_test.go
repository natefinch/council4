package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// recordingSearchAgent captures every search option it is offered, then selects
// the option whose label matches the wanted card name.
type recordingSearchAgent struct {
	wanted  string
	offered []string
}

func (*recordingSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *recordingSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		a.offered = append(a.offered, option.Label)
	}
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return []int{}
}

func manaValueCard(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    []types.Card{types.Creature},
	}}
}

// beseechSearchInstruction builds the Beseech the Queen search: find one library
// card whose mana value is at most the number of lands the controller controls,
// reveal it, and put it into hand.
func beseechSearchInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				Reveal:      true,
				Filter: game.Selection{
					ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{
						Kind:       game.DynamicAmountCountSelector,
						Multiplier: 1,
						Group: game.GroupRef(game.BattlefieldGroup(game.Selection{
							RequiredTypes: []types.Card{types.Land},
							Controller:    game.ControllerYou,
						})),
					}),
				},
			},
		},
	}
}

// TestSearchManaValueDynamicCountBoundMatchesLandCount proves the runtime honors
// a dynamic-count mana-value bound: with three lands on the battlefield, the
// searcher may find only cards whose mana value is at most three, and the
// over-bound card is neither offered nor movable out of the library.
func TestSearchManaValueDynamicCountBoundMatchesLandCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 3 {
		addLandPermanent(g, game.Player1, "Test Land")
	}
	atBound := addCardToLibrary(g, game.Player1, manaValueCard("Three Drop", 3))
	underBound := addCardToLibrary(g, game.Player1, manaValueCard("One Drop", 1))
	overBound := addCardToLibrary(g, game.Player1, manaValueCard("Four Drop", 4))
	addInstructionSpellToStack(g, []game.Instruction{beseechSearchInstruction()})
	agent := &recordingSearchAgent{wanted: "Three Drop"}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	offered := map[string]bool{}
	for _, label := range agent.offered {
		offered[label] = true
	}
	if !offered["Three Drop"] || !offered["One Drop"] {
		t.Fatalf("offered options = %v, want the mana-value <= 3 cards", agent.offered)
	}
	if offered["Four Drop"] {
		t.Fatalf("offered options = %v, an over-bound card was findable", agent.offered)
	}
	if !g.Players[game.Player1].Hand.Contains(atBound) || g.Players[game.Player1].Library.Contains(atBound) {
		t.Fatal("the chosen at-bound card was not moved to hand")
	}
	if !g.Players[game.Player1].Library.Contains(underBound) {
		t.Fatal("an unchosen matching card left the library")
	}
	if !g.Players[game.Player1].Library.Contains(overBound) {
		t.Fatal("an over-bound card was moved out of the library")
	}
}

// TestSearchManaValueDynamicCountBoundZeroLandsFindsNothing proves the bound is
// truly dynamic: with no lands controlled, the bound is zero, so no positive
// mana-value card qualifies and the search legally finds nothing.
func TestSearchManaValueDynamicCountBoundZeroLandsFindsNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	only := addCardToLibrary(g, game.Player1, manaValueCard("One Drop", 1))
	addInstructionSpellToStack(g, []game.Instruction{beseechSearchInstruction()})
	agent := &recordingSearchAgent{wanted: "One Drop"}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, label := range agent.offered {
		if label == "One Drop" {
			t.Fatal("a mana-value 1 card was findable with zero lands controlled")
		}
	}
	if !g.Players[game.Player1].Library.Contains(only) {
		t.Fatal("a card was moved out of the library despite the zero bound")
	}
}

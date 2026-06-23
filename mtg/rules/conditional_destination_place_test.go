package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conditionalDestinationAgent answers the reveal-only search choice by selecting
// the option whose label matches wanted, and answers the "put it onto the
// battlefield?" may-choice according to acceptPut.
type conditionalDestinationAgent struct {
	wanted    string
	acceptPut bool
}

func (conditionalDestinationAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a conditionalDestinationAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.acceptPut {
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

func scholarSequence() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone: zone.Library,
					Filter:     game.Selection{SubtypesAny: []types.Sub{types.Plains}},
					Reveal:     true,
					RevealOnly: true,
				},
				Amount:        game.Fixed(1),
				PublishLinked: game.LinkedKey("conditional-destination-card"),
			},
		},
		{
			Primitive: game.ConditionalDestinationPlace{
				Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "conditional-destination-card"},
				FromZone: zone.Library,
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					ControlComparison: opt.Val(game.ControlCountComparison{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						Left:      game.ControlPlayerAnyOpponent,
						Right:     game.ControlPlayerController,
						Op:        compare.GreaterThan,
					}),
				})}),
				EntryTapped: true,
				Else:        zone.Hand,
			},
		},
		{Primitive: game.ShuffleLibrary{Player: game.ControllerReference()}},
	}
}

func plainsCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Plains",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Plains},
	}}
}

// TestConditionalDestinationPlaceGateHoldsEntersBattlefieldTapped verifies that
// when the control-comparison gate holds and the controller accepts, the
// revealed library card enters the battlefield tapped under the controller.
func TestConditionalDestinationPlaceGateHoldsEntersBattlefieldTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plains := addCardToLibrary(g, game.Player1, plainsCard())
	addLandPermanent(g, game.Player2, "Opp Land 1")
	addLandPermanent(g, game.Player2, "Opp Land 2")
	addInstructionSpellToStack(g, scholarSequence())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: conditionalDestinationAgent{wanted: "Plains", acceptPut: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(plains) || g.Players[game.Player1].Hand.Contains(plains) {
		t.Fatal("revealed card should have left the library and not entered the hand")
	}
	var found *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == plains {
			found = permanent
		}
	}
	if found == nil {
		t.Fatal("revealed card did not enter the battlefield")
	}
	if found.Controller != game.Player1 {
		t.Fatalf("battlefield card controller = %v, want Player1", found.Controller)
	}
	if !found.Tapped {
		t.Fatal("battlefield card should have entered tapped")
	}
}

// TestConditionalDestinationPlaceGateFailsGoesToHand verifies that when the gate
// fails (no opponent controls more lands), the revealed card is put into the
// controller's hand instead of the battlefield.
func TestConditionalDestinationPlaceGateFailsGoesToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plains := addCardToLibrary(g, game.Player1, plainsCard())
	addLandPermanent(g, game.Player1, "Own Land")
	addInstructionSpellToStack(g, scholarSequence())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: conditionalDestinationAgent{wanted: "Plains", acceptPut: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(plains) {
		t.Fatal("revealed card should have been put into the hand when the gate fails")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == plains {
			t.Fatal("revealed card should not have entered the battlefield when the gate fails")
		}
	}
}

// TestConditionalDestinationPlaceDeclinePutGoesToHand verifies that when the gate
// holds but the controller declines the optional battlefield put, the revealed
// card is put into the controller's hand.
func TestConditionalDestinationPlaceDeclinePutGoesToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plains := addCardToLibrary(g, game.Player1, plainsCard())
	addLandPermanent(g, game.Player2, "Opp Land 1")
	addLandPermanent(g, game.Player2, "Opp Land 2")
	addInstructionSpellToStack(g, scholarSequence())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: conditionalDestinationAgent{wanted: "Plains", acceptPut: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(plains) {
		t.Fatal("declining the battlefield put should have put the card into the hand")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == plains {
			t.Fatal("revealed card should not have entered the battlefield after declining")
		}
	}
}

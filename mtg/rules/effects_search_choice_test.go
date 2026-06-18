package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// searchByNameAgent answers search choices by selecting the option whose label
// matches the wanted card name, or fails to find when wanted is empty.
type searchByNameAgent struct {
	wanted string
}

func (*searchByNameAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *searchByNameAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if a.wanted == "" {
		return []int{}
	}
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return nil
}

func TestSearchLibraryLetsPlayerChooseAmongMatchingCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	wolf := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wolf", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wolf"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(wolf) || g.Players[game.Player1].Library.Contains(wolf) {
		t.Fatal("search did not move the player-chosen matching card to hand")
	}
	if !g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("search moved an unchosen matching card out of the library")
	}
}

func TestSearchLibraryAllowsLegalFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: ""}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(bear) || g.Players[game.Player1].Hand.Contains(bear) {
		t.Fatal("search did not allow the player to legally fail to find a matching card")
	}
}

func TestSearchLibraryNilAgentFindsFirstMatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(creature) || g.Players[game.Player1].Library.Contains(creature) {
		t.Fatal("nil-agent search did not deterministically find the matching card")
	}
}

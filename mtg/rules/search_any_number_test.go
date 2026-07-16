package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// selectNamedAgent answers a search choice by selecting every offered option
// whose label is one of the wanted names, letting a test drive an "any number"
// search toward a specific subset of the matching cards.
type selectNamedAgent struct {
	wanted map[string]bool
}

func (selectNamedAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a selectNamedAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	indices := make([]int, 0, len(request.Options))
	for _, option := range request.Options {
		if a.wanted[option.Label] {
			indices = append(indices, option.Index)
		}
	}
	return indices
}

func godCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Enchantment, types.Creature},
		Subtypes: []types.Sub{types.God},
	}}
}

// anyNumberGodSearch is The World Tree's "Search your library for any number of
// God cards, put them onto the battlefield, then shuffle" library search.
func anyNumberGodSearch() game.Search {
	return game.Search{
		Amount: game.Fixed(0),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Battlefield,
			AnyNumber:   true,
			Filter: game.Selection{
				SubtypesAny: []types.Sub{types.God},
			},
		},
	}
}

// TestAnyNumberSearchFindsEveryMatchingCard exercises the upper bound of an
// "any number of" search: choosing all matches puts every God onto the
// battlefield while leaving non-matching cards untouched.
func TestAnyNumberSearchFindsEveryMatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gods := map[string]id.ID{
		"Heliod": addCardToLibrary(g, game.Player1, godCard("Heliod")),
		"Thassa": addCardToLibrary(g, game.Player1, godCard("Thassa")),
		"Erebos": addCardToLibrary(g, game.Player1, godCard("Erebos")),
	}
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, anyNumberGodSearch(), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllAgent{}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for name, cardID := range gods {
		if permanentForCard(g, cardID) == nil {
			t.Fatalf("any-number search choosing all did not put %s onto the battlefield", name)
		}
		if g.Players[game.Player1].Library.Contains(cardID) {
			t.Fatalf("any-number search left %s in the library", name)
		}
	}
	if permanentForCard(g, bear) != nil {
		t.Fatal("any-number search put a non-matching card onto the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("any-number search removed a non-matching card from the library")
	}
}

// TestAnyNumberSearchMayFindNone exercises the lower bound: an "any number of"
// search permits choosing zero even when matches are available (CR 701.19c),
// leaving the declined card in the library.
func TestAnyNumberSearchMayFindNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	heliod := addCardToLibrary(g, game.Player1, godCard("Heliod"))
	addEffectSpellToStack(g, game.Player1, anyNumberGodSearch(), nil)
	// searchByNameAgent with an empty wanted name declines every offered card.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: ""}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, heliod) != nil {
		t.Fatal("any-number search forced a find when the player chose none")
	}
	if !g.Players[game.Player1].Library.Contains(heliod) {
		t.Fatal("any-number search removed the declined card from the library")
	}
}

// TestAnyNumberSearchFindsChosenSubset exercises an intermediate choice: the
// player selects some but not all matches, and only the chosen cards move.
func TestAnyNumberSearchFindsChosenSubset(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	heliod := addCardToLibrary(g, game.Player1, godCard("Heliod"))
	thassa := addCardToLibrary(g, game.Player1, godCard("Thassa"))
	erebos := addCardToLibrary(g, game.Player1, godCard("Erebos"))
	addEffectSpellToStack(g, game.Player1, anyNumberGodSearch(), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectNamedAgent{wanted: map[string]bool{"Heliod": true, "Erebos": true}}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, heliod) == nil || permanentForCard(g, erebos) == nil {
		t.Fatal("any-number search did not put both chosen Gods onto the battlefield")
	}
	if permanentForCard(g, thassa) != nil {
		t.Fatal("any-number search put an unchosen God onto the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(thassa) {
		t.Fatal("any-number search removed the unchosen God from the library")
	}
}

package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// addColoredPermanent adds a battlefield permanent controlled by controller with
// the given colors, card types, and supertypes, so the among-controlled mana
// choice can union its colors.
func addColoredPermanent(g *game.Game, controller game.PlayerID, name string, colors []color.Color, cardTypes []types.Card, supertypes []types.Super) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Types:      cardTypes,
		Supertypes: supertypes,
		Colors:     colors,
	}})
}

// TestControlledPermanentColorsManaUnionsMatchingColors verifies the "any color
// among legendary creatures and planeswalkers you control" choice (Mox Amber)
// offers the union of colors of the matching permanents in WUBRG order, ignores
// non-matching permanents (wrong supertype, wrong type, opponent-controlled),
// and offers nothing when no matching permanent is colored.
func TestControlledPermanentColorsManaUnionsMatchingColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	choice := &game.ResolutionChoice{
		Kind:        game.ResolutionChoiceMana,
		ColorSource: game.ResolutionChoiceColorSourceControlledPermanentColors,
		Selection: &game.Selection{
			Controller:       game.ControllerYou,
			RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
			Supertypes:       []types.Super{types.Legendary},
		},
	}

	// No matching permanent: the choice is empty.
	if got := controlledPermanentColorsMana(g, game.Player1, choice); len(got) != 0 {
		t.Fatalf("empty board choice = %v, want empty", got)
	}

	// A legendary green creature and a legendary blue planeswalker you control:
	// the choice offers U and G (WUBRG order).
	addColoredPermanent(g, game.Player1, "Legend A", []color.Color{color.Green}, []types.Card{types.Creature}, []types.Super{types.Legendary})
	addColoredPermanent(g, game.Player1, "Legend B", []color.Color{color.Blue}, []types.Card{types.Planeswalker}, []types.Super{types.Legendary})

	// Non-matching permanents must not contribute: a nonlegendary red creature,
	// a legendary white artifact (wrong type), and an opponent's legendary
	// creature.
	addColoredPermanent(g, game.Player1, "Plain Creature", []color.Color{color.Red}, []types.Card{types.Creature}, nil)
	addColoredPermanent(g, game.Player1, "Legend Artifact", []color.Color{color.White}, []types.Card{types.Artifact}, []types.Super{types.Legendary})
	addColoredPermanent(g, game.Player2, "Foe Legend", []color.Color{color.Black}, []types.Card{types.Creature}, []types.Super{types.Legendary})

	got := controlledPermanentColorsMana(g, game.Player1, choice)
	if want := []mana.Color{mana.U, mana.G}; !slices.Equal(got, want) {
		t.Fatalf("among-controlled colors = %v, want %v", got, want)
	}
}

// TestControlledPermanentColorsManaColorlessOffersNothing verifies a colorless
// matching permanent contributes no colors, leaving the choice empty (a
// permanent's colors are only the five colors).
func TestControlledPermanentColorsManaColorlessOffersNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	choice := &game.ResolutionChoice{
		Kind:        game.ResolutionChoiceMana,
		ColorSource: game.ResolutionChoiceColorSourceControlledPermanentColors,
		Selection: &game.Selection{
			Controller: game.ControllerYou,
			Supertypes: []types.Super{types.Legendary},
		},
	}
	addColoredPermanent(g, game.Player1, "Colorless Legend", nil, []types.Card{types.Artifact}, []types.Super{types.Legendary})
	if got := controlledPermanentColorsMana(g, game.Player1, choice); len(got) != 0 {
		t.Fatalf("colorless permanent choice = %v, want empty", got)
	}
}

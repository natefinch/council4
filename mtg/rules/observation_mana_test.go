package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCardFaceManaColorsReadsFixedTapAbilities(t *testing.T) {
	dual := &game.CardFace{
		Name:  "Tundra",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.W),
			game.TapManaAbility(mana.U),
		},
	}
	got := cardFaceManaColors(dual, nil)
	want := []color.Color{color.White, color.Blue}
	if !slices.Equal(got, want) {
		t.Fatalf("cardFaceManaColors = %v, want %v", got, want)
	}
}

func TestCardFaceManaColorsIgnoresColorlessOnlyProduction(t *testing.T) {
	rock := &game.CardFace{
		Name:          "Colorless Rock",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}
	if got := cardFaceManaColors(rock, nil); got != nil {
		t.Fatalf("cardFaceManaColors = %v, want nil for colorless-only source", got)
	}
}

func TestCardFaceManaColorsResolvesCommanderIdentity(t *testing.T) {
	tower := &game.CardFace{
		Name:          "Command Tower",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaCommanderIdentityAbility()},
	}

	got := cardFaceManaColors(tower, []color.Color{color.Blue, color.Green})
	if want := []color.Color{color.Blue, color.Green}; !slices.Equal(got, want) {
		t.Fatalf("cardFaceManaColors = %v, want commander identity %v", got, want)
	}

	if got := cardFaceManaColors(tower, nil); got != nil {
		t.Fatalf("cardFaceManaColors = %v, want no colors for a colorless/absent commander identity", got)
	}
}

func TestCardFaceManaColorsReadsAnyColorChoice(t *testing.T) {
	anyColor := &game.CardFace{
		Name:  "Five-Color Land",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{{
			Text: "{T}: Add one mana of any color.",
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Choose{
					Choice: game.ResolutionChoice{
						Kind:   game.ResolutionChoiceMana,
						Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
					},
				},
			}}}.Ability(),
		}},
	}
	got := cardFaceManaColors(anyColor, nil)
	if !slices.Equal(got, color.AllColors()) {
		t.Fatalf("cardFaceManaColors = %v, want all five colors", got)
	}
}

func TestPermanentChosenColorManaProductionResolvesEntryChoice(t *testing.T) {
	body := game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")
	producesMana, colors := abilitiesManaProduction(
		[]game.Ability{&body},
		map[game.ChoiceKey]game.ResolutionChoiceResult{
			game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: mana.R},
		},
		nil,
	)
	if !producesMana {
		t.Fatal("chosen-color source should report ProducesMana")
	}
	if !slices.Equal(colors, []color.Color{color.Red}) {
		t.Fatalf("abilitiesManaProduction colors = %v, want [Red]", colors)
	}
}

func TestPermanentChosenColorManaProductionWithoutChoiceYieldsNoColor(t *testing.T) {
	body := game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color.")
	producesMana, colors := abilitiesManaProduction([]game.Ability{&body}, nil, nil)
	if !producesMana {
		t.Fatal("chosen-color source should still report ProducesMana")
	}
	if colors != nil {
		t.Fatalf("abilitiesManaProduction colors = %v, want nil without an entry choice", colors)
	}
}

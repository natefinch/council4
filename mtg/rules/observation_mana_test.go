package rules

import (
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
	got := cardFaceManaColors(dual)
	want := []color.Color{color.White, color.Blue}
	if !equalColors(got, want) {
		t.Fatalf("cardFaceManaColors = %v, want %v", got, want)
	}
}

func TestCardFaceManaColorsIgnoresColorlessOnlyProduction(t *testing.T) {
	rock := &game.CardFace{
		Name:          "Colorless Rock",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}
	if got := cardFaceManaColors(rock); got != nil {
		t.Fatalf("cardFaceManaColors = %v, want nil for colorless-only source", got)
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
	got := cardFaceManaColors(anyColor)
	if !equalColors(got, color.AllColors()) {
		t.Fatalf("cardFaceManaColors = %v, want all five colors", got)
	}
}

func equalColors(got, want []color.Color) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

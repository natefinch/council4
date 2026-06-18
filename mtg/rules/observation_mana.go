package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// cardFaceManaColors returns the colors of mana the face's mana abilities can
// add, deduplicated and in WUBRG order. Colorless-only production yields no
// colors. It lets an observation report a land's color fixing without exposing
// ability internals to agents.
func cardFaceManaColors(face *game.CardFace) []color.Color {
	var found colorSet
	for i := range face.ManaAbilities {
		manaAbilityColors(&face.ManaAbilities[i], &found)
	}
	return found.ordered()
}

// abilitiesManaProduction reports whether any of the effective abilities is a
// mana ability and the colors those mana abilities can add. Granted mana
// abilities are included because the input is the permanent's effective ability
// set.
func abilitiesManaProduction(abilities []game.Ability) (bool, []color.Color) {
	var found colorSet
	producesMana := false
	for _, ability := range abilities {
		body, ok := ability.(game.ManaAbility)
		if !ok {
			continue
		}
		producesMana = true
		manaAbilityColors(&body, &found)
	}
	return producesMana, found.ordered()
}

// manaAbilityColors adds to found every color the mana ability can add, reading
// fixed AddMana colors and the color options of any mana-color choice it makes.
func manaAbilityColors(body *game.ManaAbility, found *colorSet) {
	if len(body.Content.Modes) == 0 {
		return
	}
	for m := range body.Content.Modes {
		sequence := body.Content.Modes[m].Sequence
		for i := range sequence {
			addInstructionManaColors(sequence[i].Primitive, found)
		}
	}
}

func addInstructionManaColors(primitive game.Primitive, found *colorSet) {
	if primitive == nil {
		return
	}
	switch primitive.Kind() {
	case game.PrimitiveAddMana:
		add, ok := primitive.(game.AddMana)
		if !ok {
			return
		}
		if c, ok := manaColor(add.ManaColor); ok {
			found.add(c)
		}
	case game.PrimitiveChoose:
		choose, ok := primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
			return
		}
		if choose.Choice.ColorSource != game.ResolutionChoiceColorSourceStatic {
			// "any color" and commander-identity choices can yield any color.
			for _, c := range color.AllColors() {
				found.add(c)
			}
			return
		}
		for _, mc := range choose.Choice.Colors {
			if c, ok := manaColor(mc); ok {
				found.add(c)
			}
		}
	default:
	}
}

// manaColor maps a mana color to its card color, reporting false for colorless.
func manaColor(mc mana.Color) (color.Color, bool) {
	switch mc {
	case mana.W:
		return color.White, true
	case mana.U:
		return color.Blue, true
	case mana.B:
		return color.Black, true
	case mana.R:
		return color.Red, true
	case mana.G:
		return color.Green, true
	default:
		return "", false
	}
}

// colorSet collects colors while preserving a stable WUBRG output order.
type colorSet struct {
	present map[color.Color]bool
}

func (s *colorSet) add(c color.Color) {
	if s.present == nil {
		s.present = make(map[color.Color]bool, len(color.AllColors()))
	}
	s.present[c] = true
}

func (s *colorSet) ordered() []color.Color {
	if len(s.present) == 0 {
		return nil
	}
	ordered := make([]color.Color, 0, len(s.present))
	for _, c := range color.AllColors() {
		if s.present[c] {
			ordered = append(ordered, c)
		}
	}
	return ordered
}

package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// cardFaceManaColors returns the colors of mana the face's mana abilities can
// add, deduplicated and in WUBRG order. Colorless-only production yields no
// colors. It lets an observation report a land's color fixing without exposing
// ability internals to agents. A "{T}: Add one mana of the chosen color" source
// yields no colors here because the chosen color is unknowable until the
// permanent enters; abilitiesManaProduction resolves it for permanents.
func cardFaceManaColors(face *game.CardFace) []color.Color {
	var found colorSet
	for i := range face.ManaAbilities {
		manaAbilityColors(&face.ManaAbilities[i], nil, &found)
	}
	return found.ordered()
}

// abilitiesManaProduction reports whether any of the effective abilities is a
// mana ability and the colors those mana abilities can add. Granted mana
// abilities are included because the input is the permanent's effective ability
// set. entryChoices resolves entry-time color choices (CR 614.12) so a "{T}: Add
// one mana of the chosen color" source reports its chosen color.
func abilitiesManaProduction(abilities []game.Ability, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) (bool, []color.Color) {
	var found colorSet
	producesMana := false
	for _, ability := range abilities {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		producesMana = true
		manaAbilityColors(body, entryChoices, &found)
	}
	return producesMana, found.ordered()
}

// manaAbilityColors adds to found every color the mana ability can add, reading
// fixed AddMana colors, entry-time chosen colors, and the color options of any
// mana-color choice it makes.
func manaAbilityColors(body *game.ManaAbility, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult, found *colorSet) {
	if len(body.Content.Modes) == 0 {
		return
	}
	for m := range body.Content.Modes {
		sequence := body.Content.Modes[m].Sequence
		for i := range sequence {
			addInstructionManaColors(sequence[i].Primitive, entryChoices, found)
		}
	}
}

func addInstructionManaColors(primitive game.Primitive, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult, found *colorSet) {
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
		if add.EntryChoiceFrom != "" {
			if result, ok := entryChoices[add.EntryChoiceFrom]; ok {
				if c, ok := manaColor(result.Color); ok {
					found.add(c)
				}
			}
		}
	case game.PrimitiveChoose:
		choose, ok := primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
			return
		}
		if choose.Choice.ColorSource == game.ResolutionChoiceColorSourceLandsProduce {
			// The colors come from other lands' production. Such a derived
			// ability contributes no colors of its own to this report, matching
			// CR's loop-avoidance ruling and keeping landsProduceMana's scan from
			// recursing into another "lands could produce" source.
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

// abilitiesProduceColorless reports whether any of the given mana abilities
// could add colorless ({C}) mana. It reads fixed AddMana colorless output,
// entry-chosen colorless, and static mana-color choices whose options include
// colorless. Dynamic color sources contribute no colorless: "any color" and
// commander-identity choices yield only colored mana (CR 106.5), and a
// lands-produce source contributes nothing here to avoid recursion. It supports
// the "any type" lands-produce wording (Reflecting Pool), which offers colorless
// when a scanned land could produce it.
func abilitiesProduceColorless(abilities []game.Ability, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	for _, ability := range abilities {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		if manaAbilityProducesColorless(body, entryChoices) {
			return true
		}
	}
	return false
}

// manaAbilityProducesColorless reports whether the mana ability could add
// colorless mana, inspecting each AddMana and static mana-color Choose in its
// instruction sequence.
func manaAbilityProducesColorless(body *game.ManaAbility, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	if len(body.Content.Modes) == 0 {
		return false
	}
	for m := range body.Content.Modes {
		sequence := body.Content.Modes[m].Sequence
		for i := range sequence {
			if instructionProducesColorless(sequence[i].Primitive, entryChoices) {
				return true
			}
		}
	}
	return false
}

func instructionProducesColorless(primitive game.Primitive, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	if primitive == nil {
		return false
	}
	switch primitive.Kind() {
	case game.PrimitiveAddMana:
		add, ok := primitive.(game.AddMana)
		if !ok {
			return false
		}
		if add.ManaColor == mana.C {
			return true
		}
		if add.EntryChoiceFrom != "" {
			if result, ok := entryChoices[add.EntryChoiceFrom]; ok && result.Color == mana.C {
				return true
			}
		}
		return false
	case game.PrimitiveChoose:
		choose, ok := primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana ||
			choose.Choice.ColorSource != game.ResolutionChoiceColorSourceStatic {
			return false
		}
		return slices.Contains(choose.Choice.Colors, mana.C)
	default:
		return false
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

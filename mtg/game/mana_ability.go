package game

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game/mana"
)

var anyColorManaChoices = []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G}

// TapAnyColorManaAbility builds the exact tap-for-one-mana-of-any-color ability.
func TapAnyColorManaAbility() ManaAbility {
	ability := TapManaChoiceAbility(anyColorManaChoices...)
	ability.Text = "{T}: Add one mana of any color."
	return ability
}

// ManaAbilityChoiceOutput reports the fixed amount and color choices of a
// canonical choose-then-add mana ability. The returned colors are immutable
// ability data and must be treated as read-only.
func ManaAbilityChoiceOutput(body *ManaAbility) ([]mana.Color, int, bool) {
	if body == nil ||
		len(body.Content.SharedTargets) != 0 ||
		body.Content.IsModal() ||
		len(body.Content.Modes) != 1 ||
		len(body.Content.Modes[0].Targets) != 0 ||
		len(body.Content.Modes[0].Sequence) != 2 {
		return nil, 0, false
	}
	sequence := body.Content.Modes[0].Sequence
	choose, ok := sequence[0].Primitive.(Choose)
	if !ok ||
		choose.Choice.Kind != ResolutionChoiceMana ||
		len(choose.Choice.Colors) < 2 ||
		len(choose.Choice.Colors) > 6 ||
		choose.PublishChoice == "" {
		return nil, 0, false
	}
	for i, color := range choose.Choice.Colors {
		switch color {
		case mana.W, mana.U, mana.B, mana.R, mana.G, mana.C:
		default:
			return nil, 0, false
		}
		if slices.Contains(choose.Choice.Colors[:i], color) {
			return nil, 0, false
		}
	}
	add, ok := sequence[1].Primitive.(AddMana)
	if !ok ||
		add.Amount.IsDynamic() ||
		add.Amount.Value() <= 0 ||
		add.ManaColor != "" ||
		add.ChoiceFrom != choose.PublishChoice ||
		add.EntryChoiceFrom != "" ||
		add.SpendRider.Exists {
		return nil, 0, false
	}
	return choose.Choice.Colors, add.Amount.Value(), true
}

// IsTapAnyColorManaAbility reports whether body is the canonical exact
// tap-for-one-mana-of-any-color ability.
func IsTapAnyColorManaAbility(body *ManaAbility) bool {
	return body != nil && reflect.DeepEqual(*body, TapAnyColorManaAbility())
}

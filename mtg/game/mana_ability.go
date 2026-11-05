package game

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

// IsTapColorlessManaAbility reports whether body is the bare "{T}: Add {C}"
// ability that adds one colorless mana, granted by removal Auras such as
// Imprisoned in the Moon.
func IsTapColorlessManaAbility(body *ManaAbility) bool {
	return body != nil && reflect.DeepEqual(*body, TapManaAbility(mana.C))
}

// IsTapOneColorManaAbility reports whether body is a bare "{T}: Add {<color>}"
// ability that adds one mana of a single fixed color (W/U/B/R/G), as granted to
// a group ("Each creature you control with a counter on it has '{T}: Add {G}.'",
// Rishkar). It is the colored counterpart of IsTapColorlessManaAbility and runs
// through the identical bare tap-for-one-mana mechanism.
func IsTapOneColorManaAbility(body *ManaAbility) bool {
	if body == nil {
		return false
	}
	for _, color := range anyColorManaChoices {
		if reflect.DeepEqual(*body, TapManaAbility(color)) {
			return true
		}
	}
	return false
}

// TapSacrificeAnyOneColorManaAbility builds the Treasure-style granted mana
// ability "{T}, Sacrifice this artifact: Add <count> mana of any one color."
// (Goldspan Dragon, Alchemist's Talent): tap and sacrifice the host artifact to
// add count mana (count >= 2) of one color the controller chooses. text is the
// ability's printed wording, passed through so the rendered ability matches.
func TapSacrificeAnyOneColorManaAbility(text string, count int) ManaAbility {
	ability := TapManaChoiceCountAbility(text, count, anyColorManaChoices...)
	ability.AdditionalCosts = append(slices.Clone(ability.AdditionalCosts), sacrificeThisArtifactCost())
	return ability
}

func sacrificeThisArtifactCost() cost.Additional {
	return cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}
}

// IsTapSacrificeAnyOneColorManaAbility reports whether body is the Treasure-style
// granted sacrifice mana ability that adds N (>= 2) mana of one chosen color
// among the five colors.
func IsTapSacrificeAnyOneColorManaAbility(body *ManaAbility) bool {
	if body == nil {
		return false
	}
	colors, count, ok := ManaAbilityChoiceOutput(body)
	if !ok || count < 2 || !slices.Equal(colors, anyColorManaChoices) {
		return false
	}
	return reflect.DeepEqual(*body, TapSacrificeAnyOneColorManaAbility(body.Text, count))
}

// TapSacrificeAnyColorManaAbility builds the granted mana ability
// "{T}, Sacrifice this artifact: Add one mana of any color." (Ninja Pizza): tap
// and sacrifice the host artifact to add one mana of any one of the five colors
// the controller chooses. It is the count-1 any-color counterpart of the
// Treasure-style sacrifice ability. text is the ability's printed wording,
// passed through so the rendered ability matches.
func TapSacrificeAnyColorManaAbility(text string) ManaAbility {
	ability := TapAnyColorManaAbility()
	ability.Text = text
	ability.AdditionalCosts = append(slices.Clone(ability.AdditionalCosts), cost.Additional{
		Kind:   cost.AdditionalSacrificeSource,
		Text:   "Sacrifice this artifact",
		Amount: 1,
	})
	return ability
}

// IsTapSacrificeAnyColorManaAbility reports whether body is the count-1
// tap-and-sacrifice granted mana ability that adds one mana of any color
// (Ninja Pizza).
func IsTapSacrificeAnyColorManaAbility(body *ManaAbility) bool {
	if body == nil {
		return false
	}
	return reflect.DeepEqual(*body, TapSacrificeAnyColorManaAbility(body.Text))
}

// TapAnyColorCreatureSpellRestrictedManaAbility builds the granted mana ability
// "{T}: Add one mana of any color. Spend this mana only to cast a creature
// spell." (granted to each creature you control by Inga and Esika). It is the
// tap-for-one-mana-of-any-color ability whose produced mana is restricted to
// paying for creature spells. text is the ability's printed wording, passed
// through so the rendered granted ability matches.
func TapAnyColorCreatureSpellRestrictedManaAbility(text string) ManaAbility {
	rider := ManaSpendRider{
		Condition:   ManaSpendCastCreatureSpell,
		Restriction: ManaSpendRestrictedToCondition,
	}
	return TapManaChoiceWithSpendRiderAbility(text, rider, anyColorManaChoices...)
}

// IsTapAnyColorCreatureSpellRestrictedManaAbility reports whether body is the
// granted "{T}: Add one mana of any color. Spend this mana only to cast a
// creature spell." ability (Inga and Esika). It is the spend-restricted
// counterpart of IsTapAnyColorManaAbility.
func IsTapAnyColorCreatureSpellRestrictedManaAbility(body *ManaAbility) bool {
	if body == nil {
		return false
	}
	return reflect.DeepEqual(*body, TapAnyColorCreatureSpellRestrictedManaAbility(body.Text))
}

// ManaAbilityLibraryMovement reports why body can't be a mana ability under
// CR 605.1a: an activated ability is a mana ability only if "its cost and
// effect don't move any card to or from a library" (replacement effects other
// than self-replacement effects aren't considered). Such an ability, like
// Millikin's "{T}, Mill a card: Add {C}.", is an ActivatedAbility that uses the
// stack. A cost is judged by its kind and card zone. An effect is judged by
// its instructions' primitives, which must each be one known to leave every
// library alone; any other primitive is reported so its movement is decided
// before it is allowed in a mana ability. It returns "" when body qualifies.
func ManaAbilityLibraryMovement(body *ManaAbility) string {
	if body == nil {
		return ""
	}
	for _, additional := range body.AdditionalCosts {
		if additional.Kind == cost.AdditionalMill || additional.Source == zone.Library && additional.Kind != cost.AdditionalReveal {
			return "its cost moves a card from a library"
		}
	}
	for _, mode := range body.Content.Modes {
		for i := range mode.Sequence {
			instruction := &mode.Sequence[i]
			if len(instruction.TemptingOfferBody) != 0 {
				return "its effect has a tempting-offer body not classified as leaving libraries alone"
			}
			if instruction.Primitive != nil && !primitiveLeavesLibraries(instruction.Primitive) {
				return fmt.Sprintf("its effect primitive %T isn't classified as leaving libraries alone", instruction.Primitive)
			}
		}
	}
	return ""
}

// primitiveLeavesLibraries reports whether primitive is known never to move a
// card to or from a library, apart from replacement effects (CR 605.1a).
func primitiveLeavesLibraries(primitive Primitive) bool {
	switch primitive.(type) {
	case AddMana, Choose, Damage, GainLife, LoseLife, SkipNextUntap, Tap, Untap, AddCounter, RemoveCounter:
		return true
	}
	return false
}

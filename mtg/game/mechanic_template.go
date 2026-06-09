package game

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const tapManaChoiceKey = ChoiceKey("oracle-mana-color")

// CantBlockStaticBody is the complete static ability for a creature that cannot block.
var CantBlockStaticBody = StaticAbility{
	Text: "This creature can't block.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBlock,
		AffectedSource: true,
	}},
}

// CantBeCounteredStaticBody is the complete static ability for an uncounterable spell.
var CantBeCounteredStaticBody = StaticAbility{
	Text:           "This spell can't be countered.",
	ZoneOfFunction: zone.Stack,
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBeCountered,
		AffectedSource: true,
	}},
}

// WardStaticAbility builds the complete static ability for Ward with a mana cost.
func WardStaticAbility(manaCost cost.Mana) StaticAbility {
	keywordCost := append(cost.Mana(nil), manaCost...)
	return StaticAbility{
		Text: "Ward " + manaCost.String(),
		KeywordAbilities: []KeywordAbility{
			WardKeyword{Cost: keywordCost},
		},
	}
}

// EnchantStaticAbility builds the complete static ability for Enchant.
func EnchantStaticAbility(target *TargetSpec) StaticAbility {
	targetCopy := cloneTargetSpec(target)
	return StaticAbility{
		Text: "Enchant " + targetCopy.Constraint,
		KeywordAbilities: []KeywordAbility{
			EnchantKeyword{Target: targetCopy},
		},
	}
}

func cloneTargetSpec(source *TargetSpec) TargetSpec {
	target := *source
	target.Predicate.PermanentTypes = append([]types.Card(nil), target.Predicate.PermanentTypes...)
	target.Predicate.ExcludedTypes = append([]types.Card(nil), target.Predicate.ExcludedTypes...)
	target.Predicate.Colors = append([]color.Color(nil), target.Predicate.Colors...)
	target.Predicate.ExcludedColors = append([]color.Color(nil), target.Predicate.ExcludedColors...)
	if target.Selection.Exists {
		selection := target.Selection.Val
		selection.RequiredTypes = append([]types.Card(nil), selection.RequiredTypes...)
		selection.RequiredTypesAny = append([]types.Card(nil), selection.RequiredTypesAny...)
		selection.ExcludedTypes = append([]types.Card(nil), selection.ExcludedTypes...)
		selection.Supertypes = append([]types.Super(nil), selection.Supertypes...)
		selection.SubtypesAny = append([]types.Sub(nil), selection.SubtypesAny...)
		selection.ColorsAny = append([]color.Color(nil), selection.ColorsAny...)
		selection.ExcludedColors = append([]color.Color(nil), selection.ExcludedColors...)
		target.Selection = opt.Val(selection)
	}
	return target
}

// ProtectionFromColorsStaticAbility builds the complete static ability for
// protection from one or more colors.
func ProtectionFromColorsStaticAbility(colors ...color.Color) StaticAbility {
	protectedColors := append([]color.Color(nil), colors...)
	validateProtectionColors(protectedColors)
	return StaticAbility{
		Text: protectionFromColorsText(protectedColors),
		KeywordAbilities: []KeywordAbility{
			ProtectionKeyword{FromColors: protectedColors},
		},
	}
}

func validateProtectionColors(colors []color.Color) {
	if len(colors) == 0 {
		panic("game: protection requires at least one color")
	}
	seen := make(map[color.Color]struct{}, len(colors))
	for _, protectedColor := range colors {
		switch protectedColor {
		case color.White, color.Blue, color.Black, color.Red, color.Green:
		default:
			panic(fmt.Sprintf("game: invalid protection color %q", protectedColor))
		}
		if _, ok := seen[protectedColor]; ok {
			panic(fmt.Sprintf("game: duplicate protection color %q", protectedColor))
		}
		seen[protectedColor] = struct{}{}
	}
}

func protectionFromColorsText(colors []color.Color) string {
	phrases := make([]string, len(colors))
	for i, protectedColor := range colors {
		phrases[i] = "from " + strings.ToLower(string(protectedColor))
	}
	switch len(phrases) {
	case 1:
		return "Protection " + phrases[0]
	case 2:
		return "Protection " + phrases[0] + " and " + phrases[1]
	default:
		return "Protection " +
			strings.Join(phrases[:len(phrases)-1], ", ") +
			", and " +
			phrases[len(phrases)-1]
	}
}

// CyclingActivatedAbility builds the complete activated ability for Cycling
// with a mana cost.
func CyclingActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:           "Cycling " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Hand,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalDiscard,
			Text:   "Discard this card",
			Amount: 1,
			Source: zone.Hand,
		}},
		KeywordAbilities: []KeywordAbility{
			CyclingKeyword{Cost: keywordCost},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: Draw{
				Amount: Fixed(1),
				Player: ControllerReference(),
			},
		}}}.Ability(),
	}
}

// EquipActivatedAbility builds the complete activated ability for Equip with a
// mana cost.
func EquipActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:           "Equip " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Battlefield,
		Timing:         SorceryOnly,
		KeywordAbilities: []KeywordAbility{
			EquipKeyword{Cost: keywordCost},
		},
		Content: Mode{Targets: []TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "creature you control",
			Allow:      TargetAllowPermanent,
			Predicate: TargetPredicate{
				PermanentTypes: []types.Card{types.Creature},
				Controller:     ControllerYou,
			},
		}}}.Ability(),
	}
}

// TapManaAbility builds the complete "{T}: Add {X}." mana ability.
func TapManaAbility(manaColor mana.Color) ManaAbility {
	return ManaAbility{
		Text:            fmt.Sprintf("{T}: Add {%s}.", manaSymbol(manaColor)),
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{{
			Primitive: AddMana{
				Amount:    Fixed(1),
				ManaColor: manaColor,
			},
		}}}.Ability(),
	}
}

func manaSymbol(manaColor mana.Color) string {
	switch manaColor {
	case mana.W, mana.U, mana.B, mana.R, mana.G:
		return string(manaColor)
	case mana.C:
		return "C"
	default:
		panic(fmt.Sprintf("game: invalid mana color %q", manaColor))
	}
}

// TapManaChoiceAbility builds the complete tap ability for adding one mana
// chosen from two through five colors.
func TapManaChoiceAbility(colors ...mana.Color) ManaAbility {
	manaColors := append([]mana.Color(nil), colors...)
	validateManaColorChoice(manaColors)
	prompt := "Choose a color"
	if containsManaColor(manaColors, mana.C) {
		prompt = "Choose a type of mana"
	}
	return ManaAbility{
		Text:            tapManaChoiceText(manaColors),
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: prompt,
						Colors: manaColors,
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

func validateManaColorChoice(colors []mana.Color) {
	if len(colors) < 2 || len(colors) > 6 {
		panic("game: tap mana choice requires two through six mana types")
	}
	seen := make(map[mana.Color]struct{}, len(colors))
	for _, manaColor := range colors {
		switch manaColor {
		case mana.W, mana.U, mana.B, mana.R, mana.G, mana.C:
		default:
			panic(fmt.Sprintf("game: invalid mana color choice %q", manaColor))
		}
		if _, ok := seen[manaColor]; ok {
			panic(fmt.Sprintf("game: duplicate mana color choice %q", manaColor))
		}
		seen[manaColor] = struct{}{}
	}
}

func tapManaChoiceText(colors []mana.Color) string {
	if len(colors) == 5 &&
		colors[0] == mana.W &&
		colors[1] == mana.U &&
		colors[2] == mana.B &&
		colors[3] == mana.R &&
		colors[4] == mana.G {
		return "{T}: Add one mana of any color."
	}
	symbols := make([]string, len(colors))
	for i, manaColor := range colors {
		symbols[i] = fmt.Sprintf("{%s}", manaSymbol(manaColor))
	}
	if len(symbols) == 2 {
		return fmt.Sprintf("{T}: Add %s or %s.", symbols[0], symbols[1])
	}
	return fmt.Sprintf(
		"{T}: Add %s, or %s.",
		strings.Join(symbols[:len(symbols)-1], ", "),
		symbols[len(symbols)-1],
	)
}

func containsManaColor(colors []mana.Color, want mana.Color) bool {
	return slices.Contains(colors, want)
}

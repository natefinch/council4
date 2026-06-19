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

const tapManaCommanderColorKey = ChoiceKey("oracle-commander-color")

// tapManaFilterFirstKey and tapManaFilterSecondKey publish the two independent
// color choices of a filter-land mana ability (see TwoColorFilterManaAbility).
// They are distinct so the instruction sequence publishes each choice under its
// own key (CR 608.2/duplicate-key validation).
const tapManaFilterFirstKey = ChoiceKey("oracle-filter-mana-first")

const tapManaFilterSecondKey = ChoiceKey("oracle-filter-mana-second")

// CantBlockStaticBody is the complete static ability for a creature that cannot block.
var CantBlockStaticBody = StaticAbility{
	Text: "This creature can't block.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBlock,
		AffectedSource: true,
	}},
}

// CantAttackStaticBody is the complete static ability for a creature that cannot attack.
var CantAttackStaticBody = StaticAbility{
	Text: "This creature can't attack.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantAttack,
		AffectedSource: true,
	}},
}

// MustBeBlockedStaticBody is the complete static ability for a creature that
// must be blocked if able.
var MustBeBlockedStaticBody = StaticAbility{
	Text: "This creature must be blocked if able.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectMustBeBlocked,
		AffectedSource: true,
	}},
}

// CantBeBlockedStaticBody is the complete static ability for an unblockable creature.
var CantBeBlockedStaticBody = StaticAbility{
	Text: "This creature can't be blocked.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBeBlocked,
		AffectedSource: true,
	}},
}

// MustAttackStaticBody is the complete static ability for a creature that must attack.
var MustAttackStaticBody = StaticAbility{
	Text: "This creature attacks each combat if able.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectMustAttack,
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

// DoesntUntapStaticBody is the complete static ability for a permanent that does
// not untap during its controller's untap step.
var DoesntUntapStaticBody = StaticAbility{
	Text: "This permanent doesn't untap during your untap step.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectDoesntUntap,
		AffectedSource: true,
	}},
}

// CantAttackOrBlockStaticBody is the complete static ability for a creature that
// can neither attack nor block.
var CantAttackOrBlockStaticBody = StaticAbility{
	Text: "This creature can't attack or block.",
	RuleEffects: []RuleEffect{
		{Kind: RuleEffectCantAttack, AffectedSource: true},
		{Kind: RuleEffectCantBlock, AffectedSource: true},
	},
}

// NoMaximumHandSizeStaticBody is the complete static ability for "You have no
// maximum hand size." The controller never discards down to a hand-size limit.
var NoMaximumHandSizeStaticBody = StaticAbility{
	Text: "You have no maximum hand size.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectNoMaximumHandSize,
		AffectedPlayer: PlayerYou,
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
	target.Predicate.Supertypes = append([]types.Super(nil), target.Predicate.Supertypes...)
	target.Predicate.Subtypes = append([]types.Sub(nil), target.Predicate.Subtypes...)
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

// ProtectionFromTypesStaticAbility builds the static ability for protection
// from one or more card types.
func ProtectionFromTypesStaticAbility(cardTypes ...types.Card) StaticAbility {
	ts := append([]types.Card(nil), cardTypes...)
	if len(ts) == 0 {
		panic("game: protection from types requires at least one type")
	}
	return StaticAbility{
		Text:             protectionFromTypesText(ts),
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{FromTypes: ts}},
	}
}

// ProtectionFromSubtypesStaticAbility builds the static ability for protection
// from one or more creature/land subtypes.
func ProtectionFromSubtypesStaticAbility(subtypes ...types.Sub) StaticAbility {
	ss := append([]types.Sub(nil), subtypes...)
	if len(ss) == 0 {
		panic("game: protection from subtypes requires at least one subtype")
	}
	return StaticAbility{
		Text:             protectionFromSubtypesText(ss),
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{FromSubtypes: ss}},
	}
}

// ProtectionFromEverythingStaticAbility builds the static ability for
// protection from everything.
func ProtectionFromEverythingStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from everything",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Everything: true}},
	}
}

// ProtectionFromEachColorStaticAbility builds the static ability for
// protection from each color.
func ProtectionFromEachColorStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from each color",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{EachColor: true}},
	}
}

// ProtectionFromMulticoloredStaticAbility builds the static ability for
// protection from multicolored sources.
func ProtectionFromMulticoloredStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from multicolored",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Multicolored: true}},
	}
}

// ProtectionFromMonocoloredStaticAbility builds the static ability for
// protection from monocolored sources.
func ProtectionFromMonocoloredStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from monocolored",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Monocolored: true}},
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

func protectionFromTypesText(cardTypes []types.Card) string {
	phrases := make([]string, len(cardTypes))
	for i, t := range cardTypes {
		phrases[i] = "from " + strings.ToLower(string(t)) + "s"
	}
	return "Protection " + joinProtectionPhrases(phrases)
}

func protectionFromSubtypesText(subtypes []types.Sub) string {
	phrases := make([]string, len(subtypes))
	for i, s := range subtypes {
		phrases[i] = "from " + strings.ToLower(string(s)) + "s"
	}
	return "Protection " + joinProtectionPhrases(phrases)
}

func joinProtectionPhrases(phrases []string) string {
	switch len(phrases) {
	case 1:
		return phrases[0]
	case 2:
		return phrases[0] + " and " + phrases[1]
	default:
		return strings.Join(phrases[:len(phrases)-1], ", ") + ", and " + phrases[len(phrases)-1]
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

// NinjutsuActivatedAbility builds the complete hand-zone activation template
// for Ninjutsu with a mana cost.
func NinjutsuActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:           "Ninjutsu " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Hand,
		Timing:         DuringCombat,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalReturnUnblockedAttacker,
			Text:   "Return an unblocked attacker you control to its owner's hand",
			Amount: 1,
		}},
		KeywordAbilities: []KeywordAbility{
			NinjutsuKeyword{Cost: keywordCost},
		},
		Content: Mode{}.Ability(),
	}
}

// MutateStaticAbility builds the hand-zone keyword ability for Mutate.
func MutateStaticAbility(manaCost cost.Mana) StaticAbility {
	keywordCost := append(cost.Mana(nil), manaCost...)
	return StaticAbility{
		Text:           "Mutate " + manaCost.String(),
		ZoneOfFunction: zone.Hand,
		KeywordAbilities: []KeywordAbility{
			MutateKeyword{Cost: keywordCost},
		},
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

// TapChosenColorManaAbility builds the complete tap ability for "{T}: Add one
// mana of the chosen color." The color is read from the entry-time choice stored
// on the source permanent under EntryColorChoiceKey, so this ability prompts no
// choice of its own.
func TapChosenColorManaAbility(text string) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: AddMana{
					Amount:          Fixed(1),
					EntryChoiceFrom: EntryColorChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapFixedOrChosenColorManaAbility builds the complete tap ability for the
// composite "{T}: Add {C} or one mana of the chosen color." (the Gate/Thriving
// land cycle). On activation the controller chooses between the fixed color and
// the color chosen as the source permanent entered (read from EntryColorChoiceKey
// seeded on the resolving ability); one mana of the selected color is added.
func TapFixedOrChosenColorManaAbility(text string, fixed mana.Color) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:           ResolutionChoiceMana,
						Prompt:         "Choose a color",
						Colors:         []mana.Color{fixed},
						ColorSource:    ResolutionChoiceColorSourceFixedOrEntryChosen,
						EntryChoiceKey: EntryColorChoiceKey,
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

// TapManaCommanderIdentityAbility builds the complete "{T}: Add one mana of any
// color in your commander's color identity." mana ability (CR 903.4). The
// choosable colors are resolved dynamically from the controller's commander
// color identity at activation; the ability is unactivatable when that identity
// is empty.
func TapManaCommanderIdentityAbility() ManaAbility {
	return ManaAbility{
		Text:            "{T}: Add one mana of any color in your commander's color identity.",
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:        ResolutionChoiceMana,
						Prompt:      "Choose a color in your commander's color identity",
						ColorSource: ResolutionChoiceColorSourceCommanderIdentity,
					},
					PublishChoice: tapManaCommanderColorKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaCommanderColorKey,
				},
			},
		}}.Ability(),
	}
}

// TwoColorFilterManaAbility builds the activated mana ability shared by the
// "filter land" cycle (Mystic Gate, Sunken Ruins, Fetid Heath, Cascade Bluffs,
// Rugged Prairie, Graven Cairns, Twilight Mire, Wooded Bastion, Fire-Lit
// Thicket, and Flooded Grove). Their second ability reads
// "{X/Y}, {T}: Add {X}{X}, {X}{Y}, or {Y}{Y}.": paying one hybrid {X/Y} mana and
// tapping the land adds two mana, each independently either color of the fixed
// pair. The three printed combinations {X}{X}, {X}{Y}, and {Y}{Y} are exactly
// the unordered two-mana multisets over {X, Y}, so two independent color choices
// over the pair reproduce the printed output faithfully. The two choices publish
// under distinct keys so the instruction sequence is valid.
func TwoColorFilterManaAbility(first, second mana.Color) ManaAbility {
	validateFilterManaPair(first, second)
	firstSymbol := manaSymbol(first)
	secondSymbol := manaSymbol(second)
	return ManaAbility{
		Text: fmt.Sprintf(
			"{%s/%s}, {T}: Add {%s}{%s}, {%s}{%s}, or {%s}{%s}.",
			firstSymbol, secondSymbol,
			firstSymbol, firstSymbol,
			firstSymbol, secondSymbol,
			secondSymbol, secondSymbol,
		),
		ManaCost:        opt.Val(cost.Mana{cost.HybridMana(first, second)}),
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{first, second},
					},
					PublishChoice: tapManaFilterFirstKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaFilterFirstKey,
				},
			},
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{first, second},
					},
					PublishChoice: tapManaFilterSecondKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaFilterSecondKey,
				},
			},
		}}.Ability(),
	}
}

// validateFilterManaPair panics unless first and second are two distinct basic
// colors (W, U, B, R, or G), the only inputs the filter-land output body admits.
func validateFilterManaPair(first, second mana.Color) {
	for _, manaColor := range []mana.Color{first, second} {
		switch manaColor {
		case mana.W, mana.U, mana.B, mana.R, mana.G:
		default:
			panic(fmt.Sprintf("game: invalid filter mana color %q", manaColor))
		}
	}
	if first == second {
		panic(fmt.Sprintf("game: filter mana pair requires two distinct colors, got %q twice", first))
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

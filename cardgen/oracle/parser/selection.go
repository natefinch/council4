package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TriggerCardType identifies a literal card type in trigger syntax.
type TriggerCardType string

// Literal card types recognized in trigger syntax.
const (
	TriggerCardTypeUnknown      TriggerCardType = ""
	TriggerCardTypeArtifact     TriggerCardType = "TriggerCardTypeArtifact"
	TriggerCardTypeBattle       TriggerCardType = "TriggerCardTypeBattle"
	TriggerCardTypeCreature     TriggerCardType = "TriggerCardTypeCreature"
	TriggerCardTypeEnchantment  TriggerCardType = "TriggerCardTypeEnchantment"
	TriggerCardTypeInstant      TriggerCardType = "TriggerCardTypeInstant"
	TriggerCardTypeLand         TriggerCardType = "TriggerCardTypeLand"
	TriggerCardTypePlaneswalker TriggerCardType = "TriggerCardTypePlaneswalker"
	TriggerCardTypeSorcery      TriggerCardType = "TriggerCardTypeSorcery"
)

// TriggerColor identifies a literal color in trigger syntax.
type TriggerColor string

// Literal colors recognized in trigger syntax.
const (
	TriggerColorUnknown TriggerColor = ""
	TriggerColorWhite   TriggerColor = "TriggerColorWhite"
	TriggerColorBlue    TriggerColor = "TriggerColorBlue"
	TriggerColorBlack   TriggerColor = "TriggerColorBlack"
	TriggerColorRed     TriggerColor = "TriggerColorRed"
	TriggerColorGreen   TriggerColor = "TriggerColorGreen"
)

// TriggerSubtype is a canonical subtype identity in trigger syntax.
type TriggerSubtype = types.Sub

// TriggerSupertype identifies a literal supertype in trigger syntax.
type TriggerSupertype string

// Literal supertypes recognized in trigger syntax.
const (
	TriggerSupertypeUnknown   TriggerSupertype = ""
	TriggerSupertypeLegendary TriggerSupertype = "TriggerSupertypeLegendary"
	TriggerSupertypeSnow      TriggerSupertype = "TriggerSupertypeSnow"
)

// TriggerController identifies a literal controller relation in trigger syntax.
type TriggerController string

// Literal controller relations recognized in trigger syntax.
const (
	ControllerAny      TriggerController = ""
	ControllerYou      TriggerController = "ControllerYou"
	ControllerOpponent TriggerController = "ControllerOpponent"
)

// TriggerSelectionTappedState identifies a selected permanent's tapped state.
type TriggerSelectionTappedState string

// Tapped-state predicates recognized in trigger selections.
const (
	TriggerSelectionTappedAny TriggerSelectionTappedState = ""
	TriggerSelectionTapped    TriggerSelectionTappedState = "TriggerSelectionTapped"
	TriggerSelectionUntapped  TriggerSelectionTappedState = "TriggerSelectionUntapped"
)

// TriggerSelectionCombatState identifies a selected permanent's combat state.
type TriggerSelectionCombatState string

// Combat-state predicates recognized in trigger selections.
const (
	TriggerSelectionCombatAny           TriggerSelectionCombatState = ""
	TriggerSelectionAttacking           TriggerSelectionCombatState = "TriggerSelectionAttacking"
	TriggerSelectionBlocking            TriggerSelectionCombatState = "TriggerSelectionBlocking"
	TriggerSelectionAttackingOrBlocking TriggerSelectionCombatState = "TriggerSelectionAttackingOrBlocking"
)

// TriggerSelectionComparison identifies an integer comparison.
type TriggerSelectionComparison string

// Integer comparisons recognized in trigger selections.
const (
	TriggerSelectionComparisonUnknown TriggerSelectionComparison = ""
	TriggerSelectionComparisonEqual   TriggerSelectionComparison = "TriggerSelectionComparisonEqual"
	TriggerSelectionComparisonAtMost  TriggerSelectionComparison = "TriggerSelectionComparisonAtMost"
	TriggerSelectionComparisonAtLeast TriggerSelectionComparison = "TriggerSelectionComparisonAtLeast"
)

// TriggerSelectionNumber is a source-spanned integer predicate.
type TriggerSelectionNumber struct {
	Comparison TriggerSelectionComparison `json:",omitempty"`
	Value      int                        `json:",omitempty"`
	Span       shared.Span                `json:"-"`
}

// TriggerSelection is typed syntax for a permanent noun phrase in a trigger.
type TriggerSelection struct {
	RequiredTypes    []TriggerCardType           `json:",omitempty"`
	RequiredTypesAny []TriggerCardType           `json:",omitempty"`
	ExcludedTypes    []TriggerCardType           `json:",omitempty"`
	Supertypes       []TriggerSupertype          `json:",omitempty"`
	SubtypesAny      []TriggerSubtype            `json:",omitempty"`
	ColorsAny        []TriggerColor              `json:",omitempty"`
	ExcludedColors   []TriggerColor              `json:",omitempty"`
	Colorless        bool                        `json:",omitempty"`
	Multicolored     bool                        `json:",omitempty"`
	NonToken         bool                        `json:",omitempty"`
	TokenOnly        bool                        `json:",omitempty"`
	Controller       TriggerController           `json:",omitempty"`
	Tapped           TriggerSelectionTappedState `json:",omitempty"`
	CombatState      TriggerSelectionCombatState `json:",omitempty"`
	Keyword          KeywordKind                 `json:",omitempty"`
	ExcludedKeyword  KeywordKind                 `json:",omitempty"`
	ManaValue        TriggerSelectionNumber      `json:",omitzero"`
	Power            TriggerSelectionNumber      `json:",omitzero"`
	Toughness        TriggerSelectionNumber      `json:",omitzero"`

	// MatchAnyCounter records a kind-agnostic "with a counter on it" subject
	// qualifier. It compiles to the matching CompiledSelector counter dimension.
	// A kind-specific "with a <kind> counter on it" wording instead rides
	// MatchCounter and RequiredCounter below.
	MatchAnyCounter bool `json:",omitempty"`

	// MatchCounter and RequiredCounter record a kind-specific "with a <kind>
	// counter on it" subject qualifier ("a creature you control with a +1/+1
	// counter on it dies"). The matched permanent must carry a counter of
	// RequiredCounter's kind. A dying subject is matched against its last-known
	// counters (CR 603.10), the same snapshot undying and persist read, so the
	// runtime exposes the per-kind counter information this qualifier needs.
	MatchCounter    bool         `json:",omitempty"`
	RequiredCounter counter.Kind `json:",omitempty"`

	// Modified records a "modified" subject qualifier ("a modified creature you
	// control"): the matched permanent must carry a counter or have an Aura or
	// Equipment attached (CR 701.50). It compiles to Selection.MatchModified.
	Modified bool `json:",omitempty"`
	// Enchanted records an "enchanted" damage-source qualifier. Damage-trigger
	// lowering snapshots whether the source had an Aura attached at event time.
	Enchanted bool `json:",omitempty"`

	// Commander records a "commander" subject ("your commander deals combat
	// damage to a player", Archivist of Gondor). The matched permanent must be a
	// commander; it compiles to Selection.MatchCommander.
	Commander bool `json:",omitempty"`

	// Goaded records a "goaded" subject qualifier ("Whenever a goaded creature
	// attacks", Vengeful Ancestor): the matched permanent must be goaded right
	// now (CR 701.38). It compiles to Selection.MatchGoaded.
	Goaded bool `json:",omitempty"`

	// PowerAboveBase records a "with power greater than its base power" subject
	// qualifier ("one or more creatures you control ... with power greater than
	// its base power", Kutzil, Malamet Exemplar). The matched permanent's current
	// power must exceed its base power (CR 208.3). It compiles to
	// Selection.PowerAboveBase.
	PowerAboveBase bool `json:",omitempty"`

	// SubtypeFromEntryChoice records a trailing "of the chosen type" qualifier
	// ("a creature you control of the chosen type"), tying the matched permanent
	// to the creature subtype the trigger's source permanent chose as it entered.
	// It compiles to the runtime Selection.SubtypeFromSourceEntryChoice predicate.
	SubtypeFromEntryChoice bool `json:",omitempty"`

	// AnyOf records a disjunction of alternative selections, matching when the
	// subject satisfies at least one alternative ("creature or Vehicle", a mixed
	// card-type-or-subtype union the flat RequiredTypesAny/SubtypesAny fields
	// cannot express because those AND across categories). The other fields on
	// this selection remain common conjunctive requirements shared by every
	// alternative (the controller, tapped state, and excluded-self relation). It
	// compiles to Selection.AnyOf.
	AnyOf []TriggerSelection `json:",omitempty"`
}

func parseTriggerSelection(tokens []shared.Token) (TriggerSelection, bool) {
	selection := TriggerSelection{}
	if rest, kind, ok := stripTriggerCounterKindQualifier(tokens); ok {
		selection.MatchCounter = true
		selection.RequiredCounter = kind
		tokens = rest
	}
	words, ok := triggerSelectionWords(tokens)
	if !ok || len(words) == 0 {
		return TriggerSelection{}, false
	}
	words, ok = prepareTriggerSelection(words, tokens, &selection)
	if !ok {
		return TriggerSelection{}, false
	}
	if len(words) == 0 {
		return selection, selection.TokenOnly
	}
	words = consumeTriggerSelectionModifiers(words, &selection)
	return parseTriggerSelectionNoun(words, selection)
}

// stripTriggerCounterKindQualifier removes a trailing "with [a/an] <kind>
// counter(s) on it/them" qualifier from a trigger subject's tokens, returning
// the remaining base tokens and the named counter kind. It matches only the
// kind-specific form ("with a +1/+1 counter on it"); the kind-agnostic "with a
// counter on it" wording carries no counter name and is left for the word-based
// qualifier path (MatchAnyCounter). Any other shape fails closed so unrelated
// "with" qualifiers (keyword, power, toughness, mana value) keep their existing
// handling.
func stripTriggerCounterKindQualifier(tokens []shared.Token) ([]shared.Token, counter.Kind, bool) {
	n := len(tokens)
	if n < 5 {
		return tokens, 0, false
	}
	if !equalWord(tokens[n-2], "on") ||
		(!equalWord(tokens[n-1], "it") && !equalWord(tokens[n-1], "them")) {
		return tokens, 0, false
	}
	counterIndex := n - 3
	if !equalWord(tokens[counterIndex], "counter") && !equalWord(tokens[counterIndex], "counters") {
		return tokens, 0, false
	}
	withIndex := -1
	for i := counterIndex - 1; i >= 0; i-- {
		if equalWord(tokens[i], "with") {
			withIndex = i
			break
		}
	}
	if withIndex < 0 {
		return tokens, 0, false
	}
	kind, span, ok := counterNameBefore(tokens, counterIndex)
	if !ok {
		return tokens, 0, false
	}
	nameStart := withIndex + 1
	if nameStart < counterIndex && (equalWord(tokens[nameStart], "a") || equalWord(tokens[nameStart], "an")) {
		nameStart++
	}
	// The named counter-kind tokens must fill the gap exactly between the
	// optional article and the "counter" noun, so a stray modifier ("with a big
	// +1/+1 counter on it") fails closed rather than silently dropping it.
	if nameStart >= counterIndex || span.Start.Offset != tokens[nameStart].Span.Start.Offset {
		return tokens, 0, false
	}
	return tokens[:withIndex], kind, true
}

func triggerSelectionWords(tokens []shared.Token) ([]string, bool) {
	words := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if i+2 < len(tokens) &&
			equalWord(token, "and") &&
			tokens[i+1].Kind == shared.Slash &&
			equalWord(tokens[i+2], "or") {
			words = append(words, "and/or")
			i += 2
			continue
		}
		if i+2 < len(tokens) &&
			token.Kind == shared.Integer &&
			tokens[i+1].Kind == shared.Slash &&
			tokens[i+2].Kind == shared.Integer {
			words = append(words, token.Text+"/"+tokens[i+2].Text)
			i += 2
			continue
		}
		if token.Kind == shared.Integer {
			words = append(words, token.Text)
			continue
		}
		if token.Kind != shared.Word {
			return nil, false
		}
		words = append(words, strings.ToLower(token.Text))
	}
	return words, true
}

func prepareTriggerSelection(
	words []string,
	tokens []shared.Token,
	selection *TriggerSelection,
) ([]string, bool) {
	var ok bool
	words, selection.Controller, ok = cutEmbeddedTriggerController(words)
	if !ok {
		return nil, false
	}
	if len(words) > 0 && (words[len(words)-1] == "token" || words[len(words)-1] == "tokens") {
		selection.TokenOnly = true
		words = words[:len(words)-1]
	}
	if len(words) > 0 && (words[len(words)-1] == "card" || words[len(words)-1] == "cards") {
		words = words[:len(words)-1]
	}
	if len(words) == 0 {
		return words, selection.TokenOnly
	}
	switch words[len(words)-1] {
	case "tapped":
		selection.Tapped = TriggerSelectionTapped
		words = words[:len(words)-1]
	case "untapped":
		selection.Tapped = TriggerSelectionUntapped
		words = words[:len(words)-1]
	default:
	}
	if qualifier := slices.Index(words, "with"); qualifier >= 0 {
		if !parseTriggerSelectionQualifier(words[qualifier+1:], false, selection) {
			return nil, false
		}
		words = cutTriggerDistributiveEach(words[:qualifier])
	} else if qualifier := slices.Index(words, "without"); qualifier >= 0 {
		if !parseTriggerSelectionQualifier(words[qualifier+1:], true, selection) {
			return nil, false
		}
		words = cutTriggerDistributiveEach(words[:qualifier])
	}
	if len(words) > 1 {
		if powerToughness := parseTriggerPowerToughness(words[0], tokens); powerToughness.ok {
			selection.Power = powerToughness.power
			selection.Toughness = powerToughness.toughness
			words = words[1:]
		}
	}
	return words, true
}

func consumeTriggerSelectionModifiers(words []string, selection *TriggerSelection) []string {
	for len(words) > 0 {
		if len(words) >= 3 && words[0] == "attacking" && words[1] == "or" && words[2] == "blocking" {
			selection.CombatState = TriggerSelectionAttackingOrBlocking
			words = words[3:]
			continue
		}
		switch words[0] {
		case "nontoken":
			selection.NonToken = true
		case "token":
			selection.TokenOnly = true
		case "legendary":
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeLegendary)
		case "snow":
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeSnow)
		case "white", "blue", "black", "red", "green":
			selection.ColorsAny = append(selection.ColorsAny, triggerColor(words[0]))
		case "nonwhite", "nonblue", "nonblack", "nonred", "nongreen":
			selection.ExcludedColors = append(selection.ExcludedColors, triggerColor(strings.TrimPrefix(words[0], "non")))
		case "colorless":
			selection.Colorless = true
		case "multicolored":
			selection.Multicolored = true
		case "attacking":
			selection.CombatState = TriggerSelectionAttacking
		case "blocking":
			selection.CombatState = TriggerSelectionBlocking
		case "modified":
			selection.Modified = true
		case "enchanted":
			selection.Enchanted = true
		case "goaded":
			selection.Goaded = true
		default:
			return words
		}
		words = words[1:]
	}
	return words
}

func parseTriggerSelectionNoun(words []string, selection TriggerSelection) (TriggerSelection, bool) {
	if len(words) == 0 {
		return TriggerSelection{}, false
	}
	if len(words) == 3 && (words[1] == "and/or" || words[1] == "or") {
		return parseTriggerSelectionAlternativeNouns(words, selection)
	}
	var subtypeWords []string
	for _, word := range words {
		if rest, prefixed := strings.CutPrefix(word, "non"); prefixed {
			if cardType, ok := triggerCardType(rest); ok {
				selection.ExcludedTypes = append(selection.ExcludedTypes, cardType)
				continue
			}
		}
		if cardType, ok := triggerCardType(word); ok {
			if cardType != TriggerCardTypeUnknown && !slices.Contains(selection.RequiredTypes, cardType) {
				selection.RequiredTypes = append(selection.RequiredTypes, cardType)
			}
			continue
		}
		subtypeWords = append(subtypeWords, word)
	}
	if len(subtypeWords) > 0 {
		subtype := strings.Join(subtypeWords, " ")
		if subtype == "outlaw" || subtype == "outlaws" {
			selection.SubtypesAny = []TriggerSubtype{
				types.Assassin,
				types.Mercenary,
				types.Pirate,
				types.Rogue,
				types.Warlock,
			}
			return selection, true
		}
		sub, ok := recognizeSubtypePhrase(subtype)
		if !ok || !looksLikeTriggerSubtype(subtype) {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{sub}
	}
	return selection, true
}

func parseTriggerSelectionAlternativeNouns(words []string, selection TriggerSelection) (TriggerSelection, bool) {
	left, leftOK := triggerCardType(words[0])
	right, rightOK := triggerCardType(words[2])
	if leftOK && rightOK {
		selection.RequiredTypesAny = []TriggerCardType{left, right}
		return selection, true
	}
	leftSub, leftSubOK := recognizeSubtypePhrase(words[0])
	rightSub, rightSubOK := recognizeSubtypePhrase(words[2])
	if leftOK && rightSubOK {
		selection.AnyOf = []TriggerSelection{
			{RequiredTypesAny: []TriggerCardType{left}},
			{SubtypesAny: []TriggerSubtype{rightSub}},
		}
		return selection, true
	}
	if rightOK && leftSubOK {
		selection.AnyOf = []TriggerSelection{
			{SubtypesAny: []TriggerSubtype{leftSub}},
			{RequiredTypesAny: []TriggerCardType{right}},
		}
		return selection, true
	}
	if leftOK || rightOK || !leftSubOK || !rightSubOK {
		return TriggerSelection{}, false
	}
	selection.SubtypesAny = []TriggerSubtype{leftSub, rightSub}
	return selection, true
}

func cutEmbeddedTriggerController(words []string) ([]string, TriggerController, bool) {
	result := append([]string(nil), words...)
	controller := ControllerAny
	for _, relation := range []struct {
		words      []string
		controller TriggerController
	}{
		{[]string{"you", "control"}, ControllerYou},
		{[]string{"an", "opponent", "controls"}, ControllerOpponent},
		{[]string{"your", "opponents", "control"}, ControllerOpponent},
		{[]string{"you", "don't", "control"}, ControllerOpponent},
	} {
		for start := 0; start+len(relation.words) <= len(result); start++ {
			if !slices.Equal(result[start:start+len(relation.words)], relation.words) {
				continue
			}
			if controller != ControllerAny && controller != relation.controller {
				return nil, ControllerAny, false
			}
			controller = relation.controller
			result = append(result[:start], result[start+len(relation.words):]...)
			break
		}
	}
	return result, controller, len(result) > 0
}

func parseTriggerSelectionQualifier(words []string, excluded bool, selection *TriggerSelection) bool {
	if len(words) == 1 {
		keyword := triggerSelectionKeyword(words[0])
		if keyword == KeywordUnknown {
			return false
		}
		if excluded {
			selection.ExcludedKeyword = keyword
		} else {
			selection.Keyword = keyword
		}
		return true
	}
	if excluded || len(words) < 2 {
		return false
	}
	if triggerCounterAnyQualifier(words) {
		selection.MatchAnyCounter = true
		return true
	}
	if triggerPowerAboveBaseQualifier(words) {
		selection.PowerAboveBase = true
		return true
	}
	var destination *TriggerSelectionNumber
	switch {
	case words[0] == "power":
		destination = &selection.Power
	case words[0] == "toughness":
		destination = &selection.Toughness
	case len(words) >= 3 && words[0] == "base" && words[1] == "power":
		// A "base power N" subject filter has no runtime base-power
		// characteristic distinct from current power; it lowers to the same
		// power comparison the Power slot carries, the representable
		// approximation that matches at the moment the subject enters.
		destination = &selection.Power
		words = words[1:]
	case len(words) >= 3 && words[0] == "mana" && words[1] == "value":
		destination = &selection.ManaValue
		words = words[1:]
	default:
		return false
	}
	number, ok := parseTriggerSelectionNumber(words[1:])
	if !ok {
		return false
	}
	*destination = number
	return true
}

// triggerCounterAnyQualifier recognizes the kind-agnostic "with a counter on
// it/them" subject qualifier. It returns false for any other shape, including a
// kind-specific "with a <kind> counter on it" wording, so unrelated qualifiers
// keep their existing handling and unrepresentable kind-specific forms fail
// closed.
func triggerCounterAnyQualifier(words []string) bool {
	if len(words) != 4 || (words[0] != "a" && words[0] != "an") {
		return false
	}
	return (words[1] == "counter" || words[1] == "counters") &&
		words[2] == "on" &&
		(words[3] == "it" || words[3] == "them")
}

// triggerPowerAboveBaseQualifier recognizes the "power greater than its base
// power" subject qualifier (Kutzil, Malamet Exemplar): a self-relative power
// comparison requiring the matched permanent's current power to exceed its base
// power. Any other shape fails closed so unrelated qualifiers keep their existing
// handling.
func triggerPowerAboveBaseQualifier(words []string) bool {
	return slices.Equal(words, []string{"power", "greater", "than", "its", "base", "power"})
}

// cutTriggerDistributiveEach drops a distributive "each" that trails the noun
// words directly before a subject qualifier ("one or more creatures you control
// each with power greater than its base power", Kutzil). The word restates the
// per-permanent scope the batched trigger already applies to every matched
// source and carries no additional selection semantics.
func cutTriggerDistributiveEach(words []string) []string {
	if len(words) > 0 && words[len(words)-1] == "each" {
		return words[:len(words)-1]
	}
	return words
}

func parseTriggerSelectionNumber(words []string) (TriggerSelectionNumber, bool) {
	if len(words) == 0 {
		return TriggerSelectionNumber{}, false
	}
	value, ok := parseTriggerSelectionInt(words[0])
	if !ok {
		return TriggerSelectionNumber{}, false
	}
	comparison := TriggerSelectionComparisonEqual
	if len(words) == 3 && words[1] == "or" {
		switch words[2] {
		case "less":
			comparison = TriggerSelectionComparisonAtMost
		case "greater":
			comparison = TriggerSelectionComparisonAtLeast
		default:
			return TriggerSelectionNumber{}, false
		}
	} else if len(words) != 1 {
		return TriggerSelectionNumber{}, false
	}
	return TriggerSelectionNumber{Comparison: comparison, Value: value}, true
}

func parseTriggerSelectionInt(word string) (int, bool) {
	if word == "" {
		return 0, false
	}
	value := 0
	for _, r := range word {
		if r < '0' || r > '9' {
			return 0, false
		}
		value = value*10 + int(r-'0')
	}
	return value, true
}

type triggerPowerToughness struct {
	power     TriggerSelectionNumber
	toughness TriggerSelectionNumber
	ok        bool
}

func parseTriggerPowerToughness(word string, tokens []shared.Token) triggerPowerToughness {
	powerText, toughnessText, ok := strings.Cut(word, "/")
	if !ok {
		return triggerPowerToughness{}
	}
	power, powerOK := parseTriggerSelectionInt(powerText)
	toughness, toughnessOK := parseTriggerSelectionInt(toughnessText)
	if !powerOK || !toughnessOK {
		return triggerPowerToughness{}
	}
	span := shared.Span{}
	if len(tokens) >= 3 {
		span = shared.SpanOf(tokens[:3])
	}
	return triggerPowerToughness{
		power:     TriggerSelectionNumber{Comparison: TriggerSelectionComparisonEqual, Value: power, Span: span},
		toughness: TriggerSelectionNumber{Comparison: TriggerSelectionComparisonEqual, Value: toughness, Span: span},
		ok:        true,
	}
}

func triggerSelectionKeyword(word string) KeywordKind {
	switch word {
	case "deathtouch":
		return KeywordDeathtouch
	case "defender":
		return KeywordDefender
	case "flash":
		return KeywordFlash
	case "flying":
		return KeywordFlying
	case "haste":
		return KeywordHaste
	case "shadow":
		return KeywordShadow
	case "trample":
		return KeywordTrample
	default:
		return KeywordUnknown
	}
}

func cutTriggerController(words []string) ([]string, TriggerController) {
	for _, suffix := range []struct {
		words      []string
		controller TriggerController
	}{
		{[]string{"you", "control"}, ControllerYou},
		{[]string{"an", "opponent", "controls"}, ControllerOpponent},
		{[]string{"your", "opponents", "control"}, ControllerOpponent},
		{[]string{"you", "don't", "control"}, ControllerOpponent},
	} {
		if len(words) >= len(suffix.words) && slices.Equal(words[len(words)-len(suffix.words):], suffix.words) {
			return words[:len(words)-len(suffix.words)], suffix.controller
		}
	}
	return words, ControllerAny
}

func triggerColor(word string) TriggerColor {
	switch value, _ := recognizeColorWord(word); value {
	case ColorWhite:
		return TriggerColorWhite
	case ColorBlue:
		return TriggerColorBlue
	case ColorBlack:
		return TriggerColorBlack
	case ColorRed:
		return TriggerColorRed
	case ColorGreen:
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
	}
}

func triggerCardType(word string) (TriggerCardType, bool) {
	switch singularTriggerWord(word) {
	case "artifact":
		return TriggerCardTypeArtifact, true
	case "battle":
		return TriggerCardTypeBattle, true
	case "creature":
		return TriggerCardTypeCreature, true
	case "enchantment":
		return TriggerCardTypeEnchantment, true
	case "instant":
		return TriggerCardTypeInstant, true
	case "land":
		return TriggerCardTypeLand, true
	case "permanent":
		return TriggerCardTypeUnknown, true
	case "planeswalker":
		return TriggerCardTypePlaneswalker, true
	case "sorcery":
		return TriggerCardTypeSorcery, true
	default:
		return TriggerCardTypeUnknown, false
	}
}

func singularTriggerWord(word string) string {
	return strings.TrimSuffix(word, "s")
}

func looksLikeTriggerSubtype(subject string) bool {
	fields := strings.Fields(subject)
	if len(fields) == 0 || len(fields) > 2 {
		return false
	}
	for _, word := range fields {
		if _, ok := triggerCardType(word); ok ||
			strings.HasPrefix(word, "non") ||
			slices.Contains([]string{"an", "a", "the", "you", "your", "opponent", "or", "and", "but"}, word) {
			return false
		}
	}
	for _, r := range subject {
		if (r >= 'a' && r <= 'z') || r == ' ' || r == '-' || r == '\'' {
			continue
		}
		return false
	}
	return true
}

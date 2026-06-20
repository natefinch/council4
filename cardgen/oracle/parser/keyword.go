package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// KeywordKind identifies a canonical Oracle keyword. The parser owns keyword
// spelling; downstream stages consume this typed identity.
type KeywordKind string

// Oracle keywords currently consumed by semantic compilation or card generation.
const (
	KeywordUnknown          KeywordKind = ""
	KeywordAffinity         KeywordKind = "KeywordAffinity"
	KeywordAnnihilator      KeywordKind = "KeywordAnnihilator"
	KeywordCascade          KeywordKind = "KeywordCascade"
	KeywordChangeling       KeywordKind = "KeywordChangeling"
	KeywordCompanion        KeywordKind = "KeywordCompanion"
	KeywordConvoke          KeywordKind = "KeywordConvoke"
	KeywordCumulativeUpkeep KeywordKind = "KeywordCumulativeUpkeep"
	KeywordCycling          KeywordKind = "KeywordCycling"
	KeywordDeathtouch       KeywordKind = "KeywordDeathtouch"
	KeywordDefender         KeywordKind = "KeywordDefender"
	KeywordDelve            KeywordKind = "KeywordDelve"
	KeywordDevoid           KeywordKind = "KeywordDevoid"
	KeywordDisguise         KeywordKind = "KeywordDisguise"
	KeywordDoubleStrike     KeywordKind = "KeywordDoubleStrike"
	KeywordEmerge           KeywordKind = "KeywordEmerge"
	KeywordEnchant          KeywordKind = "KeywordEnchant"
	KeywordEquip            KeywordKind = "KeywordEquip"
	KeywordEscape           KeywordKind = "KeywordEscape"
	KeywordEternalize       KeywordKind = "KeywordEternalize"
	KeywordExalted          KeywordKind = "KeywordExalted"
	KeywordFirstStrike      KeywordKind = "KeywordFirstStrike"
	KeywordFlash            KeywordKind = "KeywordFlash"
	KeywordFlashback        KeywordKind = "KeywordFlashback"
	KeywordFlying           KeywordKind = "KeywordFlying"
	KeywordForetell         KeywordKind = "KeywordForetell"
	KeywordHaste            KeywordKind = "KeywordHaste"
	KeywordHexproof         KeywordKind = "KeywordHexproof"
	KeywordHorsemanship     KeywordKind = "KeywordHorsemanship"
	KeywordImprovise        KeywordKind = "KeywordImprovise"
	KeywordIndestructible   KeywordKind = "KeywordIndestructible"
	KeywordInfect           KeywordKind = "KeywordInfect"
	KeywordKicker           KeywordKind = "KeywordKicker"
	KeywordLifelink         KeywordKind = "KeywordLifelink"
	KeywordMadness          KeywordKind = "KeywordMadness"
	KeywordMenace           KeywordKind = "KeywordMenace"
	KeywordMorph            KeywordKind = "KeywordMorph"
	KeywordMutate           KeywordKind = "KeywordMutate"
	KeywordNinjutsu         KeywordKind = "KeywordNinjutsu"
	KeywordPersist          KeywordKind = "KeywordPersist"
	KeywordProtection       KeywordKind = "KeywordProtection"
	KeywordProwess          KeywordKind = "KeywordProwess"
	KeywordReadAhead        KeywordKind = "KeywordReadAhead"
	KeywordReach            KeywordKind = "KeywordReach"
	KeywordShadow           KeywordKind = "KeywordShadow"
	KeywordShroud           KeywordKind = "KeywordShroud"
	KeywordSplitSecond      KeywordKind = "KeywordSplitSecond"
	KeywordStorm            KeywordKind = "KeywordStorm"
	KeywordSuspend          KeywordKind = "KeywordSuspend"
	KeywordToxic            KeywordKind = "KeywordToxic"
	KeywordTrample          KeywordKind = "KeywordTrample"
	KeywordUndying          KeywordKind = "KeywordUndying"
	KeywordVigilance        KeywordKind = "KeywordVigilance"
	KeywordWard             KeywordKind = "KeywordWard"
	KeywordWither           KeywordKind = "KeywordWither"
	// KeywordLandcycling and the typed variants below are the landcycling
	// keyword family (CR 702.29). Each is a cycling ability whose
	// discard-from-hand activation searches the library for a land matching a
	// fixed land filter rather than drawing a card.
	KeywordLandcycling      KeywordKind = "KeywordLandcycling"
	KeywordBasicLandcycling KeywordKind = "KeywordBasicLandcycling"
	KeywordPlainscycling    KeywordKind = "KeywordPlainscycling"
	KeywordIslandcycling    KeywordKind = "KeywordIslandcycling"
	KeywordSwampcycling     KeywordKind = "KeywordSwampcycling"
	KeywordMountaincycling  KeywordKind = "KeywordMountaincycling"
	KeywordForestcycling    KeywordKind = "KeywordForestcycling"
)

var keywordNames = map[KeywordKind]string{
	KeywordAffinity:         "Affinity",
	KeywordAnnihilator:      "Annihilator",
	KeywordCascade:          "Cascade",
	KeywordChangeling:       "Changeling",
	KeywordCompanion:        "Companion",
	KeywordConvoke:          "Convoke",
	KeywordCumulativeUpkeep: "Cumulative upkeep",
	KeywordCycling:          "Cycling",
	KeywordDeathtouch:       "Deathtouch",
	KeywordDefender:         "Defender",
	KeywordDelve:            "Delve",
	KeywordDevoid:           "Devoid",
	KeywordDisguise:         "Disguise",
	KeywordDoubleStrike:     "Double strike",
	KeywordEmerge:           "Emerge",
	KeywordEnchant:          "Enchant",
	KeywordEquip:            "Equip",
	KeywordEscape:           "Escape",
	KeywordEternalize:       "Eternalize",
	KeywordExalted:          "Exalted",
	KeywordFirstStrike:      "First strike",
	KeywordFlash:            "Flash",
	KeywordFlashback:        "Flashback",
	KeywordFlying:           "Flying",
	KeywordForetell:         "Foretell",
	KeywordHaste:            "Haste",
	KeywordHexproof:         "Hexproof",
	KeywordHorsemanship:     "Horsemanship",
	KeywordImprovise:        "Improvise",
	KeywordIndestructible:   "Indestructible",
	KeywordInfect:           "Infect",
	KeywordKicker:           "Kicker",
	KeywordLifelink:         "Lifelink",
	KeywordMadness:          "Madness",
	KeywordMenace:           "Menace",
	KeywordMorph:            "Morph",
	KeywordMutate:           "Mutate",
	KeywordNinjutsu:         "Ninjutsu",
	KeywordPersist:          "Persist",
	KeywordProtection:       "Protection",
	KeywordProwess:          "Prowess",
	KeywordReadAhead:        "Read ahead",
	KeywordReach:            "Reach",
	KeywordShadow:           "Shadow",
	KeywordShroud:           "Shroud",
	KeywordSplitSecond:      "Split second",
	KeywordStorm:            "Storm",
	KeywordSuspend:          "Suspend",
	KeywordToxic:            "Toxic",
	KeywordTrample:          "Trample",
	KeywordUndying:          "Undying",
	KeywordVigilance:        "Vigilance",
	KeywordWard:             "Ward",
	KeywordWither:           "Wither",
	KeywordLandcycling:      "Landcycling",
	KeywordBasicLandcycling: "Basic landcycling",
	KeywordPlainscycling:    "Plainscycling",
	KeywordIslandcycling:    "Islandcycling",
	KeywordSwampcycling:     "Swampcycling",
	KeywordMountaincycling:  "Mountaincycling",
	KeywordForestcycling:    "Forestcycling",
}

// String returns the parser-owned canonical keyword name.
func (k KeywordKind) String() string {
	if name, ok := keywordNames[k]; ok {
		return name
	}
	return "Unknown"
}

// OracleWord returns the lowercase Oracle word(s) for a keyword, the form used in
// "creature token with <keyword>" wording (e.g. KeywordFlying -> "flying",
// KeywordFirstStrike -> "first strike"). It fails closed for the unknown keyword.
func (k KeywordKind) OracleWord() (string, bool) {
	if k == KeywordUnknown {
		return "", false
	}
	name, ok := keywordNames[k]
	if !ok {
		return "", false
	}
	return strings.ToLower(name), true
}

type keywordNameGrammar struct {
	Kind  KeywordKind `json:",omitempty"`
	Words []string    `json:",omitempty"`
}

var keywordNameGrammars = []keywordNameGrammar{
	{Kind: KeywordDoubleStrike, Words: []string{"double", "strike"}},
	{Kind: KeywordFirstStrike, Words: []string{"first", "strike"}},
	{Kind: KeywordCumulativeUpkeep, Words: []string{"cumulative", "upkeep"}},
	{Kind: KeywordReadAhead, Words: []string{"read", "ahead"}},
	{Kind: KeywordSplitSecond, Words: []string{"split", "second"}},
	{Kind: KeywordBasicLandcycling, Words: []string{"basic", "landcycling"}},
	{Kind: KeywordAffinity, Words: []string{"affinity"}},
	{Kind: KeywordAnnihilator, Words: []string{"annihilator"}},
	{Kind: KeywordCascade, Words: []string{"cascade"}},
	{Kind: KeywordChangeling, Words: []string{"changeling"}},
	{Kind: KeywordCompanion, Words: []string{"companion"}},
	{Kind: KeywordConvoke, Words: []string{"convoke"}},
	{Kind: KeywordCycling, Words: []string{"cycling"}},
	{Kind: KeywordDeathtouch, Words: []string{"deathtouch"}},
	{Kind: KeywordDefender, Words: []string{"defender"}},
	{Kind: KeywordDelve, Words: []string{"delve"}},
	{Kind: KeywordDevoid, Words: []string{"devoid"}},
	{Kind: KeywordDisguise, Words: []string{"disguise"}},
	{Kind: KeywordEmerge, Words: []string{"emerge"}},
	{Kind: KeywordEnchant, Words: []string{"enchant"}},
	{Kind: KeywordEquip, Words: []string{"equip"}},
	{Kind: KeywordEscape, Words: []string{"escape"}},
	{Kind: KeywordEternalize, Words: []string{"eternalize"}},
	{Kind: KeywordExalted, Words: []string{"exalted"}},
	{Kind: KeywordFlash, Words: []string{"flash"}},
	{Kind: KeywordFlashback, Words: []string{"flashback"}},
	{Kind: KeywordFlying, Words: []string{"flying"}},
	{Kind: KeywordForetell, Words: []string{"foretell"}},
	{Kind: KeywordHaste, Words: []string{"haste"}},
	{Kind: KeywordHexproof, Words: []string{"hexproof"}},
	{Kind: KeywordHorsemanship, Words: []string{"horsemanship"}},
	{Kind: KeywordImprovise, Words: []string{"improvise"}},
	{Kind: KeywordIndestructible, Words: []string{"indestructible"}},
	{Kind: KeywordInfect, Words: []string{"infect"}},
	{Kind: KeywordKicker, Words: []string{"kicker"}},
	{Kind: KeywordLifelink, Words: []string{"lifelink"}},
	{Kind: KeywordMadness, Words: []string{"madness"}},
	{Kind: KeywordMenace, Words: []string{"menace"}},
	{Kind: KeywordMorph, Words: []string{"morph"}},
	{Kind: KeywordMutate, Words: []string{"mutate"}},
	{Kind: KeywordNinjutsu, Words: []string{"ninjutsu"}},
	{Kind: KeywordPersist, Words: []string{"persist"}},
	{Kind: KeywordProtection, Words: []string{"protection"}},
	{Kind: KeywordProwess, Words: []string{"prowess"}},
	{Kind: KeywordReach, Words: []string{"reach"}},
	{Kind: KeywordShadow, Words: []string{"shadow"}},
	{Kind: KeywordShroud, Words: []string{"shroud"}},
	{Kind: KeywordStorm, Words: []string{"storm"}},
	{Kind: KeywordSuspend, Words: []string{"suspend"}},
	{Kind: KeywordToxic, Words: []string{"toxic"}},
	{Kind: KeywordTrample, Words: []string{"trample"}},
	{Kind: KeywordUndying, Words: []string{"undying"}},
	{Kind: KeywordVigilance, Words: []string{"vigilance"}},
	{Kind: KeywordWard, Words: []string{"ward"}},
	{Kind: KeywordWither, Words: []string{"wither"}},
	{Kind: KeywordLandcycling, Words: []string{"landcycling"}},
	{Kind: KeywordPlainscycling, Words: []string{"plainscycling"}},
	{Kind: KeywordIslandcycling, Words: []string{"islandcycling"}},
	{Kind: KeywordSwampcycling, Words: []string{"swampcycling"}},
	{Kind: KeywordMountaincycling, Words: []string{"mountaincycling"}},
	{Kind: KeywordForestcycling, Words: []string{"forestcycling"}},
}

// KeywordParameterKind identifies the grammar used by a keyword parameter.
type KeywordParameterKind string

// Typed keyword parameter shapes.
const (
	KeywordParameterNone          KeywordParameterKind = ""
	KeywordParameterManaCost      KeywordParameterKind = "KeywordParameterManaCost"
	KeywordParameterInteger       KeywordParameterKind = "KeywordParameterInteger"
	KeywordParameterEnchantTarget KeywordParameterKind = "KeywordParameterEnchantTarget"
	KeywordParameterProtection    KeywordParameterKind = "KeywordParameterProtection"
)

// ProtectionParameter is the composable typed predicate following "Protection
// from". Exactly one predicate family is populated.
type ProtectionParameter struct {
	Everything   bool        `json:",omitempty"`
	EachColor    bool        `json:",omitempty"`
	Multicolored bool        `json:",omitempty"`
	Monocolored  bool        `json:",omitempty"`
	FromColors   []Color     `json:",omitempty"`
	FromTypes    []CardType  `json:",omitempty"`
	FromSubtypes []types.Sub `json:",omitempty"`
}

type keywordParameterDetails struct {
	ManaCost      cost.Mana           `json:",omitempty"`
	Integer       int                 `json:",omitempty"`
	EnchantTarget ObjectNoun          `json:",omitempty"`
	Protection    ProtectionParameter `json:",omitzero"`
}

// KeywordParameter is source-spanned typed syntax for one keyword parameter.
// Text is parser-owned canonical text retained for diagnostics and source-stable
// compiler metadata; semantic consumers use Kind and the typed accessors.
type KeywordParameter struct {
	Kind    KeywordParameterKind `json:",omitempty"`
	Span    shared.Span          `json:"-"`
	Text    string               `json:",omitempty"`
	details *keywordParameterDetails
}

// NewManaKeywordParameter constructs a typed mana-cost parameter.
func NewManaKeywordParameter(span shared.Span, manaCost cost.Mana) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterManaCost,
		Span:    span,
		Text:    manaCost.String(),
		details: &keywordParameterDetails{ManaCost: slices.Clone(manaCost)},
	}
}

// NewIntegerKeywordParameter constructs a typed integer parameter.
func NewIntegerKeywordParameter(span shared.Span, value int) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterInteger,
		Span:    span,
		Text:    strconv.Itoa(value),
		details: &keywordParameterDetails{Integer: value},
	}
}

// NewEnchantTargetKeywordParameter constructs a typed Enchant target parameter.
func NewEnchantTargetKeywordParameter(span shared.Span, target ObjectNoun) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterEnchantTarget,
		Span:    span,
		Text:    enchantTargetName(target),
		details: &keywordParameterDetails{EnchantTarget: target},
	}
}

// NewProtectionKeywordParameter constructs a typed Protection predicate.
func NewProtectionKeywordParameter(span shared.Span, text string, protection ProtectionParameter) KeywordParameter {
	return KeywordParameter{
		Kind: KeywordParameterProtection,
		Span: span,
		Text: text,
		details: &keywordParameterDetails{
			Protection: cloneProtectionParameter(protection),
		},
	}
}

// ManaCost returns a copy of the typed mana-cost parameter.
func (p KeywordParameter) ManaCost() cost.Mana {
	if p.details == nil {
		return nil
	}
	return slices.Clone(p.details.ManaCost)
}

// Integer returns the typed integer parameter.
func (p KeywordParameter) Integer() int {
	if p.details == nil {
		return 0
	}
	return p.details.Integer
}

// EnchantTarget returns the typed Enchant target parameter.
func (p KeywordParameter) EnchantTarget() ObjectNoun {
	if p.details == nil {
		return ObjectNounUnknown
	}
	return p.details.EnchantTarget
}

// Protection returns a copy of the typed Protection predicate.
func (p KeywordParameter) Protection() ProtectionParameter {
	if p.details == nil {
		return ProtectionParameter{}
	}
	return cloneProtectionParameter(p.details.Protection)
}

// Keyword is one source-spanned recognized keyword and its typed parameter.
type Keyword struct {
	Kind      KeywordKind      `json:",omitempty"`
	NameSpan  shared.Span      `json:"-"`
	Span      shared.Span      `json:"-"`
	Text      string           `json:",omitempty"`
	Parameter KeywordParameter `json:",omitzero"`
}

// KeywordSelectorForm identifies how a selector introduces its keyword.
type KeywordSelectorForm string

// Keyword-selector forms.
const (
	KeywordSelectorFormUnknown KeywordSelectorForm = ""
	KeywordSelectorFormDirect  KeywordSelectorForm = "KeywordSelectorFormDirect"
	KeywordSelectorFormAbility KeywordSelectorForm = "KeywordSelectorFormAbility"
)

// KeywordSelector is composable "with/without <keyword>" selector syntax.
type KeywordSelector struct {
	Keyword  KeywordKind         `json:",omitempty"`
	Form     KeywordSelectorForm `json:",omitempty"`
	Span     shared.Span         `json:"-"`
	Excluded bool                `json:",omitempty"`
}

func scanKeywords(tokens []shared.Token, atoms Atoms) []Keyword {
	var keywords []Keyword
	for i := 0; i < len(tokens); i++ {
		kind, width, ok := recognizeKeywordNameAt(tokens, i)
		if !ok || kind == KeywordShadow {
			continue
		}
		nameSpan := shared.SpanOf(tokens[i : i+width])
		// A keyword word that falls inside an occurrence of the card's own name
		// (e.g. "Storm" in "Command the Storm") is part of the name, not a
		// granted ability keyword, so it must not be scanned as one.
		if atoms.SelfNameAt(nameSpan) {
			i += width - 1
			continue
		}
		end := i + width
		parameter, parameterEnd := parseKeywordParameter(kind, tokens, end, atoms)
		end = parameterEnd
		keywords = append(keywords, Keyword{
			Kind:      kind,
			NameSpan:  nameSpan,
			Span:      shared.SpanOf(tokens[i:end]),
			Text:      joinTokens(tokens[i:end]),
			Parameter: parameter,
		})
		i = end - 1
	}
	return keywords
}

func recognizeKeywordNameAt(tokens []shared.Token, start int) (KeywordKind, int, bool) {
	for _, grammar := range keywordNameGrammars {
		if atomWordsAt(tokens, start, grammar.Words...) {
			return grammar.Kind, len(grammar.Words), true
		}
	}
	return KeywordUnknown, 0, false
}

func parseKeywordParameter(
	kind KeywordKind,
	tokens []shared.Token,
	start int,
	atoms Atoms,
) (parameter KeywordParameter, end int) {
	switch kind {
	case KeywordProtection:
		return parseProtectionKeywordParameter(tokens, start, atoms)
	case KeywordEnchant:
		if start < len(tokens) {
			if target, ok := recognizeEnchantTarget(tokens[start]); ok {
				return NewEnchantTargetKeywordParameter(tokens[start].Span, target), start + 1
			}
		}
		return KeywordParameter{}, start
	default:
	}
	if manaCost, end, ok := parseKeywordManaCost(tokens, start); ok {
		return NewManaKeywordParameter(shared.SpanOf(tokens[start:end]), manaCost), end
	}
	if start < len(tokens) && tokens[start].Kind == shared.Integer {
		value, err := strconv.Atoi(tokens[start].Text)
		if err == nil {
			return NewIntegerKeywordParameter(tokens[start].Span, value), start + 1
		}
	}
	return KeywordParameter{}, start
}

func recognizeEnchantTarget(token shared.Token) (ObjectNoun, bool) {
	if token.Kind != shared.Word {
		return ObjectNounUnknown, false
	}
	switch strings.ToLower(token.Text) {
	case "artifact":
		return ObjectNounArtifact, true
	case "creature":
		return ObjectNounCreature, true
	case "enchantment":
		return ObjectNounEnchantment, true
	case "land":
		return ObjectNounLand, true
	case "permanent":
		return ObjectNounPermanent, true
	case "planeswalker":
		return ObjectNounPlaneswalker, true
	case "player":
		return ObjectNounPlayer, true
	default:
		return ObjectNounUnknown, false
	}
}

func enchantTargetName(target ObjectNoun) string {
	switch target {
	case ObjectNounArtifact:
		return "artifact"
	case ObjectNounCreature:
		return "creature"
	case ObjectNounEnchantment:
		return "enchantment"
	case ObjectNounLand:
		return "land"
	case ObjectNounPermanent:
		return "permanent"
	case ObjectNounPlaneswalker:
		return "planeswalker"
	case ObjectNounPlayer:
		return "player"
	default:
		return ""
	}
}

func parseKeywordManaCost(tokens []shared.Token, start int) (cost.Mana, int, bool) {
	end := start
	var result cost.Mana
	for end < len(tokens) && tokens[end].Kind == shared.Symbol {
		symbol, ok := parseKeywordManaSymbol(tokens[end].Text)
		if !ok {
			return nil, start, false
		}
		result = append(result, symbol)
		end++
	}
	return result, end, len(result) > 0
}

func parseKeywordManaSymbol(text string) (cost.Symbol, bool) {
	symbol, ok := strings.CutPrefix(text, "{")
	if !ok {
		return cost.Symbol{}, false
	}
	symbol, ok = strings.CutSuffix(symbol, "}")
	if !ok {
		return cost.Symbol{}, false
	}
	switch symbol {
	case "X":
		return cost.X, true
	case "C":
		return cost.C, true
	case "S":
		return cost.S, true
	case "W":
		return cost.W, true
	case "U":
		return cost.U, true
	case "B":
		return cost.B, true
	case "R":
		return cost.R, true
	case "G":
		return cost.G, true
	default:
	}
	if value, err := strconv.Atoi(symbol); err == nil {
		return cost.O(value), true
	}
	if colorName, phyrexian := strings.CutSuffix(symbol, "/P"); phyrexian {
		color, colorOK := keywordManaColor(colorName)
		if colorOK {
			return cost.PhyrexianMana(color), true
		}
		return cost.Symbol{}, false
	}
	first, second, hybrid := strings.Cut(symbol, "/")
	if !hybrid {
		return cost.Symbol{}, false
	}
	if first == "2" {
		color, colorOK := keywordManaColor(second)
		if colorOK {
			return cost.Twobrid(color), true
		}
		return cost.Symbol{}, false
	}
	firstColor, firstOK := keywordManaColor(first)
	secondColor, secondOK := keywordManaColor(second)
	if !firstOK || !secondOK {
		return cost.Symbol{}, false
	}
	return cost.HybridMana(firstColor, secondColor), true
}

func keywordManaColor(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	default:
		return "", false
	}
}

func parseProtectionKeywordParameter(
	tokens []shared.Token,
	start int,
	atoms Atoms,
) (parameter KeywordParameter, end int) {
	if start+1 >= len(tokens) || !equalWord(tokens[start], "from") {
		return KeywordParameter{}, start
	}
	if equalWord(tokens[start+1], "everything") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+2]),
			"everything",
			ProtectionParameter{Everything: true},
		), start + 2
	}
	if qualifier, ok := atoms.ColorQualifierAt(tokens[start+1].Span); ok {
		switch qualifier {
		case ColorQualifierMulticolored:
			return NewProtectionKeywordParameter(
				shared.SpanOf(tokens[start:start+2]),
				"multicolored",
				ProtectionParameter{Multicolored: true},
			), start + 2
		case ColorQualifierMonocolored:
			return NewProtectionKeywordParameter(
				shared.SpanOf(tokens[start:start+2]),
				"monocolored",
				ProtectionParameter{Monocolored: true},
			), start + 2
		default:
		}
	}
	if start+2 < len(tokens) &&
		(equalWord(tokens[start+1], "each") && equalWord(tokens[start+2], "color") ||
			equalWord(tokens[start+1], "all") &&
				(equalWord(tokens[start+2], "color") || equalWord(tokens[start+2], "colors"))) {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+3]),
			"eachcolor",
			ProtectionParameter{EachColor: true},
		), start + 3
	}
	if colors, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (Color, bool) {
		return atoms.ColorAt(token.Span)
	}); ok {
		names := make([]string, len(colors))
		for i, color := range colors {
			names[i] = colorName(color)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			strings.Join(names, ","),
			ProtectionParameter{FromColors: colors},
		), end
	}
	if cardTypes, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (CardType, bool) {
		cardType, found := atoms.CardTypeAt(token.Span)
		return cardType, found && protectionCardType(cardType)
	}); ok {
		names := make([]string, len(cardTypes))
		for i, cardType := range cardTypes {
			names[i] = cardTypeName(cardType)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			"types:"+strings.Join(names, ","),
			ProtectionParameter{FromTypes: cardTypes},
		), end
	}
	if subtypes, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (types.Sub, bool) {
		subtype, found := atoms.SubtypeAt(token.Span)
		return subtype, found && SubtypeMatchesAnyRuntimeCardType(subtype, []types.Card{types.Creature, types.Land})
	}); ok {
		names := make([]string, len(subtypes))
		for i, subtype := range subtypes {
			names[i] = string(subtype)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			"subtypes:"+strings.Join(names, ","),
			ProtectionParameter{FromSubtypes: subtypes},
		), end
	}
	return KeywordParameter{}, start
}

func parseProtectionList[T any](
	tokens []shared.Token,
	start int,
	parse func(shared.Token) (T, bool),
) (values []T, end int, ok bool) {
	first, ok := parse(tokens[start+1])
	if !ok {
		return nil, start, false
	}
	values = []T{first}
	end = start + 2
	for end < len(tokens) {
		next := end
		if tokens[next].Kind == shared.Comma {
			next++
		} else if !equalWord(tokens[next], "and") {
			break
		}
		if next < len(tokens) && equalWord(tokens[next], "and") {
			next++
		}
		if next >= len(tokens) || !equalWord(tokens[next], "from") {
			break
		}
		if next+1 >= len(tokens) {
			return nil, start, false
		}
		value, found := parse(tokens[next+1])
		if !found {
			return nil, start, false
		}
		values = append(values, value)
		end = next + 2
	}
	return values, end, true
}

func protectionCardType(cardType CardType) bool {
	switch cardType {
	case CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypeInstant,
		CardTypeLand, CardTypePlaneswalker, CardTypeSorcery:
		return true
	default:
		return false
	}
}

func colorName(color Color) string {
	switch color {
	case ColorWhite:
		return "white"
	case ColorBlue:
		return "blue"
	case ColorBlack:
		return "black"
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	default:
		return ""
	}
}

func cardTypeName(cardType CardType) string {
	switch cardType {
	case CardTypeArtifact:
		return "artifact"
	case CardTypeCreature:
		return "creature"
	case CardTypeEnchantment:
		return "enchantment"
	case CardTypeInstant:
		return "instant"
	case CardTypeLand:
		return "land"
	case CardTypePlaneswalker:
		return "planeswalker"
	case CardTypeSorcery:
		return "sorcery"
	default:
		return ""
	}
}

func cloneProtectionParameter(protection ProtectionParameter) ProtectionParameter {
	protection.FromColors = slices.Clone(protection.FromColors)
	protection.FromTypes = slices.Clone(protection.FromTypes)
	protection.FromSubtypes = slices.Clone(protection.FromSubtypes)
	return protection
}

func scanKeywordSelectors(tokens []shared.Token) []KeywordSelector {
	var selectors []KeywordSelector
	for i := range tokens {
		excluded := false
		nameStart := 0
		form := KeywordSelectorFormDirect
		switch {
		case equalWord(tokens[i], "with"):
			nameStart = i + 1
			if nameStart < len(tokens) && equalWord(tokens[nameStart], "a") {
				nameStart++
				form = KeywordSelectorFormAbility
			}
		case equalWord(tokens[i], "without"):
			excluded = true
			nameStart = i + 1
		default:
			continue
		}
		kind, width, ok := recognizeKeywordNameAt(tokens, nameStart)
		if !ok {
			continue
		}
		end := nameStart + width
		if nameStart == i+2 {
			if end >= len(tokens) || !equalWord(tokens[end], "ability") {
				continue
			}
			end++
		}
		selectors = append(selectors, KeywordSelector{
			Keyword:  kind,
			Form:     form,
			Span:     shared.SpanOf(tokens[i:end]),
			Excluded: excluded,
		})
	}
	return selectors
}

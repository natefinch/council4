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
	KeywordBloodthirst      KeywordKind = "KeywordBloodthirst"
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
	KeywordEmbalm           KeywordKind = "KeywordEmbalm"
	KeywordExalted          KeywordKind = "KeywordExalted"
	KeywordEvolve           KeywordKind = "KeywordEvolve"
	KeywordFabricate        KeywordKind = "KeywordFabricate"
	KeywordFear             KeywordKind = "KeywordFear"
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
	KeywordIntimidate       KeywordKind = "KeywordIntimidate"
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
	KeywordSkulk            KeywordKind = "KeywordSkulk"
	KeywordSplitSecond      KeywordKind = "KeywordSplitSecond"
	KeywordStorm            KeywordKind = "KeywordStorm"
	KeywordSuspend          KeywordKind = "KeywordSuspend"
	KeywordToxic            KeywordKind = "KeywordToxic"
	KeywordTrample          KeywordKind = "KeywordTrample"
	KeywordUndying          KeywordKind = "KeywordUndying"
	KeywordUnleash          KeywordKind = "KeywordUnleash"
	KeywordVigilance        KeywordKind = "KeywordVigilance"
	KeywordWard             KeywordKind = "KeywordWard"
	KeywordWither           KeywordKind = "KeywordWither"
	KeywordRiot             KeywordKind = "KeywordRiot"
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
	KeywordFlanking         KeywordKind = "KeywordFlanking"
)

var keywordNames = map[KeywordKind]string{
	KeywordAffinity:         "Affinity",
	KeywordAnnihilator:      "Annihilator",
	KeywordBloodthirst:      "Bloodthirst",
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
	KeywordEmbalm:           "Embalm",
	KeywordExalted:          "Exalted",
	KeywordEvolve:           "Evolve",
	KeywordFabricate:        "Fabricate",
	KeywordFear:             "Fear",
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
	KeywordIntimidate:       "Intimidate",
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
	KeywordSkulk:            "Skulk",
	KeywordSplitSecond:      "Split second",
	KeywordStorm:            "Storm",
	KeywordSuspend:          "Suspend",
	KeywordToxic:            "Toxic",
	KeywordTrample:          "Trample",
	KeywordUndying:          "Undying",
	KeywordUnleash:          "Unleash",
	KeywordVigilance:        "Vigilance",
	KeywordWard:             "Ward",
	KeywordWither:           "Wither",
	KeywordRiot:             "Riot",
	KeywordLandcycling:      "Landcycling",
	KeywordBasicLandcycling: "Basic landcycling",
	KeywordPlainscycling:    "Plainscycling",
	KeywordIslandcycling:    "Islandcycling",
	KeywordSwampcycling:     "Swampcycling",
	KeywordMountaincycling:  "Mountaincycling",
	KeywordForestcycling:    "Forestcycling",
	KeywordFlanking:         "Flanking",
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
	{Kind: KeywordBloodthirst, Words: []string{"bloodthirst"}},
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
	{Kind: KeywordEmbalm, Words: []string{"embalm"}},
	{Kind: KeywordExalted, Words: []string{"exalted"}},
	{Kind: KeywordEvolve, Words: []string{"evolve"}},
	{Kind: KeywordFabricate, Words: []string{"fabricate"}},
	{Kind: KeywordFear, Words: []string{"fear"}},
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
	{Kind: KeywordIntimidate, Words: []string{"intimidate"}},
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
	{Kind: KeywordSkulk, Words: []string{"skulk"}},
	{Kind: KeywordStorm, Words: []string{"storm"}},
	{Kind: KeywordSuspend, Words: []string{"suspend"}},
	{Kind: KeywordToxic, Words: []string{"toxic"}},
	{Kind: KeywordTrample, Words: []string{"trample"}},
	{Kind: KeywordUndying, Words: []string{"undying"}},
	{Kind: KeywordUnleash, Words: []string{"unleash"}},
	{Kind: KeywordVigilance, Words: []string{"vigilance"}},
	{Kind: KeywordWard, Words: []string{"ward"}},
	{Kind: KeywordWither, Words: []string{"wither"}},
	{Kind: KeywordRiot, Words: []string{"riot"}},
	{Kind: KeywordLandcycling, Words: []string{"landcycling"}},
	{Kind: KeywordPlainscycling, Words: []string{"plainscycling"}},
	{Kind: KeywordIslandcycling, Words: []string{"islandcycling"}},
	{Kind: KeywordSwampcycling, Words: []string{"swampcycling"}},
	{Kind: KeywordMountaincycling, Words: []string{"mountaincycling"}},
	{Kind: KeywordForestcycling, Words: []string{"forestcycling"}},
	{Kind: KeywordFlanking, Words: []string{"flanking"}},
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
	ChosenColor  bool        `json:",omitempty"`
}

// EnchantPredicate is the typed object restriction following an Enchant keyword.
// A permanent matches when it has any listed card type or any listed subtype
// (the union is disjunctive: "artifact or creature", "creature or Vehicle").
// Player and Opponent select a player object; Permanent selects any permanent.
// At most one of Player/Opponent/Permanent is set, and they are never combined
// with CardTypes or Subtypes. The zero value is the fail-closed unknown
// predicate.
type EnchantPredicate struct {
	Player    bool        `json:",omitempty"`
	Opponent  bool        `json:",omitempty"`
	Permanent bool        `json:",omitempty"`
	CardTypes []CardType  `json:",omitempty"`
	Subtypes  []types.Sub `json:",omitempty"`
}

// Empty reports whether the predicate carries no recognized restriction.
func (p EnchantPredicate) Empty() bool {
	return !p.Player && !p.Opponent && !p.Permanent &&
		len(p.CardTypes) == 0 && len(p.Subtypes) == 0
}

func cloneEnchantPredicate(predicate EnchantPredicate) EnchantPredicate {
	predicate.CardTypes = slices.Clone(predicate.CardTypes)
	predicate.Subtypes = slices.Clone(predicate.Subtypes)
	return predicate
}

type keywordParameterDetails struct {
	ManaCost      cost.Mana           `json:",omitempty"`
	Integer       int                 `json:",omitempty"`
	EnchantTarget EnchantPredicate    `json:",omitzero"`
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
func NewEnchantTargetKeywordParameter(span shared.Span, target EnchantPredicate) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterEnchantTarget,
		Span:    span,
		Text:    enchantTargetName(target),
		details: &keywordParameterDetails{EnchantTarget: cloneEnchantPredicate(target)},
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
func (p KeywordParameter) EnchantTarget() EnchantPredicate {
	if p.details == nil {
		return EnchantPredicate{}
	}
	return cloneEnchantPredicate(p.details.EnchantTarget)
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
	// EquipRestriction is the typed quality restriction on a restricted Equip
	// ability ("Equip legendary creature {3}", "Equip Knight {2}"), or nil for an
	// unrestricted Equip. The mana cost is still carried by Parameter.
	EquipRestriction *KeywordEquipRestriction `json:",omitempty"`
}

// KeywordEquipRestriction is the typed quality restriction on a restricted Equip
// ability: the Equipment may attach only to a creature that has every listed
// supertype and at least one of the listed subtypes (CR 301.5c). It models
// "Equip legendary creature {3}" (supertype Legendary) and "Equip <subtype>
// {N}" forms such as "Equip Knight {2}".
type KeywordEquipRestriction struct {
	Span       shared.Span `json:"-"`
	Supertypes []Supertype `json:",omitempty"`
	Subtypes   []types.Sub `json:",omitempty"`
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

// expandBushidoKeyword rewrites each printed "Bushido N" keyword line into the
// triggered ability it abbreviates: "Whenever this creature blocks or becomes
// blocked, it gets +N/+N until end of turn." (CR 702.46a). Bushido is pure
// shorthand for that combat trigger, so expanding it to canonical wording lets
// the standard trigger pipeline lower it. The rewrite is parser-owned because it
// is a wording substitution; downstream stages see only the expanded ability.
func expandBushidoKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := bushidoLineRank(line)
		if !ok {
			continue
		}
		lines[i] = "Whenever this creature blocks or becomes blocked, it gets +" +
			strconv.Itoa(rank) + "/+" + strconv.Itoa(rank) + " until end of turn."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// bushidoLineRank reports the rank N of a line that is exactly the printed
// "Bushido N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, or pair it with other
// rules text, are left untouched.
func bushidoLineRank(line string) (int, bool) {
	const prefix = "Bushido "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandAnnihilatorKeyword rewrites each printed "Annihilator N" keyword line
// into the triggered ability it abbreviates: "Whenever this creature attacks,
// defending player sacrifices N permanents of their choice." (CR 702.85a, the
// Eldrazi keyword). Annihilator is pure shorthand for that combat trigger, so
// expanding it to canonical wording lets the standard trigger pipeline lower it.
// The rewrite is parser-owned because it is a wording substitution; downstream
// stages see only the expanded ability.
func expandAnnihilatorKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := annihilatorLineRank(line)
		if !ok {
			continue
		}
		lines[i] = annihilatorCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// annihilatorCanonicalText is the triggered ability that the printed
// "Annihilator N" keyword abbreviates, with N spelled as its Oracle wording.
func annihilatorCanonicalText(rank int) string {
	if rank == 1 {
		return "Whenever this creature attacks, defending player sacrifices a permanent of their choice."
	}
	word, ok := cardinalWord(rank)
	if !ok {
		word = strconv.Itoa(rank)
	}
	return "Whenever this creature attacks, defending player sacrifices " + word + " permanents of their choice."
}

// annihilatorLineRank reports the rank N of a line that is exactly the printed
// "Annihilator N" keyword, optionally followed only by its parenthesized
// reminder text. Lines that merely contain the word elsewhere, or pair it with
// other rules text, are left untouched.
func annihilatorLineRank(line string) (int, bool) {
	const prefix = "Annihilator "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// extortCanonicalText is the triggered ability that the printed "Extort" keyword
// abbreviates (CR 702.99a).
const extortCanonicalText = "Whenever you cast a spell, you may pay {W/B}. " +
	"If you do, each opponent loses 1 life and you gain that much life."

// expandExtortKeyword rewrites each printed "Extort" keyword line into the
// triggered ability it abbreviates. Like Bushido, Extort is pure shorthand for a
// fixed triggered ability, so expanding it to canonical wording lets the standard
// trigger pipeline lower it. Multiple printed instances each expand to their own
// trigger, matching the rule that each Extort instance triggers separately. The
// rewrite is parser-owned because it is a wording substitution; downstream stages
// see only the expanded ability.
func expandExtortKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isExtortKeywordLine(line) {
			continue
		}
		lines[i] = extortCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isExtortKeywordLine reports whether a line is exactly the printed "Extort"
// keyword, optionally followed only by its parenthesized reminder text. Lines
// that merely contain the word elsewhere, or pair it with other rules text, are
// left untouched.
func isExtortKeywordLine(line string) bool {
	const keyword = "Extort"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

// modularLineRank reports the rank N of a line that is exactly the printed
// "Modular N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, that pair it with other
// rules text, or that use a variable form ("Modular—Sunburst") are left
// untouched.
func modularLineRank(line string) (int, bool) {
	const prefix = "Modular "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// modularCounterPhrase spells the enters-with-counters quantity for Modular rank
// N as Oracle text ("a +1/+1 counter", "two +1/+1 counters"). It fails closed
// for ranks outside the small-cardinal vocabulary the enters-with-counters
// static can spell.
func modularCounterPhrase(rank int) (string, bool) {
	if rank == 1 {
		return "a +1/+1 counter", true
	}
	word, ok := cardinalNumberWord(rank)
	if !ok {
		return "", false
	}
	return word + " +1/+1 counters", true
}

// cardinalNumberWord spells a small positive integer as its Oracle cardinal word
// ("two" … "ten"), the inverse of CardinalWordValue for the values a keyword
// expansion needs. It fails closed outside that range.
func cardinalNumberWord(n int) (string, bool) {
	switch n {
	case 2:
		return "two", true
	case 3:
		return "three", true
	case 4:
		return "four", true
	case 5:
		return "five", true
	case 6:
		return "six", true
	case 7:
		return "seven", true
	case 8:
		return "eight", true
	case 9:
		return "nine", true
	case 10:
		return "ten", true
	default:
		return "", false
	}
}

// expandModularKeyword rewrites each printed "Modular N" keyword line into the
// two abilities it abbreviates (CR 702.43c): a static placing N +1/+1 counters as
// the creature enters, and a dies-trigger that moves those counters onto a target
// artifact creature. Like Bushido and Extort, Modular is pure shorthand for fixed
// abilities, so expanding it to canonical wording lets the standard
// enters-with-counters and trigger pipelines lower it. The rewrite is
// parser-owned because it is a wording substitution; downstream stages see only
// the expanded abilities.
func expandModularKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := modularLineRank(line)
		if !ok {
			continue
		}
		counters, ok := modularCounterPhrase(rank)
		if !ok {
			continue
		}
		lines[i] = "This creature enters with " + counters + " on it.\n" +
			"When this creature dies, you may move all counters from this creature " +
			"onto target artifact creature."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// battleCryCanonicalText is the triggered ability that the printed "Battle cry"
// keyword abbreviates (CR 702.91a).
const battleCryCanonicalText = "Whenever this creature attacks, " +
	"each other attacking creature gets +1/+0 until end of turn."

// expandBattleCryKeyword rewrites each printed "Battle cry" keyword line into the
// triggered ability it abbreviates. Like Extort, Battle cry is pure shorthand for
// a fixed triggered ability, so expanding it to canonical wording lets the
// standard trigger pipeline lower it. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see only the expanded ability.
func expandBattleCryKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isBattleCryKeywordLine(line) {
			continue
		}
		lines[i] = battleCryCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isBattleCryKeywordLine reports whether a line is exactly the printed "Battle
// cry" keyword, optionally followed only by its parenthesized reminder text.
// Lines that merely contain the words elsewhere, or pair the keyword with other
// rules text (such as a sticker-cost prefix), are left untouched.
func isBattleCryKeywordLine(line string) bool {
	const keyword = "Battle cry"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

func scanKeywords(tokens []shared.Token, atoms Atoms) []Keyword {
	var keywords []Keyword
	for i := 0; i < len(tokens); i++ {
		kind, width, ok := recognizeKeywordNameAt(tokens, i)
		if !ok {
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
		// "flash" in the cast-permission idiom "as though they had flash" (or
		// "... it had flash") names the timing reference, not a granted Flash
		// keyword, so it is parsed by the cast-as-though-flash static instead.
		if kind == KeywordFlash && i > 0 && equalWord(tokens[i-1], "had") {
			continue
		}
		end := i + width
		var equipRestriction *KeywordEquipRestriction
		if kind == KeywordEquip {
			if restriction, manaStart, ok := parseEquipRestriction(tokens, end, atoms); ok {
				equipRestriction = restriction
				end = manaStart
			}
		}
		parameter, parameterEnd := parseKeywordParameter(kind, tokens, end, atoms)
		end = parameterEnd
		keywords = append(keywords, Keyword{
			Kind:             kind,
			NameSpan:         nameSpan,
			Span:             shared.SpanOf(tokens[i:end]),
			Text:             joinTokens(tokens[i:end]),
			Parameter:        parameter,
			EquipRestriction: equipRestriction,
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
		if predicate, end, ok := parseEnchantTargetPredicate(tokens, start, atoms); ok {
			return NewEnchantTargetKeywordParameter(shared.SpanOf(tokens[start:end]), predicate), end
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

// parseEquipRestriction recognizes the quality words of a restricted Equip
// ability ("Equip legendary creature {3}", "Equip Knight {2}", "Equip Shaman,
// Warlock, or Wizard {2}") between the Equip keyword and its mana cost. It
// consumes supertype, subtype, and the implied "creature" card-type words (plus
// list separators), returning the typed restriction and the index of the
// following mana symbol. It fails closed (ok=false) when there is no restriction
// quality, when an unrecognized word appears, or when no mana cost follows, so
// an unsupported restricted Equip stays unsupported rather than silently
// dropping the restriction.
func parseEquipRestriction(tokens []shared.Token, start int, atoms Atoms) (*KeywordEquipRestriction, int, bool) {
	restriction := &KeywordEquipRestriction{}
	j := start
	for j < len(tokens) {
		token := tokens[j]
		if token.Kind == shared.Symbol {
			break
		}
		if token.Kind == shared.Comma || equalWord(token, "or") {
			j++
			continue
		}
		if supertype, ok := atoms.SupertypeAt(token.Span); ok {
			restriction.Supertypes = append(restriction.Supertypes, supertype)
			j++
			continue
		}
		if subtype, ok := atoms.SubtypeAt(token.Span); ok {
			restriction.Subtypes = append(restriction.Subtypes, subtype)
			j++
			continue
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && cardType == CardTypeCreature {
			j++
			continue
		}
		return nil, start, false
	}
	if len(restriction.Supertypes) == 0 && len(restriction.Subtypes) == 0 {
		return nil, start, false
	}
	if j >= len(tokens) || tokens[j].Kind != shared.Symbol {
		return nil, start, false
	}
	restriction.Span = shared.SpanOf(tokens[start:j])
	return restriction, j, true
}

// parseEnchantTargetPredicate recognizes the object restriction following an
// Enchant keyword: a single player word ("player", "opponent"), the
// any-permanent word ("permanent"), or a disjunctive list of permanent card
// types and subtypes ("creature", "artifact or creature", "creature, artifact,
// or land", "Forest", "creature or Vehicle"). It consumes only the recognized
// predicate tokens and returns the index after the last one, so any trailing
// qualifier the executable backend does not support (a controller, color,
// power, or zone restriction) is left uncovered and the Enchant ability fails
// closed downstream. It returns ok=false when the first token is not a
// recognized predicate word, so an unrecognized restriction stays unsupported.
func parseEnchantTargetPredicate(tokens []shared.Token, start int, atoms Atoms) (EnchantPredicate, int, bool) {
	if start >= len(tokens) {
		return EnchantPredicate{}, start, false
	}
	switch {
	case equalWord(tokens[start], "player"):
		return EnchantPredicate{Player: true}, start + 1, true
	case equalWord(tokens[start], "opponent"):
		return EnchantPredicate{Opponent: true}, start + 1, true
	case equalWord(tokens[start], "permanent"):
		return EnchantPredicate{Permanent: true}, start + 1, true
	}
	predicate := EnchantPredicate{}
	end := start
	// items requires a separator (comma or "or") between consecutive type and
	// subtype words. Adjacent words without a separator are a single conjunctive
	// type line ("artifact creature" = an artifact creature), which a disjunctive
	// predicate cannot represent, so the second word is left uncovered to fail
	// closed rather than silently widened to a disjunction.
	expectItem := true
	for i := start; i < len(tokens); {
		token := tokens[i]
		// A comma or "or" separates list items; it is meaningful only between
		// recognized words, so end does not advance past a trailing separator.
		if token.Kind == shared.Comma || equalWord(token, "or") {
			expectItem = true
			i++
			continue
		}
		if !expectItem {
			break
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok {
			// The Enchant grammar uses singular nouns ("Enchant creature"); the
			// atom scanner also normalizes plurals, so reject a non-singular form
			// ("Enchant creatures") by leaving it uncovered to fail closed.
			if word, ok := cardTypeWord(cardType); ok && strings.EqualFold(token.Text, word) {
				predicate.CardTypes = append(predicate.CardTypes, cardType)
				expectItem = false
				i++
				end = i
				continue
			}
			break
		}
		if subtype, ok := atoms.SubtypeAt(token.Span); ok {
			if strings.EqualFold(token.Text, string(subtype)) {
				predicate.Subtypes = append(predicate.Subtypes, subtype)
				expectItem = false
				i++
				end = i
				continue
			}
			break
		}
		break
	}
	if predicate.Empty() {
		return EnchantPredicate{}, start, false
	}
	return predicate, end, true
}

// enchantTargetName renders the parser-canonical display text for an Enchant
// target predicate, retained on the keyword parameter for diagnostics.
func enchantTargetName(predicate EnchantPredicate) string {
	switch {
	case predicate.Player:
		return "player"
	case predicate.Opponent:
		return "opponent"
	case predicate.Permanent:
		return "permanent"
	}
	words := make([]string, 0, len(predicate.CardTypes)+len(predicate.Subtypes))
	for _, cardType := range predicate.CardTypes {
		if word, ok := cardTypeWord(cardType); ok {
			words = append(words, word)
		}
	}
	for _, subtype := range predicate.Subtypes {
		words = append(words, strings.ToLower(string(subtype)))
	}
	return strings.Join(words, " or ")
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
	if start+5 < len(tokens) &&
		(equalWord(tokens[start+1], "the") || equalWord(tokens[start+1], "a")) &&
		equalWord(tokens[start+2], "color") && equalWord(tokens[start+3], "of") &&
		equalWord(tokens[start+4], "your") && equalWord(tokens[start+5], "choice") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+6]),
			"color of your choice",
			ProtectionParameter{ChosenColor: true},
		), start + 6
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

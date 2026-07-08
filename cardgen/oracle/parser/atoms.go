package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// This file owns the reusable Oracle semantic atoms shared across parser
// grammar families and consumed as typed values by the compiler and card
// generation. The parser owns every spelling, vocabulary, and normalization
// decision here; downstream stages map these typed atoms onto their own
// semantic and runtime representations and never reinspect Oracle source
// spelling to recover the meanings recognized below.

// Color is a typed Oracle color atom. Its zero value is the fail-closed
// unknown color.
type Color string

// Oracle colors recognized by the parser.
const (
	ColorUnknown Color = ""
	ColorWhite   Color = "ColorWhite"
	ColorBlue    Color = "ColorBlue"
	ColorBlack   Color = "ColorBlack"
	ColorRed     Color = "ColorRed"
	ColorGreen   Color = "ColorGreen"
)

// recognizeColorWord maps a single lowercase-insensitive Oracle color word to a
// typed Color. It fails closed for any other spelling.
func recognizeColorWord(word string) (Color, bool) {
	switch strings.ToLower(word) {
	case "white":
		return ColorWhite, true
	case "blue":
		return ColorBlue, true
	case "black":
		return ColorBlack, true
	case "red":
		return ColorRed, true
	case "green":
		return ColorGreen, true
	default:
		return ColorUnknown, false
	}
}

// colorWord returns the lowercase Oracle word for a typed Color, the inverse of
// recognizeColorWord. It fails closed for the unknown color.
func colorWord(color Color) (string, bool) {
	switch color {
	case ColorWhite:
		return "white", true
	case ColorBlue:
		return "blue", true
	case ColorBlack:
		return "black", true
	case ColorRed:
		return "red", true
	case ColorGreen:
		return "green", true
	default:
		return "", false
	}
}

// ColorQualifier is a typed Oracle color-family qualifier that is not itself a
// single color.
type ColorQualifier string

// Oracle color-family qualifiers recognized by the parser.
const (
	ColorQualifierUnknown      ColorQualifier = ""
	ColorQualifierColorless    ColorQualifier = "ColorQualifierColorless"
	ColorQualifierMulticolored ColorQualifier = "ColorQualifierMulticolored"
	ColorQualifierMonocolored  ColorQualifier = "ColorQualifierMonocolored"
)

func recognizeColorQualifierWord(word string) (ColorQualifier, bool) {
	switch strings.ToLower(word) {
	case "colorless":
		return ColorQualifierColorless, true
	case "multicolored":
		return ColorQualifierMulticolored, true
	case "monocolored":
		return ColorQualifierMonocolored, true
	default:
		return ColorQualifierUnknown, false
	}
}

func recognizeColorOrNonColorWord(word string) (Color, bool) {
	if color, ok := recognizeColorWord(word); ok {
		return color, true
	}
	if rest, ok := strings.CutPrefix(strings.ToLower(word), "non"); ok {
		return recognizeColorWord(rest)
	}
	return ColorUnknown, false
}

// CardType is a typed Oracle card-type atom. Its zero value is the fail-closed
// unknown type.
type CardType string

// Oracle card types recognized by the parser.
const (
	CardTypeUnknown      CardType = ""
	CardTypeArtifact     CardType = "CardTypeArtifact"
	CardTypeBattle       CardType = "CardTypeBattle"
	CardTypeCreature     CardType = "CardTypeCreature"
	CardTypeEnchantment  CardType = "CardTypeEnchantment"
	CardTypeInstant      CardType = "CardTypeInstant"
	CardTypeLand         CardType = "CardTypeLand"
	CardTypePlaneswalker CardType = "CardTypePlaneswalker"
	CardTypeSorcery      CardType = "CardTypeSorcery"
)

// recognizeCardTypeWord maps a singular or plural Oracle card-type word to a
// typed CardType, owning the irregular "sorcery"/"sorceries" plural. It fails
// closed for any other spelling.
func recognizeCardTypeWord(word string) (CardType, bool) {
	word = strings.ToLower(word)
	switch word {
	case "artifact", "artifacts":
		return CardTypeArtifact, true
	case "battle", "battles":
		return CardTypeBattle, true
	case "creature", "creatures":
		return CardTypeCreature, true
	case "enchantment", "enchantments":
		return CardTypeEnchantment, true
	case "instant", "instants":
		return CardTypeInstant, true
	case "land", "lands":
		return CardTypeLand, true
	case "planeswalker", "planeswalkers":
		return CardTypePlaneswalker, true
	case "sorcery", "sorceries":
		return CardTypeSorcery, true
	default:
		return CardTypeUnknown, false
	}
}

// recognizeExcludedCardTypeWord maps a "non-<type>" group prefix word (e.g.
// "Nonland", "Nonartifact") to the excluded CardType, mirroring
// recognizeColorOrNonColorWord at the card-type level. It fails closed for any
// word lacking the "non" prefix or naming an unknown card type.
func recognizeExcludedCardTypeWord(word string) (CardType, bool) {
	if rest, ok := strings.CutPrefix(strings.ToLower(word), "non"); ok {
		return recognizeCardTypeWord(rest)
	}
	return CardTypeUnknown, false
}

// the inverse of recognizeCardTypeWord across every card type (including the
// non-permanent instant and sorcery types). It fails closed for the unknown
// card type.
func cardTypeWord(cardType CardType) (string, bool) {
	switch cardType {
	case CardTypeArtifact:
		return "artifact", true
	case CardTypeBattle:
		return "battle", true
	case CardTypeCreature:
		return "creature", true
	case CardTypeEnchantment:
		return "enchantment", true
	case CardTypeInstant:
		return "instant", true
	case CardTypeLand:
		return "land", true
	case CardTypePlaneswalker:
		return "planeswalker", true
	case CardTypeSorcery:
		return "sorcery", true
	default:
		return "", false
	}
}

// Supertype is a typed Oracle supertype atom.
type Supertype string

// Oracle supertypes recognized by the parser.
const (
	SupertypeUnknown   Supertype = ""
	SupertypeLegendary Supertype = "SupertypeLegendary"
	SupertypeSnow      Supertype = "SupertypeSnow"
	SupertypeBasic     Supertype = "SupertypeBasic"
	SupertypeWorld     Supertype = "SupertypeWorld"
)

// recognizeSupertypeWord maps an Oracle supertype word to a typed Supertype.
func recognizeSupertypeWord(word string) (Supertype, bool) {
	switch strings.ToLower(word) {
	case "legendary":
		return SupertypeLegendary, true
	case "snow":
		return SupertypeSnow, true
	case "basic":
		return SupertypeBasic, true
	case "world":
		return SupertypeWorld, true
	default:
		return SupertypeUnknown, false
	}
}

// supertypeWord returns the lowercase Oracle word for a typed Supertype, the
// inverse of recognizeSupertypeWord. It fails closed for the unknown supertype.
func supertypeWord(supertype Supertype) (string, bool) {
	switch supertype {
	case SupertypeLegendary:
		return "legendary", true
	case SupertypeSnow:
		return "snow", true
	case SupertypeBasic:
		return "basic", true
	case SupertypeWorld:
		return "world", true
	default:
		return "", false
	}
}

// runtimeSupertype maps a typed Oracle supertype onto its runtime supertype.
// It fails closed for the unknown supertype so downstream stages never invent a
// supertype filter the parser did not recognize.
func runtimeSupertype(supertype Supertype) (types.Super, bool) {
	switch supertype {
	case SupertypeLegendary:
		return types.Legendary, true
	case SupertypeSnow:
		return types.Snow, true
	case SupertypeBasic:
		return types.Basic, true
	case SupertypeWorld:
		return types.World, true
	default:
		return "", false
	}
}

// ObjectNoun is a typed Oracle object-noun atom: the reusable nouns that name a
// game object or player. Downstream stages decide which nouns are valid in a
// given grammar from the typed value rather than from spelling.
type ObjectNoun string

// Oracle object nouns recognized by the parser.
const (
	ObjectNounUnknown      ObjectNoun = ""
	ObjectNounAbility      ObjectNoun = "ObjectNounAbility"
	ObjectNounArtifact     ObjectNoun = "ObjectNounArtifact"
	ObjectNounCard         ObjectNoun = "ObjectNounCard"
	ObjectNounCreature     ObjectNoun = "ObjectNounCreature"
	ObjectNounEnchantment  ObjectNoun = "ObjectNounEnchantment"
	ObjectNounEquipment    ObjectNoun = "ObjectNounEquipment"
	ObjectNounLand         ObjectNoun = "ObjectNounLand"
	ObjectNounPermanent    ObjectNoun = "ObjectNounPermanent"
	ObjectNounPlaneswalker ObjectNoun = "ObjectNounPlaneswalker"
	ObjectNounOpponent     ObjectNoun = "ObjectNounOpponent"
	ObjectNounPlayer       ObjectNoun = "ObjectNounPlayer"
	ObjectNounSpell        ObjectNoun = "ObjectNounSpell"
	ObjectNounToken        ObjectNoun = "ObjectNounToken"
	ObjectNounCommander    ObjectNoun = "ObjectNounCommander"
)

// recognizeObjectNoun maps a single word token to a typed ObjectNoun. Non-word
// tokens fail closed.
func recognizeObjectNoun(token shared.Token) (ObjectNoun, bool) {
	if token.Kind != shared.Word {
		return ObjectNounUnknown, false
	}
	switch strings.ToLower(token.Text) {
	case "ability", "abilities":
		return ObjectNounAbility, true
	case "artifact", "artifacts":
		return ObjectNounArtifact, true
	case "card", "cards":
		return ObjectNounCard, true
	case "commander", "commanders":
		return ObjectNounCommander, true
	case "creature", "creatures":
		return ObjectNounCreature, true
	case "enchantment", "enchantments":
		return ObjectNounEnchantment, true
	case "equipment":
		return ObjectNounEquipment, true
	case "land", "lands":
		return ObjectNounLand, true
	case "opponent", "opponents":
		return ObjectNounOpponent, true
	case "permanent", "permanents":
		return ObjectNounPermanent, true
	case "planeswalker", "planeswalkers":
		return ObjectNounPlaneswalker, true
	case "player", "players":
		return ObjectNounPlayer, true
	case "spell", "spells":
		return ObjectNounSpell, true
	case "token", "tokens":
		return ObjectNounToken, true
	default:
		return ObjectNounUnknown, false
	}
}

func graveyardZonePhrase(tokens []shared.Token) bool {
	switch {
	case len(tokens) >= 1 && (equalWord(tokens[0], "graveyard") || equalWord(tokens[0], "graveyards")):
		return true
	case len(tokens) >= 2 &&
		(equalWord(tokens[0], "your") || equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		(equalWord(tokens[1], "graveyard") || equalWord(tokens[1], "graveyards")):
		return true
	case len(tokens) >= 3 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "single") &&
		(equalWord(tokens[2], "graveyard") || equalWord(tokens[2], "graveyards")):
		return true
	case len(tokens) >= 2 &&
		equalWord(tokens[0], "all") &&
		(equalWord(tokens[1], "graveyard") || equalWord(tokens[1], "graveyards")):
		return true
	case len(tokens) >= 2 &&
		strings.EqualFold(tokens[0].Text, "owner's") &&
		(equalWord(tokens[1], "graveyard") || equalWord(tokens[1], "graveyards")):
		return true
	case len(tokens) >= 3 &&
		equalWord(tokens[0], "an") &&
		strings.EqualFold(tokens[1].Text, "opponent's") &&
		(equalWord(tokens[2], "graveyard") || equalWord(tokens[2], "graveyards")):
		return true
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "its") || equalWord(tokens[0], "their")) &&
		(strings.EqualFold(tokens[1].Text, "owner's") || strings.EqualFold(tokens[1].Text, "owners'")) &&
		(equalWord(tokens[2], "graveyard") || equalWord(tokens[2], "graveyards")):
		return true
	default:
		return false
	}
}

func handZonePhrase(tokens []shared.Token) bool {
	switch {
	case len(tokens) >= 1 && (equalWord(tokens[0], "hand") || equalWord(tokens[0], "hands")):
		return true
	case len(tokens) >= 2 &&
		(equalWord(tokens[0], "your") || equalWord(tokens[0], "their")) &&
		(equalWord(tokens[1], "hand") || equalWord(tokens[1], "hands")):
		return true
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "its") || equalWord(tokens[0], "their")) &&
		(strings.EqualFold(tokens[1].Text, "player's") || strings.EqualFold(tokens[1].Text, "owner's") || strings.EqualFold(tokens[1].Text, "owners'")) &&
		(equalWord(tokens[2], "hand") || equalWord(tokens[2], "hands")):
		return true
	case len(tokens) >= 4 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "its") || equalWord(tokens[0], "their")) &&
		(equalWord(tokens[1], "owners") || equalWord(tokens[1], "players")) &&
		tokens[2].Kind == shared.Apostrophe &&
		(equalWord(tokens[3], "hand") || equalWord(tokens[3], "hands")):
		return true
	default:
		return false
	}
}

func battlefieldZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 1 && equalWord(tokens[0], "battlefield") ||
		len(tokens) >= 2 && equalWord(tokens[0], "the") && equalWord(tokens[1], "battlefield")
}

func libraryZonePhrase(tokens []shared.Token) bool {
	switch {
	case len(tokens) >= 1 && (equalWord(tokens[0], "library") || equalWord(tokens[0], "libraries")):
		return true
	case len(tokens) >= 2 &&
		(equalWord(tokens[0], "your") || equalWord(tokens[0], "their") ||
			equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		(equalWord(tokens[1], "library") || equalWord(tokens[1], "libraries")):
		return true
	case len(tokens) >= 2 &&
		strings.EqualFold(tokens[0].Text, "owner's") &&
		(equalWord(tokens[1], "library") || equalWord(tokens[1], "libraries")):
		return true
	case len(tokens) >= 3 &&
		equalWord(tokens[0], "an") &&
		strings.EqualFold(tokens[1].Text, "opponent's") &&
		(equalWord(tokens[2], "library") || equalWord(tokens[2], "libraries")):
		return true
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "its") || equalWord(tokens[0], "their")) &&
		(strings.EqualFold(tokens[1].Text, "owner's") || strings.EqualFold(tokens[1].Text, "owners'")) &&
		(equalWord(tokens[2], "library") || equalWord(tokens[2], "libraries")):
		return true
	default:
		return false
	}
}

func exileZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 1 && equalWord(tokens[0], "exile")
}

// commandZonePhrase recognizes "command zone" and the determined "the command
// zone" / "your command zone" forms, the origin/destination of commander-zone
// movement effects ("Put your commander into your hand from the command zone.").
func commandZonePhrase(tokens []shared.Token) bool {
	switch {
	case len(tokens) >= 2 &&
		equalWord(tokens[0], "command") &&
		equalWord(tokens[1], "zone"):
		return true
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "the") || equalWord(tokens[0], "your") || equalWord(tokens[0], "their")) &&
		equalWord(tokens[1], "command") &&
		equalWord(tokens[2], "zone"):
		return true
	default:
		return false
	}
}

// CardinalWordValue maps an Oracle cardinal number word ("one" … "ten") to its
// integer value. It owns the full small-cardinal vocabulary; callers apply
// their own numeric range policy to the typed value.
func CardinalWordValue(word string) (int, bool) {
	switch strings.ToLower(word) {
	case "one":
		return 1, true
	case "two", "twice":
		return 2, true
	case "three", "thrice":
		return 3, true
	case "four":
		return 4, true
	case "five":
		return 5, true
	case "six":
		return 6, true
	case "seven":
		return 7, true
	case "eight":
		return 8, true
	case "nine":
		return 9, true
	case "ten":
		return 10, true
	case "eleven":
		return 11, true
	case "twelve":
		return 12, true
	case "thirteen":
		return 13, true
	case "fourteen":
		return 14, true
	case "fifteen":
		return 15, true
	case "sixteen":
		return 16, true
	case "seventeen":
		return 17, true
	case "eighteen":
		return 18, true
	case "nineteen":
		return 19, true
	case "twenty":
		return 20, true
	default:
		return 0, false
	}
}

// OrdinalWordValue maps Oracle ordinal words currently recognized after parsing
// to their integer position.
func OrdinalWordValue(word string) (int, bool) {
	switch strings.ToLower(word) {
	case "first":
		return 1, true
	case "second":
		return 2, true
	case "third":
		return 3, true
	case "fourth":
		return 4, true
	case "fifth":
		return 5, true
	default:
		return 0, false
	}
}

// SelectionFlag identifies reusable selector modifiers recognized by the parser.
type SelectionFlag string

// Selection flags recognized by the parser.
const (
	SelectionFlagUnknown   SelectionFlag = ""
	SelectionFlagAnother   SelectionFlag = "SelectionFlagAnother"
	SelectionFlagOther     SelectionFlag = "SelectionFlagOther"
	SelectionFlagAttacking SelectionFlag = "SelectionFlagAttacking"
	SelectionFlagBlocking  SelectionFlag = "SelectionFlagBlocking"
	SelectionFlagTapped    SelectionFlag = "SelectionFlagTapped"
	SelectionFlagUntapped  SelectionFlag = "SelectionFlagUntapped"
	SelectionFlagToken     SelectionFlag = "SelectionFlagToken"
	SelectionFlagNonToken  SelectionFlag = "SelectionFlagNonToken"
	SelectionFlagModified  SelectionFlag = "SelectionFlagModified"
	SelectionFlagEnchanted SelectionFlag = "SelectionFlagEnchanted"
	SelectionFlagEquipped  SelectionFlag = "SelectionFlagEquipped"
)

func recognizeSelectionFlag(word string) (SelectionFlag, bool) {
	switch strings.ToLower(word) {
	case "another":
		return SelectionFlagAnother, true
	case "other":
		return SelectionFlagOther, true
	case "attacking":
		return SelectionFlagAttacking, true
	case "blocking":
		return SelectionFlagBlocking, true
	case "tapped":
		return SelectionFlagTapped, true
	case "untapped":
		return SelectionFlagUntapped, true
	case "token":
		return SelectionFlagToken, true
	case "nontoken":
		return SelectionFlagNonToken, true
	case "modified":
		return SelectionFlagModified, true
	case "enchanted":
		return SelectionFlagEnchanted, true
	case "equipped":
		return SelectionFlagEquipped, true
	default:
		return SelectionFlagUnknown, false
	}
}

// ControllerRelation identifies reusable control/ownership relation wording.
type ControllerRelation string

// Controller relations recognized by the parser.
const (
	ControllerRelationUnknown          ControllerRelation = ""
	ControllerRelationYouControl       ControllerRelation = "ControllerRelationYouControl"
	ControllerRelationYouDontControl   ControllerRelation = "ControllerRelationYouDontControl"
	ControllerRelationOpponentControls ControllerRelation = "ControllerRelationOpponentControls"
	// ControllerRelationEachOpponentControls is the distributive opponent
	// wording ("each creature each opponent controls"). It denotes the same
	// opponent-controlled set as ControllerRelationOpponentControls but a
	// different verbatim phrasing, so the byte-exact recipient reconstruction can
	// rebuild it while lowering treats both as the opponent controller.
	ControllerRelationEachOpponentControls ControllerRelation = "ControllerRelationEachOpponentControls"
	// ControllerRelationDefendingPlayerControls is the combat wording "defending
	// player controls" ("goad target creature defending player controls",
	// Coveted Peacock). It denotes the set of permanents controlled by the
	// defending player of the triggering attack.
	ControllerRelationDefendingPlayerControls ControllerRelation = "ControllerRelationDefendingPlayerControls"
	ControllerRelationYouOwn                  ControllerRelation = "ControllerRelationYouOwn"
	ControllerRelationOpponentOwns            ControllerRelation = "ControllerRelationOpponentOwns"
)

// SingularNounForms returns the candidate singular spellings for a possibly
// plural Oracle noun, most specific first, owning Oracle plural normalization
// (regular "-s", "-ies"→"-y", "-ves"→"-f"/"-fe", and "-es"). Downstream stages
// validate the returned identity candidates against their own closed subtype
// vocabularies. The input spelling is always included as the first candidate so
// already-singular nouns are returned unchanged.
func SingularNounForms(noun string) []string {
	noun = strings.TrimSpace(noun)
	candidates := []string{noun}
	if stem, ok := strings.CutSuffix(noun, "ies"); ok && stem != "" {
		candidates = append(candidates, stem+"y")
	}
	if stem, ok := strings.CutSuffix(noun, "ves"); ok && len(stem) > 1 {
		candidates = append(candidates, stem+"f", stem+"fe")
	}
	if stem, ok := strings.CutSuffix(noun, "es"); ok && stem != "" {
		candidates = append(candidates, stem)
	}
	if stem, ok := strings.CutSuffix(noun, "s"); ok && len(stem) > 1 {
		candidates = append(candidates, stem)
	}
	return candidates
}

// SubtypeMatchesCardType reports whether a parser-emitted subtype identity is
// legal for the parser card-type atom. It keeps subtype family validation on the
// parser side for compiler consumers.
func SubtypeMatchesCardType(sub types.Sub, cardType CardType) bool {
	switch cardType {
	case CardTypeArtifact:
		return types.KnownSubtypeForType(types.Artifact, sub)
	case CardTypeBattle:
		return types.KnownSubtypeForType(types.Battle, sub)
	case CardTypeCreature:
		return types.KnownSubtypeForType(types.Creature, sub)
	case CardTypeEnchantment:
		return types.KnownSubtypeForType(types.Enchantment, sub)
	case CardTypeInstant:
		return types.KnownSubtypeForType(types.Instant, sub)
	case CardTypeLand:
		return types.KnownSubtypeForType(types.Land, sub)
	case CardTypePlaneswalker:
		return types.KnownSubtypeForType(types.Planeswalker, sub)
	case CardTypeSorcery:
		return types.KnownSubtypeForType(types.Sorcery, sub)
	default:
		return false
	}
}

// SubtypeMatchesAnyRuntimeCardType reports whether sub is defined for any of the
// supplied runtime card-type families. Compiler and lowering code use this
// parser-owned check instead of reconstructing subtype meaning from source
// spelling.
func SubtypeMatchesAnyRuntimeCardType(sub types.Sub, cardTypes []types.Card) bool {
	for _, cardType := range cardTypes {
		if types.KnownSubtypeForType(cardType, sub) {
			return true
		}
	}
	return false
}

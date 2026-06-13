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
type Color uint8

// Oracle colors recognized by the parser.
const (
	ColorUnknown Color = iota
	ColorWhite
	ColorBlue
	ColorBlack
	ColorRed
	ColorGreen
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

// ColorQualifier is a typed Oracle color-family qualifier that is not itself a
// single color.
type ColorQualifier uint8

// Oracle color-family qualifiers recognized by the parser.
const (
	ColorQualifierUnknown ColorQualifier = iota
	ColorQualifierColorless
	ColorQualifierMulticolored
	ColorQualifierMonocolored
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
type CardType uint8

// Oracle card types recognized by the parser.
const (
	CardTypeUnknown CardType = iota
	CardTypeArtifact
	CardTypeBattle
	CardTypeCreature
	CardTypeEnchantment
	CardTypeInstant
	CardTypeLand
	CardTypePlaneswalker
	CardTypeSorcery
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

// Supertype is a typed Oracle supertype atom.
type Supertype uint8

// Oracle supertypes recognized by the parser.
const (
	SupertypeUnknown Supertype = iota
	SupertypeLegendary
	SupertypeSnow
	SupertypeBasic
	SupertypeWorld
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

// ObjectNoun is a typed Oracle object-noun atom: the reusable nouns that name a
// game object or player. Downstream stages decide which nouns are valid in a
// given grammar from the typed value rather than from spelling.
type ObjectNoun uint8

// Oracle object nouns recognized by the parser.
const (
	ObjectNounUnknown ObjectNoun = iota
	ObjectNounAbility
	ObjectNounArtifact
	ObjectNounCard
	ObjectNounCreature
	ObjectNounEnchantment
	ObjectNounEquipment
	ObjectNounLand
	ObjectNounPermanent
	ObjectNounPlaneswalker
	ObjectNounOpponent
	ObjectNounPlayer
	ObjectNounSpell
	ObjectNounToken
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
	default:
		return false
	}
}

func battlefieldZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 1 && equalWord(tokens[0], "battlefield") ||
		len(tokens) >= 2 && equalWord(tokens[0], "the") && equalWord(tokens[1], "battlefield")
}

func libraryZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 1 && equalWord(tokens[0], "library") ||
		len(tokens) >= 2 && equalWord(tokens[0], "your") && equalWord(tokens[1], "library")
}

func exileZonePhrase(tokens []shared.Token) bool {
	return len(tokens) >= 1 && equalWord(tokens[0], "exile")
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
type SelectionFlag uint8

// Selection flags recognized by the parser.
const (
	SelectionFlagUnknown SelectionFlag = iota
	SelectionFlagAnother
	SelectionFlagOther
	SelectionFlagAttacking
	SelectionFlagBlocking
	SelectionFlagTapped
	SelectionFlagUntapped
	SelectionFlagToken
	SelectionFlagNonToken
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
	default:
		return SelectionFlagUnknown, false
	}
}

// ControllerRelation identifies reusable control/ownership relation wording.
type ControllerRelation uint8

// Controller relations recognized by the parser.
const (
	ControllerRelationUnknown ControllerRelation = iota
	ControllerRelationYouControl
	ControllerRelationYouDontControl
	ControllerRelationOpponentControls
	ControllerRelationYouOwn
	ControllerRelationOpponentOwns
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

package parser

import "strings"

// KeepScope identifies which players a "keep one of each type" sacrifice affects.
type KeepScope uint8

const (
	// KeepScopeUnknown is the zero value for an unrecognized scope.
	KeepScopeUnknown KeepScope = iota
	// KeepScopeOpponents affects each opponent of the source's controller
	// ("Each opponent chooses ..." — Liliana, Dreadhorde General's −9).
	KeepScopeOpponents
	// KeepScopeAllPlayers affects every player ("Each player chooses ..." —
	// Cataclysm, Cataclysmic Gearhulk).
	KeepScopeAllPlayers
)

// KeepOnePerTypeSyntax is the typed payload of a recognized "keep one of each
// type" sacrifice sentence: each affected player keeps one permanent of each
// listed type and sacrifices the rest.
type KeepOnePerTypeSyntax struct {
	// Scope is the set of affected players.
	Scope KeepScope
	// Types is the ordered list of permanent types, one kept permanent per type.
	Types []CardType
	// NonlandOnly reports that the affected pool is the nonland permanents each
	// player controls ("from among the nonland permanents they control" —
	// Cataclysmic Gearhulk) rather than every permanent they control.
	NonlandOnly bool
	// ControllerChoosesForAll reports that the effect's controller ("you")
	// chooses the kept permanent of each type for every affected player, rather
	// than each player choosing among their own permanents. It backs the
	// two-sentence "For each player, you choose ... Then each player sacrifices
	// all other nonland permanents they control." form (Tragic Arrogance).
	ControllerChoosesForAll bool
}

// exactKeepOnePerTypeSacrificeEffectSyntax recognizes the "keep one of each
// type" sacrifice family verbatim and fails closed on any deviation. Three exact
// templates are supported, each parameterized by the affected-player scope ("Each
// opponent" or "Each player"):
//
//	<Subject> chooses a permanent they control of each permanent type and sacrifices the rest.
//	<Subject> chooses from among the permanents they control <list>, then sacrifices the rest.
//	<Subject> chooses <list> from among the nonland permanents they control, then sacrifices the rest.
//
// where <list> is a canonical permanent-type enumeration ("an artifact, a
// creature, an enchantment, and a land"). The parsed types are reconstructed and
// compared byte-exact so only the printed wording is accepted; any altered
// phrasing, article, ordering, or non-permanent type is rejected and the
// sentence is left unrecognized for the downstream fail-closed guard.
func exactKeepOnePerTypeSacrificeEffectSyntax(effect *EffectSyntax) bool {
	if effect.Negated || effect.Optional || effect.Amount.Known ||
		len(effect.Targets) != 0 || effect.Duration != EffectDurationNone {
		return false
	}
	text := joinedEffectText(effect.Tokens)
	for _, scope := range []struct {
		word  string
		value KeepScope
	}{
		{"Each opponent", KeepScopeOpponents},
		{"Each player", KeepScopeAllPlayers},
	} {
		if strings.EqualFold(text, scope.word+" chooses a permanent they control of each permanent type and sacrifices the rest.") {
			effect.KeepOnePerType = &KeepOnePerTypeSyntax{Scope: scope.value, Types: allPermanentCardTypes()}
			return true
		}
		if list, ok := cutFold(text, scope.word+" chooses from among the permanents they control ", ", then sacrifices the rest."); ok {
			if kept, ok := exactPermanentTypeList(list); ok {
				effect.KeepOnePerType = &KeepOnePerTypeSyntax{Scope: scope.value, Types: kept}
				return true
			}
		}
		if list, ok := cutFold(text, scope.word+" chooses ", " from among the nonland permanents they control, then sacrifices the rest."); ok {
			if kept, ok := exactPermanentTypeList(list); ok {
				effect.KeepOnePerType = &KeepOnePerTypeSyntax{Scope: scope.value, Types: kept, NonlandOnly: true}
				return true
			}
		}
	}
	return false
}

// allPermanentCardTypes returns the six permanent card types in the canonical
// rules order, backing the "of each permanent type" form.
func allPermanentCardTypes() []CardType {
	return []CardType{
		CardTypeArtifact,
		CardTypeBattle,
		CardTypeCreature,
		CardTypeEnchantment,
		CardTypeLand,
		CardTypePlaneswalker,
	}
}

// cutFold returns the substring of text between prefix and suffix, matching each
// end case-insensitively (the leading word is capitalized on a standalone
// sentence but lower-case when the clause follows a trigger). It reports false
// when either end does not match. It relies on prefix and suffix being ASCII, so
// their byte lengths index the text safely.
func cutFold(text, prefix, suffix string) (string, bool) {
	if len(text) < len(prefix)+len(suffix) {
		return "", false
	}
	if !strings.EqualFold(text[:len(prefix)], prefix) ||
		!strings.EqualFold(text[len(text)-len(suffix):], suffix) {
		return "", false
	}
	return text[len(prefix) : len(text)-len(suffix)], true
}

// exactPermanentTypeList parses a canonical permanent-type enumeration ("an
// artifact, a creature, an enchantment, and a land") into its ordered types. It
// reconstructs the canonical wording from the parsed types and compares it
// byte-exact to the input, so a non-canonical enumeration — a missing Oxford
// comma, a wrong article, a duplicated or non-permanent type — is rejected. At
// least two types are required, matching every printed member of the family.
func exactPermanentTypeList(list string) ([]CardType, bool) {
	fields := strings.Fields(list)
	var kept []CardType
	for i := 0; i < len(fields); {
		if strings.EqualFold(fields[i], "and") {
			i++
			if i >= len(fields) {
				return nil, false
			}
		}
		article := fields[i]
		if !strings.EqualFold(article, "a") && !strings.EqualFold(article, "an") {
			return nil, false
		}
		i++
		if i >= len(fields) {
			return nil, false
		}
		word := strings.TrimSuffix(fields[i], ",")
		i++
		cardType, ok := recognizeCardTypeWord(word)
		if !ok || !isPermanentCardType(cardType) {
			return nil, false
		}
		kept = append(kept, cardType)
	}
	if len(kept) < 2 {
		return nil, false
	}
	canonical, ok := renderPermanentTypeList(kept)
	if !ok || !strings.EqualFold(list, canonical) {
		return nil, false
	}
	return kept, true
}

// renderPermanentTypeList reconstructs the canonical Oracle enumeration of the
// given permanent types ("an artifact, a creature, an enchantment, and a land"),
// used to verify a parsed list is exactly the printed wording.
func renderPermanentTypeList(kept []CardType) (string, bool) {
	words := make([]string, len(kept))
	for i, cardType := range kept {
		word, ok := permanentCardTypeWord(cardType)
		if !ok {
			return "", false
		}
		words[i] = indefiniteArticle(word) + " " + word
	}
	switch len(words) {
	case 0:
		return "", false
	case 1:
		return words[0], true
	case 2:
		return words[0] + " and " + words[1], true
	default:
		return strings.Join(words[:len(words)-1], ", ") + ", and " + words[len(words)-1], true
	}
}

// isPermanentCardType reports whether cardType is one of the six permanent card
// types (artifact, battle, creature, enchantment, land, planeswalker).
func isPermanentCardType(cardType CardType) bool {
	switch cardType {
	case CardTypeArtifact, CardTypeBattle, CardTypeCreature,
		CardTypeEnchantment, CardTypeLand, CardTypePlaneswalker:
		return true
	default:
		return false
	}
}

// permanentCardTypeWord returns the lower-case singular Oracle word for a
// permanent card type.
func permanentCardTypeWord(cardType CardType) (string, bool) {
	switch cardType {
	case CardTypeArtifact:
		return "artifact", true
	case CardTypeBattle:
		return "battle", true
	case CardTypeCreature:
		return "creature", true
	case CardTypeEnchantment:
		return "enchantment", true
	case CardTypeLand:
		return "land", true
	case CardTypePlaneswalker:
		return "planeswalker", true
	default:
		return "", false
	}
}

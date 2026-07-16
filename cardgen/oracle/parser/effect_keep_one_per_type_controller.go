package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// controllerChoosesKeepPreludePrefix and controllerChoosesKeepSacrificeText are
// the two verbatim sentences of the controller-chooses keep-one-per-type form
// (Tragic Arrogance):
//
//	For each player, you choose from among the permanents that player controls <list>.
//	Then each player sacrifices all other nonland permanents they control.
//
// where <list> is a canonical permanent-type enumeration ("an artifact, a
// creature, an enchantment, and a planeswalker"). Unlike the single-sentence
// family members, the controller ("you") chooses the kept permanent of each type
// for every player, and only nonland permanents are sacrificed.
const (
	controllerChoosesKeepPreludePrefix = "For each player, you choose from among the permanents that player controls "
	controllerChoosesKeepSacrificeText = "Then each player sacrifices all other nonland permanents they control."
	controllerChoosesKeepPreludeSuffix = "."
)

// recognizeControllerChoosesKeepSequence folds the two-sentence controller-
// chooses keep-one-per-type form (Tragic Arrogance) into the generic
// KeepOnePerType payload. It requires the exact zero-effect "For each player, you
// choose from among the permanents that player controls <list>." prelude
// immediately followed by the exact "Then each player sacrifices all other
// nonland permanents they control." sacrifice. When both match, it records the
// affected scope (all players), the ordered permanent types, the nonland-only
// pool, and the controller-chooses flag on the sacrifice effect, marks that
// effect exact, and credits the prelude sentence so coverage and reference scans
// treat its tokens as belonging to the folded effect. Any other wording fails
// closed. It returns true only when the full shape matches.
func recognizeControllerChoosesKeepSequence(sentences []Sentence) bool {
	preludeIndex, kept, ok := controllerChoosesKeepPrelude(sentences)
	if !ok {
		return false
	}
	sacrificeIndex := preludeIndex + 1
	if sacrificeIndex >= len(sentences) || len(sentences[sacrificeIndex].Effects) != 1 {
		return false
	}
	effect := &sentences[sacrificeIndex].Effects[0]
	if effect.Kind != EffectSacrifice ||
		!strings.EqualFold(joinedEffectText(effect.Tokens), controllerChoosesKeepSacrificeText) {
		return false
	}
	effect.KeepOnePerType = &KeepOnePerTypeSyntax{
		Scope:                   KeepScopeAllPlayers,
		Types:                   kept,
		NonlandOnly:             true,
		ControllerChoosesForAll: true,
	}
	effect.Exact = true
	effect.ControllerChoosesKeepPreludeSpan = sentences[preludeIndex].Span
	sentences[preludeIndex].ControllerChoosesKeepPrelude = true
	return true
}

// controllerChoosesKeepPrelude returns the index and ordered permanent types of
// the first zero-effect "For each player, you choose from among the permanents
// that player controls <list>." prelude sentence, reporting whether one matches.
func controllerChoosesKeepPrelude(sentences []Sentence) (int, []CardType, bool) {
	for i := range sentences {
		if len(sentences[i].Effects) != 0 {
			continue
		}
		if kept, ok := controllerChoosesKeepPreludeTypes(sentences[i].Tokens); ok {
			return i, kept, true
		}
	}
	return 0, nil, false
}

// controllerChoosesKeepPreludeTypes parses the ordered permanent types from a
// "For each player, you choose from among the permanents that player controls
// <list>." prelude, reconstructing the canonical wording byte-exact so only the
// printed enumeration is accepted.
func controllerChoosesKeepPreludeTypes(tokens []shared.Token) ([]CardType, bool) {
	list, ok := cutFold(joinedEffectText(tokens), controllerChoosesKeepPreludePrefix, controllerChoosesKeepPreludeSuffix)
	if !ok {
		return nil, false
	}
	return exactPermanentTypeList(list)
}

// isControllerChoosesKeepPreludeTokens reports whether the sentence tokens are a
// verbatim controller-chooses keep-one-per-type prelude, gating the zero-effect
// sibling as a benign prelude candidate rather than an unrecognized sibling.
func isControllerChoosesKeepPreludeTokens(tokens []shared.Token) bool {
	_, ok := controllerChoosesKeepPreludeTypes(tokens)
	return ok
}

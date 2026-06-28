package parser

import (
	"strings"
)

// recognizeRevealTopPartitionSequence recognizes the closed effect family
// "Reveal the top N cards of your library. Put all <type> cards revealed this
// way into your hand and the rest <remainder>." (Borborygmos Enraged, Sift
// Through Sands, the tribal "reveal and gather" creatures). The controller
// reveals the top N cards of their own library, puts every revealed card
// matching the named card-type or subtype filter into their hand, and routes the
// rest to their graveyard or the bottom of their library.
//
// The parser owns the wording: it confirms the two-effect [Reveal, Put] shape,
// that the reveal draws a fixed top N from the controller's library, that the
// put clause moves "all <filter> revealed this way / from among them into your
// hand and the rest <remainder>", and extracts the remainder destination. It
// marks both effects with RevealTopPartition and records the remainder on the
// put effect; the typed count stays on the reveal's Amount and the typed filter
// on the put's Selection. The text-blind lowering reads only those typed fields.
// Any other shape, filter, or remainder fails closed, leaving the generic
// sequence path untouched.
func recognizeRevealTopPartitionSequence(sentences []Sentence) {
	effects := orderedRevealUntilEffects(sentences)
	if len(effects) != 2 {
		return
	}
	reveal := effects[0]
	put := effects[1]
	if reveal.Kind != EffectReveal || put.Kind != EffectPut ||
		reveal.Context != EffectContextController ||
		put.Context != EffectContextController {
		return
	}
	if !isRevealTopOfYourLibrary(reveal) {
		return
	}
	remainder, ok := revealPartitionPutShape(put)
	if !ok {
		return
	}
	reveal.Exact = true
	reveal.RevealTopPartition = true
	put.Exact = true
	put.RevealTopPartition = true
	put.RevealPartitionRemainder = remainder
}

// isRevealTopOfYourLibrary reports whether the reveal effect is "reveal the top
// N cards of your library" with a plain fixed positive count. Variable, range,
// or rules-derived counts the fixed-count primitive cannot model fail closed.
func isRevealTopOfYourLibrary(reveal *EffectSyntax) bool {
	if !pileSplitFixedAmount(reveal.Amount) {
		return false
	}
	text := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(reveal.Selection.Text)), ".")
	before, after, ok := strings.Cut(text, " cards of your library")
	if !ok || after != "" {
		return false
	}
	return strings.HasPrefix(before, "the top ")
}

// revealPartitionPutShape validates the put clause "all <filter> {revealed this
// way|from among them} into your hand and the rest <remainder>" and returns the
// remainder destination. It requires the typed All filter to carry a concrete
// card-type or subtype constraint and the anaphoric "revealed this way" / "from
// among them" reference back to the revealed cards. Any other shape fails closed.
func revealPartitionPutShape(put *EffectSyntax) (DigRemainderKind, bool) {
	selection := put.Selection
	if !selection.All {
		return DigRemainderGraveyard, false
	}
	if len(selection.RequiredTypesAny) == 0 && len(selection.SubtypesAny) == 0 {
		return DigRemainderGraveyard, false
	}
	if selection.ConjunctiveTypes {
		return DigRemainderGraveyard, false
	}
	text := strings.TrimSuffix(strings.TrimSpace(strings.ToLower(selection.Text)), ".")
	if !strings.HasPrefix(text, "all ") {
		return DigRemainderGraveyard, false
	}
	if !strings.Contains(text, "revealed this way") && !strings.Contains(text, "from among them") {
		return DigRemainderGraveyard, false
	}
	_, rest, ok := strings.Cut(text, " into your hand and the rest ")
	if !ok {
		return DigRemainderGraveyard, false
	}
	return revealPartitionRemainder(strings.TrimSpace(rest))
}

// revealPartitionRemainder maps the put clause's remainder phrasing to a typed
// remainder destination. Only the controller's graveyard and the bottom of the
// controller's library (in any order or a random order) are modeled.
func revealPartitionRemainder(rest string) (DigRemainderKind, bool) {
	switch rest {
	case "into your graveyard":
		return DigRemainderGraveyard, true
	case "on the bottom of your library in any order":
		return DigRemainderLibraryBottomAny, true
	case "on the bottom of your library in a random order":
		return DigRemainderLibraryBottomRandom, true
	default:
		return DigRemainderGraveyard, false
	}
}

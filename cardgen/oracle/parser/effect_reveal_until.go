package parser

import (
	"strings"

	"github.com/natefinch/council4/mtg/game/zone"
)

// recognizeRevealUntilThenPutSequence recognizes the closed effect family
// "reveal cards from the top of <library> until <player> reveal a <type> card,
// then put those cards into <zone>" (Undercity Informer, Balustrade Spy,
// Treasure Hunt). The reveal proceeds from the top of one player's library and
// stops once a card matching the named type is revealed; every revealed card is
// then moved as a group into one destination zone belonging to the same player.
//
// The parser owns the wording: it confirms the three-effect [Reveal, Reveal,
// Put] shape, that the head reveal draws from the top of a library "until" a
// boundary, that the match reveal names the boundary card type, and that the put
// clause moves "those cards" into the revealing player's graveyard or hand. It
// marks each effect with RevealUntilThenPut and records the destination on the
// put effect's ToZone so the text-blind lowering can emit a single RevealUntil
// primitive. Destinations other than graveyard and hand (battlefield, library
// bottom, exile) and cross-player destinations fail closed here.
func recognizeRevealUntilThenPutSequence(sentences []Sentence) {
	effects := orderedRevealUntilEffects(sentences)
	if len(effects) != 3 {
		return
	}
	revealUntil := effects[0]
	matchReveal := effects[1]
	put := effects[2]
	if revealUntil.Kind != EffectReveal ||
		matchReveal.Kind != EffectReveal ||
		put.Kind != EffectPut {
		return
	}

	revealText := strings.ToLower(revealUntil.Selection.Text)
	if !strings.Contains(revealText, "from the top of") ||
		!strings.Contains(revealText, "library") ||
		!strings.Contains(revealText, "until") {
		return
	}
	possessive, ok := revealedLibraryPossessive(revealText)
	if !ok {
		return
	}

	if len(matchReveal.Selection.RequiredTypesAny) == 0 &&
		len(matchReveal.Selection.ExcludedTypes) == 0 &&
		len(matchReveal.Selection.SubtypesAny) == 0 {
		return
	}

	putText := strings.ToLower(put.Selection.Text)
	if !strings.Contains(putText, "those cards") &&
		!strings.Contains(putText, "cards revealed this way") {
		return
	}
	// Split-destination forms ("put that card into <zone> and all other cards
	// revealed this way into <zone>", Hermit Druid) route the matching card and
	// the remainder to different zones. The single RevealUntil primitive moves
	// every revealed card to one destination, so any split marker fails closed.
	if strings.Contains(putText, "that card") ||
		strings.Contains(putText, "all other") ||
		strings.Contains(putText, "the rest") ||
		strings.Contains(putText, "rest of") {
		return
	}
	destination, ok := revealUntilDestination(putText, possessive)
	if !ok {
		return
	}

	revealUntil.Exact = true
	revealUntil.RevealUntilThenPut = true
	matchReveal.Exact = true
	matchReveal.RevealUntilThenPut = true
	put.Exact = true
	put.RevealUntilThenPut = true
	put.ToZone = destination
}

// orderedRevealUntilEffects returns pointers to every resolving effect across
// the ability's sentences in source order.
func orderedRevealUntilEffects(sentences []Sentence) []*EffectSyntax {
	var effects []*EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effects = append(effects, &sentences[i].Effects[j])
		}
	}
	return effects
}

// revealedLibraryPossessive extracts the possessive that owns the revealed
// library ("from the top of <possessive> library"). It returns false for any
// owner the runtime cannot resolve from the revealing player alone.
func revealedLibraryPossessive(text string) (string, bool) {
	_, after, ok := strings.Cut(text, "from the top of ")
	if !ok {
		return "", false
	}
	before, _, ok := strings.Cut(after, " library")
	if !ok {
		return "", false
	}
	possessive := strings.TrimSpace(before)
	switch possessive {
	case "their", "your", "its", "his or her":
		return possessive, true
	}
	return "", false
}

// revealUntilDestination resolves the put clause's destination zone, requiring
// the destination to belong to the same player whose library was revealed
// (matching possessive). Only graveyard and hand are modeled; every other zone
// fails closed.
func revealUntilDestination(putText, possessive string) (zone.Type, bool) {
	if strings.Contains(putText, possessive+" graveyard") {
		return zone.Graveyard, true
	}
	if strings.Contains(putText, possessive+" hand") {
		return zone.Hand, true
	}
	return zone.None, false
}

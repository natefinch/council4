package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseChooseFromEachGraveyardEffect recognizes the mass reanimation base "For
// each player, choose a creature [or planeswalker] card in that player's
// graveyard." (Breach the Multiverse) and its optional "up to one" variant "For
// each player, choose up to one creature card in that player's graveyard." The
// whole sentence collapses to one EffectChooseFromEachGraveyard whose Selection
// carries the card-type filter parsed from the "<type> card" sub-phrase and whose
// Optional records the "up to one" wording, letting the lowering stay text-blind.
// A following "Put those cards onto the battlefield under your control."
// reanimates the chosen cards.
//
// The choice is a non-targeted choose made as the spell resolves. Targeted
// variants ("... choose up to one target creature card ...", Afterlife from the
// Loam, The Moonbase) are a different mechanism, so the "target" wording fails
// closed here and flows through the generic effect parser rather than
// round-tripping as a silent non-targeted reanimation. Any other unrecognized
// shape likewise fails closed.
func parseChooseFromEachGraveyardEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Word {
			words = append(words, token)
		}
	}
	// Shortest match: "for each player choose <>=1 selection word> in that
	// player's graveyard" is 4 + 1 + 4 = 9 words.
	if len(words) < 9 {
		return nil, false
	}
	if !equalWord(words[0], "for") || !equalWord(words[1], "each") ||
		!equalWord(words[2], "player") || !equalWord(words[3], "choose") {
		return nil, false
	}
	suffix := words[len(words)-4:]
	if !equalWord(suffix[0], "in") || !equalWord(suffix[1], "that") ||
		!equalWord(suffix[2], "player's") || !equalWord(suffix[3], "graveyard") {
		return nil, false
	}
	optional := false
	selStart := 4
	if len(words) >= 12 && equalWord(words[4], "up") && equalWord(words[5], "to") && equalWord(words[6], "one") {
		optional = true
		selStart = 7
	}
	selectionWords := words[selStart : len(words)-4]
	if len(selectionWords) == 0 {
		return nil, false
	}
	// This primitive chooses without targeting, so a targeted variant must not be
	// mistaken for it; fail closed on any "target" wording so it flows to the
	// generic parser instead of silently dropping the targeting.
	for _, word := range selectionWords {
		if equalWord(word, "target") || equalWord(word, "targets") {
			return nil, false
		}
	}
	// The "<type> card" phrase must actually name at least one card type, or the
	// selection would match nothing meaningful; fail closed so an unhandled
	// wording flows to the generic parser instead of silently reanimating.
	selection := parseSelection(selectionWords, atoms)
	if len(selection.RequiredTypesAny) == 0 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectChooseFromEachGraveyard,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[3].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextEachPlayer,
		Selection:  selection,
		Optional:   optional,
		Exact:      true,
	}}, true
}

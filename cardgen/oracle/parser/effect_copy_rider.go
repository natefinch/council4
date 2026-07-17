package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// creditCopySetColorsRider folds a trailing copiable color exception
// ", except that the copy is <color>." onto a copy-stack-object effect. The
// exception is removed from the effect's bare copy clause before exactness is
// recomputed, while its typed color and span remain available to downstream
// text-blind stages.
func creditCopySetColorsRider(sentences []Sentence) {
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			if effect.Kind != EffectCopyStackObject || len(effect.CopySetColors) != 0 {
				continue
			}
			riderStart, setColor, ok := copySetColorRider(effect.Tokens)
			if !ok {
				continue
			}
			riderSpan := shared.SpanOf(effect.Tokens[riderStart:])
			effect.Tokens = effect.Tokens[:riderStart]
			effect.CopySetColors = []Color{setColor}
			effect.CopySetColorsRiderSpan = riderSpan
			effect.Exact = exactEffectSyntax(effect)
		}
	}
}

func copySetColorRider(tokens []shared.Token) (int, Color, bool) {
	for i := range tokens {
		if !equalWord(tokens[i], "except") {
			continue
		}
		j := i + 1
		if j < len(tokens) && equalWord(tokens[j], "that") {
			j++
		}
		if !effectWordsAt(tokens, j, "the", "copy", "is") {
			continue
		}
		j += 3
		if j >= len(tokens) {
			continue
		}
		setColor, ok := recognizeColorWord(tokens[j].Text)
		if !ok {
			continue
		}
		j++
		if j < len(tokens) && tokens[j].Kind == shared.Period {
			j++
		}
		if j != len(tokens) {
			continue
		}
		start := i
		if start > 0 && tokens[start-1].Kind == shared.Comma {
			start--
		}
		return start, setColor, true
	}
	return 0, ColorUnknown, false
}

// creditConjoinedCopyChooseNewTargetsRider folds a "copy <ref> and [may] choose
// a new target[s]/new targets for the copy[ies]" rider that the parser absorbs
// into the copy effect's own clause (Sevinne's Reclamation, the Chain cycle).
// Unlike the separate-sentence "You may choose new targets for the copy." rider
// (creditCopyChooseNewTargetsRider), the conjoined "and ... choose ..." tail
// shares the copy effect's sentence, so the parser reads "a new target" as a
// spurious target and amount on the copy effect and the clause never matches the
// exact bare-copy wording. This rewrites the copy effect to its bare
// "Copy <ref>." form: it drops the rider tail tokens, the spurious target, and
// the spurious amount, sets CopyMayChooseNewTargets, records the rider span for
// coverage, and recomputes exactness.
func creditConjoinedCopyChooseNewTargetsRider(sentences []Sentence) {
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			if effect.Kind != EffectCopyStackObject || effect.CopyMayChooseNewTargets {
				continue
			}
			riderStart, ok := conjoinedCopyRiderStart(effect.Tokens)
			if !ok {
				continue
			}
			riderSpan := shared.SpanOf(effect.Tokens[riderStart:])
			effect.Tokens = effect.Tokens[:riderStart]
			effect.Targets = targetsBeforeOffset(effect.Targets, riderSpan.Start.Offset)
			sentences[si].Targets = targetsBeforeOffset(sentences[si].Targets, riderSpan.Start.Offset)
			effect.Amount = EffectAmountSyntax{}
			effect.CopyMayChooseNewTargets = true
			effect.CopyChooseNewTargetsRiderSpan = riderSpan
			effect.Exact = exactEffectSyntax(effect)
		}
	}
}

// conjoinedCopyRiderStart returns the index of the "and" token that begins a
// trailing "and [may] choose a new target[s]/new targets for the copy[ies][.]"
// rider closing the clause, and reports whether one was found. The rider must
// run to the end of the effect tokens (only a trailing period may follow) so a
// mid-clause "and" introducing further effects is never mistaken for it.
func conjoinedCopyRiderStart(tokens []shared.Token) (int, bool) {
	for i := range tokens {
		if !equalWord(tokens[i], "and") {
			continue
		}
		if conjoinedCopyRiderTail(tokens[i+1:]) {
			return i, true
		}
	}
	return 0, false
}

// conjoinedCopyRiderTail reports whether body is exactly "[may] choose a new
// target[s]/new targets for the copy[ies]" optionally closed by a period.
func conjoinedCopyRiderTail(body []shared.Token) bool {
	j := 0
	if j < len(body) && equalWord(body[j], "may") {
		j++
	}
	if j >= len(body) || !equalWord(body[j], "choose") {
		return false
	}
	j++
	switch {
	case effectWordsAt(body, j, "a", "new", "target"):
		j += 3
	case effectWordsAt(body, j, "new", "targets"):
		j += 2
	default:
		return false
	}
	// The demonstrative closing the rider is "the copy" (Sevinne's Reclamation)
	// or "that copy" (the Chain cycle's "a new target for that copy"); both name
	// the freshly created copy, so accept either.
	if j >= len(body) || !equalWord(body[j], "for") {
		return false
	}
	j++
	if j >= len(body) || (!equalWord(body[j], "the") && !equalWord(body[j], "that")) {
		return false
	}
	j++
	if j >= len(body) || (!equalWord(body[j], "copy") && !equalWord(body[j], "copies")) {
		return false
	}
	j++
	if j == len(body) {
		return true
	}
	return j+1 == len(body) && body[j].Kind == shared.Period
}

// targetsBeforeOffset returns the targets whose span starts before offset,
// dropping any target introduced at or after it.
func targetsBeforeOffset(targets []TargetSyntax, offset int) []TargetSyntax {
	var kept []TargetSyntax
	for _, target := range targets {
		if target.Span.Start.Offset < offset {
			kept = append(kept, target)
		}
	}
	return kept
}

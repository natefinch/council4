package parser

import (
	"slices"
	"strings"
)

// kickerScaledSecondTargetPhrases are the canonical spellings of the trailing
// "another target for each time this spell was kicked" slot that scales a spell's
// target count with its Multikicker count (CR 702.32). The bare "it was kicked"
// spelling is admitted alongside the explicit "this spell was kicked" spelling so
// both Oracle wordings fold identically.
var kickerScaledSecondTargetPhrases = [][]string{
	{"another", "target", "for", "each", "time", "this", "spell", "was", "kicked"},
	{"another", "target", "for", "each", "time", "it", "was", "kicked"},
}

// recognizeKickerScaledTargetPreambleSentence folds the two-target "Choose <T>,
// then choose another target for each time this spell was kicked." preamble
// (Comet Storm) into a single target whose chosen count is one plus the paid
// Multikicker count. The preamble declares no standalone effect; a following
// "<source> deals X damage to each of them." sentence consumes the targets. It
// fires only for the exact two-target shape whose first slot is an ordinary exact
// target and whose second slot is the kicker-scaled "another target" phrase,
// leaving every other target list unchanged so unrelated wordings keep their
// two-target parse.
func recognizeKickerScaledTargetPreambleSentence(sentence *Sentence) {
	if len(sentence.Effects) != 0 || len(sentence.Targets) != 2 {
		return
	}
	if !sentence.Targets[0].Exact {
		return
	}
	if !isKickerScaledSecondTargetPhrase(sentence.Targets[1].Text) {
		return
	}
	sentence.Targets[0].KickerScaledCount = true
	// Extend the surviving target's span to the end of the folded "another
	// target for each time this spell was kicked" slot so executable-lowering
	// coverage still accounts for every token of the preamble sentence. The
	// connective ("then choose") and the second slot both fall within
	// [firstStart, secondEnd], so a single widened span covers them.
	if sentence.Targets[1].Span.End.Offset > sentence.Targets[0].Span.End.Offset {
		sentence.Targets[0].Span.End = sentence.Targets[1].Span.End
	}
	sentence.Targets = sentence.Targets[:1]
}

// isKickerScaledSecondTargetPhrase reports whether text is one of the canonical
// "another target for each time [this spell|it] was kicked" spellings.
func isKickerScaledSecondTargetPhrase(text string) bool {
	words := strings.Fields(strings.ToLower(text))
	for _, phrase := range kickerScaledSecondTargetPhrases {
		if slices.Equal(words, phrase) {
			return true
		}
	}
	return false
}

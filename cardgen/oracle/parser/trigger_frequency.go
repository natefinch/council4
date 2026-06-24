package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// TriggerFrequencyKind identifies a recognized per-turn cap on how many times a
// triggered ability may trigger, taken from a trailing "This ability triggers
// only <count> each turn." qualifier.
type TriggerFrequencyKind string

// Trigger frequency caps recognized by the parser.
const (
	TriggerFrequencyUnknown      TriggerFrequencyKind = ""
	TriggerFrequencyOncePerTurn  TriggerFrequencyKind = "TriggerFrequencyOncePerTurn"
	TriggerFrequencyTwicePerTurn TriggerFrequencyKind = "TriggerFrequencyTwicePerTurn"
)

// TriggerFrequencyRestriction is a recognized trailing "This ability triggers
// only once/twice each turn." qualifier on a triggered ability. Downstream
// stages consume this typed value by span instead of re-reading the qualifier
// wording.
type TriggerFrequencyRestriction struct {
	Kind TriggerFrequencyKind `json:",omitempty"`
	Span shared.Span          `json:"-"`
}

// parseTrailingTriggerFrequency recognizes a trailing "This ability triggers
// only once/twice each turn." sentence in a triggered ability's resolving body
// and returns its typed restriction, or nil when no such sentence is present.
// A trailing parenthetical reminder sentence (such as the "commit a crime"
// reminder) carries no rules meaning, so it is skipped when locating the last
// rules-bearing sentence.
func parseTrailingTriggerFrequency(source string, tokens []shared.Token) *TriggerFrequencyRestriction {
	sentences := ParseSentences(source, tokens)
	for i := len(sentences) - 1; i >= 0; i-- {
		if len(semanticEffectTokens(sentences[i].Tokens)) == 0 {
			continue
		}
		restriction, ok := parseTriggerFrequencyRestriction(sentences[i].Tokens)
		if !ok {
			return nil
		}
		return &restriction
	}
	return nil
}

func parseTriggerFrequencyRestriction(tokens []shared.Token) (TriggerFrequencyRestriction, bool) {
	fullSpan := shared.SpanOf(tokens)
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	rest, ok := cutSyntaxWords(tokens, "this", "ability", "triggers", "only")
	if !ok {
		// "Do this only once each turn." (Terrasymbiosis) caps the resolving
		// effect to one execution per turn; for an ability whose only effect is
		// the capped action this is observationally the once-per-turn trigger
		// cap, so it lowers through the same typed restriction.
		rest, ok = cutSyntaxWords(tokens, "do", "this", "only")
	}
	if !ok || len(rest) != 3 || !equalWord(rest[1], "each") || !equalWord(rest[2], "turn") {
		return TriggerFrequencyRestriction{}, false
	}
	var kind TriggerFrequencyKind
	switch {
	case equalWord(rest[0], "once"):
		kind = TriggerFrequencyOncePerTurn
	case equalWord(rest[0], "twice"):
		kind = TriggerFrequencyTwicePerTurn
	default:
		return TriggerFrequencyRestriction{}, false
	}
	return TriggerFrequencyRestriction{Kind: kind, Span: fullSpan}, true
}

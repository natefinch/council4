package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseBecomeColorEffect recognizes the one-shot continuous color-set effect
// "<subject> becomes <color>... until end of turn." (Cerulean Wisps, Niveous
// Wisps, Tidal Visionary, Raging Spirit; CR 613.1e). The subject is either the
// source ("This creature becomes colorless until end of turn.") or a single
// target ("Target permanent becomes white until end of turn."). The named color
// words SET the subject's color set; "colorless" clears it. The trailing "until
// end of turn" duration is required, distinguishing this temporary color change
// from a permanent one.
//
// Only the SET color form is recognized. The additive "becomes a <color>
// <type> in addition to its other colors and types" form (parsed by
// parseBecomeTypeEffect), the animation forms that name a base power/toughness,
// and the "the color of your choice" form (which needs a resolution-time color
// choice) all fail closed so those cards stay unsupported.
func parseBecomeColorEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	_ = atoms
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner, endOfTurn := becomeCopyTrimUntilEndOfTurn(body[:len(body)-1])
	if !endOfTurn {
		return nil, false
	}
	words := normalizedWords(inner)
	if len(words) < 3 {
		return nil, false
	}
	becomesIndex := -1
	for i, word := range words {
		if word == "becomes" {
			becomesIndex = i
			break
		}
	}
	if becomesIndex < 1 || becomesIndex+1 >= len(words) {
		return nil, false
	}
	source, ok := becomeColorSubject(words[:becomesIndex])
	if !ok {
		return nil, false
	}
	colors, colorless, ok := becomeColorColorRun(words[becomesIndex+1:])
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                      EffectBecomeColor,
		Context:                   EffectContextController,
		Span:                      sentence.Span,
		ClauseSpan:                sentence.Span,
		Text:                      sentence.Text,
		Tokens:                    append([]shared.Token(nil), body...),
		Duration:                  EffectDurationUntilEndOfTurn,
		BecomeColorColors:         colors,
		BecomeColorColorless:      colorless,
		BecomeColorSource:         source,
		BecomeColorUntilEndOfTurn: true,
	}
	return []EffectSyntax{effect}, true
}

// becomeColorSubject classifies the words before "becomes" as the source form
// ("this creature"/"this permanent", returning source=true) or a single target
// selector ("target ...", returning source=false). A target selector that
// embeds a connector or pump/grant verb marks a compound effect this recognizer
// cannot represent, so it fails closed. Any other subject fails closed.
func becomeColorSubject(words []string) (source bool, ok bool) {
	switch {
	case len(words) == 2 && words[0] == "this" &&
		(words[1] == "creature" || words[1] == "permanent"):
		return true, true
	case words[0] == "target":
		for _, word := range words[1:] {
			switch word {
			case "and", "then", "gets", "get", "gains", "gain", "loses", "lose":
				return false, false
			}
		}
		return false, true
	default:
		return false, false
	}
}

// becomeColorColorRun classifies the words after "becomes" as either the
// "colorless" clear form or a run of one or more named color words joined by
// "and". A leading "a"/"an" (the additive type forms), a leading "the" (the
// "color of your choice" form), or any trailing word fails closed.
func becomeColorColorRun(words []string) (colors []Color, colorless bool, ok bool) {
	if len(words) == 0 {
		return nil, false, false
	}
	if len(words) == 1 && words[0] == "colorless" {
		return nil, true, true
	}
	parsed := make([]Color, 0, len(words))
	index := 0
	for index < len(words) {
		runtimeColor, recognized := recognizeColorWord(words[index])
		if !recognized {
			return nil, false, false
		}
		parsed = append(parsed, runtimeColor)
		index++
		if index < len(words) && words[index] == "and" {
			index++
			if index >= len(words) {
				return nil, false, false
			}
		}
	}
	if len(parsed) == 0 {
		return nil, false, false
	}
	return parsed, false, true
}

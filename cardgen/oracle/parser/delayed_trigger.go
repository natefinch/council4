package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitDelayedTriggerEffects rewrites a single-sentence ability whose leading
// clause is a cast-event "this turn" delayed-trigger preamble ("Whenever you
// cast a spell this turn, ...", "When you next cast a creature spell this turn,
// ...") into one EffectDelayedTrigger effect carrying the sentence reparsed as a
// nested triggered ability with its turn window stripped. The preamble's cast
// verb otherwise reads as a spurious resolving cast effect that blocks lowering,
// and the post-comma body would read as an immediate effect rather than a
// delayed trigger. Rewriting fails closed: an ability the recognizer does not
// match, or whose stripped body does not reparse to exactly one triggered
// ability, is left untouched.
func emitDelayedTriggerEffects(abilities []Ability) {
	for i := range abilities {
		rewriteDelayedTriggerAbility(&abilities[i])
	}
}

func rewriteDelayedTriggerAbility(ability *Ability) {
	if len(ability.Sentences) != 1 {
		return
	}
	sentence := &ability.Sentences[0]
	tokens := semanticEffectTokens(sentence.Tokens)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	if comma <= 0 {
		return
	}
	lead := tokens[:comma]
	if !isDelayedThisTurnPreamble(lead) || !leadMentionsCast(lead) {
		return
	}
	body := delayedTriggerBodyEffect(sentence, tokens[comma].Span)
	if body == nil {
		return
	}
	inner, oneShot, ok := delayedTriggerInnerText(sentence.Text)
	if !ok {
		return
	}
	granted, ok := parseDelayedTriggerAbility(inner)
	if !ok {
		return
	}
	effect := EffectSyntax{
		Kind:                  EffectDelayedTrigger,
		Span:                  body.Span,
		VerbSpan:              body.VerbSpan,
		ClauseSpan:            body.ClauseSpan,
		Text:                  body.Text,
		DelayedTriggerAbility: &granted,
		DelayedTriggerOneShot: oneShot,
	}
	sentence.Effects = []EffectSyntax{effect}
	sentence.Targets = nil
	sentence.LegacyEffects = false
	ability.SemanticReferences = nil
	ability.SemanticKeywords = nil
	ability.ConditionBoundaries = nil
	ability.EventHistoryConditions = nil
	ability.ConditionClauses = nil
	ability.ConditionSegments = nil
}

func leadMentionsCast(lead []shared.Token) bool {
	for i := range lead {
		if equalWord(lead[i], "cast") {
			return true
		}
	}
	return false
}

// delayedTriggerBodyEffect returns the lone represented effect whose clause
// begins after the preamble comma at commaSpan, the post-comma body whose spans
// the rewritten EffectDelayedTrigger reuses so coverage credits the body clause.
// It returns nil when no single such effect exists so a body the parser split
// across multiple clauses fails closed.
func delayedTriggerBodyEffect(sentence *Sentence, commaSpan shared.Span) *EffectSyntax {
	var match *EffectSyntax
	for i := range sentence.Effects {
		effect := &sentence.Effects[i]
		if effect.Kind == EffectUnknown ||
			effect.ClauseSpan.Start.Offset < commaSpan.End.Offset {
			continue
		}
		if match != nil {
			return nil
		}
		match = effect
	}
	return match
}

// delayedTriggerInnerText reconstructs the nested triggered-ability source of a
// delayed "this turn" cast preamble by stripping the turn window and the "next"
// one-shot marker and normalizing the trigger introducer to "Whenever you cast"
// so the result is an ordinary triggered ability the pipeline parses ("Whenever
// you cast a spell this turn, <body>" -> "Whenever you cast a spell, <body>";
// "When you next cast a creature spell this turn, <body>" -> "Whenever you cast
// a creature spell, <body>"; "The next time you cast a creature spell this turn,
// <body>" -> "Whenever you cast a creature spell, <body>"). The delayed trigger
// reuses only the inner trigger pattern, so normalizing "When"/"the next time"
// to "Whenever" preserves the matched event while avoiding the provenance slot a
// one-shot "When you cast" trigger otherwise requires. oneShot reports the
// "next" forms that fire only on the first match. It fails closed on any other
// preamble shape.
func delayedTriggerInnerText(text string) (inner string, oneShot bool, ok bool) {
	trimmed := strings.TrimSpace(text)
	comma := strings.Index(trimmed, ",")
	if comma <= 0 {
		return "", false, false
	}
	preamble := strings.TrimSpace(trimmed[:comma])
	body := trimmed[comma:]
	lowered := strings.ToLower(preamble)
	if !strings.HasSuffix(lowered, "this turn") {
		return "", false, false
	}
	preamble = strings.TrimSpace(preamble[:len(preamble)-len("this turn")])
	lowered = strings.ToLower(preamble)
	switch {
	case strings.HasPrefix(lowered, "the next time you cast"):
		oneShot = true
		preamble = "Whenever you cast" + preamble[len("the next time you cast"):]
	case strings.HasPrefix(lowered, "when you next cast"):
		oneShot = true
		preamble = "Whenever you cast" + preamble[len("when you next cast"):]
	case strings.HasPrefix(lowered, "whenever you cast"):
	case strings.HasPrefix(lowered, "when you cast"):
		preamble = "Whenever you cast" + preamble[len("when you cast"):]
	default:
		return "", false, false
	}
	return strings.TrimSpace(preamble) + body, oneShot, true
}

// parseDelayedTriggerAbility reparses the reconstructed inner ability text
// through the same pipeline so downstream layers lower the delayed trigger from
// the typed inner document. It mirrors parseStaticGrantedAbility but takes raw
// text rather than a quoted token, and requires exactly one triggered ability so
// any other shape fails closed.
func parseDelayedTriggerAbility(text string) (StaticGrantedAbilitySyntax, bool) {
	document, diagnostics := Parse(text, Context{})
	if len(document.Abilities) != 1 ||
		document.Abilities[0].Kind != AbilityTriggered {
		return StaticGrantedAbilitySyntax{}, false
	}
	return StaticGrantedAbilitySyntax{
		Text:        text,
		document:    document,
		diagnostics: diagnostics,
	}, true
}

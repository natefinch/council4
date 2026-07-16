package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// exactEachOpponentGreatestPowerExileEffectSyntax recognizes the mandatory
// per-opponent choice whose independent candidate pool is that opponent's
// creatures tied for greatest power.
func exactEachOpponentGreatestPowerExileEffectSyntax(effect *EffectSyntax) bool {
	if effect.Context != EffectContextEachOpponent ||
		len(effect.Tokens) == 0 ||
		effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period ||
		!tokenWordsEqual(effect.Tokens[:len(effect.Tokens)-1],
			"Each", "opponent", "exiles", "a", "creature", "with", "the",
			"greatest", "power", "among", "creatures", "that", "player",
			"controls") {
		return false
	}
	effect.Selection = SelectionSyntax{
		Kind:             SelectionCreature,
		RequiredTypesAny: []CardType{CardTypeCreature},
	}
	effect.ExileEachOpponentChoosesGreatestPower = true
	return true
}

// exactEachOpponentCorrelatedExiledPowerDamageEffectSyntax recognizes a
// source-dealt group damage clause whose amount is different for each opponent:
// the power of the creature that same opponent exiled.
func exactEachOpponentCorrelatedExiledPowerDamageEffectSyntax(effect *EffectSyntax) bool {
	verb := -1
	for i, token := range effect.Tokens {
		if token.Span == effect.VerbSpan {
			verb = i
			break
		}
	}
	if verb < 0 ||
		effect.Tokens[len(effect.Tokens)-1].Kind != shared.Period ||
		!tokenWordsEqual(effect.Tokens[verb:len(effect.Tokens)-1],
			"deals", "damage", "to", "each", "opponent", "equal", "to", "the",
			"power", "of", "the", "creature", "they", "exiled") {
		return false
	}
	subject := shared.SpanOf(effect.Tokens[:verb])
	hasSource := false
	for _, reference := range effect.SubjectReferences {
		if spanCovers(subject, reference.Span) &&
			(reference.Kind == ReferenceSelfName || reference.Kind == ReferenceThisObject) {
			hasSource = true
			break
		}
	}
	if !hasSource {
		return false
	}
	effect.DamageEachOpponentCorrelatedExiledPower = true
	return true
}

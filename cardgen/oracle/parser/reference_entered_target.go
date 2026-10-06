package parser

import "github.com/natefinch/council4/mtg/game/zone"

// The action's own predicate names its input; a later subject names its output.
func bindEnteredTargetReferences(effects []*EffectSyntax, references []Reference, conditions []ConditionSegment) {
	for i := range references {
		reference := &references[i]
		if reference.ProducerClauseID > 0 || !libraryCardReference(*reference) ||
			reference.SubjectNoun == ObjectNounSpell || reference.SubjectNoun == ObjectNounPlayer {
			continue
		}
		var latest *EffectSyntax
		for _, effect := range effects {
			if effect.VerbSpan.End.Offset >= reference.Span.Start.Offset ||
				parserSpanContains(effect.ClauseSpan, reference.Span) ||
				libraryReferenceIsOwnActionCondition(*reference, effect.ClauseID, conditions) {
				continue
			}
			if len(effect.Targets) > 0 || enteredSubjectBarrier(effect.Kind) {
				latest = effect
			}
		}
		if latest == nil || !latest.Exact || latest.Negated || latest.DelayedTiming != DelayedTimingNone ||
			(latest.Kind != EffectPut && latest.Kind != EffectReturn) || latest.ToZone != zone.Battlefield ||
			len(latest.Targets) != 1 {
			continue
		}
		target := latest.Targets[0]
		if target.Selection.Kind == SelectionPlayer || target.Selection.Kind == SelectionOpponent ||
			target.Selection.Zone == zone.None ||
			target.Selection.Zone == zone.Battlefield || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			continue
		}
		reference.ProducerClauseID = latest.ClauseID
	}
}

func enteredSubjectBarrier(kind EffectKind) bool {
	switch kind {
	case EffectPut, EffectReturn, EffectExile, EffectReveal, EffectLookAtLibraryTop,
		EffectDig, EffectSearch, EffectChoosePermanent, EffectCreate,
		EffectManifestDread, EffectSacrifice, EffectDestroy:
		return true
	default:
		return false
	}
}

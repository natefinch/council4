package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TargetedPlayerActor validates the whole body's sole player-target declaration
// and the clause that uses it. An unindexed actor cannot name a partial set.
func (content AbilityContent) TargetedPlayerActor(effect CompiledEffect) (CompiledTarget, bool) {
	if content.subjectScope == nil || content.subjectScope.owner == nil ||
		effect.subjectScope != content.subjectScope || effect.ClauseID <= 0 ||
		effect.Order == (shared.SourceOrder{}) || effect.Negated {
		return CompiledTarget{}, false
	}
	owner := content.subjectScope.owner
	if len(content.Targets) != len(owner.Targets) {
		return CompiledTarget{}, false
	}
	playerIndex := -1
	for i, target := range content.Targets {
		canonical := owner.Targets[i]
		if target.subjectScope != content.subjectScope || !target.Exact ||
			target.Order != canonical.Order || target.Order == (shared.SourceOrder{}) ||
			target.Selector.Kind != canonical.Selector.Kind ||
			target.Selector.Zone != canonical.Selector.Zone ||
			target.Cardinality != canonical.Cardinality {
			return CompiledTarget{}, false
		}
		if target.Selector.Kind == SelectorPlayer || target.Selector.Kind == SelectorOpponent {
			if playerIndex >= 0 {
				return CompiledTarget{}, false
			}
			playerIndex = i
		}
	}
	if playerIndex < 0 {
		return CompiledTarget{}, false
	}
	target := owner.Targets[playerIndex]
	if target.Cardinality.Max <= 1 || target.Cardinality.Min < 0 ||
		target.Cardinality.CountEqualsX || target.Cardinality.MaxX {
		return CompiledTarget{}, false
	}
	var original *CompiledEffect
	for i := range owner.Effects {
		candidate := &owner.Effects[i]
		if candidate.ClauseID == effect.ClauseID {
			if original != nil {
				return CompiledTarget{}, false
			}
			original = candidate
		}
	}
	if original == nil || original.Kind != effect.Kind || original.Order != effect.Order {
		return CompiledTarget{}, false
	}
	if len(original.Targets) == 1 {
		declared := original.Targets[0]
		if len(effect.Targets) != 1 || declared.subjectScope != content.subjectScope ||
			declared.Order != target.Order || effect.Targets[0].Order != target.Order ||
			effect.Targets[0].subjectScope != content.subjectScope ||
			effect.Targets[0].Cardinality != target.Cardinality ||
			effect.Targets[0].Selector.Kind != target.Selector.Kind {
			return CompiledTarget{}, false
		}
		return target, true
	}
	if len(original.Targets) != 0 || len(effect.Targets) != 0 ||
		len(original.References) != 1 || len(effect.References) != 1 {
		return CompiledTarget{}, false
	}
	reference := original.References[0]
	local := effect.References[0]
	if reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounThose ||
		!content.OwnsSubject(reference) || !content.OwnsSubject(local) ||
		!reference.Subject.targetedPlayerActor || reference.Subject.targetOrder != target.Order ||
		reference.Subject != local.Subject || reference.Binding != local.Binding ||
		reference.NodeID != local.NodeID || reference.Order != local.Order ||
		reference.Kind != local.Kind || reference.Pronoun != local.Pronoun ||
		target.Order.End > reference.Order.Start || !original.Order.Contains(reference.Order) {
		return CompiledTarget{}, false
	}
	matches := 0
	for _, canonical := range owner.References {
		if canonical.NodeID == reference.NodeID && canonical.Order == reference.Order &&
			canonical.Kind == reference.Kind && canonical.Pronoun == reference.Pronoun {
			matches++
		}
	}
	return target, matches == 1
}

func issueTargetedPlayerActorSubjects(content *AbilityContent) {
	var producer *CompiledEffect
	var declaration CompiledTarget
	for i := range content.Effects {
		effect := &content.Effects[i]
		if len(effect.Targets) == 0 || !targetedPlayerActorEffect(effect.Kind) {
			continue
		}
		target, ok := content.TargetedPlayerActor(*effect)
		if !ok {
			continue
		}
		if producer != nil {
			return
		}
		producer, declaration = effect, target
	}
	if producer == nil {
		return
	}
	for i := range content.References {
		reference := &content.References[i]
		if reference.Binding != ReferenceBindingAmbiguous ||
			reference.Kind != ReferencePronoun || reference.Pronoun != ReferencePronounThose ||
			reference.CardIdentity ||
			reference.SubjectNoun != parser.ObjectNounUnknown && reference.SubjectNoun != parser.ObjectNounPlayer ||
			reference.Order == (shared.SourceOrder{}) || declaration.Order.End > reference.Order.Start {
			continue
		}
		owners := 0
		for _, effect := range content.Effects {
			if effect.subjectScope == content.subjectScope && effect.ClauseID > 0 &&
				!effect.Negated && targetedPlayerActorEffect(effect.Kind) &&
				len(effect.Targets) == 0 && effect.Order.Contains(reference.Order) &&
				producer.Order.End <= effect.Order.Start && effectOwnsReference(effect, *reference) {
				owners++
			}
		}
		nodes := 0
		for _, candidate := range content.References {
			if candidate.NodeID == reference.NodeID {
				nodes++
			}
		}
		if owners != 1 || nodes != 1 {
			continue
		}
		reference.Subject = ReferenceSubjectProof{
			scope: content.subjectScope, domain: ReferenceSubjectPlayerGroup, lifetime: ReferenceLifetimePlayer,
			binding: reference.Binding, kind: reference.Kind, pronoun: reference.Pronoun,
			nodeID: reference.NodeID, noun: reference.SubjectNoun,
			targetOrder: declaration.Order, targetedPlayerActor: true,
		}
	}
}

func targetedPlayerActorEffect(kind EffectKind) bool {
	switch kind {
	case EffectDraw, EffectDiscard, EffectMill, EffectGain, EffectLose, EffectSacrifice:
		return true
	default:
		return false
	}
}

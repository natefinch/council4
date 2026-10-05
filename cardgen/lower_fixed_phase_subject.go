package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type capturedContentSubject struct {
	references []int
	group      bool
	card       bool
	fromZone   zone.Type
	direct     bool
}

func lowerContentObjectReference(ctx contentCtx, reference compiler.CompiledReference, bindings referenceLoweringContext) (game.ObjectReference, bool) {
	if ctx.capturedSubject != nil {
		if ctx.capturedSubject.group || ctx.capturedSubject.card || !slices.Contains(ctx.capturedSubject.references, reference.NodeID) {
			return game.ObjectReference{}, false
		}
		return game.CapturedObjectReference(), true
	}
	return lowerObjectReference(reference, bindings)
}

// lowerFixedPhaseSubject schedules an already-modeled body against its exact
// grammatical subject. Target and event identities need no publishing action;
// products use the shared actual-result publisher selected by ClauseID.
func lowerFixedPhaseSubject(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
	sequence []game.Instruction,
	ranges [][2]int,
	effects []compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic, bool) {
	effect := ctx.content.Effects[0]
	if effect.DelayedTiming == 0 || effect.DelayedSubject.Kind == parser.DelayedSubjectUnknown {
		return game.AbilityContent{}, nil, false
	}
	switch effect.Kind {
	case compiler.EffectSacrifice, compiler.EffectDestroy, compiler.EffectExile,
		compiler.EffectReturn, compiler.EffectPut, compiler.EffectTransform,
		compiler.EffectTap, compiler.EffectUntap, compiler.EffectRemoveCounter:
	default:
		return game.AbilityContent{}, nil, false
	}
	refuse := func(detail string) (game.AbilityContent, *shared.Diagnostic, bool) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported delayed object capture", detail), true
	}
	subject := effect.DelayedSubject
	if subject.Kind == parser.DelayedSubjectUnsupported {
		return refuse("fixed-phase body has no unique compatible typed subject")
	}
	var object game.ObjectReference
	group := false
	card := false
	fromZone := zone.None
	switch subject.Kind {
	case parser.DelayedSubjectProduct:
		producer := slices.IndexFunc(effects, func(prior compiler.CompiledEffect) bool {
			return prior.ClauseID == subject.ProducerClauseID
		})
		if producer < 0 || producer >= len(ranges) {
			return refuse("fixed-phase subject producer is unavailable")
		}
		reference := compiler.CompiledReference{
			Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: producer,
		}
		_, key, ok := sequencePriorInstructionPublication([]compiler.CompiledReference{reference}, sequence, ranges, true)
		if !ok {
			return refuse("fixed-phase subject producer has no modeled actual-result publication")
		}
		object = game.LinkedObjectReference(string(key))
		// A CreateToken's requested count does not bound its actual batch:
		// replacement effects may multiply even a singular printed token.
		group = effects[producer].Kind == compiler.EffectCreate
		if effects[producer].Kind == compiler.EffectExile {
			card = true
			fromZone = zone.Exile
		}
		ctx.priorInstruction, ctx.priorLinkedKey = producer, key
	case parser.DelayedSubjectTarget:
		index := 0
		if !subject.DirectTarget {
			referenceIndex := slices.IndexFunc(ctx.content.References, func(reference compiler.CompiledReference) bool {
				return slices.Contains(subject.ReferenceNodeIDs, reference.NodeID)
			})
			if referenceIndex < 0 {
				return refuse("fixed-phase target subject reference is unavailable")
			}
			reference := ctx.content.References[referenceIndex]
			if reference.Binding != compiler.ReferenceBindingTarget {
				return refuse("fixed-phase target subject has no exact local occurrence")
			}
			index = reference.Occurrence
		}
		if index < 0 || index >= len(ctx.content.Targets) {
			return refuse("fixed-phase target subject has no exact local occurrence")
		}
		target := ctx.content.Targets[index]
		_, permanent := permanentTargetSpecWithCardinality(target)
		if target.Cardinality.Min < 0 || target.Cardinality.Min > 1 || target.Cardinality.Max != 1 ||
			!permanent || (target.Selector.Zone != zone.None && target.Selector.Zone != zone.Battlefield) {
			return refuse("fixed-phase subject is not one battlefield target")
		}
		object = game.TargetPermanentReference(index)
	case parser.DelayedSubjectSource:
		if ctx.enclosingKind == compiler.AbilitySpell && subject.CardZone == zone.None {
			return refuse("a resolving spell is not a battlefield source subject")
		}
		object = game.SourcePermanentReference()
		if subject.CardZone != zone.None {
			card, fromZone = true, subject.CardZone
		}
	case parser.DelayedSubjectEvent:
		object = game.EventPermanentReference()
		if subject.CardZone != zone.None {
			card, fromZone = true, subject.CardZone
		}
	case parser.DelayedSubjectEventRelated:
		object = game.EventRelatedPermanentReference()
	default:
		return refuse("fixed-phase subject domain is not modeled")
	}
	if len(subject.ReferenceNodeIDs) == 0 && !effect.CreatedTokensReference && !subject.DirectTarget {
		return game.AbilityContent{}, nil, false
	}
	// Delayed quantities may not read the creating stack object's X, event or
	// prior scalar result at a later phase without a modeled quantity capture.
	if effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.PowerDelta.VariableX || effect.ToughnessDelta.VariableX {
		return refuse("fixed-phase body requires an unavailable delayed quantity capture")
	}
	timing := effect.DelayedTiming
	delayedConditions := ctx.content.Conditions
	ctx.content.Conditions = nil
	ctx.content.Effects = slices.Clone(ctx.content.Effects)
	ctx.content.Effects[0].DelayedTiming = 0
	ctx.content.Targets = nil
	ctx.capturedSubject = &capturedContentSubject{
		references: subject.ReferenceNodeIDs,
		group:      group,
		card:       card,
		fromZone:   fromZone,
		direct:     subject.DirectTarget,
	}
	ctx.allowEventPronoun = true
	content, diagnostic := lowerImmediateSingleEffectSpell(cardName, ctx, syntax)
	if diagnostic != nil {
		return refuse("the captured subject's immediate body primitive or parameter is not modeled: " + diagnostic.Summary)
	}
	if len(content.SharedTargets) != 0 || content.IsModal() || len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 || len(content.Modes[0].Sequence) == 0 {
		return refuse("fixed-phase body requires unresolved future targets or modes")
	}
	if len(delayedConditions) > 1 {
		return refuse("multiple future delayed-body conditions are not modeled")
	}
	for _, condition := range delayedConditions {
		lowered, ok := lowerCapturedDelayedCondition(condition, subject, ctx)
		if !ok || group || card {
			return refuse("future delayed-body condition or its subject is not modeled")
		}
		for i := range content.Modes[0].Sequence {
			content.Modes[0].Sequence[i].Condition = opt.Val(game.EffectCondition{Condition: opt.Val(lowered)})
		}
	}
	trigger := game.DelayedTriggerDef{Timing: timing, Content: content}
	if card {
		trigger.CapturedCard = opt.Val(object)
	} else if group {
		trigger.CapturedObjectGroup = opt.Val(object)
	} else {
		trigger.CapturedObject = opt.Val(object)
	}

	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.CreateDelayedTrigger{Trigger: trigger},
	}}}.Ability(), nil, true
}

func lowerCapturedDelayedCondition(condition compiler.CompiledCondition, subject parser.DelayedSubjectOwnership, ctx contentCtx) (game.Condition, bool) {
	if !condition.Ownership.DelayedBody || condition.Intervening || condition.Resolving {
		return game.Condition{}, false
	}
	lowered, ok := lowerConditionWithReferences(condition, conditionContextEffectGate, referenceLoweringContext{
		AllowSource: true, AllowTarget: true, AllowEvent: true,
		PriorInstruction: ctx.priorInstruction, PriorLinkedKey: ctx.priorLinkedKey,
	})
	if !ok {
		return game.Condition{}, false
	}
	if lowered.Object.Exists {
		reference := condition.ObjectReference
		switch subject.Kind {
		case parser.DelayedSubjectTarget:
			ok = condition.ObjectBinding == compiler.ReferenceBindingTarget &&
				lowered.Object.Val.Kind() == game.ObjectReferenceTargetPermanent &&
				lowered.Object.Val.TargetIndex() == subject.TargetOccurrence
		case parser.DelayedSubjectSource:
			ok = condition.ObjectBinding == compiler.ReferenceBindingSource
		case parser.DelayedSubjectEvent:
			ok = condition.ObjectBinding == compiler.ReferenceBindingEventPermanent
		case parser.DelayedSubjectEventRelated:
			ok = condition.ObjectBinding == compiler.ReferenceBindingEventRelatedPermanent
		case parser.DelayedSubjectProduct:
			ok = reference != nil && reference.Binding == compiler.ReferenceBindingPriorInstructionResult &&
				reference.PriorInstruction == ctx.priorInstruction
		default:
			ok = false
		}
		if !ok {
			return game.Condition{}, false
		}
		lowered.Object = opt.Val(game.CapturedObjectReference())
		return lowered, true
	}
	switch condition.Predicate {
	case compiler.ConditionPredicateControllerLifeAtLeast,
		compiler.ConditionPredicateControllerLifeAtMost,
		compiler.ConditionPredicateControllerHandSizeExactly,
		compiler.ConditionPredicateControllerHandSizeAtLeast,
		compiler.ConditionPredicateControllerControls:
		return lowered, true
	default:
		return game.Condition{}, false
	}
}

package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

const searchUnitCategory = "structural — search group clauses, guards, or references lack exact typed ownership"

// The original clauses remain in the plan. Only their physical instruction
// ranges coincide; references and condition scopes keep their typed identities.
type searchSequenceUnit struct {
	end     int
	mode    game.Mode
	targets []compiler.CompiledTarget
}

func planSearchSequenceUnits(content compiler.AbilityContent) (map[int]searchSequenceUnit, string) {
	units := make(map[int]searchSequenceUnit)
	clauseIDs := make(map[int]bool)
	for _, effect := range content.Effects {
		if effect.ClauseID <= 0 || clauseIDs[effect.ClauseID] {
			if slices.ContainsFunc(content.Effects, func(e compiler.CompiledEffect) bool { return e.Kind == compiler.EffectSearch }) {
				return nil, "structural — search unit has a missing or duplicate original ClauseID"
			}
			return units, ""
		}
		clauseIDs[effect.ClauseID] = true
	}
	for start := 0; start < len(content.Effects); start++ {
		if content.Effects[start].Kind != compiler.EffectSearch {
			continue
		}
		group, ok := searchGroupSpec(content.Effects[start:])
		if !ok {
			return nil, "search unit — unsupported contiguous search/filter/destination parameters"
		}
		if group.RiderIndex != 0 {
			return nil, "search unit — an internal rider cannot be compressed with the search"
		}
		if group.Spec.Destination != zone.Hand && group.Spec.Destination != zone.Battlefield {
			return nil, "search unit — only hand or battlefield destinations compose"
		}
		if !group.Quantity.IsDynamic() && group.Quantity.Value() < 1 && !group.Spec.AnyNumber {
			return nil, "search unit — a fixed search quantity must be positive"
		}
		if group.Quantity.IsDynamic() && (group.Spec.AlsoGraveyard || group.Spec.RevealOnly ||
			group.Spec.ExileFaceDown || group.Spec.SplitDestination.Exists || len(group.Spec.SlotFilters) != 0) {
			return nil, "search unit — this specialized destination has no dynamic search-limit contract"
		}
		if searchSpecHasTargetBound(group.Spec) {
			return nil, "search unit — target-dependent mana-value bounds have no resolving stack context"
		}
		unit, ok := ownedSearchSequenceUnit(content, start, group)
		if !ok {
			return nil, searchUnitCategory
		}
		units[start] = unit
		start = unit.end - 1
	}
	for index, effect := range content.Effects {
		if _, start := units[index]; start || searchUnitAbsorbed(units, index) {
			continue
		}
		for _, reference := range effect.References {
			if reference.Binding == compiler.ReferenceBindingPriorInstructionResult &&
				searchUnitAbsorbed(units, reference.PriorInstruction) {
				return nil, searchUnitCategory
			}
		}
	}
	for _, condition := range content.Conditions {
		if condition.ObjectReference != nil &&
			condition.ObjectBinding == compiler.ReferenceBindingPriorInstructionResult &&
			searchUnitAbsorbed(units, condition.ObjectReference.PriorInstruction) {
			return nil, searchUnitCategory
		}
	}
	return units, ""
}

func ownedSearchSequenceUnit(content compiler.AbilityContent, start int, group searchGroup) (searchSequenceUnit, bool) {
	end := start + group.Length
	effects := content.Effects[start:end]
	var targets []compiler.CompiledTarget
	var references []compiler.CompiledReference
	for offset, effect := range effects {
		if effect.VerbOrder.End <= effect.VerbOrder.Start ||
			offset > 0 && effects[offset-1].VerbOrder.End > effect.VerbOrder.Start {
			return searchSequenceUnit{}, false
		}
		if effect.Optional || effect.Negated || effect.DelayedTiming != 0 || effect.Duration != compiler.DurationNone ||
			effect.UnsupportedDetail != "" || len(effect.OptionalActionClauseIDs) != 0 ||
			effect.OptionalPostfixElse.ActionClauseID != 0 ||
			offset > 0 && (effect.Replacement.Kind != parser.EffectReplacementNone ||
				effect.Connection == parser.EffectConnectionOtherwise) {
			return searchSequenceUnit{}, false
		}
		targets = append(targets, effect.Targets...)
		ownedReferences := referencesOutsideOwnedConditions(effect.References, content.Conditions)
		if offset == 0 && group.Spec.NameFromChoice != "" && len(ownedReferences) != 1 {
			return searchSequenceUnit{}, false
		}
		if (effect.Kind == compiler.EffectPut || effect.Kind == compiler.EffectReveal) && len(ownedReferences) != 1 {
			return searchSequenceUnit{}, false
		}
		if offset > 0 && effect.Context != effects[0].Context && effect.Context != parser.EffectContextPriorSubject {
			return searchSequenceUnit{}, false
		}
		if effect.Kind == compiler.EffectShuffle && (effect.FromZone != zone.None || effect.ToZone != zone.None) {
			return searchSequenceUnit{}, false
		}
		for _, reference := range ownedReferences {
			if effect.Kind == compiler.EffectSearch && effect.Exact &&
				effect.Context == parser.EffectContextEachPlayer &&
				reference.Kind == compiler.ReferencePronoun && reference.Pronoun == compiler.ReferencePronounTheir &&
				reference.ProducerClauseID == 0 && canonicalSearchReference(content, reference) {
				continue
			}
			if !content.OwnsSubject(reference) || !canonicalSearchReference(content, reference) {
				return searchSequenceUnit{}, false
			}
			if offset == 0 && searchNameChoiceOperand(content, start, group.Spec, reference) {
				continue
			}
			if reference.Binding == compiler.ReferenceBindingPriorInstructionResult {
				producer := reference.PriorInstruction
				if producer < start || producer >= start+offset ||
					!reference.SubjectProducerMatches(content.Effects[producer]) ||
					content.Effects[producer].Kind != compiler.EffectSearch &&
						content.Effects[producer].Kind != compiler.EffectReveal {
					return searchSequenceUnit{}, false
				}
			} else if reference.Binding != compiler.ReferenceBindingTarget || !isPlayerPronoun(reference.Pronoun) {
				return searchSequenceUnit{}, false
			}
			references = append(references, reference)
		}
	}
	// An internal branch cannot be erased by the aggregate. A modeled gate
	// governing the whole group is applied once to the physical Search.
	for _, condition := range content.Conditions {
		owners, reason := conditionClauseIndices(content.Effects, condition)
		if reason != "" {
			return searchSequenceUnit{}, false
		}
		intersects := false
		for _, owner := range owners {
			intersects = intersects || owner >= start && owner < end
		}
		if intersects && (!slices.Contains(owners, start) || !slices.Contains(owners, end-1)) {
			return searchSequenceUnit{}, false
		}
	}
	localReferences, ok := localizeTargetReferences(references, content.Targets, targets)
	if !ok {
		return searchSequenceUnit{}, false
	}
	ctx := contentCtx{content: compiler.AbilityContent{Effects: effects, Targets: targets, References: localReferences}}
	subject, ok := searchSearcher(ctx, &effects[0])
	if !ok {
		return searchSequenceUnit{}, false
	}
	for _, reference := range localReferences {
		if reference.Binding == compiler.ReferenceBindingTarget && len(subject.Targets) == 0 &&
			subject.Group.Kind == game.PlayerGroupReferenceNone {
			return searchSequenceUnit{}, false
		}
	}
	controller, controllerTargets, ok := searchController(effects[0].SearchControl, subject.Targets, group.Spec)
	if !ok || subject.Group.Kind != game.PlayerGroupReferenceNone && controller.Exists ||
		len(subject.Targets) != len(targets) {
		return searchSequenceUnit{}, false
	}
	sequence, ok := searchGroupInstructionsWithSearcher(group, subject.Player, subject.Group)
	if !ok || len(sequence) == 0 {
		return searchSequenceUnit{}, false
	}
	search, ok := sequence[0].Primitive.(game.Search)
	if !ok {
		return searchSequenceUnit{}, false
	}
	search.Controller = controller
	sequence[0].Primitive = search
	return searchSequenceUnit{
		end: end, targets: targets,
		mode: game.Mode{Targets: append(subject.Targets, controllerTargets...), Sequence: sequence},
	}, true
}

func searchUnitAbsorbed(units map[int]searchSequenceUnit, index int) bool {
	for start, unit := range units {
		if index > start && index < unit.end {
			return true
		}
	}
	return false
}

func referencesForEffects(effects []compiler.CompiledEffect) []compiler.CompiledReference {
	var references []compiler.CompiledReference
	for i := range effects {
		references = append(references, effects[i].References...)
	}
	return references
}

func canonicalSearchReference(content compiler.AbilityContent, reference compiler.CompiledReference) bool {
	matches := 0
	for _, canonical := range content.References {
		if canonical.NodeID == reference.NodeID {
			if canonical.Order != reference.Order || canonical.Order.End <= canonical.Order.Start ||
				canonical.Binding != reference.Binding || canonical.ProducerClauseID != reference.ProducerClauseID ||
				canonical.Occurrence != reference.Occurrence {
				return false
			}
			matches++
		}
	}
	return matches == 1
}

func searchNameChoiceOperand(content compiler.AbilityContent, start int, spec game.SearchSpec, reference compiler.CompiledReference) bool {
	search := content.Effects[start]
	prior := reference.PriorInstruction
	if !content.OwnsSubject(reference) || !canonicalSearchReference(content, reference) ||
		!search.SearchSameNameAsChosenObject || !search.SearchNameFromChosenPermanent ||
		spec.NameFromChoice != game.ResolutionChosenPermanentChoiceKey ||
		reference.Kind != compiler.ReferenceThatObject ||
		reference.Binding != compiler.ReferenceBindingPriorInstructionResult ||
		reference.SubjectDomain() != compiler.ReferenceSubjectPermanent ||
		reference.SubjectLifetime() != compiler.ReferenceLifetimeActualProduct ||
		prior < 0 || prior >= start {
		return false
	}
	choice := content.Effects[prior]
	if !reference.SubjectProducerMatches(choice) || choice.ClauseID <= 0 ||
		choice.VerbOrder.End <= choice.VerbOrder.Start ||
		choice.VerbOrder.End > search.VerbOrder.Start {
		return false
	}
	if slices.ContainsFunc(content.Effects[:prior], func(effect compiler.CompiledEffect) bool {
		return effect.Kind == compiler.EffectChoosePermanent
	}) || slices.ContainsFunc(content.Effects[prior+1:start], func(effect compiler.CompiledEffect) bool {
		return effect.Kind == compiler.EffectChoosePermanent
	}) {
		return false
	}
	for _, condition := range content.Conditions {
		owners, reason := conditionClauseIndices(content.Effects, condition)
		if reason != "" || slices.Contains(owners, prior) && !slices.Contains(owners, start) {
			return false
		}
	}
	choiceCtx := contentCtx{content: compiler.AbilityContent{
		Effects: []compiler.CompiledEffect{choice}, Targets: choice.Targets, References: choice.References,
	}}
	_, diagnostic := lowerChoosePermanentSpell(choiceCtx)
	return diagnostic == nil
}

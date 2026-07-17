package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// emitSourceOrder assigns every ability (and its modes) a dense source-order
// ranking and stamps it onto the nodes whose relative order or structural
// nesting the compiler consumes. The parser owns all positional reasoning here;
// the compiler later compares the emitted ranks mechanically and never inspects
// raw byte offsets to derive ordering or containment.
//
// All of an ability's ranked boundaries (start and end of every participating
// node, including its modes) are ranked together. Because a dense rank is
// strictly monotonic in the underlying offset, every pairwise order comparison
// and every span-containment test is reproduced exactly in rank space, while
// absolute positions are discarded.
func emitSourceOrder(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		var offsets []int
		visitAbilityOrderNodes(ability, func(span shared.Span, _ *shared.SourceOrder) {
			offsets = append(offsets, span.Start.Offset, span.End.Offset)
		})
		ranks := denseRanks(offsets)
		visitAbilityOrderNodes(ability, func(span shared.Span, order *shared.SourceOrder) {
			order.Start = ranks[span.Start.Offset]
			order.End = ranks[span.End.Offset]
		})
	}
}

// denseRanks maps each distinct offset to its position in ascending sorted
// order, producing a gap-free monotonic ranking.
func denseRanks(offsets []int) map[int]int {
	slices.Sort(offsets)
	ranks := make(map[int]int, len(offsets))
	next := 0
	for _, offset := range offsets {
		if _, ok := ranks[offset]; ok {
			continue
		}
		ranks[offset] = next
		next++
	}
	return ranks
}

// visitAbilityOrderNodes invokes visit for every span/order-field pair whose
// source order or containment the compiler consumes, across the ability and its
// modes. The single traversal is used for both gathering offsets and stamping
// ranks so the two phases can never diverge.
func visitAbilityOrderNodes(ability *Ability, visit func(shared.Span, *shared.SourceOrder)) {
	if ability.Trigger != nil {
		visit(ability.Trigger.Span, &ability.Trigger.Order)
	}
	if ability.CostSyntax != nil {
		visit(ability.CostSyntax.Span, &ability.CostSyntax.Order)
		for i := range ability.CostSyntax.Components {
			component := &ability.CostSyntax.Components[i]
			visit(component.Span, &component.Order)
		}
	}
	visitSentenceOrderNodes(ability.Sentences, visit)
	for i := range ability.SemanticReferences {
		reference := &ability.SemanticReferences[i]
		visit(reference.Span, &reference.Order)
	}
	for i := range ability.ConditionSegments {
		segment := &ability.ConditionSegments[i]
		visit(segment.Span, &segment.Order)
	}
	for i := range ability.TriggerConditionSegments {
		segment := &ability.TriggerConditionSegments[i]
		visit(segment.Span, &segment.Order)
	}
	if ability.Modal != nil {
		for i := range ability.Modal.Options {
			visitModeOrderNodes(&ability.Modal.Options[i], visit)
		}
	}
}

func visitModeOrderNodes(mode *Mode, visit func(shared.Span, *shared.SourceOrder)) {
	visitSentenceOrderNodes(mode.Sentences, visit)
	for i := range mode.SemanticReferences {
		reference := &mode.SemanticReferences[i]
		visit(reference.Span, &reference.Order)
	}
	for i := range mode.ConditionSegments {
		segment := &mode.ConditionSegments[i]
		visit(segment.Span, &segment.Order)
	}
}

func visitSentenceOrderNodes(sentences []Sentence, visit func(shared.Span, *shared.SourceOrder)) {
	for si := range sentences {
		sentence := &sentences[si]
		if sentence.StaticRule != nil {
			rule := sentence.StaticRule
			visit(rule.Span, &rule.Order)
			visit(rule.Operation.Span, &rule.Operation.Order)
			visit(rule.Subject.Span, &rule.Subject.Order)
		}
		for ti := range sentence.Targets {
			target := &sentence.Targets[ti]
			visit(target.Span, &target.Order)
		}
		for ei := range sentence.Effects {
			effect := &sentence.Effects[ei]
			visitEffectOrderNodes(effect, visit)
		}
	}
}

func visitEffectOrderNodes(effect *EffectSyntax, visit func(shared.Span, *shared.SourceOrder)) {
	visit(effect.Span, &effect.Order)
	visit(effect.VerbSpan, &effect.VerbOrder)
	visit(effect.Payment.Span, &effect.Payment.Order)
	for i := range effect.RepeatBody {
		visitEffectOrderNodes(&effect.RepeatBody[i], visit)
	}
}

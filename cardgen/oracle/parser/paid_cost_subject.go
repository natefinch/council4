package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// PaidCostDomain is the grammatical domain of an actual paid subject.
type PaidCostDomain uint8

// Paid cost domains recognized by the parser.
const (
	PaidCostDomainUnknown PaidCostDomain = iota
	PaidCostDomainSacrificedPermanent
	PaidCostDomainDiscardedCard
)

// PaidCostProducer identifies a cost clause and one component within it.
type PaidCostProducer struct {
	ClauseID        int
	ComponentNodeID int
}

// PaidCostBinding connects an exact condition reference to an actual cost
// component. Unbound and ambiguous grammar retains Known=false.
type PaidCostBinding struct {
	Producer       PaidCostProducer
	ConsumerNodeID int
	Domain         PaidCostDomain
	Known          bool
}

func paidCostSubjectAt(tokens []shared.Token, start int) (int, PaidCostDomain, bool) {
	if start+2 >= len(tokens) || !equalWord(tokens[start], "the") {
		return 0, PaidCostDomainUnknown, false
	}
	noun := strings.TrimSuffix(strings.ToLower(tokens[start+2].Text), "'s")
	switch {
	case equalWord(tokens[start+1], "discarded") && noun == "card":
		return 3, PaidCostDomainDiscardedCard, true
	case equalWord(tokens[start+1], "sacrificed") &&
		(noun == "permanent" || conditionAttributeComparePermanentNoun(noun)):
		return 3, PaidCostDomainSacrificedPermanent, true
	default:
		return 0, PaidCostDomainUnknown, false
	}
}

func recognizePaidCostSubjectCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	width, domain, ok := paidCostSubjectAt(body, 0)
	if !ok || width >= len(body) || (!equalWord(body[width], "was") && !equalWord(body[width], "wasn't")) {
		return ConditionClause{}, false
	}
	rest := body[width+1:]
	negated := equalWord(body[width], "wasn't")
	if len(rest) > 0 && equalWord(rest[0], "not") {
		negated = true
		rest = rest[1:]
	}
	if len(rest) > 0 && (equalWord(rest[0], "a") || equalWord(rest[0], "an")) {
		rest = rest[1:]
	}
	selection, ok := parseConditionSelection(rest, atoms)
	if !ok && len(rest) > 1 && equalWord(rest[len(rest)-1], "card") {
		selection, ok = parseConditionSelection(rest[:len(rest)-1], atoms)
	}
	if !ok && len(rest) == 1 {
		if value, known := atoms.ColorAt(rest[0].Span); known {
			selection = ConditionSelection{ColorsAny: []TriggerColor{triggerColorFromAtom(value)}}
			ok = true
		} else if value, known := conditionSupertypeAtom(rest[0].Span, atoms); known {
			selection = ConditionSelection{Supertypes: []ConditionSupertype{value}}
			ok = true
		} else if equalWord(rest[0], "multicolored") {
			selection = ConditionSelection{Multicolored: true}
			ok = true
		} else if equalWord(rest[0], "colorless") {
			selection = ConditionSelection{Colorless: true}
			ok = true
		}
	}
	if !ok {
		return ConditionClause{}, false
	}
	var subjectTypes []TriggerCardType
	if domain == PaidCostDomainSacrificedPermanent && !equalWord(body[2], "permanent") {
		noun, known := parseConditionSelection(body[2:3], atoms)
		if !known {
			return ConditionClause{}, false
		}
		subjectTypes = noun.RequiredTypes
	}
	span := shared.SpanOf(body[:width])
	return ConditionClause{
		Predicate: ConditionPredicateObjectMatches, Selection: selection, Negated: negated,
		SubjectSpan: span, HasSubjectSpan: true, SubjectRefID: atoms.ReferenceIDAt(span),
		SubjectTypes: subjectTypes,
		SubjectPast:  true,
	}, true
}

// Legacy characteristic amounts own their cost noun without a free reference.
// Only an exact typed predicate consumer introduces the new paid-cost binding.
func emitPaidCostPredicateReferences(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		ability.Atoms.references = paidCostPredicateReferences(ability.Atoms.references, ability.ConditionClauses)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				mode.Atoms.references = paidCostPredicateReferences(mode.Atoms.references, mode.ConditionClauses)
			}
		}
	}
}

func paidCostPredicateReferences(references []Reference, conditions []ConditionClause) []Reference {
	return slices.DeleteFunc(slices.Clone(references), func(reference Reference) bool {
		if reference.Kind != ReferencePaidCostSubject {
			return false
		}
		for _, condition := range conditions {
			if condition.Predicate == ConditionPredicateObjectMatches && condition.HasSubjectSpan &&
				condition.SubjectRefID == reference.NodeID && condition.SubjectSpan == reference.Span {
				return false
			}
		}
		return true
	})
}

func emitPaidCostSubjectBindings(abilities []Ability) {
	for ai := range abilities {
		ability := &abilities[ai]
		if ability.Kind != AbilityActivated && ability.Kind != AbilitySpell {
			continue
		}
		bindPaidCostReferences(abilities, ai, ability.SemanticReferences,
			paidCostSpellSentences(abilities, ai, ability.Sentences))
		if ability.Modal != nil {
			for mi := range ability.Modal.Options {
				mode := &ability.Modal.Options[mi]
				bindPaidCostReferences(abilities, ai, mode.SemanticReferences,
					paidCostSpellSentences(abilities, ai, paidCostModeSentences(ability, mi)))
			}
		}
	}
}

func bindPaidCostReferences(abilities []Ability, consumer int, references []Reference, sentences []Sentence) {
	for ri := range references {
		reference := &references[ri]
		if reference.Kind != ReferencePaidCostSubject || reference.PaidCost == nil {
			continue
		}
		binding := *reference.PaidCost
		binding.ConsumerNodeID = reference.NodeID
		if paidCostHasCompetingEffect(sentences, binding.Domain, reference.Span) {
			reference.PaidCost = &binding
			continue
		}
		candidates, producers := paidCostComponents(abilities, consumer, binding.Domain)
		if len(candidates) == 1 && paidCostRequiredProducer(abilities, consumer, producers[0]) {
			component := candidates[0]
			if component.AmountKnown && component.AmountValue == 1 && !component.AmountFromX &&
				!component.DiscardWholeHand && component.ChoiceGroup == 0 {
				binding.Producer = producers[0]
				binding.Known = true
				component.PaidSubject = &binding.Producer
			}
		}
		reference.PaidCost = &binding
	}
}

func paidCostComponents(abilities []Ability, consumer int, domain PaidCostDomain) ([]*CostComponent, []PaidCostProducer) {
	var candidates []*CostComponent
	var producers []PaidCostProducer
	for ai := range abilities {
		ability := &abilities[ai]
		if ability.CostSyntax == nil ||
			(ai != consumer && (abilities[consumer].Kind != AbilitySpell ||
				ability.Kind != AbilitySpellAdditionalCost && ability.Kind != AbilitySpellAlternativeCost)) {
			continue
		}
		for ci := range ability.CostSyntax.Components {
			component := &ability.CostSyntax.Components[ci]
			if paidCostDomainMatchesComponent(domain, component.Kind) {
				candidates = append(candidates, component)
				producers = append(producers, PaidCostProducer{ClauseID: ai + 1, ComponentNodeID: ci + 1})
			}
		}
	}
	return candidates, producers
}

// Alternatives and later declarations count toward ambiguity, not publication.
func paidCostRequiredProducer(abilities []Ability, consumer int, producer PaidCostProducer) bool {
	index := producer.ClauseID - 1
	return index == consumer ||
		index < consumer && abilities[index].Kind == AbilitySpellAdditionalCost
}

func paidCostSpellSentences(abilities []Ability, consumer int, sentences []Sentence) []Sentence {
	if abilities[consumer].Kind != AbilitySpell {
		return sentences
	}
	sentences = slices.Clone(sentences)
	for _, preceding := range abilities[:consumer] {
		if preceding.Kind != AbilitySpell {
			continue
		}
		sentences = append(sentences, preceding.Sentences...)
		if preceding.Modal != nil {
			for _, mode := range preceding.Modal.Options {
				sentences = append(sentences, mode.Sentences...)
			}
		}
	}
	return sentences
}

func paidCostHasCompetingEffect(sentences []Sentence, domain PaidCostDomain, consumer shared.Span) bool {
	for _, sentence := range sentences {
		for _, effect := range sentence.Effects {
			if effect.VerbSpan.Start.Offset < consumer.Start.Offset &&
				(domain == PaidCostDomainSacrificedPermanent && effect.Kind == EffectSacrifice ||
					domain == PaidCostDomainDiscardedCard && effect.Kind == EffectDiscard) {
				return true
			}
		}
	}
	return false
}

func paidCostDomainMatchesComponent(domain PaidCostDomain, kind CostComponentKind) bool {
	return domain == PaidCostDomainSacrificedPermanent && kind == CostComponentSacrifice ||
		domain == PaidCostDomainDiscardedCard && kind == CostComponentDiscard
}

func paidCostModeSentences(ability *Ability, modeIndex int) []Sentence {
	modal := ability.Modal
	sentences := slices.Clone(modal.Options[modeIndex].Sentences)
	if modal.ChoiceKnown && modal.MaxModes <= 1 &&
		modal.ChoiceBonus.AdditionalMaxModes == 0 && modal.ChoiceBonus.MaxModes <= 1 {
		return sentences
	}
	for _, preceding := range modal.Options[:modeIndex] {
		sentences = append(sentences, preceding.Sentences...)
	}
	return sentences
}

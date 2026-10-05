package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// DelayedSubjectKind names a grammatical antecedent, not a runtime snapshot.
type DelayedSubjectKind uint8

// Delayed subject kinds distinguish exact grammatical ownership domains.
const (
	DelayedSubjectUnknown DelayedSubjectKind = iota
	DelayedSubjectSource
	DelayedSubjectTarget
	DelayedSubjectProduct
	DelayedSubjectEvent
	DelayedSubjectEventRelated
	DelayedSubjectUnsupported
)

// DelayedSubjectOwnership preserves the exact subject of a fixed-phase body.
type DelayedSubjectOwnership struct {
	Kind                  DelayedSubjectKind `json:",omitempty"`
	ReferenceNodeIDs      []int              `json:",omitempty"`
	ProducerClauseID      int                `json:",omitempty"`
	TargetOccurrence      int                `json:",omitempty"`
	CardZone              zone.Type          `json:",omitempty"`
	CardIdentity          bool               `json:",omitempty"`
	DirectTarget          bool               `json:",omitempty"`
	OptionalAtDelayedTime bool               `json:",omitempty"`
}

func emitDelayedSubjectOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		emitDelayedSubjects(ability.Sentences, ability.SemanticReferences, ability.ConditionSegments, ability.Trigger)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				emitDelayedSubjects(mode.Sentences, mode.SemanticReferences, mode.ConditionSegments, ability.Trigger)
			}
		}
	}
}

func emitDelayedSubjects(sentences []Sentence, references []Reference, conditions []ConditionSegment, trigger *TriggerClause) {
	var effects []*EffectSyntax
	var targets []TargetSyntax
	for si := range sentences {
		targets = append(targets, sentences[si].Targets...)
		for ei := range sentences[si].Effects {
			effects = append(effects, &sentences[si].Effects[ei])
		}
	}
	for i, effect := range effects {
		if effect.DelayedTiming == DelayedTimingNone {
			continue
		}
		// The established optional-return fallback retains the original exiled
		// card, not the unavailable permanent result of the declined return.
		if i > 0 && effect.Kind == EffectReturn && effect.ToZone == zone.Battlefield &&
			effects[i-1].Optional && effects[i-1].Kind == EffectReturn &&
			slices.ContainsFunc(conditions, func(condition ConditionSegment) bool {
				return !condition.Ownership.DelayedBody &&
					condition.Ownership.ResultProducerClauseID == effects[i-1].ClauseID &&
					slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) &&
					strings.EqualFold(strings.TrimSpace(condition.Text), "If you don't")
			}) {
			continue
		}
		immediate := *effect
		immediate.References = slices.DeleteFunc(slices.Clone(effect.References), func(reference Reference) bool {
			return delayedConditionOwnsReference(conditions, reference.NodeID)
		})
		immediate.Tokens, _ = cutDelayedTiming(effect.Tokens)
		if len(immediate.Tokens) > 0 && immediate.Tokens[len(immediate.Tokens)-1].Kind == shared.Period {
			immediate.Tokens = immediate.Tokens[:len(immediate.Tokens)-1]
		}
		if effectWordsAt(immediate.Tokens, 0, "you", "may") {
			immediate.Tokens = immediate.Tokens[2:]
		}
		immediate.DelayedTiming = DelayedTimingNone
		// Optionality is carried by the enclosing ability/instruction, not by
		// the immediate action's grammatical reconstruction.
		immediate.Optional = false
		if exactEffectSyntax(&immediate) || exactDelayedPluralDisposal(&immediate) ||
			exactReferencedCardHandMove(&immediate) || exactReferencedZoneMove(&immediate) {
			effect.Exact = true
		}

		var subject DelayedSubjectOwnership
		if len(effect.Targets) == 1 {
			for occurrence, target := range targets {
				if target.Span == effect.Targets[0].Span {
					subject = DelayedSubjectOwnership{Kind: DelayedSubjectTarget, TargetOccurrence: occurrence, DirectTarget: true}
					break
				}
			}
		}
		for _, owned := range effect.References {
			if delayedConditionOwnsReference(conditions, owned.NodeID) {
				continue
			}
			index := slices.IndexFunc(references, func(reference Reference) bool {
				return reference.NodeID == owned.NodeID
			})
			if index < 0 {
				subject.Kind = DelayedSubjectUnsupported
				continue
			}
			reference := references[index]
			next := delayedReferenceSubject(reference, effects[:i], targets, references, trigger)
			if subject.Kind != DelayedSubjectUnknown &&
				(subject.Kind != next.Kind || subject.ProducerClauseID != next.ProducerClauseID ||
					subject.TargetOccurrence != next.TargetOccurrence) {
				subject.Kind = DelayedSubjectUnsupported
			} else {
				subject.Kind = next.Kind
				subject.ProducerClauseID = next.ProducerClauseID
				subject.TargetOccurrence = next.TargetOccurrence
			}
			subject.ReferenceNodeIDs = append(subject.ReferenceNodeIDs, reference.NodeID)
		}
		// Some exact token-disposal clauses have an implicit definite noun
		// rather than a collected reference. Its antecedent is still parser-owned.
		if effect.CreatedTokensReference && len(subject.ReferenceNodeIDs) == 0 {
			subject = delayedProductSubject(effects[:i])
		}
		if subject.Kind == DelayedSubjectSource && (effect.FromZone == zone.Graveyard || effect.FromZone == zone.Exile) {
			subject.CardIdentity = true
			subject.CardZone = effect.FromZone
		}
		if trigger != nil && trigger.TriggerEvent != nil &&
			(trigger.TriggerEvent.ZoneChange.Kind == TriggerEventZoneChangeDied || trigger.TriggerEvent.Kind == TriggerEventKindSacrificed) &&
			(subject.Kind == DelayedSubjectEvent ||
				subject.Kind == DelayedSubjectSource && trigger.TriggerEvent.Subject.Kind == TriggerEventSubjectSelf) {
			subject.CardZone = zone.Graveyard
		}
		if subject.Kind == DelayedSubjectSource && trigger != nil && trigger.TriggerEvent != nil &&
			trigger.TriggerEvent.ZoneChange.Kind == TriggerEventZoneChangeDied &&
			slices.ContainsFunc(effect.References, func(reference Reference) bool {
				return reference.Kind == ReferenceThisObject && strings.EqualFold(joinedEffectText(reference.Tokens), "this card")
			}) {
			subject.CardZone = zone.Graveyard
		}
		if trigger != nil && trigger.TriggerEvent != nil &&
			trigger.TriggerEvent.Kind == TriggerEventKindZoneChange &&
			trigger.TriggerEvent.Zone.MatchFromZone && trigger.TriggerEvent.Zone.FromZone.Kind == TriggerEventZoneBattlefield &&
			(subject.Kind == DelayedSubjectEvent ||
				subject.Kind == DelayedSubjectSource && trigger.TriggerEvent.Subject.Kind == TriggerEventSubjectSelf) &&
			(effect.Kind == EffectReturn && effect.ToZone == zone.Battlefield ||
				slices.ContainsFunc(effect.References, func(reference Reference) bool {
					return slices.ContainsFunc(reference.Tokens, func(token shared.Token) bool {
						return strings.EqualFold(token.Text, "card")
					})
				})) {
			subject.CardIdentity = true
		}
		effect.DelayedSubject = subject
		if effect.Optional {
			for _, sentence := range sentences {
				if !parserSpanContains(sentence.Span, effect.VerbSpan) {
					continue
				}
				end := slices.IndexFunc(sentence.Tokens, func(token shared.Token) bool {
					return token.Span.Start.Offset >= effect.OptionalSpan.Start.Offset
				})
				if end < 0 {
					continue
				}
				for ti := range end {
					if leadingDelayedTiming(sentence.Tokens[ti:end]) != DelayedTimingNone {
						effect.DelayedSubject.OptionalAtDelayedTime = true
						break
					}
				}
			}
		}
	}
}

func delayedConditionOwnsReference(conditions []ConditionSegment, nodeID int) bool {
	return slices.ContainsFunc(conditions, func(condition ConditionSegment) bool {
		return slices.Contains(condition.Ownership.ReferenceNodeIDs, nodeID)
	})
}

func exactDelayedPluralDisposal(effect *EffectSyntax) bool {
	if effect.CreatedTokensReference || slices.ContainsFunc(effect.References, func(reference Reference) bool {
		return reference.Kind == ReferencePronoun && (reference.Pronoun == PronounThose || reference.Pronoun == PronounThem)
	}) {
		var verb string
		switch effect.Kind {
		case EffectExile:
			verb = "Exile"
		case EffectSacrifice:
			verb = "Sacrifice"
		case EffectDestroy:
			verb = "Destroy"
		default:
			return false
		}
		if verb != "" && (strings.EqualFold(exactEffectClauseText(effect), verb+" those tokens.") ||
			strings.EqualFold(exactEffectClauseText(effect), verb+" the tokens.")) {
			return true
		}
	}
	if len(effect.References) != 1 {
		return false
	}
	reference := effect.References[0]
	if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName &&
		reference.Kind != ReferenceThatObject && reference.Kind != ReferencePronoun {
		return false
	}
	var verb string
	switch effect.Kind {
	case EffectExile:
		verb = "Exile"
	case EffectSacrifice:
		verb = "Sacrifice"
	case EffectDestroy:
		verb = "Destroy"
	default:
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), verb+" "+joinedEffectText(effect.References[0].Tokens)+".")
}

func exactReferencedCardHandMove(effect *EffectSyntax) bool {
	if effect.Kind != EffectPut || effect.ToZone != zone.Hand || len(effect.References) != 1 ||
		effect.References[0].Kind != ReferenceThatObject ||
		!strings.EqualFold(joinedEffectText(effect.References[0].Tokens), "that card") {
		return false
	}
	return strings.EqualFold(exactEffectClauseText(effect), "Put that card into your hand.")
}

func exactReferencedZoneMove(effect *EffectSyntax) bool {
	if len(effect.References) == 0 {
		return false
	}
	reference := effect.References[0]
	if reference.Kind != ReferenceThisObject && reference.Kind != ReferenceSelfName &&
		reference.Kind != ReferenceThatObject && reference.Kind != ReferencePronoun {
		return false
	}
	subject := joinedEffectText(reference.Tokens)
	if effect.Kind == EffectReturn && effect.ToZone == zone.Hand {
		return strings.EqualFold(exactEffectClauseText(effect), "Return "+subject+" to its owner's hand.")
	}
	if effect.Kind == EffectPut && effect.ToZone == zone.Library {
		position := ""
		switch effect.Destination {
		case EffectDestinationTop:
			position = "top"
		case EffectDestinationBottom:
			position = "the bottom"
		default:
			return false
		}
		return position != "" && strings.EqualFold(exactEffectClauseText(effect), "Put "+subject+" on "+position+" of its owner's library.")
	}
	if effect.Kind != EffectReturn || effect.ToZone != zone.Battlefield {
		return false
	}
	text := "Return " + subject + " to the battlefield"
	if effect.EntersTapped {
		text += " tapped"
	}
	if effect.EntersTransformed {
		text += " transformed"
	}
	if effect.UnderYourControl {
		text += " under your control"
	} else if effect.UnderOwnersControl {
		text += " under its owner's control"
	}
	if effect.CounterKnown {
		if !effect.CounterKind.Valid() || !effect.Amount.Known || effect.Amount.Value != 1 {
			return false
		}
		text += " with a " + effect.CounterKind.String() + " counter on it"
	}
	return strings.EqualFold(exactEffectClauseText(effect), text+".")
}

func delayedReferenceSubject(reference Reference, effects []*EffectSyntax, targets []TargetSyntax, references []Reference, trigger *TriggerClause) DelayedSubjectOwnership {
	for _, token := range reference.Tokens {
		if strings.EqualFold(token.Text, "spell") {
			return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
		}
	}
	if reference.Kind == ReferenceSelfName || reference.Kind == ReferenceThisObject {
		return DelayedSubjectOwnership{Kind: DelayedSubjectSource}
	}
	if reference.Kind != ReferenceThatObject && reference.Kind != ReferencePronoun {
		return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
	}
	// Explicit source mentions and target/product introductions compete in
	// grammatical order. Unconditional non-object clauses introduce no subject.
	bestOrder := -1
	subject := DelayedSubjectOwnership{}
	for _, prior := range references {
		if prior.Order.End > reference.Order.Start ||
			(trigger != nil && prior.Order.Start < trigger.Order.End) ||
			(prior.Kind != ReferenceSelfName && prior.Kind != ReferenceThisObject) {
			continue
		}
		if !delayedReferenceMatchesSource(reference, prior) {
			continue
		}
		if prior.Order.Start > bestOrder {
			bestOrder = prior.Order.Start
			subject = DelayedSubjectOwnership{Kind: DelayedSubjectSource}
		}
	}
	for occurrence, target := range targets {
		if target.Order.End > reference.Order.Start || target.Selection.Kind == SelectionPlayer ||
			!delayedReferenceMatchesSelection(reference, target.Selection.Kind) {
			continue
		}
		if target.Order.Start > bestOrder {
			bestOrder = target.Order.Start
			subject = DelayedSubjectOwnership{Kind: DelayedSubjectTarget, TargetOccurrence: occurrence}
		}
	}
	productFloor := bestOrder
	products := 0
	for _, effect := range effects {
		if effect.Order.End <= bestOrder || !delayedReferenceMatchesProduct(reference, effect) {
			continue
		}
		switch effect.Kind {
		case EffectCreate, EffectExile, EffectReveal, EffectDig, EffectSearch, EffectManifestDread, EffectChoosePermanent:
		case EffectReturn, EffectPut:
			if effect.ToZone != zone.Battlefield {
				continue
			}
		default:
			continue
		}
		if effect.Order.End > productFloor {
			products++
			if products > 1 && delayedReferencePlural(reference) {
				return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
			}
		}
		bestOrder = effect.Order.End
		subject = DelayedSubjectOwnership{Kind: DelayedSubjectProduct, ProducerClauseID: effect.ClauseID}
	}
	if subject.Kind != DelayedSubjectUnknown {
		return subject
	}
	if trigger == nil || trigger.TriggerEvent == nil || trigger.TriggerEvent.OneOrMore {
		return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
	}
	event := trigger.TriggerEvent
	if event.Kind == TriggerEventKindSpellCast || event.Kind == TriggerEventKindAbilityActivated {
		return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
	}
	if reference.Kind == ReferenceThatObject &&
		(event.Kind == TriggerEventKindBlock || event.Kind == TriggerEventKindBecameBlocked) &&
		selectionHasType(event.RelatedSelection, TriggerCardTypeCreature) &&
		delayedReferenceMatchesEvent(reference, event.RelatedSelection) {
		return DelayedSubjectOwnership{Kind: DelayedSubjectEventRelated}
	}
	if event.Kind == TriggerEventKindDamageDealt &&
		event.DamageRecipient.Kind == TriggerEventDamageRecipientPermanent &&
		delayedReferenceMatchesEvent(reference, event.DamageRecipient.Selection) {
		return DelayedSubjectOwnership{Kind: DelayedSubjectEvent}
	}
	if event.Kind == TriggerEventKindDamageDealt && event.DamageSource.Kind == TriggerEventSubjectSelf &&
		reference.Kind == ReferencePronoun {
		return DelayedSubjectOwnership{Kind: DelayedSubjectSource}
	}
	if event.Subject.Kind == TriggerEventSubjectSelf {
		if slices.ContainsFunc(references, func(prior Reference) bool {
			return prior.Order.Start < trigger.Order.End && !delayedReferenceMatchesSource(reference, prior)
		}) {
			return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
		}
		return DelayedSubjectOwnership{Kind: DelayedSubjectSource}
	}

	if event.Subject.Kind == TriggerEventSubjectSelection && delayedPermanentEvent(event) &&
		delayedReferenceMatchesEvent(reference, event.Subject.Selection) &&
		!slices.Contains(event.Subject.Selection.RequiredTypes, TriggerCardTypeInstant) &&
		!slices.Contains(event.Subject.Selection.RequiredTypes, TriggerCardTypeSorcery) {
		return DelayedSubjectOwnership{Kind: DelayedSubjectEvent}
	}
	return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
}

func delayedReferencePlural(reference Reference) bool {
	return reference.Pronoun == PronounThem || reference.Pronoun == PronounThose ||
		slices.ContainsFunc(reference.Tokens, func(token shared.Token) bool {
			return strings.EqualFold(token.Text, "tokens") || strings.EqualFold(token.Text, "creatures") ||
				strings.EqualFold(token.Text, "cards")
		})
}

func delayedReferenceMatchesSource(reference, source Reference) bool {
	if source.Kind != ReferenceThisObject {
		return true
	}
	known := false
	for _, token := range source.Tokens {
		var selection SelectionKind
		switch strings.ToLower(token.Text) {
		case "creature":
			selection = SelectionCreature
		case "land":
			selection = SelectionLand
		case "artifact":
			selection = SelectionArtifact
		case "enchantment":
			selection = SelectionEnchantment
		case "planeswalker":
			selection = SelectionPlaneswalker
		case "battle":
			selection = SelectionBattle
		default:
			continue
		}
		known = true
		if delayedReferenceMatchesSelection(reference, selection) {
			return true
		}
	}
	return !known
}

func delayedPermanentEvent(event *TriggerEventClause) bool {
	switch event.Kind {
	case TriggerEventKindZoneChange:
		return event.Zone.MatchFromZone && event.Zone.FromZone.Kind == TriggerEventZoneBattlefield ||
			event.Zone.MatchToZone && event.Zone.ToZone.Kind == TriggerEventZoneBattlefield
	case TriggerEventKindAttack, TriggerEventKindBlock, TriggerEventKindBecameBlocked,
		TriggerEventKindFight, TriggerEventKindCounterAdded, TriggerEventKindBecomesTapped,
		TriggerEventKindBecomesUntapped, TriggerEventKindTurnedFaceUp, TriggerEventKindSacrificed,
		TriggerEventKindBecameMonstrous, TriggerEventKindMutated, TriggerEventKindBecameTarget,
		TriggerEventKindTokenCreated, TriggerEventKindDied, TriggerEventKindAttacksUnblocked:
		return true
	default:
		return false
	}
}

func delayedReferenceMatchesEvent(reference Reference, selection TriggerSelection) bool {
	for _, token := range reference.Tokens {
		var kind TriggerCardType
		switch strings.ToLower(token.Text) {
		case "creature":
			kind = TriggerCardTypeCreature
		case "land":
			kind = TriggerCardTypeLand
		case "artifact":
			kind = TriggerCardTypeArtifact
		case "enchantment":
			kind = TriggerCardTypeEnchantment
		case "planeswalker":
			kind = TriggerCardTypePlaneswalker
		case "battle":
			kind = TriggerCardTypeBattle
		default:
			continue
		}
		return selectionHasType(selection, kind)
	}
	return true
}

func delayedReferenceMatchesSelection(reference Reference, selection SelectionKind) bool {
	if reference.Kind == ReferencePronoun {
		return true
	}
	for _, token := range reference.Tokens {
		switch strings.ToLower(token.Text) {
		case "creature":
			return selection == SelectionCreature
		case "artifact":
			return selection == SelectionArtifact
		case "land":
			return selection == SelectionLand
		case "enchantment":
			return selection == SelectionEnchantment
		case "planeswalker":
			return selection == SelectionPlaneswalker
		case "battle":
			return selection == SelectionBattle
		case "spell":
			return false
		}
	}
	return true
}

func delayedReferenceMatchesProduct(reference Reference, effect *EffectSyntax) bool {
	if reference.Kind == ReferencePronoun {
		return true
	}
	for _, token := range reference.Tokens {
		switch strings.ToLower(token.Text) {
		case "token", "tokens":
			return effect.Kind == EffectCreate
		case "card":
			return effect.Kind != EffectCreate
		case "creature":
			if effect.Kind == EffectCreate && effect.TokenPTKnown {
				return true
			}
			return slices.ContainsFunc(effect.Targets, func(target TargetSyntax) bool {
				return target.Selection.Kind == SelectionCreature
			})
		case "land", "artifact", "enchantment", "planeswalker", "battle":
			return slices.ContainsFunc(effect.Targets, func(target TargetSyntax) bool {
				return delayedReferenceMatchesSelection(reference, target.Selection.Kind)
			})
		}
	}
	return true
}

func delayedProductSubject(effects []*EffectSyntax) DelayedSubjectOwnership {
	subject := DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
	for i := len(effects) - 1; i >= 0; i-- {
		effect := effects[i]
		if effect.Kind == EffectCreate {
			if subject.Kind == DelayedSubjectProduct {
				return DelayedSubjectOwnership{Kind: DelayedSubjectUnsupported}
			}
			subject = DelayedSubjectOwnership{Kind: DelayedSubjectProduct, ProducerClauseID: effect.ClauseID}
		}
		if len(effect.Targets) != 0 || effect.Kind == EffectExile ||
			effect.Kind == EffectReveal || effect.Kind == EffectSearch {
			break
		}
	}
	return subject
}

func delayedConditionEvaluation(effect *EffectSyntax, segment *ConditionSegment, sentences []Sentence) bool {
	if effect.DelayedTiming == DelayedTimingNone {
		return false
	}
	if segment.Span.Start.Offset > effect.VerbSpan.Start.Offset {
		return true
	}
	for _, sentence := range sentences {
		if !parserSpanContains(sentence.Span, segment.Span) {
			continue
		}
		comma := shared.TopLevelIndex(sentence.Tokens, shared.Comma)
		if comma >= 0 && leadingDelayedTiming(sentence.Tokens[:comma+1]) != DelayedTimingNone {
			return true
		}
	}
	for i, token := range effect.Tokens {
		if token.Span.Start.Offset == segment.Span.Start.Offset {
			return leadingDelayedTiming(effect.Tokens[:i]) != DelayedTimingNone
		}
	}
	return false
}

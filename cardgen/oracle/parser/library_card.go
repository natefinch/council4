package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func recognizeLibraryCardEffects(sentence *Sentence) {
	for i := range sentence.Effects {
		effect := &sentence.Effects[i]
		if recognizeSingleLibraryCardProducer(effect) {
			effect.Exact = true
			for _, target := range effect.Targets {
				for j := range sentence.Targets {
					if sentence.Targets[j].Span == target.Span {
						sentence.Targets[j] = target
					}
				}
			}
		}
	}
}

func recognizeSingleLibraryCardProducer(effect *EffectSyntax) bool {
	if effect.Kind != EffectLookAtLibraryTop && effect.Kind != EffectReveal ||
		effect.Negated || effect.DelayedTiming != DelayedTimingNone {
		return false
	}
	context, targetKind, ok := singleLibraryCardProducerOwner(effect.Kind, exactEffectClauseText(effect))
	if !ok || context == EffectContextController && len(effect.Targets) != 0 ||
		context == EffectContextTarget && len(effect.Targets) != 1 {
		return false
	}
	if context == EffectContextTarget {
		target := effect.Targets[0]
		target.Selection = SelectionSyntax{Kind: targetKind}
		target.Text = "target player"
		if targetKind == SelectionOpponent {
			target.Text = "target opponent"
		}
		target.Cardinality = TargetCardinalitySyntax{Min: 1, Max: 1}
		target.Exact = true
		effect.Targets = []TargetSyntax{target}
	}
	effect.Context = context
	effect.CardSource = EffectCardSourceTopOfPlayerLibrary
	effect.Amount = EffectAmountSyntax{Known: true, Value: 1}
	return true
}

func singleLibraryCardProducerOwner(kind EffectKind, clause string) (EffectContextKind, SelectionKind, bool) {
	verb := "Reveal"
	if kind == EffectLookAtLibraryTop {
		verb = "Look at"
	}
	for _, owner := range []struct {
		text       string
		context    EffectContextKind
		targetKind SelectionKind
	}{
		{"your", EffectContextController, SelectionUnknown},
		{"target player's", EffectContextTarget, SelectionPlayer},
		{"target opponent's", EffectContextTarget, SelectionOpponent},
	} {
		if !strings.EqualFold(clause, verb+" the top card of "+owner.text+" library.") {
			continue
		}
		return owner.context, owner.targetKind, true
	}
	return EffectContextUnknown, SelectionUnknown, false
}

func libraryCardObservationBefore(tokens []shared.Token, index int) bool {
	for start := 0; start < index; start++ {
		kind := EffectReveal
		if equalWord(tokens[start], "look") {
			kind = EffectLookAtLibraryTop
		} else if !equalWord(tokens[start], "reveal") {
			continue
		}
		for end := start; end < index; end++ {
			if tokens[end].Kind != shared.Period {
				continue
			}
			if _, _, ok := singleLibraryCardProducerOwner(kind, joinTokens(tokens[start:end+1])); ok {
				return true
			}
			break
		}
	}
	return false
}

func recognizeLibraryCardAction(effect *EffectSyntax) bool {
	if effect.Kind != EffectReveal && effect.Kind != EffectPut ||
		effect.Negated || effect.DelayedTiming != DelayedTimingNone ||
		len(effect.Targets) != 0 || effect.CounterKnown {
		return false
	}
	clause := exactEffectClauseText(effect)
	for _, subject := range []string{"it", "that card", "the revealed card", "the looked-at card"} {
		if effect.Kind == EffectReveal && strings.EqualFold(clause, "Reveal "+subject+".") {
			effect.CardSource = EffectCardSourcePriorInstructionResult
			effect.FromZone = zone.Library
			effect.Amount = EffectAmountSyntax{Known: true, Value: 1}
			return true
		}
		for _, destination := range []struct {
			text     string
			zone     zone.Type
			position EffectDestinationPosition
			tapped   bool
			control  bool
			owner    bool
		}{
			{"into your hand", zone.Hand, EffectDestinationUnspecified, false, false, false},
			{"into your graveyard", zone.Graveyard, EffectDestinationUnspecified, false, false, false},
			{"on the bottom of your library", zone.Library, EffectDestinationBottom, false, false, false},
			{"on top of your library", zone.Library, EffectDestinationTop, false, false, false},
			{"onto the battlefield", zone.Battlefield, EffectDestinationUnspecified, false, false, false},
			{"onto the battlefield tapped", zone.Battlefield, EffectDestinationUnspecified, true, false, false},
			{"onto the battlefield under your control", zone.Battlefield, EffectDestinationUnspecified, false, true, false},
			{"onto the battlefield tapped under your control", zone.Battlefield, EffectDestinationUnspecified, true, true, false},
			{"into that player's hand", zone.Hand, EffectDestinationUnspecified, false, false, true},
			{"into that player's graveyard", zone.Graveyard, EffectDestinationUnspecified, false, false, true},
			{"on the bottom of that player's library", zone.Library, EffectDestinationBottom, false, false, true},
			{"on top of that player's library", zone.Library, EffectDestinationTop, false, false, true},
		} {
			if effect.Kind != EffectPut || !strings.EqualFold(clause, "Put "+subject+" "+destination.text+".") {
				continue
			}
			effect.CardSource = EffectCardSourcePriorInstructionResult
			effect.FromZone, effect.ToZone = zone.Library, destination.zone
			effect.Destination = destination.position
			effect.EntersTapped, effect.UnderYourControl = destination.tapped, destination.control
			effect.LibraryOwnerDestination = destination.owner
			return true
		}
	}
	return false
}

func emitLibraryCardReferenceOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		bindLibraryCardReferences(ability.Sentences, ability.SemanticReferences, ability.ConditionSegments)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				bindLibraryCardReferences(mode.Sentences, mode.SemanticReferences, mode.ConditionSegments)
			}
		}
	}
}

func bindLibraryCardReferences(sentences []Sentence, references []Reference, conditions []ConditionSegment) {
	var effects []*EffectSyntax
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effects = append(effects, &sentences[si].Effects[ei])
		}
	}
	emitTargetOwnerLibraryProducerClauses(effects)
	for ri := range references {
		reference := &references[ri]
		if !libraryCardReference(*reference) {
			continue
		}
		reference.LibraryCardObservation = explicitLibraryCardObservation(*reference) != EffectUnknown
		cardNoun := libraryReferenceHasCardNoun(*reference, effects, references, conditions)
		latest := 0
		cardObservation := 0
		cardEntry := 0
		reached := false
		observations := make(map[EffectKind]int)
		for _, effect := range effects {
			if effect.VerbSpan.End.Offset > reference.Span.Start.Offset ||
				parserSpanContains(effect.ClauseSpan, reference.Span) ||
				libraryReferenceIsOwnActionCondition(*reference, effect.ClauseID, conditions) {
				continue
			}
			_, _, observation := singleLibraryCardProducerOwner(effect.Kind, exactEffectClauseText(effect))
			if (observation || effect.LibraryOwnerClauseID > 0) &&
				effect.Exact && effect.CardSource == EffectCardSourceTopOfPlayerLibrary &&
				(effect.Kind == EffectReveal || effect.Kind == EffectLookAtLibraryTop) {
				latest = effect.ClauseID
				cardObservation, cardEntry = latest, 0
				reached = false
				observations[effect.Kind] = latest
				continue
			}
			if producer := linkedLibraryCardRevealProducer(*effect, latest, observations); producer != 0 {
				// Revealing an observed card preserves its original top-card identity.
				latest = producer
				observations[EffectReveal] = producer
				continue
			}
			action := *effect
			if latest > 0 && action.Kind == EffectPut && recognizeLibraryCardAction(&action) && action.ToZone == zone.Battlefield &&
				!libraryReferenceConsumesObservationElse(*reference, effect,
					libraryCardAntecedent(*reference, latest, observations), effects, references, conditions) {
				latest = effect.ClauseID
				cardEntry = latest
				reached = true
				clear(observations)
			}
			if len(effect.Targets) != 0 || effect.Kind == EffectCreate && !cardNoun ||
				effect.Kind == EffectSearch || effect.Kind == EffectChoosePermanent ||
				effect.Kind == EffectExile || effect.Kind == EffectManifestDread {
				latest = 0
				cardObservation, cardEntry = 0, 0
				reached = false
				clear(observations)
			}
		}
		producer := libraryCardAntecedent(*reference, latest, observations)
		if cardObservation > 0 && cardEntry > 0 && libraryReferenceNamesCard(*reference, conditions) &&
			slices.ContainsFunc(conditions, func(condition ConditionSegment) bool {
				return parserSpanContains(condition.Span, reference.Span)
			}) {
			producer, reached = cardObservation, false
			reference.ReachedCardProducerClauseID = cardEntry
		}
		if producer == 0 {
			continue
		}
		if !reached && reference.SubjectNoun != ObjectNounUnknown &&
			reference.SubjectNoun != ObjectNounCard {
			continue
		}
		if reached {
			for _, effect := range effects {
				if !parserSpanContains(effect.ClauseSpan, reference.Span) ||
					effect.Kind != EffectPut || !recognizeLibraryCardAction(effect) ||
					effect.ToZone == zone.Battlefield {
					continue
				}
				reference.CardIdentity = true
			}
		}
		for _, candidate := range references {
			if candidate.Span.Start.Offset >= reference.Span.Start.Offset {
				continue
			}
			if candidate.Kind != ReferenceSelfName && candidate.Kind != ReferenceThisObject {
				continue
			}
			for _, effect := range effects {
				if effect.ClauseID == producer && candidate.Span.Start.Offset > effect.VerbSpan.End.Offset {
					producer = 0
				}
			}
		}
		reference.ProducerClauseID = producer
		reference.LibraryCardObservation = !reached && (reference.LibraryCardObservation || producer > 0)
	}
	bindEnteredTargetReferences(effects, references, conditions)
	bindLibraryOwnerDestinationReferences(effects, references)
	for _, effect := range effects {
		observedSubject := false
		for _, list := range [][]Reference{effect.References, effect.SubjectReferences} {
			for i := range list {
				for _, reference := range references {
					if list[i].NodeID == reference.NodeID {
						list[i].ProducerClauseID = reference.ProducerClauseID
						list[i].LibraryCardObservation = reference.LibraryCardObservation
						list[i].CardIdentity = reference.CardIdentity
						list[i].ReachedCardProducerClauseID = reference.ReachedCardProducerClauseID
						observedSubject = observedSubject || reference.ProducerClauseID > 0
						break
					}
				}
				if observedSubject && recognizeLibraryCardAction(effect) {
					effect.Exact = true
					if list[i].ProducerClauseID > 0 && !list[i].LibraryCardObservation &&
						list[i].CardIdentity && effect.Kind == EffectPut && effect.ToZone != zone.Battlefield {
						effect.FromZone = zone.Battlefield
					}
				}
			}
		}
	}
}

func libraryReferenceIsOwnActionCondition(reference Reference, clauseID int, conditions []ConditionSegment) bool {
	for _, condition := range conditions {
		if slices.Contains(condition.Ownership.ClauseIDs, clauseID) &&
			slices.Contains(condition.Ownership.ReferenceNodeIDs, reference.NodeID) {
			return true
		}
	}
	return false
}

func libraryReferenceHasCardNoun(reference Reference, effects []*EffectSyntax, references []Reference, conditions []ConditionSegment) bool {
	if libraryReferenceNamesCard(reference, conditions) {
		return true
	}
	for _, condition := range conditions {
		for _, effect := range effects {
			if !parserSpanContains(effect.ClauseSpan, reference.Span) {
				continue
			}
			for _, owner := range condition.Ownership.ClauseIDs {
				if owner != effect.ClauseID {
					continue
				}
				for _, subject := range references {
					if subject.ProducerClauseID > 0 && parserSpanContains(condition.Span, subject.Span) {
						return true
					}
				}
			}
		}
	}
	return false
}

func libraryReferenceNamesCard(reference Reference, conditions []ConditionSegment) bool {
	if reference.CardIdentity {
		return true
	}
	for _, condition := range conditions {
		if parserSpanContains(condition.Span, reference.Span) &&
			strings.HasSuffix(strings.ToLower(condition.Text), " card") {
			return true
		}
	}
	return false
}

func libraryReferenceConsumesObservationElse(reference Reference, previous *EffectSyntax, producer int,
	effects []*EffectSyntax, references []Reference, conditions []ConditionSegment) bool {
	otherwise := false
	for _, effect := range effects {
		if parserSpanContains(effect.ClauseSpan, reference.Span) && effect.Connection == EffectConnectionOtherwise {
			action := *effect
			otherwise = recognizeLibraryCardAction(&action)
			break
		}
	}
	if !otherwise || producer <= 0 {
		return false
	}
	for _, condition := range conditions {
		for _, owner := range condition.Ownership.ClauseIDs {
			if owner != previous.ClauseID {
				continue
			}
			for _, subject := range references {
				if subject.ProducerClauseID == producer && parserSpanContains(condition.Span, subject.Span) {
					return true
				}
			}
		}
	}
	return false
}

func libraryCardReference(reference Reference) bool {
	if reference.Kind == ReferencePronoun {
		return reference.Pronoun == PronounIt || reference.Pronoun == PronounIts
	}
	if reference.Kind != ReferenceThatObject {
		return false
	}
	if reference.SubjectNoun != ObjectNounUnknown {
		return true
	}
	for _, noun := range []string{"that card", "that card's", "the revealed card", "the looked-at card"} {
		if strings.EqualFold(reference.Text, noun) {
			return true
		}
	}
	return false
}

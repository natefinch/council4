package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func libraryOwnerPossessiveAt(tokens []shared.Token, index int) bool {
	return index+2 < len(tokens) && equalWord(tokens[index], "that") &&
		equalWord(tokens[index+1], "player's") &&
		(equalWord(tokens[index+2], "library") || equalWord(tokens[index+2], "hand") ||
			equalWord(tokens[index+2], "graveyard")) && libraryCardObservationBefore(tokens, index)
}

func bindLibraryOwnerDestinationReferences(effects []*EffectSyntax, references []Reference) {
	for _, effect := range effects {
		action := *effect
		if !recognizeLibraryCardAction(&action) || !action.LibraryOwnerDestination {
			continue
		}
		producer, ownerIndex := 0, -1
		ambiguous := false
		for i, reference := range references {
			if !parserSpanContains(effect.ClauseSpan, reference.Span) ||
				reference.Span.Start.Offset < effect.VerbSpan.End.Offset {
				continue
			}
			if libraryCardReference(reference) {
				if producer != 0 || reference.ProducerClauseID <= 0 {
					ambiguous = true
				}
				producer = reference.ProducerClauseID
			} else if reference.Kind == ReferenceThatPlayer {
				if ownerIndex >= 0 {
					ambiguous = true
				}
				ownerIndex = i
			}
		}
		if ambiguous || producer <= 0 || ownerIndex < 0 {
			continue
		}
		for _, candidate := range effects {
			if candidate.ClauseID != producer || !candidate.Exact ||
				candidate.CardSource != EffectCardSourceTopOfPlayerLibrary ||
				candidate.Context != EffectContextTarget || len(candidate.Targets) != 1 {
				continue
			}
			references[ownerIndex].ProducerClauseID = producer
			present := false
			for _, reference := range action.References {
				present = present || reference.NodeID == references[ownerIndex].NodeID
			}
			if !present {
				action.References = append(action.References, references[ownerIndex])
			}
			*effect = action
		}
	}
}

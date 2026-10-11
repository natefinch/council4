package parser

func emitChosenPermanentNameOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		bindChosenPermanentNames(ability.Sentences, ability.SemanticReferences)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				bindChosenPermanentNames(mode.Sentences, mode.SemanticReferences)
			}
		}
	}
}

func bindChosenPermanentNames(sentences []Sentence, references []Reference) {
	effects := conditionEffects(sentences)
	for i := range references {
		reference := &references[i]
		if reference.Kind != ReferenceThatObject || len(reference.Tokens) != 3 ||
			!equalWord(reference.Tokens[0], "the") || !equalWord(reference.Tokens[1], "chosen") {
			continue
		}
		noun, ok := recognizeObjectNoun(reference.Tokens[2])
		if !ok {
			continue
		}
		consumer := false
		for _, effect := range effects {
			if effect.Kind == EffectSearch && effect.SearchSameNameAsChosenObject &&
				parserSpanContains(effect.ClauseSpan, reference.Span) {
				consumer = true
			}
		}
		if !consumer {
			continue
		}
		reference.SubjectNoun = noun
		var choice *EffectSyntax
		choices := 0
		for _, effect := range effects {
			if effect.Kind == EffectChoosePermanent {
				choices++
			}
		}
		ambiguous := choices != 1
		for _, effect := range effects {
			if effect.VerbSpan.Start.Offset >= reference.Span.Start.Offset {
				continue
			}
			switch effect.Kind {
			case EffectChoosePermanent:
				if choice != nil || !effect.Exact || effect.Negated ||
					!enteredNounCompatible(noun, effect.Selection.Kind) {
					ambiguous = true
				}
				choice = effect
			case EffectBolster, EffectManifest, EffectManifestDread, EffectCloak,
				EffectPopulate, EffectChooseCreatureType:
				if choice != nil {
					ambiguous = true
				}
			default:
				continue
			}
		}
		if choice != nil && !ambiguous {
			reference.ChoiceClauseID = choice.ClauseID
		}
		for _, effect := range effects {
			for _, list := range [][]Reference{effect.References, effect.SubjectReferences} {
				for j := range list {
					if list[j].NodeID == reference.NodeID {
						list[j].ChoiceClauseID = reference.ChoiceClauseID
						list[j].SubjectNoun = noun
					}
				}
			}
		}
	}
}

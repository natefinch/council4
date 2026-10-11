package parser

func bindTopExiledCardReferences(sentences []Sentence, references []Reference) {
	var effects []*EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effects = append(effects, &sentences[i].Effects[j])
		}
	}
	for i := range references {
		reference := &references[i]
		if reference.ProducerClauseID != 0 || reference.LibraryCardObservation ||
			!libraryCardReference(*reference) {
			continue
		}
		castCost := false
		for _, effect := range effects {
			if effect.Kind == EffectCast && effect.CastWithoutPayingManaCost &&
				effect.Selection.Kind == SelectionSpell && reference.Pronoun == PronounIts &&
				parserSpanContains(effect.ClauseSpan, reference.Span) {
				castCost = true
			}
		}
		if castCost {
			continue
		}
		producer := 0
		for _, effect := range effects {
			if effect.VerbSpan.End.Offset > reference.Span.Start.Offset ||
				parserSpanContains(effect.ClauseSpan, reference.Span) {
				continue
			}
			if effect.Kind == EffectExile && effect.Exact && !effect.Negated &&
				effect.CardSource == EffectCardSourceTopOfPlayerLibrary &&
				effect.Amount.Known && effect.Amount.Value == 1 {
				producer = effect.ClauseID
			} else if len(effect.Targets) > 0 || effect.Kind == EffectSearch || effect.Kind == EffectDig ||
				effect.Kind == EffectCreate || effect.Kind == EffectReveal || effect.Kind == EffectLookAtLibraryTop ||
				effect.Kind == EffectPut || effect.Kind == EffectReturn || effect.Kind == EffectExile {
				producer = 0
			}
		}
		if producer != 0 {
			reference.ProducerClauseID = producer
			reference.CardIdentity = true
		}
	}
}

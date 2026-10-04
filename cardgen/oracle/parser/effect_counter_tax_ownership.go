package parser

func emitCounterTaxConditionOwners(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		bindCounterTaxConditionOwners(ability.Sentences, ability.ConditionBoundaries)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				bindCounterTaxConditionOwners(mode.Sentences, mode.ConditionBoundaries)
			}
		}
	}
}

func bindCounterTaxConditionOwners(sentences []Sentence, boundaries []ConditionBoundary) {
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			if effect.Kind != EffectCounter ||
				effect.Payment.Form != EffectPaymentFormUnless ||
				effect.Payment.Payer != EffectPaymentPayerTargetController {
				continue
			}
			// Zero is a valid boundary identity; use -1 when no boundary owns
			// this already-recognized payment.
			effect.Payment.FailureConditionNodeID = -1
			boundary, ok := conditionBoundaryAt(boundaries, effect.Payment.Span.Start)
			if ok && boundary.Kind == ConditionIntroUnless && !boundary.Intervening && !boundary.DurationSkip {
				effect.Payment.FailureConditionNodeID = boundary.NodeID
			}
		}
	}
}

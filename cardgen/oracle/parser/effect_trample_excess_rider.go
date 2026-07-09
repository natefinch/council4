package parser

// stripTrampleExcessRiderConditionSemantics clears the condition and keyword
// accessors an ability carries alongside a source-trample-gated excess-damage
// redirect (Ram Through: "Target creature you control deals damage equal to its
// power to target creature you don't control. If the creature you control has
// trample, excess damage is dealt to that creature's controller instead.").
//
// The excess-damage recognizer already consumes the whole "If the creature you
// control has trample, excess damage is dealt ... instead." sentence into one
// EffectDealDamage carrying the RequireSourceTrample marker, so the leading "If
// ... has trample" clause is not a standalone ability condition and the trailing
// "trample" noun is not a granted keyword. Left in place, the independent
// condition-boundary and semantic-keyword scans would still surface that clause
// as an unrecognized ability condition and "trample" as a bare keyword, which the
// text-blind compiler and lowering cannot reconcile with the redirect the effect
// already models. Clearing them mirrors the animate-self and coin-flip strip
// passes: the marked effect owns the sentence, so its condition and keyword
// spans are covered by the effect.
func stripTrampleExcessRiderConditionSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasSourceTrampleExcessRider(ability) {
			continue
		}
		ability.ConditionBoundaries = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.ConditionClauses = nil
		ability.EventHistoryConditions = nil
		ability.SemanticKeywords = nil
	}
}

// abilityHasSourceTrampleExcessRider reports whether any of the ability's
// sentences carries an excess-damage-to-controller redirect gated on the source
// having trample (the RequireSourceTrample marker the excess recognizer sets).
func abilityHasSourceTrampleExcessRider(ability *Ability) bool {
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			effect := &ability.Sentences[i].Effects[j]
			if effect.Kind == EffectDealDamage && effect.RequireSourceTrample {
				return true
			}
		}
	}
	return false
}

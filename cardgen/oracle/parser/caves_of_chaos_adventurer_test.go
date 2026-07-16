package parser

import "testing"

func TestParseCavesOfChaosConditionalImpulse(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks, exile the top card of your library. "+
			"If you've completed a dungeon, you may play that card this turn without paying its mana cost. "+
			"Otherwise, you may play that card this turn.",
		Context{CardName: "Caves of Chaos Adventurer"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	var effects []EffectSyntax
	for si := range ability.Sentences {
		effects = append(effects, ability.Sentences[si].Effects...)
	}
	if len(effects) != 2 {
		t.Fatalf("effects = %#v", effects)
	}
	free := effects[0]
	if free.Kind != EffectImpulseExile ||
		!free.Exact ||
		free.Optional ||
		!free.Amount.Known ||
		free.Amount.Value != 1 ||
		free.Duration != EffectDurationThisTurn ||
		!free.ImpulseWithoutPayingManaCost {
		t.Fatalf("free impulse = %#v", free)
	}
	normal := effects[1]
	if normal.Kind != EffectImpulseExile ||
		!normal.Exact ||
		normal.Optional ||
		!normal.Amount.Known ||
		normal.Amount.Value != 1 ||
		normal.Duration != EffectDurationThisTurn ||
		normal.ImpulseWithoutPayingManaCost ||
		normal.Connection != EffectConnectionOtherwise {
		t.Fatalf("normal impulse = %#v", normal)
	}
	if len(ability.ConditionClauses) != 1 ||
		ability.ConditionClauses[0].Predicate != ConditionPredicateControllerCompletedADungeon {
		t.Fatalf("conditions = %#v", ability.ConditionClauses)
	}
}

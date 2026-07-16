package parser

import (
	"slices"
	"testing"
)

const olorinSearingLightText = "Each opponent exiles a creature with the greatest power among creatures that player controls.\n" +
	"Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled."

func TestParseEachOpponentGreatestPowerExileAndCorrelatedDamage(t *testing.T) {
	document, diagnostics := Parse(olorinSearingLightText, Context{
		CardName:         "Olórin's Searing Light",
		InstantOrSorcery: true,
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	exile := document.Abilities[0].Sentences[0].Effects[0]
	if !exile.Exact || !exile.ExileEachOpponentChoosesGreatestPower {
		t.Fatalf("exile = %#v", exile)
	}
	if exile.Context != EffectContextEachOpponent ||
		!slices.Equal(exile.Selection.RequiredTypesAny, []CardType{CardTypeCreature}) {
		t.Fatalf("exile context/selection = %v, %#v", exile.Context, exile.Selection)
	}
	damageAbility := document.Abilities[1]
	damage := damageAbility.Sentences[0].Effects[0]
	if !damage.Exact || !damage.DamageEachOpponentCorrelatedExiledPower {
		t.Fatalf("damage = %#v", damage)
	}
	if damageAbility.AbilityWord == nil || damageAbility.AbilityWord.Label != "Spell mastery" {
		t.Fatalf("ability word = %#v", damageAbility.AbilityWord)
	}
	if len(damageAbility.ConditionClauses) != 1 ||
		damageAbility.ConditionClauses[0].Predicate != ConditionPredicateGraveyardInstantOrSorceryCountAtLeast ||
		damageAbility.ConditionClauses[0].Threshold != 2 {
		t.Fatalf("conditions = %#v", damageAbility.ConditionClauses)
	}
}

package parser

import "testing"

// castEffectFromAttackTrigger parses source as a single ability, returns its two
// ordered effects (exile, then cast), and fails the test if the shape differs.
func castEffectFromAttackTrigger(t *testing.T, source string) (exile, cast EffectSyntax) {
	t.Helper()
	ability := parseSingleAbility(t, source, Context{})
	if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 2 {
		t.Fatalf("effects = %#v, want exile then cast", ability.Sentences)
	}
	return ability.Sentences[0].Effects[0], ability.Sentences[0].Effects[1]
}

// TestParseCastAnyNumberWithoutPayingTheirManaCosts proves the plural
// "cast any number of spells ... without paying their mana costs" free-cast
// clause (Etali, Primal Storm) sets both the CastWithoutPayingManaCost flag and
// the unbounded AnyNumber count, the two positive signals the exile-each-top
// free-cast lowering reads.
func TestParseCastAnyNumberWithoutPayingTheirManaCosts(t *testing.T) {
	t.Parallel()
	_, cast := castEffectFromAttackTrigger(t,
		"Whenever this creature attacks, exile the top card of each player's library, then you may cast any number of spells from among those cards without paying their mana costs.")
	if cast.Kind != EffectCast {
		t.Fatalf("kind = %v, want EffectCast", cast.Kind)
	}
	if !cast.CastWithoutPayingManaCost {
		t.Fatal("CastWithoutPayingManaCost = false, want true for the plural \"their mana costs\" form")
	}
	if !cast.Amount.AnyNumber {
		t.Fatal("Amount.AnyNumber = false, want true for \"cast any number of spells\"")
	}
}

// TestParseSingularCastWithoutPayingItsManaCostHasNoAnyNumber proves the
// singular "cast a spell ... without paying its mana cost" form still sets the
// free-cast flag but leaves AnyNumber unset, so the unbounded count stays a
// positive signal exclusive to the "any number of" wording.
func TestParseSingularCastWithoutPayingItsManaCostHasNoAnyNumber(t *testing.T) {
	t.Parallel()
	_, cast := castEffectFromAttackTrigger(t,
		"Whenever this creature attacks, exile the top card of each player's library, then you may cast a spell from among those cards without paying its mana cost.")
	if !cast.CastWithoutPayingManaCost {
		t.Fatal("CastWithoutPayingManaCost = false, want true for the singular form")
	}
	if cast.Amount.AnyNumber {
		t.Fatal("Amount.AnyNumber = true, want false for the singular \"cast a spell\" form")
	}
}

package cardgen

import (
	"testing"
)

// TestLowerPhantomPreventDamageRemovesCounter verifies the Phantom damage
// prevention replacement ("If damage would be dealt to this creature, prevent
// that damage. Remove a +1/+1 counter from this creature.") lowers to a
// replacement that prevents all damage and removes one +1/+1 counter.
func TestLowerPhantomPreventDamageRemovesCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Phantom Tiger",
		Layout:   "normal",
		TypeLine: "Creature — Cat Spirit",
		ManaCost: "{2}{G}",
		OracleText: "This creature enters with two +1/+1 counters on it.\n" +
			"If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature.",
		Power:     new("0"),
		Toughness: new("0"),
	})
	found := false
	for _, ability := range face.ReplacementAbilities {
		r := ability.Replacement
		if r.DamagePreventAll && r.DamagePreventedRemovesPlusOneCounter && r.DamageRecipientSelf {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prevent-all/remove-counter replacement found: %+v", face.ReplacementAbilities)
	}
}

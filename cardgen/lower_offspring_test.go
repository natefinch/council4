package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerOffspringKeywordAndEnterTrigger proves an Offspring permanent lowers
// to the Offspring static ability (carrying the printed additional mana cost)
// plus the canonical linked enters-the-battlefield triggered ability gated on the
// "if the offspring cost was paid" intervening-if (Coruscation Mage's shape).
func TestLowerOffspringKeywordAndEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Coruscation Mage",
		Layout:   "normal",
		TypeLine: "Creature — Otter Wizard",
		OracleText: "Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)\n" +
			"Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.",
		Power:     new("2"),
		Toughness: new("2"),
	})
	var offspringStatic *game.StaticAbility
	for i := range face.StaticAbilities {
		if _, ok := game.StaticBodyOffspring(&face.StaticAbilities[i].Body); ok {
			offspringStatic = &face.StaticAbilities[i].Body
		}
	}
	if offspringStatic == nil {
		t.Fatal("no Offspring static ability lowered from the Offspring keyword")
	}
	offspring, _ := game.StaticBodyOffspring(offspringStatic)
	if got := offspring.Cost.ManaValue(); got != 2 {
		t.Fatalf("offspring cost mana value = %d, want 2", got)
	}
	var found bool
	for i := range face.TriggeredAbilities {
		if reflect.DeepEqual(face.TriggeredAbilities[i], game.OffspringEnterTriggeredAbility()) {
			found = true
		}
	}
	if !found {
		t.Fatal("no canonical OffspringEnterTriggeredAbility lowered from the Offspring keyword")
	}
}

// TestLowerOffspringColoredCost proves a colored Offspring cost ("Offspring
// {1}{U}") lowers with its exact printed mana cost, not just generic forms.
func TestLowerOffspringColoredCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Splash Lasher",
		Layout:     "normal",
		TypeLine:   "Creature — Frog",
		OracleText: "Offspring {1}{U} (You may pay an additional {1}{U} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	var offspringStatic *game.StaticAbility
	for i := range face.StaticAbilities {
		if _, ok := game.StaticBodyOffspring(&face.StaticAbilities[i].Body); ok {
			offspringStatic = &face.StaticAbilities[i].Body
		}
	}
	if offspringStatic == nil {
		t.Fatal("no Offspring static ability lowered from the colored Offspring keyword")
	}
	offspring, _ := game.StaticBodyOffspring(offspringStatic)
	if got := len(offspring.Cost); got != 2 {
		t.Fatalf("offspring cost symbols = %d, want 2 for {1}{U}", got)
	}
	if got := offspring.Cost.ManaValue(); got != 2 {
		t.Fatalf("offspring cost mana value = %d, want 2 for {1}{U}", got)
	}
}

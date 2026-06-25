package cardgen

import (
	"strings"
	"testing"
)

// TestPrimaryDamageSourceAttribution pins the DamageSource attribution that the
// shared primaryDamageSource helper assigns to primary deal-damage instructions
// across the representative damage shapes (#1748). Each primary-damage lowering
// previously rebuilt the same source-binding branch; consolidating it must keep
// every shape's attribution byte-identical:
//   - a permanent source ("this creature", "it" bound to the source) carries an
//     explicit game.SourcePermanentReference() so the runtime credits the
//     source's last-known keywords (lifelink, deathtouch);
//   - a spell source carries no DamageSource (the spell, not a permanent, deals
//     the damage), so no DamageSource field is rendered;
//   - a follow-on rider shares the primary instruction's source attribution.
func TestPrimaryDamageSourceAttribution(t *testing.T) {
	t.Parallel()
	const sourcePermanent = "DamageSource: opt.Val(game.SourcePermanentReference())"
	tests := []struct {
		name    string
		card    *ScryfallCard
		present []string
		absent  []string
	}{
		{
			name: "creature single-target burn carries source permanent",
			card: &ScryfallCard{
				Name:       "Test Sorcerer",
				Layout:     "normal",
				TypeLine:   "Creature — Human Wizard",
				OracleText: "{T}: This creature deals 1 damage to any target.",
				Power:      new("1"),
				Toughness:  new("1"),
			},
			present: []string{
				"Primitive: game.Damage",
				"game.Fixed(1)",
				"game.AnyTargetDamageRecipient(0)",
				sourcePermanent,
			},
		},
		{
			name: "spell single-target burn omits damage source",
			card: &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: "Test Bolt deals 3 damage to any target.",
				Colors:     []string{"R"},
			},
			present: []string{
				"Primitive: game.Damage",
				"game.Fixed(3)",
				"game.AnyTargetDamageRecipient(0)",
			},
			absent: []string{"DamageSource:"},
		},
		{
			name: "spell divided damage omits damage source",
			card: &ScryfallCard{
				Name:       "Test Split",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: "Test Split deals 2 damage divided as you choose among one or two targets.",
				Colors:     []string{"R"},
			},
			present: []string{
				"Primitive: game.Damage",
				"game.Fixed(2)",
				"Divided:   true",
			},
			absent: []string{"DamageSource:"},
		},
		{
			name: "creature damage to you carries source permanent and controller",
			card: &ScryfallCard{
				Name:       "Test Brute",
				Layout:     "normal",
				TypeLine:   "Creature — Ogre Warrior",
				OracleText: "{T}: This creature deals 2 damage to you.",
				Power:      new("3"),
				Toughness:  new("3"),
			},
			present: []string{
				"Primitive: game.Damage",
				"game.Fixed(2)",
				"game.ControllerReference()",
				sourcePermanent,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(test.card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.present {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
			for _, unwanted := range test.absent {
				if strings.Contains(source, unwanted) {
					t.Fatalf("source unexpectedly contains %q:\n%s", unwanted, source)
				}
			}
		})
	}
}

// TestPrimaryDamageSourceRiderSharesSource verifies that a primary single-target
// damage clause with a follow-on "and N damage to you" rider attributes both the
// primary instruction and the rider to the same permanent source, exercising the
// shared primaryDamageSource helper threaded onto the rider instruction.
func TestPrimaryDamageSourceRiderSharesSource(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Striker",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "{T}: This creature deals 2 damage to any target and 1 damage to you.",
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	const sourcePermanent = "DamageSource: opt.Val(game.SourcePermanentReference())"
	// Both the primary "to any target" instruction and the "to you" rider must
	// carry the source-permanent attribution.
	if got := strings.Count(source, sourcePermanent); got != 2 {
		t.Fatalf("source-permanent attributions = %d, want 2 (primary + rider):\n%s", got, source)
	}
	if !strings.Contains(source, "game.Fixed(2)") ||
		!strings.Contains(source, "game.Fixed(1)") {
		t.Fatalf("expected fixed 2 primary and fixed 1 rider amounts:\n%s", source)
	}
}

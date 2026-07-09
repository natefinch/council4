package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateMassacreGirl proves the real Massacre Girl card generates fully:
// Menace, the enters ability that gives every OTHER creature -1/-1 until end of
// turn (a battlefield group excluding the source), and the this-turn delayed
// trigger that fires on each creature death to repeat the mass -1/-1. This is the
// key mechanic unlocked for Epic #2705's broadest generalization: a general
// "whenever a creature dies this turn" delayed trigger whose body is a mass P/T
// debuff excluding the source.
func TestGenerateMassacreGirl(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:      "Massacre Girl",
		Layout:    "normal",
		ManaCost:  "{3}{B}{B}",
		TypeLine:  "Legendary Creature — Human Assassin",
		Colors:    []string{"B"},
		Power:     new("4"),
		Toughness: new("4"),
		OracleText: "Menace\n" +
			"When Massacre Girl enters, each other creature gets -1/-1 until end of turn. " +
			"Whenever a creature dies this turn, each creature other than Massacre Girl gets -1/-1 until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.MenaceStaticBody",
		"game.EventPermanentEnteredBattlefield",
		// The enters -1/-1 applies to every creature except Massacre Girl herself.
		"game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference())",
		"Primitive: game.CreateDelayedTrigger",
		"game.EventPermanentDied",
		"game.DelayedWindowThisTurn",
		"PowerDelta:     -1,",
		"ToughnessDelta: -1,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateGeneralDeathDelayedTriggerSiblings proves the same general
// death-delayed-trigger recognizer correctly lowers the sibling cards that
// previously failed open (silently dropping the delayed trigger): Death Frenzy
// and Gnawing Crescendo now both emit a CreateDelayedTrigger keyed on
// EventPermanentDied and bounded to this turn.
func TestGenerateGeneralDeathDelayedTriggerSiblings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pkg     string
		card    *ScryfallCard
		wantAll []string
	}{
		{
			name: "Death Frenzy",
			pkg:  "d",
			card: &ScryfallCard{
				Name:     "Death Frenzy",
				Layout:   "normal",
				ManaCost: "{3}{B}{G}",
				TypeLine: "Sorcery",
				Colors:   []string{"B", "G"},
				OracleText: "All creatures get -2/-2 until end of turn. " +
					"Whenever a creature dies this turn, you gain 1 life.",
			},
			wantAll: []string{
				"Primitive: game.CreateDelayedTrigger",
				"game.EventPermanentDied",
				"game.DelayedWindowThisTurn",
				"Primitive: game.GainLife",
			},
		},
		{
			name: "Gnawing Crescendo",
			pkg:  "g",
			card: &ScryfallCard{
				Name:     "Gnawing Crescendo",
				Layout:   "normal",
				ManaCost: "{2}{R}",
				TypeLine: "Instant",
				Colors:   []string{"R"},
				OracleText: "Creatures you control get +2/+0 until end of turn. " +
					"Whenever a nontoken creature you control dies this turn, create a 1/1 black Rat creature token with \"This token can't block.\"",
			},
			wantAll: []string{
				"Primitive: game.CreateDelayedTrigger",
				"game.EventPermanentDied",
				"Controller:       game.TriggerControllerYou,",
				"NonToken: true",
				"game.DelayedWindowThisTurn",
				"Primitive: game.CreateToken",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(test.card, test.pkg)
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wantAll {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestGeneralDeathDelayedTriggerFailsClosed proves the recognizer never fails
// open on death-delayed-trigger wording it cannot faithfully lower. Warhost's
// Frenzy gates its "whenever a creature you control dies this turn" trigger
// behind an "if this spell was kicked" condition, and Skeletonize's trigger is
// filtered by "a creature dealt damage this way"; neither is representable, so
// both must produce no source and a diagnostic rather than silently dropping the
// delayed ability.
func TestGeneralDeathDelayedTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		pkg  string
		card *ScryfallCard
	}{
		{
			name: "Warhost's Frenzy",
			pkg:  "w",
			card: &ScryfallCard{
				Name:     "Warhost's Frenzy",
				Layout:   "normal",
				ManaCost: "{2}{R}",
				TypeLine: "Instant",
				Colors:   []string{"R"},
				OracleText: "Kicker {B} (You may pay an additional {B} as you cast this spell.)\n" +
					"Creatures you control get +2/+0 until end of turn. " +
					"If this spell was kicked, whenever a creature you control dies this turn, draw a card.",
			},
		},
		{
			name: "Skeletonize",
			pkg:  "s",
			card: &ScryfallCard{
				Name:     "Skeletonize",
				Layout:   "normal",
				ManaCost: "{4}{R}",
				TypeLine: "Instant",
				Colors:   []string{"R"},
				OracleText: "Skeletonize deals 3 damage to target creature. " +
					"When a creature dealt damage this way dies this turn, create a 1/1 black Skeleton creature token with \"{B}: Regenerate this token.\"",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(test.card, test.pkg)
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("expected fail-closed (empty source) but card generated:\n%s", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected a diagnostic explaining the fail-closed reason, got none")
			}
		})
	}
}

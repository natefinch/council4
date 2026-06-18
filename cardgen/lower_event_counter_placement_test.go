package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerEventPermanentCounterPlacement proves that "put a +1/+1 counter on
// it" / "on that creature" trigger bodies, whose pronoun denotes the permanent
// involved in the triggering event, lower to an AddCounter addressing
// game.EventPermanentReference() with the stated kind and fixed amount. This is
// the ETB / attack / combat-damage / becomes-targeted counter pattern.
func TestLowerEventPermanentCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		kind       counter.Kind
		amount     int
		optional   bool
	}{
		{
			name:       "attacking creature",
			typeLine:   "Creature — Human",
			oracleText: "Whenever a creature you control attacks, put a +1/+1 counter on it.",
			kind:       counter.PlusOnePlusOne,
			amount:     1,
		},
		{
			name:       "entering creature",
			typeLine:   "Creature — Unicorn",
			oracleText: "Whenever another creature you control enters, put a +1/+1 counter on that creature.",
			kind:       counter.PlusOnePlusOne,
			amount:     1,
		},
		{
			name:       "combat damage two counters",
			typeLine:   "Creature — Vampire",
			oracleText: "Whenever this creature deals combat damage to a player, put two +1/+1 counters on it.",
			kind:       counter.PlusOnePlusOne,
			amount:     2,
		},
		{
			name:       "becomes targeted minus counter",
			typeLine:   "Creature — Wurm",
			oracleText: "Whenever this creature becomes the target of a spell or ability, put a -1/-1 counter on it.",
			kind:       counter.MinusOneMinusOne,
			amount:     1,
		},
		{
			name:       "blocking creature",
			typeLine:   "Creature — Leech",
			oracleText: "Whenever this creature becomes blocked by a creature, put a -1/-1 counter on that creature.",
			kind:       counter.MinusOneMinusOne,
			amount:     1,
		},
		{
			name:       "optional becomes tapped",
			typeLine:   "Creature — Merfolk",
			oracleText: "Whenever this creature becomes tapped, you may put a +1/+1 counter on it.",
			kind:       counter.PlusOnePlusOne,
			amount:     1,
			optional:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			power := "2"
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Event Counter",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      &power,
				Toughness:  &power,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			ability := face.TriggeredAbilities[0]
			if ability.Optional != test.optional {
				t.Fatalf("optional = %v, want %v", ability.Optional, test.optional)
			}
			mode := ability.Content.Modes[0]
			add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
			if !ok {
				t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
			}
			if add.Object != game.EventPermanentReference() {
				t.Fatalf("object = %#v, want EventPermanentReference", add.Object)
			}
			if add.CounterKind != test.kind {
				t.Fatalf("counter kind = %v, want %v", add.CounterKind, test.kind)
			}
			if add.Amount != game.Fixed(test.amount) {
				t.Fatalf("amount = %#v, want Fixed(%d)", add.Amount, test.amount)
			}
		})
	}
}

// TestLowerEventPermanentCounterPlacementFailsClosedInSequence proves that an
// "it" counter placement that follows an object-producing instruction in the
// same ordered sequence fails closed. There the compiler binds the pronoun whose
// antecedent is the created token to the triggering event permanent, so
// accepting it would place the counters on the wrong object. Only standalone
// effects keep the EventPermanent binding.
func TestLowerEventPermanentCounterPlacementFailsClosedInSequence(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"When this enchantment enters, create a 0/0 green and blue Fractal creature token. Put three +1/+1 counters on it.",
		"When this creature enters, create a 1/1 white Soldier creature token. Put a +1/+1 counter on it.",
	} {
		_, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Sequence Counter",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: oracleText,
		})
		if len(diagnostics) == 0 {
			t.Fatalf("%q lowered without diagnostics; expected fail-closed", oracleText)
		}
	}
}

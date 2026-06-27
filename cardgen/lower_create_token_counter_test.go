package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCreateTokenThenFixedCounters verifies the ordered pair "Create a
// <token>. Put <n> +1/+1 counters on it." lowers to a token creation that
// publishes its result under a link key, followed by a fixed counter placement
// whose recipient resolves to that linked token.
func TestLowerCreateTokenThenFixedCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fractal Maker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment enters, create a 0/0 green and blue Fractal creature token. Put three +1/+1 counters on it.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then counter placement", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a token creation publishing a link", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok ||
		add.Object.Kind() != game.ObjectReferenceLinkedObject ||
		add.Object.LinkID() != string(create.PublishLinked) {
		t.Fatalf("add = %#v, want a counter placement on the linked token", mode.Sequence[1].Primitive)
	}
	if got := add.Amount; got != game.Fixed(3) {
		t.Fatalf("amount = %#v, want fixed 3", got)
	}
}

// TestLowerCreateTokenThenVariableXCounters verifies the spell form "Create a
// ... token. Put X +1/+1 counters on it." (Fractal Summoning) places the spell's
// variable X count on the just-created token.
func TestLowerCreateTokenThenVariableXCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fractal Summoning",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{X}{G}{U}",
		OracleText: "Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered: %#v", face)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then counter placement", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a token creation publishing a link", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok ||
		add.Object.Kind() != game.ObjectReferenceLinkedObject ||
		add.Object.LinkID() != string(create.PublishLinked) {
		t.Fatalf("add = %#v, want a counter placement on the linked token", mode.Sequence[1].Primitive)
	}
	if dyn := add.Amount.DynamicAmount(); !dyn.Exists || dyn.Val.Kind != game.DynamicAmountX {
		t.Fatalf("amount = %#v, want variable X", add.Amount)
	}
}

// TestLowerCreateTokenThenDynamicCounters verifies the spell form "Create a ...
// token. Put a +1/+1 counter on it for each creature your opponents control."
// (Match the Odds) places an opponent-relative dynamic count on the created
// token rather than failing closed.
func TestLowerCreateTokenThenDynamicCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Match the Odds",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{G}",
		OracleText: "Create a 1/1 white Ally creature token. Put a +1/+1 counter on it for each creature your opponents control.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered: %#v", face)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then counter placement", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a token creation publishing a link", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	if !ok ||
		add.Object.Kind() != game.ObjectReferenceLinkedObject ||
		add.Object.LinkID() != string(create.PublishLinked) ||
		!add.Amount.IsDynamic() {
		t.Fatalf("add = %#v, want a dynamic counter placement on the linked token", mode.Sequence[1].Primitive)
	}
}

// TestLowerCreateTwoTokensThenCountersUnsupported verifies the plural form
// "Create two ... tokens. ... on each of them." is not lowered by the
// single-token counter sequence: the singular link cannot model a plural
// recipient, so the card fails closed rather than placing counters on one token.
func TestLowerCreateTwoTokensThenCountersUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Twin Fractals",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create two 0/0 green and blue Fractal creature tokens. Put two +1/+1 counters on each of them.",
	})
}

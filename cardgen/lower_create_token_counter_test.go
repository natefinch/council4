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

// TestLowerCreateTwoTokensThenCountersOnEach verifies the plural form "Create
// two ... tokens. Put two +1/+1 counters on each of them." lowers to a token
// creation that publishes every created token under a link key, followed by a
// counter placement whose recipient is that linked group, so each created token
// receives the counters.
func TestLowerCreateTwoTokensThenCountersOnEach(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Twin Fractals",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create two 0/0 green and blue Fractal creature tokens. Put two +1/+1 counters on each of them.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered: %#v", face)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then counter placement", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.Amount.Value() != 2 || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a two-token creation publishing a link", mode.Sequence[0].Primitive)
	}
	add, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	key, linked := add.Group.LinkedKey()
	if !ok || !linked || key != create.PublishLinked ||
		add.Object.Kind() != game.ObjectReferenceNone ||
		add.Amount.Value() != 2 {
		t.Fatalf("add = %#v, want two counters on the linked token group", mode.Sequence[1].Primitive)
	}
}

// TestLowerCreateTokenAndCountersOnItSingleSentence proves the create-then-link
// mechanism generalizes beyond the two-sentence "Create a token. Put N counters
// on it." shape every sibling test above exercises: Fractal Anomaly instead
// joins the creation and the placement with "and" in one sentence ("Create a
// ... token and put X +1/+1 counters on it, where X is ..."), which the
// hand-written lowerCreateTokenThenCountersSequence never recognized (its own
// guard requires exactly two ordered effects). This composes through the
// generic per-effect sequence loop instead: compiler.EffectCreate is a valid
// ReferenceBindingPriorInstructionResult antecedent (cardgen/oracle/compiler/
// reference.go), so the loop publishes the create instruction's link generically
// (sequencePriorInstructionLink) and the ordinary single-effect counter-placement
// lowerer (lowerReferencedCounterPlacement) resolves "it" through it — no
// bespoke pair function was written or needed for this shape.
func TestLowerCreateTokenAndCountersOnItSingleSentence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fractal Anomaly",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{G}{U}",
		OracleText: "Create a 0/0 green and blue Fractal creature token and put X +1/+1 counters on it, where X is the number of cards you've drawn this turn.",
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

// TestLowerCreateTwoTokensAndCountersOnEachOfThemSingleSentence guards the
// silent-miscompile hazard the generic path introduces: a plural back-reference
// ("each of them") to a multi-token creation must resolve to the whole
// LinkedObjectsGroup, never to a single LinkedObjectReference. lowerObjectReference
// only ever resolves one object, so treating a plural reference as an ordinary
// single one would place the counters on just one of the created tokens instead
// of each -- and would do so silently (a fully-formed, validated CardDef with no
// diagnostic), not fail closed. This shape reaches the generic path because a
// third effect ("draw a card") pushes it off lowerCreateTokenThenCountersSequence's
// exact-two-effect guard.
func TestLowerCreateTwoTokensAndCountersOnEachOfThemSingleSentence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Probe Twin Fractals Draw",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{G}{U}",
		OracleText: "Create two 1/1 white Soldier creature tokens, put a +1/+1 counter on each of them, then draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered: %#v", face)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %#v, want create, counter placement, then draw", mode.Sequence)
	}
	create2, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create2.Amount.Value() != 2 || create2.PublishLinked == "" {
		t.Fatalf("create = %#v, want a two-token creation publishing a link", mode.Sequence[0].Primitive)
	}
	add2, ok := mode.Sequence[1].Primitive.(game.AddCounter)
	key, linked := add2.Group.LinkedKey()
	if !ok || !linked || key != create2.PublishLinked ||
		add2.Object.Kind() != game.ObjectReferenceNone ||
		add2.Amount.Value() != 1 {
		t.Fatalf("add = %#v, want counters on the whole linked token group, not a single object", mode.Sequence[1].Primitive)
	}
}

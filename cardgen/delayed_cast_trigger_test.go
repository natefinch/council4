package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableShowdownDelayedCastTrigger covers Showdown of the
// Skalds chapter II/III, a repeating event-based delayed trigger ("Whenever you
// cast a spell this turn, ..."). The chapter lowers to a CreateDelayedTrigger
// whose EventPattern matches each spell the controller casts for the rest of the
// turn, bounded by the this-turn window, firing the +1/+1 counter body each time.
func TestGenerateExecutableShowdownDelayedCastTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Showdown of the Skalds",
		Layout:   "normal",
		ManaCost: "{2}{R}{W}",
		TypeLine: "Enchantment — Saga",
		Colors:   []string{"R", "W"},
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Exile the top four cards of your library. Until the end of your next turn, you may play those cards.\n" +
			"II, III — Whenever you cast a spell this turn, put a +1/+1 counter on target creature you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Chapters: []int{2, 3}",
		"Primitive: game.CreateDelayedTrigger",
		"EventPattern: opt.Val(game.TriggerPattern{",
		"Event:      game.EventSpellCast,",
		"Controller: game.TriggerControllerYou,",
		"Window: game.DelayedWindowThisTurn,",
		"Primitive: game.AddCounter",
		"CounterKind: counter.PlusOnePlusOne,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "Timing:") && strings.Contains(source, "DelayedTriggerDef") {
		// The event-based delayed trigger must not also carry a fixed-phase
		// timing; the two are mutually exclusive.
		t.Fatalf("event delayed trigger unexpectedly rendered a Timing field:\n%s", source)
	}
}

// TestGenerateExecutableOneShotDelayedCastTrigger covers a synthetic one-shot
// event delayed trigger ("When you next cast a creature spell this turn, ...")
// whose body places a counter on a chosen target creature, exercising the
// OneShot flag and the creature-spell pattern threading through to the rendered
// CardDef on a body distinct from Summon: Fenrir's. Fenrir's own chapter II body
// (a future-cast enters-with-counters replacement) and chapter III (greatest-
// power draw) are covered end to end by TestGenerateSummonFenrir.
func TestGenerateExecutableOneShotDelayedCastTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Test Next Cast Saga",
		Layout:   "normal",
		ManaCost: "{3}{G}",
		TypeLine: "Enchantment — Saga",
		Colors:   []string{"G"},
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II, III — When you next cast a creature spell this turn, put a +1/+1 counter on target creature you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Primitive: game.CreateDelayedTrigger",
		"EventPattern: opt.Val(game.TriggerPattern{",
		"game.EventSpellCast,",
		"CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}}",
		"OneShot: true,",
		"game.DelayedWindowThisTurn,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

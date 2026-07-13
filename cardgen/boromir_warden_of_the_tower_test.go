package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerBoromirWardenOfTheTower proves the whole card composes from existing
// generic pieces: Vigilance as a static keyword; the opponent cast trigger gated
// by the "if no mana was spent to cast it" intervening condition (ManaSpentToCast
// <= 0) whose body counters the triggering stack spell through the event stack
// reference without announcing a target; and the Sacrifice ability granting the
// controlled creature group indestructible until end of turn and tempting with
// the Ring. No card-specific code is needed — the "counter that spell." parser
// recognition is the only new generic surface.
func TestLowerBoromirWardenOfTheTower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Boromir, Warden of the Tower",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Human Soldier",
		ManaCost:  "{2}{W}",
		Power:     new("3"),
		Toughness: new("3"),
		OracleText: "Vigilance\n" +
			"Whenever an opponent casts a spell, if no mana was spent to cast it, counter that spell.\n" +
			"Sacrifice Boromir: Creatures you control gain indestructible until end of turn. The Ring tempts you.",
	})

	// Vigilance keyword static.
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (Vigilance)", len(face.StaticAbilities))
	}
	if face.StaticAbilities[0].Body.Text != "Vigilance" {
		t.Fatalf("static ability text = %q, want Vigilance", face.StaticAbilities[0].Body.Text)
	}

	// Opponent cast trigger with the no-mana intervening condition and a counter
	// of the triggering stack spell.
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Fatalf("trigger event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("trigger controller = %v, want TriggerControllerOpponent", ta.Trigger.Pattern.Controller)
	}
	if !ta.Trigger.InterveningCondition.Exists {
		t.Fatalf("intervening condition missing; trigger = %+v", ta.Trigger)
	}
	aggregates := ta.Trigger.InterveningCondition.Val.Aggregates
	if len(aggregates) != 1 {
		t.Fatalf("intervening aggregates = %#v, want exactly one", aggregates)
	}
	if got := aggregates[0]; got.Aggregate != game.AggregateEventSpellManaSpentToCast ||
		got.Op != compare.LessOrEqual || got.Value != 0 {
		t.Fatalf("intervening comparison = %#v, want AggregateEventSpellManaSpentToCast <= 0", got)
	}
	tmode := ta.Content.Modes[0]
	if len(tmode.Targets) != 0 {
		t.Fatalf("counter announced targets = %#v, want none (event stack reference)", tmode.Targets)
	}
	counter, ok := tmode.Sequence[0].Primitive.(game.CounterObject)
	if !ok || counter.Object != game.EventStackObjectReference() {
		t.Fatalf("trigger primitive = %#v, want CounterObject of the event stack object", tmode.Sequence[0].Primitive)
	}

	// Sacrifice ability: controlled-group indestructible until end of turn, then
	// the Ring tempts you.
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	aa := face.ActivatedAbilities[0]
	if len(aa.AdditionalCosts) != 1 || aa.AdditionalCosts[0].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("activation costs = %#v, want a single sacrifice-source cost", aa.AdditionalCosts)
	}
	amode := aa.Content.Modes[0]
	if len(amode.Sequence) != 2 {
		t.Fatalf("sacrifice ability sequence = %d instructions, want 2", len(amode.Sequence))
	}
	apply, ok := amode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("first sacrifice instruction = %#v, want ApplyContinuous", amode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("indestructible duration = %v, want DurationUntilEndOfTurn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want exactly one", apply.ContinuousEffects)
	}
	ce := apply.ContinuousEffects[0]
	if len(ce.AddKeywords) != 1 || ce.AddKeywords[0] != game.Indestructible {
		t.Fatalf("granted keywords = %#v, want [Indestructible]", ce.AddKeywords)
	}
	if sel := ce.Group.Selection(); sel.Controller != game.ControllerYou ||
		len(sel.RequiredTypes) != 1 || sel.RequiredTypes[0] != types.Creature {
		t.Fatalf("indestructible group selection = %#v, want creatures you control", ce.Group.Selection())
	}
	if _, ok := amode.Sequence[1].Primitive.(game.RingTempts); !ok {
		t.Fatalf("second sacrifice instruction = %#v, want RingTempts", amode.Sequence[1].Primitive)
	}
}

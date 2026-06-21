package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerModalLandfallTrigger proves a modal "choose one —" body lowers on a
// permanent zone-change (landfall) trigger the same way it does on a spell-cast
// trigger, with the "Landfall —" ability word stripped and both modes lowered.
func TestLowerModalLandfallTrigger(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Felidar Retreat",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "Landfall — Whenever a land you control enters, choose one —\n" +
			"• Create a 2/2 white Cat Beast creature token.\n" +
			"• Put a +1/+1 counter on each creature you control. " +
			"Those creatures gain vigilance until end of turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("trigger event = %#v, want permanent entered battlefield", ability.Trigger.Pattern)
	}
	if !ability.Content.IsModal() {
		t.Fatalf("content = %#v, want modal", ability.Content)
	}
	if ability.Content.MinModes != 1 || ability.Content.MaxModes != 1 ||
		len(ability.Content.Modes) != 2 {
		t.Fatalf("modal range = %d..%d over %d modes, want mandatory 1..1 over two modes",
			ability.Content.MinModes, ability.Content.MaxModes, len(ability.Content.Modes))
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("mode 0 primitive = %#v, want create token", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	secondMode := ability.Content.Modes[1]
	if len(secondMode.Sequence) != 2 {
		t.Fatalf("mode 1 sequence = %#v, want counter then keyword grant", secondMode.Sequence)
	}
	if _, ok := secondMode.Sequence[0].Primitive.(game.AddCounter); !ok {
		t.Fatalf("mode 1 primitive[0] = %#v, want add counter", secondMode.Sequence[0].Primitive)
	}
	if _, ok := secondMode.Sequence[1].Primitive.(game.ApplyContinuous); !ok {
		t.Fatalf("mode 1 primitive[1] = %#v, want apply continuous", secondMode.Sequence[1].Primitive)
	}
}

// TestLowerModalDiesTrigger proves the modal body support generalizes across
// permanent zone-change triggers beyond landfall (here, a creature-dies trigger).
func TestLowerModalDiesTrigger(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Reaper",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "Whenever a creature dies, choose one —\n" +
			"• Create a 1/1 white Cat creature token.\n" +
			"• You gain 1 life.",
	})
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("trigger event = %#v, want permanent died", ability.Trigger.Pattern)
	}
	if !ability.Content.IsModal() || len(ability.Content.Modes) != 2 {
		t.Fatalf("content = %#v, want two-mode modal", ability.Content)
	}
}

// TestLowerGroupCounterThenGroupKeywordSequence proves the ordered pair "Put a
// +1/+1 counter on each creature you control. Those creatures gain <keyword>
// until end of turn." lowers to a group counter placement followed by a keyword
// grant over that same group.
func TestLowerGroupCounterThenGroupKeywordSequence(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Counter Anthem",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put a +1/+1 counter on each creature you control. Those creatures gain trample until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter then keyword grant", mode.Sequence)
	}
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.PlusOnePlusOne || !add.Group.Valid() {
		t.Fatalf("add = %#v, want +1/+1 counter over a group", mode.Sequence[0].Primitive)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want unanchored group grant until end of turn", mode.Sequence[1].Primitive)
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility ||
		len(effect.AddKeywords) != 1 ||
		effect.AddKeywords[0] != game.Trample {
		t.Fatalf("effect = %+v, want trample keyword layer", effect)
	}
	// The grant's group must be exactly the counter placement's group so "those
	// creatures" resolves to the just-counted set.
	selection := effect.Group.Selection()
	if effect.Group.Domain() != game.GroupDomainBattlefield ||
		selection.Controller != game.ControllerYou ||
		len(selection.RequiredTypes) != 1 ||
		selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("group selection = %+v, want creatures you control", selection)
	}
}

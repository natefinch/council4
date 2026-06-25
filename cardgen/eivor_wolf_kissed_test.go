package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func eivorWolfKissedCard() *ScryfallCard {
	power, toughness := "3", "3"
	return &ScryfallCard{
		Name:     "Eivor, Wolf-Kissed",
		Layout:   "normal",
		ManaCost: "{3}{R}{G}{W}",
		TypeLine: "Legendary Creature — Human Assassin Warrior",
		OracleText: "Trample, haste\n" +
			"Whenever Eivor deals combat damage to a player, you mill that many cards. You may put a Saga card and/or a land card from among them onto the battlefield.",
		Power:     &power,
		Toughness: &toughness,
	}
}

func TestGenerateExecutableCardSourceEivorWolfKissed(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(eivorWolfKissedCard(), "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.TrampleStaticBody",
		"game.HasteStaticBody",
		"Event:               game.EventDamageDealt",
		"RequireCombatDamage: true",
		"DamageRecipient:     game.DamageRecipientPlayer",
		"game.Mill{",
		"Kind:       game.DynamicAmountEventDamage",
		"PublishLinked: game.LinkedKey(\"milled-cards\")",
		"Filter:     game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Saga\")}}",
		"Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}}",
		"FromLinked: game.LinkedKey(\"milled-cards\")",
		"Zone: zone.Battlefield",
		"Optional: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerEivorMillsThatManyThenOptionalOneOfEachToBattlefield asserts the
// lowered combat-damage trigger mills the triggering combat damage and offers
// two independent optional puts (one Saga, one land) from among the milled cards
// onto the battlefield.
func TestLowerEivorMillsThatManyThenOptionalOneOfEachToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, eivorWolfKissedCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventDamageDealt || !trigger.Trigger.Pattern.RequireCombatDamage {
		t.Fatalf("trigger pattern = %#v", trigger.Trigger.Pattern)
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(trigger.Content.Modes))
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want 3", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want Mill", sequence[0].Primitive)
	}
	if mill.PublishLinked != milledCardsLinkKey {
		t.Fatalf("mill PublishLinked = %q", mill.PublishLinked)
	}
	dynamic := mill.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
		t.Fatalf("mill amount = %#v, want dynamic event damage", mill.Amount)
	}

	wantSelections := []game.Selection{
		{SubtypesAny: []types.Sub{types.Sub("Saga")}},
		{RequiredTypes: []types.Card{types.Land}},
	}
	for i, want := range wantSelections {
		put, ok := sequence[i+1].Primitive.(game.ChooseFromZone)
		if !ok {
			t.Fatalf("sequence[%d] = %#v, want ChooseFromZone", i+1, sequence[i+1].Primitive)
		}
		if !sequence[i+1].Optional {
			t.Fatalf("put %d is not optional", i)
		}
		if put.Destination.Zone != zone.Battlefield {
			t.Fatalf("put %d destination = %v, want Battlefield", i, put.Destination.Zone)
		}
		if put.Riders.FromLinked != milledCardsLinkKey {
			t.Fatalf("put %d FromLinked = %q", i, put.Riders.FromLinked)
		}
		if len(put.Filter.SubtypesAny) != len(want.SubtypesAny) ||
			len(put.Filter.RequiredTypes) != len(want.RequiredTypes) {
			t.Fatalf("put %d selection = %#v, want %#v", i, put.Filter, want)
		}
	}
}

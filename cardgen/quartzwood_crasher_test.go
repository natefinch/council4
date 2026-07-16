package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerQuartzwoodCrasher proves the full cardgen pipeline lowers Quartzwood
// Crasher into a combat-damage batch trigger (OneOrMorePerDamagedPlayer, combat
// damage to a player, trample-restricted source) whose create-token effect sizes
// an X/X green Dinosaur Beast with trample by the batch's combat damage.
func TestLowerQuartzwoodCrasher(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Quartzwood Crasher",
		Layout:    "normal",
		ManaCost:  "{4}{R}{G}",
		TypeLine:  "Creature — Dinosaur Beast",
		Colors:    []string{"R", "G"},
		Power:     new("5"),
		Toughness: new("4"),
		OracleText: "Trample\n" +
			"Whenever one or more creatures you control with trample deal combat damage to a player, " +
			"create an X/X green Dinosaur Beast creature token with trample, " +
			"where X is the amount of damage those creatures dealt to that player.",
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventDamageDealt {
		t.Errorf("event = %v, want EventDamageDealt", pattern.Event)
	}
	if !pattern.OneOrMore {
		t.Error("OneOrMore = false, want true")
	}
	if !pattern.OneOrMorePerDamagedPlayer {
		t.Error("OneOrMorePerDamagedPlayer = false, want true (per-player combat-damage coalescing)")
	}
	if !pattern.RequireCombatDamage {
		t.Error("RequireCombatDamage = false, want true")
	}
	if pattern.DamageRecipient != game.DamageRecipientPlayer {
		t.Errorf("damage recipient = %v, want DamageRecipientPlayer", pattern.DamageRecipient)
	}
	if pattern.DamageSourceSelection.Keyword != game.Trample {
		t.Errorf("damage source keyword = %v, want Trample", pattern.DamageSourceSelection.Keyword)
	}

	create, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("trigger effect = %#v, want CreateToken", face.TriggeredAbilities[0].Content)
	}
	if create.Amount.Value() != 1 || !create.Power.Exists || !create.Toughness.Exists {
		t.Fatalf("create token = %#v, want one dynamically sized token", create)
	}
	for name, size := range map[string]game.Quantity{"power": create.Power.Val, "toughness": create.Toughness.Val} {
		dynamic := size.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountTriggeringEventTotalCombatDamage {
			t.Fatalf("%s = %#v, want triggering-event total combat damage", name, size)
		}
	}

	def, ok := create.Source.TokenDefRef()
	if !ok ||
		def.Name != "Dinosaur Beast" ||
		!slices.Equal(def.Colors, []color.Color{color.Green}) ||
		!slices.Equal(def.Types, []types.Card{types.Creature}) ||
		!slices.Equal(def.Subtypes, []types.Sub{types.Dinosaur, types.Beast}) ||
		len(def.StaticAbilities) != 1 ||
		def.Power.Exists ||
		def.Toughness.Exists {
		t.Fatalf("token definition = %#v", def)
	}
}

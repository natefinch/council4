package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerWildGrowthTappedForMana verifies the "Whenever enchanted land is
// tapped for mana, its controller adds an additional {G}" aura lowers to a
// permanent-tapped trigger restricted to tapped-for-mana provenance whose body
// adds the mana to the triggering land's controller.
func TestLowerWildGrowthTappedForMana(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Growth",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant land\nWhenever enchanted land is tapped for mana, its controller adds an additional {G}.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventPermanentTapped ||
		!ability.Trigger.Pattern.RequireTappedForMana {
		t.Fatalf("trigger = %#v, want tapped-for-mana permanent-tapped pattern", ability.Trigger.Pattern)
	}
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want single add-mana instruction", ability.Content)
	}
	addMana, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddMana", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if addMana.ManaColor != mana.G || addMana.Amount.Value() != 1 {
		t.Fatalf("add mana = %#v, want one green", addMana)
	}
	if !addMana.Player.Exists ||
		addMana.Player.Val.Kind() != game.PlayerReferenceObjectController {
		t.Fatalf("add mana recipient = %#v, want triggering object's controller", addMana.Player)
	}
	object, ok := addMana.Player.Val.Object()
	if !ok || object.Kind() != game.ObjectReferenceEventPermanent {
		t.Fatalf("add mana recipient object = %#v, want event permanent", object)
	}
}

// TestLowerForbiddenOrchardSelfTappedForMana verifies the active-voice self form
// ("Whenever you tap this land for mana") lowers to a tapped-for-mana
// permanent-tapped trigger whose body creates a token under a targeted opponent's
// control (Forbidden Orchard).
func TestLowerForbiddenOrchardSelfTappedForMana(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Orchard",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add one mana of any color.\nWhenever you tap this land for mana, target opponent creates a 1/1 colorless Spirit creature token.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventPermanentTapped ||
		!ability.Trigger.Pattern.RequireTappedForMana {
		t.Fatalf("trigger = %#v, want tapped-for-mana permanent-tapped pattern", ability.Trigger.Pattern)
	}
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want single create-token instruction", ability.Content)
	}
	create, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if !create.Recipient.Exists ||
		create.Recipient.Val.Kind() != game.PlayerReferenceTargetPlayer {
		t.Fatalf("recipient = %#v, want targeted-player reference", create.Recipient)
	}
	targets := ability.Content.Modes[0].Targets
	if len(targets) != 1 || targets[0].Predicate.Player != game.PlayerOpponent {
		t.Fatalf("targets = %#v, want one opponent target", targets)
	}
}

// ("Whenever a Forest is tapped for mana") also lowers to the tapped-for-mana
// permanent-tapped trigger.
func TestLowerTappedForManaSelectionSubject(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bloom",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a Forest is tapped for mana, its controller adds an additional {G}.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	if !face.TriggeredAbilities[0].Trigger.Pattern.RequireTappedForMana {
		t.Fatalf("trigger = %#v, want tapped-for-mana provenance", face.TriggeredAbilities[0].Trigger.Pattern)
	}
}

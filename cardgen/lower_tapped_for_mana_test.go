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
	if len(targets) != 1 || targets[0].Selection.Val.Player != game.PlayerOpponent {
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

// TestLowerTappedForManaWhileMonarchInterveningCondition proves a "while you're
// the monarch" clause inside a tapped-for-mana trigger event (Regal Behemoth)
// lowers to the same intervening ControllerIsMonarch condition an "if you're the
// monarch" clause would, gating the trigger on the monarch designation.
func TestLowerTappedForManaWhileMonarchInterveningCondition(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Behemoth",
		Layout:   "normal",
		TypeLine: "Creature — Dinosaur",
		ManaCost: "{4}{G}{G}",
		OracleText: "Trample\n" +
			"Whenever you tap a land for mana while you're the monarch, add an additional one mana of any color.",
		Power:     new("5"),
		Toughness: new("5"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if !trigger.Pattern.RequireTappedForMana {
		t.Fatalf("trigger = %#v, want tapped-for-mana provenance", trigger.Pattern)
	}
	if !trigger.InterveningCondition.Exists || !trigger.InterveningCondition.Val.ControllerIsMonarch {
		t.Fatalf("intervening condition = %#v, want ControllerIsMonarch", trigger.InterveningCondition)
	}
}

// TestLowerFertileGroundAnyColorTappedForMana verifies the "its controller adds
// an additional one mana of any color" aura body (Fertile Ground, Verdant Haven,
// Buried in the Garden) lowers to a resolution-time color choice made by the
// triggering land's controller followed by an AddMana that spends that choice and
// routes the mana to the same player.
func TestLowerFertileGroundAnyColorTappedForMana(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fertile",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant land\nWhenever enchanted land is tapped for mana, its controller adds an additional one mana of any color.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventPermanentTapped ||
		!ability.Trigger.Pattern.RequireTappedForMana {
		t.Fatalf("trigger = %#v, want tapped-for-mana permanent-tapped pattern", ability.Trigger.Pattern)
	}
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 2 {
		t.Fatalf("content = %#v, want choose-then-add-mana sequence", ability.Content)
	}
	seq := ability.Content.Modes[0].Sequence

	choose, ok := seq[0].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Choose", seq[0].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("choice kind = %v, want ResolutionChoiceMana", choose.Choice.Kind)
	}
	wantColors := []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G}
	if len(choose.Choice.Colors) != len(wantColors) {
		t.Fatalf("choice colors = %#v, want the five basic colors", choose.Choice.Colors)
	}
	for i, c := range wantColors {
		if choose.Choice.Colors[i] != c {
			t.Fatalf("choice colors = %#v, want the five basic colors in WUBRG order", choose.Choice.Colors)
		}
	}
	if choose.Choice.PlayerReference == nil ||
		choose.Choice.PlayerReference.Kind() != game.PlayerReferenceObjectController {
		t.Fatalf("chooser = %#v, want triggering object's controller", choose.Choice.PlayerReference)
	}
	if choose.PublishChoice == "" {
		t.Fatalf("choose publishes no key: %#v", choose)
	}

	addMana, ok := seq[1].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddMana", seq[1].Primitive)
	}
	if addMana.Amount.Value() != 1 {
		t.Fatalf("add mana amount = %#v, want one", addMana.Amount)
	}
	if addMana.ChoiceFrom != choose.PublishChoice {
		t.Fatalf("add mana ChoiceFrom = %q, want the published choice %q", addMana.ChoiceFrom, choose.PublishChoice)
	}
	if addMana.ManaColor != "" {
		t.Fatalf("add mana color = %q, want empty (color comes from the choice)", addMana.ManaColor)
	}
	if !addMana.Player.Exists ||
		addMana.Player.Val.Kind() != game.PlayerReferenceObjectController {
		t.Fatalf("add mana recipient = %#v, want triggering object's controller", addMana.Player)
	}
}

// TestLowerReferencedPlayerAnyColorTappedForMana verifies the "that player adds
// ... one mana of any color" body — whose recipient is the triggering event's
// player rather than the ability's controller — lowers to the same choose-then-
// add-mana sequence with the event player as both chooser and recipient.
func TestLowerReferencedPlayerAnyColorTappedForMana(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Communal Bloom",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a player taps a land for mana, that player adds an additional one mana of any color.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("content = %#v, want choose-then-add-mana sequence", face.TriggeredAbilities[0].Content)
	}
	choose, ok := seq[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
		t.Fatalf("first primitive = %#v, want a mana Choose", seq[0].Primitive)
	}
	if choose.Choice.PlayerReference == nil ||
		choose.Choice.PlayerReference.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("chooser = %#v, want triggering event's player", choose.Choice.PlayerReference)
	}
	addMana, ok := seq[1].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddMana", seq[1].Primitive)
	}
	if addMana.ChoiceFrom != choose.PublishChoice || addMana.Amount.Value() != 1 {
		t.Fatalf("add mana = %#v, want one mana from the published choice", addMana)
	}
	if !addMana.Player.Exists ||
		addMana.Player.Val.Kind() != game.PlayerReferenceEventPlayer {
		t.Fatalf("add mana recipient = %#v, want triggering event's player", addMana.Player)
	}
}

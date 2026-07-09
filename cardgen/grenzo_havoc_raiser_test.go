package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func grenzoHavocRaiserCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Grenzo, Havoc Raiser",
		Layout:     "normal",
		ManaCost:   "{R}{R}",
		TypeLine:   "Legendary Creature — Goblin Rogue",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever a creature you control deals combat damage to a player, choose one —\n• Goad target creature that player controls.\n• Exile the top card of that player's library. Until end of turn, you may cast that card and you may spend mana as though it were mana of any color to cast that spell.",
	}
}

// TestLowerGrenzoHavocRaiserModalCombatDamageTrigger proves Grenzo, Havoc
// Raiser's "choose one —" modal body lowers on a combat-damage trigger with both
// modes intact: mode one goads a creature the damaged (event) player controls,
// and mode two impulse-exiles the top card of that player's library as a
// cast-only, any-color window until end of turn. This exercises the modal body
// on a combat-damage trigger, the event-player-scoped goad recognizer, and the
// event-player-library / cast / any-color impulse generalization together.
func TestLowerGrenzoHavocRaiserModalCombatDamageTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, grenzoHavocRaiserCard())

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	pattern := ability.Trigger.Pattern
	if pattern.Event != game.EventDamageDealt ||
		pattern.Controller != game.TriggerControllerYou ||
		pattern.Subject != game.TriggerSubjectDamageSource ||
		!pattern.RequireCombatDamage ||
		pattern.DamageRecipient != game.DamageRecipientPlayer {
		t.Fatalf("trigger pattern = %#v, want a controller's creature dealing combat damage to a player", pattern)
	}

	content := ability.Content
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modes chosen = [%d,%d], want exactly one", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 2 {
		t.Fatalf("modes = %d, want 2 (goad / impulse)", len(content.Modes))
	}

	// Mode one: goad a creature the event player controls.
	goadMode := content.Modes[0]
	if len(goadMode.Targets) != 1 {
		t.Fatalf("goad targets = %d, want 1", len(goadMode.Targets))
	}
	sel := goadMode.Targets[0].Selection
	if !sel.Exists || !sel.Val.ControlledByEventPlayer {
		t.Fatalf("goad target selection = %#v, want ControlledByEventPlayer", goadMode.Targets[0].Selection)
	}
	if len(sel.Val.RequiredTypesAny) != 1 || sel.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("goad target types = %#v, want any creature", sel.Val.RequiredTypesAny)
	}
	if len(goadMode.Sequence) != 1 {
		t.Fatalf("goad sequence = %#v, want a single goad", goadMode.Sequence)
	}
	goad, ok := goadMode.Sequence[0].Primitive.(game.Goad)
	if !ok || goad.Object != game.TargetPermanentReference(0) {
		t.Fatalf("mode one primitive = %#v, want goad of the chosen target", goadMode.Sequence[0].Primitive)
	}

	// Mode two: impulse-exile the top card of the event player's library.
	impulseMode := content.Modes[1]
	if len(impulseMode.Sequence) != 1 {
		t.Fatalf("impulse sequence = %#v, want a single impulse exile", impulseMode.Sequence)
	}
	impulse, ok := impulseMode.Sequence[0].Primitive.(game.ImpulseExile)
	if !ok {
		t.Fatalf("mode two primitive = %#v, want game.ImpulseExile", impulseMode.Sequence[0].Primitive)
	}
	if impulse.Player != game.EventPlayerReference() {
		t.Fatalf("impulse player = %#v, want the event player's library", impulse.Player)
	}
	if impulse.Amount.Value() != 1 || impulse.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("impulse = %#v, want one card until end of turn", impulse)
	}
	if !impulse.Cast {
		t.Fatal("impulse Cast = false, want cast-only (\"you may cast that card\")")
	}
	if !impulse.SpendAnyMana {
		t.Fatal("impulse SpendAnyMana = false, want the spend-any-color rider honored")
	}
}

// TestLowerGoadThatPlayerControlsOnCombatDamageTrigger proves the narrow
// "goad target creature that player controls" recognizer fires on a bare
// combat-damage trigger (no modal wrapper), scoping the goad target to the
// damaged player's creatures via ControlledByEventPlayer.
func TestLowerGoadThatPlayerControlsOnCombatDamageTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Havoc Goader",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Goblin",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever a creature you control deals combat damage to a player, goad target creature that player controls.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	sel := mode.Targets[0].Selection
	if !sel.Exists || !sel.Val.ControlledByEventPlayer {
		t.Fatalf("target selection = %#v, want ControlledByEventPlayer", mode.Targets[0].Selection)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Goad); !ok {
		t.Fatalf("primitive = %#v, want game.Goad", mode.Sequence[0].Primitive)
	}
}

// TestLowerImpulseExileThatPlayersLibraryCastAnyColor proves the impulse
// recognizer generalization: "exile the top card of that player's library ...
// you may cast that card and you may spend mana as though it were mana of any
// color" lowers to a cast-only, any-color impulse over the event player's
// library, even outside a modal body.
func TestLowerImpulseExileThatPlayersLibraryCastAnyColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Havoc Exiler",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Goblin",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "Whenever a creature you control deals combat damage to a player, exile the top card of that player's library. Until end of turn, you may cast that card and you may spend mana as though it were mana of any color to cast that spell.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	impulse, ok := mode.Sequence[0].Primitive.(game.ImpulseExile)
	if !ok {
		t.Fatalf("primitive = %#v, want game.ImpulseExile", mode.Sequence[0].Primitive)
	}
	if impulse.Player != game.EventPlayerReference() {
		t.Fatalf("impulse player = %#v, want the event player's library", impulse.Player)
	}
	if !impulse.Cast || !impulse.SpendAnyMana || impulse.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("impulse = %#v, want cast-only any-color until end of turn", impulse)
	}
}

// TestLowerGoadForEachOpponentThatPlayerControlsFailsClosed proves the narrow
// goad recognizer refuses a leading distributive ("for each opponent, goad
// target creature that player controls", Frenzied Gorespawn's shape): the goad
// verb no longer opens the clause, so the per-opponent iteration and event
// player would be dropped. It must fail closed rather than collapse to a single
// bare goad.
func TestLowerGoadForEachOpponentThatPlayerControlsFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Distributive Goader",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Zombie",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "When this creature enters, for each opponent, goad target creature that player controls.",
	})
	if !face.empty() {
		t.Fatalf("expected no lowered ability for the distributive goad, got %#v", face)
	}
}

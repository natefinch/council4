package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerBlockThatCreatureDamage proves a self-source combat block trigger that
// deals a fixed amount of damage to the opposing combatant ("Whenever this
// creature blocks or becomes blocked by a creature, this creature deals 3 damage
// to that creature.", Inferno Elemental) lowers onto a Damage primitive whose
// recipient is the event's related permanent and whose source is the source
// permanent.
func TestLowerBlockThatCreatureDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Inferno",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "Whenever this creature blocks or becomes blocked by a creature, this creature deals 3 damage to that creature.",
		Power:      new("3"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount != game.Fixed(3) {
		t.Fatalf("damage amount = %#v, want 3", damage.Amount)
	}
	if damage.Recipient != game.ObjectDamageRecipient(game.EventRelatedPermanentReference()) {
		t.Fatalf("damage recipient = %#v, want event related permanent", damage.Recipient)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage source = %#v, want source permanent", damage.DamageSource)
	}
}

// TestLowerBecomesBlockedThatCreatureDestroy proves a self-source became-blocked
// trigger that destroys the opposing combatant ("Whenever this creature becomes
// blocked by a creature, destroy that creature.", Sylvan Basilisk) lowers onto a
// Destroy primitive whose object is the event's related permanent, not the
// source permanent the event names as its primary subject.
func TestLowerBecomesBlockedThatCreatureDestroy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Basilisk",
		Layout:     "normal",
		TypeLine:   "Creature — Basilisk",
		OracleText: "Whenever this creature becomes blocked by a creature, destroy that creature.",
		Power:      new("2"),
		Toughness:  new("4"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok {
		t.Fatalf("primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	if destroy.Object != game.EventRelatedPermanentReference() {
		t.Fatalf("destroy object = %#v, want event related permanent", destroy.Object)
	}
}

// TestLowerBecomesBlockedThatCreatureCounter proves a self-source became-blocked
// trigger that puts a -1/-1 counter on the opposing combatant ("Whenever this
// creature becomes blocked by a creature, put a -1/-1 counter on that creature.",
// Quagmire Lamprey) lowers onto an AddCounter primitive whose object is the
// event's related permanent.
func TestLowerBecomesBlockedThatCreatureCounter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Lamprey",
		Layout:     "normal",
		TypeLine:   "Creature — Fish",
		OracleText: "Whenever this creature becomes blocked by a creature, put a -1/-1 counter on that creature.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
	}
	if add.Object != game.EventRelatedPermanentReference() {
		t.Fatalf("counter object = %#v, want event related permanent", add.Object)
	}
	if add.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("counter kind = %#v, want -1/-1", add.CounterKind)
	}
}

// TestLowerCreatureBlocksThatCreatureKeepsEventPermanent proves the non-self
// "Whenever a creature blocks" trigger still binds "that creature" to the event
// permanent (the blocking creature itself), so its controller-recipient damage
// is unaffected by the related-permanent binding scoped to self-source triggers.
func TestLowerCreatureBlocksThatCreatureKeepsEventPermanent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Strain",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature blocks, this enchantment deals 1 damage to that creature's controller.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	recipient, ok := damage.Recipient.PlayerReference()
	if !ok {
		t.Fatalf("recipient = %#v, want a player reference", damage.Recipient)
	}
	if recipient != game.ObjectControllerReference(game.EventPermanentReference()) {
		t.Fatalf("recipient = %#v, want controller of event permanent", recipient)
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAttacksItsAttackingDamage proves a self-source attack trigger that
// deals a fixed amount of damage to the defending player ("Whenever this
// creature attacks, it deals 1 damage to the player or planeswalker it's
// attacking.", Scorch Spitter) lowers onto a Damage primitive whose recipient is
// the triggering attack's defending player and whose source is the attacking
// event permanent.
func TestLowerAttacksItsAttackingDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spitter",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "Whenever this creature attacks, it deals 1 damage to the player or planeswalker it's attacking.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventAttackerDeclared {
		t.Fatalf("trigger event = %#v, want EventAttackerDeclared", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount != game.Fixed(1) {
		t.Fatalf("damage amount = %#v, want 1", damage.Amount)
	}
	if damage.Recipient != game.PlayerDamageRecipient(game.DefendingPlayerReference()) {
		t.Fatalf("damage recipient = %#v, want defending player", damage.Recipient)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("damage source = %#v, want event permanent", damage.DamageSource)
	}
}

// TestLowerControlledAttacksThatCreatureAttackingDamage proves a
// controller-scoped attack trigger that deals a fixed amount of damage to the
// defending player ("Whenever a creature you control with power 1 or less
// attacks, this enchantment deals 1 damage to the player or planeswalker that
// creature is attacking.", Cavalcade of Calamity) lowers onto a Damage primitive
// whose recipient is the triggering attack's defending player and whose source
// is the ability's own source permanent.
func TestLowerControlledAttacksThatCreatureAttackingDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Cavalcade",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control with power 1 or less attacks, this enchantment deals 1 damage to the player or planeswalker that creature is attacking.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventAttackerDeclared {
		t.Fatalf("trigger event = %#v, want EventAttackerDeclared", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount != game.Fixed(1) {
		t.Fatalf("damage amount = %#v, want 1", damage.Amount)
	}
	if damage.Recipient != game.PlayerDamageRecipient(game.DefendingPlayerReference()) {
		t.Fatalf("damage recipient = %#v, want defending player", damage.Recipient)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage source = %#v, want source permanent", damage.DamageSource)
	}
}

// TestLowerBecomesBlockedAttackingDamage proves a self-source became-blocked
// trigger that deals a fixed amount of damage to the defending player ("Whenever
// this creature becomes blocked, it deals 1 damage to the player or planeswalker
// it's attacking.", Rakdos Roustabout) lowers onto a Damage primitive whose
// recipient is the triggering attack's defending player, resolved from the
// became-blocked event.
func TestLowerBecomesBlockedAttackingDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Roustabout",
		Layout:     "normal",
		TypeLine:   "Creature — Ogre Warrior",
		OracleText: "Whenever this creature becomes blocked, it deals 1 damage to the player or planeswalker it's attacking.",
		Power:      new("3"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventAttackerBecameBlocked {
		t.Fatalf("trigger event = %#v, want EventAttackerBecameBlocked", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
	}
	if damage.Recipient != game.PlayerDamageRecipient(game.DefendingPlayerReference()) {
		t.Fatalf("damage recipient = %#v, want defending player", damage.Recipient)
	}
}

// TestLowerDynamicAttackingDamageUnsupported proves a dynamic-amount attacked
// defender damage effect whose amount counts a permanent kind ("Whenever this
// creature attacks, it deals damage to the player or planeswalker it's attacking
// equal to the number of artifacts you control.", Fathom Fleet Swordjack) fails
// closed: the recipient-before-amount word order keeps the effect inexact, so it
// is not lowered onto a Damage primitive.
func TestLowerDynamicAttackingDamageUnsupported(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Swordjack",
		Layout:     "normal",
		TypeLine:   "Creature — Orc Pirate",
		OracleText: "Whenever this creature attacks, it deals damage to the player or planeswalker it's attacking equal to the number of artifacts you control.",
		Power:      new("4"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 0 {
		t.Fatalf("got %d triggered abilities, want 0 (unsupported)", len(face.TriggeredAbilities))
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestAttachPrimitiveAttachesSourceEquipmentToTarget proves that resolving
// Attach{Attachment: SourcePermanentReference(), Target: TargetPermanentReference(0)}
// attaches the source Equipment to the chosen creature, as the Mithril Coat
// enters-the-battlefield auto-attach trigger relies on.
func TestAttachPrimitiveAttachesSourceEquipmentToTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addEquipmentPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   equipment.ObjectID,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.Attach{
		Attachment: game.SourcePermanentReference(),
		Target:     game.TargetPermanentReference(0),
	}, &TurnLog{})

	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("equipment AttachedTo = %v, want %v", equipment.AttachedTo, creature.ObjectID)
	}
	if !permanentIDsContain(creature.Attachments, equipment.ObjectID) {
		t.Fatalf("creature Attachments = %v, want to contain equipment %v", creature.Attachments, equipment.ObjectID)
	}
}

// TestAttachPrimitiveReattachesFromPriorTarget proves that attaching an already
// equipped Equipment to a new creature detaches it from the old one first.
func TestAttachPrimitiveReattachesFromPriorTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	equipment := addEquipmentPermanent(g, game.Player1)
	oldCreature := addCombatCreaturePermanent(g, game.Player1)
	newCreature := addCombatCreaturePermanent(g, game.Player1)
	if !attachPermanent(g, equipment, oldCreature) {
		t.Fatal("attachPermanent() = false, want true")
	}

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   equipment.ObjectID,
		Targets:    []game.Target{game.PermanentTarget(newCreature.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.Attach{
		Attachment: game.SourcePermanentReference(),
		Target:     game.TargetPermanentReference(0),
	}, &TurnLog{})

	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != newCreature.ObjectID {
		t.Fatalf("equipment AttachedTo = %v, want %v", equipment.AttachedTo, newCreature.ObjectID)
	}
	if permanentIDsContain(oldCreature.Attachments, equipment.ObjectID) {
		t.Fatalf("old creature still lists equipment in Attachments = %v", oldCreature.Attachments)
	}
}

// TestAttachPrimitiveAttachesEventEquipmentToTarget proves that resolving
// Attach{Attachment: EventPermanentReference(), Target: TargetPermanentReference(0)}
// attaches the triggering event permanent — the Equipment that just entered —
// to the chosen creature, as Hammer of Nazahn's "Whenever ~ or another Equipment
// you control enters, you may attach that Equipment to target creature you
// control." auto-attach trigger relies on. The entering Equipment is a different
// permanent than the trigger source, so the attachment must follow the event
// permanent rather than the source.
func TestAttachPrimitiveAttachesEventEquipmentToTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addEquipmentPermanent(g, game.Player1)
	enteringEquipment := addEquipmentPermanent(g, game.Player1)
	creature := addCombatCreaturePermanent(g, game.Player1)

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: enteringEquipment.ObjectID,
		},
		Targets: []game.Target{game.PermanentTarget(creature.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.Attach{
		Attachment: game.EventPermanentReference(),
		Target:     game.TargetPermanentReference(0),
	}, &TurnLog{})

	if !enteringEquipment.AttachedTo.Exists || enteringEquipment.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("entering equipment AttachedTo = %v, want %v", enteringEquipment.AttachedTo, creature.ObjectID)
	}
	if source.AttachedTo.Exists {
		t.Fatalf("source equipment should not be attached, AttachedTo = %v", source.AttachedTo.Val)
	}
	if !permanentIDsContain(creature.Attachments, enteringEquipment.ObjectID) {
		t.Fatalf("creature Attachments = %v, want to contain entering equipment %v", creature.Attachments, enteringEquipment.ObjectID)
	}
}

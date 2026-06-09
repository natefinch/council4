package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestEquippedCreatureSelectorGrantsKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types: []types.Card{types.Creature}},
	})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect([]game.ContinuousEffect{{
		Layer:       game.LayerAbility,
		Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
		AddKeywords: []game.Keyword{game.Deathtouch, game.Lifelink},
	}}))

	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	if !hasKeyword(g, creature, game.Deathtouch) || !hasKeyword(g, creature, game.Lifelink) {
		t.Fatal("equipped creature did not gain deathtouch and lifelink")
	}
	detachPermanent(g, equipment)
	if hasKeyword(g, creature, game.Deathtouch) || hasKeyword(g, creature, game.Lifelink) {
		t.Fatal("unattached creature retained equipment-granted keywords")
	}
}

func TestEquippedCreatureSelectorDynamicOpponentCountPT(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect([]game.ContinuousEffect{{
		Layer: game.LayerPowerToughnessModify,
		Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
		PowerDeltaDynamic: opt.Val(game.DynamicAmount{
			Kind: game.DynamicAmountOpponentCount,
		}),
		ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
			Kind: game.DynamicAmountOpponentCount,
		}),
	}}))
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("equipped creature power = %d, want base 1 + 3 opponents", got)
	}
	g.Players[game.Player4].Eliminated = true
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("equipped creature power = %d, want base 1 + 2 opponents after elimination", got)
	}
}

func TestAttachedPermanentTriggerFilterUsesLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types: []types.Card{types.Creature}},
	})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	snapshot := snapshotPermanent(g, creature, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	detachPermanent(g, equipment)

	if !triggerSourceAttachedPermanentMatches(g, equipment, game.Event{
		Kind:            game.EventDamageDealt,
		PermanentID:     creature.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
	}) {
		t.Fatal("attached-permanent trigger filter did not use damaged permanent LKI attachments")
	}
}

func TestEventDamageDynamicAmountAndAttachedDamageSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types: []types.Card{types.Creature}},
	})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        equipment.ObjectID,
		SourceCardID:    equipment.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:            game.EventDamageDealt,
			PermanentID:     creature.ObjectID,
			Amount:          4,
			DamageRecipient: game.DamageRecipientPermanent,
		},
		Targets: []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.SourceAttachedPermanentReference()),
		Amount:       game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
	}, &log)

	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("Player2 life = %d, want 36", got)
	}
}

func TestObjectPowerDynamicAmountUsesAttachedPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	})
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     equipment.ObjectID,
		SourceCardID: equipment.CardInstanceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.SourceAttachedPermanentReference()),
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:   game.DynamicAmountObjectPower,
			Object: game.SourceAttachedPermanentReference(),
		}),
	}, &log)

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("Player2 life = %d, want 37", got)
	}
}

func TestAllCreaturesExceptTargetAndOpponentPlayerSelector(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Igniter",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{game.DeathtouchStaticBody}},
	})
	other := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Other Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5})},
	})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(source.ObjectID)},
	}
	baseDamage := game.Damage{
		DamageSource: opt.Val(game.TargetPermanentReference(0)),
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:   game.DynamicAmountTargetPower,
			Object: game.TargetPermanentReference(0),
		}),
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.GroupDamageRecipient(game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.TargetPermanentReference(0))),
		DamageSource: baseDamage.DamageSource,
		Amount:       baseDamage.Amount,
	}, &log)
	if source.MarkedDamage != 0 {
		t.Fatal("source creature was damaged by all-creatures-except-target selector")
	}
	if other.MarkedDamage != 3 || !other.MarkedDeathtouchDamage {
		t.Fatalf("other creature damage/deathtouch = %d/%v, want 3/true", other.MarkedDamage, other.MarkedDeathtouchDamage)
	}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
		DamageSource: baseDamage.DamageSource,
		Amount:       baseDamage.Amount,
	}, &log)
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 37 {
			t.Fatalf("Player%d life = %d, want 37", playerID+1, got)
		}
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("Player1 life = %d, want unchanged", got)
	}
}

func equipmentWithStaticEffect(effects []game.ContinuousEffect) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Equipment",
		Types:           []types.Card{types.Artifact},
		Subtypes:        []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: effects}}},
	}
}

func damageRecipientCreatureCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}}
}

// TestDamageEquippedCreatureUnsetSourceMatchesNothing pins the legacy behavior
// that a Damage primitive whose recipient is EquippedCreature deals nothing when
// DamageSource is unset, because the source-dependent selector resolves relative
// to the (nil) DamageSource rather than the stack object's source permanent.
func TestDamageEquippedCreatureUnsetSourceMatchesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("Bearer"))
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent failed")
	}
	obj := &game.StackObject{Controller: game.Player1, SourceID: equipment.ObjectID}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient: game.GroupDamageRecipient(game.AttachedObjectGroup(game.SourcePermanentReference())),
		Amount:    game.Fixed(2),
	}, &log)

	if creature.MarkedDamage != 0 {
		t.Fatalf("equipped creature damage = %d, want 0 with unset DamageSource", creature.MarkedDamage)
	}
}

// TestDamageEquippedCreatureExplicitSourceResolvesRelativeToObject proves the
// EquippedCreature recipient resolves relative to the resolved DamageSource
// object, not the stack object's source permanent.
func TestDamageEquippedCreatureExplicitSourceResolvesRelativeToObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureA := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("CreatureA"))
	creatureB := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("CreatureB"))
	equipA := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	equipB := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect(nil))
	if !attachPermanent(g, equipA, creatureA) || !attachPermanent(g, equipB, creatureB) {
		t.Fatal("attachPermanent failed")
	}
	obj := &game.StackObject{
		Controller: game.Player1,
		SourceID:   equipA.ObjectID,
		Targets:    []game.Target{game.PermanentTarget(equipB.ObjectID)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.GroupDamageRecipient(game.AttachedObjectGroup(game.SourcePermanentReference())),
		DamageSource: opt.Val(game.TargetPermanentReference(0)),
		Amount:       game.Fixed(2),
	}, &log)

	if creatureB.MarkedDamage != 2 {
		t.Fatalf("creatureB equipped by DamageSource damage = %d, want 2", creatureB.MarkedDamage)
	}
	if creatureA.MarkedDamage != 0 {
		t.Fatalf("creatureA equipped by stack source damage = %d, want 0", creatureA.MarkedDamage)
	}
}

// TestDamageOtherCreaturesYouControlUnsetSourceMatchesNothing pins the legacy
// behavior that an OtherCreaturesYouControl recipient deals nothing when
// DamageSource is unset: ExcludeSource with no source permanent matches nothing.
func TestDamageOtherCreaturesYouControlUnsetSourceMatchesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("Source"))
	other := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("Other"))
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true})),
		Amount:    game.Fixed(2),
	}, &log)

	if other.MarkedDamage != 0 || source.MarkedDamage != 0 {
		t.Fatalf("other/source damage = %d/%d, want 0/0 with unset DamageSource", other.MarkedDamage, source.MarkedDamage)
	}
}

// TestDamageOtherCreaturesYouControlExplicitSourceResolvesRelativeToObject
// proves the OtherCreaturesYouControl recipient excludes the resolved
// DamageSource object, not the stack object's source permanent.
func TestDamageOtherCreaturesYouControlExplicitSourceResolvesRelativeToObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	stackSource := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("StackSource"))
	damageSource := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("DamageSource"))
	bystander := addCombatPermanent(g, game.Player1, damageRecipientCreatureCard("Bystander"))
	obj := &game.StackObject{
		Controller: game.Player1,
		SourceID:   stackSource.ObjectID,
		Targets:    []game.Target{game.PermanentTarget(damageSource.ObjectID)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true})),
		DamageSource: opt.Val(game.TargetPermanentReference(0)),
		Amount:       game.Fixed(2),
	}, &log)

	if damageSource.MarkedDamage != 0 {
		t.Fatalf("DamageSource creature damage = %d, want 0 (excluded as the source)", damageSource.MarkedDamage)
	}
	if stackSource.MarkedDamage != 2 {
		t.Fatalf("stack source creature damage = %d, want 2 (not the excluded source)", stackSource.MarkedDamage)
	}
	if bystander.MarkedDamage != 2 {
		t.Fatalf("bystander creature damage = %d, want 2", bystander.MarkedDamage)
	}
}

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
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect([]game.Effect{{
		Type:        game.EffectApplyContinuous,
		TargetIndex: game.TargetIndexSourcePermanent,
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			Selector:    game.EffectSelectorEquippedCreature,
			AddKeywords: []game.Keyword{game.Deathtouch, game.Lifelink},
		}},
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
	equipment := addCombatPermanent(g, game.Player1, equipmentWithStaticEffect([]game.Effect{{
		Type:        game.EffectModifyPT,
		TargetIndex: game.TargetIndexSourcePermanent,
		Selector:    game.EffectSelectorEquippedCreature,
		DynamicAmount: opt.Val(game.DynamicAmount{
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

	if !triggerSourceAttachedPermanentMatches(g, equipment, game.GameEvent{
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
		TriggerEvent: game.GameEvent{
			Kind:            game.EventDamageDealt,
			PermanentID:     creature.ObjectID,
			Amount:          4,
			DamageRecipient: game.DamageRecipientPermanent,
		},
		Targets: []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	engine.resolveEffect(g, obj, &game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 0,
		DamageSource: opt.Val(game.ObjectReference{
			Kind:        game.ObjectReferenceAttachedPermanent,
			TargetIndex: game.TargetIndexSourcePermanent,
		}),
		DynamicAmount: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
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

	engine.resolveEffect(g, obj, &game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 0,
		DamageSource: opt.Val(game.ObjectReference{
			Kind:        game.ObjectReferenceAttachedPermanent,
			TargetIndex: game.TargetIndexSourcePermanent,
		}),
		DynamicAmount: opt.Val(game.DynamicAmount{
			Kind: game.DynamicAmountObjectPower,
			Object: game.ObjectReference{
				Kind:        game.ObjectReferenceAttachedPermanent,
				TargetIndex: game.TargetIndexSourcePermanent,
			},
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
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		Abilities: []game.AbilityDef{{
			Kind:             game.StaticAbility,
			KeywordAbilities: game.SimpleKeywords(game.Deathtouch),
		}}},
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
	effect := game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 0,
		DamageSource: opt.Val(game.ObjectReference{
			Kind:        game.ObjectReferenceTargetPermanent,
			TargetIndex: 0,
		}),
		DynamicAmount: opt.Val(game.DynamicAmount{
			Kind:        game.DynamicAmountTargetPower,
			TargetIndex: 0,
		}),
	}
	log := TurnLog{}

	effect.Selector = game.EffectSelectorAllCreaturesExceptTarget
	engine.resolveEffect(g, obj, &effect, &log)
	if source.MarkedDamage != 0 {
		t.Fatal("source creature was damaged by all-creatures-except-target selector")
	}
	if other.MarkedDamage != 3 || !other.MarkedDeathtouchDamage {
		t.Fatalf("other creature damage/deathtouch = %d/%v, want 3/true", other.MarkedDamage, other.MarkedDeathtouchDamage)
	}

	effect.Selector = game.EffectSelectorNone
	effect.PlayerSelector = game.PlayerSelectorOpponents
	engine.resolveEffect(g, obj, &effect, &log)
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 37 {
			t.Fatalf("Player%d life = %d, want 37", playerID+1, got)
		}
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("Player1 life = %d, want unchanged", got)
	}
}

func equipmentWithStaticEffect(effects []game.Effect) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		Abilities: []game.AbilityDef{{
			Kind:    game.StaticAbility,
			Effects: effects,
		}}},
	}
}

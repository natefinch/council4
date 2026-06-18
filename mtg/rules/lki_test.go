package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDiesTriggerUsesLastKnownEffectiveType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Animated Land",
		Types: []types.Card{types.Land}},
	})
	one := game.PT{Value: 1}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: land.ObjectID,
			Layer:            game.LayerType,
			AddTypes:         []types.Card{types.Creature},
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: land.ObjectID,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(one),
			SetToughness:     opt.Val(one),
		},
	)

	destroyPermanent(g, land.ObjectID)

	snapshot, ok := lastKnownObject(g, land.ObjectID)
	if !ok {
		t.Fatal("missing last-known snapshot for destroyed animated land")
	}
	if !slices.Contains(snapshot.Types, types.Creature) || !snapshot.Toughness.Exists || snapshot.Toughness.Val != 1 {
		t.Fatalf("snapshot = %+v, want effective creature with toughness 1", snapshot)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature dies trigger did not use last-known effective type")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want dies trigger to draw", got)
	}
}

func TestDelayedTriggerSourceIdentitySurvivesSourceZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addCombatCreaturePermanent(g, game.Player1)
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{Timing: game.DelayedAtBeginningOfNextEndStep, Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()},
	}, nil)
	movePermanentToZone(g, source, zone.Graveyard)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want delayed trigger to resolve after source moved", got)
	}
}

func TestLinkedExileReturnOnlyUsesSameSourceLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player3)
	sourceA := addCombatCreaturePermanent(g, game.Player1)
	sourceB := addCombatCreaturePermanent(g, game.Player1)
	linkID := "linked-exile"

	objA := linkedSourceObject(sourceA)
	objA.Targets = []game.Target{game.PermanentTarget(first.ObjectID)}
	resolveInstruction(engine, g, objA, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey(linkID),
	}, nil)
	objB := linkedSourceObject(sourceB)
	objB.Targets = []game.Target{game.PermanentTarget(second.ObjectID)}
	resolveInstruction(engine, g, objB, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey(linkID),
	}, nil)

	resolveInstruction(engine, g, linkedSourceObject(sourceA), game.PutOnBattlefield{
		Source: game.LinkedBattlefieldSource(game.LinkedKey(linkID)),
	}, nil)

	if !g.Players[game.Player3].Exile.Contains(second.CardInstanceID) {
		t.Fatal("return for source A removed source B's linked exiled card")
	}
	if g.Players[game.Player2].Exile.Contains(first.CardInstanceID) {
		t.Fatal("source A's linked exiled card remained in exile")
	}
	returned := permanentByCardID(g, first.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want first card returned under owner Player2 control", returned)
	}
}

func TestDelayedLinkedExileReturnUsesOwnerControlAndEmitsEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	target.Controller = game.Player1
	source := addCombatCreaturePermanent(g, game.Player1)
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}
	key := game.LinkedKey("delayed-blink")

	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: key,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource(key),
		}}}}.Ability(),
	}}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	returned := permanentByCardID(g, target.CardInstanceID)
	if returned == nil || returned.Controller != game.Player2 {
		t.Fatalf("returned permanent = %+v, want owner Player2 control", returned)
	}
	assertEvent(t, g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID &&
			event.PermanentID == returned.ObjectID &&
			event.FromZone == zone.Exile
	})
}

func TestDelayedLinkedExileReturnFailsClosedAfterCardLeavesExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatCreaturePermanent(g, game.Player1)
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}
	key := game.LinkedKey("delayed-blink")

	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: key,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource(key),
		}}}}.Ability(),
	}}, nil)
	if !moveCardBetweenZones(g, game.Player2, target.CardInstanceID, zone.Exile, zone.Library) {
		t.Fatal("failed to move linked card out of exile")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if returned := permanentByCardID(g, target.CardInstanceID); returned != nil {
		t.Fatalf("linked card returned after leaving exile: %+v", returned)
	}
	if !g.Players[game.Player2].Library.Contains(target.CardInstanceID) {
		t.Fatal("linked card did not remain in library")
	}
}

func TestDelayedLinkedModifyPTReturnsSameTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatCreaturePermanent(g, game.Player1)
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}
	key := game.LinkedKey("delayed-target")

	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(2),
		ToughnessDelta: game.Fixed(2),
		Duration:       game.DurationUntilEndOfTurn,
		PublishLinked:  key,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Bounce{
			Object: game.LinkedObjectReference(string(key)),
		}}}}.Ability(),
	}}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("linked target was not returned to its owner's hand")
	}
	if permanentByCardID(g, target.CardInstanceID) != nil {
		t.Fatal("linked target remained on the battlefield")
	}
}

func TestDelayedLinkedApplyContinuousSacrificesSameTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player1)
	source := addCombatCreaturePermanent(g, game.Player1)
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}
	key := game.LinkedKey("delayed-sacrifice")

	resolveInstruction(engine, g, obj, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			AddKeywords: []game.Keyword{game.Flying},
		}},
		Duration:      game.DurationUntilEndOfTurn,
		PublishLinked: key,
	}, nil)
	resolveInstruction(engine, g, obj, game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
			Object: game.LinkedObjectReference(string(key)),
		}}}}.Ability(),
	}}, nil)
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if permanentByCardID(g, target.CardInstanceID) != nil {
		t.Fatal("linked target was not sacrificed from the battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("sacrificed linked target did not reach its owner's graveyard")
	}
}

func linkedSourceObject(source *game.Permanent) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   source.Controller,
	}
}

func permanentByCardID(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

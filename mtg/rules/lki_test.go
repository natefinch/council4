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
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		RequirePermanentTypes: []types.Card{types.Creature},
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController}}, nil)
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

	engine.resolveEffect(g, obj, &game.Effect{
		Type:           game.EffectCreateDelayedTrigger,
		DelayedTrigger: opt.Val(game.DelayedTriggerDef{Timing: game.DelayedAtBeginningOfNextEndStep, Effects: []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: game.TargetIndexController}}}),
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
	engine.resolveEffect(g, objA, &game.Effect{
		Type:        game.EffectExile,
		TargetIndex: 0,
		LinkID:      linkID,
	}, nil)
	objB := linkedSourceObject(sourceB)
	objB.Targets = []game.Target{game.PermanentTarget(second.ObjectID)}
	engine.resolveEffect(g, objB, &game.Effect{
		Type:        game.EffectExile,
		TargetIndex: 0,
		LinkID:      linkID,
	}, nil)

	engine.resolveEffect(g, linkedSourceObject(sourceA), &game.Effect{
		Type:   game.EffectPutOnBattlefield,
		LinkID: linkID,
	}, nil)

	if !g.Players[game.Player3].Exile.Contains(second.CardInstanceID) {
		t.Fatal("return for source A removed source B's linked exiled card")
	}
	if g.Players[game.Player2].Exile.Contains(first.CardInstanceID) {
		t.Fatal("source A's linked exiled card remained in exile")
	}
	returned := permanentByCardID(g, first.CardInstanceID)
	if returned == nil || returned.Controller != game.Player1 {
		t.Fatalf("returned permanent = %+v, want first card returned under Player1 control", returned)
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

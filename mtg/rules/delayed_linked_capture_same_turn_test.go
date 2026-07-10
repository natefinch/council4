package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// delayedSacrificeCapture is the delayed "Sacrifice it at the beginning of the
// next end step" trigger the lowering emits for a linked-object disposal: it
// freezes the linked object to a concrete id at schedule time (CapturedObject)
// and sacrifices that captured object.
func delayedSacrificeCapture(key string) *game.DelayedTriggerDef {
	return &game.DelayedTriggerDef{
		Timing:         game.DelayedAtBeginningOfNextEndStep,
		CapturedObject: opt.Val(game.LinkedObjectReference(key)),
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Sacrifice{Object: game.CapturedObjectReference()}}},
		}.Ability(),
	}
}

// TestApplyContinuousDelayedSacrificeTwiceSameTurnSacrificesBoth exercises the
// Krovikan Elementalist shape ("{U}{U}: Target creature you control gains flying
// until end of turn. Sacrifice it at the beginning of the next end step."), whose
// delayed-disposal permanent is published by ApplyContinuous rather than
// CreateToken. Krovikan Elementalist has no tap in its cost, so it is trivially
// activatable twice in one turn. Each activation must sacrifice the creature it
// pumped: the publisher clears its source-scoped link key before republishing so
// the schedule-time capture binds to this activation's creature, not the first
// activation's. Without the clear both delayed sacrifices capture the first
// creature and the second leaks.
func TestApplyContinuousDelayedSacrificeTwiceSameTurnSacrificesBoth(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Krovikan Elementalist",
		Types: []types.Card{types.Creature},
	}})
	creatureA := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creatureB := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	apply := game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			AddKeywords: []game.Keyword{game.Flying},
		}},
		Duration:      game.DurationUntilEndOfTurn,
		PublishLinked: game.LinkedKey("delayed-sacrifice-1"),
	}

	activate := func(target *game.Permanent) {
		obj := &game.StackObject{
			ID:           g.IDGen.Next(),
			Kind:         game.StackActivatedAbility,
			Controller:   game.Player1,
			SourceID:     source.ObjectID,
			SourceCardID: source.CardInstanceID,
			Targets:      []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
		}
		r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
		if !handleApplyContinuous(r, apply).succeeded {
			t.Fatal("ApplyContinuous activation did not succeed")
		}
		if !scheduleDelayedTrigger(g, obj, delayedSacrificeCapture("delayed-sacrifice-1")) {
			t.Fatal("scheduleDelayedTrigger failed")
		}
	}

	activate(creatureA)
	activate(creatureB)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, ok := permanentByObjectID(g, creatureA.ObjectID); ok {
		t.Error("first pumped creature was not sacrificed at the end step")
	}
	if _, ok := permanentByObjectID(g, creatureB.ObjectID); ok {
		t.Error("second pumped creature leaked: its delayed sacrifice must capture the creature this activation acted on, not the first activation's")
	}
}

// TestModifyPTDelayedSacrificeTwiceSameTurnSacrificesBoth exercises the same
// same-turn double-activation leak for a delayed-disposal permanent published by
// ModifyPT ("Target creature gets +2/+2 until end of turn. Sacrifice it at the
// beginning of the next end step."). ModifyPT, like ApplyContinuous, must clear
// its source-scoped link key before republishing so each activation's delayed
// sacrifice binds to the creature it pumped.
func TestModifyPTDelayedSacrificeTwiceSameTurnSacrificesBoth(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Pump Source",
		Types: []types.Card{types.Creature},
	}})
	creatureA := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creatureB := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	modify := game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(2),
		ToughnessDelta: game.Fixed(2),
		Duration:       game.DurationUntilEndOfTurn,
		PublishLinked:  game.LinkedKey("delayed-sacrifice-1"),
	}

	activate := func(target *game.Permanent) {
		obj := &game.StackObject{
			ID:           g.IDGen.Next(),
			Kind:         game.StackActivatedAbility,
			Controller:   game.Player1,
			SourceID:     source.ObjectID,
			SourceCardID: source.CardInstanceID,
			Targets:      []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
		}
		r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
		if !handleModifyPT(r, modify).succeeded {
			t.Fatal("ModifyPT activation did not succeed")
		}
		if !scheduleDelayedTrigger(g, obj, delayedSacrificeCapture("delayed-sacrifice-1")) {
			t.Fatal("scheduleDelayedTrigger failed")
		}
	}

	activate(creatureA)
	activate(creatureB)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, ok := permanentByObjectID(g, creatureA.ObjectID); ok {
		t.Error("first pumped creature was not sacrificed at the end step")
	}
	if _, ok := permanentByObjectID(g, creatureB.ObjectID); ok {
		t.Error("second pumped creature leaked: its delayed sacrifice must capture the creature this activation acted on, not the first activation's")
	}
}

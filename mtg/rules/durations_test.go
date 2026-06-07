package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestUntilEndOfTurnPTModifierUsesRuntimeContinuousEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		TargetIndex:    0,
		PowerDelta:     game.Fixed(3),
		ToughnessDelta: game.Fixed(3),
		Duration:       game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(g.ContinuousEffects))
	}
	if creature.TemporaryPowerModifier != 0 || creature.TemporaryToughnessModifier != 0 {
		t.Fatalf("legacy temporary modifiers = +%d/+%d, want 0/0", creature.TemporaryPowerModifier, creature.TemporaryToughnessModifier)
	}
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power = %d, want 5", got)
	}
}

func TestUntilEndOfTurnPTModifierSnapshotsDynamicX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		TargetIndex: 0,
		Duration:    game.DurationUntilEndOfTurn,
		PowerDelta: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountX,
		}),
		ToughnessDelta: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountX,
		}),
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty")
	}
	obj.XValue = 3

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(g.ContinuousEffects))
	}
	effect := g.ContinuousEffects[0]
	if effect.PowerDelta != 3 || effect.ToughnessDelta != 3 {
		t.Fatalf("continuous effect deltas = +%d/+%d, want +3/+3", effect.PowerDelta, effect.ToughnessDelta)
	}
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power = %d, want 5", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 5 {
		t.Fatalf("effective toughness = %d ok=%v, want 5 true", got, ok)
	}
}

func TestUntilEndOfTurnPTModifierSnapshotsDynamicTargetPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		TargetIndex: 0,
		Duration:    game.DurationUntilEndOfTurn,
		PowerDelta: game.Dynamic(game.DynamicAmount{
			Kind:        game.DynamicAmountTargetPower,
			TargetIndex: 0,
		}),
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(g.ContinuousEffects))
	}
	if got := g.ContinuousEffects[0].PowerDelta; got != 4 {
		t.Fatalf("snapshotted power delta = %d, want current power 4", got)
	}
	if got := effectivePower(g, creature); got != 8 {
		t.Fatalf("effective power = %d, want doubled current power 8", got)
	}
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	if got := effectivePower(g, creature); got != 9 {
		t.Fatalf("effective power after later counter = %d, want snapshotted delta plus new counter = 9", got)
	}
}

func TestCleanupExpiresTemporaryContinuousEffectsButKeepsCountersAndStaticEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addAnthemPermanent(g, game.Player1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		TargetIndex:    0,
		PowerDelta:     game.Fixed(3),
		ToughnessDelta: game.Fixed(3),
		Duration:       game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	expireCleanupDurations(g)

	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects after cleanup = %d, want 0", len(g.ContinuousEffects))
	}
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power after cleanup = %d, want base 2 + counter 1 + anthem 1 = 4", got)
	}
}

func TestUntilYourNextTurnDurationExpiresAtThatPlayersNextTurnStart(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerPowerToughnessModify,
		PowerDelta:       2,
		ToughnessDelta:   2,
		Duration:         game.DurationUntilYourNextTurn,
		CreatedTurn:      1,
		ExpiresFor:       game.Player1,
	})

	g.Turn.TurnNumber = 2
	g.Turn.ActivePlayer = game.Player2
	expireTurnStartDurations(g)
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power on another player's turn = %d, want 4", got)
	}

	g.Turn.ActivePlayer = game.Player1
	expireTurnStartDurations(g)
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power on controller's next turn = %d, want expired base 2", got)
	}
	expireTurnStartDurations(g)
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after second expiry = %d, want idempotent base 2", got)
	}
}

func TestThisTurnDurationExpiresAtCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerPowerToughnessModify,
		PowerDelta:       2,
		ToughnessDelta:   2,
		Duration:         game.DurationThisTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})

	expireCleanupDurations(g)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after this-turn expiry = %d, want base 2", got)
	}
}

func TestCleanupChecksSBAsAfterDurationExpiry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	creature.Counters.Add(counter.MinusOneMinusOne, 1)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerPowerToughnessModify,
		ToughnessDelta:   1,
		Duration:         game.DurationUntilEndOfTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})
	if _, ok := permanentDeathReason(g, creature); ok {
		t.Fatal("creature should survive before cleanup while temporary toughness effect applies")
	}

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("0-toughness creature survived cleanup after duration expiry")
	}
}

func TestDelayedNextEndStepTriggerFiresOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	addEffectSpellToStack(g, game.Player1, game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}}},
			}.Ability(),
		},
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after end step = %d, want 0", len(g.DelayedTriggers))
	}
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("player hand size = %d, want delayed trigger to draw 1 card", g.Players[game.Player1].Hand.Size())
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("player hand size after second ending phase = %d, want trigger to fire once", g.Players[game.Player1].Hand.Size())
	}
}

func TestDelayedNextEndStepTriggersUseAPNAPStackOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.DelayedTriggers = append(g.DelayedTriggers,
		game.DelayedTrigger{
			ID:         1,
			Controller: game.Player2,
			Timing:     game.DelayedAtBeginningOfNextEndStep,
			Ability:    game.TriggeredAbilityBody{},
		},
		game.DelayedTrigger{
			ID:         2,
			Controller: game.Player1,
			Timing:     game.DelayedAtBeginningOfNextEndStep,
			Ability:    game.TriggeredAbilityBody{},
		},
	)

	putBeginningOfEndStepDelayedTriggersOnStack(g)

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	if objects[0].Controller != game.Player1 || objects[1].Controller != game.Player2 {
		t.Fatalf("stack controllers bottom-to-top = %v/%v, want APNAP Player1/Player2", objects[0].Controller, objects[1].Controller)
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestUntilEndOfTurnPTModifierUsesRuntimeContinuousEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
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
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationUntilEndOfTurn,
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

// TestUntilEndOfTurnPTModifierSnapshotsNegativeDynamicX covers the "-X/-X"
// X-cost shrink (e.g. Death Wind), whose deltas lower to the runtime X amount
// with a -1 multiplier. The snapshot must subtract the chosen X from both power
// and toughness.
func TestUntilEndOfTurnPTModifierSnapshotsNegativeDynamicX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationUntilEndOfTurn,
		PowerDelta: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountX,
			Multiplier: -1,
		}),
		ToughnessDelta: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountX,
			Multiplier: -1,
		}),
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty")
	}
	obj.XValue = 2

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(g.ContinuousEffects))
	}
	effect := g.ContinuousEffects[0]
	if effect.PowerDelta != -2 || effect.ToughnessDelta != -2 {
		t.Fatalf("continuous effect deltas = %+d/%+d, want -2/-2", effect.PowerDelta, effect.ToughnessDelta)
	}
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want 3", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 3 {
		t.Fatalf("effective toughness = %d ok=%v, want 3 true", got, ok)
	}
}

func TestUntilEndOfTurnPTModifierSnapshotsDynamicTargetPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationUntilEndOfTurn,
		PowerDelta: game.Dynamic(game.DynamicAmount{
			Kind:   game.DynamicAmountTargetPower,
			Object: game.TargetPermanentReference(0),
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
		Object:         game.TargetPermanentReference(0),
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
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
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

func TestDelayedNextUpkeepTriggerFiresInUpkeepOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Delayed Card"}})
	addEffectSpellToStack(g, game.Player1, game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		},
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if len(g.DelayedTriggers) != 1 || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("delayed trigger fired before upkeep: triggers=%d hand=%d", len(g.DelayedTriggers), g.Players[game.Player1].Hand.Size())
	}

	g.Turn.TurnNumber++
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after upkeep = %d, want 0", len(g.DelayedTriggers))
	}
	if g.Players[game.Player1].Hand.Size() != 2 {
		t.Fatalf("player hand size = %d, want delayed draw plus draw-step draw", g.Players[game.Player1].Hand.Size())
	}
}

func TestDelayedSourceCardPermanentExileFollowsReturnedCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     creature.ObjectID,
		SourceCardID: creature.CardInstanceID,
		Controller:   game.Player1,
	}, &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Exile{Object: game.SourceCardPermanentReference()}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("source-card permanent remained on battlefield")
	}
	if !g.Players[game.Player1].Exile.Contains(creature.CardInstanceID) {
		t.Fatal("source-card permanent was not exiled")
	}
}

func TestDelayedSourceCardPermanentSacrificeFailsClosedWhenSourceLeft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}, &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Sacrifice{Object: game.SourceCardPermanentReference()}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
	if !movePermanentToZone(g, source, zone.Hand) {
		t.Fatal("failed to move source card from battlefield")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if _, ok := permanentByObjectID(g, other.ObjectID); !ok {
		t.Fatal("unresolved source-card reference sacrificed another permanent")
	}
}

func TestDelayedSourceCardReturnMovesCardFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	if !movePermanentToZone(g, creature, zone.Graveyard) {
		t.Fatal("failed to move source card to graveyard")
	}
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     creature.ObjectID,
		SourceCardID: creature.CardInstanceID,
		Controller:   game.Player1,
	}, &game.DelayedTriggerDef{
		Timing: game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceSource},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if !g.Players[game.Player1].Hand.Contains(creature.CardInstanceID) {
		t.Fatal("source card was not returned from graveyard")
	}
}

func TestDelayedNextEndStepTriggersUseAPNAPStackOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.DelayedTriggers = append(g.DelayedTriggers,
		game.DelayedTrigger{
			ID:         1,
			Controller: game.Player2,
			Timing:     game.DelayedAtBeginningOfNextEndStep,
			Ability:    game.TriggeredAbility{},
		},
		game.DelayedTrigger{
			ID:         2,
			Controller: game.Player1,
			Timing:     game.DelayedAtBeginningOfNextEndStep,
			Ability:    game.TriggeredAbility{},
		},
	)

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	NewEngine(nil).putTriggeredAbilitiesOnStack(g)

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	if objects[0].Controller != game.Player1 || objects[1].Controller != game.Player2 {
		t.Fatalf("stack controllers bottom-to-top = %v/%v, want APNAP Player1/Player2", objects[0].Controller, objects[1].Controller)
	}
}

func TestDelayedNextEndStepCreatedDuringEndStepWaitsForNextBoundary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStack(g)

	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		Controller: game.Player1,
		Timing:     game.DelayedAtBeginningOfNextEndStep,
		Ability:    game.TriggeredAbility{Content: game.Mode{}.Ability()},
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("next-end-step trigger created during the end step fired after its beginning")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 waiting", len(g.DelayedTriggers))
	}

	g.Turn.TurnNumber++
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delayed trigger did not fire at the next end-step boundary")
	}
	if len(g.DelayedTriggers) != 0 || g.Stack.Size() != 1 {
		t.Fatalf("after next end step: delayed=%d stack=%d, want 0/1", len(g.DelayedTriggers), g.Stack.Size())
	}
}

// --- Issue #225: source-tied control durations ---

// makeCreaturePermanent is a minimal helper that adds a creature permanent
// controlled by controller with the given name.
func makeCreaturePermanent(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// applySourceTiedControlEffect simulates an activated ability from source
// giving controller control of target with the provided duration.
func applySourceTiedControlEffect(g *game.Game, controller game.PlayerID, source, target *game.Permanent, duration game.EffectDuration) bool {
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   controller,
	}
	return applyTypedContinuousEffects(g, obj, target, []game.ContinuousEffect{{
		Layer:         game.LayerControl,
		NewController: opt.Val(game.Player1),
	}}, duration)
}

// TestSourceOnBattlefieldControlDurationExpiresWhenSourceLeaves verifies that
// DurationForAsLongAsSourceOnBattlefield expires at SBA cadence when the
// source permanent leaves the battlefield.
func TestSourceOnBattlefieldControlDurationExpiresWhenSourceLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Control Enchantment",
	}})
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")

	if !applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsSourceOnBattlefield) {
		t.Fatal("applyTypedContinuousEffects returned false")
	}
	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller before source leaves = %v, want Player1", got)
	}

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after source leaves = %v, want Player2 (original)", got)
	}
}

// TestSourceOnBattlefieldControlDurationPersistsWhileSourcePresent verifies
// that DurationForAsLongAsSourceOnBattlefield does NOT expire while the source
// remains on the battlefield.
func TestSourceOnBattlefieldControlDurationPersistsWhileSourcePresent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Control Enchantment",
	}})
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")

	applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsSourceOnBattlefield)
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller while source present = %v, want Player1", got)
	}
}

func TestSourceTiedControlDurationsExpireWhenSourcePhasesOut(t *testing.T) {
	for _, test := range []struct {
		name     string
		duration game.EffectDuration
	}{
		{"source on battlefield", game.DurationForAsLongAsSourceOnBattlefield},
		{"you control source", game.DurationForAsLongAsYouControlSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := makeCreaturePermanent(g, game.Player1, "Phasing Source")
			target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")
			applySourceTiedControlEffect(g, game.Player1, source, target, test.duration)

			source.PhasedOut = true
			engine.applyStateBasedActions(g)

			if got := effectiveController(g, target); got != game.Player2 {
				t.Fatalf("controller after source phases out = %v, want Player2", got)
			}
			if len(g.ContinuousEffects) != 0 {
				t.Fatalf("continuous effects after source phases out = %d, want 0", len(g.ContinuousEffects))
			}
		})
	}
}

func TestSourceTiedControlDurationExpiresBeforeLegendRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addLegendaryPermanent(g, game.Player1, "Godo")
	target := addLegendaryPermanent(g, game.Player2, "Godo")
	source := makeCreaturePermanent(g, game.Player1, "Control Source")
	applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsSourceOnBattlefield)
	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller before phasing = %v, want Player1", got)
	}

	source.PhasedOut = true
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after source phases out = %v, want Player2", got)
	}
	if _, ok := permanentByObjectID(g, original.ObjectID); !ok {
		t.Fatal("original legendary permanent left battlefield")
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("legend rule acted before phased source's control duration expired")
	}
}

func TestConditionalControlDurationsExpireToFixedPointBeforeLegendRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addLegendaryPermanent(g, game.Player1, "Godo")
	target := addLegendaryPermanent(g, game.Player2, "Godo")
	anchor := makeCreaturePermanent(g, game.Player1, "Anchor")
	dependentSource := makeCreaturePermanent(g, game.Player2, "Dependent Source")

	applySourceTiedControlEffect(
		g,
		game.Player1,
		anchor,
		dependentSource,
		game.DurationForAsLongAsSourceOnBattlefield,
	)
	applySourceTiedControlEffect(
		g,
		game.Player1,
		dependentSource,
		target,
		game.DurationForAsLongAsYouControlSource,
	)
	if got := effectiveController(g, dependentSource); got != game.Player1 {
		t.Fatalf("dependent source controller before phasing = %v, want Player1", got)
	}
	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("target controller before phasing = %v, want Player1", got)
	}

	anchor.PhasedOut = true
	engine.applyStateBasedActions(g)

	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects after fixed-point expiration = %d, want 0", len(g.ContinuousEffects))
	}
	if got := effectiveController(g, dependentSource); got != game.Player2 {
		t.Fatalf("dependent source controller after expiration = %v, want Player2", got)
	}
	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("target controller after expiration = %v, want Player2", got)
	}
	if _, ok := permanentByObjectID(g, original.ObjectID); !ok {
		t.Fatal("original legendary permanent left battlefield")
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); !ok {
		t.Fatal("legend rule acted before dependent duration expired")
	}
}

// TestYouControlSourceDurationExpiresWhenSourceLeaves verifies that
// DurationForAsLongAsYouControlSource expires when the source leaves the
// battlefield.
func TestYouControlSourceDurationExpiresWhenSourceLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := makeCreaturePermanent(g, game.Player1, "Merieke-style Creature")
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")

	applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsYouControlSource)

	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller before source leaves = %v, want Player1", got)
	}

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after source leaves = %v, want Player2 (original)", got)
	}
}

// TestYouControlSourceDurationExpiresWhenControllerLosesSource verifies that
// DurationForAsLongAsYouControlSource expires when the effect controller no
// longer controls the source (i.e. someone else takes control of it).
func TestYouControlSourceDurationExpiresWhenControllerLosesSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := makeCreaturePermanent(g, game.Player1, "Merieke-style Creature")
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")

	applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsYouControlSource)

	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller before control change = %v, want Player1", got)
	}

	// Player2 gains control of the source via a permanent-duration effect.
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		SourceObjectID:   g.IDGen.Next(),
		Controller:       game.Player2,
		Timestamp:        game.Timestamp(g.IDGen.Next()),
		Duration:         game.DurationPermanent,
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
	})

	// SBAs: Player1 no longer controls source, so the target effect expires.
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after source changes hands = %v, want Player2 (original owner)", got)
	}
}

// makeAuraAttachedTo creates a legal "Enchant creature" Aura permanent
// controlled by controller and attaches it to target, making target
// "enchanted".
func makeAuraAttachedTo(g *game.Game, controller game.PlayerID, target *game.Permanent, name string) *game.Permanent {
	aura := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Predicate: game.TargetPredicate{
					PermanentTypes: []types.Card{types.Creature},
				},
			}}},
		}},
	}})
	attachPermanent(g, aura, target)
	return aura
}

// TestControlledCreatureEnchantedDurationExpiresWhenAuraLeaves verifies that the
// attachment-dependent control duration (Rootwater Matriarch) expires at SBA
// cadence when the controlled creature is no longer enchanted.
func TestControlledCreatureEnchantedDurationExpiresWhenAuraLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := makeCreaturePermanent(g, game.Player1, "Rootwater Matriarch")
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")
	aura := makeAuraAttachedTo(g, game.Player2, target, "Some Aura")

	if !applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsControlledCreatureEnchanted) {
		t.Fatal("applyTypedContinuousEffects returned false")
	}
	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller while enchanted = %v, want Player1", got)
	}

	// The Aura leaves the battlefield: the creature is no longer enchanted.
	if !movePermanentToZone(g, aura, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after creature unenchanted = %v, want Player2 (original)", got)
	}
}

// TestControlledCreatureEnchantedDurationPersistsWhileEnchanted verifies that
// the attachment-dependent control duration does NOT expire while the
// controlled creature remains enchanted.
func TestControlledCreatureEnchantedDurationPersistsWhileEnchanted(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := makeCreaturePermanent(g, game.Player1, "Rootwater Matriarch")
	target := makeCreaturePermanent(g, game.Player2, "Stolen Creature")
	makeAuraAttachedTo(g, game.Player2, target, "Some Aura")

	applySourceTiedControlEffect(g, game.Player1, source, target, game.DurationForAsLongAsControlledCreatureEnchanted)
	engine.applyStateBasedActions(g)

	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller while enchanted = %v, want Player1", got)
	}
}

// TestSourceTiedDurationFailsClosedForSpellSource verifies that
// ApplyContinuous fails closed when a source-tied duration is used with a
// spell source (not a battlefield permanent).
func TestSourceTiedDurationFailsClosedForSpellSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	target := makeCreaturePermanent(g, game.Player2, "Target Creature")

	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		Duration: game.DurationForAsLongAsSourceOnBattlefield,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after spell with source-on-battlefield duration = %v, want Player2 (unchanged)", got)
	}
	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects = %d, want 0 (spell source rejected)", len(g.ContinuousEffects))
	}
}

// TestSourceTiedDurationFailsClosedForSpellSourceYouControl is the same test
// for DurationForAsLongAsYouControlSource.
func TestSourceTiedDurationFailsClosedForSpellSourceYouControl(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	target := makeCreaturePermanent(g, game.Player2, "Target Creature")

	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		Duration: game.DurationForAsLongAsYouControlSource,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after spell with you-control-source duration = %v, want Player2 (unchanged)", got)
	}
	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects = %d, want 0 (spell source rejected)", len(g.ContinuousEffects))
	}
}

// TestExistingUntilEOTControlUnchanged verifies that existing
// DurationUntilEndOfTurn control effects from #224 are unaffected by the new
// source-tied expiry path.
func TestExistingUntilEOTControlUnchanged(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	target := makeCreaturePermanent(g, game.Player2, "Target Creature")

	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller after gain-control until EOT = %v, want Player1", got)
	}
	// SBAs must NOT expire until-EOT effects.
	engine.applyStateBasedActions(g)
	if got := effectiveController(g, target); got != game.Player1 {
		t.Fatalf("controller after SBAs = %v, want Player1 (until-EOT should survive SBAs)", got)
	}
	// Cleanup step expires it.
	expireCleanupDurations(g)
	if got := effectiveController(g, target); got != game.Player2 {
		t.Fatalf("controller after cleanup = %v, want Player2 (original)", got)
	}
}

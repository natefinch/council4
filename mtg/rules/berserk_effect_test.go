package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const berserkTargetLink = game.LinkedKey("berserk-target")

func berserkAbility() game.AbilityContent {
	target := game.TargetPermanentReference(0)
	captured := game.CapturedObjectReference()
	return game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		}},
		Sequence: []game.Instruction{
			{Primitive: game.ApplyContinuous{
				Object: opt.Val(target),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: []game.Keyword{game.Trample},
				}},
				Duration:      game.DurationUntilEndOfTurn,
				PublishLinked: berserkTargetLink,
			}},
			{Primitive: game.ModifyPT{
				Object: target,
				PowerDelta: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountTargetPower,
					Object: target,
				}),
				Duration: game.DurationUntilEndOfTurn,
			}},
			{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing:         game.DelayedAtBeginningOfNextEndStep,
				CapturedObject: opt.Val(game.LinkedObjectReference(string(berserkTargetLink))),
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Destroy{Object: captured},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
						Object:                 opt.Val(captured),
						ObjectAttackedThisTurn: true,
					})}),
				}}}.Ability(),
			}}},
		},
	}.Ability()
}

func addBerserkSpellToStack(g *game.Game, target *game.Permanent) {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:         "Berserk Mechanics",
			Types:        []types.Card{types.Instant},
			SpellAbility: opt.Val(berserkAbility()),
		}},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		SourceCardID: sourceID,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	})
}

func resolveNextEndStep(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("next-end-step delayed trigger did not fire")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestBerserkSnapshotsEffectivePowerBeforePump(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addBerserkSpellToStack(g, creature)
	creature.Counters.Add(counter.PlusOnePlusOne, 1)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, creature); got != 8 {
		t.Fatalf("effective power = %d, want doubled effective power 8", got)
	}
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("target did not gain trample")
	}
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	if got := effectivePower(g, creature); got != 9 {
		t.Fatalf("power after later counter = %d, want snapshotted pump plus counter = 9", got)
	}
}

func TestBerserkNegativePowerUsesZeroForDynamicAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, -2)
	addBerserkSpellToStack(g, creature)

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 2 {
		t.Fatalf("continuous effect count = %d, want two", len(g.ContinuousEffects))
	}
	if g.ContinuousEffects[1].PowerDelta != 0 {
		t.Fatalf("power delta = %d, want zero", g.ContinuousEffects[1].PowerDelta)
	}
	if got := effectivePermanentValues(g, creature).power; got != -2 {
		t.Fatalf("raw effective power = %d, want unchanged -2", got)
	}
}

func TestBerserkIllegalTargetCountersSpellAndSchedulesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addBerserkSpellToStack(g, creature)
	if !movePermanentToZone(g, creature, zone.Graveyard) {
		t.Fatal("failed to remove target")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.ContinuousEffects) != 0 || len(g.DelayedTriggers) != 0 {
		t.Fatalf("effects = %d delayed = %d, want none after illegal target",
			len(g.ContinuousEffects), len(g.DelayedTriggers))
	}
}

func TestBerserkDestroysCapturedObjectIfItAttackedBeforeOrAfterResolution(t *testing.T) {
	for _, attackBefore := range []bool{true, false} {
		name := "after"
		if attackBefore {
			name = "before"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
			if attackBefore {
				declareAttackedHistory(g, creature)
			}
			addBerserkSpellToStack(g, creature)
			engine.resolveTopOfStack(g, &TurnLog{})
			if !attackBefore {
				declareAttackedHistory(g, creature)
			}

			resolveNextEndStep(t, g, engine)

			if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
				t.Fatal("attacking captured object was not destroyed")
			}
		})
	}
}

func TestBerserkDoesNotDestroyCapturedObjectThatDidNotAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addBerserkSpellToStack(g, creature)
	engine.resolveTopOfStack(g, &TurnLog{})

	resolveNextEndStep(t, g, engine)

	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("nonattacking captured object was destroyed")
	}
}

func TestBerserkCapturedObjectSurvivesControlChangeUntilDelayedDestroy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addBerserkSpellToStack(g, creature)
	engine.resolveTopOfStack(g, &TurnLog{})
	creature.Controller = game.Player1
	declareAttackedHistory(g, creature)

	resolveNextEndStep(t, g, engine)

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("captured object was not destroyed after changing controllers")
	}
}

func TestBerserkDestroysCapturedTokenObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatTokenCreaturePermanent(g, game.Player2, 3)
	declareAttackedHistory(g, creature)
	addBerserkSpellToStack(g, creature)
	engine.resolveTopOfStack(g, &TurnLog{})

	resolveNextEndStep(t, g, engine)

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("captured attacking token was not destroyed")
	}
}

func TestBerserkDoesNotDestroyPhasedOutCapturedObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	declareAttackedHistory(g, creature)
	addBerserkSpellToStack(g, creature)
	engine.resolveTopOfStack(g, &TurnLog{})
	creature.PhasedOut = true

	resolveNextEndStep(t, g, engine)

	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("phased-out captured object was destroyed")
	}
}

func TestBerserkCapturedObjectDoesNotFollowCardThroughLeaveAndReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	declareAttackedHistory(g, creature)
	addBerserkSpellToStack(g, creature)
	engine.resolveTopOfStack(g, &TurnLog{})
	cardID := creature.CardInstanceID
	if !movePermanentToZone(g, creature, zone.Exile) {
		t.Fatal("failed to exile captured target")
	}
	returned := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          creature.Owner,
		Controller:     creature.Controller,
	}
	g.Battlefield = append(g.Battlefield, returned)

	resolveNextEndStep(t, g, engine)

	if _, ok := permanentByObjectID(g, returned.ObjectID); !ok {
		t.Fatal("returned new object was destroyed")
	}
}

func TestBerserkDestroyUsesStandardIndestructibleAndRegenerationRules(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Game, *game.Permanent)
		check func(*testing.T, *game.Permanent)
	}{
		{
			name: "indestructible",
			setup: func(g *game.Game, creature *game.Permanent) {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID:               g.IDGen.Next(),
					AffectedObjectID: creature.ObjectID,
					Layer:            game.LayerAbility,
					AddKeywords:      []game.Keyword{game.Indestructible},
					Duration:         game.DurationPermanent,
				})
			},
			check: func(t *testing.T, creature *game.Permanent) {
				t.Helper()
				if creature.Tapped {
					t.Fatal("indestructible creature was tapped")
				}
			},
		},
		{
			name: "regeneration",
			setup: func(_ *game.Game, creature *game.Permanent) {
				creature.RegenerationShields = 1
			},
			check: func(t *testing.T, creature *game.Permanent) {
				t.Helper()
				if !creature.Tapped || creature.RegenerationShields != 0 {
					t.Fatalf("regenerated creature tapped=%v shields=%d, want true/0",
						creature.Tapped, creature.RegenerationShields)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
			test.setup(g, creature)
			declareAttackedHistory(g, creature)
			addBerserkSpellToStack(g, creature)
			engine.resolveTopOfStack(g, &TurnLog{})

			resolveNextEndStep(t, g, engine)

			live, ok := permanentByObjectID(g, creature.ObjectID)
			if !ok {
				t.Fatal("standard destruction protection did not preserve creature")
			}
			test.check(t, live)
		})
	}
}

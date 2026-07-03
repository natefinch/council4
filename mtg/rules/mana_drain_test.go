package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func manaDrainInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextMainPhase,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:   game.DynamicAmountCapturedTargetManaValue,
					Object: game.CapturedTargetStackObjectReference(0),
				}),
				ManaColor: mana.C,
			}}}}.Ability(),
		}}},
	}
}

func manaDrainTargetFace(uncounterable bool) *game.CardFace {
	face := &game.CardFace{
		Name:     "Variable Spell",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.U}),
		Types:    []types.Card{types.Sorcery},
	}
	if uncounterable {
		face.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
	}
	return face
}

func TestManaDrainCapturesTargetManaValueBeforeCountering(t *testing.T) {
	for _, uncounterable := range []bool{false, true} {
		t.Run(map[bool]string{false: "countered", true: "uncounterable"}[uncounterable], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.ActivePlayer = game.Player2
			target := addStackSpellWithFace(g, game.Player2, manaDrainTargetFace(uncounterable))
			target.XValue = 3
			addInstructionSpellToStackForController(
				g,
				game.Player1,
				manaDrainInstructions(),
				[]game.Target{game.StackObjectTarget(target.ID)},
			)

			engine.resolveTopOfStack(g, &TurnLog{})

			_, targetRemains := stackObjectByID(g, target.ID)
			if targetRemains != uncounterable {
				t.Fatalf("target remains = %v, want %v", targetRemains, uncounterable)
			}
			if len(g.DelayedTriggers) != 1 ||
				g.DelayedTriggers[0].CapturedTargetManaValueLKI[0] != 4 {
				t.Fatalf("delayed trigger = %#v, want captured mana value 4", g.DelayedTriggers)
			}

			g.TriggerEventCursor = len(g.Events)
			g.Turn.Step = game.StepPostcombatMain
			emitBeginningOfStepEvent(g, game.StepPostcombatMain)
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Mana Drain fired during another player's main phase")
			}

			g.Turn.TurnNumber++
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Step = game.StepNone
			emitBeginningOfStepEvent(g, game.StepPrecombatMain)
			if !engine.putTriggeredAbilitiesOnStack(g) || len(g.DelayedTriggers) != 0 {
				t.Fatalf("next-main trigger did not fire once: stack=%d delayed=%d", g.Stack.Size(), len(g.DelayedTriggers))
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 4 {
				t.Fatalf("colorless mana = %d, want 4", got)
			}

			g.TriggerEventCursor = len(g.Events)
			emitBeginningOfStepEvent(g, game.StepPostcombatMain)
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("one-shot next-main trigger fired a second time")
			}
		})
	}
}

func TestTargetStackObjectManaValueUsesCounteredTargetLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addStackSpellWithFace(g, game.Player2, manaDrainTargetFace(false))
	target.XValue = 3
	resolving := &game.StackObject{
		Targets: []game.Target{game.StackObjectTarget(target.ID)},
	}

	if !counterTargetStackObject(g, resolving, 0, false, game.CounteredSpellGraveyard) {
		t.Fatal("target spell was not countered")
	}
	dynamic := game.DynamicAmount{
		Kind:   game.DynamicAmountObjectManaValue,
		Object: game.TargetStackObjectReference(0),
	}
	if got := dynamicObjectManaValue(g, resolving, &dynamic); got != 4 {
		t.Fatalf("countered target mana value = %d, want 4", got)
	}
}

func TestManaDrainCapturesTokenBackedAlternateFaceManaValue(t *testing.T) {
	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Prepared Front",
		ManaCost: opt.Val(cost.Mana{cost.U}),
	}, Alternate: opt.Val(game.CardFace{
		Name:     "Prepared Spell",
		ManaCost: opt.Val(cost.Mana{cost.X, cost.R, cost.R}),
	})}
	value, ok := stackObjectManaValue(
		game.NewGame([game.NumPlayers]game.PlayerConfig{}),
		&game.StackObject{
			Kind:           game.StackSpell,
			Face:           game.FaceAlternate,
			SourceTokenDef: tokenDef,
			XValue:         3,
		},
	)
	if !ok || value != 5 {
		t.Fatalf("stackObjectManaValue() = %d, %v, want 5, true", value, ok)
	}
}

func TestNextMainDelayedTriggerUsesCombinedAPNAPOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	ordinary := addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepPrecombatMain,
	}, nil, nil)
	delayedSource := g.IDGen.Next()
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		SourceObjectID: delayedSource,
		Controller:     game.Player1,
		Timing:         game.DelayedAtBeginningOfNextMainPhase,
		Ability:        game.TriggeredAbility{Content: game.Mode{}.Ability()},
	})
	g.TriggerEventCursor = len(g.Events)
	emitBeginningOfStepEvent(g, game.StepPrecombatMain)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("main-phase triggers were not put on the stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 ||
		objects[0].SourceID != delayedSource ||
		objects[1].SourceID != ordinary.ObjectID {
		t.Fatalf("stack bottom-to-top = %#v, want active-player delayed then nonactive ordinary", objects)
	}
}

func TestNextMainDelayedTriggerCanFireAtPostcombatMainTheSameTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Step = game.StepPrecombatMain
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		Controller: game.Player1,
		Timing:     game.DelayedAtBeginningOfNextMainPhase,
		Ability:    game.TriggeredAbility{Content: game.Mode{}.Ability()},
	})
	g.TriggerEventCursor = len(g.Events)
	emitBeginningOfStepEvent(g, game.StepPostcombatMain)

	if !engine.putTriggeredAbilitiesOnStack(g) || len(g.DelayedTriggers) != 0 {
		t.Fatalf("same-turn postcombat main did not fire next-main trigger: stack=%d delayed=%d", g.Stack.Size(), len(g.DelayedTriggers))
	}
}

package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type numberChoiceAgent struct {
	number int
}

func (numberChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a numberChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if option.Label == strconv.Itoa(a.number) {
			return []int{option.Index}
		}
	}
	return nil
}

type triggerOrderAgent struct {
	optionCount int
	order       []int
}

func (*triggerOrderAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *triggerOrderAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.optionCount = len(request.Options)
	return append([]int(nil), a.order...)
}

func TestDelayedNextTurnUpkeepBoundedDrawChoice(t *testing.T) {
	for _, number := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(number), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			for range 2 {
				addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Delayed Draw"}})
			}

			key := game.ChoiceKey("draw-count")
			if !scheduleDelayedTrigger(g, &game.StackObject{Controller: game.Player2}, &game.DelayedTriggerDef{
				Timing: game.DelayedAtBeginningOfNextUpkeep,
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:      game.ResolutionChoiceNumber,
								MinNumber: 0,
								MaxNumber: 2,
							},
							PublishChoice: key,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountChosenNumber,
								ResultKey: game.ResultKey(key),
							}),
							Player: game.ControllerReference(),
						},
					},
				}}.Ability(),
			}) {
				t.Fatal("scheduleDelayedTrigger() = false")
			}

			g.Turn.Step = game.StepUpkeep
			emitBeginningOfStepEvent(g, game.StepUpkeep)
			engine.putTriggeredAbilitiesOnStack(g)
			if !g.Stack.IsEmpty() || len(g.DelayedTriggers) != 1 {
				t.Fatalf("trigger fired in current turn: stack=%d delayed=%d", g.Stack.Size(), len(g.DelayedTriggers))
			}

			g.Turn.TurnNumber++
			emitBeginningOfStepEvent(g, game.StepUpkeep)
			engine.putTriggeredAbilitiesOnStack(g)
			if g.Stack.Size() != 1 || len(g.DelayedTriggers) != 0 {
				t.Fatalf("next-turn upkeep scheduling: stack=%d delayed=%d, want 1/0", g.Stack.Size(), len(g.DelayedTriggers))
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player2: numberChoiceAgent{number: number},
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if got := g.Players[game.Player2].Hand.Size(); got != number {
				t.Fatalf("hand size = %d, want chosen draw %d", got, number)
			}

			g.Turn.TurnNumber++
			emitBeginningOfStepEvent(g, game.StepUpkeep)
			engine.putTriggeredAbilitiesOnStack(g)
			if !g.Stack.IsEmpty() {
				t.Fatal("one-shot delayed trigger fired a second time")
			}
		})
	}
}

func TestDelayedAndOrdinaryUpkeepTriggersShareControllerOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 2
	g.Turn.Step = game.StepUpkeep
	ordinary := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, nil, nil)
	delayedSource := g.IDGen.Next()
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		SourceObjectID: delayedSource,
		Controller:     game.Player1,
		CreatedTurn:    1,
		Timing:         game.DelayedAtBeginningOfNextUpkeep,
		Ability:        game.TriggeredAbility{Content: game.Mode{}.Ability()},
	})
	g.TriggerEventCursor = len(g.Events)
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	agent := &triggerOrderAgent{order: []int{0, 1}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("upkeep triggers were not put on the stack")
	}

	if agent.optionCount != 2 {
		t.Fatalf("trigger order choices = %d, want delayed and ordinary triggers together", agent.optionCount)
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	if objects[0].SourceID != ordinary.ObjectID || objects[1].SourceID != delayedSource {
		t.Fatalf("stack sources bottom-to-top = %v/%v, want ordinary then delayed", objects[0].SourceID, objects[1].SourceID)
	}
}

func TestDelayedAndOrdinaryUpkeepTriggersUseCombinedAPNAPOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 2
	g.Turn.Step = game.StepUpkeep
	ordinary := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, nil, nil)
	delayedSource := g.IDGen.Next()
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		SourceObjectID: delayedSource,
		Controller:     game.Player2,
		CreatedTurn:    1,
		Timing:         game.DelayedAtBeginningOfNextUpkeep,
		Ability:        game.TriggeredAbility{Content: game.Mode{}.Ability()},
	})
	g.TriggerEventCursor = len(g.Events)
	emitBeginningOfStepEvent(g, game.StepUpkeep)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("upkeep triggers were not put on the stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	if objects[0].SourceID != ordinary.ObjectID || objects[0].Controller != game.Player1 ||
		objects[1].SourceID != delayedSource || objects[1].Controller != game.Player2 {
		t.Fatalf("stack bottom-to-top = (%v,%v)/(%v,%v), want AP ordinary then NAP delayed",
			objects[0].SourceID, objects[0].Controller, objects[1].SourceID, objects[1].Controller)
	}
}

func TestCounterThenDelayedDrawUsesCounteredTargetControllerLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Delayed Draw"}})
	target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
	key := game.ChoiceKey("draw-count")
	targetController := game.CapturedTargetControllerReference(0)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.Choose{
					Choice: game.ResolutionChoice{
						Kind:            game.ResolutionChoiceNumber,
						PlayerReference: &targetController,
						MinNumber:       0,
						MaxNumber:       2,
					},
					PublishChoice: key,
				}},
				{Primitive: game.Draw{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:      game.DynamicAmountChosenNumber,
						ResultKey: game.ResultKey(key),
					}),
					Player: targetController,
				}},
			}}.Ability(),
		}}},
	}, []game.Target{game.StackObjectTarget(target.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, target.ID); ok {
		t.Fatal("target spell remained after counter")
	}
	if len(g.DelayedTriggers) != 1 ||
		g.DelayedTriggers[0].Controller != game.Player1 ||
		g.DelayedTriggers[0].CapturedTargetControllerLKI[0] != game.Player2 {
		t.Fatalf("delayed triggers = %+v, want source controller with target-controller LKI", g.DelayedTriggers)
	}
	g.Turn.TurnNumber++
	g.Turn.Step = game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	engine.putTriggeredAbilitiesOnStack(g)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: numberChoiceAgent{number: 1},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 1 || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("hands = %d/%d, want countered target controller to draw", g.Players[game.Player1].Hand.Size(), g.Players[game.Player2].Hand.Size())
	}
}

func TestDelayedCapturedControllerDoesNotCollideWithLocalTargetLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Captured Draw"}})
	localTarget := addStackSpell(g, game.Player3, "Local Target", []types.Card{types.Sorcery})
	g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
		Controller:                  game.Player1,
		CreatedTurn:                 1,
		Timing:                      game.DelayedAtBeginningOfNextUpkeep,
		CapturedTargetControllerLKI: map[int]game.PlayerID{0: game.Player2},
		Ability: game.TriggeredAbility{Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					StackObjectKinds: []game.StackObjectKind{game.StackSpell},
				},
			}},
			Sequence: []game.Instruction{
				{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
				{Primitive: game.Draw{
					Amount: game.Fixed(1),
					Player: game.CapturedTargetControllerReference(0),
				}},
			},
		}.Ability()},
	})
	g.Turn.TurnNumber = 2
	g.Turn.Step = game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("delayed trigger was not put on the stack")
	}
	delayed, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("delayed trigger missing from stack")
	}
	if len(delayed.Targets) != 1 || delayed.Targets[0].StackObjectID != localTarget.ID {
		t.Fatalf("delayed local targets = %+v, want local stack object %v", delayed.Targets, localTarget.ID)
	}
	if delayed.CapturedTargetControllerLKI[0] != game.Player2 {
		t.Fatalf("captured controller = %v, want Player2", delayed.CapturedTargetControllerLKI[0])
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, localTarget.ID); ok {
		t.Fatal("ordinary local target reference did not counter the local spell")
	}
	if delayed.TargetControllerLKI[0] != game.Player3 {
		t.Fatalf("local target controller LKI = %v, want Player3", delayed.TargetControllerLKI[0])
	}
	if g.Players[game.Player2].Hand.Size() != 1 || g.Players[game.Player3].Hand.Size() != 0 {
		t.Fatalf("hands Player2/Player3 = %d/%d, want captured enclosing controller to draw",
			g.Players[game.Player2].Hand.Size(), g.Players[game.Player3].Hand.Size())
	}
}

func TestCounterThenDelayedDrawsPreserveControllersWhenCounterFails(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{Name: "Delayed Draw"}})
	}
	protectedID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected Spell",
		Types:           []types.Card{types.Sorcery},
		StaticAbilities: []game.StaticAbility{game.CantBeCounteredStaticBody},
	}})
	g.Players[game.Player2].Hand.Remove(protectedID)
	protected := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   protectedID,
		Controller: game.Player2,
	}
	g.Stack.Push(protected)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.CapturedTargetControllerReference(0),
			}}}}.Ability(),
		}}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}}},
	}, []game.Target{game.StackObjectTarget(protected.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, protected.ID); !ok {
		t.Fatal("can't-be-countered target left the stack")
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want 2 despite failed counter", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].Controller != game.Player1 || g.DelayedTriggers[1].Controller != game.Player1 {
		t.Fatalf("delayed controllers = %v/%v, want source controller for both", g.DelayedTriggers[0].Controller, g.DelayedTriggers[1].Controller)
	}

	g.Turn.TurnNumber++
	g.Turn.Step = game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	engine.putTriggeredAbilitiesOnStack(g)
	objects := g.Stack.Objects()
	if len(objects) != 3 {
		t.Fatalf("stack objects = %d, want protected spell plus two delayed triggers", len(objects))
	}
	if objects[1].Controller != game.Player1 || objects[2].Controller != game.Player1 {
		t.Fatalf("delayed stack controllers bottom-to-top = %v/%v, want source controller for both", objects[1].Controller, objects[2].Controller)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player2].Hand.Size() != 1 {
		t.Fatalf("hands after independent delayed draws = %d/%d, want 1/1", g.Players[game.Player1].Hand.Size(), g.Players[game.Player2].Hand.Size())
	}
}

func TestCounterThenDelayedDrawsDoNotScheduleWithIllegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
	instructions := []game.Instruction{
		{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
		{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextUpkeep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}}},
	}
	addInstructionSpellToStackForController(g, game.Player1, instructions, []game.Target{game.StackObjectTarget(target.ID)})
	counter, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("counter spell missing")
	}
	counter.TargetCounts = []int{1}
	card, ok := g.GetCardInstance(counter.SourceID)
	if !ok {
		t.Fatal("counter spell card missing")
	}
	card.Def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{
				StackObjectKinds: []game.StackObjectKind{game.StackSpell},
			},
		}},
		Sequence: instructions,
	}.Ability())
	if _, ok := g.Stack.RemoveByID(target.ID); !ok {
		t.Fatal("failed to remove target")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want none when all targets are illegal", len(g.DelayedTriggers))
	}
}

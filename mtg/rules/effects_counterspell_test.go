package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDamageAndLoseLifeEffectsCanEliminatePlayers(t *testing.T) {
	tests := []struct {
		name      string
		primitive game.Primitive
	}{
		{name: "damage", primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}},
		{name: "lose life", primitive: game.LoseLife{Amount: game.Fixed(3), Player: game.TargetPlayerReference(0)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player2].Life = 3
			addEffectSpellToStack(g, game.Player1, tt.primitive, []game.Target{game.PlayerTarget(game.Player2)})

			engine.resolveTopOfStack(g, &TurnLog{})
			losses := engine.applyStateBasedActions(g)

			if len(losses) != 1 {
				t.Fatalf("losses = %d, want 1", len(losses))
			}
			if losses[0].Player != game.Player2 {
				t.Fatalf("loss player = %v, want %v", losses[0].Player, game.Player2)
			}
			if !g.Players[game.Player2].Eliminated {
				t.Fatal("player 2 was not eliminated")
			}
		})
	}
}

func TestFailedDrawEffectLogsAndEliminatesPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)
	losses := engine.applyStateBasedActions(g)
	log.Losses = append(log.Losses, losses...)

	if len(log.Draws) != 1 {
		t.Fatalf("draw logs = %d, want 1", len(log.Draws))
	}
	if !log.Draws[0].Failed {
		t.Fatal("draw log did not record failed draw")
	}
	if len(log.Losses) != 1 {
		t.Fatalf("loss logs = %d, want 1", len(log.Losses))
	}
	if log.Losses[0].Player != game.Player1 || log.Losses[0].Reason != LossReasonEmptyLibraryDraw {
		t.Fatalf("loss log = %+v, want player %v reason %q", log.Losses[0], game.Player1, LossReasonEmptyLibraryDraw)
	}
	if !g.Players[game.Player1].Eliminated {
		t.Fatal("player 1 was not eliminated")
	}
}

func TestMillScryAndSurveilLibraryEffectsUseDeterministicFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third"}})
	addEffectSpellToStack(g, game.Player1, game.Scry{Amount: game.Fixed(2), Player: game.ControllerReference()}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after scry = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Surveil{Amount: game.Fixed(2), Player: game.ControllerReference()}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after surveil = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Mill{Amount: game.Fixed(2), Player: game.ControllerReference()}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(third) || !g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("mill did not move top two cards to graveyard")
	}
	if got := g.Players[game.Player1].Library.All(); len(got) != 1 || got[0] != top {
		t.Fatalf("library after mill = %+v, want only original bottom card", got)
	}
}

func TestCounterEffectCountersTargetStackObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := g.IDGen.Next()
	g.CardInstances[targetID] = &game.CardInstance{
		ID: targetID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Target Spell",
			Types: []types.Card{types.Sorcery}},
		},
		Owner: game.Player2,
	}
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.CounterObject{Object: game.TargetStackObjectReference(0)}, []game.Target{game.StackObjectTarget(targetObj.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, targetObj.ID); ok {
		t.Fatal("target stack object remained after counter effect")
	}
	if !g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestCounterEffectCountersAbilityWithoutMovingSourceCard(t *testing.T) {
	for _, kind := range []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name:  "Ability Source",
				Types: []types.Card{types.Artifact},
			}})
			targetObj := &game.StackObject{
				ID:           g.IDGen.Next(),
				Kind:         kind,
				SourceID:     source.ObjectID,
				SourceCardID: source.CardInstanceID,
				Controller:   game.Player2,
			}
			g.Stack.Push(targetObj)
			addEffectSpellToStack(g, game.Player1, game.CounterObject{Object: game.TargetStackObjectReference(0)}, []game.Target{game.StackObjectTarget(targetObj.ID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			if _, ok := stackObjectByID(g, targetObj.ID); ok {
				t.Fatal("target ability remained after counter effect")
			}
			if _, ok := permanentByObjectID(g, source.ObjectID); !ok {
				t.Fatal("countering ability removed source permanent")
			}
			if g.Players[game.Player2].Graveyard.Contains(source.CardInstanceID) {
				t.Fatal("countering ability moved source card to graveyard")
			}
		})
	}
}

func TestCounterEffectCannotCounterUnknownFutureStackObjectKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackObjectKind(99),
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.CounterObject{Object: game.TargetStackObjectReference(0)}, []game.Target{game.StackObjectTarget(targetObj.ID)})
	counterObj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("counter spell missing from stack")
	}
	counterObj.TargetCounts = []int{1}
	counterCard, ok := g.GetCardInstance(counterObj.SourceID)
	if !ok {
		t.Fatal("counter spell card missing")
	}
	counterCard.Def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{
				StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
			},
		}},
		Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
	}.Ability())

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, targetObj.ID); !ok {
		t.Fatal("unknown future stack-object kind was countered")
	}
}

func TestCounterStackObjectRejectsUnknownFutureKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackObjectKind(99),
		Controller: game.Player2,
	}
	g.Stack.Push(target)

	if counterStackObject(g, target.ID) {
		t.Fatal("counterStackObject accepted unknown future stack-object kind")
	}
	if _, ok := stackObjectByID(g, target.ID); !ok {
		t.Fatal("unknown future stack-object kind left the stack")
	}
}

func TestCounterEffectCannotCounterProtectedCreatureSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Counter Shield",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeCountered,
				AffectedController: game.ControllerYou,
				SpellTypes:         []types.Card{types.Creature},
			}},
		}}},
	})
	targetID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Protected Creature",
		Types: []types.Card{types.Creature}},
	})
	g.Players[game.Player2].Hand.Remove(targetID)
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.CounterObject{Object: game.TargetStackObjectReference(0)}, []game.Target{game.StackObjectTarget(targetObj.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, targetObj.ID); !ok {
		t.Fatal("creature spell was countered despite can't-be-countered rule effect")
	}
	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("protected spell moved to graveyard")
	}
}

func TestCounterEffectCannotCounterSourceSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
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

	ordinaryID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ordinary Spell",
		Types: []types.Card{types.Sorcery},
	}})
	g.Players[game.Player2].Hand.Remove(ordinaryID)
	ordinary := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   ordinaryID,
		Controller: game.Player2,
	}
	g.Stack.Push(ordinary)

	if counterStackObject(g, protected.ID) {
		t.Fatal("source spell was countered despite its can't-be-countered rule effect")
	}
	if _, ok := stackObjectByID(g, protected.ID); !ok {
		t.Fatal("protected source spell left the stack")
	}
	if !counterStackObject(g, ordinary.ID) {
		t.Fatal("source-scoped rule effect protected another spell")
	}
}

func TestUnlessPaysCounterUsesTargetControllerPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addInstructionSpellToStackForController(
		g,
		game.Player1,
		unlessPaysCounterInstructions(cost.Mana{cost.G}),
		[]game.Target{game.StackObjectTarget(target.ID)},
	)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if _, ok := stackObjectByID(g, target.ID); !ok {
		t.Fatal("target spell was countered even though its controller paid")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Player != game.Player2 {
		t.Fatalf("choices = %+v, want payment choice for target controller", log.Choices)
	}
}

func TestUnlessPaysCounterCountersWhenPaymentDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addInstructionSpellToStackForController(
		g,
		game.Player1,
		unlessPaysCounterInstructions(cost.Mana{cost.G}),
		[]game.Target{game.StackObjectTarget(target.ID)},
	)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := stackObjectByID(g, target.ID); ok {
		t.Fatal("target spell remained on stack after payment was declined")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.SourceID) {
		t.Fatal("countered target spell did not move to graveyard")
	}
}

func TestUnlessPaysCounterSkipsPaymentWhenTargetGone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Sorcery})
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addInstructionSpellToStackForController(
		g,
		game.Player1,
		unlessPaysCounterInstructions(cost.Mana{cost.G}),
		[]game.Target{game.StackObjectTarget(target.ID)},
	)
	if _, ok := g.Stack.RemoveByID(target.ID); !ok {
		t.Fatal("failed to remove target stack object")
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no payment choice for missing target", log.Choices)
	}
}

func unlessPaysCounterInstructions(manaCost cost.Mana) []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay " + manaCost.String() + "?",
				Payer:    opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
				ManaCost: opt.Val(manaCost),
			}},
			PublishResult: "unless-paid",
		},
		{
			Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       "unless-paid",
				Succeeded: game.TriFalse,
			}),
		},
	}
}

func TestExcessDamageCanFeedLaterEffectAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(5), Recipient: game.AnyTargetDamageRecipient(0)},
			PublishResult: "damage",
		},
		{
			Primitive: game.Damage{
				Recipient: game.AnyTargetDamageRecipient(1),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectExcessDamage,
					ResultKey: "damage",
				}),
			},
		},
	}, []game.Target{game.PermanentTarget(target.ObjectID), game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("player 2 life = %d, want 37 from excess damage", got)
	}
}

func TestZeroExcessDamageDoesNotSatisfySuccessCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0), ResultAmountKind: game.EffectResultAmountExcessDamage},
			PublishResult: "excess",
		},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(5)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "excess", Succeeded: game.TriTrue}),
		},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("player 1 life = %d, want no gain from zero excess damage", got)
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestSelfETBTriggerGoesOnStackAndResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	spellID := addCardToHand(g, game.Player1, triggeredCreature(game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil))
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast triggered creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("ETB trigger was not put on stack")
	}

	obj := g.Stack.Peek()
	if obj == nil || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != spellID {
		t.Fatalf("top of stack = %+v, want triggered ability from %v", obj, spellID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want trigger to draw one card", got)
	}
}

func TestDeathTriggerGoesOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:              game.EventPermanentDied,
		MatchPermanentType: true,
		PermanentType:      game.TypeCreature,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)
	creature := addCombatCreaturePermanent(g, game.Player2)

	destroyPermanent(g, creature.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("death trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want death trigger to draw one card", got)
	}
}

func TestSelfDeathTriggerUsesLeftBattlefieldSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	permanent := addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	destroyPermanent(g, permanent.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-death trigger was not put on stack")
	}
	obj := g.Stack.Peek()
	if obj == nil || obj.SourceID != permanent.ObjectID || obj.SourceCardID != permanent.CardInstanceID {
		t.Fatalf("top of stack = %+v, want self-death trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want self-death trigger to draw one card", got)
	}
}

func TestTokenTriggersUseTokenDefinition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	token := triggeredCreature(game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	permanent := createTokenPermanent(g, game.Player1, token)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token ETB trigger was not put on stack")
	}
	obj := g.Stack.Peek()
	if obj == nil || obj.SourceID != permanent.ObjectID || obj.SourceCardID != 0 || obj.SourceTokenDef != token {
		t.Fatalf("top of stack = %+v, want token trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want token ETB trigger to draw one card", got)
	}
}

func TestTokenSelfETBTriggerDoesNotMatchOtherToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "First Drawn"})
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Second Drawn"})
	token := triggeredCreature(game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	createTokenPermanent(g, game.Player1, token)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first token ETB trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	second := createTokenPermanent(g, game.Player1, token)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second token ETB trigger was not put on stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 1 {
		t.Fatalf("stack size = %d, want only the second token's self trigger", len(objects))
	}
	if objects[0].SourceID != second.ObjectID {
		t.Fatalf("trigger source = %v, want second token %v", objects[0].SourceID, second.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want exactly two token ETB triggers resolved", got)
	}
}

func TestTokenSelfDeathTriggerUsesLeftBattlefieldSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	token := triggeredCreature(game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	permanent := createTokenPermanent(g, game.Player1, token)
	destroyPermanent(g, permanent.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token self-death trigger was not put on stack")
	}
	obj := g.Stack.Peek()
	if obj == nil || obj.SourceID != permanent.ObjectID || obj.SourceCardID != 0 || obj.SourceTokenDef != token {
		t.Fatalf("top of stack = %+v, want token self-death trigger source %+v", obj, permanent)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want token self-death trigger to draw one card", got)
	}
}

func TestTokenSelfDeathTriggerDoesNotMatchOtherToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	token := triggeredCreature(game.TriggerPattern{
		Event:  game.EventPermanentDied,
		Source: game.TriggerSourceSelf,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	first := createTokenPermanent(g, game.Player1, token)
	createTokenPermanent(g, game.Player1, token)
	destroyPermanent(g, first.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token self-death trigger was not put on stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 1 {
		t.Fatalf("stack size = %d, want only the dying token's self trigger", len(objects))
	}
	if objects[0].SourceID != first.ObjectID {
		t.Fatalf("trigger source = %v, want dying token %v", objects[0].SourceID, first.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want exactly one token self-death trigger resolved", got)
	}
}

func TestDamageTriggerGoesOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:           game.EventDamageDealt,
		Player:          game.TriggerPlayerOpponent,
		DamageRecipient: game.DamageRecipientPlayer,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)

	dealPlayerDamage(g, 0, 0, game.Player2, game.Player2, 1, false)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want damage trigger to draw one card", got)
	}
}

func TestDrawTriggerChoosesDeterministicLegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Effect{{Type: game.EffectDamage, Amount: 1, TargetIndex: 0}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})

	if _, ok := engine.drawCard(g, game.Player1); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 39 {
		t.Fatalf("player 2 life = %d, want deterministic opponent target to lose 1", got)
	}
}

func TestTriggerTargetChoiceCanBeMadeByAgent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Effect{{Type: game.EffectDamage, Amount: 1, TargetIndex: 0}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player1); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("player 2 life = %d, want agent to choose another target", got)
	}
	if got := g.Players[game.Player3].Life; got != 39 {
		t.Fatalf("player 3 life = %d, want chosen target to lose 1", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceTarget || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want recorded target choice without fallback", log.Choices)
	}
}

func TestOptionalTriggeredAbilityChoiceHappensOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addCardToLibrary(g, game.Player2, &game.CardDef{Name: "Triggering Drawn"})
	addOptionalTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event: game.EventCardDrawn,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("optional trigger was not put on stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want optional trigger on stack before may choice", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("player 1 hand size = %d, want optional trigger declined on resolution", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceMay || log.Choices[0].Selected[0] != 0 {
		t.Fatalf("choices = %+v, want no may choice recorded", log.Choices)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "declined" {
		t.Fatalf("resolves = %+v, want declined optional trigger", log.Resolves)
	}
}

func TestCastTriggerGoesOnStackAboveCastSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
	}, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}}, nil)
	spellID := addCardToHand(g, game.Player1, greenInstant())
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cast trigger was not put on stack")
	}

	obj := g.Stack.Peek()
	if obj == nil || obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want cast trigger above cast spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want cast trigger to draw one card", got)
	}
}

func TestInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	t.Run("not put on stack when false at trigger time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 40
		addTriggeredPermanentWithCondition(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}})
		addCardToLibrary(g, game.Player2, &game.CardDef{Name: "Drawn"})
		if _, ok := engine.drawCard(g, game.Player2); !ok {
			t.Fatal("drawCard() = false, want true")
		}
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("intervening-if false trigger was put on stack")
		}
	})
	t.Run("does not resolve when false at resolution time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 41
		addTriggeredPermanentWithCondition(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}})
		addCardToLibrary(g, game.Player2, &game.CardDef{Name: "Event Drawn"})
		addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Trigger Drawn"})
		if _, ok := engine.drawCard(g, game.Player2); !ok {
			t.Fatal("drawCard() = false, want true")
		}
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("intervening-if true trigger was not put on stack")
		}
		g.Players[game.Player1].Life = 40
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		if g.Players[game.Player1].Hand.Size() != 0 {
			t.Fatal("intervening-if false on resolution still applied effect")
		}
		if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
			t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
		}
	})
}

func TestInterveningIfUsesEffectiveControllerAtTriggerTime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 10
	g.Players[game.Player2].Life = 41
	triggerSource := addTriggeredPermanentWithCondition(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}})
	newController := game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: triggerSource.ObjectID,
		Layer:            game.LayerControl,
		NewController:    &newController,
	})
	addCardToLibrary(g, game.Player3, &game.CardDef{Name: "Event Drawn"})
	addCardToLibrary(g, game.Player2, &game.CardDef{Name: "Trigger Drawn"})

	if _, ok := engine.drawCard(g, game.Player3); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger controlled by Player2 should use Player2 life for intervening-if")
	}
	obj := g.Stack.Peek()
	if obj == nil || obj.Controller != game.Player2 {
		t.Fatalf("trigger controller = %+v, want Player2", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 1 {
		t.Fatal("trigger did not resolve for effective controller")
	}
}

func TestTriggeredAbilitiesUseAPNAPStackOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player3, &game.CardDef{Name: "Drawn"})
	addTriggeredPermanent(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addTriggeredPermanent(g, game.Player2, game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)

	if _, ok := engine.drawCard(g, game.Player3); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw triggers were not put on stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want two triggers", len(objects))
	}
	if objects[0].Controller != game.Player1 || objects[1].Controller != game.Player2 {
		t.Fatalf("stack controllers bottom-to-top = %v, %v; want active player's trigger below next player's trigger", objects[0].Controller, objects[1].Controller)
	}
}

func TestTriggeredAbilitiesUseAgentOrderWithinController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	first := addTriggeredPermanent(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	second := addTriggeredPermanent(g, game.Player1, game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{Name: "Drawn"})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1, 0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("draw triggers were not put on stack")
	}

	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want two triggers", len(objects))
	}
	if objects[0].SourceID != second.ObjectID || objects[1].SourceID != first.ObjectID {
		t.Fatalf("stack sources bottom-to-top = %v, %v; want agent order %v, %v", objects[0].SourceID, objects[1].SourceID, second.ObjectID, first.ObjectID)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceOrder || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want recorded order choice without fallback", log.Choices)
	}
}

func addTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern game.TriggerPattern, effects []game.Effect, targets []game.TargetSpec) *game.Permanent {
	return addCombatPermanent(g, controller, triggeredCreature(pattern, effects, targets))
}

func addOptionalTriggeredPermanent(g *game.Game, controller game.PlayerID, pattern game.TriggerPattern, effects []game.Effect, targets []game.TargetSpec) *game.Permanent {
	card := triggeredCreature(pattern, effects, targets)
	card.Abilities[0].Optional = true
	return addCombatPermanent(g, controller, card)
}

func addTriggeredPermanentWithCondition(g *game.Game, controller game.PlayerID, pattern game.TriggerPattern, lifeAtLeast int, effects []game.Effect) *game.Permanent {
	permanent := addTriggeredPermanent(g, controller, pattern, effects, nil)
	card := g.GetCardInstance(permanent.CardInstanceID)
	card.Def.Abilities[0].Trigger.InterveningIfControllerLifeAtLeast = lifeAtLeast
	return permanent
}

func triggeredCreature(pattern game.TriggerPattern, effects []game.Effect, targets []game.TargetSpec) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{
		Name:      "Triggered Creature",
		Types:     []game.CardType{game.TypeCreature},
		ManaCost:  greenCost(),
		Power:     &pt,
		Toughness: &pt,
		Abilities: []game.AbilityDef{
			{
				Kind: game.TriggeredAbility,
				Trigger: &game.TriggerCondition{
					Type:    game.TriggerWhenever,
					Pattern: pattern,
				},
				Effects: effects,
				Targets: targets,
			},
		},
	}
}

type choiceOnlyAgent struct {
	choices [][]int
	next    int
}

func (a *choiceOnlyAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	return action.Pass()
}

func (a *choiceOnlyAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if a.next >= len(a.choices) {
		return nil
	}
	choice := append([]int(nil), a.choices[a.next]...)
	a.next++
	return choice
}

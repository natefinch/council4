package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludesCyclingFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want cycling activation", legal)
	}
}

func TestCyclingDiscardsCardAndDrawsOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}

	if !forest.Tapped {
		t.Fatal("cycling mana cost did not tap available land")
	}
	if g.Players[game.Player1].Hand.Contains(cyclingID) {
		t.Fatal("cycled card remained in hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cyclingID) {
		t.Fatal("cycled card was not discarded to graveyard")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 cycling ability", g.Stack.Size())
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != cyclingID || obj.SourceCardID != cyclingID || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("cycling stack object = %+v, want activated ability sourced from cycled card", obj)
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.CardID == cyclingID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	})
	assertEvent(t, g.Events, game.EventCycled, func(event game.Event) bool {
		return event.Controller == game.Player1 &&
			event.Player == game.Player1 &&
			event.CardID == cyclingID &&
			event.SourceID == cyclingID
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("cycling did not draw the top card")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0", g.Stack.Size())
	}
}

func TestCyclingEventTriggerFiresOnlyOnCycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycling Trigger Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCycled,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !discardCardFromHand(g, game.Player1, discardID) {
		t.Fatal("discardCardFromHand() = false")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycling trigger fired for ordinary discard")
	}

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycling trigger did not fire for cycling")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want cycling triggered ability", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want cycling trigger to draw one card", got)
	}
}

func TestCycleOrDiscardTriggerFiresForCyclingAndOrdinaryDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 1"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn 2"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDiscarded,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	cyclingID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !discardCardFromHand(g, game.Player1, discardID) {
		t.Fatal("discardCardFromHand() = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycle-or-discard trigger did not fire for ordinary discard")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cycle-or-discard trigger did not fire for cycling discard")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want two trigger draws", got)
	}
}

func TestHandCyclingGrantEnablesMatchingCardsOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Land}}, cost.Mana{cost.R})
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycling Land", Types: []types.Card{types.Land}}})
	creatureID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Not a Land", Types: []types.Card{types.Creature}}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.ActivateAbility(landID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want granted land cycling", legal)
	}
	if actionsContain(legal, action.ActivateAbility(creatureID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want no cycling for nonmatching creature", legal)
	}

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(landID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for granted cycling")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != landID || obj.SourceCardID != landID || obj.AbilityIndex != 0 {
		t.Fatalf("cycling stack object = %+v, want cycled card source", obj)
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.Player == game.Player1 && event.CardID == landID
	})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("resolved granted cycling did not draw a card")
	}
}

func TestHandCyclingGrantForCreatureCardsUsesGrantedCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Creature}}, cost.Mana{cost.O(1), cost.U})
	creatureID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycling Creature", Types: []types.Card{types.Creature}}})
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(creatureID, 0, nil, 0)) {
		t.Fatal("granted creature cycling was not legal with {1}{U} available")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(creatureID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for granted creature cycling")
	}
}

func TestHandCyclingGrantTracksSourceStateAndController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Land}}, cost.Mana{cost.R})
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycling Land", Types: []types.Card{types.Land}}})
	opponentLandID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent Land", Types: []types.Card{types.Land}}})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player2, types.Mountain)

	g.Turn.PriorityPlayer = game.Player1
	source.PhasedOut = true
	if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(landID, 0, nil, 0)) {
		t.Fatal("phased-out grant source still granted cycling")
	}
	source.PhasedOut = false
	source.Controller = game.Player2
	if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(landID, 0, nil, 0)) {
		t.Fatal("grant still affected previous controller's hand")
	}
	g.Turn.PriorityPlayer = game.Player2
	if !actionsContain(engine.legalActions(g, game.Player2), action.ActivateAbility(opponentLandID, 0, nil, 0)) {
		t.Fatal("grant did not move to new controller's hand")
	}
	g.Battlefield = nil
	if actionsContain(engine.legalActions(g, game.Player2), action.ActivateAbility(opponentLandID, 0, nil, 0)) {
		t.Fatal("removed grant source still granted cycling")
	}
}

func TestHandCyclingGrantDoesNotDuplicatePrintedCycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Creature}}, cost.Mana{cost.R})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.R}))
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if countActivationActionsForSource(engine.legalActions(g, game.Player1), cyclingID) != 1 {
		t.Fatalf("legal actions = %+v, want one same-cost cycling action", engine.legalActions(g, game.Player1))
	}
}

func TestHandCyclingGrantAddsDifferentCostToPrintedCycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addHandCyclingGrantPermanent(g, game.Player1, game.Selection{RequiredTypes: []types.Card{types.Creature}}, cost.Mana{cost.R})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(1)}))
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)
	if countActivationActionsForSource(legal, cyclingID) != 2 {
		t.Fatalf("legal actions = %+v, want printed and granted cycling actions", legal)
	}
}

func TestCyclingCostReductionPreservesColoredSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:             game.CostModifierAbility,
		AbilityKeyword:   game.Cycling,
		GenericReduction: 2,
	})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(2), cost.U}))
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was not legal with reduced {2}{U} cost and one Island")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want reduced Cycling to be payable")
	}
	if !island.Tapped {
		t.Fatal("reduced Cycling did not still require the colored mana symbol")
	}
	if forest.Tapped {
		t.Fatal("reduced Cycling paid generic mana after the generic portion was reduced away")
	}
}

func TestCyclingCostReductionDoesNotPayUnpayableColoredSymbols(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:             game.CostModifierAbility,
		AbilityKeyword:   game.Cycling,
		GenericReduction: 2,
	})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(2), cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was legal even though the reduced cost still required {U}")
	}
}

func TestCyclingCostReplacementPaysZeroRegardlessOfPrintedCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:           game.CostModifierAbility,
		AbilityKeyword: game.Cycling,
		SetManaCost:    opt.Val(cost.Mana{}),
	})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(2), cost.U}))
	g.Turn.PriorityPlayer = game.Player1

	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was not legal with zero replacement cost")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want zero replacement Cycling to be payable")
	}
}

func TestCyclingCostReplacementStillAppliesCostIncreases(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:           game.CostModifierAbility,
		AbilityKeyword: game.Cycling,
		SetManaCost:    opt.Val(cost.Mana{}),
	})
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:            game.CostModifierAbility,
		AbilityKeyword:  game.Cycling,
		GenericIncrease: 1,
	})
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(2), cost.U}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was not legal with replacement cost plus {1} increase")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want replacement plus increase to be payable")
	}
	if !forest.Tapped {
		t.Fatal("replacement cost ignored additional generic increase")
	}
}

func TestCyclingCostReplacementCanRequireHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanentWithCondition(g, game.Player1, game.CostModifier{
		Kind:           game.CostModifierAbility,
		AbilityKeyword: game.Cycling,
		SetManaCost:    opt.Val(cost.Mana{}),
	}, opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerHandSize, Op: compare.GreaterOrEqual, Value: 7}}}))
	for range 5 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Filler"}})
	}
	cyclingID := addCardToHand(g, game.Player1, typedCyclingCard(types.Creature, cost.Mana{cost.O(2), cost.U}))
	g.Turn.PriorityPlayer = game.Player1

	if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was legal before hand-size condition was satisfied")
	}

	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Seventh Card"}})
	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(cyclingID, 0, nil, 0)) {
		t.Fatal("cycling was not legal after hand-size condition was satisfied")
	}
}

func TestFirstCycleEachTurnCostReplacementExpiresAfterCycling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCyclingCostModifierPermanent(g, game.Player1, game.CostModifier{
		Kind:               game.CostModifierAbility,
		AbilityKeyword:     game.Cycling,
		SetManaCost:        opt.Val(cost.Mana{}),
		FirstCycleEachTurn: true,
	})
	firstID := addCardToHand(g, game.Player1, cyclingCard())
	secondID := addCardToHand(g, game.Player1, cyclingCard())
	g.Turn.PriorityPlayer = game.Player1

	if !actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(firstID, 0, nil, 0)) {
		t.Fatal("first Cycling activation was not legal for zero")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(firstID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want first Cycling activation to be free")
	}
	if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(secondID, 0, nil, 0)) {
		t.Fatal("second Cycling activation was still free in the same turn")
	}
}

func cyclingCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Cycling Test Card",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		}},
	}
}

func typedCyclingCard(cardType types.Card, manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Typed Cycling Test Card",
		Types: []types.Card{cardType},
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(manaCost),
		}},
	}
}

func countActivationActionsForSource(actions []action.Action, sourceID id.ID) int {
	count := 0
	for _, legal := range actions {
		if activate, ok := legal.ActivateAbilityPayload(); ok && activate.SourceID == sourceID {
			count++
		}
	}
	return count
}

func addHandCyclingGrantPermanent(g *game.Game, controller game.PlayerID, selection game.Selection, manaCost cost.Mana) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Cycling Granter",
			Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectGrantHandCardAbility,
					AffectedPlayer: game.PlayerYou,
					CardSelection:  selection,
					GrantedAbility: game.CyclingActivatedAbility(manaCost),
				}},
			}},
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func addCyclingCostModifierPermanent(g *game.Game, controller game.PlayerID, modifier game.CostModifier) *game.Permanent {
	return addCyclingCostModifierPermanentWithCondition(g, controller, modifier, opt.V[game.Condition]{})
}

func addCyclingCostModifierPermanentWithCondition(g *game.Game, controller game.PlayerID, modifier game.CostModifier, condition opt.V[game.Condition]) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Cycling Modifier",
			Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{
				Condition: condition,
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCostModifier,
					AffectedPlayer: game.PlayerYou,
					CostModifier:   modifier,
				}},
			}},
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func actionsContain(actions []action.Action, want action.Action) bool {
	for _, got := range actions {
		if actionsEqual(got, want) {
			return true
		}
	}
	return false
}

func landcyclingCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Landcycling Test Card",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{
			game.LandcyclingActivatedAbility(cost.Mana{cost.O(1)}, game.SearchSpec{
				Filter: game.Selection{
					RequiredTypes: []types.Card{types.Land},
					Supertypes:    []types.Super{types.Basic},
				},
			}),
		}},
	}
}

func TestLandcyclingSearchesBasicLandToHandOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, landcyclingCard())
	basicID := addCardToLibrary(g, game.Player1, basicLandDef(types.Forest))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonbasic", Types: []types.Card{types.Land}}})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for landcycling")
	}
	if !forest.Tapped {
		t.Fatal("landcycling mana cost did not tap available land")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("landcycled card was not discarded to graveyard")
	}
	assertEvent(t, g.Events, game.EventCycled, func(event game.Event) bool {
		return event.CardID == cardID && event.Player == game.Player1
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(basicID) || g.Players[game.Player1].Library.Contains(basicID) {
		t.Fatal("landcycling did not move the searched basic land to hand")
	}
}

// firstCycleTriggerPattern matches "Whenever you cycle another card for the
// first time each turn" (Valiant Rescuer): a cycle trigger gated on the first
// cycle occurrence of the turn.
func firstCycleTriggerPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                      game.EventCycled,
		Player:                     game.TriggerPlayerYou,
		PlayerEventOrdinalThisTurn: 1,
		ExcludeSelf:                true,
	}
}

// TestCycleOrdinalThreadsOccurrences confirms successive cycles by the same
// player are numbered 1, 2, ... within a turn and the count resets next turn.
func TestCycleOrdinalThreadsOccurrences(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	emitEvent(g, game.Event{Kind: game.EventCycled, Controller: game.Player1, Player: game.Player1, CardID: id.ID(1)})
	emitEvent(g, game.Event{Kind: game.EventCycled, Controller: game.Player1, Player: game.Player1, CardID: id.ID(2)})
	assertCycleOrdinal(t, g, id.ID(1), 1)
	assertCycleOrdinal(t, g, id.ID(2), 2)

	g.Turn.TurnNumber++
	markCurrentTurnEventStart(g)
	emitEvent(g, game.Event{Kind: game.EventCycled, Controller: game.Player1, Player: game.Player1, CardID: id.ID(3)})
	assertCycleOrdinal(t, g, id.ID(3), 1)
}

// TestFirstCycleEachTurnTriggerGatesOnFirstOccurrence confirms a
// first-cycle-each-turn trigger fires for the first cycle of the turn and not a
// later cycle the same turn.
func TestFirstCycleEachTurnTriggerGatesOnFirstOccurrence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, firstCycleTriggerPattern(),
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	firstCard := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventCycled, Controller: game.Player1, Player: game.Player1, CardID: firstCard})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first-cycle trigger did not fire for the first cycle of the turn")
	}

	secondCard := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventCycled, Controller: game.Player1, Player: game.Player1, CardID: secondCard})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first-cycle trigger fired for a later cycle the same turn")
	}
}

func assertCycleOrdinal(t *testing.T, g *game.Game, cardID id.ID, want int) {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventCycled && event.CardID == cardID {
			if event.PlayerEventOrdinalThisTurn != want {
				t.Fatalf("cycle ordinal for %v = %d, want %d", cardID, event.PlayerEventOrdinalThisTurn, want)
			}
			return
		}
	}
	t.Fatalf("no cycle event found for card %v", cardID)
}

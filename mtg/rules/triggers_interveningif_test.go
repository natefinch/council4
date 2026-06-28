package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	t.Run("not put on stack when false at trigger time", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 40
		addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		if _, ok := engine.drawCard(g, game.Player2, false); !ok {
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
		addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Event Drawn"}})
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Drawn"}})
		if _, ok := engine.drawCard(g, game.Player2, false); !ok {
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

func TestKickedInterveningIfChecksEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := game.TriggerCondition{InterveningIfEventPermanentWasKicked: true}
	if triggerInterveningIf(g, nil, game.Player1, &trigger, &game.Event{}) {
		t.Fatal("unkicked enter event satisfied kicked intervening-if")
	}
	if !triggerInterveningIf(g, nil, game.Player1, &trigger, &game.Event{KickerPaid: true}) {
		t.Fatal("kicked enter event did not satisfy kicked intervening-if")
	}
}

func TestWasCastInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	t.Run("not put on stack when permanent was not cast", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentWasCast: true,
		})
		emitEvent(g, game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			CardID:      source.CardInstanceID,
			PermanentID: source.ObjectID,
		})
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("non-cast permanent enter trigger was put on stack")
		}
	})

	t.Run("cast fact is rechecked on resolution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentWasCast: true,
		})
		emitEvent(g, game.Event{
			Kind:         game.EventPermanentEnteredBattlefield,
			CardID:       source.CardInstanceID,
			PermanentID:  source.ObjectID,
			EnterWasCast: true,
		})
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("cast permanent enter trigger was not put on stack")
		}
		obj, ok := g.Stack.Peek()
		if !ok {
			t.Fatal("missing triggered ability")
		}
		obj.TriggerEvent.EnterWasCast = false
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
			t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
		}
	})
}

func TestEnteredOrCastFromGraveyardInterveningIfChecksEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anyTrigger := game.TriggerCondition{InterveningIfEventPermanentEnteredOrCastFromGraveyard: true}
	cases := []struct {
		name  string
		event *game.Event
		want  bool
	}{
		{"entered from graveyard", &game.Event{FromZone: zone.Graveyard, ToZone: zone.Battlefield}, true},
		{"cast from graveyard", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Graveyard}, true},
		{"entered from exile", &game.Event{FromZone: zone.Exile, ToZone: zone.Battlefield}, false},
		{"cast from hand", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Hand}, false},
		{"cast-from-graveyard zone ignored when not cast", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterCastFromZone: zone.Graveyard}, false},
		{"opponent graveyard still counts for any-graveyard", &game.Event{FromZone: zone.Graveyard, ToZone: zone.Battlefield, Player: game.Player2}, true},
		{"nil event", nil, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := triggerInterveningIf(g, nil, game.Player1, &anyTrigger, test.event); got != test.want {
				t.Fatalf("triggerInterveningIf() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestEnteredOrCastFromControllerGraveyardInterveningIfChecksOwnership(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := game.TriggerCondition{InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: true}
	cases := []struct {
		name  string
		event *game.Event
		want  bool
	}{
		{"reanimated from your graveyard", &game.Event{FromZone: zone.Graveyard, ToZone: zone.Battlefield, Player: game.Player1}, true},
		{"reanimated from opponent graveyard", &game.Event{FromZone: zone.Graveyard, ToZone: zone.Battlefield, Player: game.Player2}, false},
		{"you cast it from your graveyard", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, Player: game.Player1, EnterWasCast: true, EnterCastFromZone: zone.Graveyard, EnterHasCastController: true, EnterCastController: game.Player1}, true},
		{"opponent cast it from your graveyard", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, Player: game.Player1, EnterWasCast: true, EnterCastFromZone: zone.Graveyard, EnterHasCastController: true, EnterCastController: game.Player2}, false},
		{"cast from your hand", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, Player: game.Player1, EnterWasCast: true, EnterCastFromZone: zone.Hand, EnterHasCastController: true, EnterCastController: game.Player1}, false},
		{"nil event", nil, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := triggerInterveningIf(g, nil, game.Player1, &trigger, test.event); got != test.want {
				t.Fatalf("triggerInterveningIf() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestCastFromControllerHandInterveningIfChecksEnterEvent verifies that the "if
// you cast it from your hand" enter-trigger intervening condition fires only
// when the entering permanent was cast by the trigger controller from their
// hand, and fails closed for every other provenance.
func TestCastFromControllerHandInterveningIfChecksEnterEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := game.TriggerCondition{InterveningIfEventPermanentWasCastFromControllerHand: true}
	cases := []struct {
		name  string
		event *game.Event
		want  bool
	}{
		{"you cast it from your hand", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Hand, EnterHasCastController: true, EnterCastController: game.Player1}, true},
		{"opponent cast it from hand", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Hand, EnterHasCastController: true, EnterCastController: game.Player2}, false},
		{"you cast it from graveyard", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Graveyard, EnterHasCastController: true, EnterCastController: game.Player1}, false},
		{"entered from hand without cast", &game.Event{FromZone: zone.Hand, ToZone: zone.Battlefield, EnterCastFromZone: zone.Hand}, false},
		{"cast from hand without controller provenance", &game.Event{FromZone: zone.Stack, ToZone: zone.Battlefield, EnterWasCast: true, EnterCastFromZone: zone.Hand}, false},
		{"nil event", nil, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := triggerInterveningIf(g, nil, game.Player1, &trigger, test.event); got != test.want {
				t.Fatalf("triggerInterveningIf() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestEnteredOrCastFromGraveyardInterveningIfCheckedOnStack(t *testing.T) {
	t.Run("reanimated permanent fires", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: true,
		})
		emitEvent(g, game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			CardID:      source.CardInstanceID,
			PermanentID: source.ObjectID,
			Player:      source.Owner,
			FromZone:    zone.Graveyard,
			ToZone:      zone.Battlefield,
		})
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("reanimated enter trigger was not put on stack")
		}
	})

	t.Run("ordinary hand cast does not fire", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
			InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: true,
		})
		emitEvent(g, game.Event{
			Kind:              game.EventPermanentEnteredBattlefield,
			CardID:            source.CardInstanceID,
			PermanentID:       source.ObjectID,
			Player:            source.Owner,
			FromZone:          zone.Stack,
			ToZone:            zone.Battlefield,
			EnterWasCast:      true,
			EnterCastFromZone: zone.Hand,
		})
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("hand-cast enter trigger was put on stack")
		}
	})
}

func TestControlsPermanentInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addSelfEnterInterveningTrigger(g, &game.TriggerCondition{
		InterveningCondition: opt.Val(game.Condition{
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
			}),
		}),
	})
	enter := game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		CardID:      source.CardInstanceID,
		PermanentID: source.ObjectID,
	}
	emitEvent(g, enter)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger was put on stack without an artifact")
	}

	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	artifact.PhasedOut = true
	emitEvent(g, enter)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger counted a phased-out artifact")
	}

	artifact.PhasedOut = false
	emitEvent(g, enter)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controls-artifact enter trigger was not put on stack")
	}
	artifact.PhasedOut = true
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestStepControlsPermanentInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Controller: game.TriggerControllerYou,
		Step:       game.StepUpkeep,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningCondition = opt.Val(game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{
			Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
		}),
	})
	event := game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: game.Player1,
		Player:     game.Player1,
		Step:       game.StepUpkeep,
	}

	emitEvent(g, event)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("step trigger fired without a controlled artifact")
	}
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	emitEvent(g, event)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("step trigger did not fire with a controlled artifact")
	}
	artifact.PhasedOut = true
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestStepLifeInterveningIfCheckedWhenTriggeringAndResolving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Controller: game.TriggerControllerYou,
		Step:       game.StepUpkeep,
	}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningCondition = opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 10}},
	})
	event := game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: game.Player1,
		Player:     game.Player1,
		Step:       game.StepUpkeep,
	}

	g.Players[game.Player1].Life = 9
	emitEvent(g, event)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("step trigger fired below the life threshold")
	}
	g.Players[game.Player1].Life = 10
	emitEvent(g, event)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("step trigger did not fire at the life threshold")
	}
	g.Players[game.Player1].Life = 9
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestInterveningIfUsesEffectiveControllerAtTriggerTime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 10
	g.Players[game.Player2].Life = 41
	triggerSource := addTriggeredPermanentWithCondition(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, 41, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}})
	newController := game.Player2
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: triggerSource.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(newController),
	})
	addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Event Drawn"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Drawn"}})

	if _, ok := engine.drawCard(g, game.Player3, false); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger controlled by Player2 should use Player2 life for intervening-if")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Controller != game.Player2 {
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
	addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)

	if _, ok := engine.drawCard(g, game.Player3, false); !ok {
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
	first := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	second := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1, 0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
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

func TestSimultaneousCounterTriggerCanTargetEarlierTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: game.EventCardDrawn}, nil, nil)
	addTriggeredPermanent(
		g,
		game.Player1,
		&game.TriggerPattern{Event: game.EventCardDrawn},
		[]game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
		[]game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{
				StackObjectKinds: []game.StackObjectKind{game.StackTriggeredAbility},
			},
		}},
	)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw triggers were not put on stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack size = %d, want two triggers", len(objects))
	}
	if len(objects[1].Targets) != 1 || objects[1].Targets[0] != game.StackObjectTarget(objects[0].ID) {
		t.Fatalf("counter trigger targets = %+v, want earlier trigger %v", objects[1].Targets, objects[0].ID)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Stack.IsEmpty() {
		t.Fatal("counter trigger did not remove earlier simultaneous trigger")
	}
}

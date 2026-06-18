package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// --- Issue #125: state-change and counter-added trigger matching ---

func TestSelfTapTriggerMatchesSelfTapEvent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:  game.EventPermanentTapped,
		Source: game.TriggerSourceSelf,
	}
	event := game.Event{
		Kind:           game.EventPermanentTapped,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("self-tap trigger did not match own tap event")
	}

	other := addCombatCreaturePermanent(g, game.Player2)
	eventOther := game.Event{
		Kind:           game.EventPermanentTapped,
		SourceObjectID: other.ObjectID,
		CardID:         other.CardInstanceID,
		PermanentID:    other.ObjectID,
		Controller:     game.Player2,
	}
	if triggerMatchesEvent(g, source, pattern, eventOther) {
		t.Fatal("self-tap trigger matched an opponent's tap event")
	}
}

func TestSelfUntapTriggerMatchesSelfUntapEvent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:  game.EventPermanentUntapped,
		Source: game.TriggerSourceSelf,
	}
	event := game.Event{
		Kind:           game.EventPermanentUntapped,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("self-untap trigger did not match own untap event")
	}

	other := addCombatCreaturePermanent(g, game.Player2)
	eventOther := game.Event{
		Kind:           game.EventPermanentUntapped,
		SourceObjectID: other.ObjectID,
		CardID:         other.CardInstanceID,
		PermanentID:    other.ObjectID,
		Controller:     game.Player2,
	}
	if triggerMatchesEvent(g, source, pattern, eventOther) {
		t.Fatal("self-untap trigger matched an opponent's untap event")
	}
}

func TestCounterKindFilterMatchesSameKind(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
	}
	event := game.Event{
		Kind:           game.EventCountersAdded,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		CounterKind:    counter.PlusOnePlusOne,
		Amount:         2,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("counter-kind filter did not match matching counter kind")
	}
}

func TestCounterKindFilterRejectsDifferentKind(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
	}
	event := game.Event{
		Kind:           game.EventCountersAdded,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		CounterKind:    counter.MinusOneMinusOne,
		Amount:         1,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("counter-kind filter matched wrong counter kind")
	}
}

func TestCounterKindFilterRejectsNonCounterEvent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
	}
	// EventPermanentTapped has zero CounterKind; MatchCounterKind guard fires.
	event := game.Event{
		Kind:           game.EventPermanentTapped,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("counter-kind filter matched a non-counter event")
	}
}

func TestSelfTapTriggerGoesOnStack(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentTapped,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	setPermanentTapped(g, source, true)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-tap trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want self-tap trigger to draw one card", got)
	}
}

func TestSelfCounterAddedTriggerGoesOnStack(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
		OneOrMore:        true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	addCountersToPermanent(g, source, counter.PlusOnePlusOne, 2)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self counter-added trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want counter trigger to draw one card", got)
	}
}

func TestSelfCounterAddedTriggerDoesNotFireForWrongKind(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	addCountersToPermanent(g, source, counter.MinusOneMinusOne, 1)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter trigger fired for wrong counter kind")
	}
}

func TestSelfCounterAddedOneOrMoreCoalesces(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
		OneOrMore:        true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	simultaneousID := g.IDGen.Next()
	emitEvent(g, game.Event{
		Kind:           game.EventCountersAdded,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		CounterKind:    counter.PlusOnePlusOne,
		Amount:         1,
		SimultaneousID: simultaneousID,
	})
	emitEvent(g, game.Event{
		Kind:           game.EventCountersAdded,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		CounterKind:    counter.PlusOnePlusOne,
		Amount:         1,
		SimultaneousID: simultaneousID,
	})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("counter trigger not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one coalesced trigger", got)
	}
}

// --- Issue #318: controller-scoped counter-added trigger matching ---

func TestControllerScopedCounterTriggerMatchesOtherControlledCreature(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Controller:       game.TriggerControllerYou,
		ExcludeSelf:      true,
		OneOrMore:        true,
		MatchCounterKind: true,
		CounterKind:      counter.PlusOnePlusOne,
		SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}

	other := addCombatCreaturePermanent(g, game.Player1)
	eventOther := game.Event{
		Kind:        game.EventCountersAdded,
		PermanentID: other.ObjectID,
		Controller:  game.Player1,
		CounterKind: counter.PlusOnePlusOne,
		Amount:      1,
	}
	if !triggerMatchesEvent(g, source, pattern, eventOther) {
		t.Fatal("controller-scoped counter trigger did not match another controlled creature")
	}

	eventSelf := game.Event{
		Kind:           game.EventCountersAdded,
		SourceObjectID: source.ObjectID,
		CardID:         source.CardInstanceID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
		CounterKind:    counter.PlusOnePlusOne,
		Amount:         1,
	}
	if triggerMatchesEvent(g, source, pattern, eventSelf) {
		t.Fatal("controller-scoped counter trigger matched its own counter event despite ExcludeSelf")
	}

	opponent := addCombatCreaturePermanent(g, game.Player2)
	eventOpponent := game.Event{
		Kind:        game.EventCountersAdded,
		PermanentID: opponent.ObjectID,
		Controller:  game.Player2,
		CounterKind: counter.PlusOnePlusOne,
		Amount:      1,
	}
	if triggerMatchesEvent(g, source, pattern, eventOpponent) {
		t.Fatal("controller-scoped counter trigger matched an opponent's counter event")
	}
}

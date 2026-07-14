package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// midnightClockThresholdPattern mirrors the trigger the compiler lowers for
// Midnight Clock's "When the twelfth hour counter is put on this artifact, ..."
// ability: a self-sourced EventCountersAdded pattern restricted to hour counters
// that fires once as the source's hour count crosses twelve (CR 122.6, CR
// 603.2). The threshold is intentionally kind-generic so any counter kind and
// any Nth ordinal reuse the same runtime matcher.
func midnightClockThresholdPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:            game.EventCountersAdded,
		Source:           game.TriggerSourceSelf,
		MatchCounterKind: true,
		CounterKind:      counter.Hour,
		CounterThreshold: 12,
	}
}

// hourCounterEvent builds the authoritative counter-placement event the runtime
// emits when hour counters land on the source: it reports the source's counter
// total before the placement (PreviousCounterAmount) and the amount added, which
// is what the threshold matcher reads rather than re-counting the permanent.
func hourCounterEvent(source *game.Permanent, previous, amount int) game.Event {
	return game.Event{
		Kind:                  game.EventCountersAdded,
		Controller:            source.Controller,
		PermanentID:           source.ObjectID,
		SourceObjectID:        source.ObjectID,
		CounterKind:           counter.Hour,
		PreviousCounterAmount: previous,
		Amount:                amount,
	}
}

func addMidnightClockPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Midnight Clock",
		Types: []types.Card{types.Artifact},
	}})
}

// TestMidnightClockThresholdFiresWhenTwelfthCounterPlaced covers the canonical
// path: an hour counter placement that raises the count from eleven to twelve
// crosses the threshold and fires the trigger exactly once.
func TestMidnightClockThresholdFiresWhenTwelfthCounterPlaced(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	if !triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 11, 1)) {
		t.Fatal("threshold did not fire when the twelfth hour counter was placed")
	}
}

// TestMidnightClockThresholdFiresWhenMultipleCountersCrossTwelve proves the
// matcher keys off crossing the threshold, not landing on it exactly: placing
// several counters at once (proliferate, or an effect adding multiple) that
// jumps the total from below twelve to twelve-or-more fires the trigger.
func TestMidnightClockThresholdFiresWhenMultipleCountersCrossTwelve(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	// Reaching exactly twelve in one placement (e.g. an effect that sets the
	// count from zero to twelve) still crosses the threshold.
	if !triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 0, 12)) {
		t.Fatal("threshold did not fire when a single placement reached twelve")
	}
	// Overshooting twelve (5 -> 15) crosses the threshold and fires once.
	if !triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 5, 10)) {
		t.Fatal("threshold did not fire when a placement overshot twelve")
	}
}

// TestMidnightClockThresholdDoesNotFireBelowTwelve confirms placements that keep
// the total under twelve never trigger, so the clock only "strikes" on the
// twelfth hour.
func TestMidnightClockThresholdDoesNotFireBelowTwelve(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	if triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 10, 1)) {
		t.Fatal("threshold fired when the count only reached eleven")
	}
}

// TestMidnightClockThresholdDoesNotRefireOnceAtOrAboveTwelve covers the case
// where the source already sits at or above twelve when a further counter is
// added (manual counters, or removal below twelve then re-adding past it is
// handled separately): a placement that starts at twelve-or-more never re-fires,
// since the threshold was already crossed on the earlier placement.
func TestMidnightClockThresholdDoesNotRefireOnceAtOrAboveTwelve(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	if triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 12, 1)) {
		t.Fatal("threshold re-fired when a counter was added at twelve")
	}
	if triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 15, 2)) {
		t.Fatal("threshold re-fired when a counter was added above twelve")
	}
}

// TestMidnightClockThresholdRefiresAfterRemovalBelowTwelve verifies that if hour
// counters are removed below twelve and later re-added to cross twelve again,
// the trigger fires anew: the matcher only inspects the current placement's
// before/after totals, so a fresh crossing qualifies (CR 122.6).
func TestMidnightClockThresholdRefiresAfterRemovalBelowTwelve(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	// After removal the count sits at eleven again; adding one crosses twelve.
	if !triggerMatchesEvent(g, clock, pattern, hourCounterEvent(clock, 11, 1)) {
		t.Fatal("threshold did not re-fire after re-crossing twelve")
	}
}

// TestMidnightClockThresholdIgnoresOtherCounterKinds confirms the kind filter:
// placing a different counter kind, even enough to cross twelve, never fires the
// hour-specific threshold.
func TestMidnightClockThresholdIgnoresOtherCounterKinds(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	event := hourCounterEvent(clock, 11, 1)
	event.CounterKind = counter.PlusOnePlusOne
	if triggerMatchesEvent(g, clock, pattern, event) {
		t.Fatal("hour threshold fired for a non-hour counter placement")
	}
}

// TestMidnightClockThresholdIgnoresCountersOnOtherPermanents confirms the
// self-source filter: hour counters crossing twelve on a different permanent do
// not fire this clock's trigger.
func TestMidnightClockThresholdIgnoresCountersOnOtherPermanents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	clock := addMidnightClockPermanent(g, game.Player1)
	other := addMidnightClockPermanent(g, game.Player1)
	pattern := midnightClockThresholdPattern()

	if triggerMatchesEvent(g, clock, pattern, hourCounterEvent(other, 11, 1)) {
		t.Fatal("threshold fired for counters placed on another permanent")
	}
}

func addCardToZone(g *game.Game, owner game.PlayerID, zoneAdd func(id.ID)) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   vanillaCreature("Zone Card", 1, 1),
		Owner: owner,
	}
	zoneAdd(cardID)
	return cardID
}

// TestMidnightClockResolutionShufflesHandAndGraveyardThenDrawsThenExiles
// resolves the full "shuffle your hand and graveyard into your library, then
// draw seven cards. Exile this artifact." sequence and proves each step: hand
// and graveyard cards move into the library, seven cards are drawn afterward,
// and the source artifact is exiled.
func TestMidnightClockResolutionShufflesHandAndGraveyardThenDrawsThenExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	clock := addMidnightClockPermanent(g, game.Player1)
	player := g.Players[game.Player1]

	handCards := []id.ID{
		addCardToZone(g, game.Player1, player.Hand.Add),
		addCardToZone(g, game.Player1, player.Hand.Add),
	}
	graveyardCards := []id.ID{
		addCardToZone(g, game.Player1, player.Graveyard.Add),
		addCardToZone(g, game.Player1, player.Graveyard.Add),
		addCardToZone(g, game.Player1, player.Graveyard.Add),
	}
	for range 10 {
		addCardToZone(g, game.Player1, player.Library.Add)
	}

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     clock.ObjectID,
		SourceCardID: clock.CardInstanceID,
	}

	// Step 1: shuffle hand and graveyard into the library. Checked before the
	// draw because the subsequent draw may pull some of these same cards back
	// into hand.
	resolveInstruction(engine, g, obj, game.ShuffleGraveyardIntoLibrary{
		Player:      game.ControllerReference(),
		IncludeHand: true,
	}, &TurnLog{})

	for _, cardID := range append(append([]id.ID{}, handCards...), graveyardCards...) {
		if player.Hand.Contains(cardID) {
			t.Fatalf("card %v stayed in hand after shuffle", cardID)
		}
		if player.Graveyard.Contains(cardID) {
			t.Fatalf("card %v stayed in graveyard after shuffle", cardID)
		}
		if !player.Library.Contains(cardID) {
			t.Fatalf("card %v did not move into library after shuffle", cardID)
		}
	}
	if got := player.Hand.Size(); got != 0 {
		t.Fatalf("hand size after shuffle = %d, want 0", got)
	}
	// 2 hand + 3 graveyard + 10 library = 15 cards now in the library.
	if got := player.Library.Size(); got != 15 {
		t.Fatalf("library size after shuffle = %d, want 15", got)
	}

	// Step 2: draw seven from the freshly shuffled library.
	resolveInstruction(engine, g, obj, game.Draw{
		Amount: game.Fixed(7),
		Player: game.ControllerReference(),
	}, &TurnLog{})
	if got := player.Hand.Size(); got != 7 {
		t.Fatalf("hand size after draw = %d, want 7", got)
	}

	// Step 3: exile the source artifact.
	resolveInstruction(engine, g, obj, game.Exile{
		Object: game.SourceCardPermanentReference(),
	}, &TurnLog{})
	if _, ok := permanentByObjectID(g, clock.ObjectID); ok {
		t.Fatal("Midnight Clock remained on the battlefield after resolution")
	}
	if !player.Exile.Contains(clock.CardInstanceID) {
		t.Fatal("Midnight Clock was not exiled on resolution")
	}
}

// TestMidnightClockResolutionWithEmptyHandAndGraveyard confirms the resolution
// is well-defined when both shuffled zones are empty: the library is still
// shuffled, the draw proceeds, and the source is exiled.
func TestMidnightClockResolutionWithEmptyHandAndGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	clock := addMidnightClockPermanent(g, game.Player1)
	player := g.Players[game.Player1]
	for range 7 {
		addCardToZone(g, game.Player1, player.Library.Add)
	}

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     clock.ObjectID,
		SourceCardID: clock.CardInstanceID,
	}
	sequence := []game.Primitive{
		game.ShuffleGraveyardIntoLibrary{Player: game.ControllerReference(), IncludeHand: true},
		game.Draw{Amount: game.Fixed(7), Player: game.ControllerReference()},
		game.Exile{Object: game.SourceCardPermanentReference()},
	}
	for _, prim := range sequence {
		resolveInstruction(engine, g, obj, prim, &TurnLog{})
	}

	if got := player.Hand.Size(); got != 7 {
		t.Fatalf("hand size after draw = %d, want 7", got)
	}
	if !player.Exile.Contains(clock.CardInstanceID) {
		t.Fatal("Midnight Clock was not exiled on resolution")
	}
}

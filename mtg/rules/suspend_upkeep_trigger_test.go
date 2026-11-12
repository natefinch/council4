package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Suspend's upkeep time-counter removal is a triggered ability that uses the
// stack (#3789).
//
// Source review: official WotC Comprehensive Rules TXT, effective 2026-11-06
// (the current edition listed on https://magic.wizards.com/en/rules),
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829:
// 702.62a (suspend's second ability, "At the beginning of your upkeep, if this
// card is suspended, remove a time counter from it," is a triggered ability
// that functions in exile), 702.62b (a suspended card is in exile, has suspend
// and has a time counter), 603.4 (the intervening "if" is checked when the
// ability triggers and again as it resolves), 603.3b and 503.1a (beginning of
// upkeep triggers go on the stack in the order their controller chooses
// before the active player gets priority), 108.4a (the owner controls a
// card's abilities in exile), 400.7, 500.8 and 707.10.
// Ancestral Vision, oracle_id 9728dec9-d482-4c7a-8cdc-44d010dc878d, pinned
// corpus and https://api.scryfall.com/cards/9ec075ba-db56-4dcf-b1dc-fe6270b7ab36
// (rulings 2024-02-02, checked 2026-11-12): "If the first triggered ability of
// suspend (the one that removes time counters) is countered, no time counter
// is removed. The ability will trigger again at the beginning of the card's
// owner's next upkeep." Phyrexian Arena, oracle_id
// ee579a32-a048-4335-b966-231ba731cdea: "At the beginning of your upkeep, you
// draw a card and lose 1 life." Stifle, oracle_id
// b3b00911-ece7-4484-bc36-f211ce72b6cc: "Counter target activated or
// triggered ability."

// addSuspendedAncestralVision puts the registered Ancestral Vision into
// playerID's exile suspended with counters time counters.
func addSuspendedAncestralVision(t *testing.T, g *game.Game, playerID game.PlayerID, counters int) id.ID {
	t.Helper()
	return addSuspendedCard(g, playerID, registryCard(t, "Ancestral Vision"), counters)
}

// suspendUpkeepRemovals counts the time counters removed from cardID.
func suspendUpkeepRemovals(g *game.Game, cardID id.ID) int {
	removed := 0
	for _, event := range g.Events {
		if event.Kind == game.EventCountersRemoved && event.CardID == cardID && event.CounterKind == counter.Time {
			removed += event.Amount
		}
	}
	return removed
}

// TestSuspendUpkeepTriggerOrderedWithPhyrexianArena proves suspend's upkeep
// ability goes on the stack with Phyrexian Arena's trigger in the order Player
// 1 chooses, removes nothing until it resolves, and then removes one time
// counter. An opponent's suspended card doesn't trigger on Player 1's upkeep.
func TestSuspendUpkeepTriggerOrderedWithPhyrexianArena(t *testing.T) {
	for _, suspendFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("suspend resolves first %t", suspendFirst), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			arena := addCombatPermanent(g, game.Player1, registryCard(t, "Phyrexian Arena"))
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Arena draw"}})
			cardID := addSuspendedAncestralVision(t, g, game.Player1, 2)
			opponentCardID := addSuspendedAncestralVision(t, g, game.Player2, 2)
			var requests []game.ChoiceRequest
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoiceOrder {
					return request.DefaultSelection
				}
				requests = append(requests, request)
				// The last ability placed resolves first.
				order := make([]int, 0, len(request.Options))
				last := -1
				for _, option := range request.Options {
					isSuspend := option.Label == fmt.Sprintf("source=%v ability=0", cardID)
					if isSuspend == suspendFirst {
						last = option.Index
					} else {
						order = append(order, option.Index)
					}
				}
				return append(order, last)
			}}

			engine.beginSuspendUpkeep(g, game.Player1, suspendAgents(agent))

			if len(requests) != 1 || len(requests[0].Options) != 2 || g.Stack.Size() != 2 {
				t.Fatalf("order requests = %+v stack = %+v, want Player 1 to order the suspend and Arena triggers",
					requests, g.Stack.Objects())
			}
			if g.ExileCounterCount(cardID, counter.Time) != 2 || suspendUpkeepRemovals(g, cardID) != 0 {
				t.Fatal("time counter was removed before the upkeep trigger resolved")
			}
			lifeBefore := g.Players[game.Player1].Life
			if suspendFirst {
				requireSuspendUpkeepTrigger(t, g, cardID)
				engine.resolveTopOfStack(g, &TurnLog{})
				if g.ExileCounterCount(cardID, counter.Time) != 1 || g.Players[game.Player1].Life != lifeBefore {
					t.Fatal("suspend trigger did not resolve before Arena's trigger")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			} else {
				top, _ := g.Stack.Peek()
				if top.SourceID != arena.ObjectID {
					t.Fatalf("stack top = %+v, want Arena's trigger", top)
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if g.ExileCounterCount(cardID, counter.Time) != 2 || g.Players[game.Player1].Life != lifeBefore-1 {
					t.Fatal("Arena's trigger did not resolve before the suspend trigger")
				}
				requireSuspendUpkeepTrigger(t, g, cardID)
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, suspendAgents(agent), &TurnLog{})
			if !g.Stack.IsEmpty() || g.ExileCounterCount(cardID, counter.Time) != 1 || !cardIsSuspended(g, cardID) {
				t.Fatalf("stack = %+v, want Ancestral Vision still suspended with 1 counter", g.Stack.Objects())
			}
			removed := requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, 2, 1)
			if removed.Controller != game.Player1 {
				t.Fatalf("removal controller = %v, want the owner", removed.Controller)
			}
			if g.ExileCounterCount(opponentCardID, counter.Time) != 2 || suspendUpkeepRemovals(g, opponentCardID) != 0 {
				t.Fatal("an opponent's suspended card lost a time counter on Player 1's upkeep")
			}
		})
	}
}

// suspendUpkeepStifler casts Stifle once, targeting suspend's upkeep trigger
// when it is on top of the stack.
type suspendUpkeepStifler struct {
	g       *game.Game
	stifle  id.ID
	stifled bool
}

func (a *suspendUpkeepStifler) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	if top, ok := a.g.Stack.Peek(); !a.stifled && ok && top.SuspendUpkeep {
		for _, act := range legal {
			if cast, ok := act.CastSpellPayload(); ok && cast.CardID == a.stifle {
				a.stifled = true
				return act
			}
		}
	}
	return action.Pass()
}

func (*suspendUpkeepStifler) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

// TestStifledSuspendUpkeepTriggerKeepsTimeCounter plays Player 1's upkeep step:
// players get priority with the upkeep trigger on the stack, an opponent's
// Stifle counters it, and Ancestral Vision keeps its last time counter, stays
// suspended and isn't cast.
func TestStifledSuspendUpkeepTriggerKeepsTimeCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 1)
	stifleID := addCardToHand(g, game.Player2, registryCard(t, "Stifle"))
	addBasicLandPermanent(g, game.Player2, types.Island)
	agents := simPassPolicies()
	stifler := &suspendUpkeepStifler{g: g, stifle: stifleID}
	agents[game.Player2] = stifler
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.TriggerEventCursor = len(g.Events)

	engine.runUpkeepStep(g, agents, &TurnLog{})

	if !stifler.stifled || !g.Players[game.Player2].Graveyard.Contains(stifleID) {
		t.Fatal("Player 2 did not get priority to Stifle the suspend upkeep trigger")
	}
	if !g.Stack.IsEmpty() || g.ExileCounterCount(cardID, counter.Time) != 1 || !cardIsSuspended(g, cardID) ||
		suspendUpkeepRemovals(g, cardID) != 0 || len(suspendCastEvents(g, cardID)) != 0 {
		t.Fatalf("stack = %+v; the countered upkeep trigger removed a time counter", g.Stack.Objects())
	}
}

// TestSuspendUpkeepTriggerRemovesLastCounterIntoFreeCast resolves the upkeep
// trigger for Ancestral Vision's last time counter; the removal triggers the
// cast ability, which goes on the stack and casts the card for free.
func TestSuspendUpkeepTriggerRemovesLastCounterIntoFreeCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 1)
	target := game.PlayerTarget(game.Player2)
	agents := suspendAgents(&suspendTargetAgent{target: target})
	drawn := make([]id.ID, 3)
	for i := range drawn {
		drawn[i] = addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprintf("Vision draw %d", i)}})
	}

	engine.beginSuspendUpkeep(g, game.Player1, agents)
	requireSuspendUpkeepTrigger(t, g, cardID)
	if g.Stack.Size() != 1 || g.ExileCounterCount(cardID, counter.Time) != 1 {
		t.Fatal("last time counter was removed before the upkeep trigger resolved")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if g.ExileCounterCount(cardID, counter.Time) != 0 || cardIsSuspended(g, cardID) {
		t.Fatal("upkeep trigger did not remove the last time counter")
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	requireSuspendCastTrigger(t, g, cardID)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	spell, _ := requireSuspendCast(t, g, cardID)
	if spell.ManaSpentToCast != 0 || len(spell.Targets) != 1 || spell.Targets[0] != target {
		t.Fatalf("Ancestral Vision cast = %+v, want a free cast targeting Player 2", spell)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	for _, drawnID := range drawn {
		if !g.Players[game.Player2].Hand.Contains(drawnID) {
			t.Fatalf("Player 2 did not draw %v", drawnID)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved Ancestral Vision did not go to its owner's graveyard")
	}
}

// TestSuspendUpkeepTriggerAfterCardLeavesExile moves the suspended card out of
// exile and back with new time counters in response: the returned card is a
// new object (CR 400.7), so the trigger's "if this card is suspended" check
// fails (CR 603.4) and it removes no counter.
func TestSuspendUpkeepTriggerAfterCardLeavesExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 2)
	engine.beginSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	requireSuspendUpkeepTrigger(t, g, cardID)

	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("moving the card out of exile and back failed")
	}
	g.AddExileCounter(cardID, counter.Time, 2)
	if !cardIsSuspended(g, cardID) {
		t.Fatal("returned card is not suspended")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.ExileCounterCount(cardID, counter.Time) != 2 || suspendUpkeepRemovals(g, cardID) != 0 {
		t.Fatal("upkeep trigger removed a time counter from the card's new object")
	}
}

// TestSuspendUpkeepTriggersInAdditionalUpkeepStep plays a beginning phase with
// Paradox Haze's additional upkeep step: suspend's upkeep ability triggers in
// each upkeep, so Ancestral Vision with two time counters is cast and resolves
// in the additional upkeep.
func TestSuspendUpkeepTriggersInAdditionalUpkeepStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 2)
	addParadoxHazeAura(g, game.Player1, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw step card"}})
	for i := range 3 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprintf("Vision draw %d", i)}})
	}
	agents := simPassPolicies()
	agents[game.Player1] = &suspendTargetAgent{target: game.PlayerTarget(game.Player2)}
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.TriggerEventCursor = len(g.Events)

	log := TurnLog{}
	engine.runBeginningPhase(g, agents, &log)

	if got := countSteps(&log, game.StepUpkeep); got != 2 {
		t.Fatalf("upkeep steps = %d, want 2", got)
	}
	if suspendUpkeepRemovals(g, cardID) != 2 || len(suspendCastEvents(g, cardID)) != 1 {
		t.Fatalf("time counters removed = %d casts = %d, want one removal per upkeep and one cast",
			suspendUpkeepRemovals(g, cardID), len(suspendCastEvents(g, cardID)))
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player2].Hand.Size() != 3 {
		t.Fatal("Ancestral Vision did not resolve in the additional upkeep step")
	}
}

// TestSuspendUpkeepTriggerClonesIndependently clones the game with the upkeep
// trigger on the stack: the clone keeps the trigger's card identity and the
// card's counters, and resolving it in either game leaves the other unchanged.
func TestSuspendUpkeepTriggerClonesIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 2)
	engine.beginSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	original := requireSuspendUpkeepTrigger(t, g, cardID)
	clone := g.Clone()

	cloned := requireSuspendUpkeepTrigger(t, clone, cardID)
	if cloned.ID != original.ID || clone.ExileCounterCount(cardID, counter.Time) != 2 {
		t.Fatalf("cloned trigger = %+v, want the same trigger and counters", cloned)
	}
	engine.resolveTopOfStack(clone, &TurnLog{})
	if clone.ExileCounterCount(cardID, counter.Time) != 1 || g.ExileCounterCount(cardID, counter.Time) != 2 {
		t.Fatal("resolving the clone's trigger changed the original")
	}
	requireSuspendUpkeepTrigger(t, g, cardID)

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.ExileCounterCount(cardID, counter.Time) != 1 || !g.Stack.IsEmpty() || !clone.Stack.IsEmpty() {
		t.Fatal("original and clone resolved the upkeep trigger differently")
	}
}

// TestCopiedSuspendUpkeepTriggerRemovesAnotherCounter copies the upkeep
// ability (CR 707.10): the copy removes a time counter from the same card for
// its controller, and the original removes the last one, triggering the cast
// ability once for the card's owner.
func TestCopiedSuspendUpkeepTriggerRemovesAnotherCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedAncestralVision(t, g, game.Player1, 2)
	engine.beginSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	obj := requireSuspendUpkeepTrigger(t, g, cardID)
	addEffectSpellToStack(g, game.Player2, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.SuspendUpkeep || copied.ID == obj.ID || copied.Controller != game.Player2 {
		t.Fatalf("stack top = %+v, want Player 2's copy of the upkeep trigger", copied)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	removed := requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, 2, 1)
	if removed.Controller != game.Player2 || g.ExileCounterCount(cardID, counter.Time) != 1 {
		t.Fatalf("copy removal = %+v, want Player 2 removing one counter", removed)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendCastTrigger(t, g, cardID)
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want one cast trigger", g.Stack.Objects())
	}
}

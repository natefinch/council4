package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Rebound's delayed recast is a triggered ability that uses the stack (#3611).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     702.88a (rebound exiles the resolving spell "and, at the beginning of
//     your next upkeep, you may cast this card from exile without paying its
//     mana cost"), 702.88b (that cast follows the alternative-cost rules),
//     603.7a-c (a delayed triggered ability triggers once at the specified
//     time and affects the objects it was bound to; a card that changed zones
//     is a new object it can't find), 603.7f (a delayed trigger created by a
//     replacement from a static ability has that ability's source and is
//     controlled by the player who controlled it), 603.3b (the active player
//     orders their simultaneous triggers as they go on the stack), 117.3b
//     (players get priority while the trigger is on the stack) and 400.7.
//   - Trumpeting Herd, oracle_id 69501650-ed48-4ebf-9287-b0338c5bc5d5,
//     Scryfall rulings 2019-06-14 for b0f3b68e-f616-4687-bc2d-075165162cd1:
//     "Create a 3/3 green Elephant creature token. Rebound"; the recast is
//     optional and a card not cast stays exiled with no later chance.
//   - Stifle, oracle_id b3b00911-ece7-4484-bc36-f211ce72b6cc: "Counter target
//     activated or triggered ability. (Mana abilities can't be targeted.)"
//   - Phyrexian Arena, oracle_id ee579a32-a048-4335-b966-231ba731cdea: "At the
//     beginning of your upkeep, you draw a card and lose 1 life."

type reboundTriggerFixture struct {
	registry *cards.Registry
	engine   *Engine
	game     *game.Game
	cardID   id.ID
}

// newReboundTriggerFixture has Player1 cast and resolve registered Trumpeting
// Herd from hand, leaving it in exile with its rebound trigger pending.
func newReboundTriggerFixture(t *testing.T) *reboundTriggerFixture {
	t.Helper()
	registry, herd, _ := reboundReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID, _ := castTrumpetingHerd(t, g, engine, herd)
	engine.resolveTopOfStack(g, &TurnLog{})
	if controlledElephants(g, game.Player1) != 1 || !g.Players[game.Player1].Exile.Contains(cardID) ||
		!reboundTriggerPending(g, cardID) {
		t.Fatal("Trumpeting Herd did not rebound into exile with a pending trigger")
	}
	return &reboundTriggerFixture{registry: registry, engine: engine, game: g, cardID: cardID}
}

func (f *reboundTriggerFixture) def(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := f.registry.Lookup(name)
	if def == nil {
		t.Fatalf("registered %s missing", name)
	}
	return def
}

// requireReboundTriggerOnTop checks that g's top stack object is Player1's
// rebound trigger for cardID and that the card is still waiting in exile.
func requireReboundTriggerOnTop(t *testing.T, g *game.Game, cardID id.ID) *game.StackObject {
	t.Helper()
	top, ok := g.Stack.Peek()
	if !ok || top.Kind != game.StackTriggeredAbility || !top.Rebound || top.Controller != game.Player1 ||
		top.SourceCardID != cardID || top.CapturedCardID != cardID ||
		top.InlineTrigger == nil || top.InlineTrigger.Text != reboundDelayedTriggerText {
		t.Fatalf("stack top = %+v, want Player1's rebound trigger for the card", top)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || reboundTriggerPending(g, cardID) {
		t.Fatal("rebound trigger left the card or remained pending once on the stack")
	}
	return top
}

// noChoiceAgents fail the test if any player is offered a choice with options.
func noChoiceAgents(t *testing.T) [game.NumPlayers]PlayerAgent {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	for player := range agents {
		agents[player] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			if len(request.Options) != 0 {
				t.Fatalf("unexpected choice %+v", request)
			}
			return request.DefaultSelection
		}}
	}
	return agents
}

// TestReboundTriggerUsesStackAtControllersNextUpkeep proves rebound puts a
// triggered ability on the stack at its controller's next upkeep rather than
// casting immediately; the cast happens only when the trigger resolves (players
// can respond first; see TestStifleCountersReboundTrigger).
func TestReboundTriggerUsesStackAtControllersNextUpkeep(t *testing.T) {
	t.Parallel()
	f := newReboundTriggerFixture(t)
	g := f.game

	beginUpkeep(f.engine, g, game.Player2, noChoiceAgents(t))
	if g.Stack.Size() != 0 || !reboundTriggerPending(g, f.cardID) {
		t.Fatal("rebound triggered at an opponent's upkeep")
	}

	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want only the rebound trigger", g.Stack.Objects())
	}
	requireReboundTriggerOnTop(t, g, f.cardID)
	var observed bool
	for _, obj := range g.Stack.Objects() {
		observed = observed || obj.SourceZone == zone.Exile && obj.Kind == game.StackSpell
	}
	if observed || controlledElephants(g, game.Player1) != 1 {
		t.Fatal("rebound cast before its trigger resolved")
	}

	f.engine.resolveTopOfStack(g, &TurnLog{})
	spell, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 1 || spell.Kind != game.StackSpell || spell.SourceID != f.cardID ||
		spell.SourceZone != zone.Exile || spell.Controller != game.Player1 {
		t.Fatalf("stack = %+v, want Trumpeting Herd cast from exile", g.Stack.Objects())
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if controlledElephants(g, game.Player1) != 2 || !g.Players[game.Player1].Graveyard.Contains(f.cardID) ||
		reboundTriggerPending(g, f.cardID) {
		t.Fatal("rebound cast did not resolve once into the graveyard")
	}

	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	if g.Stack.Size() != 0 {
		t.Fatal("rebound triggered again at a later upkeep")
	}
}

// TestStifleCountersReboundTrigger proves the rebound trigger is an ordinary
// triggered ability: countering it casts nothing and the card stays exiled with
// no later chance.
func TestStifleCountersReboundTrigger(t *testing.T) {
	t.Parallel()
	f := newReboundTriggerFixture(t)
	g := f.game
	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	trigger := requireReboundTriggerOnTop(t, g, f.cardID)

	stifleID := addCardToHand(g, game.Player2, f.def(t, "Stifle"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !f.engine.applyAction(g, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(trigger.ID)}, 0, nil)) {
		t.Fatal("Stifle could not target the rebound trigger")
	}
	f.engine.resolveTopOfStackWithChoices(g, noChoiceAgents(t), &TurnLog{})
	if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(f.cardID) ||
		controlledElephants(g, game.Player1) != 1 || reboundTriggerPending(g, f.cardID) {
		t.Fatalf("Stifle did not counter the rebound trigger: stack=%+v", g.Stack.Objects())
	}
	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(f.cardID) {
		t.Fatal("countered rebound trigger got a later chance")
	}
}

// TestReboundTriggerIgnoresCardThatLeftExile proves a card that leaves exile
// while the trigger is on the stack is a new object the trigger can't cast,
// even after it returns to exile (CR 603.7c, 400.7).
func TestReboundTriggerIgnoresCardThatLeftExile(t *testing.T) {
	t.Parallel()
	f := newReboundTriggerFixture(t)
	g := f.game
	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	requireReboundTriggerOnTop(t, g, f.cardID)
	if !moveCardBetweenZones(g, game.Player1, f.cardID, zone.Exile, zone.Hand) ||
		!moveCardBetweenZones(g, game.Player1, f.cardID, zone.Hand, zone.Exile) {
		t.Fatal("could not move Trumpeting Herd out of exile and back")
	}
	f.engine.resolveTopOfStackWithChoices(g, noChoiceAgents(t), &TurnLog{})
	if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(f.cardID) ||
		controlledElephants(g, game.Player1) != 1 {
		t.Fatal("rebound trigger cast a card that had left exile")
	}
}

// TestReboundTriggerOrderedWithOtherUpkeepTrigger proves the rebound trigger
// and Phyrexian Arena's upkeep trigger go on the stack together in the order
// their controller chooses (CR 603.3b).
func TestReboundTriggerOrderedWithOtherUpkeepTrigger(t *testing.T) {
	t.Parallel()
	for _, reboundOnTop := range []bool{true, false} {
		name := "Arena resolves first"
		if reboundOnTop {
			name = "rebound resolves first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f := newReboundTriggerFixture(t)
			g := f.game
			arena := addCombatPermanent(g, game.Player1, f.def(t, "Phyrexian Arena"))
			addCardToLibrary(g, game.Player1, f.def(t, "Stifle"))
			var requests []game.ChoiceRequest
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Prompt != "Order triggered abilities." {
					return request.DefaultSelection
				}
				requests = append(requests, request)
				// Abilities are put on the stack in the chosen order, so the
				// last chosen resolves first.
				order := make([]int, 0, len(request.Options))
				var last int
				for _, option := range request.Options {
					isArena := option.Label == fmt.Sprintf("source=%v ability=0", arena.ObjectID)
					if isArena == reboundOnTop {
						order = append(order, option.Index)
					} else {
						last = option.Index
					}
				}
				return append(order, last)
			}}
			beginUpkeep(f.engine, g, game.Player1, agents)
			if len(requests) != 1 || len(requests[0].Options) != 2 || g.Stack.Size() != 2 {
				t.Fatalf("order requests = %+v stack = %+v, want both triggers ordered once",
					requests, g.Stack.Objects())
			}
			top, _ := g.Stack.Peek()
			if top.Rebound != reboundOnTop {
				t.Fatalf("stack top rebound = %v, want %v", top.Rebound, reboundOnTop)
			}
			lifeBefore := g.Players[game.Player1].Life
			handBefore := g.Players[game.Player1].Hand.Size()
			if reboundOnTop {
				requireReboundTriggerOnTop(t, g, f.cardID)
				f.engine.resolveTopOfStack(g, &TurnLog{})
				// The recast spell resolves before Arena's trigger.
				spell, _ := g.Stack.Peek()
				if g.Stack.Size() != 2 || spell.SourceID != f.cardID || spell.Kind != game.StackSpell {
					t.Fatalf("stack = %+v, want Trumpeting Herd above Arena's trigger", g.Stack.Objects())
				}
				f.engine.resolveTopOfStack(g, &TurnLog{})
				if controlledElephants(g, game.Player1) != 2 || g.Players[game.Player1].Life != lifeBefore {
					t.Fatal("Trumpeting Herd did not resolve before Arena's trigger")
				}
				f.engine.resolveTopOfStack(g, &TurnLog{})
			} else {
				f.engine.resolveTopOfStack(g, &TurnLog{})
				if g.Players[game.Player1].Life != lifeBefore-1 || controlledElephants(g, game.Player1) != 1 {
					t.Fatal("Arena's trigger did not resolve before the rebound trigger")
				}
				requireReboundTriggerOnTop(t, g, f.cardID)
				f.engine.resolveTopOfStack(g, &TurnLog{})
				f.engine.resolveTopOfStack(g, &TurnLog{})
			}
			if g.Stack.Size() != 0 || controlledElephants(g, game.Player1) != 2 ||
				g.Players[game.Player1].Life != lifeBefore-1 || g.Players[game.Player1].Hand.Size() != handBefore+1 {
				t.Fatal("both upkeep triggers did not resolve exactly once")
			}
		})
	}
}

// TestReboundTriggerCloneAndReplay proves the pending and stacked rebound
// trigger are game-owned: clones resolve independently of the original and
// replay the same outcome.
func TestReboundTriggerCloneAndReplay(t *testing.T) {
	t.Parallel()
	f := newReboundTriggerFixture(t)
	g := f.game
	pending := g.Clone()

	beginUpkeep(f.engine, g, game.Player1, noChoiceAgents(t))
	trigger := requireReboundTriggerOnTop(t, g, f.cardID)
	stacked := g.Clone()

	// Declining in the stacked clone leaves the original's trigger untouched.
	decline := [game.NumPlayers]PlayerAgent{game.Player1: &cascadeCastChoiceAgent{decline: true}}
	f.engine.resolveTopOfStackWithChoices(stacked, decline, &TurnLog{})
	if stacked.Stack.Size() != 0 || !stacked.Players[game.Player1].Exile.Contains(f.cardID) {
		t.Fatal("declined rebound trigger cast the card in the clone")
	}
	if top, ok := g.Stack.Peek(); !ok || top.ID != trigger.ID || !top.Rebound {
		t.Fatal("clone's resolution changed the original's stacked trigger")
	}

	// The original and the replayed pending clone reach the same outcome.
	resolveReboundTriggers(f.engine, g, [game.NumPlayers]PlayerAgent{})
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if !reboundTriggerPending(pending, f.cardID) || pending.Stack.Size() != 0 {
		t.Fatal("original's upkeep changed the pending clone")
	}
	runReboundUpkeep(f.engine, pending, game.Player1, [game.NumPlayers]PlayerAgent{})
	f.engine.resolveTopOfStack(pending, &TurnLog{})
	for _, replay := range []*game.Game{g, pending} {
		if controlledElephants(replay, game.Player1) != 2 || !replay.Players[game.Player1].Graveyard.Contains(f.cardID) ||
			replay.Stack.Size() != 0 || reboundTriggerPending(replay, f.cardID) {
			t.Fatal("rebound trigger outcome differs between original and clone")
		}
	}
}

// TestReboundTriggerThroughUpkeepStep runs the real upkeep step with passing
// players: the trigger goes on the stack in the upkeep priority window, its
// recast resolves and only then does the step end.
func TestReboundTriggerThroughUpkeepStep(t *testing.T) {
	t.Parallel()
	f := newReboundTriggerFixture(t)
	g := f.game
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.TriggerEventCursor = len(g.Events)
	f.engine.runUpkeepStep(g, simPassPolicies(), &TurnLog{})
	if g.Stack.Size() != 0 || controlledElephants(g, game.Player1) != 2 ||
		!g.Players[game.Player1].Graveyard.Contains(f.cardID) || reboundTriggerPending(g, f.cardID) {
		t.Fatalf("upkeep step did not resolve the rebound trigger and recast: stack=%+v", g.Stack.Objects())
	}
}

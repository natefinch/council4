package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Engine gap #3629 (sub-issue of #3625). Default-registry cards
// checked against the pinned corpus 6da9ced0…a1b4 (oracle_id): Aether
// Spellbomb 4b033a0a-c1ae-44d7-9662-72cbbfda024b {1} "{U}, Sacrifice this
// artifact: Return target creature to its owner's hand. {1}, Sacrifice this
// artifact: Draw a card." (no rulings); Raugrin Crystal
// ae4a8470-4517-42c7-a7f6-ef7635e10fe5 "{T}: Add {U}, {R}, or {W}. Cycling {2}
// ({2}, Discard this card: Draw a card.)" (no rulings); Reassembling Skeleton
// 9dbc3530-b278-4c8d-b2cc-a09dfac9d5e5 "{1}{B}: Return this card from your
// graveyard to the battlefield tapped."; Waterlogged Grove
// 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa and Silent Clearing
// fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life: Add {G} or {U}." /
// "{W} or {B}." (payment-time mana abilities the planner never activates on
// its own); Grizzly Bears 14c8f55d-d177-4c25-a931-ebeb9e6062a0; Future Sight
// 93a0e938-b916-47ec-9803-369e885e253c "Play with the top card of your library
// revealed. …". Millikin fa4dffda-6f04-4d0b-829d-28a1a5794dee, named by the
// issue, is deliberately not used: its pinned Oracle text reads "{T}, Mill a
// card: Add {C}. (Activate only as an instant. …)" and its 2026-08-05 ruling
// says "Millikin's ability is not a mana ability. … it can't be activated in
// the middle of paying for a cost.", the engine's classification of it being
// #3690. Comprehensive Rules effective 2026-09-25: 602.2b "The remainder of
// the process for activating an ability is identical to the process for
// casting a spell listed in rules 601.2b–i. … An activated ability's analog to
// a spell's mana cost (as referenced in rule 601.2f) is its activation cost.";
// 601.2f the total cost becomes locked in; 601.2g "If the total cost includes
// a mana payment, the player then has a chance to activate mana abilities …
// Mana abilities must be activated before costs are paid."; 601.2h "Partial
// payments are not allowed."; 605.1a a mana ability's "cost and effect don't
// move any card to or from a library"; 605.3a a mana ability may be activated
// "whenever they are … activating an ability that requires a mana payment";
// 401.5 a changed library top stays hidden until "the ability becomes
// activated"; 733.1 "If a player … starts to take an action but can't legally
// complete it, the entire action is reversed and any payments already made are
// canceled. … Each player may also reverse any legal mana abilities that
// player activated while making the illegal play …"; 733.2 "the player who had
// priority retains it and may take another action or pass. The player may
// redo the reversed action in a legal way …".

// activationWindowRoute activates an ability from one zone a source can
// function from, with a two-land set of payment-time mana abilities in lands:
// lands[:len(lands)-1] can't pay it while all of them can, by activating the
// abilities in order while paying.
type activationWindowRoute struct {
	name  string
	lands []string
	order []string
	// build adds the source and anything its activation needs to g, returning
	// the ID its activations name as their source.
	build func(t *testing.T, g *game.Game) id.ID
	// activation reports whether act is the route's activation of source.
	activation func(act action.Action, source id.ID) bool
	// done checks the activation is on the stack with its costs paid; undone
	// checks nothing of it remains.
	done, undone func(t *testing.T, g *game.Game, source id.ID)
}

// activationOf matches an activation of source's ability with targets
// targets.
func activationOf(targets int) func(action.Action, id.ID) bool {
	return func(act action.Action, source id.ID) bool {
		activate, ok := act.ActivateAbilityPayload()
		return ok && activate.SourceID == source && len(activate.Targets) == targets
	}
}

// requireActivatedFrom checks the top of the stack is an activated ability of
// the card cardID.
func requireActivatedFrom(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceCardID != cardID {
		t.Fatalf("stack top = %+v, want an ability activated from %v", obj, cardID)
	}
}

func requireEmptyStack(t *testing.T, g *game.Game) {
	t.Helper()
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want the activation reversed", g.Stack.Size())
	}
}

func activationWindowRoutes() []activationWindowRoute {
	spellbomb := func(t *testing.T, g *game.Game) id.ID {
		addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
		return addCombatPermanent(g, game.Player1, registryCardDef(t, "Aether Spellbomb")).ObjectID
	}
	return []activationWindowRoute{
		{
			name:       "battlefield Aether Spellbomb",
			lands:      []string{"Silent Clearing", "Waterlogged Grove"},
			order:      []string{groveMana},
			build:      spellbomb,
			activation: activationOf(1),
			done: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				if _, ok := permanentByObjectID(g, source); ok {
					t.Fatal("Aether Spellbomb wasn't sacrificed to pay the cost")
				}
				obj, ok := g.Stack.Peek()
				if !ok || obj.SourceID != source || len(obj.Targets) != 1 {
					t.Fatalf("stack top = %+v, want Spellbomb's targeted ability", obj)
				}
			},
			undone: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				if _, ok := permanentByObjectID(g, source); !ok {
					t.Fatal("Aether Spellbomb left the battlefield, want the activation reversed")
				}
				requireEmptyStack(t, g)
			},
		},
		{
			name:  "hand cycling Raugrin Crystal",
			lands: []string{"Waterlogged Grove", "Silent Clearing"},
			order: []string{groveMana, clearingMana},
			build: func(t *testing.T, g *game.Game) id.ID {
				return addCardToHand(g, game.Player1, registryCardDef(t, "Raugrin Crystal"))
			},
			activation: activationOf(0),
			done: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				if !g.Players[game.Player1].Graveyard.Contains(source) {
					t.Fatal("Raugrin Crystal wasn't discarded to pay the cycling cost")
				}
				requireActivatedFrom(t, g, source)
			},
			undone: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				if !g.Players[game.Player1].Hand.Contains(source) {
					t.Fatal("Raugrin Crystal left the hand, want the cycling reversed")
				}
				requireEmptyStack(t, g)
			},
		},
		{
			name:  "graveyard Reassembling Skeleton",
			lands: []string{"Waterlogged Grove", "Silent Clearing"},
			order: []string{groveMana, clearingMana},
			build: func(t *testing.T, g *game.Game) id.ID {
				return addCardToGraveyard(g, game.Player1, registryCardDef(t, "Reassembling Skeleton"))
			},
			activation: activationOf(0),
			done: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				requireActivatedFrom(t, g, source)
			},
			undone: func(t *testing.T, g *game.Game, source id.ID) {
				t.Helper()
				if !g.Players[game.Player1].Graveyard.Contains(source) {
					t.Fatal("Reassembling Skeleton left the graveyard")
				}
				requireEmptyStack(t, g)
			},
		},
	}
}

// findRouteActivation is route's activation of source among player1's legal
// actions.
func findRouteActivation(e *Engine, g *game.Game, route activationWindowRoute, source id.ID) (action.Action, bool) {
	for _, act := range e.legalActions(g, game.Player1) {
		if route.activation(act, source) {
			return act, true
		}
	}
	return action.Action{}, false
}

// activationWindowAgent activates mana abilities while paying as its
// manaOrderAgent does, choosing blue or black mana, and records each
// payment-time mana window it was offered, whether Player1's unchanged
// library top (top) stayed revealed during it, and the deepest nesting of
// library-top windows at any choice.
type activationWindowAgent struct {
	manaOrderAgent
	top                       id.ID
	windows, hidden, maxDepth int
}

func newActivationWindowAgent(order []string, top id.ID) *activationWindowAgent {
	return &activationWindowAgent{manaOrderAgent: manaOrderAgent{order: order, colors: []string{"U", "B"}}, top: top}
}

func (a *activationWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		a.windows++
		if topCardShown(NewObservation(obs.g, game.Player1), game.Player1) != a.top {
			a.hidden++
		}
	}
	a.maxDepth = max(a.maxDepth, len(obs.g.LibraryTopWindows))
	return a.manaOrderAgent.ChooseChoice(obs, request)
}

// TestActivationPaysWithPaymentTimeManaAbilities activates, under Future
// Sight, an ability from each zone a source functions from: Aether Spellbomb's
// targeted ability on the battlefield ({U}), Raugrin Crystal's cycling from
// hand ({2}) and Reassembling Skeleton's from the graveyard ({1}{B}). The only
// mana sources are lands whose mana abilities the payment planner never
// activates on its own. Without the last land the activation isn't offered;
// with it, it is, and the lands are activated in the payment-time mana window
// once the total cost is locked (CR 602.2b, 601.2f-g, 605.3a) and pay it (CR
// 601.2h). Each land's color choice is made with the land's own window nested
// in the activation's, and the unchanged revealed top card stays revealed
// throughout (CR 401.5 withholds only a changed top).
func TestActivationPaysWithPaymentTimeManaAbilities(t *testing.T) {
	t.Parallel()
	for _, route := range activationWindowRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			source := route.build(t, g)
			for _, land := range route.lands[:len(route.lands)-1] {
				addCombatPermanent(g, game.Player1, registryCardDef(t, land))
			}
			if _, ok := findRouteActivation(e, g, route, source); ok {
				t.Fatal("activation offered without enough mana abilities to pay it")
			}
			addCombatPermanent(g, game.Player1, registryCardDef(t, route.lands[len(route.lands)-1]))
			act, ok := findRouteActivation(e, g, route, source)
			if !ok {
				t.Fatal("activation payable with payment-time mana abilities wasn't offered")
			}
			life := g.Players[game.Player1].Life
			agent := newActivationWindowAgent(route.order, top)

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied {
				t.Fatalf("outcome = %v, want the ability activated", outcome)
			}
			route.done(t, g, source)
			if agent.activated != len(route.order) || agent.windows < len(route.order) {
				t.Fatalf("activated %d mana abilities over %d window prompts, want %d", agent.activated, agent.windows, len(route.order))
			}
			p := g.Players[game.Player1]
			if p.ManaPool.Total() != 0 || p.Life != life-len(route.order) {
				t.Fatalf("pool = %d, life = %d; want the activated lands' mana spent and %d life paid", p.ManaPool.Total(), p.Life, len(route.order))
			}
			if agent.hidden != 0 || agent.maxDepth != 2 {
				t.Fatalf("top hidden at %d window prompts, deepest library-top windows = %d; want the unchanged top revealed and the land's window nested in the activation's", agent.hidden, agent.maxDepth)
			}
			assertTopShown(t, g, "Future Sight", top)
		})
	}
}

// TestOptionalActivationManaWindow activates Aether Spellbomb's "{1},
// Sacrifice this artifact: Draw a card." with {C} already in the pool and
// Waterlogged Grove untapped. The cost is payable from the pool, but the cost
// includes a mana payment, so the payment-time mana window is still offered
// (CR 602.2b, 601.2g): declining it pays from the pool, while activating the
// Grove pays with either mana and leaves the other.
func TestOptionalActivationManaWindow(t *testing.T) {
	t.Parallel()
	for _, activate := range []bool{false, true} {
		name := "declined"
		if activate {
			name = "Grove activated"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			spellbomb := addCombatPermanent(g, game.Player1, registryCardDef(t, "Aether Spellbomb")).ObjectID
			grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			p := g.Players[game.Player1]
			p.ManaPool.Add(mana.C, 1)
			life := p.Life
			act, ok := findRouteActivation(e, g, activationWindowRoute{activation: activationOf(0)}, spellbomb)
			if !ok {
				t.Fatal("Spellbomb's draw ability wasn't offered with {C} in the pool")
			}
			var order []string
			if activate {
				order = []string{groveMana}
			}
			agent := newActivationWindowAgent(order, top)

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied {
				t.Fatalf("outcome = %v, want Spellbomb's draw ability activated", outcome)
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != spellbomb || len(obj.Targets) != 0 {
				t.Fatalf("stack top = %+v, want Spellbomb's draw ability", obj)
			}
			if _, ok := permanentByObjectID(g, spellbomb); ok {
				t.Fatal("Aether Spellbomb wasn't sacrificed")
			}
			if agent.windows == 0 {
				t.Fatal("the payment-time mana window wasn't offered for a cost payable from the pool")
			}
			wantPool, wantLife := 0, life
			if activate {
				wantPool, wantLife = 1, life-1
			}
			if grove.Tapped != activate || p.ManaPool.Total() != wantPool || p.Life != wantLife {
				t.Fatalf("Grove tapped = %t, pool = %d, life = %d; want %t, %d, %d", grove.Tapped, p.ManaPool.Total(), p.Life, activate, wantPool, wantLife)
			}
			assertTopShown(t, g, "Future Sight", top)
		})
	}
}

// TestUnpaidActivationIsReversed begins each route's activation, offered
// because both lands together could pay it, but activates all but the last
// of the mana abilities it needs while paying. The locked cost can't then be
// paid, so the activation is reversed (CR 733.1): nothing of the activation
// remains, and the lands' mana abilities are reversed with it, untapping them
// and restoring the life paid. A repeated reversed activation is exhausted
// (CR 733.2 lets the player redo it, but a deterministic player moves on).
func TestUnpaidActivationIsReversed(t *testing.T) {
	t.Parallel()
	for _, route := range activationWindowRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			source := route.build(t, g)
			for _, land := range route.lands {
				addCombatPermanent(g, game.Player1, registryCardDef(t, land))
			}
			act, ok := findRouteActivation(e, g, route, source)
			if !ok {
				t.Fatal("activation payable with both lands wasn't offered")
			}
			activate, _ := act.ActivateAbilityPayload()
			life := g.Players[game.Player1].Life
			attempts := &castAttempts{}

			for attempt := range 2 {
				agent := newActivationWindowAgent(route.order[:len(route.order)-1], top)
				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)
				if outcome != actionReversed || agent.activated != len(route.order)-1 {
					t.Fatalf("attempt %d: outcome = %v after %d mana abilities, want the activation reversed", attempt, outcome, agent.activated)
				}
				attempts.reverseActivationAttempt(activate)
				route.undone(t, g, source)
				requireLandsUnused(t, g, life, route.lands...)
				if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
					t.Fatalf("pool = %d, want the reversed mana gone", got)
				}
				assertTopShown(t, g, "Future Sight", top)
			}
			if !attempts.actionExhausted(act) {
				t.Fatal("activation reversed twice isn't exhausted")
			}
		})
	}
}

// activationLoopAgent begins one activation whenever it is offered and never
// activates a mana ability while paying.
type activationLoopAgent struct {
	route      activationWindowRoute
	source     id.ID
	activation int
}

func (a *activationLoopAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if a.route.activation(act, a.source) {
			a.activation++
			return act
		}
	}
	return action.Pass()
}

func (*activationLoopAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

// TestActivationReversalsInPriorityLoop runs a player who begins each route's
// activation whenever it is offered but declines every payment-time mana
// window, with only the lands to pay. Each unpaid activation is reversed and
// the player keeps priority (CR 733.2); once the repeat is exhausted the
// player passes, and the step ends with nothing activated or paid.
func TestActivationReversalsInPriorityLoop(t *testing.T) {
	t.Parallel()
	for _, route := range activationWindowRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, _, _ := libraryTopActionGame(t, "Future Sight", nil)
			source := route.build(t, g)
			for _, land := range route.lands {
				addCombatPermanent(g, game.Player1, registryCardDef(t, land))
			}
			life := g.Players[game.Player1].Life
			agent := &activationLoopAgent{route: route, source: source}

			e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})

			if agent.activation < 2 {
				t.Fatalf("activations begun = %d, want repeated reversed attempts", agent.activation)
			}
			route.undone(t, g, source)
			requireLandsUnused(t, g, life, route.lands...)
		})
	}
}

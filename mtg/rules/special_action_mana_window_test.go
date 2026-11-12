package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Engine gap #3630. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Saw It Coming
// 90edaf33-d0ab-47e0-8f6a-6fba38286e6e {1}{U}{U} "Counter target spell.
// Foretell {1}{U} (During your turn, you may pay {2} and exile this card from
// your hand face down. Cast it on a later turn for its foretell cost.)";
// Slickshot Show-Off 74c7a029-59aa-4511-8e4c-f8f167cda4e2 (Plot {1}{R});
// Willbender 0aae277e-e58e-4115-b5fd-0459451e17ec (Morph {1}{U}); Waterlogged
// Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa and Silent Clearing
// fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life: Add {G} or {U}." /
// "{W} or {B}." (payment-time mana abilities the planner never activates on
// its own); Phyrexian Altar 8d02b297-97c4-4379-9862-0a462400f66f "Sacrifice a
// creature: Add one mana of any color."; Mana Confluence
// d0ee5bdc-2b69-4b73-9a20-ffcc18783b29 "{T}, Pay 1 life: Add one mana of any
// color."; Future Sight 93a0e938-b916-47ec-9803-369e885e253c "Play with the
// top card of your library revealed. …". The suspend route uses the test card
// suspendSorcery. Millikin fa4dffda-6f04-4d0b-829d-28a1a5794dee, named by the
// issue, isn't used: under 605.1a and its 2026-08-05 ruling its ability is not
// a mana ability (#3690). Comprehensive Rules effective 2026-09-25: 116.1
// special actions; 116.2b turning a face-down permanent face up, 116.2f
// suspend, 116.2h foretell and 116.2k plot are special actions; 116.3 the
// player receives priority afterwards; 605.3a "A player may activate an
// activated mana ability whenever they have priority, whenever they are
// casting a spell or activating an ability that requires a mana payment, or
// whenever a rule or effect asks for a mana payment (even if it's in the
// middle of casting or resolving a spell or activating or resolving an
// ability)."; 702.37e, 701.40b, 701.58b turning face up for its cost;
// 702.62a, 702.143a, 702.170a the suspend, foretell and plot costs; 601.2h
// "Partial payments are not allowed."; 401.5 "If the top card of the player's
// library changes while a player is taking a special action …, the new card
// won't be revealed and can't be looked at until the player has finished
// taking that special action."; 733.1 "If a player … starts to take an action
// but can't legally complete it, the entire action is reversed and any
// payments already made are canceled. … Each player may also reverse any
// legal mana abilities that player activated while making the illegal play
// …"; 733.2 the player who had priority retains it.

// specialActionRoute takes one special action with a mana cost. With pool's
// mana floating, lands[:len(lands)-1] can't pay its cost while all of them
// can, by activating their mana abilities in order while paying.
type specialActionRoute struct {
	name  string
	pool  []mana.Color
	lands []string
	order []string
	// build adds the card or permanent the action uses and returns the action.
	build func(t *testing.T, g *game.Game) action.Action
	// done and undone check that the action was, or wasn't, taken.
	done, undone func(t *testing.T, g *game.Game, act action.Action)
}

func specialActionRoutes() []specialActionRoute {
	handCard := func(act action.Action) (id.ID, bool) {
		switch act.Kind {
		case action.ActionForetellCard:
			p, ok := act.ForetellCardPayload()
			return p.CardID, ok
		case action.ActionPlotCard:
			p, ok := act.PlotCardPayload()
			return p.CardID, ok
		case action.ActionSuspendCard:
			p, ok := act.SuspendCardPayload()
			return p.CardID, ok
		}
		return 0, false
	}
	exiled := func(t *testing.T, g *game.Game, act action.Action, recorded func(*game.Game, id.ID) bool) {
		t.Helper()
		card, _ := handCard(act)
		p := g.Players[game.Player1]
		if p.Hand.Contains(card) || !p.Exile.Contains(card) || !recorded(g, card) {
			t.Fatal("the card wasn't exiled from hand by the special action")
		}
	}
	inHand := func(t *testing.T, g *game.Game, act action.Action, recorded func(*game.Game, id.ID) bool) {
		t.Helper()
		card, _ := handCard(act)
		p := g.Players[game.Player1]
		if !p.Hand.Contains(card) || p.Exile.Contains(card) || recorded(g, card) {
			t.Fatal("the reversed special action's card isn't back in hand")
		}
	}
	foretold := func(g *game.Game, c id.ID) bool { _, ok := g.ForetoldCards[c]; return ok }
	plotted := func(g *game.Game, c id.ID) bool { _, ok := g.PlottedCards[c]; return ok }
	suspended := func(g *game.Game, c id.ID) bool {
		return g.ExileCounterCount(c, counter.Time) == 3 && cardIsSuspended(g, c)
	}
	faceUp := func(t *testing.T, g *game.Game, act action.Action, want bool) {
		t.Helper()
		p, _ := act.TurnFaceUpPayload()
		willbender, ok := permanentByObjectID(g, p.PermanentID)
		if !ok || willbender.FaceDown == want || (want && willbender.FaceDownKind != game.FaceDownNone) {
			t.Fatalf("Willbender face up = %t, want %t", ok && !willbender.FaceDown, want)
		}
	}
	return []specialActionRoute{
		{
			name:  "foretell Saw It Coming",
			lands: []string{"Waterlogged Grove", "Silent Clearing"},
			order: []string{groveMana, clearingMana},
			build: func(t *testing.T, g *game.Game) action.Action {
				return action.ForetellCard(addCardToHand(g, game.Player1, registryCardDef(t, "Saw It Coming")))
			},
			done: func(t *testing.T, g *game.Game, act action.Action) {
				exiled(t, g, act, foretold)
				card, _ := handCard(act)
				if !g.Players[game.Player1].Exile.IsFaceDown(card) {
					t.Fatal("the foretold card wasn't exiled face down")
				}
			},
			undone: func(t *testing.T, g *game.Game, act action.Action) { inHand(t, g, act, foretold) },
		},
		{
			name:  "plot Slickshot Show-Off",
			pool:  []mana.Color{mana.R},
			lands: []string{"Waterlogged Grove"},
			order: []string{groveMana},
			build: func(t *testing.T, g *game.Game) action.Action {
				return action.PlotCard(addCardToHand(g, game.Player1, registryCardDef(t, "Slickshot Show-Off")))
			},
			done:   func(t *testing.T, g *game.Game, act action.Action) { exiled(t, g, act, plotted) },
			undone: func(t *testing.T, g *game.Game, act action.Action) { inHand(t, g, act, plotted) },
		},
		{
			name:  "suspend",
			lands: []string{"Waterlogged Grove", "Silent Clearing"},
			order: []string{groveMana, clearingMana},
			build: func(_ *testing.T, g *game.Game) action.Action {
				return action.SuspendCard(addCardToHand(g, game.Player1, suspendSorcery(3, cost.Mana{cost.U, cost.B})))
			},
			done:   func(t *testing.T, g *game.Game, act action.Action) { exiled(t, g, act, suspended) },
			undone: func(t *testing.T, g *game.Game, act action.Action) { inHand(t, g, act, suspended) },
		},
		{
			name:  "turn Willbender face up",
			lands: []string{"Waterlogged Grove", "Silent Clearing"},
			order: []string{groveMana, clearingMana},
			build: func(t *testing.T, g *game.Game) action.Action {
				return action.TurnFaceUp(addFaceDownWillbender(t, g).ObjectID)
			},
			done:   func(t *testing.T, g *game.Game, act action.Action) { faceUp(t, g, act, true) },
			undone: func(t *testing.T, g *game.Game, act action.Action) { faceUp(t, g, act, false) },
		},
	}
}

func addFaceDownWillbender(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	willbender := addCombatPermanent(g, game.Player1, registryCardDef(t, "Willbender"))
	willbender.FaceDown, willbender.FaceDownKind, willbender.FaceDownFace = true, game.FaceDownMorph, game.FaceFront
	return willbender
}

func specialActionOffered(e *Engine, g *game.Game, want action.Action) bool {
	for _, act := range e.legalActions(g, game.Player1) {
		if reflect.DeepEqual(act, want) {
			return true
		}
	}
	return false
}

func (r specialActionRoute) setup(t *testing.T, g *game.Game, lands []string) action.Action {
	t.Helper()
	for _, color := range r.pool {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	for _, land := range lands {
		addCombatPermanent(g, game.Player1, registryCardDef(t, land))
	}
	return r.build(t, g)
}

// TestSpecialActionPaysWithPaymentTimeManaAbilities takes, under Future
// Sight, each special action with a mana cost — foretelling Saw It Coming
// ({2}), plotting Slickshot Show-Off ({1}{R} with {R} floating), suspending a
// card ({U}{B}) and turning a face-down Willbender face up ({1}{U}). The only
// other mana sources are lands whose mana abilities the payment planner never
// activates on its own. Without the last land the action isn't offered; with
// it, it is, and the lands are activated while paying (CR 605.3a) and pay the
// cost in full (CR 601.2h). Each land's color choice is made with the land's
// own window nested in the action's, and the unchanged revealed top card
// stays revealed throughout (CR 401.5 withholds only a changed top).
func TestSpecialActionPaysWithPaymentTimeManaAbilities(t *testing.T) {
	t.Parallel()
	for _, route := range specialActionRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			act := route.setup(t, g, route.lands[:len(route.lands)-1])
			if specialActionOffered(e, g, act) {
				t.Fatal("special action offered without enough mana abilities to pay it")
			}
			addCombatPermanent(g, game.Player1, registryCardDef(t, route.lands[len(route.lands)-1]))
			if !specialActionOffered(e, g, act) {
				t.Fatal("special action payable with payment-time mana abilities wasn't offered")
			}
			life := g.Players[game.Player1].Life
			agent := newActivationWindowAgent(route.order, top)

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied {
				t.Fatalf("outcome = %v, want the special action taken", outcome)
			}
			route.done(t, g, act)
			if agent.activated != len(route.order) || agent.windows < len(route.order) {
				t.Fatalf("activated %d mana abilities over %d window prompts, want %d", agent.activated, agent.windows, len(route.order))
			}
			p := g.Players[game.Player1]
			if p.ManaPool.Total() != 0 || p.Life != life-len(route.order) {
				t.Fatalf("pool = %d, life = %d; want the lands' mana spent and %d life paid", p.ManaPool.Total(), p.Life, len(route.order))
			}
			if agent.hidden != 0 || agent.maxDepth != 2 {
				t.Fatalf("top hidden at %d window prompts, deepest library-top windows = %d; want the unchanged top revealed and the land's window nested in the action's", agent.hidden, agent.maxDepth)
			}
			if g.Stack.Size() != 0 || g.Turn.PriorityPlayer != game.Player1 {
				t.Fatal("special action used the stack or passed priority (CR 116.1, 116.3)")
			}
			assertTopShown(t, g, "Future Sight", top)
		})
	}
}

// TestOptionalSpecialActionManaWindow foretells Saw It Coming with {C}{C}
// already floating and Waterlogged Grove untapped. The {2} is payable from
// the pool, but the cost is a mana payment, so the payment-time mana window
// is still offered (CR 605.3a): declining it pays from the pool, while
// activating the Grove pays with either mana and leaves one.
func TestOptionalSpecialActionManaWindow(t *testing.T) {
	t.Parallel()
	for _, activate := range []bool{false, true} {
		name := "declined"
		if activate {
			name = "Grove activated"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			p := g.Players[game.Player1]
			p.ManaPool.Add(mana.C, 2)
			life := p.Life
			sawItComing := addCardToHand(g, game.Player1, registryCardDef(t, "Saw It Coming"))
			var order []string
			if activate {
				order = []string{groveMana}
			}
			agent := newActivationWindowAgent(order, top)

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.ForetellCard(sawItComing), reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied || !p.Exile.Contains(sawItComing) {
				t.Fatalf("outcome = %v, want Saw It Coming foretold", outcome)
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

// TestUnpaidSpecialActionIsReversed begins each route's special action,
// offered because all its lands could pay it, but activates all but the last
// of the mana abilities it needs while paying. The cost can't then be paid
// (CR 601.2h), so the action is reversed (CR 733.1): the card stays in hand or
// the permanent face down, and the lands' mana abilities are reversed with it,
// untapping them and restoring the life paid. A repeated reversed action is
// exhausted (CR 733.2 lets the player redo it, but a deterministic player
// moves on).
func TestUnpaidSpecialActionIsReversed(t *testing.T) {
	t.Parallel()
	for _, route := range specialActionRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
			act := route.setup(t, g, route.lands)
			if !specialActionOffered(e, g, act) {
				t.Fatal("special action payable with its lands wasn't offered")
			}
			life := g.Players[game.Player1].Life
			pool := g.Players[game.Player1].ManaPool.Total()
			attempts := &castAttempts{}

			for attempt := range 2 {
				agent := newActivationWindowAgent(route.order[:len(route.order)-1], top)
				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, attempts)
				if outcome != actionReversed || agent.activated != len(route.order)-1 {
					t.Fatalf("attempt %d: outcome = %v after %d mana abilities, want the special action reversed", attempt, outcome, agent.activated)
				}
				route.undone(t, g, act)
				requireLandsUnused(t, g, life, route.lands...)
				if got := g.Players[game.Player1].ManaPool.Total(); got != pool {
					t.Fatalf("pool = %d, want %d with the reversed mana gone", got, pool)
				}
				assertTopShown(t, g, "Future Sight", top)
			}
			if !attempts.actionExhausted(act) {
				t.Fatal("special action reversed twice isn't exhausted")
			}
		})
	}
}

// specialActionLoopAgent begins one special action whenever it is offered and
// never activates a mana ability while paying.
type specialActionLoopAgent struct {
	want  action.Action
	begun int
}

func (a *specialActionLoopAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if reflect.DeepEqual(act, a.want) {
			a.begun++
			return act
		}
	}
	return action.Pass()
}

func (*specialActionLoopAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

// TestSpecialActionReversalsInPriorityLoop runs a player who begins each
// route's special action whenever it is offered but declines every
// payment-time mana window, with only the lands to pay. Each unpaid action is
// reversed and the player keeps priority (CR 733.2); once the repeat is
// exhausted the player passes, and the step ends with nothing taken or paid.
func TestSpecialActionReversalsInPriorityLoop(t *testing.T) {
	t.Parallel()
	for _, route := range specialActionRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, _, _ := libraryTopActionGame(t, "Future Sight", nil)
			act := route.setup(t, g, route.lands)
			life := g.Players[game.Player1].Life
			agent := &specialActionLoopAgent{want: act}

			e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})

			if agent.begun < 2 {
				t.Fatalf("special actions begun = %d, want repeated reversed attempts", agent.begun)
			}
			route.undone(t, g, act)
			requireLandsUnused(t, g, life, route.lands...)
		})
	}
}

// TestSpecialActionReversedWhenManaAbilityRemovesItsPermanent begins turning
// a face-down Willbender face up for {1}{U}, its controller's only creature,
// offered because Waterlogged Grove and Silent Clearing could pay it. While
// paying, the player activates the Grove for {U} and then Phyrexian Altar,
// sacrificing the face-down Willbender itself for {U}. The mana would pay the
// cost, but the permanent the action turns face up is gone, so the action
// can't be completed and is reversed (CR 733.1). Neither mana ability moved a
// card to or from a library, so both are reversed: Willbender is back on the
// battlefield face down, the Grove untapped and the life restored.
func TestSpecialActionReversedWhenManaAbilityRemovesItsPermanent(t *testing.T) {
	t.Parallel()
	g, e, top, _ := libraryTopActionGame(t, "Future Sight", nil)
	for _, land := range []string{"Waterlogged Grove", "Silent Clearing"} {
		addCombatPermanent(g, game.Player1, registryCardDef(t, land))
	}
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
	willbender := addFaceDownWillbender(t, g)
	act := action.TurnFaceUp(willbender.ObjectID)
	if !specialActionOffered(e, g, act) {
		t.Fatal("turning Willbender face up wasn't offered")
	}
	life := g.Players[game.Player1].Life
	agent := newActivationWindowAgent([]string{groveMana, "Sacrifice a creature"}, top)

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

	if outcome != actionReversed || agent.activated != 2 {
		t.Fatalf("outcome = %v after %d mana abilities, want the action reversed after both", outcome, agent.activated)
	}
	faceDown, ok := permanentByObjectID(g, willbender.ObjectID)
	if !ok || !faceDown.FaceDown || faceDown.FaceDownKind != game.FaceDownMorph || g.Players[game.Player1].Graveyard.Contains(willbender.CardInstanceID) {
		t.Fatal("Willbender's sacrifice wasn't reversed with the action, leaving it face down")
	}
	requireLandsUnused(t, g, life, "Waterlogged Grove", "Silent Clearing")
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("pool = %d, want the reversed mana gone", got)
	}
	assertTopShown(t, g, "Future Sight", top)
}

// TestNestedManaAbilityDuringSpecialActionKeepsLibraryTopHidden foretells
// Saw It Coming under Future Sight, paying the {2} with Phyrexian Altar and
// then Mana Confluence while paying (CR 605.3a). The Altar's sacrifice of
// Homing Golem puts the Golem on top of the library (its own replacement
// effect, so the Altar's ability is still a mana ability, CR 605.1a).
// Confluence's window nests in the foretell's after the top already changed;
// the special action's snapshot governs, so the new top stays hidden through
// Confluence's color choice and is revealed once the special action is
// finished (CR 401.5).
func TestNestedManaAbilityDuringSpecialActionKeepsLibraryTopHidden(t *testing.T) {
	t.Parallel()
	g, e, start, _ := libraryTopActionGame(t, "Future Sight", nil)
	g.Players[game.Player1].Life = 20
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
	golem := addCombatPermanent(g, game.Player1, homingGolemDef()).CardInstanceID
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Mana Confluence"))
	sawItComing := addCardToHand(g, game.Player1, registryCardDef(t, "Saw It Coming"))
	agent := &nestedManaWindowAgent{libraryTopWindowAgent: &libraryTopWindowAgent{owner: game.Player1, start: start}}

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.ForetellCard(sawItComing), reboundKickerAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied || !g.Players[game.Player1].Exile.Contains(sawItComing) {
		t.Fatalf("outcome = %v, want Saw It Coming foretold", outcome)
	}
	if top, _ := g.Players[game.Player1].Library.Top(); len(agent.activated) != 2 || top != golem {
		t.Fatalf("activated %v, want Phyrexian Altar (putting Homing Golem on top) then Mana Confluence", agent.activated)
	}
	if !agent.colorChoice || agent.depth != 2 {
		t.Fatalf("Confluence's color choice made = %t with %d windows open, want it inside the foretell's window", agent.colorChoice, agent.depth)
	}
	agent.assertHiddenWhileChanged(t)
	if g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player1].Life != 19 {
		t.Fatal("the foretell cost wasn't paid with the Altar's and Confluence's mana")
	}
	assertTopShown(t, g, "Future Sight", golem)
}

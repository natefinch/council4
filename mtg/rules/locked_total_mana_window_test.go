package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3585 (remainder of #3479). Default-registry cards
// checked against the pinned corpus: Waterlogged Grove
// 0ab6bfbd-d2e1-4c4c-9f91-6f69c5b8e3bb "{T}, Pay 1 life: Add {G} or {U}. /
// {1}, {T}, Sacrifice this land: Draw a card."; Nurturing Peatland
// 2744ac83-a79f-4042-8720-688b5adda382 "{T}, Pay 1 life: Add {B} or {G}. /
// {1}, {T}, Sacrifice this land: Draw a card."; Thalia, Guardian of Thraben
// c9f8b8fb-1cd8-450e-a1fe-892e7a323479 "First strike / Noncreature spells cost
// {1} more to cast."; Sol Ring 8ee443cc-e17a-493b-9c93-1f9e141a30e4 {1}
// "{T}: Add {C}{C}."; Conduit of Worlds 8380eb8d-d1c3-4f96-b3b9-54845188c1d1
// "{T}: Choose target nonland permanent card in your graveyard. If you haven't
// cast a spell this turn, you may cast that card. ..."; Electrodominance
// 85e29bdc-938e-4132-813e-70c5e922efaf "... You may cast a spell with mana
// value X or less from your hand without paying its mana cost."; Vault Skirge
// {1}{B/P} and Etherium Sculptor "Artifact spells you cast cost {1} less to
// cast." (see phyrexian_payment_window_test.go). Comprehensive Rules effective
// 2026-09-25: 601.2b announces how a Phyrexian symbol will be paid; 601.2f
// locks the total cost including increases; 601.2g "If the total cost includes
// a mana payment, the player then has a chance to activate mana abilities";
// 601.2h pays it; 702.35 madness, 702.62a suspend and 702.88b rebound cast
// during resolution or a trigger; 733.1 reverses a cast that can't be
// completed. Neither land's ability is activated automatically, so a locked
// total is offered only when the caster's payment-time mana abilities could
// actually pay it: one land can't make a {2} total, both can.

// lockedTotalManaAgent accepts optional casts, activates every payment-time
// mana ability whose label contains ability (default "Pay 1 life"), records
// the free-spell options it was offered, and takes each Phyrexian payment's
// default, recording it.
type lockedTotalManaAgent struct {
	ability         string
	windows         int
	activated       int
	phyrexianLabels []string
	spellLabels     [][]string
}

func (*lockedTotalManaAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *lockedTotalManaAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Prompt == "Choose a spell to cast without paying its mana cost":
		labels := make([]string, len(request.Options))
		for i, option := range request.Options {
			labels[i] = option.Label
		}
		a.spellLabels = append(a.spellLabels, labels)
	case request.Kind == game.ChoicePayment && strings.HasPrefix(request.Prompt, "Pay {") && strings.HasSuffix(request.Prompt, "/P}"):
		for _, option := range request.Options {
			if slices.Contains(request.DefaultSelection, option.Index) {
				a.phyrexianLabels = append(a.phyrexianLabels, option.Label)
			}
		}
	case request.Prompt == "Activate a mana ability while paying?":
		a.windows++
		ability := a.ability
		if ability == "" {
			ability = "Pay 1 life"
		}
		for _, option := range request.Options {
			if strings.Contains(option.Label, ability) {
				a.activated++
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// lockedTotalRoute builds a game in which casting a card through one route
// needs a locked {2} total once thalias Thalias are on the battlefield, and
// returns the cast and its outcome checks.
type lockedTotalRoute struct {
	name    string
	thalias int
	build   func(t *testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent))
	cast    func(t *testing.T, g *game.Game, cardID id.ID)
	notCast func(t *testing.T, g *game.Game, cardID id.ID)
}

func lockedTotalRoutes() []lockedTotalRoute {
	reg := cards.NewDefaultRegistry()
	withGrove := func(g *game.Game) {
		addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	}
	stackTop := func(t *testing.T, g *game.Game, cardID id.ID, isCopy bool) {
		t.Helper()
		obj, ok := g.Stack.Peek()
		if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID || obj.Copy != isCopy {
			t.Fatalf("stack top = %+v, want the free cast of %v", obj, cardID)
		}
	}
	return []lockedTotalRoute{
		{
			name: "free cast from hand", thalias: 2,
			build: func(*testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				withGrove(g)
				cardID := addCardToHand(g, game.Player1, freeCastTestInstant("Locked Total Spell"))
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
						game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{})
				}
			},
			cast: func(t *testing.T, g *game.Game, cardID id.ID) { stackTop(t, g, cardID, false) },
			notCast: func(t *testing.T, g *game.Game, cardID id.ID) {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Hand.Contains(cardID) {
					t.Fatal("the free cast from hand was offered")
				}
			},
		},
		{
			name: "free copy of an exiled card", thalias: 2,
			build: func(*testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				withGrove(g)
				cardID := addCardToExile(g, game.Player1, freeCastTestInstant("Locked Total Spell"))
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.castFreeCopyOfCard(g, game.Player1, cardID, agents, &TurnLog{})
				}
			},
			cast: func(t *testing.T, g *game.Game, cardID id.ID) {
				stackTop(t, g, cardID, true)
				if !g.Players[game.Player1].Exile.Contains(cardID) {
					t.Fatal("casting a copy moved the exiled card")
				}
			},
			notCast: func(t *testing.T, g *game.Game, cardID id.ID) {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(cardID) {
					t.Fatal("the free copy was offered")
				}
			},
		},
		{
			name: "Electrodominance spell choice", thalias: 2,
			build: func(t *testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				cardID := addCardToHand(g, game.Player1, freeCastTestInstant("Locked Total Spell"))
				electroID := addCardToHand(g, game.Player1, reg.Lookup("Electrodominance"))
				g.Players[game.Player1].ManaPool.Add(mana.R, 2)
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
				if !e.applyAction(g, game.Player1, action.CastSpell(electroID, []game.Target{game.PlayerTarget(game.Player2)}, 1, nil)) {
					t.Fatal("casting Electrodominance failed")
				}
				withGrove(g)
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				}
			},
			cast: func(t *testing.T, g *game.Game, cardID id.ID) { stackTop(t, g, cardID, false) },
			notCast: func(t *testing.T, g *game.Game, cardID id.ID) {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Hand.Contains(cardID) {
					t.Fatal("the spell was cast through Electrodominance")
				}
			},
		},
		{
			name: "rebound", thalias: 2,
			build: func(t *testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				cardID := addCardToHand(g, game.Player1, reboundTestSpell("Locked Total Rebound Spell"))
				g.Players[game.Player1].ManaPool.Add(mana.G, 1)
				castIntoReboundExile(t, e, g, cardID)
				withGrove(g)
				setReboundUpkeep(g)
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					runReboundUpkeep(e, g, game.Player1, agents)
				}
			},
			cast:    func(t *testing.T, g *game.Game, cardID id.ID) { requireReboundCast(t, g, cardID) },
			notCast: requireReboundNotCast,
		},
		{
			name: "suspend", thalias: 2,
			build: func(*testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
				withGrove(g)
				setReboundUpkeep(g)
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.processSuspendUpkeep(g, game.Player1, agents, &TurnLog{})
				}
			},
			cast:    func(t *testing.T, g *game.Game, cardID id.ID) { requireSuspendCast(t, g, cardID) },
			notCast: requireSuspendNotCast,
		},
		{
			name: "madness", thalias: 1,
			build: func(t *testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g := newEffectiveCascadeGame()
				cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{cost.O(1)}))
				withGrove(g)
				discardIntoMadness(t, e, g, cardID)
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				}
			},
			cast:    func(t *testing.T, g *game.Game, cardID id.ID) { requireMadnessCast(t, g, cardID) },
			notCast: requireMadnessNotCast,
		},
		{
			name: "paid cast through Conduit of Worlds", thalias: 1,
			build: func(t *testing.T) (*Engine, *game.Game, id.ID, func([game.NumPlayers]PlayerAgent)) {
				e, g, cardID := conduitPaidCastGame(t, "Sol Ring")
				return e, g, cardID, func(agents [game.NumPlayers]PlayerAgent) {
					e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				}
			},
			cast: func(t *testing.T, g *game.Game, cardID id.ID) { requirePaidCast(t, g, cardID, zone.Graveyard, 2) },
			notCast: func(t *testing.T, g *game.Game, cardID id.ID) {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
					t.Fatal("Sol Ring was cast through Conduit of Worlds")
				}
				if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
					t.Fatal("Conduit's can't-cast-additional lock applied without a cast")
				}
			},
		},
	}
}

// TestLockedTotalOfferedOnlyWhenManaAbilitiesCanPayIt casts on each route that
// locks its total before the payment-time mana ability window: a free cast
// from hand, a free copy of an exiled card, Electrodominance's spell choice,
// rebound, suspend, madness and a paid cast through Conduit of Worlds. Thalia
// makes each locked total {2} (CR 601.2f). Waterlogged Grove alone can make
// only one mana, so the cast is not offered and the game is unchanged: no
// window opens and the Grove stays untapped. With Nurturing Peatland too the
// abilities can make the {2}: the cast is offered, both are activated in the
// window (CR 601.2g) and the {2} is paid (CR 601.2h).
func TestLockedTotalOfferedOnlyWhenManaAbilitiesCanPayIt(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, route := range lockedTotalRoutes() {
		for _, peatland := range []bool{false, true} {
			name := route.name + "/Grove alone"
			if peatland {
				name = route.name + "/Grove and Peatland"
			}
			t.Run(name, func(t *testing.T) {
				_, g, cardID, run := route.build(t)
				for range route.thalias {
					addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
				}
				if peatland {
					addCombatPermanent(g, game.Player1, reg.Lookup("Nurturing Peatland"))
				}
				life := g.Players[game.Player1].Life
				agent := &lockedTotalManaAgent{}

				run(reboundKickerAgents(agent))

				grove := reboundPermanentNamed(t, g, "Waterlogged Grove")
				if !peatland {
					route.notCast(t, g, cardID)
					if agent.windows != 0 || agent.activated != 0 || grove.Tapped || g.Players[game.Player1].Life != life {
						t.Fatalf("windows=%d activations=%d grove tapped=%t life=%d, want no proposal and the game unchanged",
							agent.windows, agent.activated, grove.Tapped, g.Players[game.Player1].Life)
					}
					for _, labels := range agent.spellLabels {
						if slices.ContainsFunc(labels, func(label string) bool { return strings.Contains(label, "Locked Total") }) {
							t.Fatalf("spell choices = %v, want the unpayable spell not offered", agent.spellLabels)
						}
					}
					return
				}
				route.cast(t, g, cardID)
				obj, _ := g.Stack.Peek()
				if agent.activated != 2 || !grove.Tapped || !reboundPermanentNamed(t, g, "Nurturing Peatland").Tapped ||
					g.Players[game.Player1].Life != life-2 || obj.ManaSpentToCast != 2 {
					t.Fatalf("activations=%d grove tapped=%t life=%d mana spent=%d, want both lands to pay the {2} total",
						agent.activated, grove.Tapped, g.Players[game.Player1].Life, obj.ManaSpentToCast)
				}
			})
		}
	}
}

// TestPhyrexianManaDefaultsToManaOnlyWhenManaAbilitiesCanPayIt casts the
// registered Vault Skirge through Conduit of Worlds with Etherium Sculptor,
// leaving a {B/P} total. Its payment is announced before the window (CR
// 601.2b), and an announced mana payment is binding, so the announcement
// defaults to mana only when the caster's mana abilities could pay it.
// Waterlogged Grove can't make {B}, so it defaults to 2 life and the cast
// completes for life with no window and the Grove untapped. Nurturing Peatland
// can make {B}, so it defaults to {B}, and the Peatland pays it in the window
// (CR 601.2g-h).
func TestPhyrexianManaDefaultsToManaOnlyWhenManaAbilitiesCanPayIt(t *testing.T) {
	for _, tc := range []struct {
		name        string
		permanents  []string
		wantDefault string
		wantMana    int
	}{
		{name: "Grove alone", permanents: []string{"Etherium Sculptor"}, wantDefault: "Pay 2 life"},
		{name: "Grove and Peatland", permanents: []string{"Etherium Sculptor", "Nurturing Peatland"},
			wantDefault: "Pay B mana", wantMana: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g, cardID := conduitPaidCastGame(t, "Vault Skirge", tc.permanents...)
			life := g.Players[game.Player1].Life
			agent := &lockedTotalManaAgent{ability: "Add {B} or {G}"}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

			requirePaidCast(t, g, cardID, zone.Graveyard, tc.wantMana)
			if !slices.Equal(agent.phyrexianLabels, []string{tc.wantDefault}) {
				t.Fatalf("Phyrexian defaults = %v, want %q", agent.phyrexianLabels, tc.wantDefault)
			}
			grove := reboundPermanentNamed(t, g, "Waterlogged Grove")
			if tc.wantMana == 0 {
				if agent.windows != 0 || grove.Tapped || g.Players[game.Player1].Life != life-2 {
					t.Fatalf("windows=%d grove tapped=%t life=%d, want {B/P} paid with 2 life and no window",
						agent.windows, grove.Tapped, g.Players[game.Player1].Life)
				}
				return
			}
			if agent.activated != 1 || grove.Tapped || !reboundPermanentNamed(t, g, "Nurturing Peatland").Tapped ||
				g.Players[game.Player1].Life != life-1 {
				t.Fatalf("activations=%d grove tapped=%t life=%d, want the Peatland to pay the announced {B} for 1 life",
					agent.activated, grove.Tapped, g.Players[game.Player1].Life)
			}
		})
	}
}

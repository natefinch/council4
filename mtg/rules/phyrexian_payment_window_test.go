package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-09-25 (media.wizards.com
// MagicCompRules 20260925). CR 107.4f: a Phyrexian mana symbol "can be paid
// either with one mana of its color or by paying 2 life"; CR 601.2b: "If a cost
// that will be paid as the spell is being cast includes Phyrexian mana symbols,
// the player announces whether they intend to pay 2 life or a corresponding
// colored mana cost for each of those symbols"; CR 601.2f locks the total cost;
// CR 601.2g: "If the total cost includes a mana payment, the player then has a
// chance to activate mana abilities"; CR 601.2h pays the total. Scryfall
// rulings for Vault Skirge (2011-06-01): "you choose how to pay for each
// Phyrexian mana symbol at the same time you would choose modes or choose a
// value for X"; for K'rrik (2019-08-23): "K'rrik's ability doesn't modify or
// reduce costs you pay. It changes only how you may pay those costs." Oracle
// (pinned corpus): Vault Skirge 64cf5d59-7bcd-4b0b-a160-c8468d4c0f60 {1}{B/P};
// K'rrik, Son of Yawgmoth cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd "For each {B} in
// a cost, you may pay 2 life rather than pay that mana."; Conduit of Worlds
// ed14be15-8f8d-4fe3-a147-f5da8ed873bf "{T}: Choose target nonland permanent
// card in your graveyard. If you haven't cast a spell this turn, you may cast
// that card."; Etherium Sculptor 96b87445-6362-4a17-91b2-cbf203fd03fd
// "Artifact spells you cast cost {1} less to cast."; Waterlogged Grove
// 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or {U}.";
// Thalia, Guardian of Thraben "Noncreature spells cost {1} more to cast."
// These scenarios cover Engine gap #3470 on the paid, free and madness cast
// routes that open a payment-time mana ability window.

// phyrexianWindowAgent announces each Phyrexian payment as 2 life when payLife
// is set, or chooses 2 life after a mana ability window when lifeAfterWindow is
// set. It prefers kicked free casts, accepts optional casts, and activates the
// payment-time mana ability containing ability once. It records the order of
// Phyrexian payment choices, target choices and mana ability windows.
type phyrexianWindowAgent struct {
	payLife         bool
	lifeAfterWindow bool
	ability         string
	steps           []string
	activated       int
}

func (*phyrexianWindowAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *phyrexianWindowAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	pick := func(label string) []int {
		for _, option := range request.Options {
			if label != "" && strings.Contains(option.Label, label) {
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	}
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case strings.Contains(request.Prompt, "without paying its mana cost"):
		return pick("with kicker")
	case request.Prompt == "Activate a mana ability while paying?":
		a.steps = append(a.steps, "window")
		if a.activated == 0 {
			if selected := pick(a.ability); selected[0] != 0 {
				a.activated++
				return selected
			}
		}
		return request.DefaultSelection
	case request.Prompt == "Choose spell targets.":
		a.steps = append(a.steps, "target")
	case request.Kind == game.ChoicePayment && strings.HasPrefix(request.Prompt, "Pay {") && strings.HasSuffix(request.Prompt, "/P}"):
		a.steps = append(a.steps, "phyrexian")
		if a.payLife || a.lifeAfterWindow && slices.Contains(a.steps, "window") {
			return pick("Pay 2 life")
		}
	}
	return request.DefaultSelection
}

// TestPaidCastPhyrexianLifeAnnouncedBeforeManaWindow casts the registered Vault
// Skirge through Conduit of Worlds with Waterlogged Grove available. The
// Phyrexian payment is announced before any mana ability window (CR 601.2b).
// With Etherium Sculptor removing the {1} and {B/P} announced as 2 life, the
// total includes no mana payment, so no window opens and the Grove stays
// untapped (CR 601.2g). Announcing {B} opens the window. Without the Sculptor
// the {1} still opens it, and the Grove pays the {1} while life pays {B/P}.
// Each cast runs in a clone without changing the original game.
func TestPaidCastPhyrexianLifeAnnouncedBeforeManaWindow(t *testing.T) {
	for _, tc := range []struct {
		name          string
		permanents    []string
		payLife       bool
		poolB         int
		ability       string
		wantSteps     []string
		wantActivated int
		wantLifeLost  int
		wantManaSpent int
	}{
		{name: "life covers the reduced total", permanents: []string{"Etherium Sculptor"}, payLife: true, ability: "Pay 1 life",
			wantSteps: []string{"phyrexian"}, wantLifeLost: 2},
		{name: "mana announced opens the window", permanents: []string{"Etherium Sculptor"}, poolB: 1, ability: "no such ability",
			wantSteps: []string{"phyrexian", "window"}, wantManaSpent: 1},
		{name: "life with a generic mana payment", payLife: true, ability: "Pay 1 life",
			wantSteps: []string{"phyrexian", "window"}, wantActivated: 1, wantLifeLost: 3, wantManaSpent: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g, cardID := conduitPaidCastGame(t, "Vault Skirge", tc.permanents...)
			g.Players[game.Player1].ManaPool.Add(mana.B, tc.poolB)
			life := g.Players[game.Player1].Life
			clone := g.Clone()
			agent := &phyrexianWindowAgent{payLife: tc.payLife, ability: tc.ability}

			e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(agent), &TurnLog{})

			requirePaidCast(t, clone, cardID, zone.Graveyard, tc.wantManaSpent)
			grove := reboundPermanentNamed(t, clone, "Waterlogged Grove")
			if !slices.Equal(agent.steps, tc.wantSteps) || agent.activated != tc.wantActivated ||
				grove.Tapped != (tc.wantActivated == 1) || clone.Players[game.Player1].Life != life-tc.wantLifeLost {
				t.Fatalf("steps=%v activated=%d grove tapped=%t life=%d, want steps %v, %d activations and life %d",
					agent.steps, agent.activated, grove.Tapped, clone.Players[game.Player1].Life, tc.wantSteps, tc.wantActivated, life-tc.wantLifeLost)
			}
			if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Life != life ||
				reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped {
				t.Fatal("casting in the clone changed the original game")
			}
		})
	}
}

// TestFreeCastPhyrexianKickerAnnouncedBeforeManaWindow casts a spell from hand
// without paying its mana cost, paying a {G/P} kicker. With the kicker
// announced as 2 life the total includes no mana payment and no window opens;
// Thalia's {1} tax still opens it and the Grove pays the tax (CR 601.2b,
// 601.2f-g). K'rrik doesn't change the total, so a {B} kicker remains a mana
// payment: the window opens and K'rrik's life option is taken while paying
// (CR 601.2h).
func TestFreeCastPhyrexianKickerAnnouncedBeforeManaWindow(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name          string
		kicker        cost.Symbol
		permanents    []string
		thalia        bool
		wantSteps     []string
		wantActivated int
		wantLifeLost  int
		wantManaSpent int
	}{
		{name: "life covers the kicker", kicker: cost.PhyrexianMana(mana.G), wantSteps: []string{"phyrexian"}, wantLifeLost: 2},
		{name: "life with a tax mana payment", kicker: cost.PhyrexianMana(mana.G), thalia: true,
			wantSteps: []string{"phyrexian", "window"}, wantActivated: 1, wantLifeLost: 3, wantManaSpent: 1},
		{name: "K'rrik keeps the mana payment", kicker: cost.B, permanents: []string{"K'rrik, Son of Yawgmoth"},
			wantSteps: []string{"window", "phyrexian"}, wantLifeLost: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			for _, name := range tc.permanents {
				addCombatPermanent(g, game.Player1, reg.Lookup(name))
			}
			if tc.thalia {
				addCombatPermanent(g, game.Player2, reg.Lookup("Thalia, Guardian of Thraben"))
			}
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			def := freeCastTestInstant("Phyrexian Kicker Spell")
			def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
				game.KickerKeyword{Cost: cost.Mana{tc.kicker}},
			}}}
			cardID := addCardToHand(g, game.Player1, def)
			life := g.Players[game.Player1].Life
			ability := "no such ability"
			if tc.wantActivated != 0 {
				ability = "Pay 1 life"
			}
			agent := &phyrexianWindowAgent{payLife: true, ability: ability}

			ok := e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
				game.Selection{}, opt.V[int]{}, false, reboundKickerAgents(agent), &TurnLog{})

			obj, found := g.Stack.Peek()
			if !ok || !found || obj.SourceID != cardID || !obj.KickerPaid || obj.ManaSpentToCast != tc.wantManaSpent {
				t.Fatalf("cast=%t stack top=%+v, want the kicked free cast with %d mana spent", ok, obj, tc.wantManaSpent)
			}
			if !slices.Equal(agent.steps, tc.wantSteps) || agent.activated != tc.wantActivated ||
				grove.Tapped != (tc.wantActivated == 1) || g.Players[game.Player1].Life != life-tc.wantLifeLost {
				t.Fatalf("steps=%v activated=%d grove tapped=%t life=%d, want steps %v, %d activations and life %d",
					agent.steps, agent.activated, grove.Tapped, g.Players[game.Player1].Life, tc.wantSteps, tc.wantActivated, life-tc.wantLifeLost)
			}
		})
	}
}

// TestMadnessPhyrexianLifeAnnouncedBeforeManaWindow casts madness costs with a
// Phyrexian symbol announced as 2 life. {G/P} alone includes no mana payment,
// so no window opens; {1}{G/P} opens it and the Grove pays the {1} (CR 601.2b,
// 601.2g, 702.35a).
func TestMadnessPhyrexianLifeAnnouncedBeforeManaWindow(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	for _, tc := range []struct {
		name          string
		madness       cost.Mana
		wantSteps     []string
		wantActivated int
		wantLifeLost  int
		wantManaSpent int
	}{
		{name: "life only", madness: cost.Mana{cost.PhyrexianMana(mana.G)}, wantSteps: []string{"phyrexian"}, wantLifeLost: 2},
		{name: "life and mana", madness: cost.Mana{cost.O(1), cost.PhyrexianMana(mana.G)},
			wantSteps: []string{"phyrexian", "window"}, wantActivated: 1, wantLifeLost: 3, wantManaSpent: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addCardToHand(g, game.Player1, madnessSorcery(tc.madness))
			grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
			discardIntoMadness(t, e, g, cardID)
			life := g.Players[game.Player1].Life
			agent := &phyrexianWindowAgent{payLife: true, ability: "Pay 1 life"}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

			obj, _ := requireMadnessCast(t, g, cardID)
			if !slices.Equal(agent.steps, tc.wantSteps) || agent.activated != tc.wantActivated ||
				grove.Tapped != (tc.wantActivated == 1) || g.Players[game.Player1].Life != life-tc.wantLifeLost ||
				obj.ManaSpentToCast != tc.wantManaSpent {
				t.Fatalf("steps=%v activated=%d grove tapped=%t life=%d mana spent=%d, want steps %v, %d activations, life %d and %d mana",
					agent.steps, agent.activated, grove.Tapped, g.Players[game.Player1].Life, obj.ManaSpentToCast,
					tc.wantSteps, tc.wantActivated, life-tc.wantLifeLost, tc.wantManaSpent)
			}
		})
	}
}

// TestAnnouncedPhyrexianManaIsPaidWithMana announces a {B/P} as a mana payment
// on each route, then activates Waterlogged Grove, which can add only {G} or
// {U}. The announcement is binding (CR 601.2b, 601.2h): the symbol can't fall
// back to 2 life, so the proposal is reversed (CR 733) with the Grove untapped
// and its life refunded. The paid Vault Skirge returns to the graveyard, the
// madness card goes to the graveyard, and the free cast retries unkicked.
func TestAnnouncedPhyrexianManaIsPaidWithMana(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	blackPhyrexian := cost.PhyrexianMana(mana.B)
	t.Run("paid", func(t *testing.T) {
		e, g, cardID := conduitPaidCastGame(t, "Vault Skirge", "Etherium Sculptor")
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{ability: "Pay 1 life"}

		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

		if g.Stack.Size() != 0 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("Vault Skirge was cast although its announced {B} was not paid")
		}
		requireAnnouncedManaReversed(t, g, agent, reboundPermanentNamed(t, g, "Waterlogged Grove"), life)
	})
	t.Run("madness", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		cardID := addCardToHand(g, game.Player1, madnessSorcery(cost.Mana{blackPhyrexian}))
		grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
		discardIntoMadness(t, e, g, cardID)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{ability: "Pay 1 life"}

		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

		requireMadnessNotCast(t, g, cardID)
		requireAnnouncedManaReversed(t, g, agent, grove, life)
	})
	t.Run("free", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		grove := addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
		def := freeCastTestInstant("Phyrexian Kicker Spell")
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
			game.KickerKeyword{Cost: cost.Mana{blackPhyrexian}},
		}}}
		cardID := addCardToHand(g, game.Player1, def)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{ability: "Pay 1 life"}

		ok := e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
			game.Selection{}, opt.V[int]{}, false, reboundKickerAgents(agent), &TurnLog{})

		obj, found := g.Stack.Peek()
		if !ok || !found || obj.SourceID != cardID || obj.KickerPaid {
			t.Fatalf("cast=%t stack top=%+v, want the unkicked retry after the unpaid kicker", ok, obj)
		}
		requireAnnouncedManaReversed(t, g, agent, grove, life)
	})
}

func requireAnnouncedManaReversed(t *testing.T, g *game.Game, agent *phyrexianWindowAgent, grove *game.Permanent, life int) {
	t.Helper()
	if agent.activated != 1 || !slices.Equal(agent.steps[:2], []string{"phyrexian", "window"}) {
		t.Fatalf("steps=%v activated=%d, want {B} announced as mana, then the Grove activated", agent.steps, agent.activated)
	}
	if grove.Tapped || g.Players[game.Player1].Life != life || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("grove tapped=%t life=%d pool=%d, want the Grove activation reversed and life %d",
			grove.Tapped, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Total(), life)
	}
}

// TestPaidCastPhyrexianAnnouncedBeforeTargets casts the registered Gitaxian
// Probe ({U/P}, "Look at target player's hand.") from the graveyard during
// resolution. Its Phyrexian payment is announced with its cost before its
// target is chosen (CR 601.2b-c), and paying 2 life leaves the total without a
// mana payment.
func TestPaidCastPhyrexianAnnouncedBeforeTargets(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	cardID := addCardToGraveyard(g, game.Player1, cards.NewDefaultRegistry().Lookup("Gitaxian Probe"))
	life := g.Players[game.Player1].Life
	agent := &phyrexianWindowAgent{payLife: true}

	ok := e.castPaidTargetedSpell(g, game.Player1, cardID, zone.Graveyard, reboundKickerAgents(agent), &TurnLog{})

	if !ok {
		t.Fatal("Gitaxian Probe was not cast")
	}
	requirePaidCast(t, g, cardID, zone.Graveyard, 0)
	obj, _ := g.Stack.Peek()
	if !slices.Equal(agent.steps, []string{"phyrexian", "target"}) || len(obj.Targets) != 1 || g.Players[game.Player1].Life != life-2 {
		t.Fatalf("steps=%v targets=%v life=%d, want the {U/P} announced before the target and life %d",
			agent.steps, obj.Targets, g.Players[game.Player1].Life, life-2)
	}
}

// TestKrrikPaysLifeForAnnouncedPhyrexianMana casts the registered Vault Skirge
// through Conduit of Worlds with Etherium Sculptor and K'rrik, Son of
// Yawgmoth, with {B} available. Announcing {B/P} as {B} keeps a mana payment
// and opens the window; K'rrik then lets the caster pay 2 life rather than that
// {B} while paying (CR 601.2h; K'rrik ruling 2019-08-23: "If you choose to pay
// {B} this way, K'rrik's ability allows you to pay life rather than pay that
// mana."), so the black mana stays in the pool.
func TestKrrikPaysLifeForAnnouncedPhyrexianMana(t *testing.T) {
	e, g, cardID := conduitPaidCastGame(t, "Vault Skirge", "Etherium Sculptor", "K'rrik, Son of Yawgmoth")
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	life := g.Players[game.Player1].Life
	agent := &phyrexianWindowAgent{lifeAfterWindow: true, ability: "no such ability"}

	e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.ManaSpentToCast != 0 {
		t.Fatalf("stack top=%+v, want Vault Skirge cast without spending mana", obj)
	}
	if !slices.Equal(agent.steps, []string{"phyrexian", "window", "phyrexian"}) ||
		g.Players[game.Player1].Life != life-2 || g.Players[game.Player1].ManaPool.Amount(mana.B) != 1 {
		t.Fatalf("steps=%v life=%d black=%d, want K'rrik's life option after the window, life %d and {B} kept",
			agent.steps, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Amount(mana.B), life-2)
	}
}

// TestAnnouncedPhyrexianLifeLeavesManaForOtherSymbols announces each {B/P} as
// 2 life with only {B} available and no mana ability to activate. Locking the
// total honors that announcement (CR 601.2b, 601.2f): the {B} pays the {1}
// rather than the {B/P}, so each route casts its spell instead of rejecting a
// payable total. The paid route casts the registered Vault Skirge ({1}{B/P})
// through Conduit of Worlds; the free route pays a {1}{B/P} kicker; madness
// pays {1}{B/P} and, with two announced symbols, {1}{B/P}{B/P}.
func TestAnnouncedPhyrexianLifeLeavesManaForOtherSymbols(t *testing.T) {
	blackPhyrexian := cost.PhyrexianMana(mana.B)
	requireLifeAndMana := func(t *testing.T, g *game.Game, agent *phyrexianWindowAgent, obj *game.StackObject, life, lifeLost int) {
		t.Helper()
		if obj.ManaSpentToCast != 1 || g.Players[game.Player1].ManaPool.Total() != 0 ||
			g.Players[game.Player1].Life != life-lifeLost || slices.Contains(agent.steps, "window") {
			t.Fatalf("mana spent=%d pool=%d life=%d steps=%v, want {B} spent on {1}, life %d and no window",
				obj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total(), g.Players[game.Player1].Life, agent.steps, life-lifeLost)
		}
	}
	t.Run("paid", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		conduit := addConduit(g, game.Player1)
		cardID := addCardToGraveyard(g, game.Player1, cards.NewDefaultRegistry().Lookup("Vault Skirge"))
		setUpConduitMainPhase(g)
		if !activateConduitTargeting(t, g, e, conduit, cardID) {
			t.Fatal("activating Conduit of Worlds failed")
		}
		g.Players[game.Player1].ManaPool.Add(mana.B, 1)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{payLife: true}

		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

		requirePaidCast(t, g, cardID, zone.Graveyard, 1)
		obj, _ := g.Stack.Peek()
		requireLifeAndMana(t, g, agent, obj, life, 2)
	})
	t.Run("free", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		def := freeCastTestInstant("Phyrexian Kicker Spell")
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
			game.KickerKeyword{Cost: cost.Mana{cost.O(1), blackPhyrexian}},
		}}}
		cardID := addCardToHand(g, game.Player1, def)
		g.Players[game.Player1].ManaPool.Add(mana.B, 1)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{payLife: true}

		ok := e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
			game.Selection{}, opt.V[int]{}, false, reboundKickerAgents(agent), &TurnLog{})

		obj, found := g.Stack.Peek()
		if !ok || !found || obj.SourceID != cardID || !obj.KickerPaid {
			t.Fatalf("cast=%t stack top=%+v, want the kicked free cast", ok, obj)
		}
		requireLifeAndMana(t, g, agent, obj, life, 2)
	})
	for _, tc := range []struct {
		name     string
		madness  cost.Mana
		lifeLost int
	}{
		{name: "madness", madness: cost.Mana{cost.O(1), blackPhyrexian}, lifeLost: 2},
		{name: "madness with two announced symbols", madness: cost.Mana{blackPhyrexian, cost.O(1), blackPhyrexian}, lifeLost: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addCardToHand(g, game.Player1, madnessSorcery(tc.madness))
			discardIntoMadness(t, e, g, cardID)
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			life := g.Players[game.Player1].Life
			agent := &phyrexianWindowAgent{payLife: true}

			e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})

			obj, _ := requireMadnessCast(t, g, cardID)
			requireLifeAndMana(t, g, agent, obj, life, tc.lifeLost)
		})
	}
}

package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3587. Default-registry cards checked against the
// pinned corpus (oracle_id): Overeager Apprentice
// 3ea71eed-e1ae-4e28-a4f5-44ff115949f4 {2}{B} "Discard a card, Sacrifice this
// creature: Add {B}{B}{B}."; Molt Tender 9a39ca64-e1a9-42ef-8efd-6aa3edde3d93
// {G} "{T}: Mill a card. ... / {T}, Exile a card from your graveyard: Add one
// mana of any color."; Lightning Bolt 4457ed35-7c10-48c8-9776-456485fdf070
// {R} "Lightning Bolt deals 3 damage to any target."; Sol Ring
// 6ad8011d-3471-4369-9d68-b264cc027487 {1} "{T}: Add {C}{C}."; Thalia,
// Guardian of Thraben 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells
// cost {1} more to cast."; Waterlogged Grove
// 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or {U}.";
// Conduit of Worlds ed14be15-8f8d-4fe3-a147-f5da8ed873bf "{T}: Choose target
// nonland permanent card in your graveyard. If you haven't cast a spell this
// turn, you may cast that card. ..."; Grizzly Bears
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0 {1}{G}. Comprehensive Rules effective
// 2026-09-25: 601.2a "the player first moves that card (or that copy of a
// card) from where it is to the stack"; 601.2f locks the total cost; 601.2g
// "the player then has a chance to activate mana abilities"; 707.12 a cast copy
// of a card is not that card, which stays where it is; 733.1 reverses a cast
// whose cost isn't paid. The mana abilities are activated only after the move,
// so a cast is offered only when they can pay without the card being cast.

// proposedSpellManaAgent accepts optional casts, activates the named mana
// abilities in order, one at each payment-time window, and records every
// prompt it is shown.
type proposedSpellManaAgent struct {
	abilities []string
	prompts   []string
}

func (*proposedSpellManaAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *proposedSpellManaAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.prompts = append(a.prompts, request.Prompt)
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Prompt == "Activate a mana ability while paying?" && len(a.abilities) > 0:
		for _, option := range request.Options {
			if strings.Contains(option.Label, a.abilities[0]) {
				a.abilities = a.abilities[1:]
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func requireProposedSpellCast(t *testing.T, g *game.Game, cardID id.ID, fromZone zone.Type, manaSpent int) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.Copy || obj.SourceID != cardID || obj.SourceZone != fromZone || obj.ManaSpentToCast != manaSpent {
		t.Fatalf("stack top = %+v, want the cast of card %v from %v with %d mana spent", obj, cardID, fromZone, manaSpent)
	}
}

// requireNoProposedCast checks that nothing was offered or changed: the
// probe ran only on clones, so the live game's cards, IDs and events are as
// they were.
func requireNoProposedCast(t *testing.T, g *game.Game, agent *proposedSpellManaAgent, cardID id.ID, fromZone zone.Type, before *game.Game, permanents ...string) {
	t.Helper()
	if len(agent.prompts) != 0 {
		t.Fatalf("prompts = %q, want no cast offered", agent.prompts)
	}
	if source := castSourceZone(g.Players[game.Player1], fromZone); source == nil || !source.Contains(cardID) || g.Stack.Size() != before.Stack.Size() {
		t.Fatal("the card left its zone or the stack changed although no cast was offered")
	}
	if g.IDGen.Current() != before.IDGen.Current() || len(g.Events) != len(before.Events) ||
		g.Players[game.Player1].Hand.Size() != before.Players[game.Player1].Hand.Size() ||
		g.Players[game.Player1].Graveyard.Size() != before.Players[game.Player1].Graveyard.Size() ||
		g.Players[game.Player1].Exile.Size() != before.Players[game.Player1].Exile.Size() ||
		g.Players[game.Player1].ManaPool.Total() != before.Players[game.Player1].ManaPool.Total() {
		t.Fatal("offering the cast changed the live game")
	}
	for _, name := range permanents {
		if reboundPermanentNamed(t, g, name).Tapped {
			t.Fatalf("%s was tapped although no cast was offered", name)
		}
	}
}

// TestFreeCastFromHandProbesManaAfterMovingSpell offers a free cast of
// Lightning Bolt from hand with Thalia's {1} tax and Overeager Apprentice as
// the only mana source. Casting Bolt moves it to the stack first (CR 601.2a),
// so when Bolt is the only card in hand the Apprentice can't pay its discard
// cost at the payment-time window (CR 601.2g) and no cast is offered. With
// Grizzly Bears also in hand, the cast is offered and the Apprentice discards
// the Bears to pay the tax.
func TestFreeCastFromHandProbesManaAfterMovingSpell(t *testing.T) {
	setup := func(t *testing.T, extra bool) (*Engine, *game.Game, id.ID) {
		e, g := newEffectiveCascadeGame()
		addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
		addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Overeager Apprentice"))
		cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Lightning Bolt"))
		if extra {
			addCardToHand(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		}
		return e, g, cardID
	}
	t.Run("only card in hand", func(t *testing.T) {
		e, g, cardID := setup(t, false)
		before := g.Clone()
		agent := &proposedSpellManaAgent{abilities: []string{"Discard a card"}}
		if e.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
			game.Selection{}, opt.V[int]{}, false, true, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("free cast succeeded with only the spell itself to discard")
		}
		requireNoProposedCast(t, g, agent, cardID, zone.Hand, before)
		reboundPermanentNamed(t, g, "Overeager Apprentice")
	})
	t.Run("another card in hand", func(t *testing.T) {
		e, g, cardID := setup(t, true)
		agent := &proposedSpellManaAgent{abilities: []string{"Discard a card"}}
		if !e.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand,
			game.Selection{}, opt.V[int]{}, false, true, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatalf("free cast failed; prompts = %q", agent.prompts)
		}
		requireProposedSpellCast(t, g, cardID, zone.Hand, 1)
		player := g.Players[game.Player1]
		if player.Hand.Size() != 0 || player.Graveyard.Size() != 2 {
			t.Fatalf("hand=%d graveyard=%d, want Bears discarded and the Apprentice sacrificed", player.Hand.Size(), player.Graveyard.Size())
		}
	})
}

// TestFreeCastFromGraveyardProbesManaAfterMovingSpell offers a free cast of Sol
// Ring from the graveyard with Thalia's {1} tax and Molt Tender as the only
// mana source. Sol Ring is on the stack before mana abilities are activated
// (CR 601.2a, 601.2g), so Molt Tender can't exile it, and with no other
// graveyard card no cast is offered. With Grizzly Bears also in the graveyard,
// Molt Tender exiles the Bears to pay the tax.
func TestFreeCastFromGraveyardProbesManaAfterMovingSpell(t *testing.T) {
	setup := func(t *testing.T, extra bool) (*Engine, *game.Game, id.ID) {
		e, g := newEffectiveCascadeGame()
		addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
		addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Molt Tender"))
		cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Sol Ring"))
		if extra {
			addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		}
		return e, g, cardID
	}
	t.Run("only card in graveyard", func(t *testing.T) {
		e, g, cardID := setup(t, false)
		before := g.Clone()
		agent := &proposedSpellManaAgent{abilities: []string{"Exile a card"}}
		if e.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard,
			game.Selection{}, opt.V[int]{}, false, true, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("free cast succeeded with only the spell itself to exile")
		}
		requireNoProposedCast(t, g, agent, cardID, zone.Graveyard, before, "Molt Tender")
	})
	t.Run("another card in graveyard", func(t *testing.T) {
		e, g, cardID := setup(t, true)
		agent := &proposedSpellManaAgent{abilities: []string{"Exile a card"}}
		if !e.chooseAndCastFreeSpellFromSourceMay(g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard,
			game.Selection{}, opt.V[int]{}, false, true, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatalf("free cast failed; prompts = %q", agent.prompts)
		}
		requireProposedSpellCast(t, g, cardID, zone.Graveyard, 1)
		player := g.Players[game.Player1]
		if player.Graveyard.Size() != 0 || player.Exile.Size() != 1 || !reboundPermanentNamed(t, g, "Molt Tender").Tapped {
			t.Fatalf("graveyard=%d exile=%d, want Bears exiled by a tapped Molt Tender", player.Graveyard.Size(), player.Exile.Size())
		}
	})
}

// TestFreeCastCopyProbesManaWithCardInPlace casts a copy of Sol Ring, which
// stays in the graveyard: the copy is not the card (CR 707.12), so only the
// copy is put on the stack and Molt Tender may exile the Sol Ring card itself
// to pay Thalia's tax.
func TestFreeCastCopyProbesManaWithCardInPlace(t *testing.T) {
	e, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Molt Tender"))
	cardID := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Sol Ring"))
	agent := &proposedSpellManaAgent{abilities: []string{"Exile a card"}}
	if !e.chooseAndCastFreeSpell(g, g.Players[game.Player1], game.Player1, cardID, zone.Graveyard, game.Selection{}, opt.V[int]{},
		freeCastSpell{copyCard: true}, true, reboundKickerAgents(agent), &TurnLog{}) {
		t.Fatalf("free copy cast failed; prompts = %q", agent.prompts)
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Copy || !obj.CastCopy || obj.ManaSpentToCast != 1 {
		t.Fatalf("stack top = %+v, want the cast copy with Thalia's tax paid", obj)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || !reboundPermanentNamed(t, g, "Molt Tender").Tapped {
		t.Fatal("Molt Tender did not exile the copied Sol Ring card to pay")
	}
}

// TestConduitPaidCastLocksTotalAfterMovingSpell casts Sol Ring through Conduit
// of Worlds with Thalia's tax, so the locked total is {2}, and Waterlogged
// Grove and Molt Tender as the mana sources. The total is locked with Sol Ring
// already on the stack (CR 601.2a, 601.2f), so when it is the only graveyard
// card Molt Tender can't exile it and the cast isn't proposed: Sol Ring stays
// in the graveyard, no window opens and Conduit applies no lock. With Grizzly
// Bears also in the graveyard, both sources pay the {2}.
func TestConduitPaidCastLocksTotalAfterMovingSpell(t *testing.T) {
	setup := func(t *testing.T, extra bool) (*Engine, *game.Game, id.ID) {
		e, g, cardID := conduitPaidCastGame(t, "Sol Ring", "Molt Tender")
		addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Thalia, Guardian of Thraben"))
		if extra {
			addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		}
		return e, g, cardID
	}
	t.Run("only card in graveyard", func(t *testing.T) {
		e, g, cardID := setup(t, false)
		ids := g.IDGen.Current()
		agent := &proposedSpellManaAgent{abilities: []string{"Exile a card"}}
		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
		for _, prompt := range agent.prompts {
			if prompt == "Activate a mana ability while paying?" {
				t.Fatal("the payment-time window opened for a cast that can't be paid")
			}
		}
		if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Stack.Size() != 0 || g.IDGen.Current() != ids {
			t.Fatal("Sol Ring was proposed although its locked total can't be paid")
		}
		if reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || reboundPermanentNamed(t, g, "Molt Tender").Tapped {
			t.Fatal("a mana source was used for a cast that wasn't proposed")
		}
		if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
			t.Fatal("Conduit's can't-cast-additional lock applied without a cast")
		}
	})
	t.Run("another card in graveyard", func(t *testing.T) {
		e, g, cardID := setup(t, true)
		life := g.Players[game.Player1].Life
		agent := &proposedSpellManaAgent{abilities: []string{"Exile a card", "Pay 1 life"}}
		e.resolveTopOfStackWithChoices(g, reboundKickerAgents(agent), &TurnLog{})
		requirePaidCast(t, g, cardID, zone.Graveyard, 2)
		player := g.Players[game.Player1]
		if player.Life != life-1 || player.Graveyard.Size() != 0 || player.Exile.Size() != 1 ||
			!reboundPermanentNamed(t, g, "Molt Tender").Tapped || !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped {
			t.Fatalf("life=%d graveyard=%d exile=%d, want Grove and Molt Tender (exiling Bears) to pay", player.Life, player.Graveyard.Size(), player.Exile.Size())
		}
		if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
			t.Fatal("Conduit's can't-cast-additional lock was not applied after the cast")
		}
	})
}

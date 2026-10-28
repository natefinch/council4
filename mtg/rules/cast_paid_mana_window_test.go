package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-09-25 (media.wizards.com
// MagicCompRules 20260925). CR 601.2f locks the total cost, including cost
// increases and reductions; CR 601.2g: "If the total cost includes a mana
// payment, the player then has a chance to activate mana abilities"; CR 605.3a
// limits activating mana abilities mid-cast to that window; CR 733.1 reverses an
// unpaid cast. Oracle (pinned corpus): Conduit of Worlds 8380eb8d-d1c3-4f96-
// b3b9-54845188c1d1 "{T}: Choose target nonland permanent card in your
// graveyard. If you haven't cast a spell this turn, you may cast that card. If
// you do, you can't cast additional spells this turn. Activate only as a
// sorcery."; Ornithopter 305078a5-ac18-4721-bba2-3434eba5b1cf {0} "Flying";
// Sol Ring 8ee443cc-e17a-493b-9c93-1f9e141a30e4 {1}; Etherium Sculptor
// b0556f20-1ea8-4ab5-9a09-dec9959f017b "Artifact spells you cast cost {1} less
// to cast."; Thalia, Guardian of Thraben 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee
// "Noncreature spells cost {1} more to cast."; Waterlogged Grove 0ab6bfbd-
// d2e1-4c4c-9f91-6f69c5b8e3bb "{T}, Pay 1 life: Add {G} or {U}."; Panglacial
// Wurm 063bb28b-5e32-4f31-a208-16b653edf413 "While you're searching your
// library, you may cast this card from your library." These scenarios cover
// Engine gap #3464 on both paid-cast entry routes (resolution-time targeted
// casts and casts while searching a library).

// paidCastManaAgent accepts every optional cast and otherwise behaves as
// paymentManaAgent at the payment-time mana ability window.
type paidCastManaAgent struct{ *paymentManaAgent }

func (a paidCastManaAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	return a.paymentManaAgent.ChooseChoice(obs, request)
}

// conduitPaidCastGame puts Conduit of Worlds, Waterlogged Grove and the named
// permanents onto Player1's battlefield, the named card into Player1's
// graveyard, and activates Conduit targeting it.
func conduitPaidCastGame(t *testing.T, target string, permanents ...string) (*Engine, *game.Game, id.ID) {
	t.Helper()
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	source := addConduit(g, game.Player1)
	addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	for _, name := range permanents {
		addCombatPermanent(g, game.Player1, reg.Lookup(name))
	}
	cardID := addCardToGraveyard(g, game.Player1, reg.Lookup(target))
	setUpConduitMainPhase(g)
	if !activateConduitTargeting(t, g, e, source, cardID) {
		t.Fatal("activating Conduit of Worlds failed")
	}
	return e, g, cardID
}

func requirePaidCast(t *testing.T, g *game.Game, cardID id.ID, fromZone zone.Type, manaSpent int) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID || obj.SourceZone != fromZone {
		t.Fatalf("stack top = %+v, want the paid cast of %v from %v", obj, cardID, fromZone)
	}
	if obj.ManaSpentToCast != manaSpent || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana spent=%d pool=%d, want %d spent and an empty pool", obj.ManaSpentToCast, g.Players[game.Player1].ManaPool.Total(), manaSpent)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.CardID == cardID {
			return
		}
	}
	t.Fatal("no spell-cast event for the paid cast")
}

// TestPaidCastZeroTotalOpensNoManaAbilityWindow casts, through Conduit of
// Worlds, a printed {0} card and a {1} artifact whose locked total is reduced to
// {0}. Neither total includes a mana payment, so no payment-time mana ability
// window opens and Waterlogged Grove can't be tapped mid-cast (CR 601.2g,
// 605.3a). Each cast runs in a clone without changing the original game.
func TestPaidCastZeroTotalOpensNoManaAbilityWindow(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		permanents   []string
	}{
		{name: "printed zero", target: "Ornithopter"},
		{name: "reduced to zero", target: "Sol Ring", permanents: []string{"Etherium Sculptor"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g, cardID := conduitPaidCastGame(t, tc.target, tc.permanents...)
			life := g.Players[game.Player1].Life
			clone := g.Clone()
			agent := &paymentManaAgent{ability: "Pay 1 life"}
			e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(paidCastManaAgent{agent}), &TurnLog{})

			requirePaidCast(t, clone, cardID, zone.Graveyard, 0)
			if agent.windows != 0 || reboundPermanentNamed(t, clone, "Waterlogged Grove").Tapped || clone.Players[game.Player1].Life != life {
				t.Fatalf("windows=%d life=%d, want no mana ability window and an untapped Grove", agent.windows, clone.Players[game.Player1].Life)
			}
			if !spellCastProhibited(clone, game.Player1, vanillaCreatureDef()) {
				t.Fatal("Conduit's can't-cast-additional lock was not applied after the cast")
			}
			if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Stack.Size() != 1 {
				t.Fatal("casting in the clone changed the original game")
			}
		})
	}
}

// TestPaidCastManaTotalUsesManaAbilityWindow casts Sol Ring through Conduit of
// Worlds with Waterlogged Grove as the only mana source. The locked total
// includes a mana payment ({1}, or {1} with Thalia's tax offset by Etherium
// Sculptor), so the window opens and the Grove pays it (CR 601.2f-g). When the
// caster doesn't activate the Grove, the proposal is reversed: Sol Ring returns
// to the graveyard and Conduit applies no lock (CR 733.1).
func TestPaidCastManaTotalUsesManaAbilityWindow(t *testing.T) {
	for _, tc := range []struct {
		name       string
		permanents []string
	}{
		{name: "printed cost"},
		{name: "tax and reduction", permanents: []string{"Thalia, Guardian of Thraben", "Etherium Sculptor"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g, cardID := conduitPaidCastGame(t, "Sol Ring", tc.permanents...)
			life := g.Players[game.Player1].Life

			clone := g.Clone()
			agent := &paymentManaAgent{ability: "Pay 1 life"}
			e.resolveTopOfStackWithChoices(clone, reboundKickerAgents(paidCastManaAgent{agent}), &TurnLog{})
			requirePaidCast(t, clone, cardID, zone.Graveyard, 1)
			if agent.windows != 1 || agent.activated != 1 || !reboundPermanentNamed(t, clone, "Waterlogged Grove").Tapped ||
				clone.Players[game.Player1].Life != life-1 {
				t.Fatalf("windows=%d activated=%d life=%d, want the Grove to pay {1}", agent.windows, agent.activated, clone.Players[game.Player1].Life)
			}

			declined := g.Clone()
			agent = &paymentManaAgent{ability: "no such ability"}
			e.resolveTopOfStackWithChoices(declined, reboundKickerAgents(paidCastManaAgent{agent}), &TurnLog{})
			if agent.windows != 1 || declined.Stack.Size() != 0 || !declined.Players[game.Player1].Graveyard.Contains(cardID) ||
				reboundPermanentNamed(t, declined, "Waterlogged Grove").Tapped || declined.Players[game.Player1].Life != life {
				t.Fatalf("windows=%d stack=%d, want the unpaid proposal reversed into the graveyard", agent.windows, declined.Stack.Size())
			}
			if spellCastProhibited(declined, game.Player1, vanillaCreatureDef()) {
				t.Fatal("Conduit locked casting although the card was not cast")
			}
		})
	}
}

// TestSearchPaidCastZeroManaTotalOpensNoManaAbilityWindow casts a {0}
// Panglacial Wurm variant with an additional "Pay 2 life" cost while searching.
// The total has no mana payment, so the life is paid without opening a mana
// ability window or tapping Waterlogged Grove (CR 601.2g-h).
func TestSearchPaidCastZeroManaTotalOpensNoManaAbilityWindow(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := authored.PanglacialWurm.New()
	def.ManaCost = opt.Val(cost.Mana{})
	def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2, Text: "Pay 2 life"}}
	wurm := addCardToLibrary(g, game.Player1, def)
	addCombatPermanent(g, game.Player1, reg.Lookup("Waterlogged Grove"))
	life := g.Players[game.Player1].Life
	agent := &paymentManaAgent{ability: "Pay 1 life"}
	policies := simPassPolicies()
	policies[game.Player1] = paidCastManaAgent{agent}

	NewEngine(nil).searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true}, 1)
	requirePaidCast(t, g, wurm, zone.Library, 0)
	if agent.windows != 0 || reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped || g.Players[game.Player1].Life != life-2 {
		t.Fatalf("windows=%d life=%d, want only the additional life paid and no mana ability window", agent.windows, g.Players[game.Player1].Life)
	}
}

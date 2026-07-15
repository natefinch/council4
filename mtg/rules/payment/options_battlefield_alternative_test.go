package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// battlefieldAlternativeState exercises the runtime gating of a board-state
// conditional alternative cost (Blasphemous Edict). PermanentMatchesSelection
// reports a permanent as one of the counted creatures only when it is in the
// creature set and the evaluator asked for the creature card type, so a wrong
// selection would undercount and fail the offering assertions. Every other State
// method is inherited from fakePaymentState.
type battlefieldAlternativeState struct {
	fakePaymentState

	creature map[id.ID]bool
}

func (s battlefieldAlternativeState) PermanentMatchesSelection(p *game.Permanent, sel game.Selection) bool {
	if p == nil || !s.creature[p.ObjectID] {
		return false
	}
	return len(sel.RequiredTypesAny) == 1 && sel.RequiredTypesAny[0] == types.Creature
}

// blasphemousEdictLikeCard is a Blasphemous-Edict-shaped spell: {3}{B}{B} with a
// "Pay {B}" alternative gated by thirteen or more creatures on the battlefield.
func blasphemousEdictLikeCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Edict Test",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
		Types:    []types.Card{types.Sorcery},
		AlternativeCosts: []cost.Alternative{{
			Label:                  "Pay {B}",
			ManaCost:               opt.Val(cost.Mana{cost.B}),
			Condition:              cost.AlternativeConditionPermanentsOnBattlefield,
			ConditionCount:         13,
			ConditionPermanentType: types.Creature,
		}},
	}}
}

// battlefieldWithCreatures builds a state whose battlefield holds count counted
// creatures. When controllers are given the creatures are spread across them in
// round-robin, proving the count spans all players rather than one controller.
func battlefieldWithCreatures(count int, controllers ...game.PlayerID) battlefieldAlternativeState {
	if len(controllers) == 0 {
		controllers = []game.PlayerID{game.Player1}
	}
	perms := make([]*game.Permanent, 0, count)
	creature := make(map[id.ID]bool, count)
	for i := range count {
		p := &game.Permanent{ObjectID: id.ID(i + 1), Controller: controllers[i%len(controllers)]}
		perms = append(perms, p)
		creature[p.ObjectID] = true
	}
	return battlefieldAlternativeState{
		fakePaymentState: fakePaymentState{battlefield: perms},
		creature:         creature,
	}
}

// TestBattlefieldAlternativeOfferedAtThreshold proves the board-state gate is a
// minimum ("thirteen or more"): the "Pay {B}" alternative is withheld at twelve
// counted creatures and offered at thirteen and fourteen, while the normal cost
// is always available and the alternative carries exactly its {B} replacement.
func TestBattlefieldAlternativeOfferedAtThreshold(t *testing.T) {
	t.Parallel()
	card := blasphemousEdictLikeCard()
	for _, tc := range []struct {
		name    string
		count   int
		offered bool
	}{
		{"twelve creatures", 12, false},
		{"thirteen creatures", 13, true},
		{"fourteen creatures", 14, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			state := battlefieldWithCreatures(tc.count)
			options := spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
			if _, ok := spellOptionByLabel(options, "Normal cost"); !ok {
				t.Fatal("normal-cost option missing")
			}
			alternative, ok := spellOptionByLabel(options, "Pay {B}")
			if ok != tc.offered {
				t.Fatalf("Pay {B} offered = %t, want %t (count %d)", ok, tc.offered, tc.count)
			}
			if tc.offered {
				if alternative.manaCost == nil || alternative.manaCost.String() != "{B}" {
					t.Fatalf("alternative mana cost = %#v, want {B}", alternative.manaCost)
				}
			}
		})
	}
}

// TestBattlefieldAlternativeCountsAllPlayers proves the threshold counts
// creatures controlled by every player: thirteen creatures split across two
// players satisfy the gate even though no single player controls thirteen.
func TestBattlefieldAlternativeCountsAllPlayers(t *testing.T) {
	t.Parallel()
	card := blasphemousEdictLikeCard()
	state := battlefieldWithCreatures(13, game.Player1, game.Player2)
	options := spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Pay {B}"); !ok {
		t.Fatal("Pay {B} not offered when thirteen creatures are split across two players")
	}
}

// TestBattlefieldAlternativeExcludesPhasedOutAndNonMatching proves the count
// ignores phased-out permanents and permanents that are not the counted type:
// a board of thirteen creatures with one phased out (plus extra non-creatures)
// counts as twelve, so the gate stays closed.
func TestBattlefieldAlternativeExcludesPhasedOutAndNonMatching(t *testing.T) {
	t.Parallel()
	card := blasphemousEdictLikeCard()
	state := battlefieldWithCreatures(13)
	// Phase one counted creature out, dropping the effective count to twelve.
	state.battlefield[0].PhasedOut = true
	// Add non-creature permanents that must never count toward the threshold.
	for i := range 5 {
		state.battlefield = append(state.battlefield, &game.Permanent{ObjectID: id.ID(100 + i), Controller: game.Player1})
	}
	options := spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Pay {B}"); ok {
		t.Fatal("Pay {B} offered when only twelve non-phased creatures are present")
	}
	// Phasing the creature back in restores the thirteenth and opens the gate.
	state.battlefield[0].PhasedOut = false
	options = spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Pay {B}"); !ok {
		t.Fatal("Pay {B} not offered when thirteen creatures are present alongside non-creatures")
	}
}

// TestBattlefieldAlternativePreservesAdditionalCosts proves a satisfied
// board-state alternative still carries the spell's required additional costs
// (CR 601.2f): the alternative replaces only the mana cost.
func TestBattlefieldAlternativePreservesAdditionalCosts(t *testing.T) {
	t.Parallel()
	card := blasphemousEdictLikeCard()
	card.AdditionalCosts = []cost.Additional{{
		Kind:   cost.AdditionalSacrifice,
		Text:   "sacrifice a creature",
		Amount: 1,
	}}
	state := battlefieldWithCreatures(13)
	options := spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
	alternative, ok := spellOptionByLabel(options, "Pay {B}")
	if !ok {
		t.Fatal("Pay {B} not offered")
	}
	if len(alternative.additionalCosts) != 1 || alternative.additionalCosts[0].Kind != cost.AdditionalSacrifice {
		t.Fatalf("alternative additional costs = %#v, want the required sacrifice preserved", alternative.additionalCosts)
	}
}

// TestBattlefieldAlternativeIndexStability proves a board-state alternative that
// is not offered (its gate unsatisfied) does not renumber the remaining
// alternatives: a later unconditional alternative keeps its position-derived
// cost-choice index whether or not the earlier one is available.
func TestBattlefieldAlternativeIndexStability(t *testing.T) {
	t.Parallel()
	card := blasphemousEdictLikeCard()
	// Append a second, unconditional alternative after the board-state one.
	card.AlternativeCosts = append(card.AlternativeCosts, cost.Alternative{
		Label:    "Pay {G}",
		ManaCost: opt.Val(cost.Mana{cost.G}),
	})

	// Too few creatures: the board-state alternative (index 1) is skipped, but
	// the unconditional one keeps index 2 rather than collapsing to 1.
	fewCreatures := battlefieldWithCreatures(0)
	options := spellCostOptionsForZoneAndKicker(fewCreatures, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Pay {B}"); ok {
		t.Fatal("Pay {B} offered with no creatures on the battlefield")
	}
	green, ok := spellOptionByLabel(options, "Pay {G}")
	if !ok {
		t.Fatal("unconditional Pay {G} alternative missing")
	}
	if green.index != 2 {
		t.Fatalf("Pay {G} index = %d, want stable index 2 even when Pay {B} is skipped", green.index)
	}

	// With the gate satisfied both alternatives are offered at their stable
	// position-derived indices.
	enough := battlefieldWithCreatures(13)
	options = spellCostOptionsForZoneAndKicker(enough, game.Player1, card, zone.Hand, false, 0, false, nil)
	black, ok := spellOptionByLabel(options, "Pay {B}")
	if !ok {
		t.Fatal("Pay {B} not offered with thirteen creatures")
	}
	if black.index != 1 {
		t.Fatalf("Pay {B} index = %d, want 1", black.index)
	}
	green, ok = spellOptionByLabel(options, "Pay {G}")
	if !ok {
		t.Fatal("Pay {G} not offered")
	}
	if green.index != 2 {
		t.Fatalf("Pay {G} index = %d, want 2", green.index)
	}
}

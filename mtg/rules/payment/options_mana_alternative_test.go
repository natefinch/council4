package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// manaAlternativeState is a configurable State for exercising the runtime
// gating of mana-only conditional alternative costs. It overrides only the two
// predicates those conditions consult, inheriting every other State method from
// fakePaymentState.
type manaAlternativeState struct {
	fakePaymentState

	opponentGainedLife bool
	attackingCreatures int
}

func (s manaAlternativeState) OpponentGainedLifeThisTurn(game.PlayerID) bool {
	return s.opponentGainedLife
}

func (s manaAlternativeState) AttackingCreatureCount() int {
	return s.attackingCreatures
}

func spellOptionByLabel(options []spellCostOption, label string) (spellCostOption, bool) {
	for _, option := range options {
		if option.label == label {
			return option, true
		}
	}
	return spellCostOption{}, false
}

// TestManaAlternativeCostOfferedOnlyWhenConditionSatisfied proves an opponent's
// life-gain gate on a mana-only alternative cost is enforced at cast time: the
// "Pay {B}" option appears alongside the normal cost only when the condition is
// satisfied, and disappears entirely when it is not, while the normal-cost
// option always remains.
func TestManaAlternativeCostOfferedOnlyWhenConditionSatisfied(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Needle Test",
		ManaCost: opt.Val(cost.Mana{cost.O(5), cost.B, cost.B}),
		Types:    []types.Card{types.Instant},
		AlternativeCosts: []cost.Alternative{{
			Label:     "Pay {B}",
			ManaCost:  opt.Val(cost.Mana{cost.B}),
			Condition: cost.AlternativeConditionOpponentGainedLifeThisTurn,
		}},
	}}

	satisfied := manaAlternativeState{opponentGainedLife: true}
	options := spellCostOptionsForZoneAndKicker(satisfied, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Normal cost"); !ok {
		t.Fatal("normal-cost option missing when condition satisfied")
	}
	alternative, ok := spellOptionByLabel(options, "Pay {B}")
	if !ok {
		t.Fatal("mana alternative option missing when condition satisfied")
	}
	if alternative.manaCost == nil || alternative.manaCost.String() != "{B}" {
		t.Fatalf("alternative mana cost = %#v, want {B}", alternative.manaCost)
	}

	unsatisfied := manaAlternativeState{opponentGainedLife: false}
	options = spellCostOptionsForZoneAndKicker(unsatisfied, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Normal cost"); !ok {
		t.Fatal("normal-cost option missing when condition not satisfied")
	}
	if _, ok := spellOptionByLabel(options, "Pay {B}"); ok {
		t.Fatal("mana alternative option offered when condition not satisfied")
	}
}

// TestZeroManaAlternativeCostIsDistinctOption proves that a {0} alternative cost
// appears as its own payable cast option (an explicit zero mana cost), separate
// from the normal cost, when its attacking-creatures gate is satisfied.
func TestZeroManaAlternativeCostIsDistinctOption(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		Name:     "Zero Test",
		ManaCost: opt.Val(cost.Mana{cost.O(4), cost.U, cost.U}),
		Types:    []types.Card{types.Instant},
		AlternativeCosts: []cost.Alternative{{
			Label:          "Pay {0}",
			ManaCost:       opt.Val(cost.Mana{cost.O(0)}),
			Condition:      cost.AlternativeConditionCreaturesAttacking,
			ConditionCount: 3,
		}},
	}}

	enough := manaAlternativeState{attackingCreatures: 3}
	options := spellCostOptionsForZoneAndKicker(enough, game.Player1, card, zone.Hand, false, 0, false, nil)
	zero, ok := spellOptionByLabel(options, "Pay {0}")
	if !ok {
		t.Fatal("{0} alternative option missing when three creatures attack")
	}
	if zero.manaCost == nil {
		t.Fatal("{0} alternative option has no explicit mana cost")
	}
	if len(*zero.manaCost) != 1 || zero.manaCost.String() != "{0}" {
		t.Fatalf("{0} alternative mana cost = %#v, want a single {0} symbol", zero.manaCost)
	}

	tooFew := manaAlternativeState{attackingCreatures: 2}
	options = spellCostOptionsForZoneAndKicker(tooFew, game.Player1, card, zone.Hand, false, 0, false, nil)
	if _, ok := spellOptionByLabel(options, "Pay {0}"); ok {
		t.Fatal("{0} alternative option offered with too few attackers")
	}
}

// TestManaAlternativeCostPreservesAdditionalCosts proves that choosing the
// mana-only alternative still requires the spell's additional costs (CR 601.2f):
// the alternative replaces only the mana cost, so a required additional cost
// carries onto the alternative option.
func TestManaAlternativeCostPreservesAdditionalCosts(t *testing.T) {
	t.Parallel()
	additional := cost.Additional{
		Kind:   cost.AdditionalSacrifice,
		Text:   "sacrifice a creature",
		Amount: 1,
	}
	card := &game.CardDef{CardFace: game.CardFace{
		Name:            "Additional Test",
		ManaCost:        opt.Val(cost.Mana{cost.O(3), cost.W}),
		Types:           []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{additional},
		AlternativeCosts: []cost.Alternative{{
			Label:            "Pay {W}",
			ManaCost:         opt.Val(cost.Mana{cost.W}),
			Condition:        cost.AlternativeConditionCreaturesAttacking,
			ConditionCount:   1,
			ConditionExactly: true,
		}},
	}}

	state := manaAlternativeState{attackingCreatures: 1}
	options := spellCostOptionsForZoneAndKicker(state, game.Player1, card, zone.Hand, false, 0, false, nil)
	alternative, ok := spellOptionByLabel(options, "Pay {W}")
	if !ok {
		t.Fatal("mana alternative option missing")
	}
	if len(alternative.additionalCosts) != 1 || alternative.additionalCosts[0].Kind != cost.AdditionalSacrifice {
		t.Fatalf("alternative additional costs = %#v, want the required sacrifice preserved", alternative.additionalCosts)
	}
}

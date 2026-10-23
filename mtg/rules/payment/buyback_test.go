package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestBuybackCostOptionsPreserveAlternativeAndLockedTotals(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		Name: "Buyback payment", Types: []types.Card{types.Instant},
		ManaCost:         opt.Val(cost.Mana{cost.O(2)}),
		AlternativeCosts: []cost.Alternative{{Label: "Alternate", ManaCost: opt.Val(cost.Mana{cost.O(1)})}},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{Cost: cost.Mana{cost.O(3)}}},
		}},
	}}
	request := redirectRequest(game.Player1, card)
	state := choicePaymentState{life: 40, pool: poolOf(manaUnits(mana.C, 10)),
		modifiers: []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 2, GenericReduction: 1}}}
	options := spellCostOptionsForRequest(state, request)
	if len(options) != 4 {
		t.Fatalf("normal/alternative buyback expansion has %d options, want 4", len(options))
	}
	for i, want := range []struct {
		paid        bool
		alternative int
		total       int
	}{{false, 0, 3}, {true, 0, 6}, {false, 1, 2}, {true, 1, 5}} {
		prepared := New(state).PrepareSpellCosts(request)[i]
		request.PreparedCost = &prepared
		request.Prefs = &Preferences{AlternativeIndex: prepared.Summary.Index}
		plan, ok := buildSpellCostPlan(state, request)
		if !ok || plan.option.index != i || plan.option.alternativeIndex != want.alternative ||
			plan.option.buybackPaid != want.paid || genericCostAmount(plan.option.manaCost) != want.total {
			t.Fatalf("prepared option %d changed announcement/total: %+v", i, plan.option)
		}
		// Payment of a prepared total must not expand optional costs again or
		// recompute taxes/reductions after mana abilities alter their sources.
		changed := state
		changed.modifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 100}}
		plan, ok = buildSpellCostPlan(changed, request)
		if !ok || genericCostAmount(plan.option.manaCost) != want.total {
			t.Fatalf("locked buyback total changed after preparation: %+v", plan.option)
		}
		request.PreparedCost = nil
		request.Prefs = nil
	}
}

func TestBuybackWithoutManaCostDoesNotCreateCastPermission(t *testing.T) {
	t.Parallel()
	card := &game.CardDef{CardFace: game.CardFace{
		Name: "No mana cost", Types: []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{Cost: cost.Mana{cost.O(3)}}},
		}},
	}}
	request := redirectRequest(game.Player1, card)
	options := spellCostOptionsForRequest(fakePaymentState{}, request)
	if len(options) != 1 || options[0].manaCost != nil || options[0].buybackPaid {
		t.Fatal("buyback substituted for an unpayable absent mana cost")
	}
}

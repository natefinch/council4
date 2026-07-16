package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// redirectLightningCard builds the Redirect Lightning card definition: printed
// cost {R} plus the printed choice among alternative additional costs "pay 5
// life or pay {2}". It mirrors the curated mtg/cards/r/redirect_lightning.go so
// the payment planner is exercised on the real card's cost shape.
func redirectLightningCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Redirect Lightning",
			ManaCost: opt.Val(cost.Mana{cost.R}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			AdditionalCostChoices: []cost.AdditionalChoice{{
				Options: []cost.AdditionalChoiceOption{
					{
						Label: "Pay 5 life",
						Costs: []cost.Additional{{
							Kind:   cost.AdditionalPayLife,
							Text:   "pay 5 life",
							Amount: 5,
						}},
					},
					{
						Label: "Pay {2}",
						Mana:  cost.Mana{cost.O(2)},
					},
				},
			}},
		},
	}
}

// choicePaymentState is a configurable State for exercising additional-cost
// choice payment. It reports one caster life total and one floating mana pool
// (fakePaymentState has no tappable sources, so the pool is the only mana), plus
// optional spell cost modifiers, inheriting every other method.
type choicePaymentState struct {
	fakePaymentState

	life      int
	pool      mana.Pool
	modifiers []game.CostModifier
}

func (s choicePaymentState) Player(playerID game.PlayerID) (*game.Player, bool) {
	return &game.Player{ID: playerID, Life: s.life, ManaPool: s.pool}, true
}

func (s choicePaymentState) CostModifiersForSpell(game.PlayerID, *game.CardDef, id.ID, zone.Type, []game.Target, bool, bool) []game.CostModifier {
	return s.modifiers
}

// poolOf builds a floating mana pool from color/amount pairs.
func poolOf(pairs ...struct {
	color  mana.Color
	amount int
}) mana.Pool {
	pool := mana.NewPool()
	for _, pair := range pairs {
		pool.Add(pair.color, pair.amount)
	}
	return pool
}

func manaUnits(col mana.Color, amount int) struct {
	color  mana.Color
	amount int
} {
	return struct {
		color  mana.Color
		amount int
	}{col, amount}
}

func redirectRequest(caster game.PlayerID, card *game.CardDef) SpellRequest {
	return SpellRequest{
		PlayerID:   caster,
		CardID:     id.ID(1),
		SourceZone: zone.Hand,
		Card:       card,
	}
}

// TestRedirectLightningExpandsToTwoStableBranches proves the printed additional
// cost choice expands into exactly two payable cost options with stable
// sequential indices: a "Pay 5 life" branch that keeps the printed {R} cost and
// carries the pay-5-life additional cost, and a "Pay {2}" branch whose {2} is
// added to (never replacing) the printed {R}. The printed cost surviving on both
// branches is what lets taxes and reductions apply to each.
func TestRedirectLightningExpandsToTwoStableBranches(t *testing.T) {
	t.Parallel()
	card := redirectLightningCard()
	options := spellCostOptionsForRequest(fakePaymentState{}, redirectRequest(game.Player1, card))
	if len(options) != 2 {
		t.Fatalf("options = %d, want two additional-cost choice branches", len(options))
	}

	life := options[0]
	if life.index != 0 {
		t.Fatalf("life branch index = %d, want 0", life.index)
	}
	if life.label != "Pay 5 life" {
		t.Fatalf("life branch label = %q, want \"Pay 5 life\"", life.label)
	}
	if life.manaCost == nil || !slices.Equal(*life.manaCost, cost.Mana{cost.R}) {
		t.Fatalf("life branch mana = %#v, want printed {R} unchanged", life.manaCost)
	}
	if len(life.additionalCosts) != 1 || life.additionalCosts[0].Kind != cost.AdditionalPayLife || life.additionalCosts[0].Amount != 5 {
		t.Fatalf("life branch additional costs = %#v, want a single pay-5-life", life.additionalCosts)
	}

	pay := options[1]
	if pay.index != 1 {
		t.Fatalf("mana branch index = %d, want 1", pay.index)
	}
	if pay.label != "Pay {2}" {
		t.Fatalf("mana branch label = %q, want \"Pay {2}\"", pay.label)
	}
	if pay.manaCost == nil || !slices.Equal(*pay.manaCost, cost.Mana{cost.R, cost.O(2)}) {
		t.Fatalf("mana branch mana = %#v, want printed {R} plus additive {2}", pay.manaCost)
	}
	if len(pay.additionalCosts) != 0 {
		t.Fatalf("mana branch additional costs = %#v, want none", pay.additionalCosts)
	}
}

// TestRedirectLightningBranchPayability proves the payment planner enforces each
// branch's payability independently: a branch is offered only when the caster
// can actually pay it. Insufficient life removes the life branch, insufficient
// mana removes the mana branch, and the choice survives for a non-active-player
// caster in a multiplayer game.
func TestRedirectLightningBranchPayability(t *testing.T) {
	t.Parallel()
	card := redirectLightningCard()

	labels := func(summaries []SpellOptionSummary) []string {
		out := make([]string, 0, len(summaries))
		for _, summary := range summaries {
			out = append(out, summary.Label)
		}
		return out
	}

	t.Run("both branches available", func(t *testing.T) {
		t.Parallel()
		state := choicePaymentState{life: 40, pool: poolOf(manaUnits(mana.R, 1), manaUnits(mana.C, 2))}
		got := labels(payableSpellOptionsFromState(state, redirectRequest(game.Player1, card)))
		if !slices.Equal(got, []string{"Pay 5 life", "Pay {2}"}) {
			t.Fatalf("payable branches = %v, want both", got)
		}
	})

	t.Run("insufficient life drops the life branch", func(t *testing.T) {
		t.Parallel()
		state := choicePaymentState{life: 3, pool: poolOf(manaUnits(mana.R, 1), manaUnits(mana.C, 2))}
		got := labels(payableSpellOptionsFromState(state, redirectRequest(game.Player1, card)))
		if !slices.Equal(got, []string{"Pay {2}"}) {
			t.Fatalf("payable branches with 3 life = %v, want only the mana branch", got)
		}
	})

	t.Run("insufficient mana drops the mana branch", func(t *testing.T) {
		t.Parallel()
		state := choicePaymentState{life: 40, pool: poolOf(manaUnits(mana.R, 1))}
		got := labels(payableSpellOptionsFromState(state, redirectRequest(game.Player1, card)))
		if !slices.Equal(got, []string{"Pay 5 life"}) {
			t.Fatalf("payable branches with only {R} = %v, want only the life branch", got)
		}
	})

	t.Run("neither branch payable", func(t *testing.T) {
		t.Parallel()
		state := choicePaymentState{life: 3, pool: poolOf(manaUnits(mana.R, 1))}
		got := payableSpellOptionsFromState(state, redirectRequest(game.Player1, card))
		if len(got) != 0 {
			t.Fatalf("payable branches = %v, want none", labels(got))
		}
	})

	t.Run("non-active caster in multiplayer", func(t *testing.T) {
		t.Parallel()
		// Player2 casts at instant speed while Player1 is active; the caster's own
		// life and mana are what gate the branches.
		state := choicePaymentState{life: 40, pool: poolOf(manaUnits(mana.R, 1), manaUnits(mana.C, 2))}
		got := labels(payableSpellOptionsFromState(state, redirectRequest(game.Player2, card)))
		if !slices.Equal(got, []string{"Pay 5 life", "Pay {2}"}) {
			t.Fatalf("payable branches for Player2 caster = %v, want both", got)
		}
	})
}

// TestAdditionalCostChoiceTaxAndReductionApplyToEachBranch proves a spell cost
// tax or reduction folds onto every additional-cost choice branch, because each
// branch keeps the printed cost. A {2} tax raises both the {R} life branch and
// the {R}{2} mana branch; a {2} generic reduction lowers the mana branch and
// floors the life branch at its colored requirement (generic never goes below
// zero and the {R} is untouched).
func TestAdditionalCostChoiceTaxAndReductionApplyToEachBranch(t *testing.T) {
	t.Parallel()
	card := redirectLightningCard()
	options := spellCostOptionsForRequest(fakePaymentState{}, redirectRequest(game.Player1, card))
	if len(options) != 2 {
		t.Fatalf("options = %d, want two branches", len(options))
	}
	lifeBranch, manaBranch := options[0], options[1]

	apply := func(modifier game.CostModifier, option spellCostOption) cost.Mana {
		state := choicePaymentState{life: 40, modifiers: []game.CostModifier{modifier}}
		modified := applyCostModifiers(state, costModificationContext{
			player:     game.Player1,
			card:       card,
			cardID:     id.ID(1),
			sourceZone: zone.Hand,
			option:     option,
		})
		if modified.manaCost == nil {
			return nil
		}
		return *modified.manaCost
	}

	tax := game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 2}
	if got := apply(tax, lifeBranch); !slices.Equal(got, cost.Mana{cost.O(2), cost.R}) {
		t.Fatalf("taxed life branch = %#v, want {2}{R}", got)
	}
	if got := apply(tax, manaBranch); !slices.Equal(got, cost.Mana{cost.O(4), cost.R}) {
		t.Fatalf("taxed mana branch = %#v, want {4}{R}", got)
	}

	reduction := game.CostModifier{Kind: game.CostModifierSpell, GenericReduction: 2}
	if got := apply(reduction, lifeBranch); !slices.Equal(got, cost.Mana{cost.R}) {
		t.Fatalf("reduced life branch = %#v, want {R} (generic floored at zero, colored untouched)", got)
	}
	if got := apply(reduction, manaBranch); !slices.Equal(got, cost.Mana{cost.R}) {
		t.Fatalf("reduced mana branch = %#v, want {R} ({2} reduced away)", got)
	}
}

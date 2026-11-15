package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func TestCompletePlanOwnsAbilityCostMetadata(t *testing.T) {
	for _, variable := range []bool{false, true} {
		t.Run(map[bool]string{false: "fixed", true: "X"}[variable], func(t *testing.T) {
			source := &game.Permanent{ObjectID: 1, Controller: game.Player1}
			state := choicePaymentState{fakePaymentState: fakePaymentState{battlefield: []*game.Permanent{source}}, life: 40, pool: poolOf(manaUnits(mana.C, 1))}
			modifiedCost := cost.Mana{cost.O(1)}
			req := AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				CostModifiers: []game.CostModifier{{SetManaCost: opt.Val(modifiedCost)}}}
			if variable {
				modifiedCost[0] = cost.X
				req.XValue = 1
			}
			request := ForAbility(req)
			planner := New(state)
			plan, ok, reason := planner.Complete(request, Partial{})
			if !ok {
				t.Fatalf("Complete: %s", reason)
			}
			req.CostModifiers[0].SetManaCost.Val[0] = cost.O(2)
			if _, ok, reason := planner.Apply(plan); !ok {
				t.Fatalf("unchanged-game witness borrowed caller metadata: %s", reason)
			}
		})
	}
}

type countedCompletionState struct {
	fakePaymentState
	queries int
}

func (s *countedCompletionState) PermanentMatchesSelection(*game.Permanent, game.Selection) bool {
	s.queries++
	return true
}

func TestImpossibleFixedCostsRejectWithoutSubsetEnumeration(t *testing.T) {
	for _, second := range []int{6, 7} {
		state := &countedCompletionState{}
		for i := range 12 {
			state.battlefield = append(state.battlefield, &game.Permanent{ObjectID: id.ID(i + 1), Controller: game.Player1})
		}
		request := ForGeneric(GenericRequest{PlayerID: game.Player1, AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Amount: 6},
			{Kind: cost.AdditionalSacrifice, Amount: second},
		}})
		if _, ok, _ := New(state).Complete(request, Partial{}); ok != (second == 6) {
			t.Fatalf("six plus %d consuming selections from 12: Complete = %v", second, ok)
		}
		// Neither a contradiction nor an exact assignment should generate the
		// Cartesian product of fixed-count subsets (review regression).
		if state.queries > 1000 {
			t.Fatalf("fixed costs required %d eligibility reads", state.queries)
		}
	}
}

func TestComponentOptionsAdvanceOneObjectWithWitness(t *testing.T) {
	state := &countedCompletionState{}
	for i := range 12 {
		state.battlefield = append(state.battlefield, &game.Permanent{ObjectID: id.ID(i + 1), Controller: game.Player1})
	}
	planner := New(state)
	request := ForGeneric(GenericRequest{PlayerID: game.Player1, AdditionalCosts: []cost.Additional{
		{Kind: cost.AdditionalSacrifice, Amount: 6},
		{Kind: cost.AdditionalSacrifice, Amount: 6},
	}})
	step := Step{Kind: StepComponent, Index: 0}
	choices := planner.Options(request, Partial{}, step)
	if len(choices) != 12 {
		t.Fatalf("next-object choices = %d, want 12, not all six-element subsets", len(choices))
	}
	for _, choice := range choices {
		if len(choice.Component.Objects) != 1 || choice.Component.Finished || len(choice.Plan.Components()[0].Objects) != 6 {
			t.Fatal("next object did not retain a complete witness")
		}
	}
	partial := Partial{}
	for range 6 {
		choices = planner.Options(request, partial, step)
		if len(choices) == 0 {
			t.Fatal("a completable prefix lost all next choices")
		}
		partial = partial.With(step, choices[0])
	}
	if !partial.Components[0].Finished {
		t.Fatal("the sixth object did not finish the fixed-count component")
	}
	step.Index = 1
	choices = planner.Options(request, partial, step)
	if len(choices) != 6 {
		t.Fatalf("second component choices = %d, want only six unconsumed objects", len(choices))
	}
}

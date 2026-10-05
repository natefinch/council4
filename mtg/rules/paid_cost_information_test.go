package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostManaValuePreservesLegacyEventSnapshots(t *testing.T) {
	subject := selectionSubject{
		kind: subjectEventPermanent,
		event: game.Event{TokenDef: &game.CardDef{CardFace: game.CardFace{
			ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		}}},
		snapshot: &game.ObjectSnapshot{},
	}
	value, known := subject.manaValue()
	if !known || value != 3 {
		t.Fatal("additive frozen mana-value metadata changed legacy event-snapshot behavior")
	}
	subject.snapshot.ManaValue = opt.Val(0)
	if value, known := subject.manaValue(); !known || value != 0 {
		t.Fatal("known frozen zero was replaced by printed event traits")
	}
}

func TestPaidCostUnavailablePublicationCannotEnableComplement(t *testing.T) {
	r := &effectResolver{
		game: game.NewGame([game.NumPlayers]game.PlayerConfig{}),
		obj:  &game.StackObject{Controller: game.Player1},
	}

	publisher := game.Instruction{
		Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
			Object: opt.Val(game.PaidCostReference("missing", game.PaidCostDiscard, "")),
		})}),
		PublishCondition: "payment-traits",
	}
	for _, stale := range []bool{false, true} {
		r.conditionEvaluations = map[game.ConditionKey]bool{"payment-traits": stale}
		if r.instructionConditionSatisfied(&publisher) {
			t.Fatal("unavailable cost facts satisfied the publisher")
		}
		if _, published := r.conditionEvaluations["payment-traits"]; published {
			t.Fatal("unavailable cost facts published a success-shaped boolean")
		}
		if r.instructionConditionSatisfied(&game.Instruction{
			ConditionGate: "payment-traits", ConditionGateNegate: true,
		}) {
			t.Fatal("complement consumed an unavailable paid-cost predicate")
		}
	}
}

func TestPaidCostSharedAvailabilityRejectsUnavailablePredicateDomains(t *testing.T) {
	obj := &game.StackObject{Controller: game.Player1, PaidCostSubjects: []game.PaidCostSubject{{
		Key: "cost", Kind: game.PaidCostSacrifice, CharacteristicsKnown: true,
		Snapshot: game.ObjectSnapshot{ObjectID: 4, Types: []types.Card{types.Creature}},
	}}}
	for _, selection := range []game.Selection{
		{Power: opt.Val(compare.Int{Op: compare.Equal})},
		{AnyOf: []game.Selection{{RequiredTypes: []types.Card{types.Creature}},
			{Power: opt.Val(compare.Int{Op: compare.Equal})}}},
		{Tapped: game.TriFalse},
		{MatchNoCounters: true},
		{Owner: game.OwnerYou},
	} {
		condition := game.Condition{
			Object:        opt.Val(game.PaidCostReference("cost", game.PaidCostSacrifice, "")),
			ObjectMatches: opt.Val(selection), Negate: true,
		}
		r := &effectResolver{game: game.NewGame([game.NumPlayers]game.PlayerConfig{}), obj: obj,
			conditionEvaluations: map[game.ConditionKey]bool{"traits": false}}
		publisher := game.Instruction{
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(condition)}), PublishCondition: "traits",
		}
		if objectConditionInformationAvailable(r.game, conditionContext{obj: obj}, &condition) ||
			r.instructionConditionSatisfied(&publisher) {
			t.Fatal("unavailable characteristic/domain enabled a negated paid-cost predicate")
		}
		if _, published := r.conditionEvaluations["traits"]; published {
			t.Fatal("unavailable paid-cost predicate left a complement-enabling boolean")
		}
	}
}

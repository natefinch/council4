package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestValidateCardDefRejectsNonPositiveOpponentCastSpellsCount proves the
// opponent-cast-spells alternative cost requires a positive threshold: a
// non-positive ConditionCount is an invalid alternative cost, while the real
// "three or more" threshold validates.
func TestValidateCardDefRejectsNonPositiveOpponentCastSpellsCount(t *testing.T) {
	t.Parallel()
	newCard := func(count int) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:     "Opponent Cast Alt",
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U, cost.U}),
			Types:    []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{{
				Label:          "Pay {0}",
				ManaCost:       opt.Val(cost.Mana{cost.O(0)}),
				Condition:      cost.AlternativeConditionOpponentCastSpellsThisTurn,
				ConditionCount: count,
			}},
		}}
	}

	if issues := ValidateCardDef(newCard(0)); !hasCardDefIssue(issues, CardDefIssueInvalidAlternativeCost) {
		t.Fatalf("count 0 issues = %+v, want %s", issues, CardDefIssueInvalidAlternativeCost)
	}
	if issues := ValidateCardDef(newCard(3)); hasCardDefIssue(issues, CardDefIssueInvalidAlternativeCost) {
		t.Fatalf("count 3 issues = %+v, want no invalid-alternative-cost issue", issues)
	}
}

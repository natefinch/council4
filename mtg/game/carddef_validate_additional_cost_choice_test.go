package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// redirectLifeOrManaChoice is the Redirect Lightning shaped choice: pay 5 life
// or pay an additive {2}.
func redirectLifeOrManaChoice() cost.AdditionalChoice {
	return cost.AdditionalChoice{
		Options: []cost.AdditionalChoiceOption{
			{
				Label: "Pay 5 life",
				Costs: []cost.Additional{{Kind: cost.AdditionalPayLife, Text: "pay 5 life", Amount: 5}},
			},
			{
				Label: "Pay {2}",
				Mana:  cost.Mana{cost.O(2)},
			},
		},
	}
}

func TestValidateCardDefAcceptsAdditionalCostChoice(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:                  "Redirect Lightning",
		ManaCost:              opt.Val(cost.Mana{cost.R}),
		Types:                 []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{redirectLifeOrManaChoice()},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("valid additional-cost choice issues = %+v, want none", issues)
	}
}

func TestValidateCardDefRejectsAdditionalCostChoiceWithAlternativeCosts(t *testing.T) {
	// The additive additional-cost choice and the printed-cost-replacing
	// alternative costs both key off the expanded cost-option index, so a face
	// carrying both must fail closed rather than risk mis-paying.
	card := &CardDef{CardFace: CardFace{
		Name:                  "Overloaded Choice",
		ManaCost:              opt.Val(cost.Mana{cost.R}),
		Types:                 []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{redirectLifeOrManaChoice()},
		AlternativeCosts: []cost.Alternative{{
			Label:    "Pay {0}",
			ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAdditionalCostChoice) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAdditionalCostChoice)
	}
}

func TestValidateCardDefRejectsSingleBranchAdditionalCostChoice(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:     "Lonely Branch",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{{
			Options: []cost.AdditionalChoiceOption{{Label: "Pay {2}", Mana: cost.Mana{cost.O(2)}}},
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAdditionalCostChoice) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAdditionalCostChoice)
	}
}

func TestValidateCardDefRejectsEmptyAdditionalCostChoiceBranch(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:     "Empty Branch",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{{
			Options: []cost.AdditionalChoiceOption{
				{Label: "Pay {2}", Mana: cost.Mana{cost.O(2)}},
				{Label: "Pay nothing"},
			},
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAdditionalCostChoice) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAdditionalCostChoice)
	}
}

func TestValidateCardDefRejectsAdditionalCostChoiceBranchCarryingChoiceGroup(t *testing.T) {
	// The branch itself is the choice, so a branch cost must not also carry the
	// comparable-cost ChoiceGroup tag (that would double-encode the choice).
	card := &CardDef{CardFace: CardFace{
		Name:     "Double Encoded",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCostChoices: []cost.AdditionalChoice{{
			Options: []cost.AdditionalChoiceOption{
				{Label: "Pay {2}", Mana: cost.Mana{cost.O(2)}},
				{
					Label: "Sacrifice a creature",
					Costs: []cost.Additional{{
						Kind:        cost.AdditionalSacrifice,
						Text:        "sacrifice a creature",
						Amount:      1,
						ChoiceGroup: 1,
					}},
				},
			},
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAdditionalCostChoice) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAdditionalCostChoice)
	}
}

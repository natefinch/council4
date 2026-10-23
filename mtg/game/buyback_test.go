package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestBuybackKeywordValidationAndClone(t *testing.T) {
	keyword := BuybackKeyword{Cost: cost.Mana{cost.O(3)},
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}}
	def := &CardDef{CardFace: CardFace{
		Name: "Buyback declaration", Types: []types.Card{types.Instant},
		StaticAbilities: []StaticAbility{{
			ZoneOfFunction: zone.Stack, KeywordAbilities: []KeywordAbility{keyword},
		}},
	}}
	if issues := ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("inspectable buyback declaration rejected: %+v", issues)
	}
	read, ok := def.BuybackKeyword()
	if !ok || !def.HasKeyword(Buyback) || read.Cost[0] != cost.O(3) {
		t.Fatal("buyback declaration not discoverable")
	}
	cloned := keyword.cloneKeywordAbility().(BuybackKeyword)
	cloned.Cost[0] = cost.O(9)
	cloned.AdditionalCosts[0].Amount = 9
	if keyword.Cost[0] != cost.O(3) || keyword.AdditionalCosts[0].Amount != 1 {
		t.Fatal("keyword clone aliases mutable cost slices")
	}
	for _, invalid := range []struct {
		name     string
		types    []types.Card
		keywords []KeywordAbility
	}{
		{name: "empty", types: []types.Card{types.Instant}, keywords: []KeywordAbility{BuybackKeyword{}}},
		{name: "unparameterized", types: []types.Card{types.Instant}, keywords: SimpleKeywords(Buyback)},
		{name: "multiple", types: []types.Card{types.Instant}, keywords: []KeywordAbility{keyword, keyword}},
		{name: "buyback X", types: []types.Card{types.Instant}, keywords: []KeywordAbility{BuybackKeyword{Cost: cost.Mana{cost.X}}}},
		{name: "buyback non-mana X", types: []types.Card{types.Instant}, keywords: []KeywordAbility{BuybackKeyword{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, AmountFromX: true}},
		}}},
		{name: "permanent", types: []types.Card{types.Creature}, keywords: []KeywordAbility{keyword}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			def.Types = invalid.types
			def.StaticAbilities[0].KeywordAbilities = invalid.keywords
			if !hasCardDefIssue(ValidateCardDef(def), CardDefIssueInvalidKeywordAbility) {
				t.Fatal("unsupported buyback declaration passed validation")
			}
		})
	}
}

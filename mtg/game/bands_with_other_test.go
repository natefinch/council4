package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: official CR effective 2026-09-25 (verified copy retained
// under .cardwork), 702.22a: "bands with other [quality]" is a parameterized
// form of banding; 702.22m: multiple "bands with other" abilities of the same
// kind are redundant. The rules scenarios live in
// mtg/rules/combat_bands_with_other_test.go.

// TestBandsWithOtherKeywordValidationAndClone checks that the "bands with
// other" keyword carries its quality, needs a non-empty valid quality, reports
// the BandsWithOther keyword and clones its quality without sharing storage.
func TestBandsWithOtherKeywordValidationAndClone(t *testing.T) {
	t.Parallel()
	legendary := Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}}
	card := func(quality Selection) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:            "Bands Test Creature",
			Types:           []types.Card{types.Creature},
			StaticAbilities: []StaticAbility{BandsWithOtherStaticAbility(quality, "legendary creatures")},
		}}
	}

	if issues := ValidateCardDef(card(legendary)); hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("valid bands with other rejected: %+v", issues)
	}
	if issues := ValidateCardDef(card(Selection{})); !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("bands with other without a quality accepted: %+v", issues)
	}

	body := BandsWithOtherStaticAbility(legendary, "legendary creatures")
	if body.Text != "Bands with other legendary creatures" {
		t.Fatalf("text = %q", body.Text)
	}
	keywords := StaticBodyBandsWithOtherKeywords(&body)
	if len(keywords) != 1 || KeywordAbilityKind(keywords[0]) != BandsWithOther || len(keywords[0].Quality.Supertypes) != 1 {
		t.Fatalf("body keywords = %+v; want bands with other legendary creatures", keywords)
	}
	keyword := keywords[0]
	dinosaurs := Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Dinosaur}}
	both := StaticAbility{KeywordAbilities: []KeywordAbility{BandsWithOtherKeyword{Quality: legendary}, BandsWithOtherKeyword{Quality: dinosaurs}}}
	if got := StaticBodyBandsWithOtherKeywords(&both); len(got) != 2 || len(got[1].Quality.SubtypesAny) != 1 {
		t.Fatalf("two-quality body keywords = %+v; want both", got)
	}
	if _, ok := StaticBodyLandwalkKeyword(&body); ok {
		t.Fatal("bands with other read as landwalk")
	}

	cloned := cloneKeywordAbilities(body.KeywordAbilities)
	cloned[0].(BandsWithOtherKeyword).Quality.Supertypes[0] = types.Basic
	if keyword.Quality.Supertypes[0] != types.Legendary {
		t.Fatal("cloned quality shares storage with the original")
	}
}

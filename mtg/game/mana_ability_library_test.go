package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3690. Comprehensive Rules effective 2026-09-25,
// 605.1a: an activated ability is a mana ability only if, among other
// criteria, it doesn't move cards to or from a library as part of its cost or
// effect, ignoring replacement effects other than self-replacement effects.
// Millikin's "{T}, Mill a card: Add {C}." is the ruled example (Millikin
// ruling 2026-08-05: "Millikin's ability is not a mana ability."); revealing
// a library's top card leaves it there.

func manaAbilityCard(ability ManaAbility) *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:          "Mana Test",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []ManaAbility{ability},
	}}
}

func addColorlessContent(extra ...Instruction) AbilityContent {
	sequence := append([]Instruction{{Primitive: AddMana{Amount: Fixed(1), ManaColor: mana.C}}}, extra...)
	return AbilityContent{Modes: []Mode{{Sequence: sequence}}, MinModes: 1, MaxModes: 1}
}

func libraryMovementIssues(card *CardDef) []CardDefIssue {
	var found []CardDefIssue
	for _, issue := range ValidateCardDef(card) {
		if strings.Contains(issue.Message, "CR 605.1a") {
			found = append(found, issue)
		}
	}
	return found
}

// TestValidateCardDefRejectsLibraryMovingManaAbility proves a mana ability
// definition whose cost or effect moves a card to or from a library, directly
// or through a granted ability, is rejected so it is modelled as an activated
// ability that uses the stack (CR 605.1a).
func TestValidateCardDefRejectsLibraryMovingManaAbility(t *testing.T) {
	t.Parallel()
	tap := cost.Additional{Kind: cost.AdditionalTap}
	tests := []struct {
		name string
		card *CardDef
	}{
		{"mill cost", manaAbilityCard(ManaAbility{
			Text:            "{T}, Mill a card: Add {C}.",
			AdditionalCosts: []cost.Additional{tap, {Kind: cost.AdditionalMill, Text: "Mill a card", Amount: 1}},
			Content:         addColorlessContent(),
		})},
		{"exile from library cost", manaAbilityCard(ManaAbility{
			Text:            "{T}, Exile the top card of your library: Add {C}.",
			AdditionalCosts: []cost.Additional{tap, {Kind: cost.AdditionalExile, Source: zone.Library, Amount: 1}},
			Content:         addColorlessContent(),
		})},
		{"draw effect", manaAbilityCard(ManaAbility{
			Text:            "{T}: Add {C}. Draw a card.",
			AdditionalCosts: []cost.Additional{tap},
			Content:         addColorlessContent(Instruction{Primitive: Draw{Amount: Fixed(1)}}),
		})},
		{"granted mill cost", &CardDef{CardFace: CardFace{
			Name:  "Mana Grant Test",
			Types: []types.Card{types.Enchantment},
			StaticAbilities: []StaticAbility{{
				ZoneOfFunction: zone.Battlefield,
				ContinuousEffects: []ContinuousEffect{{
					Duration: DurationPermanent,
					Group:    ObjectControlledGroup(SourcePermanentReference(), Selection{RequiredTypes: []types.Card{types.Creature}}),
					Layer:    LayerAbility,
					AddAbilities: []Ability{&ManaAbility{
						Text:            "{T}, Mill a card: Add {C}.",
						AdditionalCosts: []cost.Additional{tap, {Kind: cost.AdditionalMill, Text: "Mill a card", Amount: 1}},
						Content:         addColorlessContent(),
					}},
				}},
			}},
		}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if issues := libraryMovementIssues(tc.card); len(issues) != 1 || issues[0].Code != CardDefIssueInvalidAbilityBody {
				t.Fatalf("CR 605.1a issues = %+v, want one invalid-ability-body issue", issues)
			}
		})
	}
}

// TestValidateCardDefAllowsLibraryNeutralManaAbility proves mana abilities
// whose costs and effects leave every library alone still validate: paying
// life, sacrificing a creature, and revealing a card from the library.
func TestValidateCardDefAllowsLibraryNeutralManaAbility(t *testing.T) {
	t.Parallel()
	tap := cost.Additional{Kind: cost.AdditionalTap}
	for name, ability := range map[string]ManaAbility{
		"pay life": {
			Text:            "{T}, Pay 1 life: Add {C}.",
			AdditionalCosts: []cost.Additional{tap, {Kind: cost.AdditionalPayLife, Amount: 1}},
			Content:         addColorlessContent(),
		},
		"sacrifice": {
			Text:            "Sacrifice a creature: Add {C}.",
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature}},
			Content:         addColorlessContent(),
		},
		"reveal from library": {
			Text:            "{T}, Reveal the top card of your library: Add {C}.",
			AdditionalCosts: []cost.Additional{tap, {Kind: cost.AdditionalReveal, Source: zone.Library, Amount: 1}},
			Content:         addColorlessContent(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if issues := libraryMovementIssues(manaAbilityCard(ability)); len(issues) != 0 {
				t.Fatalf("CR 605.1a issues = %+v, want none", issues)
			}
		})
	}
}

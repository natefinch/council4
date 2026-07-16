package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestValidateKickedModalReplacementRange(t *testing.T) {
	t.Parallel()
	content := AbilityContent{
		Modes:    []Mode{{}, {}, {}},
		MinModes: 1,
		MaxModes: 1,
		ModeChoiceBonus: ModeChoiceBonus{
			Condition:    ModeChoiceConditionSpellKicked,
			ReplaceRange: true,
			MinModes:     0,
			MaxModes:     3,
		},
	}
	card := func(content AbilityContent) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:         "Kicked Modal",
			Types:        []types.Card{types.Instant},
			SpellAbility: opt.Val(content),
		}}
	}
	if issues := ValidateCardDef(card(content)); len(issues) != 0 {
		t.Fatalf("valid replacement range issues = %+v", issues)
	}
	for name, mutate := range map[string]func(*AbilityContent){
		"not replacement": func(content *AbilityContent) {
			content.ModeChoiceBonus.ReplaceRange = false
		},
		"negative minimum": func(content *AbilityContent) {
			content.ModeChoiceBonus.MinModes = -1
		},
		"maximum below minimum": func(content *AbilityContent) {
			content.ModeChoiceBonus.MinModes = 2
			content.ModeChoiceBonus.MaxModes = 1
		},
		"maximum exceeds modes": func(content *AbilityContent) {
			content.ModeChoiceBonus.MaxModes = 4
		},
		"additive and replacement": func(content *AbilityContent) {
			content.ModeChoiceBonus.AdditionalMaxModes = 1
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			invalid := content
			mutate(&invalid)
			if issues := ValidateCardDef(card(invalid)); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
			}
		})
	}
}

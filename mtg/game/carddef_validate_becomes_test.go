package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestValidateCardDefAllowsEntersBecomesReplacement proves a well-formed group
// ETB characteristic replacement (Displaced Dinosaurs) passes validation.
func TestValidateCardDefAllowsEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name:       "Displaced Dinosaurs",
		Types:      []types.Card{types.Enchantment},
		OracleText: "As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types.",
		ReplacementAbilities: []ReplacementAbility{
			EntersBecomesGroupReplacement(
				"As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types.",
				EntersBecomesGroupParams{
					Controller:    TriggerControllerYou,
					Historic:      true,
					AddTypes:      []types.Card{types.Creature},
					AddSubtypes:   []types.Sub{types.Dinosaur},
					BasePower:     opt.Val(7),
					BaseToughness: opt.Val(7),
				},
			),
		},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

// TestValidateCardDefRejectsEmptyEntersBecomesReplacement proves an
// enters-becomes replacement that grants no characteristics is flagged.
func TestValidateCardDefRejectsEmptyEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name:  "Empty Becomes",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []ReplacementAbility{{
			Replacement: ReplacementEffect{EntersBecomesCharacteristic: true},
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

// TestValidateCardDefRejectsHalfBasePTEntersBecomesReplacement proves an
// enters-becomes replacement that sets only one of base power or toughness is
// flagged, so an entrant never gets a half-defined P/T.
func TestValidateCardDefRejectsHalfBasePTEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name:  "Half PT Becomes",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []ReplacementAbility{{
			Replacement: ReplacementEffect{
				EntersBecomesCharacteristic: true,
				EntersBecomesAddTypes:       []types.Card{types.Creature},
				EntersBecomesBasePower:      opt.Val(7),
			},
		}},
	}}
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

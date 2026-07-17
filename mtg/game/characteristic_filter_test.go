package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestCloneCharacteristicFiltersDeepCopiesBranches(t *testing.T) {
	original := []CharacteristicFilter{{
		Types:      []types.Card{types.Artifact},
		Supertypes: []types.Super{types.Legendary},
		Subtypes:   []types.Sub{types.Wizard},
	}}
	cloned := cloneCharacteristicFilters(original)
	cloned[0].Types[0] = types.Creature
	cloned[0].Supertypes[0] = types.Basic
	cloned[0].Subtypes[0] = types.Avatar
	if original[0].Types[0] != types.Artifact ||
		original[0].Supertypes[0] != types.Legendary ||
		original[0].Subtypes[0] != types.Wizard {
		t.Fatalf("clone mutated original: %#v", original)
	}
}

func TestCloneControlledTriggerDoublerSnapshot(t *testing.T) {
	original := Event{
		Kind: EventZoneChanged,
		ControlledTriggerDoublers: &ControlledTriggerDoublerSnapshot{
			Doublers: []ControlledTriggerDoubler{{
				SourceID:        1,
				Controller:      Player1,
				PermanentFilter: []CharacteristicFilter{{Types: []types.Card{types.Artifact}}},
				Enters:          true,
			}},
			PermanentControllers: []PermanentControllerSnapshot{{SourceID: 2, Controller: Player1}},
		},
	}
	cloned := cloneEvent(original)
	cloned.ControlledTriggerDoublers.Doublers[0].PermanentFilter[0].Types[0] = types.Creature
	cloned.ControlledTriggerDoublers.PermanentControllers[0].Controller = Player2

	if got := original.ControlledTriggerDoublers.Doublers[0].PermanentFilter[0].Types[0]; got != types.Artifact {
		t.Fatalf("original filter type = %v, want artifact", got)
	}
	if got := original.ControlledTriggerDoublers.PermanentControllers[0].Controller; got != Player1 {
		t.Fatalf("original controller = %v, want Player1", got)
	}
}

func TestValidateCharacteristicFilteredRuleEffects(t *testing.T) {
	valid := &CardDef{CardFace: CardFace{
		Name:  "Valid",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{RuleEffects: []RuleEffect{
			{
				Kind:           RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer: PlayerYou,
				SpellCharacteristicFilters: []CharacteristicFilter{
					{Supertypes: []types.Super{types.Legendary}},
					{Types: []types.Card{types.Artifact}},
				},
			},
			{
				Kind: RuleEffectAdditionalTriggerForControlledPermanent,
				TriggerCausePermanentFilters: []CharacteristicFilter{
					{Supertypes: []types.Super{types.Legendary}},
					{Types: []types.Card{types.Artifact}},
				},
				TriggerCausePermanentEnters: true,
				TriggerCausePermanentLeaves: true,
			},
		}}},
	}}
	if issues := ValidateCardDef(valid); len(issues) != 0 {
		t.Fatalf("valid characteristic filters produced issues: %#v", issues)
	}

	invalid := *valid
	invalid.StaticAbilities = []StaticAbility{{RuleEffects: []RuleEffect{{
		Kind:                         RuleEffectAdditionalTriggerForControlledPermanent,
		TriggerCausePermanentFilters: []CharacteristicFilter{{}},
	}}}}
	if issues := ValidateCardDef(&invalid); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("empty characteristic branch was accepted: %#v", issues)
	}
}

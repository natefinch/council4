package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestPreventDamageOneShotSourceFilterValidation proves the one-shot
// chosen-source fields are declared only on an all-damage one-shot shield.
func TestPreventDamageOneShotSourceFilterValidation(t *testing.T) {
	t.Parallel()
	valid := PreventDamage{
		Player: ControllerReference(), All: true, OneShot: true,
		SourceTypes: []types.Card{types.Artifact}, SourceColorFromEntryChoice: true,
	}
	if err := valid.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want a filtered one-shot shield accepted", err)
	}
	group := PreventDamage{Player: ControllerReference(), All: true, OneShot: true, AndControlledCreatures: true}
	if err := group.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want a you-and-your-creatures shield accepted", err)
	}
	for name, invalid := range map[string]PreventDamage{
		"one-shot without All":         {Player: ControllerReference(), Amount: Fixed(2), OneShot: true},
		"types without one-shot":       {Player: ControllerReference(), All: true, SourceTypes: []types.Card{types.Land}},
		"entry color without one-shot": {Player: ControllerReference(), All: true, SourceColorFromEntryChoice: true},
		"creatures without player":     {Object: SourcePermanentReference(), All: true, OneShot: true, AndControlledCreatures: true},
		"creatures with by-source":     {Player: ControllerReference(), All: true, BySource: true, AndControlledCreatures: true},
	} {
		if err := invalid.validatePrimitive(nil, false); err == nil {
			t.Errorf("%s: validatePrimitive() = nil, want rejection", name)
		}
	}
}

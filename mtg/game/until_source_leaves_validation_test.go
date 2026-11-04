package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestMovePermanentUntilSourceLeavesRequiresExile proves the CR 610.3
// duration is declared only on an exile.
func TestMovePermanentUntilSourceLeavesRequiresExile(t *testing.T) {
	t.Parallel()
	valid := MovePermanent{Object: TargetPermanentReference(0), Destination: zone.Exile, UntilSourceLeaves: true}
	if err := valid.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want an exile until-source-leaves move accepted", err)
	}
	invalid := MovePermanent{Object: TargetPermanentReference(0), Destination: zone.Hand, UntilSourceLeaves: true}
	err := invalid.validatePrimitive(nil, false)
	if err == nil || !strings.Contains(err.Error(), "until-source-leaves") {
		t.Fatalf("validatePrimitive() = %v, want a non-exile until-source-leaves rejection", err)
	}
}

// TestForEachPlayerUntilSourceLeavesRequiresExile proves the distributive
// route accepts the CR 610.3 duration only on exile removal.
func TestForEachPlayerUntilSourceLeavesRequiresExile(t *testing.T) {
	t.Parallel()
	valid := ForEachPlayer{
		Scope:             AllPlayersReference(),
		Chooser:           ControllerReference(),
		Selection:         Selection{RequiredTypes: []types.Card{types.Creature}},
		Removal:           DistributiveRemovalExile,
		LinkedKey:         LinkedKey("exile-until-leaves"),
		UntilSourceLeaves: true,
	}
	if err := valid.validatePrimitive(nil, false); err != nil {
		t.Fatalf("validatePrimitive() = %v, want an exile until-source-leaves removal accepted", err)
	}
	invalid := valid
	invalid.Removal = DistributiveRemovalDestroy
	err := invalid.validatePrimitive(nil, false)
	if err == nil || !strings.Contains(err.Error(), "until-source-leaves") {
		t.Fatalf("validatePrimitive() = %v, want a destroy until-source-leaves rejection", err)
	}
}

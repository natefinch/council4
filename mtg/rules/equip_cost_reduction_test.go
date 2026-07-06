package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// TestEquipCostReductionGatedByMonarch proves the self-referential equip cost
// reduction "This ability costs {3} less to activate if you're the monarch."
// (Crown of Gondor): the Equip {4} activation costs {4} while its controller is
// not the monarch and {1} once they hold the crown. An opponent's monarchy does
// not reduce the controller's equip cost.
func TestEquipCostReductionGatedByMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	equipBody := game.EquipCostReductionActivatedAbility(cost.Mana{cost.O(4)}, game.CostModifier{
		Kind:               game.CostModifierAbility,
		GenericReduction:   3,
		ReductionCondition: opt.Val(game.Condition{ControllerIsMonarch: true}),
	})

	if got := effectiveActivatedAbilityCost(g, game.Player1, nil, &equipBody); manaString(got) != "{4}" {
		t.Fatalf("equip cost while not monarch = %q, want {4}", manaString(got))
	}

	g.Players[game.Player1].IsMonarch = true
	if got := effectiveActivatedAbilityCost(g, game.Player1, nil, &equipBody); manaString(got) != "{1}" {
		t.Fatalf("equip cost while monarch = %q, want {1}", manaString(got))
	}

	g.Players[game.Player1].IsMonarch = false
	g.Players[game.Player2].IsMonarch = true
	if got := effectiveActivatedAbilityCost(g, game.Player1, nil, &equipBody); manaString(got) != "{4}" {
		t.Fatalf("equip cost while opponent is monarch = %q, want {4}", manaString(got))
	}
}

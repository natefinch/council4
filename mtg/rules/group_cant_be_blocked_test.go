package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGroupCantBeBlockedThisTurnAppliesToControllersCreatures models Keeper of
// Keys: a one-shot "creatures you control can't be blocked this turn" applies the
// restriction to every creature the effect's controller controls (not a single
// permanent), leaves opponents' creatures blockable, and expires at cleanup. The
// group rule effect carries no object anchor; the runtime scopes it by the
// affected-controller relation and required card types.
func TestGroupCantBeBlockedThisTurnAppliesToControllersCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	mine2 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponent := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:               game.RuleEffectCantBeBlocked,
		Controller:         game.Player1,
		AffectedController: game.ControllerYou,
		PermanentTypes:     []types.Card{types.Creature},
		Duration:           game.DurationThisTurn,
		CreatedTurn:        g.Turn.TurnNumber,
	})

	if !ruleEffectProhibitsBeingBlocked(g, mine1) || !ruleEffectProhibitsBeingBlocked(g, mine2) {
		t.Fatal("group can't-be-blocked did not apply to both of the controller's creatures")
	}
	if ruleEffectProhibitsBeingBlocked(g, opponent) {
		t.Fatal("group can't-be-blocked wrongly applied to an opponent's creature")
	}

	expireRuleEffects(g)

	if ruleEffectProhibitsBeingBlocked(g, mine1) {
		t.Fatal("group can't-be-blocked still applied after cleanup expiry")
	}
}

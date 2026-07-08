package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// TestGenericDoesNotActivateManaAbilityStandalone checks that activating a mana
// ability on its own is not scored above passing. Mana is spent through the
// payment system as the agent pays for spells and abilities, so activating a mana
// ability standalone only floats mana that empties at end of step — and a
// mana-neutral one that pays for itself (Skyshroud Elf, "{1}: Add {R} or {W}")
// would otherwise be activated without end, spinning the priority loop.
func TestGenericDoesNotActivateManaAbilityStandalone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rock := addObservedPermanent(g, game.Player1, manaRockDef("Rock", 2))
	act := action.ActivateAbility(rock.ObjectID, 0, nil, 0)

	strategy := GenericStrategy{}
	if score := strategy.ScoreAction(rules.NewObservation(g, game.Player1), act); score > scorePass {
		t.Fatalf("standalone mana-ability activation scored %v, want at or below pass %v", score, scorePass)
	}
}

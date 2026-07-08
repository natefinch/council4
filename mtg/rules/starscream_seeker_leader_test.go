package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
)

// newStarscreamSeekerLeader puts the real Starscream card onto controller's
// battlefield as its back face (Starscream, Seeker Leader) so its "Whenever
// Starscream deals combat damage to a player, if there is no monarch, that
// player becomes the monarch." trigger runs through the real engine path.
func newStarscreamSeekerLeader(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.StarscreamPowerHungry)
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

// TestStarscreamCombatDamageBecomesMonarch proves Starscream, Seeker Leader's
// back-face ability on the real card: when Starscream deals combat damage to a
// player and there is no monarch, that damaged player becomes the monarch.
func TestStarscreamCombatDamageBecomesMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	starscream := newStarscreamSeekerLeader(g, game.Player1)

	dealPlayerDamage(g, starscream.ObjectID, starscream.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Starscream combat-damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("damaged opponent did not become the monarch after Starscream's combat damage")
	}
}

// TestStarscreamCombatDamageMonarchGate proves the NoMonarch intervening
// condition on the real card: while a monarch already exists, Starscream's
// combat damage to a non-monarch opponent leaves the crown untouched.
func TestStarscreamCombatDamageMonarchGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player3].IsMonarch = true
	starscream := newStarscreamSeekerLeader(g, game.Player1)

	dealPlayerDamage(g, starscream.ObjectID, starscream.ObjectID, game.Player1, game.Player2, 2, true)
	// The NoMonarch gate fails, so no become-monarch trigger goes on the stack.
	engine.putTriggeredAbilitiesOnStack(g)

	if g.Players[game.Player2].IsMonarch {
		t.Fatal("opponent became monarch despite an existing monarch")
	}
	if !g.Players[game.Player3].IsMonarch {
		t.Fatal("existing monarch lost the crown")
	}
}

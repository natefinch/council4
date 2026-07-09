package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addTapScopedAlly puts a vanilla creature the given player controls onto the
// battlefield, tapped or untapped, so Saryth's tap-state-scoped anthems can be
// observed against a permanent in a known tap state.
func addTapScopedAlly(g *game.Game, controller game.PlayerID, name string, tapped bool) *game.Permanent {
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	permanent.Tapped = tapped
	return permanent
}

// TestSarythTapStateScopedAnthemsAreSelective proves the real generated Saryth,
// the Viper's Fang applies its two tap-state-scoped affected groups selectively:
// "Other tapped creatures you control have deathtouch." grants deathtouch only to
// OTHER tapped creatures its controller controls, and "Other untapped creatures
// you control have hexproof." grants hexproof only to OTHER untapped creatures.
// A tapped ally gains deathtouch but not hexproof; an untapped ally gains
// hexproof but not deathtouch; Saryth itself (untapped, but excluded from its own
// "other" groups) gains neither; and an opponent's tapped creature is unaffected
// because the groups are scoped to the source's controller.
func TestSarythTapStateScopedAnthemsAreSelective(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	saryth := addCombatPermanent(g, game.Player1, cards.SarythTheViperSFang())

	tappedAlly := addTapScopedAlly(g, game.Player1, "Tapped Ally", true)
	untappedAlly := addTapScopedAlly(g, game.Player1, "Untapped Ally", false)
	opponentTapped := addTapScopedAlly(g, game.Player2, "Rival", true)

	obs := observe(g, game.Player1)

	tappedView := findPermanentView(t, obs, tappedAlly.ObjectID)
	if !tappedView.HasKeyword(game.Deathtouch) {
		t.Error("tapped ally should have deathtouch")
	}
	if tappedView.HasKeyword(game.Hexproof) {
		t.Error("tapped ally should NOT have hexproof (it is tapped, not untapped)")
	}

	untappedView := findPermanentView(t, obs, untappedAlly.ObjectID)
	if !untappedView.HasKeyword(game.Hexproof) {
		t.Error("untapped ally should have hexproof")
	}
	if untappedView.HasKeyword(game.Deathtouch) {
		t.Error("untapped ally should NOT have deathtouch (it is untapped, not tapped)")
	}

	sarythView := findPermanentView(t, obs, saryth.ObjectID)
	if sarythView.HasKeyword(game.Hexproof) || sarythView.HasKeyword(game.Deathtouch) {
		t.Errorf("Saryth is excluded from its own groups: deathtouch=%v hexproof=%v, want both false",
			sarythView.HasKeyword(game.Deathtouch), sarythView.HasKeyword(game.Hexproof))
	}

	rivalView := findPermanentView(t, obs, opponentTapped.ObjectID)
	if rivalView.HasKeyword(game.Deathtouch) || rivalView.HasKeyword(game.Hexproof) {
		t.Errorf("opponent's tapped creature must be unaffected: deathtouch=%v hexproof=%v",
			rivalView.HasKeyword(game.Deathtouch), rivalView.HasKeyword(game.Hexproof))
	}
}

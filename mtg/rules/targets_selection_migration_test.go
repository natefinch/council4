package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestPermanentTargetLegalityFlowsThroughSelection verifies that, after the
// TargetPredicate->Selection migration, permanent characteristic legality is
// driven solely by the spec's Selection. A spec carrying a Selection that
// requires a red creature you control must accept only permanents matching
// every characteristic and reject any that fail one. This is the canonical
// matchSelection path (CR 115.1, CR 608.2b).
func TestPermanentTargetLegalityFlowsThroughSelection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	redAlly := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	whiteAlly := addColoredCreaturePermanent(g, game.Player1, 2, color.White)
	redEnemy := addColoredCreaturePermanent(g, game.Player2, 2, color.Red)
	redArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Relic",
		Types:  []types.Card{types.Artifact},
		Colors: []color.Color{color.Red},
	}})

	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature},
			ColorsAny:        []color.Color{color.Red},
			Controller:       game.ControllerYou,
		}),
	}

	cases := []struct {
		name      string
		permanent *game.Permanent
		want      bool
	}{
		{"red creature you control", redAlly, true},
		{"wrong color creature you control", whiteAlly, false},
		{"red creature an opponent controls", redEnemy, false},
		{"red noncreature you control", redArtifact, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := targetMatchesSpec(g, game.Player1, 0, &spec, game.PermanentTarget(tc.permanent.ObjectID))
			if got != tc.want {
				t.Fatalf("targetMatchesSpec(%s) = %t, want %t", tc.name, got, tc.want)
			}
		})
	}
}

// TestSelectionTargetPreservesHexproofAndProtection verifies that routing
// permanent legality through Selection leaves the hexproof, shroud, and
// protection-from-color interactions unchanged: a red source choosing a
// "target creature" Selection may not target a creature with shroud, an
// opponent-controlled hexproof creature, or a creature with protection from
// red, but may target its controller's own hexproof creature and an ordinary
// creature (CR 702.11e, CR 702.16c, CR 702.18e).
func TestSelectionTargetPreservesHexproofAndProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	ordinary := addColoredCreaturePermanent(g, game.Player2, 2, color.Green)
	shroud := addShroudPermanent(g, game.Player2)
	enemyHexproof := addHexproofPermanent(g, game.Player2)
	ownHexproof := addHexproofPermanent(g, game.Player1)
	protectedFromRed := addProtectionFromColorPermanent(g, game.Player2, color.Red)

	redSource := &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Bolt",
		Types:  []types.Card{types.Instant},
		Colors: []color.Color{color.Red},
	}}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature},
		}),
	}

	cases := []struct {
		name      string
		permanent *game.Permanent
		want      bool
	}{
		{"ordinary creature", ordinary, true},
		{"own hexproof creature", ownHexproof, true},
		{"shroud creature", shroud, false},
		{"opponent hexproof creature", enemyHexproof, false},
		{"protection from red creature", protectedFromRed, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targets := []game.Target{game.PermanentTarget(tc.permanent.ObjectID)}
			got := targetsMatchSpecSlice(g, game.Player1, redSource, 0, &spec, targets)
			if got != tc.want {
				t.Fatalf("targetsMatchSpecSlice(%s) = %t, want %t", tc.name, got, tc.want)
			}
		})
	}
}

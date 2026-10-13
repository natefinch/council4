package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCompiledPermanentUnionExplicitTailCandidates(t *testing.T) {
	for _, tc := range []struct {
		tail   string
		flying bool
		pt     int
		want   bool
	}{
		{"with flying", false, 3, false},
		{"with flying", true, 3, true},
		{"without flying", true, 3, false},
		{"without flying", false, 3, true},
		{"with power 3 or less", false, 4, false},
		{"with power 3 or less", false, 3, true},
		{"with toughness 3 or less", false, 4, false},
		{"with toughness 3 or less", false, 3, true},
	} {
		for _, vehicle := range []bool{false, true} {
			t.Run(tc.tail+map[bool]string{false: "/creature", true: "/vehicle"}[vehicle], func(t *testing.T) {
				def := compiledPermanentUnion(t, "Tap target creature "+tc.tail+" or Vehicle "+tc.tail+".")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				face := game.CardFace{
					Name: "Tail Candidate", Types: []types.Card{types.Creature},
					Power: opt.Val(game.PT{Value: tc.pt}), Toughness: opt.Val(game.PT{Value: tc.pt}),
				}
				if vehicle {
					face.Types = []types.Card{types.Artifact, types.Creature}
					face.Subtypes = []types.Sub{types.Vehicle}
				}
				if tc.flying {
					face.StaticAbilities = []game.StaticAbility{game.FlyingStaticBody}
				}
				target := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: face})
				choices := targetChoicesForSpell(g, game.Player1, def, nil, game.CastBranch{})
				offered := slices.ContainsFunc(choices.choices, func(choice []game.Target) bool {
					return slices.Contains(choice, game.PermanentTarget(target.ObjectID))
				})
				if offered != tc.want {
					t.Errorf("vehicle=%v flying=%v PT=%d offered=%v want=%v", vehicle, tc.flying, tc.pt, offered, tc.want)
				}
			})
		}
	}
}

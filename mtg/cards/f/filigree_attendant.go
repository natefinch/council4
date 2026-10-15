package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FiligreeAttendant is the card definition for Filigree Attendant.
//
// Type: Artifact Creature — Homunculus
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flying
//	Filigree Attendant's power is equal to the number of artifacts you control.
var FiligreeAttendant = newFiligreeAttendant

func newFiligreeAttendant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Filigree Attendant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:       []color.Color{color.Blue},
			Types:        []types.Card{types.Artifact, types.Creature},
			Subtypes:     []types.Sub{types.Homunculus},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			Filigree Attendant's power is equal to the number of artifacts you control.
		`,
		},
	}
}

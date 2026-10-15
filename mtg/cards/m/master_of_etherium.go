package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MasterOfEtherium is the card definition for Master of Etherium.
//
// Type: Artifact Creature — Vedalken Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	Master of Etherium's power and toughness are each equal to the number of artifacts you control.
//	Other artifact creatures you control get +1/+1.
var MasterOfEtherium = newMasterOfEtherium

func newMasterOfEtherium() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Master of Etherium",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Vedalken, types.Wizard},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Other artifact creatures you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Master of Etherium's power and toughness are each equal to the number of artifacts you control.
			Other artifact creatures you control get +1/+1.
		`,
		},
	}
}

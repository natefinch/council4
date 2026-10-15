package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkorpekhLord is the card definition for Skorpekh Lord.
//
// Type: Artifact Creature — Necron Noble
// Cost: {2}{B}
//
// Oracle text:
//
//	Menace
//	Command Protocols — Other artifact creatures you control get +1/+0 and have menace.
//	Unearth {2}{B} ({2}{B}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
var SkorpekhLord = newSkorpekhLord

func newSkorpekhLord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Skorpekh Lord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Necron, types.Noble},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Text: "Command Protocols — Other artifact creatures you control get +1/+0 and have menace.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.UnearthActivatedAbility(cost.Mana{cost.O(2), cost.B}),
			},
			OracleText: `
			Menace
			Command Protocols — Other artifact creatures you control get +1/+0 and have menace.
			Unearth {2}{B} ({2}{B}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
		`,
		},
	}
}

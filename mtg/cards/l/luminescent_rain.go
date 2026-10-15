package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LuminescentRain is the card definition for Luminescent Rain.
//
// Type: Instant
// Cost: {2}{G}
//
// Oracle text:
//
//	Choose a creature type. You gain 2 life for each permanent you control of that type.
var LuminescentRain = newLuminescentRain

func newLuminescentRain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Luminescent Rain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:          game.ResolutionChoiceSubtype,
								Prompt:        "Choose a creature type",
								SubtypeOfType: types.Creature,
							},
							PublishChoice: game.SpellChosenTypeChoiceKey,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 2,
								Group:      game.BattlefieldGroup(game.Selection{SubtypeChoice: game.SubtypeChoiceResolution, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a creature type. You gain 2 life for each permanent you control of that type.
		`,
		},
	}
}

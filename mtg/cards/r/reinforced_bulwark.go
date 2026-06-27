package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReinforcedBulwark is the card definition for Reinforced Bulwark.
//
// Type: Artifact Creature — Wall
// Cost: {3}
//
// Oracle text:
//
//	Defender
//	{T}: Prevent the next 1 damage that would be dealt to you this turn.
var ReinforcedBulwark = newReinforcedBulwark()

func newReinforcedBulwark() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Reinforced Bulwark",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Prevent the next 1 damage that would be dealt to you this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player: game.ControllerReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			{T}: Prevent the next 1 damage that would be dealt to you this turn.
		`,
		},
	}
}

package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkiriLineSlinger is the card definition for Akiri, Line-Slinger.
//
// Type: Legendary Creature — Kor Soldier Ally
// Cost: {R}{W}
//
// Oracle text:
//
//	First strike, vigilance
//	Akiri gets +1/+0 for each artifact you control.
//	Partner (You can have two commanders if both have partner.)
var AkiriLineSlinger = newAkiriLineSlinger

func newAkiriLineSlinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Akiri, Line-Slinger",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Soldier, types.Ally},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Akiri gets +1/+0 for each artifact you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountSelector,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
							}),
						},
					},
				},
				game.PartnerStaticBody,
			},
			OracleText: `
			First strike, vigilance
			Akiri gets +1/+0 for each artifact you control.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}

package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FaeburrowElder is the card definition for Faeburrow Elder.
//
// Type: Creature — Treefolk Druid
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Vigilance
//	This creature gets +1/+1 for each color among permanents you control.
//	{T}: For each color among permanents you control, add one mana of that color.
var FaeburrowElder = newFaeburrowElder

func newFaeburrowElder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Faeburrow Elder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Druid},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "This creature gets +1/+1 for each color among permanents you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountColorCountInGroup,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountColorCountInGroup,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Group:       game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
							}),
						},
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaEachControlledColorAbility("{T}: For each color among permanents you control, add one mana of that color.", game.Selection{Controller: game.ControllerYou}),
			},
			OracleText: `
			Vigilance
			This creature gets +1/+1 for each color among permanents you control.
			{T}: For each color among permanents you control, add one mana of that color.
		`,
		},
	}
}

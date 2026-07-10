package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KolaghanForerunners is the card definition for Kolaghan Forerunners.
//
// Type: Creature — Human Berserker
// Cost: {2}{R}
//
// Oracle text:
//
//	Trample
//	Kolaghan Forerunners's power is equal to the number of creatures you control.
//	Dash {2}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var KolaghanForerunners = newKolaghanForerunners

func newKolaghanForerunners() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kolaghan Forerunners",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:       []color.Color{color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Berserker},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Trample
			Kolaghan Forerunners's power is equal to the number of creatures you control.
			Dash {2}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}

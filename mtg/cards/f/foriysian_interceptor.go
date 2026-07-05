package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForiysianInterceptor is the card definition for Foriysian Interceptor.
//
// Type: Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Defender
//	This creature can block an additional creature each combat.
var ForiysianInterceptor = newForiysianInterceptor()

func newForiysianInterceptor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Foriysian Interceptor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.DefenderStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Defender
			This creature can block an additional creature each combat.
		`,
		},
	}
}

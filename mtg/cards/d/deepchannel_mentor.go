package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeepchannelMentor is the card definition for Deepchannel Mentor.
//
// Type: Creature — Merfolk Rogue
// Cost: {5}{U}
//
// Oracle text:
//
//	Blue creatures you control can't be blocked.
var DeepchannelMentor = newDeepchannelMentor

func newDeepchannelMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Deepchannel Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Blue creatures you control can't be blocked.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeBlocked,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
							AffectedSelection:  game.Selection{ColorsAny: []color.Color{color.Blue}},
						},
					},
				},
			},
			OracleText: `
			Blue creatures you control can't be blocked.
		`,
		},
	}
}

package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BorneUponAWind is the card definition for Borne Upon a Wind.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	You may cast spells this turn as though they had flash.
//	Draw a card.
var BorneUponAWind = newBorneUponAWind

func newBorneUponAWind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Borne Upon a Wind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectCastSpellsAsThoughFlash,
									Duration:       game.DurationPermanent,
									AffectedPlayer: game.PlayerYou,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may cast spells this turn as though they had flash.
			Draw a card.
		`,
		},
	}
}

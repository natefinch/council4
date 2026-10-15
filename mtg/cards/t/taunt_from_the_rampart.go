package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TauntFromTheRampart is the card definition for Taunt from the Rampart.
//
// Type: Sorcery
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Goad all creatures your opponents control. Until your next turn, those creatures can't block. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)
var TauntFromTheRampart = newTauntFromTheRampart

func newTauntFromTheRampart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Taunt from the Rampart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Goad{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:               game.RuleEffectCantBlock,
									Duration:           game.DurationPermanent,
									AffectedController: game.ControllerOpponent,
									PermanentTypes:     []types.Card{types.Creature},
									AffectedSelection:  game.Selection{MatchGoaded: true},
								},
							},
							Duration: game.DurationUntilYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Goad all creatures your opponents control. Until your next turn, those creatures can't block. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)
		`,
		},
	}
}

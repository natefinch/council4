package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QueenMotherRamonda is the card definition for Queen Mother Ramonda.
//
// Type: Legendary Creature — Human Noble
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	When Queen Mother Ramonda enters, you become the monarch.
//	As long as you're the monarch, creatures with power 2 or less can't attack you.
var QueenMotherRamonda = newQueenMotherRamonda

func newQueenMotherRamonda() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Queen Mother Ramonda",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerIsMonarch: true,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                      game.RuleEffectCantAttack,
							DefendingPlayer:           game.PlayerYou,
							DefendingPlayerDirectOnly: true,
							PermanentTypes:            []types.Card{types.Creature},
							AffectedSelection:         game.Selection{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Queen Mother Ramonda enters, you become the monarch.
			As long as you're the monarch, creatures with power 2 or less can't attack you.
		`,
		},
	}
}

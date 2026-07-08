package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShaleskinBruiser is the card definition for Shaleskin Bruiser.
//
// Type: Creature — Beast
// Cost: {6}{R}
//
// Oracle text:
//
//	Trample
//	Whenever this creature attacks, it gets +3/+0 until end of turn for each other attacking Beast.
var ShaleskinBruiser = newShaleskinBruiser

func newShaleskinBruiser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Shaleskin Bruiser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object: game.EventPermanentReference(),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 3,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Beast")}, CombatState: game.CombatStateAttacking, ExcludeSource: true}),
									}),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever this creature attacks, it gets +3/+0 until end of turn for each other attacking Beast.
		`,
		},
	}
}

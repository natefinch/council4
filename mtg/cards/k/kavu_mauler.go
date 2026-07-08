package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KavuMauler is the card definition for Kavu Mauler.
//
// Type: Creature — Kavu
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Trample
//	Whenever this creature attacks, it gets +1/+1 until end of turn for each other attacking Kavu.
var KavuMauler = newKavuMauler

func newKavuMauler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kavu Mauler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu},
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
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Kavu")}, CombatState: game.CombatStateAttacking, ExcludeSource: true}),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Kavu")}, CombatState: game.CombatStateAttacking, ExcludeSource: true}),
									}),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever this creature attacks, it gets +1/+1 until end of turn for each other attacking Kavu.
		`,
		},
	}
}

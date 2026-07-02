package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RubblebeltRaiders is the card definition for Rubblebelt Raiders.
//
// Type: Creature — Human Warrior
// Cost: {1}{R/G}{R/G}{R/G}
//
// Oracle text:
//
//	Whenever this creature attacks, put a +1/+1 counter on it for each attacking creature you control.
var RubblebeltRaiders = newRubblebeltRaiders()

func newRubblebeltRaiders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Rubblebelt Raiders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
									}),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, put a +1/+1 counter on it for each attacking creature you control.
		`,
		},
	}
}

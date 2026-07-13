package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteelburrChampion is the card definition for SteelburrChampion.
//
// Type: Creature — Mouse Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	Offspring {1}{W} (You may pay an additional {1}{W} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Vigilance
//	Whenever an opponent casts a noncreature spell, put a +1/+1 counter on this creature.
var SteelburrChampion = newSteelburrChampion

func newSteelburrChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Steelburr Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mouse, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(1), cost.W}),
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerOpponent,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {1}{W} (You may pay an additional {1}{W} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Vigilance
			Whenever an opponent casts a noncreature spell, put a +1/+1 counter on this creature.
		`,
		},
	}
}

package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GimliCounterOfKills is the card definition for Gimli, Counter of Kills.
//
// Type: Legendary Creature — Dwarf Warrior
// Cost: {3}{R}
//
// Oracle text:
//
//	Trample
//	Whenever a creature an opponent controls dies, Gimli deals 1 damage to that creature's controller.
var GimliCounterOfKills = newGimliCounterOfKills()

func newGimliCounterOfKills() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gimli, Counter of Kills",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dwarf, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerOpponent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.EventPermanentReference())),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever a creature an opponent controls dies, Gimli deals 1 damage to that creature's controller.
		`,
		},
	}
}

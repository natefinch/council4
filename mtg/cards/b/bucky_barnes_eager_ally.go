package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BuckyBarnesEagerAlly is the card definition for Bucky Barnes, Eager Ally.
//
// Type: Legendary Creature — Human Soldier Hero
// Cost: {1}{W}
//
// Oracle text:
//
//	Vigilance
//	When Bucky Barnes dies, look at the top four cards of your library. You may reveal an Equipment, Hero, or Soldier card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var BuckyBarnesEagerAlly = newBuckyBarnesEagerAlly

func newBuckyBarnesEagerAlly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Bucky Barnes, Eager Ally",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(4),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Equipment"), types.Sub("Hero"), types.Sub("Soldier")}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			When Bucky Barnes dies, look at the top four cards of your library. You may reveal an Equipment, Hero, or Soldier card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}

package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UreniOfTheUnwritten is the card definition for Ureni of the Unwritten.
//
// Type: Legendary Creature — Spirit Dragon
// Cost: {4}{G}{U}{R}
//
// Oracle text:
//
//	Flying, trample
//	Whenever Ureni enters or attacks, look at the top eight cards of your library. You may put a Dragon creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
var UreniOfTheUnwritten = newUreniOfTheUnwritten()

func newUreniOfTheUnwritten() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ureni of the Unwritten",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit, types.Dragon},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerDeclared,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:      game.ControllerReference(),
									Look:        game.Fixed(8),
									Take:        game.Fixed(1),
									Remainder:   game.DigRemainderLibraryBottom,
									Filter:      opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Dragon")}}),
									TakeUpTo:    true,
									Destination: zone.Battlefield,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, trample
			Whenever Ureni enters or attacks, look at the top eight cards of your library. You may put a Dragon creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}

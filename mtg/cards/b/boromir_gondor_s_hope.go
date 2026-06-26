package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoromirGondorSHope is the card definition for Boromir, Gondor's Hope.
//
// Type: Legendary Creature — Human Warrior
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Whenever Boromir enters or attacks, look at the top six cards of your library. You may reveal a Human or artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var BoromirGondorSHope = newBoromirGondorSHope()

func newBoromirGondorSHope() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Boromir, Gondor's Hope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
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
									Player:    game.ControllerReference(),
									Look:      game.Fixed(6),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, SubtypesAny: []types.Sub{types.Sub("Human")}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever Boromir enters or attacks, look at the top six cards of your library. You may reveal a Human or artifact card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}

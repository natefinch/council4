package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SandsteppeMastodon is the card definition for Sandsteppe Mastodon.
//
// Type: Creature — Elephant
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Reach
//	When this creature enters, bolster 5. (Choose a creature with the least toughness among creatures you control and put five +1/+1 counters on it.)
var SandsteppeMastodon = newSandsteppeMastodon

func newSandsteppeMastodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sandsteppe Mastodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
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
								Primitive: game.Bolster{
									Amount: game.Fixed(5),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			When this creature enters, bolster 5. (Choose a creature with the least toughness among creatures you control and put five +1/+1 counters on it.)
		`,
		},
	}
}

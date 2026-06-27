package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FierceEmpath is the card definition for Fierce Empath.
//
// Type: Creature — Elf
// Cost: {2}{G}
//
// Oracle text:
//
//	When this creature enters, you may search your library for a creature card with mana value 6 or greater, reveal it, put it into your hand, then shuffle.
var FierceEmpath = newFierceEmpath()

func newFierceEmpath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fierce Empath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
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
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Hand,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 6})},
										Reveal:      true,
									},
									Amount: game.Fixed(1),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you may search your library for a creature card with mana value 6 or greater, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}

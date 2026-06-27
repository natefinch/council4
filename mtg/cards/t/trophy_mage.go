package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrophyMage is the card definition for Trophy Mage.
//
// Type: Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	When this creature enters, you may search your library for an artifact card with mana value 3, reveal it, put it into your hand, then shuffle.
var TrophyMage = newTrophyMage()

func newTrophyMage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Trophy Mage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Artifact}, ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 3})},
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
			When this creature enters, you may search your library for an artifact card with mana value 3, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}

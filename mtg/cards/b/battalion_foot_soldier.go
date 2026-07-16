package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BattalionFootSoldier is the card definition for Battalion Foot Soldier.
//
// Type: Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	When this creature enters, you may search your library for any number of cards named Battalion Foot Soldier, reveal them, put them into your hand, then shuffle.
var BattalionFootSoldier = newBattalionFootSoldier

func newBattalionFootSoldier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Battalion Foot Soldier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
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
										Name:        "Battalion Foot Soldier",
										Reveal:      true,
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, you may search your library for any number of cards named Battalion Foot Soldier, reveal them, put them into your hand, then shuffle.
		`,
		},
	}
}

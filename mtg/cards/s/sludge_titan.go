package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SludgeTitan is the card definition for Sludge Titan.
//
// Type: Creature — Zombie Giant
// Cost: {4}{B/G}{B/G}
//
// Oracle text:
//
//	Trample
//	Whenever this creature enters or attacks, mill five cards. You may put a creature card and/or a land card from among them into your hand.
var SludgeTitan = newSludgeTitan()

func newSludgeTitan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Sludge Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Giant},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
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
								Primitive: game.Mill{
									Amount:        game.Fixed(5),
									Player:        game.ControllerReference(),
									PublishLinked: game.LinkedKey("milled-cards"),
								},
							},
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Graveyard,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Hand,
									},
									Riders: game.ChooseRiders{
										FromLinked: game.LinkedKey("milled-cards"),
									},
									Prompt: "Choose a card to return to your hand",
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever this creature enters or attacks, mill five cards. You may put a creature card and/or a land card from among them into your hand.
		`,
		},
	}
}

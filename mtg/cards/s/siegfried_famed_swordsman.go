package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SiegfriedFamedSwordsman is the card definition for Siegfried, Famed Swordsman.
//
// Type: Legendary Creature — Human Warrior Rogue
// Cost: {3}{B}
//
// Oracle text:
//
//	Menace
//	When Siegfried enters, mill three cards. Then put X +1/+1 counters on Siegfried, where X is twice the number of creature cards in your graveyard.
var SiegfriedFamedSwordsman = newSiegfriedFamedSwordsman

func newSiegfriedFamedSwordsman() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Siegfried, Famed Swordsman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
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
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 2,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Graveyard,
										Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace
			When Siegfried enters, mill three cards. Then put X +1/+1 counters on Siegfried, where X is twice the number of creature cards in your graveyard.
		`,
		},
	}
}

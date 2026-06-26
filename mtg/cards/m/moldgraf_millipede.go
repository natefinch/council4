package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoldgrafMillipede is the card definition for Moldgraf Millipede.
//
// Type: Creature — Insect Horror
// Cost: {4}{G}
//
// Oracle text:
//
//	When this creature enters, mill three cards, then put a +1/+1 counter on this creature for each creature card in your graveyard. (To mill three cards, put the top three cards of your library into your graveyard.)
var MoldgrafMillipede = newMoldgrafMillipede()

func newMoldgrafMillipede() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Moldgraf Millipede",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect, types.Horror},
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
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
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
			When this creature enters, mill three cards, then put a +1/+1 counter on this creature for each creature card in your graveyard. (To mill three cards, put the top three cards of your library into your graveyard.)
		`,
		},
	}
}

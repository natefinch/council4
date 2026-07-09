package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GhaltaStampedeTyrant is the card definition for Ghalta, Stampede Tyrant.
//
// Type: Legendary Creature — Elder Dinosaur
// Cost: {5}{G}{G}{G}
//
// Oracle text:
//
//	Trample
//	When Ghalta enters, put any number of creature cards from your hand onto the battlefield.
var GhaltaStampedeTyrant = newGhaltaStampedeTyrant

func newGhaltaStampedeTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ghalta, Stampede Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elder, types.Dinosaur},
			Power:      opt.Val(game.PT{Value: 12}),
			Toughness:  opt.Val(game.PT{Value: 12}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
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
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
									Count:      game.ChooseAnyNumber,
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Prompt: "Choose a card to put onto the battlefield",
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			When Ghalta enters, put any number of creature cards from your hand onto the battlefield.
		`,
		},
	}
}

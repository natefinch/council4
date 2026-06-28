package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GraveDefiler is the card definition for Grave Defiler.
//
// Type: Creature — Zombie
// Cost: {3}{B}
//
// Oracle text:
//
//	When this creature enters, reveal the top four cards of your library. Put all Zombie cards revealed this way into your hand and the rest on the bottom of your library in any order.
//	{1}{B}: Regenerate this creature.
var GraveDefiler = newGraveDefiler()

func newGraveDefiler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grave Defiler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{B}: Regenerate this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
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
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(4),
									Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}},
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, reveal the top four cards of your library. Put all Zombie cards revealed this way into your hand and the rest on the bottom of your library in any order.
			{1}{B}: Regenerate this creature.
		`,
		},
	}
}

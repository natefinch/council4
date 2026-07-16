package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// InameDeathAspect is the card definition for Iname, Death Aspect.
//
// Type: Legendary Creature — Spirit
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	When Iname enters, you may search your library for any number of Spirit cards, put them into your graveyard, then shuffle.
var InameDeathAspect = newInameDeathAspect

func newInameDeathAspect() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Iname, Death Aspect",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
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
										Destination: zone.Graveyard,
										Filter:      game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit")}},
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
			When Iname enters, you may search your library for any number of Spirit cards, put them into your graveyard, then shuffle.
		`,
		},
	}
}

package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GodoBanditWarlord is the card definition for Godo, Bandit Warlord.
//
// Type: Legendary Creature — Human Barbarian
// Cost: {5}{R}
//
// Oracle text:
//
//	When Godo enters, you may search your library for an Equipment card, put it onto the battlefield, then shuffle.
//	Whenever Godo attacks for the first time each turn, untap it and all Samurai you control. After this phase, there is an additional combat phase.
var GodoBanditWarlord = newGodoBanditWarlord

func newGodoBanditWarlord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Godo, Bandit Warlord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Barbarian},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
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
										Destination: zone.Battlefield,
										Filter:      game.Selection{SubtypesAny: []types.Sub{types.Sub("Equipment")}},
									},
									Amount: game.Fixed(1),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.EventPermanentReference(),
								},
							},
							{
								Primitive: game.AddExtraPhases{
									Combat: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Godo enters, you may search your library for an Equipment card, put it onto the battlefield, then shuffle.
			Whenever Godo attacks for the first time each turn, untap it and all Samurai you control. After this phase, there is an additional combat phase.
		`,
		},
	}
}

package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArdozCobblerOfWar is the card definition for Ardoz, Cobbler of War.
//
// Type: Legendary Creature — Goblin Shaman
// Cost: {1}{R}
//
// Oracle text:
//
//	Haste
//	Whenever Ardoz or another creature you control enters, that creature gets +2/+0 until end of turn.
//	{3}{R}: Create a 1/1 red Goblin creature token with haste. Activate only as a sorcery.
var ArdozCobblerOfWar = newArdozCobblerOfWar

func newArdozCobblerOfWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ardoz, Cobbler of War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Goblin, types.Shaman},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{R}: Create a 1/1 red Goblin creature token with haste. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(ardozCobblerOfWarToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Ardoz or another creature you control enters, that creature gets +2/+0 until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
							SubjectSelectionOrSelf: true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.EventPermanentReference(),
											PowerDelta: game.Fixed(2),
											Duration:   game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Haste
			Whenever Ardoz or another creature you control enters, that creature gets +2/+0 until end of turn.
			{3}{R}: Create a 1/1 red Goblin creature token with haste. Activate only as a sorcery.
		`,
		},
	}
}

var ardozCobblerOfWarToken = newArdozCobblerOfWarToken()

func newArdozCobblerOfWarToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}

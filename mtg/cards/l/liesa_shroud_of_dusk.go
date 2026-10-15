package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LiesaShroudOfDusk is the card definition for Liesa, Shroud of Dusk.
//
// Type: Legendary Creature — Angel
// Cost: {2}{W}{W}{B}
//
// Oracle text:
//
//	Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times.
//	Flying, lifelink
//	Whenever a player casts a spell, they lose 2 life.
var LiesaShroudOfDusk = newLiesaShroudOfDusk

func newLiesaShroudOfDusk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Liesa, Shroud of Dusk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:           "Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times.",
					ZoneOfFunction: zone.Command,
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectPayLifeForCommanderTax,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
						},
					},
				},
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell, they lose 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(2),
											Player: game.EventPlayerReference(),
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
			Rather than pay {2} for each previous time you've cast this spell from the command zone this game, pay 2 life that many times.
			Flying, lifelink
			Whenever a player casts a spell, they lose 2 life.
		`,
		},
	}
}

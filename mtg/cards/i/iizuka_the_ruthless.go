package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IizukaTheRuthless is the card definition for Iizuka the Ruthless.
//
// Type: Legendary Creature — Human Samurai
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)
//	{2}{R}, Sacrifice a Samurai: Samurai creatures you control gain double strike until end of turn.
var IizukaTheRuthless = newIizukaTheRuthless

func newIizukaTheRuthless() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Iizuka the Ruthless",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Samurai},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{R}, Sacrifice a Samurai: Samurai creatures you control gain double strike until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Samurai",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Samurai},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Samurai")}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.DoubleStrike,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventBlockerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerBecameBlocked,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.EventPermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)
			{2}{R}, Sacrifice a Samurai: Samurai creatures you control gain double strike until end of turn.
		`,
		},
	}
}

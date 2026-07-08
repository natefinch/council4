package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FleshCarver is the card definition for Flesh Carver.
//
// Type: Creature — Human Wizard
// Cost: {2}{B}
//
// Oracle text:
//
//	Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this creature.
//	When this creature dies, create an X/X black Horror creature token, where X is this creature's power.
var FleshCarver = newFleshCarver

func newFleshCarver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Flesh Carver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IntimidateStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this creature.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
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
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(fleshCarverToken),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: 1,
										Object:     game.SourcePermanentReference(),
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: 1,
										Object:     game.SourcePermanentReference(),
									})),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
			{1}{B}, Sacrifice another creature: Put two +1/+1 counters on this creature.
			When this creature dies, create an X/X black Horror creature token, where X is this creature's power.
		`,
		},
	}
}

var fleshCarverToken = newFleshCarverToken()

func newFleshCarverToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Horror",
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Horror},
		},
	}
}

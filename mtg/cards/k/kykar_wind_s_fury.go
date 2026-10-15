package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KykarWindSFury is the card definition for Kykar, Wind's Fury.
//
// Type: Legendary Creature — Bird Wizard
// Cost: {1}{U}{R}{W}
//
// Oracle text:
//
//	Flying
//	Whenever you cast a noncreature spell, create a 1/1 white Spirit creature token with flying.
//	Sacrifice a Spirit: Add {R}.
var KykarWindSFury = newKykarWindSFury

func newKykarWindSFury() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Kykar, Wind's Fury",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bird, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Sacrifice a Spirit: Add {R}.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Spirit",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Spirit},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.R,
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
					Text: "Whenever you cast a noncreature spell, create a 1/1 white Spirit creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(kykarWindSFuryToken),
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
			Flying
			Whenever you cast a noncreature spell, create a 1/1 white Spirit creature token with flying.
			Sacrifice a Spirit: Add {R}.
		`,
		},
	}
}

var kykarWindSFuryToken = newKykarWindSFuryToken()

func newKykarWindSFuryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}

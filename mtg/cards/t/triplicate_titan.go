package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TriplicateTitan is the card definition for Triplicate Titan.
//
// Type: Artifact Creature — Golem
// Cost: {9}
//
// Oracle text:
//
//	Flying, vigilance, trample
//	When this creature dies, create a 3/3 colorless Golem artifact creature token with flying, a 3/3 colorless Golem artifact creature token with vigilance, and a 3/3 colorless Golem artifact creature token with trample.
var TriplicateTitan = newTriplicateTitan()

func newTriplicateTitan() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Triplicate Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
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
									Source: game.TokenDef(triplicateTitanToken),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(triplicateTitanToken2),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(triplicateTitanToken3),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance, trample
			When this creature dies, create a 3/3 colorless Golem artifact creature token with flying, a 3/3 colorless Golem artifact creature token with vigilance, and a 3/3 colorless Golem artifact creature token with trample.
		`,
		},
	}
}

var triplicateTitanToken = newTriplicateTitanToken()

func newTriplicateTitanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Golem",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}

var triplicateTitanToken2 = newTriplicateTitanToken2()

func newTriplicateTitanToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Golem",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}

var triplicateTitanToken3 = newTriplicateTitanToken3()

func newTriplicateTitanToken3() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Golem",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}

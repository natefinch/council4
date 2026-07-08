package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WurmcoilEngine is the card definition for Wurmcoil Engine.
//
// Type: Artifact Creature — Phyrexian Wurm
// Cost: {6}
//
// Oracle text:
//
//	Deathtouch, lifelink
//	When this creature dies, create a 3/3 colorless Phyrexian Wurm artifact creature token with deathtouch and a 3/3 colorless Phyrexian Wurm artifact creature token with lifelink.
var WurmcoilEngine = newWurmcoilEngine

func newWurmcoilEngine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wurmcoil Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wurm},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.LifelinkStaticBody,
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
									Source: game.TokenDef(wurmcoilEngineToken),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(wurmcoilEngineToken2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Deathtouch, lifelink
			When this creature dies, create a 3/3 colorless Phyrexian Wurm artifact creature token with deathtouch and a 3/3 colorless Phyrexian Wurm artifact creature token with lifelink.
		`,
		},
	}
}

var wurmcoilEngineToken = newWurmcoilEngineToken()

func newWurmcoilEngineToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Wurm",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wurm},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
		},
	}
}

var wurmcoilEngineToken2 = newWurmcoilEngineToken2()

func newWurmcoilEngineToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Wurm",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Wurm},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
		},
	}
}

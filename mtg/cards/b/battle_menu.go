package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BattleMenu is the card definition for Battle Menu.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Choose one —
//	• Attack — Create a 2/2 white Knight creature token.
//	• Ability — Target creature gets +0/+4 until end of turn.
//	• Magic — Destroy target creature with power 4 or greater.
//	• Item — You gain 4 life.
var BattleMenu = newBattleMenu

func newBattleMenu() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Battle Menu",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Attack — Create a 2/2 white Knight creature token.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(battleMenuToken),
								},
							},
						},
					},
					game.Mode{
						Text: "Ability — Target creature gets +0/+4 until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									ToughnessDelta: game.Fixed(4),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Magic — Destroy target creature with power 4 or greater.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with power 4 or greater",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Item — You gain 4 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(4),
									Player: game.ControllerReference(),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Attack — Create a 2/2 white Knight creature token.
			• Ability — Target creature gets +0/+4 until end of turn.
			• Magic — Destroy target creature with power 4 or greater.
			• Item — You gain 4 life.
		`,
		},
	}
}

var battleMenuToken = newBattleMenuToken()

func newBattleMenuToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}

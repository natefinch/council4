package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkromaSWill is the card definition for Akroma's Will.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Choose one. If you control a commander as you cast this spell, you may choose both instead.
//	• Creatures you control gain flying, vigilance, and double strike until end of turn.
//	• Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.
var AkromaSWill = newAkromaSWill

func newAkromaSWill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Akroma's Will",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Creatures you control gain flying, vigilance, and double strike until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Flying, game.Vigilance, game.DoubleStrike},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Lifelink, game.Indestructible},
											AddAbilities: []game.Ability{
												new(game.ProtectionFromEachColorStaticAbility()),
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1},
			}),
			OracleText: `
			Choose one. If you control a commander as you cast this spell, you may choose both instead.
			• Creatures you control gain flying, vigilance, and double strike until end of turn.
			• Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.
		`,
		},
	}
}

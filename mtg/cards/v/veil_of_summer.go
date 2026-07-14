package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VeilOfSummer is the card definition for Veil of Summer.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Draw a card if an opponent has cast a blue or black spell this turn. Spells you control can't be countered this turn. You and permanents you control gain hexproof from blue and from black until end of turn. (You and they can't be the targets of blue or black spells or abilities your opponents control.)
var VeilOfSummer = newVeilOfSummer

func newVeilOfSummer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Veil of Summer",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
									Event:         game.EventSpellCast,
									Controller:    game.TriggerControllerOpponent,
									CardSelection: game.Selection{ColorsAny: []color.Color{color.Blue, color.Black}},
								}, Window: game.EventHistoryCurrentTurn}),
							}),
						}),
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:               game.RuleEffectCantBeCountered,
									AffectedController: game.ControllerYou,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectPlayerHexproof,
									AffectedPlayer: game.PlayerYou,
									Protection:     game.ProtectionKeyword{FromColors: []color.Color{color.Blue, color.Black}},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									AddAbilities: []game.Ability{
										new(game.HexproofFromColorsStaticAbility(color.Blue, color.Black)),
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card if an opponent has cast a blue or black spell this turn. Spells you control can't be countered this turn. You and permanents you control gain hexproof from blue and from black until end of turn. (You and they can't be the targets of blue or black spells or abilities your opponents control.)
		`,
		},
	}
}

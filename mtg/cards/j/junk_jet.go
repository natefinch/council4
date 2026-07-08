package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JunkJet is the card definition for Junk Jet.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	When this Equipment enters, create a Junk token. (It's an artifact with "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.")
//	{3}, Sacrifice another artifact: Double equipped creature's power until end of turn.
//	Equip {1}
var JunkJet = newJunkJet

func newJunkJet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Junk Jet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}, Sacrifice another artifact: Double equipped creature's power until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceAttachedPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerPowerToughnessModify,
											DoublePower: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(junkJetToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this Equipment enters, create a Junk token. (It's an artifact with "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.")
			{3}, Sacrifice another artifact: Double equipped creature's power until end of turn.
			Equip {1}
		`,
		},
	}
}

var junkJetToken = newJunkJetToken()

func newJunkJetToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Junk",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Junk},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(1),
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

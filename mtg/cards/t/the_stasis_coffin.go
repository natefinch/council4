package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheStasisCoffin is the card definition for The Stasis Coffin.
//
// Type: Legendary Artifact
// Cost: {3}
//
// Oracle text:
//
//	{2}, {T}, Exile The Stasis Coffin: You gain protection from everything until your next turn.
var TheStasisCoffin = newTheStasisCoffin

func newTheStasisCoffin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Stasis Coffin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Exile The Stasis Coffin: You gain protection from everything until your next turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile The Stasis Coffin",
							Amount:      1,
							Source:      zone.Battlefield,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:           game.RuleEffectPlayerProtection,
													Duration:       game.DurationPermanent,
													AffectedPlayer: game.PlayerYou,
													Protection: game.ProtectionKeyword{
														Everything: true,
													},
												},
											},
											Duration: game.DurationUntilYourNextTurn,
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
			{2}, {T}, Exile The Stasis Coffin: You gain protection from everything until your next turn.
		`,
		},
	}
}

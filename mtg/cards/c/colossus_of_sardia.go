package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ColossusOfSardia is the card definition for Colossus of Sardia.
//
// Type: Artifact Creature — Golem
// Cost: {9}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	This creature doesn't untap during your untap step.
//	{9}: Untap this creature. Activate only during your upkeep.
var ColossusOfSardia = newColossusOfSardia

func newColossusOfSardia() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Colossus of Sardia",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "This creature doesn't untap during your untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{9}: Untap this creature. Activate only during your upkeep.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(9),
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.DuringUpkeep,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourcePermanentReference(),
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
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			This creature doesn't untap during your untap step.
			{9}: Untap this creature. Activate only during your upkeep.
		`,
		},
	}
}

package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CogworkAssembler is the card definition for Cogwork Assembler.
//
// Type: Artifact Creature — Assembly-Worker
// Cost: {3}
//
// Oracle text:
//
//	{7}: Create a token that's a copy of target artifact. That token gains haste. Exile it at the beginning of the next end step.
var CogworkAssembler = newCogworkAssembler

func newCogworkAssembler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Cogwork Assembler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.AssemblyWorker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{7}: Create a token that's a copy of target artifact. That token gains haste. Exile it at the beginning of the next end step.",
					ManaCost:       opt.Val(cost.Mana{cost.O(7)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:      game.TokenCopySourceObject,
										Object:      game.TargetPermanentReference(0),
										AddKeywords: []game.Keyword{game.Haste},
									}),
								},
							},
							{
								Primitive: game.Exile{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{7}: Create a token that's a copy of target artifact. That token gains haste. Exile it at the beginning of the next end step.
		`,
		},
	}
}

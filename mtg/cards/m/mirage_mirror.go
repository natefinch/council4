package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MirageMirror is the card definition for Mirage Mirror.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{2}: This artifact becomes a copy of target artifact, creature, enchantment, or land until end of turn.
var MirageMirror = newMirageMirror

func newMirageMirror() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mirage Mirror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: This artifact becomes a copy of target artifact, creature, enchantment, or land until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact, creature, enchantment, or land",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.BecomeCopy{
											Object:         game.TargetPermanentReference(0),
											UntilEndOfTurn: true,
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
			{2}: This artifact becomes a copy of target artifact, creature, enchantment, or land until end of turn.
		`,
		},
	}
}

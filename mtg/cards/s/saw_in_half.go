package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SawInHalf is the card definition for Saw in Half.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	Destroy target creature. If that creature dies this way, its controller creates two tokens that are copies of that creature, except their power is half that creature's power and their toughness is half that creature's toughness. Round up each time.
var SawInHalf = newSawInHalf

func newSawInHalf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Saw in Half",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
						PublishResult: game.ResultKey("dies-this-way-copy"),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:                     game.TokenCopySourceObject,
								Object:                     game.TargetPermanentReference(0),
								HalvePowerToughnessRoundUp: true,
							}),
							Recipient: opt.Val(game.AffectedTargetControllerReference(0)),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "dies-this-way-copy",
							Succeeded: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature. If that creature dies this way, its controller creates two tokens that are copies of that creature, except their power is half that creature's power and their toughness is half that creature's toughness. Round up each time.
		`,
		},
	}
}

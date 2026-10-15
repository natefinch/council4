package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LandMine is the card definition for Land Mine.
//
// Type: Token Artifact
//
// Oracle text:
//   {R}, Sacrifice this artifact: This artifact deals 2 damage to target attacking creature without flying.

// LandMineTokenb143de5d1be54f12a3679e1d33ac5e05 is the card definition for Land Mine.
var LandMineTokenb143de5d1be54f12a3679e1d33ac5e05 = newLandMineTokenb143de5d1be54f12a3679e1d33ac5e05()

func newLandMineTokenb143de5d1be54f12a3679e1d33ac5e05() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Land Mine",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R}, Sacrifice this artifact: This artifact deals 2 damage to target attacking creature without flying.",
					ManaCost: opt.Val(cost.Mana{
						cost.R,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target attacking creature without flying",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking, ExcludedKeyword: game.Flying}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			{R}, Sacrifice this artifact: This artifact deals 2 damage to target attacking creature without flying.
		`,
		},
	}
}

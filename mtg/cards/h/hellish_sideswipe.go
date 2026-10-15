package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HellishSideswipe is the card definition for Hellish Sideswipe.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact or creature.
//	Destroy target creature or Vehicle. If the sacrificed permanent was a Vehicle, draw a card.
var HellishSideswipe = newHellishSideswipe

func newHellishSideswipe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Hellish Sideswipe",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact or creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
					PermanentTypeAlt:   types.Creature,
					SubjectKey:         "paid-cost-1-1",
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.PaidCostReference("paid-cost-1-1", game.PaidCostSacrifice, "")),
								ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice an artifact or creature.
			Destroy target creature or Vehicle. If the sacrificed permanent was a Vehicle, draw a card.
		`,
		},
	}
}

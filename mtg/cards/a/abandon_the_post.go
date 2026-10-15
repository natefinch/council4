package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbandonThePost is the card definition for Abandon the Post.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Up to two target creatures can't block this turn.
//	Flashback {3}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var AbandonThePost = newAbandonThePost

func newAbandonThePost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Abandon the Post",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {3}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.R,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:     game.RuleEffectCantBlock,
									Duration: game.DurationPermanent,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(1)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:     game.RuleEffectCantBlock,
									Duration: game.DurationPermanent,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Up to two target creatures can't block this turn.
			Flashback {3}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

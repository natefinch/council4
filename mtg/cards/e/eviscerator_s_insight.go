package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EvisceratorSInsight is the card definition for Eviscerator's Insight.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact or creature.
//	Draw two cards.
//	Flashback {4}{B} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
var EvisceratorSInsight = newEvisceratorSInsight

func newEvisceratorSInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Eviscerator's Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {4}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.B,
							},
						},
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact or creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
					PermanentTypeAlt:   types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice an artifact or creature.
			Draw two cards.
			Flashback {4}{B} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
		`,
		},
	}
}

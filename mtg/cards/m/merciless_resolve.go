package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MercilessResolve is the card definition for Merciless Resolve.
//
// Type: Instant
// Cost: {2}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature or land.
//	Draw two cards.
var MercilessResolve = newMercilessResolve

func newMercilessResolve() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Merciless Resolve",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature or land",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
					PermanentTypeAlt:   types.Land,
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
			As an additional cost to cast this spell, sacrifice a creature or land.
			Draw two cards.
		`,
		},
	}
}

package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkeletalScrying is the card definition for Skeletal Scrying.
//
// Type: Instant
// Cost: {X}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile X cards from your graveyard.
//	You draw X cards and you lose X life.
var SkeletalScrying = newSkeletalScrying

func newSkeletalScrying() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Skeletal Scrying",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:        cost.AdditionalExile,
					Text:        "exile X cards from your graveyard",
					AmountFromX: true,
					Source:      zone.Graveyard,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, exile X cards from your graveyard.
			You draw X cards and you lose X life.
		`,
		},
	}
}

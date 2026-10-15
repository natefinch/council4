package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpittingImage is the card definition for Spitting Image.
//
// Type: Sorcery
// Cost: {4}{G/U}{G/U}
//
// Oracle text:
//
//	Create a token that's a copy of target creature.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var SpittingImage = newSpittingImage

func newSpittingImage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Spitting Image",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
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
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceObject,
								Object: game.TargetPermanentReference(0),
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a token that's a copy of target creature.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}

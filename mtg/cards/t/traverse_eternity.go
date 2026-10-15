package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TraverseEternity is the card definition for Traverse Eternity.
//
// Type: Sorcery
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Draw cards equal to the greatest mana value among historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)
var TraverseEternity = newTraverseEternity

func newTraverseEternity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Traverse Eternity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountGreatestManaValueInGroup,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact}}, game.Selection{Supertypes: []types.Super{types.Legendary}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Saga")}}}, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw cards equal to the greatest mana value among historic permanents you control. (Artifacts, legendaries, and Sagas are historic.)
		`,
		},
	}
}

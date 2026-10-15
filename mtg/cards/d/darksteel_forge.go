package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelForge is the card definition for Darksteel Forge.
//
// Type: Artifact
// Cost: {9}
//
// Oracle text:
//
//	Artifacts you control have indestructible. (Effects that say "destroy" don't destroy them. Artifact creatures with indestructible can't be destroyed by damage.)
var DarksteelForge = newDarksteelForge

func newDarksteelForge() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Forge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Artifacts you control have indestructible. (Effects that say \"destroy\" don't destroy them. Artifact creatures with indestructible can't be destroyed by damage.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Indestructible},
						},
					},
				},
			},
			OracleText: `
			Artifacts you control have indestructible. (Effects that say "destroy" don't destroy them. Artifact creatures with indestructible can't be destroyed by damage.)
		`,
		},
	}
}

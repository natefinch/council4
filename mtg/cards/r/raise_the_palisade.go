package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RaiseThePalisade is the card definition for Raise the Palisade.
//
// Type: Sorcery
// Cost: {4}{U}
//
// Oracle text:
//
//	Choose a creature type. Return all creatures that aren't of the chosen type to their owners' hands.
var RaiseThePalisade = newRaiseThePalisade

func newRaiseThePalisade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Raise the Palisade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:          game.ResolutionChoiceSubtype,
								Prompt:        "Choose a creature type",
								SubtypeOfType: types.Creature,
							},
							PublishChoice: game.SpellChosenTypeChoiceKey,
						},
					},
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceResolutionExcluded}),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a creature type. Return all creatures that aren't of the chosen type to their owners' hands.
		`,
		},
	}
}

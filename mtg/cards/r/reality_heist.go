package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RealityHeist is the card definition for Reality Heist.
//
// Type: Instant
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
//	Look at the top seven cards of your library. You may reveal up to two artifact cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.
var RealityHeist = newRealityHeist

func newRealityHeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Reality Heist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(7),
							Take:      game.Fixed(2),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
			Look at the top seven cards of your library. You may reveal up to two artifact cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}

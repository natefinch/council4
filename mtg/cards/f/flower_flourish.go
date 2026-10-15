package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Flower is the card definition for Flower // Flourish.
//
// Type: Sorcery // Sorcery
// Cost: {G/W} // {4}{G}{W}
// Face: Flourish — Sorcery ({4}{G}{W})
//
// Oracle text:
//
//	Search your library for a basic Forest or Plains card, reveal it, put it into your hand, then shuffle.
var Flower = newFlower

func newFlower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Flower",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors: []color.Color{color.White, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Forest, types.Plains}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a basic Forest or Plains card, reveal it, put it into your hand, then shuffle.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Flourish",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.White, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     2,
									ToughnessDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +2/+2 until end of turn.
		`,
		}),
	}
}

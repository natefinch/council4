package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Perennation is the card definition for Perennation.
//
// Type: Sorcery
// Cost: {3}{W}{B}{G}
//
// Oracle text:
//
//	Return target permanent card from your graveyard to the battlefield with a hexproof counter and an indestructible counter on it.
var Perennation = newPerennation

func newPerennation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Perennation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							EntryCounters: []game.CounterPlacement{
								game.CounterPlacement{
									Kind:   counter.Hexproof,
									Amount: 1,
								},
								game.CounterPlacement{
									Kind:   counter.Indestructible,
									Amount: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target permanent card from your graveyard to the battlefield with a hexproof counter and an indestructible counter on it.
		`,
		},
	}
}

package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MemoryPlunder is the card definition for Memory Plunder.
//
// Type: Instant
// Cost: {U/B}{U/B}{U/B}{U/B}
//
// Oracle text:
//
//	You may cast target instant or sorcery card from an opponent's graveyard without paying its mana cost.
var MemoryPlunder = newMemoryPlunder

func newMemoryPlunder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Memory Plunder",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target instant or sorcery card from an opponent's graveyard without paying its mana cost",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Zone:   zone.Graveyard,
							Card:   game.CardReference{Kind: game.CardReferenceTarget},
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may cast target instant or sorcery card from an opponent's graveyard without paying its mana cost.
		`,
		},
	}
}

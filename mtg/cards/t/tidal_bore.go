package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TidalBore is the card definition for Tidal Bore.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	You may return an Island you control to its owner's hand rather than pay this spell's mana cost.
//	You may tap or untap target creature.
var TidalBore = newTidalBore

func newTidalBore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tidal Bore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Return an Island you control to its owner's hand",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReturnToHand,
							Text:        "return an Island you control to its owner's hand",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Island},
						},
					},
				},
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
						Primitive: game.TapOrUntap{
							Object: game.TargetPermanentReference(0),
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may return an Island you control to its owner's hand rather than pay this spell's mana cost.
			You may tap or untap target creature.
		`,
		},
	}
}

package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShepherdOfTheFlock is the card definition for Shepherd of the Flock // Usher to Safety.
//
// Type: Creature — Human Peasant // Instant — Adventure
// Cost: {1}{W} // {W}
// Face: Usher to Safety — Instant — Adventure ({W})
//
// Oracle text:
//
//	Usher to Safety
//	Return target permanent you control to its owner's hand. (Then exile this card. You may cast the creature later from exile.)
var ShepherdOfTheFlock = newShepherdOfTheFlock

func newShepherdOfTheFlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shepherd of the Flock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Peasant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Usher to Safety",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target permanent you control to its owner's hand. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}

package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CruelRevival is the card definition for Cruel Revival.
//
// Type: Instant
// Cost: {4}{B}
//
// Oracle text:
//
//	Destroy target non-Zombie creature. It can't be regenerated. Return up to one target Zombie card from your graveyard to your hand.
var CruelRevival = newCruelRevival

func newCruelRevival() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cruel Revival",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target non-Zombie creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Zombie")}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target Zombie card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target non-Zombie creature. It can't be regenerated. Return up to one target Zombie card from your graveyard to your hand.
		`,
		},
	}
}

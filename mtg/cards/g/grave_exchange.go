package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GraveExchange is the card definition for Grave Exchange.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Return target creature card from your graveyard to your hand. Target player sacrifices a creature of their choice.
var GraveExchange = newGraveExchange()

func newGraveExchange() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grave Exchange",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.SacrificePermanents{
							Amount:    game.Fixed(1),
							Player:    game.TargetPlayerReference(1),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to your hand. Target player sacrifices a creature of their choice.
		`,
		},
	}
}

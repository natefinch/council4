package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HorrifyingRevelation is the card definition for Horrifying Revelation.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target player discards a card, then mills a card.
var HorrifyingRevelation = newHorrifyingRevelation

func newHorrifyingRevelation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Horrifying Revelation",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(1),
							Player:      game.TargetPlayerReference(0),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards a card, then mills a card.
		`,
		},
	}
}

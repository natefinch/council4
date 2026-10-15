package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InquisitionOfKozilek is the card definition for Inquisition of Kozilek.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target player reveals their hand. You choose a nonland card from it with mana value 3 or less. That player discards that card.
var InquisitionOfKozilek = newInquisitionOfKozilek

func newInquisitionOfKozilek() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Inquisition of Kozilek",
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
						Primitive: game.ChooseDiscardFromHand{
							Player:       game.TargetPlayerReference(0),
							ExcludeLand:  true,
							MaxManaValue: opt.Val(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player reveals their hand. You choose a nonland card from it with mana value 3 or less. That player discards that card.
		`,
		},
	}
}

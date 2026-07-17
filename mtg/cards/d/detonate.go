package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Detonate is the card definition for Detonate.
//
// Type: Sorcery
// Cost: {X}{R}
//
// Oracle text:
//
//	Destroy target artifact with mana value X. It can't be regenerated. Detonate deals X damage to that artifact's controller.
var Detonate = newDetonate

func newDetonate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Detonate",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:       1,
						MaxTargets:       1,
						Constraint:       "target artifact with mana value X",
						Allow:            game.TargetAllowPermanent,
						Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
						ManaValueEqualsX: true,
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
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact with mana value X. It can't be regenerated. Detonate deals X damage to that artifact's controller.
		`,
		},
	}
}

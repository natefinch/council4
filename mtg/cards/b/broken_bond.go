package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BrokenBond is the card definition for Broken Bond.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Destroy target artifact or enchantment. You may put a land card from your hand onto the battlefield.
var BrokenBond = newBrokenBond

func newBrokenBond() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Broken Bond",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.ChooseFromZone{
							Player:     game.ControllerReference(),
							SourceZone: zone.Hand,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to put onto the battlefield",
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact or enchantment. You may put a land card from your hand onto the battlefield.
		`,
		},
	}
}

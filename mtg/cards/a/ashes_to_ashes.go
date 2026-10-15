package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AshesToAshes is the card definition for Ashes to Ashes.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Exile two target nonartifact creatures. Ashes to Ashes deals 5 damage to you.
var AshesToAshes = newAshesToAshes

func newAshesToAshes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ashes to Ashes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 2,
						MaxTargets: 2,
						Constraint: "two target nonartifact creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Artifact}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.PlayerDamageRecipient(game.ControllerReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile two target nonartifact creatures. Ashes to Ashes deals 5 damage to you.
		`,
		},
	}
}

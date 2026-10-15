package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Browbeat is the card definition for Browbeat.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Any player may have Browbeat deal 5 damage to them. If no one does, target player draws three cards.
var Browbeat = newBrowbeat

func newBrowbeat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Browbeat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.AllPlayersReference()),
						PublishResult:      game.ResultKey("group-may-have-action"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.TargetPlayerReference(0),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "group-may-have-action",
							Accepted: game.TriFalse,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Any player may have Browbeat deal 5 damage to them. If no one does, target player draws three cards.
		`,
		},
	}
}

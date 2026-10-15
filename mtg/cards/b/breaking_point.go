package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreakingPoint is the card definition for Breaking Point.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Any player may have Breaking Point deal 6 damage to them. If no one does, destroy all creatures. Creatures destroyed this way can't be regenerated.
var BreakingPoint = newBreakingPoint

func newBreakingPoint() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Breaking Point",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(6),
							Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.AllPlayersReference()),
						PublishResult:      game.ResultKey("group-may-have-action"),
					},
					{
						Primitive: game.Destroy{
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							PreventRegeneration: true,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "group-may-have-action",
							Accepted: game.TriFalse,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Any player may have Breaking Point deal 6 damage to them. If no one does, destroy all creatures. Creatures destroyed this way can't be regenerated.
		`,
		},
	}
}

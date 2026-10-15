package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BookBurning is the card definition for Book Burning.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Any player may have Book Burning deal 6 damage to them. If no one does, target player mills six cards.
var BookBurning = newBookBurning

func newBookBurning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Book Burning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
							Amount:    game.Fixed(6),
							Recipient: game.PlayerDamageRecipient(game.GroupOfferMemberReference()),
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.AllPlayersReference()),
						PublishResult:      game.ResultKey("group-may-have-action"),
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(6),
							Player:      game.TargetPlayerReference(0),
							Destination: zone.Graveyard,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "group-may-have-action",
							Accepted: game.TriFalse,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Any player may have Book Burning deal 6 damage to them. If no one does, target player mills six cards.
		`,
		},
	}
}

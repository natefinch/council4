package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EnterTheGodEternals is the card definition for Enter the God-Eternals.
//
// Type: Sorcery
// Cost: {2}{U}{U}{B}
//
// Oracle text:
//
//	Enter the God-Eternals deals 4 damage to target creature and you gain life equal to the damage dealt this way. Target player mills four cards. Amass Zombies 4. (Put four +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)
var EnterTheGodEternals = newEnterTheGodEternals

func newEnterTheGodEternals() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Enter the God-Eternals",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
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
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						PublishResult: game.ResultKey("damage-dealt-this-way"),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								ResultKey:  game.ResultKey("damage-dealt-this-way"),
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(4),
							Player:      game.TargetPlayerReference(1),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.Amass{
							Amount:  game.Fixed(4),
							Subtype: types.Zombie,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Enter the God-Eternals deals 4 damage to target creature and you gain life equal to the damage dealt this way. Target player mills four cards. Amass Zombies 4. (Put four +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)
		`,
		},
	}
}

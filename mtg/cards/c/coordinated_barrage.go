package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoordinatedBarrage is the card definition for Coordinated Barrage.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Choose a creature type. Coordinated Barrage deals damage to target attacking or blocking creature equal to the number of permanents you control of the chosen type.
var CoordinatedBarrage = newCoordinatedBarrage

func newCoordinatedBarrage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Coordinated Barrage",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking or blocking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:          game.ResolutionChoiceSubtype,
								Prompt:        "Choose a creature type",
								SubtypeOfType: types.Creature,
							},
							PublishChoice: game.SpellChosenTypeChoiceKey,
						},
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypeChoice: game.SubtypeChoiceSourceEntry, Controller: game.ControllerYou}),
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a creature type. Coordinated Barrage deals damage to target attacking or blocking creature equal to the number of permanents you control of the chosen type.
		`,
		},
	}
}

package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfBounty is the card definition for Curse of Bounty.
//
// Type: Enchantment — Aura Curse
// Cost: {1}{G}
//
// Oracle text:
//
//	Enchant player
//	Whenever enchanted player is attacked, untap all nonland permanents you control. Each opponent attacking that player untaps all nonland permanents they control.
var CurseOfBounty = newCurseOfBounty

func newCurseOfBounty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Curse of Bounty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Curse},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "player",
					Allow:      game.TargetAllowPlayer,
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                                 game.EventAttackerDeclared,
							OneOrMore:                             true,
							AttackedPlayerIsSourceEnchantedPlayer: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
								},
							},
							{
								Primitive: game.Untap{
									Group: game.PlayerControlledGroup(game.GroupOfferMemberReference(), game.Selection{ExcludedTypes: []types.Card{types.Land}}),
								},
								ForEachPlayerGroup: opt.Val(game.OpponentsAttackingTriggerPlayerReference()),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant player
			Whenever enchanted player is attacked, untap all nonland permanents you control. Each opponent attacking that player untaps all nonland permanents they control.
		`,
		},
	}
}

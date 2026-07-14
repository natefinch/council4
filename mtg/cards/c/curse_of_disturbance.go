package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfDisturbance is the card definition for Curse of Disturbance.
//
// Type: Enchantment — Aura Curse
// Cost: {2}{B}
//
// Oracle text:
//
//	Enchant player
//	Whenever enchanted player is attacked, create a 2/2 black Zombie creature token. Each opponent attacking that player does the same.
var CurseOfDisturbance = newCurseOfDisturbance

func newCurseOfDisturbance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Curse of Disturbance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(curseOfDisturbanceToken),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount:         game.Fixed(1),
									Source:         game.TokenDef(curseOfDisturbanceToken),
									RecipientGroup: game.OpponentsAttackingTriggerPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant player
			Whenever enchanted player is attacked, create a 2/2 black Zombie creature token. Each opponent attacking that player does the same.
		`,
		},
	}
}

var curseOfDisturbanceToken = newCurseOfDisturbanceToken()

func newCurseOfDisturbanceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}

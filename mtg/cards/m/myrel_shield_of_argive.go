package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MyrelShieldOfArgive is the card definition for Myrel, Shield of Argive.
//
// Type: Legendary Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.
//	Whenever Myrel attacks, create X 1/1 colorless Soldier artifact creature tokens, where X is the number of Soldiers you control.
var MyrelShieldOfArgive = newMyrelShieldOfArgive

func newMyrelShieldOfArgive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Myrel, Shield of Argive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectCantCastSpells,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerOpponent,
							RestrictedDuringControllerTurn: true,
						},
						game.RuleEffect{
							Kind:                           game.RuleEffectCantActivateAbilities,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerOpponent,
							PermanentTypes:                 []types.Card{types.Artifact, types.Creature, types.Enchantment},
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Myrel attacks, create X 1/1 colorless Soldier artifact creature tokens, where X is the number of Soldiers you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Soldier")}, Controller: game.ControllerYou}),
											}),
											Source: game.TokenDef(myrelShieldOfArgiveToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.
			Whenever Myrel attacks, create X 1/1 colorless Soldier artifact creature tokens, where X is the number of Soldiers you control.
		`,
		},
	}
}

var myrelShieldOfArgiveToken = newMyrelShieldOfArgiveToken()

func newMyrelShieldOfArgiveToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

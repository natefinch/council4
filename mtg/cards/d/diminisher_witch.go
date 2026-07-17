package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DiminisherWitch is the card definition for Diminisher Witch.
//
// Type: Creature — Human Warlock
// Cost: {2}{U}
//
// Oracle text:
//
//	Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
//	When this creature enters, if it was bargained, create a Cursed Role token attached to target creature an opponent controls. (If you control another Role on it, put that one into the graveyard. Enchanted creature is 1/1.)
var DiminisherWitch = newDiminisherWitch

func newDiminisherWitch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Diminisher Witch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warlock},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.BargainStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                           "if it was bargained",
						InterveningIfEventPermanentWasBargained: true,
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(diminisherWitchToken),
									EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
			When this creature enters, if it was bargained, create a Cursed Role token attached to target creature an opponent controls. (If you control another Role on it, put that one into the graveyard. Enchanted creature is 1/1.)
		`,
		},
	}
}

var diminisherWitchToken = newDiminisherWitchToken()

func newDiminisherWitchToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Cursed Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetPower:     opt.Val(game.PT{Value: 1}),
							SetToughness: opt.Val(game.PT{Value: 1}),
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has base power and toughness 1/1.
		`,
		},
	}
}

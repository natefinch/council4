package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReturnTriumphant is the card definition for Return Triumphant.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	Return target creature card with mana value 3 or less from your graveyard to the battlefield. Create a Young Hero Role token attached to it. (Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it." If you put another Role on the creature later, put this one into the graveyard.)
var ReturnTriumphant = newReturnTriumphant

func newReturnTriumphant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Return Triumphant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card with mana value 3 or less from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:          game.Fixed(1),
							Source:          game.TokenDef(returnTriumphantToken),
							EntryAttachedTo: opt.Val(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card with mana value 3 or less from your graveyard to the battlefield. Create a Young Hero Role token attached to it. (Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it." If you put another Role on the creature later, put this one into the graveyard.)
		`,
		},
	}
}

var returnTriumphantToken = newReturnTriumphantToken()

func newReturnTriumphantToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Young Hero Role",
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
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:  game.EventAttackerDeclared,
											Source: game.TriggerSourceSelf,
										},
										InterveningIf: "if its toughness is 3 or less",
										InterveningCondition: opt.Val(game.Condition{
											Object:        opt.Val(game.EventPermanentReference()),
											ObjectMatches: opt.Val(game.Selection{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
										}),
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.AddCounter{
													Amount:      game.Fixed(1),
													Object:      game.EventPermanentReference(),
													CounterKind: counter.PlusOnePlusOne,
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it."
		`,
		},
	}
}

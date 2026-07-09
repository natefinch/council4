package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlelaCunningConqueror is the card definition for Alela, Cunning Conqueror.
//
// Type: Legendary Creature — Faerie Warlock
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	Flying
//	Whenever you cast your first spell during each opponent's turn, create a 1/1 black Faerie Rogue creature token with flying.
//	Whenever one or more Faeries you control deal combat damage to a player, goad target creature that player controls.
var AlelaCunningConqueror = newAlelaCunningConqueror

func newAlelaCunningConqueror() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Alela, Cunning Conqueror",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Faerie, types.Warlock},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							CastDuringTurn:             game.TriggerTurnNotYours,
							PlayerEventOrdinalThisTurn: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(alelaCunningConquerorToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							OneOrMore:             true,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Faerie")}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature that player controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ControlledByEventPlayer: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Goad{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever you cast your first spell during each opponent's turn, create a 1/1 black Faerie Rogue creature token with flying.
			Whenever one or more Faeries you control deal combat damage to a player, goad target creature that player controls.
		`,
		},
	}
}

var alelaCunningConquerorToken = newAlelaCunningConquerorToken()

func newAlelaCunningConquerorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Faerie Rogue",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}

package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CombatCalligrapher is the card definition for Combat Calligrapher.
//
// Type: Creature — Bird Cleric
// Cost: {3}{W}
//
// Oracle text:
//
//	Flying
//	Inklings can't attack you or planeswalkers you control.
//	Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent.
var CombatCalligrapher = newCombatCalligrapher

func newCombatCalligrapher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Combat Calligrapher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Inklings can't attack you or planeswalkers you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCantAttack,
							Duration:          game.DurationPermanent,
							DefendingPlayer:   game.PlayerYou,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Inkling}},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerOpponent,
							StackObjectKind:          game.StackSpell,
							AttackRecipient:          game.AttackRecipientPlayer,
							OneOrMore:                true,
							OneOrMorePerAttackTarget: true,
							CounterKind:              counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:                 game.Fixed(1),
											Source:                 game.TokenDef(combatCalligrapherToken),
											Recipient:              opt.Val(game.EventPlayerReference()),
											EntryTapped:            true,
											EntryAttackingDefender: opt.Val(game.DefendingPlayerReference()),
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
			Flying
			Inklings can't attack you or planeswalkers you control.
			Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent.
		`,
		},
	}
}

var combatCalligrapherToken = newCombatCalligrapherToken()

func newCombatCalligrapherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Inkling",
			Colors:    []color.Color{color.White, color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Inkling},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}

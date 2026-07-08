package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MinnWilyIllusionist is the card definition for Minn, Wily Illusionist.
//
// Type: Legendary Creature — Gnome Wizard
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Whenever you draw your second card each turn, create a 1/1 blue Illusion creature token with "This token gets +1/+0 for each other Illusion you control."
//	Whenever an Illusion you control dies, you may put a permanent card with mana value less than or equal to that creature's power from your hand onto the battlefield.
var MinnWilyIllusionist = newMinnWilyIllusionist

func newMinnWilyIllusionist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Minn, Wily Illusionist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Gnome, types.Wizard},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(minnWilyIllusionistToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Illusion")}},
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you draw your second card each turn, create a 1/1 blue Illusion creature token with "This token gets +1/+0 for each other Illusion you control."
			Whenever an Illusion you control dies, you may put a permanent card with mana value less than or equal to that creature's power from your hand onto the battlefield.
		`,
		},
	}
}

var minnWilyIllusionistToken = newMinnWilyIllusionistToken()

func newMinnWilyIllusionistToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Illusion")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
						},
					},
				},
			},
		},
	}
}

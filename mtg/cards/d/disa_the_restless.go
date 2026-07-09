package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DisaTheRestless is the card definition for Disa the Restless.
//
// Type: Legendary Creature — Human Scout
// Cost: {2}{B}{R}{G}
//
// Oracle text:
//
//	Whenever a Lhurgoyf permanent card is put into your graveyard from anywhere other than the battlefield, put it onto the battlefield.
//	Whenever one or more creatures you control deal combat damage to a player, create a Tarmogoyf token.
var DisaTheRestless = newDisaTheRestless

func newDisaTheRestless() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Disa the Restless",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Scout},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							ExcludeFromZone:  true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Lhurgoyf")}},
						},
					},
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
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(disaTheRestlessToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a Lhurgoyf permanent card is put into your graveyard from anywhere other than the battlefield, put it onto the battlefield.
			Whenever one or more creatures you control deal combat damage to a player, create a Tarmogoyf token.
		`,
		},
	}
}

var disaTheRestlessToken = newDisaTheRestlessToken()

func newDisaTheRestlessToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tarmogoyf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lhurgoyf},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueCardTypesAmongAllGraveyards, Offset: 1}),
		},
	}
}

package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BasriTomorrowSChampion is the card definition for Basri, Tomorrow's Champion.
//
// Type: Legendary Creature — Human Knight
// Cost: {W}
//
// Oracle text:
//
//	{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with lifelink. (An exerted creature won't untap during your next untap step.)
//	Cycling {2}{W} ({2}{W}, Discard this card: Draw a card.)
//	When you cycle this card, Cats you control gain hexproof and indestructible until end of turn.
var BasriTomorrowSChampion = newBasriTomorrowSChampion()

func newBasriTomorrowSChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Basri, Tomorrow's Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Knight},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with lifelink. (An exerted creature won't untap during your next untap step.)",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind: cost.AdditionalExert,
							Text: "Exert Basri",
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(basriTomorrowSChampionToken),
								},
							},
						},
					}.Ability(),
				},
				game.CyclingActivatedAbility(cost.Mana{cost.O(2), cost.W}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventCycled,
							Source: game.TriggerSourceSelf,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cat")}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Hexproof,
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}, {T}, Exert Basri: Create a 1/1 white Cat creature token with lifelink. (An exerted creature won't untap during your next untap step.)
			Cycling {2}{W} ({2}{W}, Discard this card: Draw a card.)
			When you cycle this card, Cats you control gain hexproof and indestructible until end of turn.
		`,
		},
	}
}

var basriTomorrowSChampionToken = newBasriTomorrowSChampionToken()

func newBasriTomorrowSChampionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
		},
	}
}

package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FeastingTrollKing is the card definition for Feasting Troll King.
//
// Type: Creature — Troll Noble
// Cost: {2}{G}{G}{G}{G}
//
// Oracle text:
//
//	Vigilance, trample
//	When this creature enters, if you cast it from your hand, create three Food tokens. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	Sacrifice three Foods: Return this card from your graveyard to the battlefield. Activate only during your turn.
var FeastingTrollKing = newFeastingTrollKing

func newFeastingTrollKing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Feasting Troll King",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Noble},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice three Foods: Return this card from your graveyard to the battlefield. Activate only during your turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice three Foods",
							Amount:      3,
							SubtypesAny: cost.SubtypeSet{types.Sub("Food")},
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.DuringYourTurn,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you cast it from your hand",
						InterveningIfEventPermanentWasCastFromControllerHand: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(feastingTrollKingToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, trample
			When this creature enters, if you cast it from your hand, create three Food tokens. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			Sacrifice three Foods: Return this card from your graveyard to the battlefield. Activate only during your turn.
		`,
		},
	}
}

var feastingTrollKingToken = newFeastingTrollKingToken()

func newFeastingTrollKingToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

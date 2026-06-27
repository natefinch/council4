package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DefendersOfHumanity is the card definition for Defenders of Humanity.
//
// Type: Enchantment
// Cost: {X}{2}{W}
//
// Oracle text:
//
//	When this enchantment enters, create X 2/2 white Astartes Warrior creature tokens with vigilance.
//	{X}{2}{W}, Exile this enchantment: Create X 2/2 white Astartes Warrior creature tokens with vigilance. Activate only if you control no creatures and only during your turn.
var DefendersOfHumanity = newDefendersOfHumanity()

func newDefendersOfHumanity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Defenders of Humanity",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{X}{2}{W}, Exile this enchantment: Create X 2/2 white Astartes Warrior creature tokens with vigilance. Activate only if you control no creatures and only during your turn.",
					ManaCost: opt.Val(cost.Mana{cost.X, cost.O(2), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile this enchantment",
							Amount: 1,
							Source: zone.Battlefield,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.DuringYourTurn,
					ActivationCondition: opt.Val(game.Condition{
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							MinCount:  1,
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Source: game.TokenDef(defendersOfHumanityToken),
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
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Source: game.TokenDef(defendersOfHumanityToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, create X 2/2 white Astartes Warrior creature tokens with vigilance.
			{X}{2}{W}, Exile this enchantment: Create X 2/2 white Astartes Warrior creature tokens with vigilance. Activate only if you control no creatures and only during your turn.
		`,
		},
	}
}

var defendersOfHumanityToken = newDefendersOfHumanityToken()

func newDefendersOfHumanityToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Astartes Warrior",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Astartes, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}

package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TChakaVenerableKing is the card definition for T'Chaka, Venerable King.
//
// Type: Legendary Creature — Human Noble Hero
// Cost: {G}{W}
//
// Oracle text:
//
//	When T'Chaka enters, mill three cards, then you may put an artifact or land card from among the milled cards into your hand.
//	{3}, Exile this card from your graveyard: You become the monarch. Activate only if you control your commander.
var TChakaVenerableKing = newTChakaVenerableKing()

func newTChakaVenerableKing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "T'Chaka, Venerable King",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}, Exile this card from your graveyard: You become the monarch. Activate only if you control your commander.",
					ManaCost: opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile this card from your graveyard",
							Amount: 1,
							Source: zone.Graveyard,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					ActivationCondition: opt.Val(game.Condition{
						ControllerControlsCommander: true,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
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
								Primitive: game.Dig{
									Player:   game.ControllerReference(),
									Look:     game.Fixed(3),
									Take:     game.Fixed(1),
									Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Land}}),
									TakeUpTo: true,
									Reveal:   true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When T'Chaka enters, mill three cards, then you may put an artifact or land card from among the milled cards into your hand.
			{3}, Exile this card from your graveyard: You become the monarch. Activate only if you control your commander.
		`,
		},
	}
}

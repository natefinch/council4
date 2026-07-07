package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KingSolomonSFrogs is the card definition for King Solomon's Frogs.
//
// Type: Legendary Artifact
// Cost: {3}{W}
//
// Oracle text:
//
//	Flash
//	When King Solomon's Frogs enters, if you cast it, for each opponent, exile up to one target permanent that player controls with mana value 3 or greater. For each permanent exiled this way, its controller draws a card.
//	{3}, {T}, Exile King Solomon's Frogs: You become the monarch.
var KingSolomonSFrogs = newKingSolomonSFrogs()

func newKingSolomonSFrogs() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "King Solomon's Frogs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}, {T}, Exile King Solomon's Frogs: You become the monarch.",
					ManaCost: opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalExileSource,
							Text:   "Exile King Solomon's Frogs",
							Amount: 1,
							Source: zone.Battlefield,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
						InterveningIf: "if you cast it",
						InterveningIfEventPermanentWasCastByController: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ExileForEachOpponent{
									Chooser:   game.ControllerReference(),
									Selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})},
									LinkedKey: game.LinkedKey("exiled-for-each-opponent"),
								},
							},
							{
								Primitive: game.DrawForEachExiled{
									LinkedKey: game.LinkedKey("exiled-for-each-opponent"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			When King Solomon's Frogs enters, if you cast it, for each opponent, exile up to one target permanent that player controls with mana value 3 or greater. For each permanent exiled this way, its controller draws a card.
			{3}, {T}, Exile King Solomon's Frogs: You become the monarch.
		`,
		},
	}
}

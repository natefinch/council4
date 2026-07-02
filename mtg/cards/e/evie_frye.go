package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EvieFrye is the card definition for Evie Frye.
//
// Type: Legendary Creature — Human Assassin
// Cost: {1}{U}
//
// Oracle text:
//
//	Partner with Jacob Frye (When this creature enters, target player may put Jacob into their hand from their library, then shuffle.)
//	{1}, {T}: Draw a card, then discard a card. When you discard a creature card this way, target creature you control can't be blocked this turn.
var EvieFrye = newEvieFrye()

func newEvieFrye() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Evie Frye",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Draw a card, then discard a card. When you discard a creature card this way, target creature you control can't be blocked this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.ChooseDiscardFromHand{
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Partner with Jacob Frye (When this creature enters, target player may put Jacob into their hand from their library, then shuffle.)
			{1}, {T}: Draw a card, then discard a card. When you discard a creature card this way, target creature you control can't be blocked this turn.
		`,
		},
	}
}

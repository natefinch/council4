package b

import (
	"github.com/natefinch/council4/mtg/cards/common"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BridgeworksBattle is the card definition for Bridgeworks Battle // Tanglespan Bridgeworks.
//
// Type: Sorcery // Land
// Face: Bridgeworks Battle — Sorcery ({2}{G})
// Face: Tanglespan Bridgeworks — Land
//
// Front oracle text:
//
//	Target creature you control gets +2/+2 until end of turn. It fights up to
//	one target creature you don't control. (Each deals damage equal to its power
//	to the other.)
//
// Back oracle text:
//
//	As this land enters, you may pay 3 life. If you don't, it enters tapped.
//	{T}: Add {G}.
var BridgeworksBattle = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bridgeworks Battle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			OracleText: `
				Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
			`,
			SpellAbility: opt.Val(
				game.Mode{
					Targets: []game.TargetSpec{
						{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "creature you control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerYou,
							},
						},
						{

							MinTargets: 0,
							MaxTargets: 1,
							Constraint: "creature you don't control",
							Allow:      game.TargetAllowPermanent,
							Predicate: game.TargetPredicate{
								PermanentTypes: []types.Card{
									types.Creature,
								},
								Controller: game.ControllerNotYou,
							},
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.ModifyPT{
								TargetIndex:    0,
								PowerDelta:     game.Fixed(2),
								ToughnessDelta: game.Fixed(2),
								Duration:       game.DurationUntilEndOfTurn,
							},
						},
						{
							Primitive: game.Fight{
								TargetIndex:        0,
								RelatedTargetIndex: opt.Val(1),
							},
							Description: "target creature you control fights up to one target creature you don't control",
						},
					},
				}.Ability(),
			),
		},
		Layout: game.LayoutModalDFC,
	}

	back := game.CardFace{
		Name:  "Tanglespan Bridgeworks",
		Types: []types.Card{types.Land},
		OracleText: `
			As this land enters, you may pay 3 life. If you don't, it enters tapped.
			{T}: Add {G}.
		`,
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 3 life. If you don't, it enters tapped.", game.ResolutionPayment{
				Prompt: "Pay 3 life?",
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalPayLife, Amount: 3, Text: "Pay 3 life"},
				},
			}),
		},
	}

	back.ManaAbilities = append(back.ManaAbilities, common.TapForOne(mana.G))

	card.Back = opt.Val(back)
	return card
}()

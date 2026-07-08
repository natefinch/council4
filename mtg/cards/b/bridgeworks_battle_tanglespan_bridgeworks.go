package b

import (
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
// Face: Tanglespan Bridgeworks — Land
//
// Oracle text:
//
//	Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
var BridgeworksBattle = newBridgeworksBattle

func newBridgeworksBattle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bridgeworks Battle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Fight{
							Object:        game.TargetPermanentReference(0),
							RelatedObject: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature you control gets +2/+2 until end of turn. It fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Tanglespan Bridgeworks",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 3 life. If you don't, it enters tapped.", game.ResolutionPayment{
					Prompt: "Pay 3 life?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Amount: 3,
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may pay 3 life. If you don't, it enters tapped.
			{T}: Add {G}.
		`,
		}),
	}
}

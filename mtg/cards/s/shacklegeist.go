package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Shacklegeist is the card definition for Shacklegeist.
//
// Type: Creature — Spirit
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying
//	This creature can block only creatures with flying.
//	Tap two untapped Spirits you control: Tap target creature you don't control.
var Shacklegeist = newShacklegeist()

func newShacklegeist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Shacklegeist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap two untapped Spirits you control: Tap target creature you don't control.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap two untapped Spirits you control",
							Amount:      2,
							SubtypesAny: cost.SubtypeSet{types.Spirit},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you don't control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			This creature can block only creatures with flying.
			Tap two untapped Spirits you control: Tap target creature you don't control.
		`,
		},
	}
}

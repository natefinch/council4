package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SarythTheViperSFang is the card definition for Saryth, the Viper's Fang.
//
// Type: Legendary Creature — Human Warlock
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Other tapped creatures you control have deathtouch.
//	Other untapped creatures you control have hexproof.
//	{1}, {T}: Untap another target creature or land you control.
var SarythTheViperSFang = newSarythTheViperSFang

func newSarythTheViperSFang() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Saryth, the Viper's Fang",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warlock},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Deathtouch,
							},
						},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Hexproof,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Untap another target creature or land you control.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature or land you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}, Controller: game.ControllerYou, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Other tapped creatures you control have deathtouch.
			Other untapped creatures you control have hexproof.
			{1}, {T}: Untap another target creature or land you control.
		`,
		},
	}
}

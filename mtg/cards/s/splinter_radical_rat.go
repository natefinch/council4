package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SplinterRadicalRat is the card definition for Splinter, Radical Rat.
//
// Type: Legendary Creature — Mutant Ninja Rat
// Cost: {1}{W/B}{W/B}
//
// Oracle text:
//
//	If a triggered ability of a Ninja creature you control triggers, that ability triggers an additional time.
//	{1}{U}: Target Ninja can't be blocked this turn.
var SplinterRadicalRat = newSplinterRadicalRat

func newSplinterRadicalRat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Splinter, Radical Rat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.B),
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Ninja, types.Rat},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
							AffectedSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Ninja")}},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{U}: Target Ninja can't be blocked this turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Ninja",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Ninja")}}),
							},
						},
						Sequence: []game.Instruction{
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
			If a triggered ability of a Ninja creature you control triggers, that ability triggers an additional time.
			{1}{U}: Target Ninja can't be blocked this turn.
		`,
		},
	}
}

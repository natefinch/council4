package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArchwayOfInnovation is the card definition for Archway of Innovation.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control an Island.
//	{T}: Add {U}.
//	{U}, {T}: The next spell you cast this turn has improvise. (Your artifacts can help cast that spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
var ArchwayOfInnovation = newArchwayOfInnovation

func newArchwayOfInnovation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Archway of Innovation",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{U}, {T}: The next spell you cast this turn has improvise. (Your artifacts can help cast that spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)",
					ManaCost:        opt.Val(cost.Mana{cost.U}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:                   game.RuleEffectGrantSpellKeyword,
											AffectedController:     game.ControllerYou,
											GrantedKeyword:         game.Improvise,
											AppliesToNextSpellOnly: true,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control an Island.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control an Island.
			{T}: Add {U}.
			{U}, {T}: The next spell you cast this turn has improvise. (Your artifacts can help cast that spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
		`,
		},
	}
}

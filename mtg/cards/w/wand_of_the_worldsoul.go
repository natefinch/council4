package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WandOfTheWorldsoul is the card definition for Wand of the Worldsoul.
//
// Type: Artifact
// Cost: {2}{W}
//
// Oracle text:
//
//	This artifact enters tapped.
//	{T}: Add {W}.
//	{T}: The next spell you cast this turn has convoke.
var WandOfTheWorldsoul = newWandOfTheWorldsoul

func newWandOfTheWorldsoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wand of the Worldsoul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: The next spell you cast this turn has convoke.",
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
											GrantedKeyword:         game.Convoke,
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
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			OracleText: `
			This artifact enters tapped.
			{T}: Add {W}.
			{T}: The next spell you cast this turn has convoke.
		`,
		},
	}
}

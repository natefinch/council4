package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ProgenitorSIcon is the card definition for Progenitor's Icon.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose a creature type.
//	{T}: Add one mana of any color.
//	{T}: The next spell of the chosen type you cast this turn can be cast as though it had flash.
var ProgenitorSIcon = newProgenitorSIcon

func newProgenitorSIcon() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Progenitor's Icon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: The next spell of the chosen type you cast this turn can be cast as though it had flash.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "The next spell of the chosen type you cast this turn can be cast as though it had flash.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:                   game.RuleEffectCastSpellsAsThoughFlash,
													Duration:               game.DurationPermanent,
													AffectedPlayer:         game.PlayerYou,
													AppliesToNextSpellOnly: true,
													SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,
												},
											},
											Duration: game.DurationThisTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			As this artifact enters, choose a creature type.
			{T}: Add one mana of any color.
			{T}: The next spell of the chosen type you cast this turn can be cast as though it had flash.
		`,
		},
	}
}

package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrecognitionField is the card definition for Precognition Field.
//
// Type: Enchantment
// Cost: {3}{U}
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may cast instant and sorcery spells from the top of your library.
//	{3}: Exile the top card of your library.
var PrecognitionField = newPrecognitionField

func newPrecognitionField() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Precognition Field",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				game.StaticAbility{
					Text: "You may cast instant and sorcery spells from the top of your library.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Instant, types.Sorcery},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}: Exile the top card of your library.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(1),
											Player:      game.ControllerReference(),
											Destination: zone.Exile,
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
			OracleText: `
			You may look at the top card of your library any time.
			You may cast instant and sorcery spells from the top of your library.
			{3}: Exile the top card of your library.
		`,
		},
	}
}

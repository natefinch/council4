package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MysticForge is the card definition for Mystic Forge.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may cast artifact spells and colorless spells from the top of your library.
//	{T}, Pay 1 life: Exile the top card of your library.
var MysticForge = newMysticForge

func newMysticForge() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mystic Forge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				game.StaticAbility{
					Text: "You may cast artifact spells and colorless spells from the top of your library.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Artifact},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
							SpellColorless: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Pay 1 life: Exile the top card of your library.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 1 life",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
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
			You may cast artifact spells and colorless spells from the top of your library.
			{T}, Pay 1 life: Exile the top card of your library.
		`,
		},
	}
}

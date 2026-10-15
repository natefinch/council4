package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GollumSBite is the card definition for Gollum's Bite.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Target creature gets -2/-2 until end of turn.
//	{3}{B}, Exile this card from your graveyard: The Ring tempts you. Activate only as a sorcery.
var GollumSBite = newGollumSBite

func newGollumSBite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gollum's Bite",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{B}, Exile this card from your graveyard: The Ring tempts you. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this card from your graveyard",
							Amount:      1,
							Source:      zone.Graveyard,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.RingTempts{
											Player: game.ControllerReference(),
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
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-2),
							ToughnessDelta: game.Fixed(-2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets -2/-2 until end of turn.
			{3}{B}, Exile this card from your graveyard: The Ring tempts you. Activate only as a sorcery.
		`,
		},
	}
}

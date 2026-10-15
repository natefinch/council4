package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Mutagen is the card definition for Mutagen.
//
// Type: Token Artifact — Mutagen
//
// Oracle text:
//   {1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.

// MutagenToken4c1e395a94c04e83a621b3387bfebbdf is the card definition for Mutagen.
var MutagenToken4c1e395a94c04e83a621b3387bfebbdf = newMutagenToken4c1e395a94c04e83a621b3387bfebbdf()

func newMutagenToken4c1e395a94c04e83a621b3387bfebbdf() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Mutagen",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Mutagen},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this token",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.
		`,
		},
	}
}

package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TowerOfTheMagistrate is the card definition for Tower of the Magistrate.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}, {T}: Target creature gains protection from artifacts until end of turn.
var TowerOfTheMagistrate = newTowerOfTheMagistrate

func newTowerOfTheMagistrate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Tower of the Magistrate",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Target creature gains protection from artifacts until end of turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
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
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddAbilities: []game.Ability{
												new(game.ProtectionFromTypesStaticAbility(types.Artifact)),
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{1}, {T}: Target creature gains protection from artifacts until end of turn.
		`,
		},
	}
}

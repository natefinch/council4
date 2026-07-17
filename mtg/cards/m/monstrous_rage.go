package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MonstrousRage is the card definition for Monstrous Rage.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Target creature gets +2/+0 until end of turn. Create a Monster Role token attached to it. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has trample.)
var MonstrousRage = newMonstrousRage

func newMonstrousRage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Monstrous Rage",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
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
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:          game.Fixed(1),
							Source:          game.TokenDef(monstrousRageToken),
							EntryAttachedTo: opt.Val(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +2/+0 until end of turn. Create a Monster Role token attached to it. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has trample.)
		`,
		},
	}
}

var monstrousRageToken = newMonstrousRageToken()

func newMonstrousRageToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Monster Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has trample.
		`,
		},
	}
}

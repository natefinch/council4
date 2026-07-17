package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoyalTreatment is the card definition for Royal Treatment.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Target creature you control gains hexproof until end of turn. Create a Royal Role token attached to that creature. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has ward {1}.)
var RoyalTreatment = newRoyalTreatment

func newRoyalTreatment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Royal Treatment",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Hexproof,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount:          game.Fixed(1),
							Source:          game.TokenDef(royalTreatmentToken),
							EntryAttachedTo: opt.Val(game.TargetPermanentReference(0)),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature you control gains hexproof until end of turn. Create a Royal Role token attached to that creature. (If you control another Role on it, put that one into the graveyard. Enchanted creature gets +1/+1 and has ward {1}.)
		`,
		},
	}
}

var royalTreatmentToken = newRoyalTreatmentToken()

func newRoyalTreatmentToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Royal Role",
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
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
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
			Enchanted creature gets +1/+1 and has ward {1}.
			(Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {1}.)
		`,
		},
	}
}

package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StreetUrchin is the card definition for Street Urchin.
//
// Type: Legendary Enchantment — Background
// Cost: {1}{R}
//
// Oracle text:
//
//	Commander creatures you own have "{1}, Sacrifice another creature or an artifact: This creature deals 1 damage to any target."
var StreetUrchin = newStreetUrchin

func newStreetUrchin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Street Urchin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Owner: game.OwnerYou, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:     "{1}, Sacrifice another creature or an artifact: This creature deals 1 damage to any target.",
									ManaCost: opt.Val(cost.Mana{cost.O(1)}),
									AdditionalCosts: []cost.Additional{
										{
											Kind:               cost.AdditionalSacrifice,
											Text:               "Sacrifice another creature or an artifact",
											Amount:             1,
											MatchPermanentType: true,
											PermanentType:      types.Creature,
											PermanentTypeAlt:   types.Artifact,
											ExcludeSource:      true,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "any target",
												Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.Damage{
													Amount:       game.Fixed(1),
													Recipient:    game.AnyTargetDamageRecipient(0),
													DamageSource: opt.Val(game.SourcePermanentReference()),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Commander creatures you own have "{1}, Sacrifice another creature or an artifact: This creature deals 1 damage to any target."
		`,
		},
	}
}

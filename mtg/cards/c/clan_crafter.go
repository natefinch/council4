package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ClanCrafter is the card definition for Clan Crafter.
//
// Type: Legendary Enchantment — Background
// Cost: {1}{U}
//
// Oracle text:
//
//	Commander creatures you own have "{2}, Sacrifice an artifact: Put a +1/+1 counter on this creature and draw a card."
var ClanCrafter = newClanCrafter

func newClanCrafter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Clan Crafter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Background},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:     "{2}, Sacrifice an artifact: Put a +1/+1 counter on this creature and draw a card.",
									ManaCost: opt.Val(cost.Mana{cost.O(2)}),
									AdditionalCosts: []cost.Additional{
										{
											Kind:               cost.AdditionalSacrifice,
											Text:               "Sacrifice an artifact",
											Amount:             1,
											MatchPermanentType: true,
											PermanentType:      types.Artifact,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.AddCounter{
													Amount:      game.Fixed(1),
													Object:      game.SourcePermanentReference(),
													CounterKind: counter.PlusOnePlusOne,
												},
											},
											{
												Primitive: game.Draw{
													Amount: game.Fixed(1),
													Player: game.ControllerReference(),
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
			Commander creatures you own have "{2}, Sacrifice an artifact: Put a +1/+1 counter on this creature and draw a card."
		`,
		},
	}
}

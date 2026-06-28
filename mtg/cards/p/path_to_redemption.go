package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PathToRedemption is the card definition for Path to Redemption.
//
// Type: Enchantment — Aura
// Cost: {1}{W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature can't attack or block.
//	{5}, Sacrifice this Aura: Exile enchanted creature. Create a 1/1 white Ally creature token. Activate only during your turn.
var PathToRedemption = newPathToRedemption()

func newPathToRedemption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Path to Redemption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantAttack,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantBlock,
							AffectedAttached: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{5}, Sacrifice this Aura: Exile enchanted creature. Create a 1/1 white Ally creature token. Activate only during your turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(5)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this Aura",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.DuringYourTurn,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(pathToRedemptionToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature can't attack or block.
			{5}, Sacrifice this Aura: Exile enchanted creature. Create a 1/1 white Ally creature token. Activate only during your turn.
		`,
		},
	}
}

var pathToRedemptionToken = newPathToRedemptionToken()

func newPathToRedemptionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

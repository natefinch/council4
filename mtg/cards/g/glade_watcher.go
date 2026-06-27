package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GladeWatcher is the card definition for Glade Watcher.
//
// Type: Creature — Elemental
// Cost: {1}{G}
//
// Oracle text:
//
//	Defender
//	Formidable — {G}: This creature can attack this turn as though it didn't have defender. Activate only if creatures you control have total power 8 or greater.
var GladeWatcher = newGladeWatcher()

func newGladeWatcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Glade Watcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "Formidable — {G}: This creature can attack this turn as though it didn't have defender. Activate only if creatures you control have total power 8 or greater.",
					ManaCost:       opt.Val(cost.Mana{cost.G}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}},
							TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCanAttackAsThoughDefender,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			Formidable — {G}: This creature can attack this turn as though it didn't have defender. Activate only if creatures you control have total power 8 or greater.
		`,
		},
	}
}

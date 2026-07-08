package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MobileFort is the card definition for Mobile Fort.
//
// Type: Artifact Creature — Wall
// Cost: {4}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	{3}: This creature gets +3/-1 until end of turn and can attack this turn as though it didn't have defender. Activate only once each turn.
var MobileFort = newMobileFort

func newMobileFort() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mobile Fort",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}: This creature gets +3/-1 until end of turn and can attack this turn as though it didn't have defender. Activate only once each turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3)}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(3),
									ToughnessDelta: game.Fixed(-1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
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
			Defender (This creature can't attack.)
			{3}: This creature gets +3/-1 until end of turn and can attack this turn as though it didn't have defender. Activate only once each turn.
		`,
		},
	}
}

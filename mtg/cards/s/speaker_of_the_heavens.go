package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpeakerOfTheHeavens is the card definition for Speaker of the Heavens.
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//
//	Vigilance, lifelink (Attacking doesn't cause this creature to tap. Damage dealt by this creature also causes you to gain that much life.)
//	{T}: Create a 4/4 white Angel creature token with flying. Activate only if you have at least 7 life more than your starting life total and only as a sorcery.
var SpeakerOfTheHeavens = newSpeakerOfTheHeavens

func newSpeakerOfTheHeavens() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Speaker of the Heavens",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.LifelinkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Create a 4/4 white Angel creature token with flying. Activate only if you have at least 7 life more than your starting life total and only as a sorcery.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Timing:          game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLifeAboveStarting, Op: compare.GreaterOrEqual, Value: 7}},
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(speakerOfTheHeavensToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, lifelink (Attacking doesn't cause this creature to tap. Damage dealt by this creature also causes you to gain that much life.)
			{T}: Create a 4/4 white Angel creature token with flying. Activate only if you have at least 7 life more than your starting life total and only as a sorcery.
		`,
		},
	}
}

var speakerOfTheHeavensToken = newSpeakerOfTheHeavensToken()

func newSpeakerOfTheHeavensToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}

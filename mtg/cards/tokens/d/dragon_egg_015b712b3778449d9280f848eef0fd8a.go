package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Dragon Egg
//
// Type: Token Creature — Dragon Egg
//
// Oracle text:
//   Defender
//   When this creature dies, create a 2/2 red Dragon creature token with flying and "{R}: This creature gets +1/+0 until end of turn."

// DragonEggToken015b712b3778449d9280f848eef0fd8a is the card definition for Dragon Egg.
var DragonEggToken015b712b3778449d9280f848eef0fd8a = newDragonEggToken015b712b3778449d9280f848eef0fd8a()

func newDragonEggToken015b712b3778449d9280f848eef0fd8a() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Dragon Egg",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon, types.Egg},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(dragonEggToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			When this creature dies, create a 2/2 red Dragon creature token with flying and "{R}: This creature gets +1/+0 until end of turn."
		`,
		},
	}
}

var dragonEggToken = newDragonEggToken()

func newDragonEggToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{R}: This creature gets +1/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

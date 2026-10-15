package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MonicaRambeau is the card definition for Monica Rambeau // Photon, Living Light.
//
// Type: Legendary Creature — Human Hero // Legendary Creature — Elemental Hero
// Face: Photon, Living Light — Legendary Creature — Elemental Hero ({2}{R}{W}{W})
//
// Oracle text:
//
//	Flying, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	{2}{R}{W}{W}: Transform Monica Rambeau. Activate only as a sorcery.
var MonicaRambeau = newMonicaRambeau

func newMonicaRambeau() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Monica Rambeau",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProwessStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{R}{W}{W}: Transform Monica Rambeau. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.R,
						cost.W,
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			{2}{R}{W}{W}: Transform Monica Rambeau. Activate only as a sorcery.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name: "Photon, Living Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elemental, types.Hero},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HexproofStaticBody,
				game.ProwessStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a noncreature spell, put a +1/+1 counter on each other creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying, hexproof, prowess
			Whenever you cast a noncreature spell, put a +1/+1 counter on each other creature you control.
		`,
		}),
	}
}

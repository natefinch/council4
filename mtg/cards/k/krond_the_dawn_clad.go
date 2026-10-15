package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KrondTheDawnClad is the card definition for Krond the Dawn-Clad.
//
// Type: Legendary Creature — Archon
// Cost: {G}{G}{G}{W}{W}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	Whenever Krond attacks, if it's enchanted, exile target permanent.
var KrondTheDawnClad = newKrondTheDawnClad

func newKrondTheDawnClad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Krond the Dawn-Clad",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
				cost.G,
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Archon},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Krond attacks, if it's enchanted, exile target permanent.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if it's enchanted",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if it's enchanted",
							Object:            opt.Val(game.EventPermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{MatchEnchanted: true}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
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
			Flying, vigilance
			Whenever Krond attacks, if it's enchanted, exile target permanent.
		`,
		},
	}
}

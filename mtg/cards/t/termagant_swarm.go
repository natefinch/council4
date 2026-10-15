package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TermagantSwarm is the card definition for Termagant Swarm.
//
// Type: Creature — Tyranid
// Cost: {X}{G}
//
// Oracle text:
//
//	Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
//	Death Frenzy — When this creature dies, create a number of 1/1 green Tyranid creature tokens equal to this creature's power.
var TermagantSwarm = newTermagantSwarm

func newTermagantSwarm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Termagant Swarm",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.RavenousDrawTriggeredAbility(),
				game.TriggeredAbility{
					Text: "Death Frenzy — When this creature dies, create a number of 1/1 green Tyranid creature tokens equal to this creature's power.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.SourcePermanentReference(),
											}),
											Source: game.TokenDef(termagantSwarmToken),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.RavenousEntersWithCountersReplacement(),
			},
			OracleText: `
			Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
			Death Frenzy — When this creature dies, create a number of 1/1 green Tyranid creature tokens equal to this creature's power.
		`,
		},
	}
}

var termagantSwarmToken = newTermagantSwarmToken()

func newTermagantSwarmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Tyranid",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

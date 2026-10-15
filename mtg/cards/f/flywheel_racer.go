package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlywheelRacer is the card definition for Flywheel Racer.
//
// Type: Artifact — Vehicle
// Cost: {2}
//
// Oracle text:
//
//	Vigilance
//	{T}: Add one mana of any color. Activate only if this permanent is a creature.
//	Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)
var FlywheelRacer = newFlywheelRacer

func newFlywheelRacer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Flywheel Racer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(1),
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add one mana of any color. Activate only if this permanent is a creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ActivationCondition: opt.Val(game.Condition{
						Text:              "only if this permanent is a creature",
						Object:            opt.Val(game.SourcePermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
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
			Vigilance
			{T}: Add one mana of any color. Activate only if this permanent is a creature.
			Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}

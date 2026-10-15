package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LeoninOfTheLostPride is the card definition for Leonin of the Lost Pride.
//
// Type: Creature — Cat Warrior
// Cost: {1}{W}
//
// Oracle text:
//
//	When this creature dies, exile target card from an opponent's graveyard.
var LeoninOfTheLostPride = newLeoninOfTheLostPride

func newLeoninOfTheLostPride() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leonin of the Lost Pride",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature dies, exile target card from an opponent's graveyard.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from an opponent's graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
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
			When this creature dies, exile target card from an opponent's graveyard.
		`,
		},
	}
}

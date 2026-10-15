package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForbiddenOrchard is the card definition for Forbidden Orchard.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any color.
//	Whenever you tap this land for mana, target opponent creates a 1/1 colorless Spirit creature token.
var ForbiddenOrchard = newForbiddenOrchard

func newForbiddenOrchard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Forbidden Orchard",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you tap this land for mana, target opponent creates a 1/1 colorless Spirit creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Source:               game.TriggerSourceSelf,
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
							RequireTappedForMana: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:    game.Fixed(1),
											Source:    game.TokenDef(forbiddenOrchardToken),
											Recipient: opt.Val(game.TargetPlayerReference(0)),
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
			{T}: Add one mana of any color.
			Whenever you tap this land for mana, target opponent creates a 1/1 colorless Spirit creature token.
		`,
		},
	}
}

var forbiddenOrchardToken = newForbiddenOrchardToken()

func newForbiddenOrchardToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

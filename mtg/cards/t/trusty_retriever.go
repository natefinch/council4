package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrustyRetriever is the card definition for Trusty Retriever.
//
// Type: Creature — Dog
// Cost: {3}{W}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Put a +1/+1 counter on this creature.
//	• Return target artifact or enchantment card from your graveyard to your hand.
var TrustyRetriever = newTrustyRetriever()

func newTrustyRetriever() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Trusty Retriever",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Put a +1/+1 counter on this creature.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Return target artifact or enchantment card from your graveyard to your hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact or enchantment card from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			When this creature enters, choose one —
			• Put a +1/+1 counter on this creature.
			• Return target artifact or enchantment card from your graveyard to your hand.
		`,
		},
	}
}

package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OltecArchaeologists is the card definition for Oltec Archaeologists.
//
// Type: Creature — Human Artificer Scout
// Cost: {4}{W}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Return target artifact card from your graveyard to your hand.
//	• Scry 3. (Look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var OltecArchaeologists = newOltecArchaeologists

func newOltecArchaeologists() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Oltec Archaeologists",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Artificer, types.Scout},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
								Text: "Return target artifact card from your graveyard to your hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact card from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
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
							game.Mode{
								Text: "Scry 3. (Look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)",
								Sequence: []game.Instruction{
									{
										Primitive: game.Scry{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
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
			• Return target artifact card from your graveyard to your hand.
			• Scry 3. (Look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}

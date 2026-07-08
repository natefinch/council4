package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KamiOfRestlessShadows is the card definition for Kami of Restless Shadows.
//
// Type: Creature — Spirit
// Cost: {4}{B}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Return up to one target Ninja or Rogue creature card from your graveyard to your hand.
//	• Put target creature card from your graveyard on top of your library.
var KamiOfRestlessShadows = newKamiOfRestlessShadows

func newKamiOfRestlessShadows() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Kami of Restless Shadows",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
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
								Text: "Return up to one target Ninja or Rogue creature card from your graveyard to your hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 0,
										MaxTargets: 1,
										Constraint: "up to one target Ninja or Rogue creature card from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Ninja"), types.Sub("Rogue")}, Controller: game.ControllerYou}),
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
								Text: "Put target creature card from your graveyard on top of your library.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card from your graveyard",
										Allow:      game.TargetAllowCard,
										TargetZone: zone.Graveyard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.MoveCard{
											Card:        game.CardReference{Kind: game.CardReferenceTarget},
											FromZone:    zone.Graveyard,
											Destination: zone.Library,
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
			• Return up to one target Ninja or Rogue creature card from your graveyard to your hand.
			• Put target creature card from your graveyard on top of your library.
		`,
		},
	}
}

package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SerraAvatar is the card definition for Serra Avatar.
//
// Type: Creature — Avatar
// Cost: {4}{W}{W}{W}
//
// Oracle text:
//
//	Serra Avatar's power and toughness are each equal to your life total.
//	When Serra Avatar is put into a graveyard from anywhere, shuffle it into its owner's library.
var SerraAvatar = newSerraAvatar

func newSerraAvatar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Serra Avatar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Avatar},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Serra Avatar is put into a graveyard from anywhere, shuffle it into its owner's library.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchToZone:     true,
							ToZone:          zone.Graveyard,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ShufflePermanentIntoLibrary{
											Object: game.EventPermanentReference(),
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
			Serra Avatar's power and toughness are each equal to your life total.
			When Serra Avatar is put into a graveyard from anywhere, shuffle it into its owner's library.
		`,
		},
	}
}

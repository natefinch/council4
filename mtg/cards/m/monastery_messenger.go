package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MonasteryMessenger is the card definition for Monastery Messenger.
//
// Type: Creature — Bird Scout
// Cost: {2/U}{2/R}{2/W}
//
// Oracle text:
//
//	Flying, vigilance
//	When this creature enters, put up to one target noncreature, nonland card from your graveyard on top of your library.
var MonasteryMessenger = newMonasteryMessenger()

func newMonasteryMessenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Monastery Messenger",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.U),
				cost.Twobrid(mana.R),
				cost.Twobrid(mana.W),
			}),
			Colors:    []color.Color{color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Scout},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target noncreature, nonland card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}, Controller: game.ControllerYou}),
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
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance
			When this creature enters, put up to one target noncreature, nonland card from your graveyard on top of your library.
		`,
		},
	}
}

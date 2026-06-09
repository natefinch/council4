package b

import (
	"github.com/natefinch/council4/mtg/cards/common"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BugenhagenWiseElder is the card definition for Bugenhagen, Wise Elder.
//
// Type: types.Legendary Creature — Human Shaman
// Cost: {1}{G}
//
// Oracle text:
//
//	Reach
//	At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.
//	{T}: Add one mana of any color.
var BugenhagenWiseElder = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bugenhagen, Wise Elder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Shaman},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: `
				Reach
				At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.
				{T}: Add one mana of any color.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.ReachStaticBody,
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepUpkeep,
				},
				InterveningIf: "if you control a creature with power 7 or greater",
				InterveningCondition: opt.Val(game.Condition{
					ControllerControls: game.PermanentFilter{
						Types: []types.Card{
							types.Creature,
						},
						Power: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 7,
						}),
					},
				}),
			},
			Content: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability(),
		},
	)

	card.ManaAbilities = []game.ManaAbility{common.TapForOneOfAny("bugenhagen-color")}

	return card
}()

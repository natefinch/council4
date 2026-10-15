package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EleshNornMotherOfMachines is the card definition for Elesh Norn, Mother of Machines.
//
// Type: Legendary Creature — Phyrexian Praetor
// Cost: {4}{W}
//
// Oracle text:
//
//	Vigilance
//	If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
//	Permanents entering don't cause abilities of permanents your opponents control to trigger.
var EleshNornMotherOfMachines = newEleshNornMotherOfMachines

func newEleshNornMotherOfMachines() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elesh Norn, Mother of Machines",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Praetor},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectAdditionalTriggerForEnteringPermanent,
							Duration: game.DurationPermanent,
						},
					},
				},
				game.StaticAbility{
					Text: "Permanents entering don't cause abilities of permanents your opponents control to trigger.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectSuppressOpponentEnteringTriggers,
							Duration: game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			Vigilance
			If a permanent entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
			Permanents entering don't cause abilities of permanents your opponents control to trigger.
		`,
		},
	}
}

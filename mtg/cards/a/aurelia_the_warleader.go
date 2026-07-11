package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AureliaTheWarleader is the card definition for Aurelia, the Warleader.
//
// Type: Legendary Creature — Angel
// Cost: {2}{R}{R}{W}{W}
//
// Oracle text:
//
//	Flying, vigilance, haste
//	Whenever Aurelia attacks for the first time each turn, untap all creatures you control. After this phase, there is an additional combat phase.
var AureliaTheWarleader = newAureliaTheWarleader

func newAureliaTheWarleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Aurelia, the Warleader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								},
							},
							{
								Primitive: game.AddExtraPhases{
									Combat: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, vigilance, haste
			Whenever Aurelia attacks for the first time each turn, untap all creatures you control. After this phase, there is an additional combat phase.
		`,
		},
	}
}

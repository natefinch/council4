package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RiveteersAscendancy is the card definition for Riveteers Ascendancy.
//
// Type: Enchantment
// Cost: {B}{R}{G}
//
// Oracle text:
//
//	Whenever you sacrifice a creature, you may return target creature card with lesser mana value from your graveyard to the battlefield tapped. Do this only once each turn.
var RiveteersAscendancy = newRiveteersAscendancy

func newRiveteersAscendancy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Riveteers Ascendancy",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green, color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentSacrificed,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Optional:           true,
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card with lesser mana value from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValueLessThanEventPermanent: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									EntryTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you sacrifice a creature, you may return target creature card with lesser mana value from your graveyard to the battlefield tapped. Do this only once each turn.
		`,
		},
	}
}

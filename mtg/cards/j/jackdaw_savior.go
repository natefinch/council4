package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JackdawSavior is the card definition for Jackdaw Savior.
//
// Type: Creature — Bird Cleric
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Whenever this creature or another creature you control with flying dies, return another target creature card with lesser mana value from your graveyard to the battlefield.
var JackdawSavior = newJackdawSavior

func newJackdawSavior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Jackdaw Savior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentDied,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature card with lesser mana value from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValueLessThanEventPermanent: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever this creature or another creature you control with flying dies, return another target creature card with lesser mana value from your graveyard to the battlefield.
		`,
		},
	}
}

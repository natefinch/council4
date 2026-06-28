package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TerisianMindbreaker is the card definition for Terisian Mindbreaker.
//
// Type: Artifact Creature — Juggernaut
// Cost: {7}
//
// Oracle text:
//
//	Whenever this creature attacks, defending player mills half their library, rounded up.
//	Unearth {1}{U}{U}{U} ({1}{U}{U}{U}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
var TerisianMindbreaker = newTerisianMindbreaker()

func newTerisianMindbreaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Terisian Mindbreaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.UnearthActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.U, cost.U}),
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
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:      game.DynamicAmountCountCardsInZone,
										Divisor:   2,
										RoundUp:   true,
										Player:    func() *game.PlayerReference { ref := game.DefendingPlayerReference(); return &ref }(),
										CardZone:  zone.Library,
										Selection: &game.Selection{},
									}),
									Player: game.DefendingPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, defending player mills half their library, rounded up.
			Unearth {1}{U}{U}{U} ({1}{U}{U}{U}: Return this card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step or if it would leave the battlefield. Unearth only as a sorcery.)
		`,
		},
	}
}

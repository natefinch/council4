package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GoblinDarkDwellers is the card definition for Goblin Dark-Dwellers.
//
// Type: Creature — Goblin
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	When this creature enters, you may cast target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
var GoblinDarkDwellers = newGoblinDarkDwellers

func newGoblinDarkDwellers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Dark-Dwellers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
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
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CastForFree{
									Player:            game.ControllerReference(),
									Zone:              zone.Graveyard,
									Card:              game.CardReference{Kind: game.CardReferenceTarget},
									ExileOnResolution: true,
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			When this creature enters, you may cast target instant or sorcery card with mana value 3 or less from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
		`,
		},
	}
}

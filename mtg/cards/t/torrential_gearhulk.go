package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TorrentialGearhulk is the card definition for Torrential Gearhulk.
//
// Type: Artifact Creature — Construct
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flash
//	When this creature enters, you may cast target instant card from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
var TorrentialGearhulk = newTorrentialGearhulk()

func newTorrentialGearhulk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Torrential Gearhulk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
								Constraint: "target instant card from your graveyard without paying its mana cost",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant}, Controller: game.ControllerYou}),
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
			Flash
			When this creature enters, you may cast target instant card from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
		`,
		},
	}
}

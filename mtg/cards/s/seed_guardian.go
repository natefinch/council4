package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SeedGuardian is the card definition for Seed Guardian.
//
// Type: Creature — Elemental
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Reach
//	When this creature dies, create an X/X green Elemental creature token, where X is the number of creature cards in your graveyard.
var SeedGuardian = newSeedGuardian

func newSeedGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Seed Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(seedGuardianToken),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Graveyard,
										Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Graveyard,
										Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
									})),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			When this creature dies, create an X/X green Elemental creature token, where X is the number of creature cards in your graveyard.
		`,
		},
	}
}

var seedGuardianToken = newSeedGuardianToken()

func newSeedGuardianToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Elemental",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Elemental},
		},
	}
}

package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BladewingDeathlessTyrant is the card definition for Bladewing, Deathless Tyrant.
//
// Type: Legendary Creature — Dragon Skeleton
// Cost: {5}{B}{R}
//
// Oracle text:
//
//	Flying, haste
//	Whenever Bladewing deals combat damage to a player or planeswalker, for each creature card in your graveyard, create a 2/2 black Zombie Knight creature token with menace.
var BladewingDeathlessTyrant = newBladewingDeathlessTyrant

func newBladewingDeathlessTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Bladewing, Deathless Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon, types.Skeleton},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Bladewing deals combat damage to a player or planeswalker, for each creature card in your graveyard, create a 2/2 black Zombie Knight creature token with menace.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventDamageDealt,
							Source:                   game.TriggerSourceSelf,
							Subject:                  game.TriggerSubjectDamageSource,
							StackObjectKind:          game.StackSpell,
							DamageRecipient:          game.DamageRecipientPlayer | game.DamageRecipientPermanent,
							DamageRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
							RequireCombatDamage:      true,
							CounterKind:              counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountCardsInZone,
												Multiplier: 1,
												Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
												CardZone:   zone.Graveyard,
												Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
											}),
											Source: game.TokenDef(bladewingDeathlessTyrantToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying, haste
			Whenever Bladewing deals combat damage to a player or planeswalker, for each creature card in your graveyard, create a 2/2 black Zombie Knight creature token with menace.
		`,
		},
	}
}

var bladewingDeathlessTyrantToken = newBladewingDeathlessTyrantToken()

func newBladewingDeathlessTyrantToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie Knight",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}

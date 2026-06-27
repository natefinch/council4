package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpawnGangCommander is the card definition for Spawn-Gang Commander.
//
// Type: Creature — Eldrazi Goblin
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	When you cast this spell, create three 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
//	{1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any target.
var SpawnGangCommander = newSpawnGangCommander()

func newSpawnGangCommander() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Spawn-Gang Commander",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Goblin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any target.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.C}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice an Eldrazi",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Eldrazi},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(spawnGangCommanderToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			When you cast this spell, create three 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
			{1}{C}, Sacrifice an Eldrazi: This creature deals 2 damage to any target.
		`,
		},
	}
}

var spawnGangCommanderToken = newSpawnGangCommanderToken()

func newSpawnGangCommanderToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

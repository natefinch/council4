package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	tokenc "github.com/natefinch/council4/mtg/cards/tokens/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	defilerOfSoulsOracleID  = "34894b92-7ebd-49be-9a67-7dc6428aaab9"
	woebringerDemonOracleID = "21ea9501-9a0f-43a3-a2cc-e9648325c45e"
	wasitoraOracleID        = "13fc4168-fcc2-4011-804f-211d5d86b7dd"
)

const (
	woebringerSacrificeResult game.ResultKey = "21ea9501-9a0f-43a3-a2cc-e9648325c45e/sacrifice"
	wasitoraSacrificeResult   game.ResultKey = "13fc4168-fcc2-4011-804f-211d5d86b7dd/sacrifice"
)

var DefilerOfSouls = cardset.Entry{
	OracleID:     defilerOfSoulsOracleID,
	Name:         "Defiler of Souls",
	New:          newDefilerOfSouls,
	Verification: "mtg/rules/upkeep_sacrifice_eddicts_test.go: TestDefilerOfSoulsEachPlayerUpkeepSacrificesMonocoloredCreature",
}

var WoebringerDemon = cardset.Entry{
	OracleID:     woebringerDemonOracleID,
	Name:         "Woebringer Demon",
	New:          newWoebringerDemon,
	Verification: "mtg/rules/upkeep_sacrifice_eddicts_test.go: TestWoebringerDemonUpkeepSacrificeAndSelfSacrificeFallback",
}

var WasitoraNekoruQueen = cardset.Entry{
	OracleID:     wasitoraOracleID,
	Name:         "Wasitora, Nekoru Queen",
	New:          newWasitoraNekoruQueen,
	Verification: "mtg/rules/upkeep_sacrifice_eddicts_test.go: TestWasitoraCombatDamageSacrificeAndCatDragonFallback",
}

func newDefilerOfSouls() *game.CardDef {
	text := "Flying\nAt the beginning of each player's upkeep, that player sacrifices a monocolored creature of their choice."
	triggerText := "At the beginning of each player's upkeep, that player sacrifices a monocolored creature of their choice."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Defiler of Souls", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B, cost.R}),
			Colors: []color.Color{color.Black, color.Red}, Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Demon},
			Power: opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 5}), OracleText: text,
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: triggerText,
				Trigger: game.TriggerCondition{
					Type:    game.TriggerAt,
					Pattern: game.TriggerPattern{Event: game.EventBeginningOfStep, Step: game.StepUpkeep},
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.SacrificePermanents{
						Player: game.EventPlayerReference(), Amount: game.Fixed(1),
						Selection: monocoloredCreatureSelection(),
					},
				}}}.Ability(),
			}},
		},
	}
}

func newWoebringerDemon() *game.CardDef {
	text := "Flying\nAt the beginning of each player's upkeep, that player sacrifices a creature of their choice. If the player can't, sacrifice this creature."
	triggerText := "At the beginning of each player's upkeep, that player sacrifices a creature of their choice. If the player can't, sacrifice this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Woebringer Demon", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
			Colors: []color.Color{color.Black}, Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Demon},
			Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4}), OracleText: text,
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: triggerText,
				Trigger: game.TriggerCondition{
					Type:    game.TriggerAt,
					Pattern: game.TriggerPattern{Event: game.EventBeginningOfStep, Step: game.StepUpkeep},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Player: game.EventPlayerReference(), Amount: game.Fixed(1),
							Selection: creatureSelection(),
						},
						PublishResult: woebringerSacrificeResult,
					},
					{
						Primitive: game.Sacrifice{Object: game.SourcePermanentReference()},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key: woebringerSacrificeResult, Succeeded: game.TriFalse,
						}),
					},
				}}.Ability(),
			}},
		},
	}
}

func newWasitoraNekoruQueen() *game.CardDef {
	text := "Flying, trample\nWhenever Wasitora deals combat damage to a player, that player sacrifices a creature of their choice. If the player can't, you create a 3/3 black, red, and green Cat Dragon creature token with flying."
	triggerText := "Whenever Wasitora deals combat damage to a player, that player sacrifices a creature of their choice. If the player can't, you create a 3/3 black, red, and green Cat Dragon creature token with flying."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Wasitora, Nekoru Queen", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.R, cost.G}),
			Colors: []color.Color{color.Black, color.Red, color.Green}, Supertypes: []types.Super{types.Legendary},
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Cat, types.Dragon},
			Power: opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 4}), OracleText: text,
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.TrampleStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: triggerText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event: game.EventDamageDealt, Controller: game.TriggerControllerYou,
						Source: game.TriggerSourceSelf, Subject: game.TriggerSubjectDamageSource,
						DamageRecipient: game.DamageRecipientPlayer, RequireCombatDamage: true,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Player: game.EventPlayerReference(), Amount: game.Fixed(1), Selection: creatureSelection(),
						},
						PublishResult: wasitoraSacrificeResult,
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1), Source: game.TokenDef(tokenc.CatDragonTokenafdf740007234f0db3d846ad3a698908),
							Recipient: opt.Val(game.ControllerReference()),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key: wasitoraSacrificeResult, Succeeded: game.TriFalse,
						}),
					},
				}}.Ability(),
			}},
		},
	}
}

func creatureSelection() game.Selection {
	return game.Selection{RequiredTypes: []types.Card{types.Creature}}
}

func monocoloredCreatureSelection() game.Selection {
	colors := []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green}
	alternatives := make([]game.Selection, 0, len(colors))
	for _, selected := range colors {
		excluded := make([]color.Color, 0, len(colors)-1)
		for _, candidate := range colors {
			if candidate != selected {
				excluded = append(excluded, candidate)
			}
		}
		alternatives = append(alternatives, game.Selection{
			ColorsAny: []color.Color{selected}, ExcludedColors: excluded,
		})
	}
	return game.Selection{RequiredTypes: []types.Card{types.Creature}, AnyOf: alternatives}
}

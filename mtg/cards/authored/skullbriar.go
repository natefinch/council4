package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Skullbriar uses ordinary shared operations and a typed zone-transition policy.
var Skullbriar = cardset.Entry{
	OracleID: "088ac883-6f44-4046-bf19-81c7fd952e55", Name: "Skullbriar, the Walking Grave", New: newSkullbriar,
	Verification: "mtg/rules/skullbriar_test.go: TestSkullbriarProductionCombatTriggerAndSimulator",
}

func newSkullbriar() *game.CardDef {
	const trigger = "Whenever Skullbriar deals combat damage to a player, put a +1/+1 counter on it."
	const retention = "Counters remain on Skullbriar as it moves to any zone other than a player's hand or library."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Skullbriar, the Walking Grave", ManaCost: opt.Val(cost.Mana{cost.B, cost.G}),
			Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Zombie, types.Elemental}, Colors: []color.Color{color.Black, color.Green},
			Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			OracleText: "Haste\n" + trigger + "\n" + retention,
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				{Text: retention, ZoneTransitionPolicy: game.ZoneTransitionRetainCountersExceptHandLibrary},
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: trigger,
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
					Event: game.EventDamageDealt, Source: game.TriggerSourceSelf, Subject: game.TriggerSubjectDamageSource,
					RequireCombatDamage: true, DamageRecipient: game.DamageRecipientPlayer,
				}},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddCounter{
					Object: game.SourcePermanentReference(), CounterKind: counter.PlusOnePlusOne, Amount: game.Fixed(1),
				}}}}.Ability(),
			}},
		},
	}
}

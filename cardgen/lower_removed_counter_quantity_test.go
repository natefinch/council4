package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

const ashlingRemovedCounterText = "{1}{R}: Put a +1/+1 counter on Ashling. If this is the third time this ability has resolved this turn, remove all +1/+1 counters from Ashling, and it deals that much damage to each creature and each player."

func TestLowerRemovedCounterQuantityConsumers(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Remove two charge counters from this artifact. You gain that much life.",
		"Remove all charge counters from this artifact. Draw that many cards.",
		"Remove a charge counter from this artifact. Untap this artifact. You gain life equal to the number of counters removed this way.",
		"Remove two +1/+1 counters from target creature. Draw that many cards.",
		"Remove a +1/+1 counter from each creature you control. You gain life equal to the number of counters removed this way.",
		"Quantity deals 7 damage to any target. Remove two charge counters from this artifact. You gain that much life.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Quantity", Layout: "normal", TypeLine: "Artifact", OracleText: "{1}: " + text}
			face := lowerSingleFace(t, card)
			sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
			last := len(sequence) - 1
			producer := 0
			if sequence[0].Primitive.Kind() == game.PrimitiveDamage {
				producer = 1
			}
			if sequence[producer].PublishResult != removedCounterQuantityKey(producer+1) ||
				!sequence[last].ResultGate.Exists || !sequence[last].ResultGate.Val.AmountAvailable ||
				sequence[last].ResultGate.Val.Key != sequence[producer].PublishResult {
				t.Fatalf("sequence = %#v", sequence)
			}
		})
	}
}

func TestLowerAshlingActualQuantityComposition(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{Name: "Ashling", Layout: "normal", TypeLine: "Legendary Creature - Elemental Shaman", OracleText: ashlingRemovedCounterText}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].CountsResolutionsThisTurn = true",
		`ActivatedAbilities[0].Content.Modes[0].Sequence[1].PublishResult = "removed-counter-quantity-2"`,
		"ActivatedAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn = 3",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[2].ResultGate.Val.AmountAvailable = true",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[3].ResultGate.Val.AmountAvailable = true",
	)
	face := lowerSingleFace(t, card)
	sequence := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 4 {
		t.Fatalf("instructions=%d, want placement/removal/two damage groups", len(sequence))
	}
	remove, ok := sequence[1].Primitive.(game.RemoveCounter)
	if !ok || remove.AllKinds || remove.CounterKind != counter.PlusOnePlusOne ||
		!remove.Amount.IsDynamic() || remove.Amount.DynamicAmount().Val.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("remove = %#v", sequence[1].Primitive)
	}
	for _, i := range []int{2, 3} {
		assertConditionConsumer(t, sequence[1], sequence[i], false)
		damage, ok := sequence[i].Primitive.(game.Damage)
		if !ok {
			t.Fatalf("primitive = %T", sequence[i].Primitive)
		}
		dynamic := damage.Amount.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountPreviousEffectResult ||
			dynamic.Val.ResultKey != sequence[1].PublishResult {
			t.Fatalf("damage = %#v", damage)
		}
	}
}

func TestRemovedCounterQuantityNearMisses(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"You gain life equal to the number of counters removed this way. Remove a charge counter from this artifact.",
		"Remove all charge counters from each artifact you control. Draw that many cards.",
		"Remove two counters from this artifact. Draw that many cards.",
		"Remove a charge counter from this artifact. Remove a charge counter from this artifact. You gain life equal to the number of counters removed this way.",
		"Remove a charge counter from this artifact. Draw a card. You gain that much life.",
		"Remove all charge counters from this artifact except one. Draw that many cards.",
		"Choose two \u2014\n\u2022 Remove all charge counters from this artifact. Draw that many cards.\n\u2022 Remove all charge counters from this artifact. You gain that much life.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Quantity Near Miss", Layout: "normal", TypeLine: "Artifact", OracleText: "{1}: " + text})
	}
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Quantity Near Miss", Layout: "normal", TypeLine: "Artifact",
		OracleText: "Whenever you gain life, remove a charge counter from this artifact. Draw a card. You gain that much life.",
	})
}

func TestRemovedCounterQuantityLinkIsTextBlind(t *testing.T) {
	t.Parallel()
	effects := []compiler.CompiledEffect{
		{ClauseID: 17, Kind: compiler.EffectRemoveCounter, Text: "diagnostic-only"},
		{ClauseID: 92, Amount: compiler.CompiledAmount{DynamicKind: compiler.DynamicAmountRemovedCounterCount, ProducerClauseID: 17}},
	}
	sequence := []game.Instruction{
		{Primitive: game.RemoveCounter{Object: game.SourcePermanentReference(), Amount: game.Fixed(2)}},
		{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountPreviousEffectResult, ResultKey: removedCounterQuantityKey(17),
		})}},
	}
	if reason := linkRemovedCounterQuantities(effects, [][2]int{{0, 1}, {1, 2}}, sequence); reason != "" ||
		sequence[0].PublishResult != removedCounterQuantityKey(17) ||
		!sequence[1].ResultGate.Exists || sequence[1].ResultGate.Val.Key != removedCounterQuantityKey(17) ||
		!sequence[1].ResultGate.Val.AmountAvailable {
		t.Fatalf("reason=%q sequence=%#v", reason, sequence)
	}
}

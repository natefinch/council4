package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestRepeatTargetProjectionPreservesNestedOperands(t *testing.T) {
	t.Parallel()
	inner := game.RepeatProcess{
		Times: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(0)}),
		Body: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Explore{Creature: game.TargetPermanentReference(0)},
				Optional: true, PublishOptionalDecision: "choice",
				LocalProducts: game.LocalProducts{Results: []game.ResultKey{"receipt"}}},
			{Primitive: game.Tap{Object: game.TargetPermanentReference(0)},
				OptionalDecisionGate: "choice", PublishResult: "receipt"},
			{Primitive: game.Draw{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)}},
		}}.Ability(),
	}

	outer := game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{
		Sequence: []game.Instruction{{Primitive: inner}},
	}.Ability()}
	transformed, ok := rebaseTargetedPrimitive(outer, 2, 1)
	if !ok {
		t.Fatal("nested repeat could not be projected")
	}
	projectedOuter := captureTestPrimitive[game.RepeatProcess](t, transformed)
	got := captureTestPrimitive[game.RepeatProcess](t, projectedOuter.Body.Modes[0].Sequence[0].Primitive)
	sequence := got.Body.Modes[0].Sequence
	if captureTestPrimitive[game.Explore](t, sequence[0].Primitive).Creature != game.TargetPermanentReference(2) ||
		captureTestPrimitive[game.Tap](t, sequence[1].Primitive).Object != game.TargetPermanentReference(2) ||
		captureTestPrimitive[game.Draw](t, sequence[2].Primitive).Player != game.TargetPlayerReference(2) ||
		got.Times.DynamicAmount().Val.Object != game.TargetPermanentReference(2) {
		t.Fatal("one nested target-bearing primitive or repeat quantity retained a local index")
	}
	if sequence[0].PublishOptionalDecision != "choice" || !sequence[0].Optional ||
		sequence[1].OptionalDecisionGate != "choice" || sequence[1].PublishResult != "receipt" ||
		!reflect.DeepEqual(sequence[0].LocalProducts, inner.Body.Modes[0].Sequence[0].LocalProducts) {
		t.Fatal("target projection changed optionality, receipts, or local product ownership")
	}
	if captureTestPrimitive[game.Explore](t, inner.Body.Modes[0].Sequence[0].Primitive).Creature != game.TargetPermanentReference(0) {
		t.Fatal("target projection mutated the original body")
	}
	if remapTargetedSequence([]game.Instruction{{Primitive: outer}}, nil) {
		t.Fatal("missing target map silently retained nested local references")
	}
}

func TestRepeatProjectionTransformsEnvelopeAndAllQuantityOperands(t *testing.T) {
	t.Parallel()
	player := game.TargetPlayerReference(0)
	for _, amount := range []game.DynamicAmount{
		{Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Multiplier: 1},
		{Kind: game.DynamicAmountCountSelector, Group: game.PlayerControlledGroup(player, game.Selection{}), Multiplier: 1},
		{Kind: game.DynamicAmountMaxOf, Operands: []game.DynamicAmount{
			{Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Multiplier: 1},
		}},
	} {
		repeat := game.RepeatProcess{Times: game.Dynamic(amount), Body: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)},
		}}}.Ability()}
		got, ok := rebaseTargetedPrimitive(repeat, 2, 0)
		if !ok {
			t.Fatal("modeled repeat quantity could not be projected")
		}
		dynamic := captureTestPrimitive[game.RepeatProcess](t, got).Times.DynamicAmount().Val
		if dynamic.Player != nil && *dynamic.Player != game.TargetPlayerReference(2) {
			t.Fatal("repeat count retained a local player reference")
		}
		if anchor, exists := dynamic.Group.PlayerAnchor(); exists && anchor != game.TargetPlayerReference(2) {
			t.Fatal("repeat count retained a local group anchor")
		}
		if len(dynamic.Operands) != 0 && *dynamic.Operands[0].Player != game.TargetPlayerReference(2) {
			t.Fatal("repeat count retained a local nested operand")
		}
		if amount.Player != nil && *amount.Player != game.TargetPlayerReference(0) ||
			len(amount.Operands) != 0 && *amount.Operands[0].Player != game.TargetPlayerReference(0) {
			t.Fatal("projection mutated a shared quantity pointer or operand")
		}
		if anchor, exists := amount.Group.PlayerAnchor(); exists && anchor != game.TargetPlayerReference(0) {
			t.Fatal("projection mutated the original group")
		}
		if _, ok := rebaseTargetedPrimitive(repeat, 3, 0); !ok {
			t.Fatal("the shared definition could not build a second independent remap")
		}
		if dynamic.Player != nil && *dynamic.Player != game.TargetPlayerReference(2) ||
			len(dynamic.Operands) != 0 && *dynamic.Operands[0].Player != game.TargetPlayerReference(2) {
			t.Fatal("building a second remap mutated the first")
		}
		if remapTargetedSequence([]game.Instruction{{Primitive: repeat}}, nil) {
			t.Fatal("missing target map accepted quantity-only target references")
		}
	}
	condition := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
		Object:        opt.Val(game.TargetPermanentReference(0)),
		ObjectMatches: opt.Val(game.Selection{Keyword: game.Flying}),
	})})
	repeat := game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Explore{Creature: game.TargetPermanentReference(0)},
		Condition: condition, Optional: true, OptionalActor: opt.Val(game.TargetPlayerReference(0)),
	}}}.Ability()}
	got, ok := rebaseTargetedPrimitive(repeat, 2, 0)
	if !ok {
		t.Fatal("modeled repeat envelope could not be projected")
	}
	instruction := captureTestPrimitive[game.RepeatProcess](t, got).Body.Modes[0].Sequence[0]
	if instruction.Condition.Val.Condition.Val.Object.Val != game.TargetPermanentReference(2) ||
		instruction.OptionalActor.Val != game.TargetPlayerReference(2) {
		t.Fatal("repeat envelope retained a clause-local target")
	}
	if condition.Val.Condition.Val.Object.Val != game.TargetPermanentReference(0) ||
		repeat.Body.Modes[0].Sequence[0].OptionalActor.Val != game.TargetPlayerReference(0) {
		t.Fatal("projection mutated a shared condition or optional actor")
	}
	repeat.Body.Modes[0].Sequence[0].Primitive = game.Explore{Creature: game.SourcePermanentReference()}
	repeat.Body.Modes[0].Sequence[0].OptionalActor = opt.V[game.PlayerReference]{}
	if remapTargetedSequence([]game.Instruction{{Primitive: repeat}}, nil) {
		t.Fatal("missing target map accepted an envelope-only target reference")
	}
}

func TestRepeatProjectionRefusesIndependentTargetScopes(t *testing.T) {
	t.Parallel()
	for _, body := range []game.AbilityContent{
		{},
		{Modes: []game.Mode{{}, {}}},
		{SharedTargets: []game.TargetSpec{{}}, Modes: []game.Mode{{}}},
		game.Mode{Targets: []game.TargetSpec{{}}, Sequence: []game.Instruction{
			{Primitive: game.Explore{Creature: game.TargetPermanentReference(0)}},
		}}.Ability(),
	} {
		if _, ok := rebaseTargetedPrimitive(game.RepeatProcess{Times: game.Fixed(2), Body: body}, 1, 1); ok {
			t.Fatalf("independently scoped repeat body accepted: %+v", body)
		}
	}
}

func TestRepeatProjectionPreservesFrozenReferenceDomains(t *testing.T) {
	t.Parallel()
	for _, object := range []game.ObjectReference{
		game.SourcePermanentReference(), game.EventPermanentReference(),
		game.LinkedObjectReference("prior"), game.CapturedObjectReference(),
		game.CapturedTargetStackObjectReference(1),
	} {
		player := game.CapturedTargetControllerReference(1)
		amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountObjectPower, Object: object, Player: &player})
		repeat := game.RepeatProcess{Times: amount, Body: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Explore{Creature: game.EventPermanentReference()},
		}}}.Ability()}
		projected, ok := rebaseTargetedPrimitive(repeat, 3, 2)
		if !ok {
			t.Fatal("frozen-domain references should not require a chosen-target map")
		}
		got := captureTestPrimitive[game.RepeatProcess](t, projected).Times.DynamicAmount().Val
		if got.Object != object || *got.Player != player || got.Player == &player {
			t.Fatal("projection rebound a frozen domain or retained the old mutable player pointer")
		}
	}
	for _, amount := range []game.DynamicAmount{
		{Kind: game.DynamicAmountTargetPower},
		{Kind: game.DynamicAmountObjectPower, Object: game.AllTargetPermanentsReference(0)},
		{Kind: game.DynamicAmountCountSelector, Group: game.PlayerGroupControlledGroup(game.TargetedPlayersReference(), game.Selection{})},
	} {
		if _, ok := transformQuantity(game.Dynamic(amount), rebaseTransform(2, 1)); ok {
			t.Fatal("implicit or whole-target-group quantity bypassed its target-domain proof")
		}
	}
}

func TestRepeatedExploreAfterPlayerOrCardTarget(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{
		"Target player draws a card.",
		"Exile target card from a graveyard.",
		"Counter target creature spell.",
		"Counter target spell an opponent controls.",
	} {
		for _, count := range []string{"two", "X"} {
			t.Run(prefix+"/"+count, func(t *testing.T) {
				t.Parallel()
				card := &ScryfallCard{Name: "Test Survey", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{X}{G}",
					OracleText: prefix + " Target creature explores " + count + " times."}
				assertCardPaths(t, card,
					"SpellAbility.Val.Modes[0].Targets[1]",
					"Sequence[1].Primitive.(game.RepeatProcess).Body.Modes[0].Sequence[0].Primitive.(game.Explore).Creature.targetIndex = 1",
				)
				if count == "X" {
					assertCardPaths(t, card, "Sequence[1].Primitive.(game.RepeatProcess).Times.dynamic.Kind = game.DynamicAmountX")
				} else {
					assertCardPaths(t, card, "Sequence[1].Primitive.(game.RepeatProcess).Times.fixed = 2")
				}
			})
		}
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledTargetedActorScope(t *testing.T) *targetedPlayerActorProjection {
	t.Helper()
	compilation, diagnostics := compileTestOracle("Any number of target players each mill two cards.",
		parser.Context{InstantOrSorcery: true}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	body := compilation.Abilities[0].Content
	scope := targetedPlayerActorScope(body, body.Effects[0])
	if scope == nil {
		t.Fatal("canonical complete-body target set did not issue an actor projection")
	}
	return scope
}

func TestTargetedPlayerActorVisitedQuantityFamilies(t *testing.T) {
	t.Parallel()
	group := game.TargetedPlayersReference()
	for _, test := range []visitedQuantityCase{
		{"draw", func(q game.Quantity) game.Primitive { return game.Draw{PlayerGroup: group, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Draw](t, p).Amount
			}},
		{"discard", func(q game.Quantity) game.Primitive { return game.Discard{PlayerGroup: group, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.Discard](t, p).Amount
			}},
		{"mill", func(q game.Quantity) game.Primitive {
			return game.MoveTopOfLibrary{PlayerGroup: group, Amount: q, Destination: zone.Graveyard}
		}, func(t *testing.T, p game.Primitive) game.Quantity {
			return captureTestPrimitive[game.MoveTopOfLibrary](t, p).Amount
		}},
		{"gain", func(q game.Quantity) game.Primitive { return game.GainLife{PlayerGroup: group, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.GainLife](t, p).Amount
			}},
		{"lose", func(q game.Quantity) game.Primitive { return game.LoseLife{PlayerGroup: group, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.LoseLife](t, p).Amount
			}},
		{"sacrifice", func(q game.Quantity) game.Primitive { return game.SacrificePermanents{PlayerGroup: group, Amount: q} },
			func(t *testing.T, p game.Primitive) game.Quantity {
				return captureTestPrimitive[game.SacrificePermanents](t, p).Amount
			}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			scope := compiledTargetedActorScope(t)
			fixed := test.build(game.Fixed(2))
			if _, ok := rebaseTargetedPrimitive(fixed, 0, 0); ok {
				t.Fatal("low-level identity mapping without complete-body proof accepted an unindexed actor")
			}
			projected, ok := transformPrimitiveTargetIndicesWithActorScope(fixed, rebaseTransform(2, 1), scope)
			if !ok || !targetedPlayerActor(projected) || test.amount(t, projected).Value() != 2 {
				t.Fatal("owned whole actor set or fixed scalar changed under nonplayer-prefix projection")
			}
			player := game.TargetPlayerReference(0)
			scalar := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountCountCardsInZone,
				Player: &player, CardZone: zone.Hand, Selection: new(game.Selection), Multiplier: 1})
			projected, ok = transformPrimitiveTargetIndicesWithActorScope(test.build(scalar), rebaseTransform(2, 1), scope)
			if !ok || *test.amount(t, projected).DynamicAmount().Val.Player != game.TargetPlayerReference(2) ||
				*scalar.DynamicAmount().Val.Player != player {
				t.Fatal("actor allowance bypassed scalar target projection or mutated its original operand")
			}
			unscoped := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountCountSelector,
				Group: game.PlayerGroupControlledGroup(group, game.Selection{}), Multiplier: 1})
			if _, ok := transformPrimitiveTargetIndicesWithActorScope(test.build(unscoped), rebaseTransform(2, 1), scope); ok {
				t.Fatal("actor allowance was misused as a whole-target group-quantity allowance")
			}
			scope.order.Start++
			if _, ok := transformPrimitiveTargetIndicesWithActorScope(fixed, rebaseTransform(0, 0), scope); ok {
				t.Fatal("drifted actor declaration identity was accepted")
			}
		})
	}
}

func TestCourtTargetedPlayerActorWholeBodyAndPartialRefusals(t *testing.T) {
	t.Parallel()
	const body = "Any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead."
	for _, prefix := range []string{"", "Counter target spell. "} {
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Actor Set Probe", Layout: "normal", TypeLine: "Sorcery",
				OracleText: prefix + body + " You gain 1 life."}
			assertCardPaths(t, card,
				"PlayerGroup.Kind = game.PlayerGroupReferenceTargetedPlayers",
				"Amount.fixed = 2", "Amount.fixed = 10",
				"PublishCondition = \"condition-", "ConditionGateNegate = true",
				"Primitive.(game.GainLife).Amount.fixed = 1",
			)
		})
	}

	for _, text := range []string{
		"Target player draws a card. " + body,
		body + " Target player draws a card.",
		body + " Up to two target players each draw a card.",
		"Exile target card from a graveyard. " + body,
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Partial Actor Set Probe", Layout: "normal",
				TypeLine: "Sorcery", OracleText: text}, "actor target-set scope")
		})
	}
}

func TestTargetedPlayerActorProofIsNotAnOrdinarySubject(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileTestOracle("Any number of target players each mill two cards. Each of those players mills ten cards.",
		parser.Context{InstantOrSorcery: true}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	body := compilation.Abilities[0].Content
	reference := body.Effects[1].References[0]
	if !body.OwnsSubject(reference) || reference.SubjectDomain() != compiler.ReferenceSubjectPlayerGroup ||
		reference.Binding != compiler.ReferenceBindingAmbiguous {
		t.Fatal("canonical group-actor issuance changed binding or did not retain exact private body ownership")
	}
	bindings := referenceLoweringContext{AllowSource: true, AllowTarget: true, AllowEvent: true, PriorInstruction: 0, PriorLinkedKey: "prior"}
	if _, ok := lowerObjectReference(reference, bindings); ok {
		t.Fatal("group actor proof became a general singleton object")
	}
	if _, ok := lowerPlayerReference(reference, bindings); ok {
		t.Fatal("group actor proof became a general singleton player or payment actor")
	}
	if _, ok := lowerCardReference(reference, bindings); ok {
		t.Fatal("group actor proof became a card or prior card result")
	}
}

func TestTargetedPlayerActorDoesNotRetagCardReadbacks(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Exile target card from a graveyard. Any number of target players each mill two cards. Each of those players mills ten cards.",
		"Exile target card from a graveyard. Any number of target players each mill two cards. Return those cards to their owners' hands.",
	} {
		compilation, diagnostics := compileTestOracle(text, parser.Context{InstantOrSorcery: true}, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		body := compilation.Abilities[0].Content
		effect := body.Effects[len(body.Effects)-1]
		if targetedPlayerActorScope(body, effect) != nil {
			t.Fatal("generic card-result readback acquired player actor-set authority")
		}
		assertCardUnsupported(t, &ScryfallCard{Name: "Ambiguous Card Actor Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
	}
}

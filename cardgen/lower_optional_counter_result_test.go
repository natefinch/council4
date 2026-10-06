package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestOptionalCounterResultBoundary(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		paths      []string
		absent     []string
	}{
		{"single optional counter", "You may counter target spell. If that spell is countered this way, you gain 2 life. You gain 1 life.",
			[]string{"Sequence[0].Optional = true", "Sequence[0].PublishResult = \"if-you-do\"",
				"Sequence[1].ResultGate.Val.Succeeded = game.TriTrue"},
			[]string{"Sequence[1].Optional = true", "Sequence[2].ResultGate.Exists = true"}},
		{"independent decisions around tax", "You may discard a card and gain 3 life. Counter target spell unless its controller pays {3}. If that spell is countered this way, you gain 2 life. You gain 1 life. You may discard a card and gain 5 life.",
			[]string{"Sequence[0].PublishOptionalDecision", "Sequence[1].OptionalDecisionGate",
				"Sequence[2].Primitive.(game.Pay)", "Sequence[3].Primitive.(game.CounterObject)",
				"Sequence[3].PublishResult = \"if-you-do\"", "Sequence[4].ResultGate.Val.Succeeded = game.TriTrue",
				"Sequence[6].PublishOptionalDecision", "Sequence[7].OptionalDecisionGate"},
			[]string{"Sequence[2].Optional = true", "Sequence[3].Optional = true",
				"Sequence[3].OptionalDecisionGate", "Sequence[5].ResultGate.Exists = true"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Optional Counter Boundary", Layout: "normal", TypeLine: "Instant", OracleText: tt.text}
			assertCardPaths(t, card, tt.paths...)
			assertCardPathsAbsent(t, card, tt.absent...)
		})
	}
	for _, text := range []string{
		"You may counter target spell and gain 2 life. If that spell is countered this way, draw a card.",
		"You may counter target spell unless its controller pays {3}. If that spell is countered this way, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Optional Counter Refusal", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
		})
	}
}

func TestOptionalCounterTypedResultOwnership(t *testing.T) {
	for _, compound := range []bool{false, true} {
		document, diagnostics := parser.Parse(
			"You may counter target spell. You gain 1 life. If that spell is countered this way, you gain 2 life.",
			parser.Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compiled, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		content := compiled.Abilities[0].Content
		if compound {
			content.Effects[0].OptionalActionClauseIDs = []int{content.Effects[0].ClauseID, content.Effects[1].ClauseID}
			content.Effects[0].Context = parser.EffectContextController
			plan, valid, handled := planScopedResultFlow(content)
			if valid || !handled || plan.failureCategory != "structural — whole optional action outcome not modeled" {
				t.Fatalf("compound counter outcome: valid=%v handled=%v reason=%q", valid, handled, plan.failureCategory)
			}
		} else {
			content.Conditions[0].Ownership.ResultSubjectTargetOccurrence = 999
			plan, valid, handled := planScopedResultFlow(content)
			if valid || !handled || plan.failureCategory != counterOwnershipCategory {
				t.Fatalf("wrong counter occurrence: valid=%v handled=%v reason=%q", valid, handled, plan.failureCategory)
			}
		}
	}
}

func TestOptionalCounterPublicationSelection(t *testing.T) {
	taxSequence := func() []game.Instruction {
		return []game.Instruction{
			{Primitive: game.Pay{}, PublishResult: "payment"},
			{Primitive: game.CounterObject{}, ResultGate: opt.Val(game.InstructionResultGate{
				Key: "payment", Succeeded: game.TriFalse,
			})},
		}
	}
	for _, tt := range []struct {
		name           string
		sequence       func() []game.Instruction
		optional, gate bool
		want           bool
		publisher      int
	}{
		{"actual tax counter", taxSequence, false, false, true, 1},
		{"single optional counter", func() []game.Instruction {
			return []game.Instruction{{Primitive: game.CounterObject{}}}
		}, true, false, true, 0},
		{"optional expanded result", taxSequence, true, false, false, 0},
		{"foreign payment gate", func() []game.Instruction {
			sequence := taxSequence()
			sequence[1].ResultGate.Val.Key = "another-payment"
			return sequence
		}, false, false, false, 0},
		{"wrong payment predicate", func() []game.Instruction {
			sequence := taxSequence()
			sequence[1].ResultGate.Val.Succeeded = game.TriTrue
			return sequence
		}, false, false, false, 0},
		{"single foreign gate", func() []game.Instruction {
			return []game.Instruction{{Primitive: game.CounterObject{},
				ResultGate: opt.Val(game.InstructionResultGate{Key: "foreign", Succeeded: game.TriTrue})}}
		}, false, false, false, 0},
		{"already optional", func() []game.Instruction {
			sequence := taxSequence()
			sequence[1].Optional = true
			return sequence
		}, false, false, false, 0},
		{"existing result publisher", func() []game.Instruction {
			sequence := taxSequence()
			sequence[1].PublishResult = "another-result"
			return sequence
		}, false, false, false, 0},
		{"overlapping result gate", taxSequence, false, true, false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			flow := scopedResultFlow{
				publishers: map[int]game.ResultKey{0: "actual"}, optional: map[int]bool{0: tt.optional},
			}
			if tt.optional {
				flow.actions = optionalActionGroups{
					owners: map[int]int{0: 0}, keys: map[int]game.OptionalDecisionKey{0: "decision"},
					controllerActor: map[int]bool{0: true},
				}
			}
			if tt.gate {
				flow.gates = map[int]game.InstructionResultGate{0: {Key: "outer", Succeeded: game.TriTrue}}
			}
			sequence := tt.sequence()
			reason, valid := flow.apply(0, sequence)
			if valid != tt.want || !valid && reason == "" {
				t.Fatalf("valid=%v reason=%q, want %v", valid, reason, tt.want)
			}
			if valid && (sequence[tt.publisher].PublishResult != "actual" ||
				sequence[tt.publisher].Optional != tt.optional ||
				tt.publisher == 1 && sequence[0].PublishResult != "payment") {
				t.Fatalf("publisher/decision/payment overwritten: %#v", sequence)
			}
		})
	}
}

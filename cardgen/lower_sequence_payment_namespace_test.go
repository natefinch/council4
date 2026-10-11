package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

func TestSequencePaymentActorProofRefusals(t *testing.T) {
	for _, tt := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
	}{
		{"missing owner", func(c *compiler.AbilityContent) { c.Conditions[1].Ownership.ResultProducerClauseID = 0 }},
		{"forward owner", func(c *compiler.AbilityContent) {
			c.Conditions[1].Ownership.ResultProducerClauseID = c.Effects[1].ClauseID
		}},
		{"foreign actor node", func(c *compiler.AbilityContent) { c.Conditions[1].Ownership.ReferenceNodeIDs = []int{999} }},
		{"duplicate actor", func(c *compiler.AbilityContent) { c.References = append(c.References, c.References[1]) }},
		{"unproved actor", func(c *compiler.AbilityContent) { c.References[1].Subject = compiler.ReferenceSubjectProof{} }},
		{"wrong actor target", func(c *compiler.AbilityContent) { c.References[1].Occurrence = 999 }},
		{"foreign owned actor", func(c *compiler.AbilityContent) { c.Conditions[1].Ownership.ResultSubjectReferenceNodeID = 999 }},
		{"foreign owned target", func(c *compiler.AbilityContent) { c.Conditions[1].Ownership.ResultSubjectTargetOccurrence = 999 }},
		{"foreign failure clause", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ClauseIDs = []int{c.Effects[1].ClauseID} }},
		{"missing tax clause", func(c *compiler.AbilityContent) { c.Effects[0].ClauseID = 0 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			compiled, diagnostics := compileTestOracle(
				"Counter target spell unless its controller pays {3}. If they do, you gain 2 life.",
				parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compiled.Abilities[0].Content
			tt.mutate(&content)
			flow, ok := planOptionalFlow(content)
			if !ok {
				return
			}
			if _, _, ok := planSequenceConditions(content, flow); ok {
				t.Fatal("malformed payment identity was accepted")
			}
		})
	}
	for _, text := range []string{
		"Counter target spell unless its controller pays {3}. If you do, draw a card.",
		"Counter target spell. You gain 1 life. If they do, draw a card.",
		"Counter target spell unless its controller pays {3}. They discard a card. If they do, draw a card.",
		"Counter up to two target spells unless their controller pays {3}. If they do, draw a card.",
		"Counter target spell unless its controller pays {3}. Counter another target spell unless its controller pays {2}.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Payment Near Miss", TypeLine: "Instant", Layout: "normal", OracleText: text})
	}
}

func TestSequencePaymentNamespaces(t *testing.T) {
	for _, text := range []string{
		"Counter target spell unless its controller pays {3}. Counter target spell unless its controller pays {2}. You gain 1 life.",
		"Counter target spell unless its controller pays {3}. You gain 1 life. Counter target spell unless its controller pays {2}.",
		"Counter target spell unless its controller pays {3}. If they do, you gain 2 life.",
		"Counter target spell unless its controller pays {3}. You gain 1 life. If they do, you gain 2 life.",
		"Counter target spell unless its controller pays {3}. If they do, you gain 2 life. If that spell is countered this way, you gain 3 life.",
	} {
		t.Run(text, func(t *testing.T) {
			defs, diagnostics, err := CompileCardDefs(&ScryfallCard{
				Name: "Payment Namespace Probe", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v %+v", err, diagnostics)
			}
			mode := defs[0].SpellAbility.Val.Modes[0]
			keys := map[game.ResultKey]bool{}
			for i, instruction := range mode.Sequence {
				if _, ok := instruction.Primitive.(game.Pay); !ok {
					continue
				}
				key := instruction.PublishResult
				if key == "" || key == unlessPaidResultKey || keys[key] {
					t.Fatalf("payment key is missing, unscoped or aliased: %q", key)
				}
				keys[key] = true
				if !slices.Contains(instruction.LocalProducts.Results, key) {
					t.Fatal("payment publisher lacks frame-local lifetime")
				}
				if i+1 >= len(mode.Sequence) || !mode.Sequence[i+1].ResultGate.Exists ||
					mode.Sequence[i+1].ResultGate.Val.Key != key ||
					mode.Sequence[i+1].ResultGate.Val.Succeeded != game.TriFalse {
					t.Fatal("counter does not consume its own failed payment")
				}
			}
			if len(keys) == 0 {
				t.Fatal("missing payment")
			}
			if len(keys) == 1 {
				paymentKey := mode.Sequence[0].PublishResult
				consumer := 2
				if !mode.Sequence[consumer].ResultGate.Exists {
					consumer++
				}
				if mode.Sequence[consumer].ResultGate.Val.Key != paymentKey ||
					mode.Sequence[consumer].ResultGate.Val.Succeeded != game.TriTrue {
					t.Fatal("payer success consumer did not bind the payment")
				}
				if mode.Sequence[1].PublishResult != "" && (mode.Sequence[1].PublishResult == paymentKey ||
					mode.Sequence[3].ResultGate.Val.Key != mode.Sequence[1].PublishResult) {
					t.Fatal("actual counter success did not retain its distinct producer")
				}
			}
		})
	}
}

func TestSequencePaymentNamespaceCompilerOwnership(t *testing.T) {
	for _, text := range []string{
		"Counter target spell unless its controller pays {3}. Counter another target spell unless its controller pays {2}.",
		"Counter target spell unless its controller pays {3}. If they do, you gain 2 life.",
		"Counter target spell unless its controller pays {3}. You gain 1 life. If they do, you gain 2 life.",
		"Counter target spell unless its controller pays {3}. If that spell is countered this way, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			compiled, diagnostics := compileTestOracle(text,
				parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compiled.Abilities[0].Content
			ownedPayments := map[int]int{}
			for _, effect := range content.Effects {
				if effectOwnsCounterTax(effect) {
					if _, duplicate := ownedPayments[effect.Payment.FailureConditionNodeID]; duplicate {
						t.Fatal("payment failure condition is not unique")
					}
					ownedPayments[effect.Payment.FailureConditionNodeID] = effect.ClauseID
				}
			}
			for _, condition := range content.Conditions {
				if condition.Predicate == compiler.ConditionPredicateTargetControllerDoesNotPay {
					clauseID, owned := ownedPayments[condition.NodeID]
					if !owned || len(condition.Ownership.ClauseIDs) != 1 || condition.Ownership.ClauseIDs[0] != clauseID {
						t.Fatal("payment boundary lost its clause owner")
					}
				}
				if condition.Predicate == compiler.ConditionPredicatePriorInstructionAccepted &&
					!counterPaymentActorCondition(content, 0, condition) {
					t.Fatal("embedded payment actor lacks exact owned target reference")
				}
			}
		})
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestExplicitPlayerResultGateOwnsEventAction(t *testing.T) {
	for _, text := range []string{
		"Whenever this creature deals damage to an opponent, that player discards a card at random. If the player does, they draw a card.",
		"Whenever this creature deals damage to an opponent, that player discards a card at random. You gain 1 life. If the player does, they draw a card.",
		"Whenever an opponent draws a card, that player may pay 2 life. If the player doesn't, you create a Treasure token.",
	} {
		t.Run(text, func(t *testing.T) {
			compiled, diagnostics := compileTestOracle(text, parser.Context{}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compiled.Abilities[0].Content
			condition := content.Conditions[0]
			if condition.Ownership.ResultProducerClauseID != content.Effects[0].ClauseID ||
				condition.Ownership.ResultSubjectTargetOccurrence != -1 ||
				len(condition.Ownership.ReferenceNodeIDs) != 1 ||
				condition.Ownership.ReferenceNodeIDs[0] != condition.Ownership.ResultSubjectReferenceNodeID {
				t.Fatal("explicit player gate lost exact action/reference ownership")
			}
			actor, ok := ownedResultActorReference(content, condition.Ownership.ResultSubjectReferenceNodeID)
			if !ok || actor.Binding != compiler.ReferenceBindingEventPlayer ||
				actor.SubjectDomain() != compiler.ReferenceSubjectPlayer {
				t.Fatal("THE-player gate borrowed a controller, object or default actor")
			}
			if _, ok := planOptionalFlow(content); !ok {
				t.Fatal("exact singular event-player action did not acquire its own result flow")
			}
			for _, mutate := range []func(*compiler.AbilityContent){
				func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ReferenceNodeIDs = nil },
				func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultSubjectReferenceNodeID = 99 },
				func(c *compiler.AbilityContent) {
					c.Conditions[0].Ownership.ReferenceNodeIDs = []int{c.Effects[0].SubjectReferences[0].NodeID}
				},
				func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 0 },
				func(c *compiler.AbilityContent) {
					c.Conditions[0].Ownership.ResultProducerClauseID = c.Effects[len(c.Effects)-1].ClauseID
				},
				func(c *compiler.AbilityContent) {
					c.Effects[0].SubjectReferences[0].Subject = compiler.ReferenceSubjectProof{}
				},
				func(c *compiler.AbilityContent) {
					c.Effects[0].SubjectReferences[0].Binding = compiler.ReferenceBindingSource
				},
			} {
				fresh, _ := compileTestOracle(text, parser.Context{}, compiler.Context{})
				content := fresh.Abilities[0].Content
				mutate(&content)
				if _, ok := planOptionalFlow(content); ok {
					t.Fatal("malformed explicit player actor/action acquired a result flow")
				}
			}
			assertCardPaths(t, &ScryfallCard{Name: "Player Action", Layout: "normal", TypeLine: "Enchantment", OracleText: text},
				"PublishResult", "ResultGate")
		})
	}
	for _, text := range []string{
		"Whenever an opponent draws a card, that player may pay 2 life. If you do, draw a card.",
		"Whenever an opponent draws a card, that player may pay 2 life. If you don't, draw a card.",
		"Whenever an opponent draws a card, each player may pay 2 life. If the player doesn't, you create a Treasure token.",
		"That player may pay 2 life. If the player doesn't, you create a Treasure token.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Player Action Near Miss", Layout: "normal", TypeLine: "Enchantment", OracleText: text})
	}
}

package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestFoldedPlayerDecisionOwnership(t *testing.T) {
	for _, tc := range []struct {
		text  string
		spell bool
	}{
		{"Whenever an opponent draws a card, that player may pay {2}. If the player doesn't, you create a Treasure token.", false},
		{"Whenever a player casts a spell, that player may pay {2}. If the player does, they draw a card.", false},
		{"At the beginning of the upkeep of enchanted creature's controller, that player may pay {4}. If the player does, untap the creature.", false},
		{"Return target creature to its owner's hand. Then that creature's controller may pay {U}{U}. If the player does, they may copy this spell and may choose a new target for that copy.", true},
	} {
		t.Run(tc.text, func(t *testing.T) {
			fresh := func() compiler.AbilityContent {
				t.Helper()
				compiled, diagnostics := compileTestOracle(tc.text, parser.Context{InstantOrSorcery: tc.spell}, compiler.Context{})
				if len(diagnostics) != 0 {
					t.Fatal(diagnostics)
				}
				return compiled.Abilities[0].Content
			}
			content := fresh()
			condition := content.Conditions[0]
			if len(condition.Ownership.ReferenceNodeIDs) != 1 {
				t.Fatal("folded decision lost its canonical condition actor")
			}
			actor, ok := ownedResultActorReference(content, condition.Ownership.ReferenceNodeIDs[0])
			if !ok || !content.OwnsPaymentDecision(actor) {
				t.Fatal("folded decision borrowed an ordinary player/object operand")
			}
			for _, effect := range content.Effects {
				for _, reference := range effect.References {
					if reference.NodeID == actor.NodeID {
						t.Fatal("condition actor leaked into the gated payload")
					}
				}
			}
			for _, mutate := range []func(*compiler.AbilityContent){
				func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ReferenceNodeIDs = nil },
				func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ClauseIDs = []int{99} },
				func(c *compiler.AbilityContent) { c.Conditions[0].NodeID = 99 },
				func(c *compiler.AbilityContent) { c.Conditions[0].Order = shared.SourceOrder{} },
				func(c *compiler.AbilityContent) {
					for i := range c.Effects {
						if c.Effects[i].Payment.Form != "" {
							c.Effects[i].Payment.Payer = parser.EffectPaymentPayerController
						}
					}
				},
				func(c *compiler.AbilityContent) {
					for i := range c.Effects {
						if c.Effects[i].Payment.Form != "" {
							c.Effects[i].Payment.Order = shared.SourceOrder{}
						}
					}
				},
				func(c *compiler.AbilityContent) {
					for i := range c.Effects {
						if c.Effects[i].Payment.Form != "" {
							c.Effects[i].Payment.Order.End = c.Conditions[0].Order.End + 1
						}
					}
				},
				func(c *compiler.AbilityContent) {
					local, _ := ownedResultActorReference(*c, actor.NodeID)
					c.References = append(c.References, local)
				},
			} {
				content := fresh()
				localActor, ok := ownedResultActorReference(content, actor.NodeID)
				if !ok || !content.OwnsPaymentDecision(localActor) {
					t.Fatal("mutation baseline has no owned payment decision")
				}
				mutate(&content)
				if content.OwnsPaymentDecision(localActor) {
					t.Fatal("foreign or malformed folded decision acquired payment ownership")
				}
			}
			if fresh().OwnsPaymentDecision(actor) {
				t.Fatal("another compilation borrowed the original decision proof")
			}
			content.Conditions[0].Span = shared.Span{}
			for i := range content.Effects {
				content.Effects[i].Payment.Span = shared.Span{}
			}
			if !content.OwnsPaymentDecision(actor) {
				t.Fatal("folded decision used byte spans rather than typed order and ownership")
			}
		})
	}
}

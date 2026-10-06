package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func optionalEnteredReceiptContent(t *testing.T) compiler.AbilityContent {
	t.Helper()
	document, diagnostics := parser.Parse(
		"Look at the top card of your library. You may put it onto the battlefield. That creature gains haste until end of turn.",
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	return compilation.Abilities[0].Content
}

func TestOptionalEnteredSubjectReceiptUsesExistingOwner(t *testing.T) {
	content := optionalEnteredReceiptContent(t)
	for _, tc := range []struct {
		name   string
		change func(*compiler.AbilityContent)
		refuse bool
	}{
		{"eligible singleton", func(*compiler.AbilityContent) {}, false},
		{"foreign subject", func(c *compiler.AbilityContent) {
			c.References[len(c.References)-1] = optionalEnteredReceiptContent(t).References[1]
		}, true},
		{"incompatible producer", func(c *compiler.AbilityContent) {
			c.Effects[1].ToZone = zone.Hand
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy := content
			copy.References = append([]compiler.CompiledReference(nil), content.References...)
			copy.Effects = append([]compiler.CompiledEffect(nil), content.Effects...)
			tc.change(&copy)
			groups, reason := planOptionalActionGroups(copy.Effects, map[int]int{1: 0, 2: 1, 3: 2})
			if reason != "" {
				t.Fatal(reason)
			}
			owner, key := groups.owners[1], groups.keys[1]
			reason = groups.requireEnteredSubjectReceipts(copy)
			if (reason != "") != tc.refuse {
				t.Fatalf("receipt demand reason=%q, refusal=%v", reason, tc.refuse)
			}
			if groups.owners[1] != owner || groups.keys[1] != key || groups.compound[1] {
				t.Fatal("demand changed optional ownership, key construction, or action group")
			}
			if !tc.refuse && !groups.requiredReceipts[owner] {
				t.Fatal("eligible entered subject did not request its existing owner receipt")
			}
		})
	}
}

func TestOptionalEnteredSubjectReceiptCardDefs(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Scoped Entry Receipt", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Look at the top card of your library. You may put it onto the battlefield. That creature gains haste until end of turn.",
	}, "SpellAbility.Val.Modes[0].Sequence[1].PublishOptionalDecision",
		"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.PutOnBattlefield).PublishLinked",
		"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.ApplyContinuous).Object.Val")
	assertCardPathsAbsent(t, &ScryfallCard{
		Name: "Unrelated Optional", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may draw a card. You gain 2 life.",
	}, "SpellAbility.Val.Modes[0].Sequence[0].PublishOptionalDecision")
}

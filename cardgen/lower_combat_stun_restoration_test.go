package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCombatStunRelatedPossessiveRestoration(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.",
		"Whenever this creature blocks a creature, tap that creature. That creature doesn't untap during its controller's next untap step.",
		"Whenever equipped creature blocks a creature, that creature doesn't untap during its controller's next untap step.",
		"Whenever enchanted creature blocks a creature, that creature doesn't untap during its controller's next untap step.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			typeLine := "Creature"
			if strings.HasPrefix(text, "Whenever equipped") {
				typeLine = "Artifact — Equipment"
			}
			if strings.HasPrefix(text, "Whenever enchanted") {
				typeLine = "Enchantment — Aura"
				text = "Enchant creature\n" + text
			}
			card := &ScryfallCard{Name: "Combat Stun", Layout: "normal", TypeLine: typeLine, OracleText: text}
			assertCardPaths(t, card, "Object.kind = game.ObjectReferenceEventRelatedPermanent")
		})
	}
	assertCardPaths(t, &ScryfallCard{
		Name: "Nested Stun", Layout: "normal", TypeLine: "Creature",
		OracleText: "When this creature enters, create two 0/2 blue Illusion creature tokens with \"Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.\"",
	}, "Object.kind = game.ObjectReferenceEventRelatedPermanent")
}

func TestStaticEquippedConditionOwnsActualSubject(t *testing.T) {
	t.Parallel()
	const text = "This creature can't attack or block unless it's equipped."
	assertCardPaths(t, &ScryfallCard{
		Name: "Equipment Condition", Layout: "normal", TypeLine: "Creature", OracleText: text,
	}, "MatchEquipped = true", "Object.Val.kind = game.ObjectReferenceSourcePermanent", "Kind = game.RuleEffectCantAttack", "Kind = game.RuleEffectCantBlock")
	compile := func() compiler.CompiledAbility {
		compilation, diagnostics := compileTestOracle(text, parser.Context{}, compiler.Context{})
		if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
			t.Fatalf("compile: %+v / %+v", compilation, diagnostics)
		}
		return compilation.Abilities[0]
	}
	for _, invalid := range []string{"missing", "foreign", "wrong identity"} {
		t.Run(invalid, func(t *testing.T) {
			t.Parallel()
			ability := compile()
			condition := ability.Static.Declarations[0].Condition
			if condition == nil || condition.ObjectReference == nil {
				t.Fatal("missing compiled condition subject")
			}
			switch invalid {
			case "missing":
				condition.ObjectReference = nil
			case "foreign":
				foreign := compile()
				condition.ObjectReference = foreign.Static.Declarations[0].Condition.ObjectReference
			case "wrong identity":
				condition.SubjectRefID++
			default:
				t.Fatal("unknown invalid condition")
			}
			_, handled, diagnostic := lowerStaticDeclarations(ability, &parser.Ability{})
			if !handled || diagnostic == nil {
				t.Fatal("unowned condition was admitted")
			}
		})
	}
}

package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileCombatCalligrapherMechanics(t *testing.T) {
	t.Parallel()
	source := "Inklings can't attack you or planeswalkers you control.\n" +
		"Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 2 {
		t.Fatalf("compilation = %#v, diagnostics = %#v", compilation, diagnostics)
	}

	declaration := compilation.Abilities[0].Static.Declarations[0]
	if declaration.Rule.Kind != StaticRuleCantAttackYou ||
		declaration.Group.Domain != StaticGroupBattlefield ||
		len(declaration.Group.Selection.SubtypesAny) != 1 ||
		declaration.Group.Selection.SubtypesAny[0] != types.Inkling {
		t.Fatalf("static declaration = %#v", declaration)
	}

	triggered := compilation.Abilities[1]
	pattern := triggered.Trigger.Pattern
	if pattern.Event != TriggerEventAttackerDeclared ||
		pattern.Controller != ControllerAny ||
		pattern.Player != TriggerPlayerOpponent ||
		pattern.AttackRecipient != TriggerAttackRecipientPlayer ||
		!pattern.OneOrMore ||
		!pattern.OneOrMorePerAttackTarget {
		t.Fatalf("trigger pattern = %#v", pattern)
	}
	if len(triggered.Content.Effects) != 1 {
		t.Fatalf("effects = %#v", triggered.Content.Effects)
	}
	effect := triggered.Content.Effects[0]
	if effect.Kind != EffectCreate ||
		effect.Context != parser.EffectContextEventPlayer ||
		effect.TokenAttackDefender != parser.AttackDefenderThatOpponent {
		t.Fatalf("effect = %#v", effect)
	}
	if !effect.Exact ||
		!effect.TokenPTKnown ||
		effect.TokenPower != 2 ||
		effect.TokenToughness != 1 ||
		len(effect.Selector.ColorsAny()) != 2 ||
		len(effect.Selector.SubtypesAny()) != 1 ||
		len(effect.TokenKeywords) != 1 ||
		effect.TokenKeywords[0] != parser.KeywordFlying {
		t.Fatalf("token shape: exact=%v pt=%v %d/%d colors=%v subtypes=%v keywords=%v",
			effect.Exact, effect.TokenPTKnown, effect.TokenPower, effect.TokenToughness,
			effect.Selector.ColorsAny(), effect.Selector.SubtypesAny(), effect.TokenKeywords)
	}
	if len(triggered.Content.References) != 0 {
		t.Fatalf("references = %#v, want defender correlation carried by typed effect", triggered.Content.References)
	}
	if len(triggered.Content.Keywords) != 0 {
		t.Fatalf("ability keywords = %#v, want token keywords carried by effect", triggered.Content.Keywords)
	}
}

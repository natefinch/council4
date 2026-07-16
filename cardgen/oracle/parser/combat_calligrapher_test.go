package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseSubtypeAttackRestrictionAgainstController(t *testing.T) {
	t.Parallel()
	source := "Inklings can't attack you or planeswalkers you control."
	atoms := atomsFor(t, source, "")
	if len(atoms.Subtypes()) != 1 || atoms.Subtypes()[0].Identity != types.Inkling {
		t.Fatalf("subtype atoms = %#v, want Inkling", atoms.Subtypes())
	}
	if subject, ok := parsePluralSubtypeGroupSubject(lexedWords(t, source), atoms); !ok {
		t.Fatal("plural subtype subject not recognized")
	} else if subject.Kind != EffectStaticSubjectAllCreatureSubtype {
		t.Fatalf("subject = %#v", subject)
	}
	declarations := parseStaticDeclarationSyntax(t, source, Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	declaration := declarations[0]
	if declaration.Kind != StaticDeclarationRule ||
		declaration.Subject.Kind != StaticDeclarationSubjectGroup ||
		declaration.Subject.Group.Kind != EffectStaticSubjectAllCreatureSubtype ||
		declaration.Subject.Group.Subtype != types.Inkling ||
		declaration.Rule.Operation.Kind != StaticRuleOperationAttack ||
		!staticRuleQualifiersAre(declaration.Rule.Qualifiers, StaticRuleQualifierDefenderYou) {
		t.Fatalf("declaration = %#v", declaration)
	}
}

func TestParsePlayerAttacksOpponentCreatesCorrelatedToken(t *testing.T) {
	t.Parallel()
	source := "Whenever a player attacks one of your opponents, that attacking player creates a tapped 2/1 white and black Inkling creature token with flying that's attacking that opponent."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("document = %#v, diagnostics = %#v", document, diagnostics)
	}
	ability := document.Abilities[0]
	event := ability.Trigger.TriggerEvent
	if event == nil ||
		event.Kind != TriggerEventKindAttack ||
		event.Actor.Kind != TriggerEventActorPlayer ||
		event.Player.Kind != TriggerPlayerSelectorOpponent ||
		event.AttackRecipient.Kind != TriggerEventAttackRecipientPlayer ||
		!event.OneOrMore ||
		!event.OneOrMorePerAttackTarget {
		t.Fatalf("trigger event = %#v", event)
	}
	if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	effect := ability.Sentences[0].Effects[0]
	if effect.Kind != EffectCreate ||
		effect.Context != EffectContextEventPlayer ||
		effect.AttackDefender != AttackDefenderThatOpponent ||
		!effect.Selection.Tapped ||
		!effect.Selection.Attacking ||
		len(effect.Selection.ColorsAny) != 2 ||
		len(effect.Selection.SubtypesAny) != 1 ||
		effect.Selection.SubtypesAny[0] != types.Inkling ||
		len(effect.TokenKeywords) != 1 ||
		effect.TokenKeywords[0] != KeywordFlying {
		t.Fatalf("effect = %#v", effect)
	}
}

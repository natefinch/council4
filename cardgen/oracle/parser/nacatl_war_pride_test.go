package parser

import "testing"

const nacatlWarPrideOracle = "This creature must be blocked by exactly one creature if able.\n" +
	"Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of creatures defending player controls. Exile the tokens at the beginning of the next end step."

func TestParseNacatlWarPrideReusableMechanics(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(nacatlWarPrideOracle, Context{CardName: "Nacatl War-Pride"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want static and attack trigger", len(document.Abilities))
	}
	static := document.Abilities[0]
	if len(static.ConditionBoundaries) != 0 || len(static.StaticDeclarations) != 1 {
		t.Fatalf("static ability = %#v", static)
	}
	rule := static.StaticDeclarations[0].Rule
	if !staticRuleQualifiersAre(rule.Qualifiers, StaticRuleQualifierExactlyOneCreature, StaticRuleQualifierIfAble) {
		t.Fatalf("static rule = %#v", rule)
	}
	triggered := document.Abilities[1]
	if len(triggered.Sentences) != 2 ||
		len(triggered.Sentences[0].Effects) != 1 ||
		len(triggered.Sentences[1].Effects) != 1 {
		t.Fatalf("triggered ability = %#v", triggered)
	}
	create := triggered.Sentences[0].Effects[0]
	if !create.Exact || !create.TokenCopyOfSource || !create.TokenCopyEntersTapped ||
		!create.TokenCopyAttacksDefender ||
		create.Amount.Selection == nil ||
		create.Amount.Selection.Controller != SelectionControllerDefendingPlayer {
		t.Fatalf("create effect = %#v", create)
	}
	exile := triggered.Sentences[1].Effects[0]
	if !exile.Exact || !exile.CreatedTokensReference || exile.DelayedTiming != DelayedTimingNextEndStep {
		t.Fatalf("exile effect = %#v", exile)
	}
}

func TestParseQualifiedDefendingPlayerCreatureCount(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of green creatures defending player controls. Exile the tokens at the beginning of the next end step.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	selection := document.Abilities[0].Sentences[0].Effects[0].Amount.Selection
	if selection == nil ||
		selection.Controller != SelectionControllerDefendingPlayer ||
		len(selection.ColorsAny) != 1 ||
		selection.ColorsAny[0] != ColorGreen {
		t.Fatalf("selection = %#v", selection)
	}
}

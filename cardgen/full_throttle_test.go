package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

const fullThrottleOracle = "After this main phase, there are two additional combat phases.\n" +
	"At the beginning of each combat this turn, untap all creatures that attacked this turn."

func TestLowerFullThrottle(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sorcery",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{4}{R}{R}",
		OracleText: fullThrottleOracle,
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", sequence)
	}
	extra, ok := sequence[0].Primitive.(game.AddExtraPhases)
	if !ok || extra.Combat || extra.CombatCount != 2 || extra.Main || extra.Beginning {
		t.Fatalf("extra phases = %#v, want exactly two combats", sequence[0].Primitive)
	}
	delayed, ok := sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("delayed primitive = %T, want game.CreateDelayedTrigger", sequence[1].Primitive)
	}
	if delayed.Trigger.OneShot ||
		delayed.Trigger.Window != game.DelayedWindowThisTurn ||
		!delayed.Trigger.EventPattern.Exists ||
		delayed.Trigger.EventPattern.Val.Event != game.EventBeginningOfStep ||
		delayed.Trigger.EventPattern.Val.Step != game.StepBeginningOfCombat {
		t.Fatalf("delayed trigger = %#v", delayed.Trigger)
	}
	inner := delayed.Trigger.Content.Modes[0].Sequence
	untap, ok := inner[0].Primitive.(game.Untap)
	if !ok || untap.Group.Domain() != game.GroupDomainAttackedThisTurn {
		t.Fatalf("delayed content = %#v, want attacked-this-turn group untap", inner)
	}
}

func TestRenderFullThrottleMechanics(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Sorcery",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{4}{R}{R}",
		OracleText: fullThrottleOracle,
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.AddExtraPhases{",
		"CombatCount: 2,",
		"game.CreateDelayedTrigger{",
		"game.DelayedWindowThisTurn",
		"game.StepBeginningOfCombat",
		"game.AttackedThisTurnGroup(",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerFullThrottleMechanicsIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(fullThrottleOracle, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}

	extraContent := compilation.Abilities[0].Content
	extraContent.Effects[0].Text = "lowering must use the typed combat count"
	extra, diagnostic := lowerAdditionalCombatPhase(contentCtx{content: extraContent})
	if diagnostic != nil {
		t.Fatalf("extra-combat diagnostic = %#v", diagnostic)
	}
	primitive, ok := extra.Modes[0].Sequence[0].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("extra-combat primitive = %T, want game.AddExtraPhases", extra.Modes[0].Sequence[0].Primitive)
	}
	if primitive.CombatCount != 2 {
		t.Fatalf("extra phases = %#v, want typed count 2", primitive)
	}

	delayedContent := compilation.Abilities[1].Content
	delayedContent.Effects[0].Text = "lowering must use the typed delayed trigger"
	delayed, diagnostic := lowerDelayedTriggerContent(contentCtx{content: delayedContent})
	if diagnostic != nil {
		t.Fatalf("delayed-trigger diagnostic = %#v", diagnostic)
	}
	create, ok := delayed.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("delayed-trigger primitive = %T, want game.CreateDelayedTrigger", delayed.Modes[0].Sequence[0].Primitive)
	}
	untap, ok := create.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Untap)
	if !ok {
		t.Fatalf(
			"delayed-trigger content primitive = %T, want game.Untap",
			create.Trigger.Content.Modes[0].Sequence[0].Primitive,
		)
	}
	if untap.Group.Domain() != game.GroupDomainAttackedThisTurn {
		t.Fatalf("untap group domain = %v, want attacked this turn", untap.Group.Domain())
	}

	untapDocument, diagnostics := parser.Parse(
		"Untap all creatures that attacked this turn.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("untap parse diagnostics = %#v", diagnostics)
	}
	untapCompilation, diagnostics := compiler.Compile(untapDocument, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("untap compile diagnostics = %#v", diagnostics)
	}
	untapContent := untapCompilation.Abilities[0].Content
	untapContent.Effects[0].Text = "lowering must use the typed historical group"
	loweredUntap, diagnostic := lowerUntapSpell(contentCtx{content: untapContent})
	if diagnostic != nil {
		t.Fatalf("historical-untap diagnostic = %#v", diagnostic)
	}
	historical, ok := loweredUntap.Modes[0].Sequence[0].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("historical untap primitive = %T, want game.Untap", loweredUntap.Modes[0].Sequence[0].Primitive)
	}
	if historical.Group.Domain() != game.GroupDomainAttackedThisTurn {
		t.Fatalf("historical untap group domain = %v, want attacked this turn", historical.Group.Domain())
	}
}

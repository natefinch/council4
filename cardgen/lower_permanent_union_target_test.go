package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestPermanentUnionTargetVerbs(t *testing.T) {
	for _, tc := range []struct {
		text string
		path string
	}{
		{"Return target creature or Vehicle to its owner's hand.", "Primitive.(game.MovePermanent).Destination = zone.Hand"},
		{"Tap target creature or Vehicle.", "Primitive.(game.Tap).Object.kind = game.ObjectReferenceTargetPermanent"},
		{"Untap target creature or Vehicle.", "Primitive.(game.Untap).Object.kind = game.ObjectReferenceTargetPermanent"},
		{"Destroy target creature or Vehicle.", "Primitive.(game.Destroy).Object.kind = game.ObjectReferenceTargetPermanent"},
		{"Exile target creature or Vehicle.", "Primitive.(game.MovePermanent).Destination = zone.Exile"},
		{"Put two +1/+1 counters on target creature or Vehicle.", "Primitive.(game.AddCounter).Amount.fixed = 2"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Permanent Union", Layout: "normal", TypeLine: "Instant", OracleText: tc.text,
			},
				"Modes[0].Targets[0].Allow = game.TargetAllowPermanent",
				"Modes[0].Targets[0].Selection.Val.AnyOf[0].RequiredTypesAny[0] = types.Creature",
				"Modes[0].Targets[0].Selection.Val.AnyOf[1].SubtypesAny[0] = types.Vehicle",
				tc.path,
			)
		})
	}
}
func TestPermanentUnionTargetComposition(t *testing.T) {
	for _, prefix := range []string{
		"Target player gains 1 life. ",
		"Return target creature card from your graveyard to your hand. ",
		"Counter target spell. ",
	} {
		t.Run(prefix, func(t *testing.T) {
			card := &ScryfallCard{
				Name: "Union Composition", Layout: "normal", TypeLine: "Instant",
				OracleText: prefix + "Return target creature or Vehicle to its owner's hand. Its controller loses 2 life. Its owner gains 3 life.",
			}
			assertCardPaths(t, card,
				"Modes[0].Targets[1].Selection.Val.AnyOf[0].RequiredTypesAny[0] = types.Creature",
				"Modes[0].Targets[1].Selection.Val.AnyOf[1].SubtypesAny[0] = types.Vehicle",
				"Modes[0].Sequence[1].Primitive.(game.MovePermanent).Object.targetIndex = 1",
				"Modes[0].Sequence[2].Primitive.(game.LoseLife).Player.kind = game.PlayerReferenceObjectController",
				"Modes[0].Sequence[2].Primitive.(game.LoseLife).Player.object.Val.targetIndex = 1",
				"Modes[0].Sequence[3].Primitive.(game.GainLife).Player.kind = game.PlayerReferenceObjectOwner",
				"Modes[0].Sequence[3].Primitive.(game.GainLife).Player.object.Val.targetIndex = 1",
			)
		})
	}
}

func TestPermanentUnionTargetInheritedObject(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Union Inheritance", Layout: "normal", TypeLine: "Instant",
		OracleText: "Tap target creature or Vehicle. You gain 1 life. Untap it.",
	},
		"Modes[0].Sequence[0].Primitive.(game.Tap).Object.kind = game.ObjectReferenceTargetPermanent",
		"Modes[0].Sequence[2].Primitive.(game.Untap).Object.kind = game.ObjectReferenceTargetPermanent",
	)
}

func TestClauseDeclaredTargetReturnAfterCardExile(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		t.Run(noun, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Declared Target", Layout: "normal", TypeLine: "Instant",
				OracleText: "Exile target card from a graveyard. Return target " + noun + " to its owner's hand.",
			}, "Modes[0].Sequence[1].Primitive.(game.MovePermanent).Object.targetIndex = 1",
				"Modes[0].Sequence[1].Primitive.(game.MovePermanent).Destination = zone.Hand")
		})
	}
}

func TestClauseDeclaredTargetDoesNotInventFollowupOwnership(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		assertCardUnsupported(t, &ScryfallCard{
			Name: "Unowned Followup", Layout: "normal", TypeLine: "Instant",
			OracleText: "Exile target card from a graveyard. Return target " + noun +
				" to its owner's hand. Its controller loses 2 life. Its owner gains 3 life.",
		})
	}
}

func TestPermanentUnionTargetFilters(t *testing.T) {
	for _, tc := range []struct {
		text string
		path string
	}{
		{"Tap target creature or Vehicle you control.", "Selection.Val.Controller = game.ControllerYou"},
		{"Destroy target blue creature or artifact.", "Selection.Val.AnyOf[0].ColorsAny[0] = color.Blue"},
		{"Untap target creature or tapped Vehicle.", "Selection.Val.AnyOf[1].Tapped = game.TriTrue"},
		{"Exile target artifact creature or Vehicle.", "Selection.Val.AnyOf[0].RequiredTypes[0] = types.Artifact"},
		{"Return target creature with mana value 3 or less or Vehicle with mana value 3 or less to its owner's hand.", "Selection.Val.AnyOf[1].ManaValue.Val.Value = 3"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Union Filters", Layout: "normal", TypeLine: "Instant", OracleText: tc.text,
			}, tc.path)
		})
	}
}

func TestPermanentUnionTargetRefusals(t *testing.T) {
	for _, text := range []string{
		"Return target creature or Vehicle to its controller's hand.",
		"Return target creature or Vehicle to your hand.",
		"Return target creature or Vehicle to its owner's hand face down.",
		"Return all target creature or Vehicle to its owner's hand.",
		"Return each target creature or Vehicle to its owner's hand.",
		"Tap up to one target creature or Vehicle.",
		"Tap two target creatures or Vehicles.",
		"Tap any number of target creatures or Vehicles.",
		"Tap target creature you control or Vehicle.",
		"Tap target blue red creature or artifact.",
		"Tap target creature or Vehicle you own.",
		"Tap target creature or Vehicle from your graveyard.",
		"Tap target creature or Vehicle card.",
		"Tap target creature or player.",
		"Tap target creature or spell.",
		"Tap target creature or Vehicle with mana value X.",
		"Tap target creature or Vehicle with an additional rider.",
		"Tap target creature or Vehicle that is blue.",
		"Tap target creature or Vehicle with ward {2}.",
		"Tap target creature or Vehicle. That spell gains flying until end of turn.",
		"Tap target creature or Vehicle. Its copy gains flying until end of turn.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Union Refusal", Layout: "normal", TypeLine: "Instant", OracleText: text,
			})
		})
	}
}

func TestPermanentUnionTargetAdapterProofOwnership(t *testing.T) {
	const text = "Return target creature or Vehicle to its owner's hand."
	compile := func() (parser.Document, compiler.AbilityContent) {
		document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return document, compilation.Abilities[0].Content
	}
	for _, tc := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
	}{
		{"missing", func(c *compiler.AbilityContent) { c.References[0].Subject = compiler.ReferenceSubjectProof{} }},
		{"foreign", func(c *compiler.AbilityContent) {
			_, foreign := compile()
			c.References[0] = foreign.References[0]
		}},
		{"node", func(c *compiler.AbilityContent) { c.References[0].NodeID++ }},
		{"clause", func(c *compiler.AbilityContent) { c.References[0].ProducerClauseID++ }},
		{"occurrence", func(c *compiler.AbilityContent) { c.References[0].Occurrence++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document, content := compile()
			tc.mutate(&content)
			ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span, enclosingKind: compiler.AbilitySpell}
			if _, diagnostic := lowerContent("Union Ownership", ctx, &document.Abilities[0]); diagnostic == nil {
				t.Fatal("modified or foreign proof admitted")
			}
		})
	}
}

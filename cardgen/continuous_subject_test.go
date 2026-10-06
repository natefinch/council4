package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// These tests exercise the shared continuous-effect composition helpers directly,
// covering each recipient a continuous effect can be routed to: the source
// permanent, a single permanent target, a static creature group, and an arbitrary
// resolved object. Every continuous-effect lowerer (double power/toughness, set
// base power/toughness, color change, animation, keyword grant/loss, switch
// power/toughness) reaches one of these recipients through this same path, so the
// coverage here backs all of those families.

// applyContinuousOf extracts the single ApplyContinuous primitive a one-shot
// continuous mode produces, failing the test for any other shape.
func applyContinuousOf(t *testing.T, content game.AbilityContent) game.ApplyContinuous {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("got %d modes, want 1", len(content.Modes))
	}
	mode := content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("got %d instructions, want 1", len(mode.Sequence))
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive is %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	return apply
}

func sampleContinuousEffects() []game.ContinuousEffect {
	return []game.ContinuousEffect{
		{Layer: game.LayerPowerToughnessSet},
		{Layer: game.LayerType, AddEveryCreatureType: true},
	}
}

// TestContinuousSourceMode covers the source recipient: the effects bind to the
// source permanent and carry no static group.
func TestContinuousSourceMode(t *testing.T) {
	ctx := contentCtx{content: referenceSubjectFixture(t, "{1}: This creature gains flying until end of turn.", false)}
	content, diagnostic := continuousSourceMode(ctx, sampleContinuousEffects(), game.DurationUntilEndOfTurn)
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	apply := applyContinuousOf(t, content)
	if !apply.Object.Exists || !reflect.DeepEqual(apply.Object.Val, game.SourcePermanentReference()) {
		t.Fatalf("Object = %#v, want source permanent reference", apply.Object)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("Duration = %v, want until end of turn", apply.Duration)
	}
	for _, effect := range apply.ContinuousEffects {
		if !reflect.DeepEqual(effect.Group, game.GroupReference{}) {
			t.Fatalf("source effect carries a group: %#v", effect.Group)
		}
	}
}

// TestContinuousObjectMode covers an arbitrary resolved-object recipient.
func TestContinuousObjectMode(t *testing.T) {
	object := game.TargetPermanentReference(2)
	content := continuousObjectMode(object, sampleContinuousEffects(), game.DurationUntilEndOfTurn)
	apply := applyContinuousOf(t, content)
	if !apply.Object.Exists || !reflect.DeepEqual(apply.Object.Val, object) {
		t.Fatalf("Object = %#v, want %#v", apply.Object, object)
	}
}

// TestContinuousGroupMode covers the static-group recipient: the group rides on
// every continuous effect, the caller's slice is left untouched, and no per-object
// Object binding is set.
func TestContinuousGroupMode(t *testing.T) {
	group := game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})
	effects := sampleContinuousEffects()
	content := continuousGroupMode(group, effects, game.DurationUntilEndOfTurn)
	apply := applyContinuousOf(t, content)
	if apply.Object.Exists {
		t.Fatalf("group mode set an Object: %#v", apply.Object)
	}
	if len(apply.ContinuousEffects) != len(effects) {
		t.Fatalf("got %d effects, want %d", len(apply.ContinuousEffects), len(effects))
	}
	for i, effect := range apply.ContinuousEffects {
		if !reflect.DeepEqual(effect.Group, group) {
			t.Fatalf("effect %d group = %#v, want %#v", i, effect.Group, group)
		}
	}
	for i, effect := range effects {
		if !reflect.DeepEqual(effect.Group, game.GroupReference{}) {
			t.Fatalf("caller effect %d was mutated with group %#v", i, effect.Group)
		}
	}
}

// TestContinuousSubjectModeRoutesRecipients drives the continuousSubjectMode
// router through its source, group, and single-target branches, the three
// recipients whose recognition it shares across effect families, and confirms a
// disallowed subject fails closed.
func TestContinuousSubjectModeRoutesRecipients(t *testing.T) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, &shared.Diagnostic{Summary: "unsupported"}
	}
	emptyCtx := contentCtx{content: compiler.AbilityContent{}}

	t.Run("source", func(t *testing.T) {
		ctx := contentCtx{content: referenceSubjectFixture(t, "{1}: This creature gains flying until end of turn.", false)}
		content, diag := continuousSubjectMode(ctx, &compiler.CompiledEffect{}, sampleContinuousEffects(),
			game.DurationUntilEndOfTurn, continuousSubjectOptions{SourceForm: true}, unsupported)
		if diag != nil {
			t.Fatalf("unexpected diagnostic: %v", diag)
		}
		apply := applyContinuousOf(t, content)
		if !reflect.DeepEqual(apply.Object.Val, game.SourcePermanentReference()) {
			t.Fatalf("Object = %#v, want source permanent reference", apply.Object)
		}
	})

	t.Run("group", func(t *testing.T) {
		effect := &compiler.CompiledEffect{StaticSubject: compiler.StaticSubjectAllCreatures}
		content, diag := continuousSubjectMode(emptyCtx, effect, sampleContinuousEffects(),
			game.DurationUntilEndOfTurn, continuousSubjectOptions{AllowGroup: true}, unsupported)
		if diag != nil {
			t.Fatalf("unexpected diagnostic: %v", diag)
		}
		apply := applyContinuousOf(t, content)
		if apply.Object.Exists {
			t.Fatalf("group routing set an Object: %#v", apply.Object)
		}
		for _, effect := range apply.ContinuousEffects {
			if reflect.DeepEqual(effect.Group, game.GroupReference{}) {
				t.Fatal("group routing left an effect without a group")
			}
		}
	})

	t.Run("group disallowed fails closed", func(t *testing.T) {
		effect := &compiler.CompiledEffect{StaticSubject: compiler.StaticSubjectAllCreatures}
		_, diag := continuousSubjectMode(emptyCtx, effect, sampleContinuousEffects(),
			game.DurationUntilEndOfTurn, continuousSubjectOptions{AllowTarget: true}, unsupported)
		if diag == nil {
			t.Fatal("group subject lowered despite AllowGroup being unset")
		}
	})

	t.Run("target", func(t *testing.T) {
		ctx := contentCtx{content: compiler.AbilityContent{
			Targets: []compiler.CompiledTarget{targetCreatureFixture()},
		}}
		content, diag := continuousSubjectMode(ctx, &compiler.CompiledEffect{}, sampleContinuousEffects(),
			game.DurationUntilEndOfTurn, continuousSubjectOptions{AllowTarget: true}, unsupported)
		if diag != nil {
			t.Fatalf("unexpected diagnostic: %v", diag)
		}
		apply := applyContinuousOf(t, content)
		if !reflect.DeepEqual(apply.Object.Val, game.TargetPermanentReference(0)) {
			t.Fatalf("Object = %#v, want first target permanent reference", apply.Object)
		}
	})
}

// targetCreatureFixture builds a minimal single "target creature" the
// permanentTargetSpecWithCardinality reducer accepts, so the router's target
// branch can be exercised without the parser.
func targetCreatureFixture() compiler.CompiledTarget {
	return compiler.CompiledTarget{
		Exact:       true,
		Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Selector:    compiler.CompiledSelector{Kind: compiler.SelectorCreature},
	}
}

// TestContinuousReferenceObject confirms the referenced-object subject resolver
// accepts every binding the runtime's ApplyContinuous can resolve — source,
// source-attached, and triggering event permanent — so a continuous effect can
// name any of them, while gating a bare target back-reference on the referenced
// object context and failing closed on a player binding it cannot represent.
func TestContinuousReferenceObject(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		spell     bool
		reference compiler.CompiledReference
		effect    compiler.CompiledEffect
		want      game.ObjectReference
		wantOK    bool
	}{
		{
			name:      "source",
			text:      "{1}: This creature gains flying until end of turn.",
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingSource},
			want:      game.SourcePermanentReference(),
			wantOK:    true,
		},
		{
			name:      "source attached",
			text:      "{1}: Double equipped creature's power until end of turn.",
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingSourceAttached},
			want:      game.SourceAttachedPermanentReference(),
			wantOK:    true,
		},
		{
			name:      "event permanent",
			text:      "Whenever another creature enters, that creature gains flying until end of turn.",
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingEventPermanent},
			want:      game.EventPermanentReference(),
			wantOK:    true,
		},
		{
			name:      "target back-reference with referenced-object context",
			text:      "Tap target creature. That creature gains flying until end of turn.",
			spell:     true,
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingTarget, Occurrence: 0},
			effect:    compiler.CompiledEffect{Context: parser.EffectContextReferencedObject},
			want:      game.TargetPermanentReference(0),
			wantOK:    true,
		},
		{
			name:      "target back-reference without referenced-object context fails closed",
			text:      "Tap target creature. That creature gains flying until end of turn.",
			spell:     true,
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingTarget, Occurrence: 0},
			effect:    compiler.CompiledEffect{Context: parser.EffectContextController},
			wantOK:    false,
		},
		{
			name:      "player binding fails closed",
			text:      "Whenever this creature deals combat damage to a player, that player draws a card.",
			reference: compiler.CompiledReference{Binding: compiler.ReferenceBindingEventPlayer},
			wantOK:    false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reference := referenceSubjectFixtureBinding(t, tc.text, tc.spell, tc.reference.Binding)
			object, ok := continuousReferenceObject(reference, &tc.effect)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && !reflect.DeepEqual(object, tc.want) {
				t.Fatalf("object = %#v, want %#v", object, tc.want)
			}
		})
	}
}

// Unproven source back-references cannot become subjects; genuine resolving
// spell cards and original permanent sources use distinct compiler proofs.
func TestContinuousReferenceObjectSpellSourceBackReference(t *testing.T) {
	source := compiler.CompiledReference{Binding: compiler.ReferenceBindingSource}

	backReference := &compiler.CompiledEffect{Context: parser.EffectContextReferencedObject}
	if _, ok := continuousReferenceObject(source, backReference); ok {
		t.Fatal("spell source back-reference resolved, want fail closed")
	}

	selfReference := &compiler.CompiledEffect{Context: parser.EffectContextSource}
	if _, ok := continuousReferenceObject(source, selfReference); ok {
		t.Fatal("unproven handmade self-reference resolved")
	}
	face := lowerSingleFace(t, &ScryfallCard{
		Name: "Self Exile", Layout: "normal", TypeLine: "Instant", OracleText: "Exile this spell.",
	})
	move, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.MoveResolvingSpell)
	if !ok || move.Destination != zone.Exile {
		t.Fatal("normally compiled self spell did not use its resolving card")
	}
	permanent := referenceSubjectFixtureBinding(t, "{1}: This creature gains flying until end of turn.", false, compiler.ReferenceBindingSource)
	object, ok := continuousReferenceObject(permanent, backReference)
	if !ok {
		t.Fatal("permanent-ability source reference did not resolve")
	}
	if !reflect.DeepEqual(object, game.SourcePermanentReference()) {
		t.Fatalf("permanent-ability object = %#v, want original source permanent reference", object)
	}
}

// Retain this historical test name while removing its unsafe caller-selected
// card-following expectation: live mutation names the original source object.
func TestContinuousReferenceObjectSourceAsCard(t *testing.T) {
	reference := referenceSubjectFixtureBinding(t, "{1}: This creature gains flying until end of turn.", false, compiler.ReferenceBindingSource)
	object, ok := continuousReferenceObject(reference, &compiler.CompiledEffect{})
	if !ok {
		t.Fatal("source-as-card reference did not resolve")
	}
	if !reflect.DeepEqual(object, game.SourcePermanentReference()) {
		t.Fatalf("object = %#v, want original source permanent reference", object)
	}
}

func referenceSubjectFixture(t *testing.T, text string, spell bool) compiler.AbilityContent {
	t.Helper()
	document, diagnostics := parser.Parse(text, parser.Context{CardName: "Subject Fixture", InstantOrSorcery: spell})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
		t.Fatalf("normal subject compilation: %v", diagnostics)
	}
	return compilation.Abilities[0].Content
}

func referenceSubjectFixtureBinding(t *testing.T, text string, spell bool, binding compiler.ReferenceBinding) compiler.CompiledReference {
	t.Helper()
	content := referenceSubjectFixture(t, text, spell)
	if binding == compiler.ReferenceBindingSourceAttached && len(content.Effects) == 1 {
		if reference, ok := content.Source.AttachedObjectSubject(content.Effects[0]); ok {
			return reference
		}
	}
	for _, reference := range content.References {
		if reference.Binding == binding && reference.SubjectSupported() {
			return reference
		}
	}
	t.Fatalf("normal compilation has no validated subject binding %v: %s", binding, text)
	return compiler.CompiledReference{}
}

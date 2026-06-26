package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseStaticPluralSubtypeGroupSubject(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source      string
		kind        EffectStaticSubjectKind
		subtypesAny []types.Sub
	}{
		"controlled conjunction": {
			source:      "Skeletons and Zombies you control get +1/+1.",
			kind:        EffectStaticSubjectControlledCreatureSubtype,
			subtypesAny: []types.Sub{types.Skeleton, types.Zombie},
		},
		"controlled oxford list": {
			source:      "Skeletons, Vampires, and Zombies you control get +1/+1.",
			kind:        EffectStaticSubjectControlledCreatureSubtype,
			subtypesAny: []types.Sub{types.Skeleton, types.Vampire, types.Zombie},
		},
		"other controlled conjunction": {
			source:      "Other Robots and Constructs you control get +1/+1.",
			kind:        EffectStaticSubjectOtherControlledCreatureSubtype,
			subtypesAny: []types.Sub{types.Robot, types.Construct},
		},
		"battlefield other single": {
			source:      "Other Goblins get +1/+1.",
			kind:        EffectStaticSubjectOtherCreatureSubtype,
			subtypesAny: []types.Sub{types.Goblin},
		},
		"battlefield all single": {
			source:      "All Saprolings get +1/+1.",
			kind:        EffectStaticSubjectAllCreatureSubtype,
			subtypesAny: []types.Sub{types.Saproling},
		},
		"battlefield bare conjunction": {
			source:      "Servos and Thopters get +1/+1.",
			kind:        EffectStaticSubjectAllCreatureSubtype,
			subtypesAny: []types.Sub{types.Servo, types.Thopter},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			declarations := parseStaticDeclarationSyntax(t, test.source, Context{})
			if len(declarations) != 1 {
				t.Fatalf("declarations = %#v, want one", declarations)
			}
			subject := declarations[0].Subject
			if subject.Kind != StaticDeclarationSubjectGroup || subject.Group.Kind != test.kind {
				t.Fatalf("subject = %#v, want group %s", subject, test.kind)
			}
			if !slices.Equal(subject.Group.SubtypesAny, test.subtypesAny) {
				t.Fatalf("subtypesAny = %#v, want %#v", subject.Group.SubtypesAny, test.subtypesAny)
			}
		})
	}
}

// TestParseStaticPluralSubtypeGroupSubjectDefersSingleControlled confirms the
// single-subtype controlled plural form ("Goblins you control get ...") is still
// owned by the established single-subtype production, so its output is unchanged
// by the conjunction recognizer.
func TestParseStaticPluralSubtypeGroupSubjectDefersSingleControlled(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t, "Goblins you control get +1/+1.", Context{})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	subject := declarations[0].Subject
	if subject.Kind != StaticDeclarationSubjectGroup ||
		subject.Group.Kind != EffectStaticSubjectControlledCreatureSubtype {
		t.Fatalf("subject = %#v, want controlled creature subtype group", subject)
	}
	if subject.Group.Subtype != types.Goblin || len(subject.Group.SubtypesAny) != 0 {
		t.Fatalf("subtype = %q subsAny = %#v, want single Goblin with no SubtypesAny",
			subject.Group.Subtype, subject.Group.SubtypesAny)
	}
}

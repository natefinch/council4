package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileGandalfStaticMechanics(t *testing.T) {
	for name, tc := range map[string]struct {
		source string
		check  func(*testing.T, StaticDeclaration)
	}{
		"flash permission": {
			source: "You may cast legendary spells and artifact spells as though they had flash.",
			check: func(t *testing.T, declaration StaticDeclaration) {
				if declaration.CastAsThoughFlash == nil || len(declaration.CastAsThoughFlash.Filters) != 2 {
					t.Fatalf("declaration = %#v", declaration)
				}
				filters := declaration.CastAsThoughFlash.Filters
				if len(filters[0].Supertypes) != 1 || filters[0].Supertypes[0] != types.Legendary ||
					len(filters[1].Types) != 1 || filters[1].Types[0] != types.Artifact {
					t.Fatalf("filters = %#v", filters)
				}
			},
		},
		"trigger multiplier": {
			source: "If a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
			check: func(t *testing.T, declaration StaticDeclaration) {
				cause := declaration.ControlledMultiplier.CausePermanentZoneChange
				if cause == nil || !cause.Enters || !cause.Leaves || len(cause.Filters) != 2 {
					t.Fatalf("cause = %#v", cause)
				}
				if len(cause.Filters[0].Supertypes) != 1 || cause.Filters[0].Supertypes[0] != types.Legendary ||
					len(cause.Filters[1].Types) != 1 || cause.Filters[1].Types[0] != types.Artifact {
					t.Fatalf("filters = %#v", cause.Filters)
				}
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			compilation, diagnostics := compileSource(tc.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 ||
				compilation.Abilities[0].Static == nil ||
				len(compilation.Abilities[0].Static.Declarations) != 1 {
				t.Fatalf("compilation = %#v", compilation)
			}
			tc.check(t, compilation.Abilities[0].Static.Declarations[0])
		})
	}
}

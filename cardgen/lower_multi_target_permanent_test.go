package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerDualTargetBounce proves the dual-target bounce "Return target <A> and
// target <B> to their owners' hands." lowers to two single-target specs in
// Oracle order and one Bounce per slot, each addressing its own target index.
func TestLowerDualTargetBounce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		controllers [2]game.ControllerRelation
		permTypes   [2]types.Card
	}{
		{
			name:        "you control and you don't control",
			oracleText:  "Return target permanent you control and target permanent you don't control to their owners' hands.",
			controllers: [2]game.ControllerRelation{game.ControllerYou, game.ControllerNotYou},
			permTypes:   [2]types.Card{"", ""},
		},
		{
			name:        "creature and land",
			oracleText:  "Return target creature and target land to their owners' hands.",
			controllers: [2]game.ControllerRelation{game.ControllerAny, game.ControllerAny},
			permTypes:   [2]types.Card{types.Creature, types.Land},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Dual Bounce",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 2 {
				t.Fatalf("targets = %#v, want two specs", mode.Targets)
			}
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
			}
			for i := range mode.Targets {
				spec := mode.Targets[i]
				if spec.MinTargets != 1 || spec.MaxTargets != 1 {
					t.Fatalf("spec[%d] cardinality = {%d,%d}, want {1,1}", i, spec.MinTargets, spec.MaxTargets)
				}
				if spec.Allow != game.TargetAllowPermanent {
					t.Fatalf("spec[%d] allow = %v, want TargetAllowPermanent", i, spec.Allow)
				}
				if spec.Predicate.Controller != test.controllers[i] {
					t.Fatalf("spec[%d] controller = %v, want %v", i, spec.Predicate.Controller, test.controllers[i])
				}
				if test.permTypes[i] != "" {
					if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != test.permTypes[i] {
						t.Fatalf("spec[%d] types = %v, want [%v]", i, spec.Predicate.PermanentTypes, test.permTypes[i])
					}
				}
				bounce, ok := mode.Sequence[i].Primitive.(game.Bounce)
				if !ok || bounce.Object != game.TargetPermanentReference(i) {
					t.Fatalf("sequence[%d] = %#v, want Bounce of TargetPermanentReference(%d)", i, mode.Sequence[i].Primitive, i)
				}
			}
		})
	}
}

// TestLowerMultiTargetPermanentVerbs proves plural and optional permanent
// destroy, tap, untap, and bounce wordings lower to a single multi-target spec
// carrying the cardinality range and one verb primitive per slot, each
// addressing its own target index — the same machinery the exile path uses.
func TestLowerMultiTargetPermanentVerbs(t *testing.T) {
	t.Parallel()
	type primCheck func(t *testing.T, i int, prim game.Primitive)
	destroyCheck := func(t *testing.T, i int, prim game.Primitive) {
		d, ok := prim.(game.Destroy)
		if !ok || d.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] = %#v, want Destroy of TargetPermanentReference(%d)", i, prim, i)
		}
	}
	tapCheck := func(t *testing.T, i int, prim game.Primitive) {
		p, ok := prim.(game.Tap)
		if !ok || p.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] = %#v, want Tap of TargetPermanentReference(%d)", i, prim, i)
		}
	}
	untapCheck := func(t *testing.T, i int, prim game.Primitive) {
		p, ok := prim.(game.Untap)
		if !ok || p.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] = %#v, want Untap of TargetPermanentReference(%d)", i, prim, i)
		}
	}
	bounceCheck := func(t *testing.T, i int, prim game.Primitive) {
		p, ok := prim.(game.Bounce)
		if !ok || p.Object != game.TargetPermanentReference(i) {
			t.Fatalf("sequence[%d] = %#v, want Bounce of TargetPermanentReference(%d)", i, prim, i)
		}
	}

	tests := []struct {
		name       string
		oracleText string
		minTargets int
		maxTargets int
		permType   types.Card
		check      primCheck
	}{
		{"destroy two creatures", "Destroy two target creatures.", 2, 2, types.Creature, destroyCheck},
		{"destroy up to two artifacts", "Destroy up to two target artifacts.", 0, 2, types.Artifact, destroyCheck},
		{"destroy three permanents", "Destroy three target permanents.", 3, 3, "", destroyCheck},
		{"tap up to two creatures", "Tap up to two target creatures.", 0, 2, types.Creature, tapCheck},
		{"untap two lands", "Untap two target lands.", 2, 2, types.Land, untapCheck},
		{"bounce two creatures", "Return two target creatures to their owners' hands.", 2, 2, types.Creature, bounceCheck},
		{"bounce up to three creatures", "Return up to three target creatures to their owners' hands.", 0, 3, types.Creature, bounceCheck},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Multi Verb",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %#v, want one spec", mode.Targets)
			}
			spec := mode.Targets[0]
			if spec.MinTargets != test.minTargets || spec.MaxTargets != test.maxTargets {
				t.Fatalf("cardinality = {%d,%d}, want {%d,%d}", spec.MinTargets, spec.MaxTargets, test.minTargets, test.maxTargets)
			}
			if spec.Allow != game.TargetAllowPermanent {
				t.Fatalf("allow = %v, want TargetAllowPermanent", spec.Allow)
			}
			if test.permType != "" {
				if len(spec.Predicate.PermanentTypes) != 1 || spec.Predicate.PermanentTypes[0] != test.permType {
					t.Fatalf("predicate types = %v, want [%v]", spec.Predicate.PermanentTypes, test.permType)
				}
			}
			if len(mode.Sequence) != test.maxTargets {
				t.Fatalf("sequence len = %d, want %d", len(mode.Sequence), test.maxTargets)
			}
			for i := range mode.Sequence {
				test.check(t, i, mode.Sequence[i].Primitive)
			}
		})
	}
}

// TestLowerMultiTargetPermanentSingleTargetUnchanged proves the single-target
// destroy, tap, untap, and bounce paths are untouched: each still lowers to one
// spec with one verb instruction addressing TargetPermanentReference(0).
func TestLowerMultiTargetPermanentSingleTargetUnchanged(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		check      func(t *testing.T, prim game.Primitive)
	}{
		{"destroy", "Destroy target creature.", func(t *testing.T, prim game.Primitive) {
			if d, ok := prim.(game.Destroy); !ok || d.Object != game.TargetPermanentReference(0) {
				t.Fatalf("prim = %#v, want Destroy of TargetPermanentReference(0)", prim)
			}
		}},
		{"tap", "Tap target creature.", func(t *testing.T, prim game.Primitive) {
			if p, ok := prim.(game.Tap); !ok || p.Object != game.TargetPermanentReference(0) {
				t.Fatalf("prim = %#v, want Tap of TargetPermanentReference(0)", prim)
			}
		}},
		{"bounce", "Return target creature to its owner's hand.", func(t *testing.T, prim game.Primitive) {
			if p, ok := prim.(game.Bounce); !ok || p.Object != game.TargetPermanentReference(0) {
				t.Fatalf("prim = %#v, want Bounce of TargetPermanentReference(0)", prim)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Single Verb",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 || mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
				t.Fatalf("targets = %#v, want one {1,1} spec", mode.Targets)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
			}
			test.check(t, mode.Sequence[0].Primitive)
		})
	}
}

// TestLowerMultiTargetPermanentFailClosed proves shapes the executable backend
// cannot represent exactly stay rejected with a diagnostic and no partial card.
func TestLowerMultiTargetPermanentFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{"destroy subtype qualifier", "Destroy up to two target Goblin creatures."},
		{"destroy tapped qualifier", "Destroy two target tapped creatures."},
		{"destroy unbounded any number", "Destroy any number of target creatures."},
		{"tap attacking qualifier", "Tap two target attacking creatures."},
		{"bounce subtype qualifier", "Return up to two target Goblin creatures to their owners' hands."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Reject Verb",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

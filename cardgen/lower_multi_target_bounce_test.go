package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerMultiTargetBounceExcludedTypeAndUpToOne proves the bounce paths the
// parser-exactness widening unlocks: a multi-target excluded-type bounce ("up to
// two target nonland permanents") and the optional single-slot bounce ("up to
// one target ... to its owner's hand"), including its "other" and "you control"
// qualifiers. Each lowers to one multi-target spec carrying the cardinality
// range, the predicate, and one Bounce per slot addressing its own target index.
func TestLowerMultiTargetBounceExcludedTypeAndUpToOne(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		minTargets int
		maxTargets int
		permTypes  []types.Card
		excluded   []types.Card
		controller game.ControllerRelation
		another    bool
	}{
		{
			name:       "up to two nonland permanents",
			oracleText: "Return up to two target nonland permanents to their owners' hands.",
			minTargets: 0, maxTargets: 2,
			excluded: []types.Card{types.Land},
		},
		{
			name:       "six nonland permanents",
			oracleText: "Return six target nonland permanents to their owners' hands.",
			minTargets: 6, maxTargets: 6,
			excluded: []types.Card{types.Land},
		},
		{
			name:       "up to one creature its owner",
			oracleText: "Return up to one target creature to its owner's hand.",
			minTargets: 0, maxTargets: 1,
			permTypes: []types.Card{types.Creature},
		},
		{
			name:       "up to one nonland permanent its owner",
			oracleText: "Return up to one target nonland permanent to its owner's hand.",
			minTargets: 0, maxTargets: 1,
			excluded: []types.Card{types.Land},
		},
		{
			name:       "up to one other permanent you control its owner",
			oracleText: "Return up to one other target permanent you control to its owner's hand.",
			minTargets: 0, maxTargets: 1,
			controller: game.ControllerYou,
			another:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Multi Bounce",
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
			if !cardSlicesEqual(spec.Predicate.PermanentTypes, test.permTypes) {
				t.Fatalf("permanent types = %v, want %v", spec.Predicate.PermanentTypes, test.permTypes)
			}
			if !cardSlicesEqual(spec.Predicate.ExcludedTypes, test.excluded) {
				t.Fatalf("excluded types = %v, want %v", spec.Predicate.ExcludedTypes, test.excluded)
			}
			if spec.Predicate.Controller != test.controller {
				t.Fatalf("controller = %v, want %v", spec.Predicate.Controller, test.controller)
			}
			if spec.Predicate.Another != test.another {
				t.Fatalf("another = %v, want %v", spec.Predicate.Another, test.another)
			}
			if len(mode.Sequence) != test.maxTargets {
				t.Fatalf("sequence len = %d, want %d", len(mode.Sequence), test.maxTargets)
			}
			for i := range mode.Sequence {
				p, ok := mode.Sequence[i].Primitive.(game.Bounce)
				if !ok || p.Object != game.TargetPermanentReference(i) {
					t.Fatalf("sequence[%d] = %#v, want Bounce of TargetPermanentReference(%d)", i, mode.Sequence[i].Primitive, i)
				}
			}
		})
	}
}

// TestLowerUpToOneTargetTappedBounce proves the "up to one target <qualified>"
// optional single-target form reuses the mandatory single-target reconstruction
// for qualifiers (here a tapped-state filter) the plural multi-target form does
// not express, both as a stand-alone spell and as the enters-the-battlefield
// trigger printed on Peerless Ropemaster. Each lowers to one optional permanent
// target (MinTargets 0, MaxTargets 1) carrying the tapped predicate and a single
// Bounce addressing target index 0.
func TestLowerUpToOneTargetTappedBounce(t *testing.T) {
	t.Parallel()
	assertTappedBounce := func(t *testing.T, mode game.Mode) {
		t.Helper()
		if len(mode.Targets) != 1 {
			t.Fatalf("targets = %#v, want one spec", mode.Targets)
		}
		spec := mode.Targets[0]
		if spec.MinTargets != 0 || spec.MaxTargets != 1 {
			t.Fatalf("cardinality = {%d,%d}, want {0,1}", spec.MinTargets, spec.MaxTargets)
		}
		if spec.Allow != game.TargetAllowPermanent {
			t.Fatalf("allow = %v, want TargetAllowPermanent", spec.Allow)
		}
		if !cardSlicesEqual(spec.Predicate.PermanentTypes, []types.Card{types.Creature}) {
			t.Fatalf("permanent types = %v, want [Creature]", spec.Predicate.PermanentTypes)
		}
		if spec.Predicate.Tapped != game.TriTrue {
			t.Fatalf("tapped = %v, want TriTrue", spec.Predicate.Tapped)
		}
		if len(mode.Sequence) != 1 {
			t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
		}
		p, ok := mode.Sequence[0].Primitive.(game.Bounce)
		if !ok || p.Object != game.TargetPermanentReference(0) {
			t.Fatalf("sequence[0] = %#v, want Bounce of TargetPermanentReference(0)", mode.Sequence[0].Primitive)
		}
	}

	t.Run("spell", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Tapped Bounce",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Return up to one target tapped creature to its owner's hand.",
		})
		assertTappedBounce(t, face.SpellAbility.Val.Modes[0])
	})

	t.Run("enters trigger", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Peerless Ropemaster",
			Layout:     "normal",
			TypeLine:   "Creature — Human Rogue",
			Power:      new("2"),
			Toughness:  new("2"),
			OracleText: "When this creature enters, return up to one target tapped creature to its owner's hand.",
		})
		if len(face.TriggeredAbilities) != 1 {
			t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
		}
		assertTappedBounce(t, face.TriggeredAbilities[0].Content.Modes[0])
	})
}

func cardSlicesEqual(got, want []types.Card) bool {
	return slices.Equal(got, want)
}

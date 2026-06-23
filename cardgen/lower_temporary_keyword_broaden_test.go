package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLowerTemporaryKeywordGrantBroadTargets covers the broadened temporary
// keyword grant target filter: the subject may now be any permanent target the
// canonical permanentTargetSpec already accepts for destroy/exile/tap — a bare
// subtype noun, a non-creature card type, a color, or a tapped qualifier — not
// only a plain "creature"/"permanent" target. Each lowers to one
// until-end-of-turn ApplyContinuous keyword grant on the target.
func TestLowerTemporaryKeywordGrantBroadTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracle       string
		wantKeywords []game.Keyword
		wantTypes    []types.Card
		wantSubtypes []types.Sub
	}{
		{
			name:         "bare subtype noun target",
			oracle:       "Target Human gains double strike until end of turn.",
			wantKeywords: []game.Keyword{game.DoubleStrike},
			wantSubtypes: []types.Sub{types.Sub("Human")},
		},
		{
			name:         "artifact card-type target",
			oracle:       "Target artifact gains indestructible until end of turn.",
			wantKeywords: []game.Keyword{game.Indestructible},
			wantTypes:    []types.Card{types.Artifact},
		},
		{
			name:         "color-filtered creature target",
			oracle:       "Target black creature gains flying until end of turn.",
			wantKeywords: []game.Keyword{game.Flying},
			wantTypes:    []types.Card{types.Creature},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Broad Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			if mode.Targets[0].MinTargets != 1 || mode.Targets[0].MaxTargets != 1 {
				t.Fatalf("cardinality = [%d,%d], want [1,1]", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
			}
			if tc.wantTypes != nil && !reflect.DeepEqual(mode.Targets[0].Predicate.PermanentTypes, tc.wantTypes) {
				t.Fatalf("permanent types = %v, want %v", mode.Targets[0].Predicate.PermanentTypes, tc.wantTypes)
			}
			if tc.wantSubtypes != nil && !reflect.DeepEqual(mode.Targets[0].Predicate.Subtypes, tc.wantSubtypes) {
				t.Fatalf("subtypes = %v, want %v", mode.Targets[0].Predicate.Subtypes, tc.wantSubtypes)
			}
			if len(mode.Sequence) != 1 {
				t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
			}
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
				t.Fatalf("object = %#v, want target permanent 0", apply.Object)
			}
			if apply.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", apply.Duration)
			}
			effect := apply.ContinuousEffects[0]
			if effect.Layer != game.LayerAbility ||
				!reflect.DeepEqual(effect.AddKeywords, tc.wantKeywords) {
				t.Fatalf("continuous effect = %+v, want %v keyword grant", effect, tc.wantKeywords)
			}
		})
	}
}

// TestLowerTemporaryKeywordGrantOptionalTarget covers the broadened optional
// ("up to one") target keyword grant. The single TargetSpec carries the [0,1]
// cardinality and one ApplyContinuous slot addresses the chosen target; a
// declined slot leaves the unresolved index the runtime no-ops.
func TestLowerTemporaryKeywordGrantOptionalTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Grant",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Up to one target creature gains first strike and vigilance until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if mode.Targets[0].MinTargets != 0 || mode.Targets[0].MaxTargets != 1 {
		t.Fatalf("cardinality = [%d,%d], want [0,1]", mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
		t.Fatalf("object = %#v, want target permanent 0", apply.Object)
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility ||
		!reflect.DeepEqual(effect.AddKeywords, []game.Keyword{game.FirstStrike, game.Vigilance}) {
		t.Fatalf("continuous effect = %+v, want first strike + vigilance grant", effect)
	}
}

// TestLowerTemporaryKeywordLossBroadTargets covers the broadened temporary
// keyword loss target filter: like the grant path, a targeted "loses" spell now
// accepts any permanent target permanentTargetSpec accepts (subtype noun,
// conjunctive card types, tapped qualifier), each lowering to one
// until-end-of-turn ApplyContinuous keyword removal on the target.
func TestLowerTemporaryKeywordLossBroadTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracle       string
		wantKeywords []game.Keyword
		wantSubtypes []types.Sub
		wantTypes    []types.Card
		wantTapped   bool
	}{
		{
			name:         "bare subtype noun target",
			oracle:       "Target Human loses flying until end of turn.",
			wantKeywords: []game.Keyword{game.Flying},
			wantSubtypes: []types.Sub{types.Sub("Human")},
		},
		{
			name:         "conjunctive card-type target",
			oracle:       "Target artifact creature loses indestructible until end of turn.",
			wantKeywords: []game.Keyword{game.Indestructible},
			wantTypes:    []types.Card{types.Artifact, types.Creature},
		},
		{
			name:         "tapped creature target",
			oracle:       "Target tapped creature loses flying until end of turn.",
			wantKeywords: []game.Keyword{game.Flying},
			wantTypes:    []types.Card{types.Creature},
			wantTapped:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Broad Loss",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			pred := mode.Targets[0].Predicate
			if tc.wantSubtypes != nil && !reflect.DeepEqual(pred.Subtypes, tc.wantSubtypes) {
				t.Fatalf("subtypes = %v, want %v", pred.Subtypes, tc.wantSubtypes)
			}
			if tc.wantTypes != nil && !reflect.DeepEqual(pred.PermanentTypes, tc.wantTypes) {
				t.Fatalf("permanent types = %v, want %v", pred.PermanentTypes, tc.wantTypes)
			}
			if tc.wantTapped && pred.Tapped != game.TriTrue {
				t.Fatalf("tapped = %v, want true", pred.Tapped)
			}
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
				t.Fatalf("object = %#v, want target permanent 0", apply.Object)
			}
			if apply.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", apply.Duration)
			}
			effect := apply.ContinuousEffects[0]
			if effect.Layer != game.LayerAbility ||
				!reflect.DeepEqual(effect.RemoveKeywords, tc.wantKeywords) {
				t.Fatalf("continuous effect = %+v, want %v keyword loss", effect, tc.wantKeywords)
			}
		})
	}
}

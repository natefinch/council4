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
			if tc.wantTypes != nil && !reflect.DeepEqual(mode.Targets[0].Selection.Val.RequiredTypesAny, tc.wantTypes) {
				t.Fatalf("permanent types = %v, want %v", mode.Targets[0].Selection.Val.RequiredTypesAny, tc.wantTypes)
			}
			if tc.wantSubtypes != nil && !reflect.DeepEqual(mode.Targets[0].Selection.Val.SubtypesAny, tc.wantSubtypes) {
				t.Fatalf("subtypes = %v, want %v", mode.Targets[0].Selection.Val.SubtypesAny, tc.wantSubtypes)
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
		name            string
		oracle          string
		wantKeywords    []game.Keyword
		wantSubtypes    []types.Sub
		wantTypes       []types.Card
		wantConjunctive bool
		wantTapped      bool
	}{
		{
			name:         "bare subtype noun target",
			oracle:       "Target Human loses flying until end of turn.",
			wantKeywords: []game.Keyword{game.Flying},
			wantSubtypes: []types.Sub{types.Sub("Human")},
		},
		{
			name:            "conjunctive card-type target",
			oracle:          "Target artifact creature loses indestructible until end of turn.",
			wantKeywords:    []game.Keyword{game.Indestructible},
			wantTypes:       []types.Card{types.Artifact, types.Creature},
			wantConjunctive: true,
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
			pred := mode.Targets[0].Selection.Val
			if tc.wantSubtypes != nil && !reflect.DeepEqual(pred.SubtypesAny, tc.wantSubtypes) {
				t.Fatalf("subtypes = %v, want %v", pred.SubtypesAny, tc.wantSubtypes)
			}
			if tc.wantTypes != nil {
				got := pred.RequiredTypesAny
				if tc.wantConjunctive {
					got = pred.RequiredTypes
				}
				if !reflect.DeepEqual(got, tc.wantTypes) {
					t.Fatalf("permanent types = %v, want %v", got, tc.wantTypes)
				}
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

// TestLowerTemporaryLandwalkGrant covers the broadened temporary keyword grant
// for the landwalk evasion family (CR 702.14). Landwalk is parameterized by a
// land subtype, so it lowers to a granted LandwalkKeyword static-ability body via
// AddAbilities — exactly like the permanent forestwalk grant — carried for the
// until-end-of-turn duration, rather than a simple AddKeywords enum value.
func TestLowerTemporaryLandwalkGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		oracle  string
		subtype types.Sub
	}{
		{"forestwalk", "Target creature gains forestwalk until end of turn.", types.Forest},
		{"islandwalk", "Target creature gains islandwalk until end of turn.", types.Island},
		{"swampwalk", "Target creature gains swampwalk until end of turn.", types.Swamp},
		{"mountainwalk", "Target creature gains mountainwalk until end of turn.", types.Mountain},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Landwalk Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
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
			if effect.Layer != game.LayerAbility {
				t.Fatalf("layer = %v, want ability", effect.Layer)
			}
			if len(effect.AddKeywords) != 0 {
				t.Fatalf("keywords = %v, want none (landwalk is ability-backed)", effect.AddKeywords)
			}
			if len(effect.AddAbilities) != 1 {
				t.Fatalf("abilities = %d, want 1 granted landwalk ability", len(effect.AddAbilities))
			}
			static, ok := effect.AddAbilities[0].(*game.StaticAbility)
			if !ok {
				t.Fatalf("ability = %T, want *game.StaticAbility", effect.AddAbilities[0])
			}
			landwalk, ok := game.StaticBodyLandwalkKeyword(static)
			if !ok || landwalk.Subtype != tc.subtype {
				t.Fatalf("landwalk = %+v ok=%v, want subtype %v", landwalk, ok, tc.subtype)
			}
		})
	}
}

// TestLowerTemporarySimpleCombatKeywordGrant covers the broadened temporary
// keyword grant for the combat keywords the runtime already models as simple
// continuous keywords but the parser previously refused (infect, wither,
// horsemanship, skulk). Each lowers to a single AddKeywords enum value granted
// until end of turn, like the established trample/flying grants.
func TestLowerTemporarySimpleCombatKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		oracle  string
		keyword game.Keyword
	}{
		{"infect", "Target creature gains infect until end of turn.", game.Infect},
		{"wither", "Target creature gains wither until end of turn.", game.Wither},
		{"horsemanship", "Target creature gains horsemanship until end of turn.", game.Horsemanship},
		{"skulk", "Target creature gains skulk until end of turn.", game.Skulk},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Combat Keyword Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("duration = %v, want until end of turn", apply.Duration)
			}
			effect := apply.ContinuousEffects[0]
			if len(effect.AddAbilities) != 0 {
				t.Fatalf("abilities = %v, want none", effect.AddAbilities)
			}
			if !reflect.DeepEqual(effect.AddKeywords, []game.Keyword{tc.keyword}) {
				t.Fatalf("keywords = %v, want [%v]", effect.AddKeywords, tc.keyword)
			}
		})
	}
}

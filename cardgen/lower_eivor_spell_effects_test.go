package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestLowerMultiTargetTemporaryKeywordGrant covers the broadened multi-target
// "each" distribution of an until-end-of-turn keyword grant: a plural ("two
// target creatures each gain ...") or optional-multi ("up to two target creatures
// each gain ...") subject grants the keyword to every chosen target. Each lowers
// to one ApplyContinuous per target slot, mirroring the single-target form.
func TestLowerMultiTargetTemporaryKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		minTargets int
		maxTargets int
		keyword    game.Keyword
	}{
		{
			name:       "up to two you control gain lifelink",
			oracle:     "Up to two target creatures you control each gain lifelink until end of turn.",
			minTargets: 0,
			maxTargets: 2,
			keyword:    game.Lifelink,
		},
		{
			name:       "two target creatures gain trample",
			oracle:     "Two target creatures each gain trample until end of turn.",
			minTargets: 2,
			maxTargets: 2,
			keyword:    game.Trample,
		},
		{
			name:       "up to three target creatures gain flying",
			oracle:     "Up to three target creatures each gain flying until end of turn.",
			minTargets: 0,
			maxTargets: 3,
			keyword:    game.Flying,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Multi Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			if mode.Targets[0].MinTargets != tc.minTargets || mode.Targets[0].MaxTargets != tc.maxTargets {
				t.Fatalf("cardinality = [%d,%d], want [%d,%d]",
					mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets, tc.minTargets, tc.maxTargets)
			}
			if len(mode.Sequence) != tc.maxTargets {
				t.Fatalf("sequence = %d instructions, want %d", len(mode.Sequence), tc.maxTargets)
			}
			for i := range mode.Sequence {
				apply, ok := mode.Sequence[i].Primitive.(game.ApplyContinuous)
				if !ok {
					t.Fatalf("instruction %d = %T, want game.ApplyContinuous", i, mode.Sequence[i].Primitive)
				}
				if apply.Object != opt.Val(game.TargetPermanentReference(i)) {
					t.Fatalf("instruction %d object = %#v, want target permanent %d", i, apply.Object, i)
				}
				if apply.Duration != game.DurationUntilEndOfTurn {
					t.Fatalf("instruction %d duration = %v, want until end of turn", i, apply.Duration)
				}
				if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, []game.Keyword{tc.keyword}) {
					t.Fatalf("instruction %d keywords = %v, want %v",
						i, apply.ContinuousEffects[0].AddKeywords, tc.keyword)
				}
			}
		})
	}
}

// TestLowerControlledSourceKeywordGrant covers the broadened keyword grant whose
// duration is tied to the source's lifetime: "for as long as you control this
// <noun>" lasts as long as the source stays under its controller's control
// (Aegis Angel's enters trigger, Tale of Tinúviel chapter I). Each lowers to one
// ApplyContinuous keyword grant on the single target with the
// DurationForAsLongAsYouControlSource source-lifetime duration. The grant is
// exercised in an enters trigger, the verified resolving context for the form.
func TestLowerControlledSourceKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeLine string
		oracle   string
		keyword  game.Keyword
	}{
		{
			name:     "another target permanent gains indestructible (Aegis Angel)",
			typeLine: "Creature — Angel",
			oracle:   "When this creature enters, another target permanent gains indestructible for as long as you control this creature.",
			keyword:  game.Indestructible,
		},
		{
			name:     "target creature you control gains flying",
			typeLine: "Creature — Bird",
			oracle:   "When this creature enters, target creature you control gains flying for as long as you control this creature.",
			keyword:  game.Flying,
		},
		{
			name:     "target creature gains hexproof on an enchantment",
			typeLine: "Enchantment",
			oracle:   "When this enchantment enters, target creature you control gains hexproof for as long as you control this enchantment.",
			keyword:  game.Hexproof,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Controlled Source Grant",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			if len(mode.Targets) != 1 || mode.Targets[0].MaxTargets != 1 {
				t.Fatalf("targets = %#v, want one single target", mode.Targets)
			}
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Object != opt.Val(game.TargetPermanentReference(0)) {
				t.Fatalf("object = %#v, want target permanent 0", apply.Object)
			}
			if apply.Duration != game.DurationForAsLongAsYouControlSource {
				t.Fatalf("duration = %v, want for as long as you control source", apply.Duration)
			}
			if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, []game.Keyword{tc.keyword}) {
				t.Fatalf("keywords = %v, want %v", apply.ContinuousEffects[0].AddKeywords, tc.keyword)
			}
		})
	}
}

// TestLowerSagaSplitExileReturn covers the split O-Ring across Saga chapters:
// one chapter exiles a target permanent and a later chapter returns "the exiled
// card to the battlefield" (The Princess Takes Flight chapters I and III). The
// exile must publish the exile-until-leaves linked key and the return must read
// it, linking the two chapters at runtime without synthesizing a leave trigger.
func TestLowerSagaSplitExileReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Princess Saga",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Exile up to one target creature.\n" +
			"II — Target creature you control gets +2/+2 and gains flying until end of turn.\n" +
			"III — Return the exiled card to the battlefield under its owner's control.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	exile, ok := face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Exile)
	if !ok {
		t.Fatalf("chapter I primitive = %T, want game.Exile", face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if exile.ExileLinkedKey != exileUntilLeavesKey {
		t.Fatalf("chapter I exile key = %q, want %q", exile.ExileLinkedKey, exileUntilLeavesKey)
	}
	put, ok := face.ChapterAbilities[2].Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("chapter III primitive = %T, want game.PutOnBattlefield", face.ChapterAbilities[2].Content.Modes[0].Sequence[0].Primitive)
	}
	linked, ok := put.Source.LinkedKey()
	if !ok || linked != exileUntilLeavesKey {
		t.Fatalf("chapter III return source = %#v, want linked %q", put.Source, exileUntilLeavesKey)
	}
	// No leave-the-battlefield return trigger is synthesized: the explicit
	// chapter III return already releases the exiled card.
	for i := range face.TriggeredAbilities {
		if face.TriggeredAbilities[i].Trigger.Pattern.Event == game.EventZoneChanged {
			t.Fatalf("unexpected synthesized leave trigger: %#v", face.TriggeredAbilities[i])
		}
	}
}

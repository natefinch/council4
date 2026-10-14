package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestReplacementProgramAllowsIntrinsicOriginMetadata(t *testing.T) {
	replacement := ReplacementEffect{
		Program:    ProgramReference{ID: "test/land-replacement", Version: "1"},
		MatchEvent: EventCardDrawn, FromSourceAbility: true,
	}
	if err := replacement.ValidateProgram(); err != nil {
		t.Fatalf("runtime provenance was mistaken for a replacement action: %v", err)
	}
}

func TestValidateBasicLandTypeSet(t *testing.T) {
	for _, test := range []struct {
		name    string
		layer   ContinuousLayer
		subtype types.Sub
		valid   bool
	}{
		{"Mountain", LayerType, types.Mountain, true},
		{"Forest", LayerType, types.Forest, true},
		{"ability layer", LayerAbility, types.Mountain, false},
		{"nonbasic subtype", LayerType, types.Locus, false},
		{"creature subtype", LayerType, types.Elf, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			issues := ValidateCardDef(&CardDef{CardFace: CardFace{
				Name: "Typed land SET", Types: []types.Card{types.Enchantment},
				StaticAbilities: []StaticAbility{{ContinuousEffects: []ContinuousEffect{{
					Layer: test.layer, AffectedSource: true, SetBasicLandTypes: []types.Sub{test.subtype},
				}}}},
			}})
			if (len(issues) == 0) != test.valid {
				t.Fatalf("valid=%v, issues=%+v", test.valid, issues)
			}
		})
	}
}

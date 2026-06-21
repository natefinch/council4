package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerGraveyardAnthemStatic proves the Incarnation cycle ("As long as this
// card is in your graveyard and you control a <land>, creatures you control have
// <keyword>") lowers to a static ability that functions from the graveyard, is
// gated by a controls-a-land condition, and grants the keyword to a controlled
// group rather than only the source.
func TestLowerGraveyardAnthemStatic(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		subtype    types.Sub
	}{
		"anger": {
			oracleText: "Haste\nAs long as this card is in your graveyard and you control a Mountain, creatures you control have haste.",
			subtype:    types.Sub("Mountain"),
		},
		"wonder": {
			oracleText: "Flying\nAs long as this card is in your graveyard and you control an Island, creatures you control have flying.",
			subtype:    types.Sub("Island"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Incarnation",
				Layout:     "normal",
				TypeLine:   "Creature — Incarnation",
				OracleText: test.oracleText,
				ManaCost:   "{2}{R}",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			var graveyard int
			found := false
			for i := range face.StaticAbilities {
				body := &face.StaticAbilities[i].Body
				if body.ZoneOfFunction != zone.Graveyard {
					continue
				}
				found = true
				graveyard++
				if !body.Condition.Exists {
					t.Fatalf("graveyard static missing condition: %#v", body)
				}
				matching := body.Condition.Val.ControlsMatching
				if !matching.Exists {
					t.Fatalf("ControlsMatching missing: %#v", body.Condition.Val)
				}
				if !slices.Equal(matching.Val.Selection.SubtypesAny, []types.Sub{test.subtype}) {
					t.Fatalf("control selection = %#v, want subtype %s", matching.Val.Selection, test.subtype)
				}
				if len(body.ContinuousEffects) == 0 {
					t.Fatalf("graveyard static missing continuous effects: %#v", body)
				}
				for j := range body.ContinuousEffects {
					if body.ContinuousEffects[j].AffectedSource {
						t.Fatalf("graveyard anthem must affect a controlled group, not the source: %#v", body.ContinuousEffects[j])
					}
				}
			}
			if !found {
				t.Fatalf("no graveyard-function static ability lowered: %#v", face.StaticAbilities)
			}
			if graveyard != 1 {
				t.Fatalf("graveyard static count = %d, want 1", graveyard)
			}
		})
	}
}

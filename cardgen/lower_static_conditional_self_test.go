package cardgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerSelfStatic lowers a single conditional self static creature face and
// returns its sole static ability.
func lowerSelfStatic(t *testing.T, oracleText string) game.StaticAbility {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Conditional Self",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: oracleText,
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	return face.StaticAbilities[0].Body
}

// TestLowerLeadingConditionSelfStatic proves that a leading "As long as ..."
// condition whose clause contains a player possession verb ("you have", "an
// opponent has") lowers the same self characteristic/keyword static the trailing
// form already produced, rather than tripping over a phantom keyword-grant
// effect.
func TestLowerLeadingConditionSelfStatic(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		condition  game.Condition
		power      int
		toughness  int
		keywords   []game.Keyword
	}{
		"leading life pt and keyword": {
			oracleText: "As long as you have 30 or more life, this creature gets +5/+5 and has flying.",
			condition:  game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 30}}},
			power:      5,
			toughness:  5,
			keywords:   []game.Keyword{game.Flying},
		},
		"trailing life pt and keyword": {
			oracleText: "This creature gets +5/+5 and has flying as long as you have 30 or more life.",
			condition:  game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 30}}},
			power:      5,
			toughness:  5,
			keywords:   []game.Keyword{game.Flying},
		},
		"leading hand size": {
			oracleText: "As long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike.",
			condition:  game.Condition{ControllerHandSizeAtLeast: 7},
			power:      2,
			toughness:  1,
			keywords:   []game.Keyword{game.FirstStrike},
		},
		"leading hand empty keyword only": {
			oracleText: "As long as you have no cards in hand, this creature has double strike.",
			condition:  game.Condition{ControllerHandEmpty: true},
			keywords:   []game.Keyword{game.DoubleStrike},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := lowerSelfStatic(t, test.oracleText)
			assertSelfCondition(t, &ability, &test.condition)
			assertSelfContinuous(t, &ability, test.power, test.toughness, test.keywords)
		})
	}
}

// TestLowerControlObjectConditionSelfStatic proves the richer "you control
// <object>" condition matchers (token, multicolored, typed subtype) lower onto
// the runtime ControlsMatching selection.
func TestLowerControlObjectConditionSelfStatic(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		selection  game.Selection
		power      int
		toughness  int
		keywords   []game.Keyword
	}{
		"token": {
			oracleText: "As long as you control a token, this creature gets +2/+0 and has trample.",
			selection:  game.Selection{TokenOnly: true},
			power:      2,
			keywords:   []game.Keyword{game.Trample},
		},
		"another multicolored permanent": {
			oracleText: "As long as you control another multicolored permanent, this creature gets +1/+1 and has first strike.",
			selection:  game.Selection{Multicolored: true, ExcludeSource: true},
			power:      1,
			toughness:  1,
			keywords:   []game.Keyword{game.FirstStrike},
		},
		"typed subtype creature": {
			oracleText: "As long as you control a Griffin creature, this creature gets +3/+3 and has flying.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				SubtypesAny:   []types.Sub{types.Griffin},
			},
			power:     3,
			toughness: 3,
			keywords:  []game.Keyword{game.Flying},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := lowerSelfStatic(t, test.oracleText)
			if !ability.Condition.Exists {
				t.Fatalf("condition missing: %#v", ability)
			}
			matching := ability.Condition.Val.ControlsMatching
			if !matching.Exists {
				t.Fatalf("ControlsMatching missing: %#v", ability.Condition.Val)
			}
			got := matching.Val.Selection
			if !selectionEqual(got, test.selection) {
				t.Fatalf("selection = %#v, want %#v", got, test.selection)
			}
			assertSelfContinuous(t, &ability, test.power, test.toughness, test.keywords)
		})
	}
}

func assertSelfCondition(t *testing.T, ability *game.StaticAbility, want *game.Condition) {
	t.Helper()
	if !ability.Condition.Exists {
		t.Fatalf("condition missing: %#v", ability)
	}
	got := ability.Condition.Val
	got.Text = ""
	if !reflect.DeepEqual(got, *want) {
		t.Fatalf("condition = %#v, want %#v", got, want)
	}
}

func assertSelfContinuous(t *testing.T, ability *game.StaticAbility, power, toughness int, keywords []game.Keyword) {
	t.Helper()
	var gotPower, gotToughness int
	var gotKeywords []game.Keyword
	for index := range ability.ContinuousEffects {
		effect := &ability.ContinuousEffects[index]
		if !effect.AffectedSource {
			t.Fatalf("continuous effect not source-affecting: %#v", effect)
		}
		switch effect.Layer {
		case game.LayerPowerToughnessModify:
			gotPower += effect.PowerDelta
			gotToughness += effect.ToughnessDelta
		case game.LayerAbility:
			gotKeywords = append(gotKeywords, effect.AddKeywords...)
		default:
			t.Fatalf("unexpected layer: %#v", effect)
		}
	}
	if gotPower != power || gotToughness != toughness {
		t.Fatalf("p/t delta = %d/%d, want %d/%d", gotPower, gotToughness, power, toughness)
	}
	if !keywordsEqual(gotKeywords, keywords) {
		t.Fatalf("keywords = %#v, want %#v", gotKeywords, keywords)
	}
}

func keywordsEqual(got, want []game.Keyword) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// selectionEqual compares the selection fields exercised by these tests,
// treating nil and empty slices as equal.
func selectionEqual(got, want game.Selection) bool {
	return slices.Equal(got.RequiredTypes, want.RequiredTypes) &&
		slices.Equal(got.SubtypesAny, want.SubtypesAny) &&
		slices.Equal(got.ColorsAny, want.ColorsAny) &&
		slices.Equal(got.Supertypes, want.Supertypes) &&
		got.Colorless == want.Colorless &&
		got.Multicolored == want.Multicolored &&
		got.TokenOnly == want.TokenOnly &&
		got.ExcludeSource == want.ExcludeSource &&
		got.Tapped == want.Tapped
}

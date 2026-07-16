package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestRenderDevotionNotCreatureSourceKeepsColors proves the source-generation
// backend renders the Theros God devotion static faithfully: the generated Go
// source carries the devotion aggregate, its red color, the five threshold, and
// the creature-type removal at the type layer. This guards the natural next step
// of adding a God to the curated card set, where the Colors field must survive
// rendering rather than being silently dropped.
func TestRenderDevotionNotCreatureSourceKeepsColors(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Purphoros, God of the Forge",
		Layout:   "normal",
		ManaCost: "{3}{R}",
		TypeLine: "Legendary Enchantment Creature — God",
		OracleText: "Indestructible\n" +
			"As long as your devotion to red is less than five, Purphoros isn't a creature.\n" +
			"Whenever another creature you control enters, Purphoros deals 2 damage to each opponent.\n" +
			"{2}{R}: Creatures you control get +1/+0 until end of turn.",
		Colors:    []string{"R"},
		Power:     new("6"),
		Toughness: new("5"),
	}, "purphoros")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.AggregateControllerDevotion",
		"[]color.Color{color.Red}",
		"compare.LessThan",
		"game.LayerType",
		"[]types.Card{types.Creature}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// findDevotionNotCreatureStatic returns the sole static ability whose condition
// is a devotion comparison and whose continuous effect removes the creature
// type, failing the test if the face does not carry exactly that shape.
func findDevotionNotCreatureStatic(t *testing.T, face loweredFaceAbilities) game.StaticAbility {
	t.Helper()
	var found []game.StaticAbility
	for _, static := range face.StaticAbilities {
		body := static.Body
		if !body.Condition.Exists || len(body.Condition.Val.Aggregates) != 1 {
			continue
		}
		if body.Condition.Val.Aggregates[0].Aggregate != game.AggregateControllerDevotion {
			continue
		}
		found = append(found, body)
	}
	if len(found) != 1 {
		t.Fatalf("devotion-not-creature statics = %d, want 1 (statics: %#v)", len(found), face.StaticAbilities)
	}
	return found[0]
}

func assertRemovesCreatureType(t *testing.T, body game.StaticAbility) {
	t.Helper()
	if len(body.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(body.ContinuousEffects))
	}
	effect := body.ContinuousEffects[0]
	if effect.Layer != game.LayerType {
		t.Fatalf("layer = %v, want LayerType", effect.Layer)
	}
	if !effect.AffectedSource {
		t.Fatal("effect must affect the source permanent")
	}
	if len(effect.RemoveTypes) != 1 || effect.RemoveTypes[0] != types.Creature {
		t.Fatalf("RemoveTypes = %v, want [Creature]", effect.RemoveTypes)
	}
	if len(effect.AddTypes) != 0 || len(effect.SetTypes) != 0 {
		t.Fatalf("effect must only remove the creature type, got Add=%v Set=%v", effect.AddTypes, effect.SetTypes)
	}
}

func assertDevotionCondition(t *testing.T, body game.StaticAbility, colors []color.Color, threshold int) {
	t.Helper()
	agg := body.Condition.Val.Aggregates[0]
	if agg.Op != compare.LessThan {
		t.Fatalf("op = %v, want LessThan", agg.Op)
	}
	if agg.Value != threshold {
		t.Fatalf("threshold = %d, want %d", agg.Value, threshold)
	}
	if len(agg.Colors) != len(colors) {
		t.Fatalf("colors = %v, want %v", agg.Colors, colors)
	}
	for i, c := range colors {
		if agg.Colors[i] != c {
			t.Fatalf("colors = %v, want %v", agg.Colors, colors)
		}
	}
}

// TestLowerDevotionNotCreaturePurphoros proves the whole Purphoros, God of the
// Forge card lowers with no diagnostics: the devotion-gated "isn't a creature"
// static removes the creature type while red devotion is below five, and its ETB
// damage trigger and pump activated ability lower alongside it.
func TestLowerDevotionNotCreaturePurphoros(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Purphoros, God of the Forge",
		Layout:   "normal",
		ManaCost: "{3}{R}",
		TypeLine: "Legendary Enchantment Creature — God",
		OracleText: "Indestructible\n" +
			"As long as your devotion to red is less than five, Purphoros isn't a creature.\n" +
			"Whenever another creature you control enters, Purphoros deals 2 damage to each opponent.\n" +
			"{2}{R}: Creatures you control get +1/+0 until end of turn.",
		Colors:    []string{"R"},
		Power:     new("6"),
		Toughness: new("5"),
	})
	body := findDevotionNotCreatureStatic(t, face)
	assertDevotionCondition(t, body, []color.Color{color.Red}, 5)
	assertRemovesCreatureType(t, body)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1 (ETB damage)", len(face.TriggeredAbilities))
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1 (pump)", len(face.ActivatedAbilities))
	}
}

// TestLowerDevotionNotCreatureTwoColor proves the two-color God wording ("your
// devotion to white and black is less than seven") lowers with both colors and
// the seven threshold. It uses a minimal two-color god so the assertion isolates
// the devotion static from unrelated abilities of any specific real card.
func TestLowerDevotionNotCreatureTwoColor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Testros, God of Trials",
		Layout:   "normal",
		ManaCost: "{1}{W}{B}",
		TypeLine: "Legendary Enchantment Creature — God",
		OracleText: "Indestructible\n" +
			"As long as your devotion to white and black is less than seven, Testros isn't a creature.",
		Colors:    []string{"W", "B"},
		Power:     new("3"),
		Toughness: new("5"),
	})
	body := findDevotionNotCreatureStatic(t, face)
	assertDevotionCondition(t, body, []color.Color{color.White, color.Black}, 7)
	assertRemovesCreatureType(t, body)
}

// TestLowerDevotionNotCreatureFailsClosed proves near-miss wordings do not lower
// to the devotion-not-creature static. A different type-removal ("isn't an
// artifact"), a non-self subject, and a non-devotion gate must each fail closed
// so the compiler never overfits to devotion wording it does not fully model.
func TestLowerDevotionNotCreatureFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "not an artifact instead of creature",
			oracleText: "As long as your devotion to red is less than five, Nixon isn't an artifact.",
		},
		{
			name:       "non-self subject",
			oracleText: "As long as your devotion to red is less than five, that creature isn't a creature.",
		},
		{
			name:       "greater than instead of less than",
			oracleText: "As long as your devotion to red is greater than five, Nixon isn't a creature.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Nixon",
				Layout:     "normal",
				ManaCost:   "{3}{R}",
				TypeLine:   "Legendary Enchantment Creature — God",
				OracleText: tc.oracleText,
				Colors:     []string{"R"},
				Power:      new("6"),
				Toughness:  new("5"),
			})
			for _, static := range face.StaticAbilities {
				if static.Body.Condition.Exists && len(static.Body.Condition.Val.Aggregates) == 1 &&
					static.Body.Condition.Val.Aggregates[0].Aggregate == game.AggregateControllerDevotion {
					t.Fatalf("near-miss wording lowered to a devotion-not-creature static: %q", tc.oracleText)
				}
			}
		})
	}
}

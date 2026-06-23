package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerSummonMagusSisters verifies the Final Fantasy Saga "Summon: Magus
// Sisters" lowers end to end. Its grouped chapter "I, II, III — Choose one at
// random —" is a saga chapter whose body is a modal ability with the random
// mode-selection primitive: one chapter ability covering chapters 1, 2, and 3,
// whose content carries RandomModes and three flavor-named options (Combine
// Powers!, Defense!, Fight!). Haste lowers as a static keyword.
func TestLowerSummonMagusSisters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Summon: Magus Sisters",
		Layout:   "saga",
		TypeLine: "Enchantment Creature — Saga Faerie",
		ManaCost: "{4}{G}",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II, III — Choose one at random —\n" +
			"• Combine Powers! — Put three +1/+1 counters on target creature.\n" +
			"• Defense! — Put a shield counter on target creature. You gain 3 life.\n" +
			"• Fight! — This creature fights up to one target creature an opponent controls.\n" +
			"Haste",
		Power:     new("5"),
		Toughness: new("5"),
	})

	if len(face.ChapterAbilities) != 1 {
		t.Fatalf("got %d chapter abilities, want 1", len(face.ChapterAbilities))
	}
	chapter := face.ChapterAbilities[0]
	if !slices.Equal(chapter.Chapters, []int{1, 2, 3}) {
		t.Fatalf("chapter numbers = %v, want [1 2 3]", chapter.Chapters)
	}
	content := chapter.Content
	if !content.RandomModes {
		t.Error("chapter content RandomModes = false, want true")
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("mode range = [%d,%d], want [1,1]", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 3 {
		t.Fatalf("got %d modes, want 3", len(content.Modes))
	}

	// Combine Powers!: three +1/+1 counters on a single target creature.
	combine := content.Modes[0]
	if len(combine.Targets) != 1 {
		t.Fatalf("Combine Powers targets = %d, want 1", len(combine.Targets))
	}
	addCounter, ok := combine.Sequence[0].Primitive.(game.AddCounter)
	if !ok || addCounter.CounterKind != counter.PlusOnePlusOne || addCounter.Amount != game.Fixed(3) {
		t.Fatalf("Combine Powers primitive = %+v, want AddCounter +1/+1 x3", combine.Sequence[0].Primitive)
	}

	// Defense!: a shield counter on a target creature plus three life gained.
	defense := content.Modes[1]
	if len(defense.Targets) != 1 {
		t.Fatalf("Defense targets = %d, want 1", len(defense.Targets))
	}
	if len(defense.Sequence) != 2 {
		t.Fatalf("Defense sequence length = %d, want 2", len(defense.Sequence))
	}
	shield, ok := defense.Sequence[0].Primitive.(game.AddCounter)
	if !ok || shield.CounterKind != counter.Shield {
		t.Fatalf("Defense primitive[0] = %+v, want AddCounter shield", defense.Sequence[0].Primitive)
	}
	gain, ok := defense.Sequence[1].Primitive.(game.GainLife)
	if !ok || gain.Amount != game.Fixed(3) {
		t.Fatalf("Defense primitive[1] = %+v, want GainLife 3", defense.Sequence[1].Primitive)
	}

	// Fight!: this creature fights up to one opponent-controlled creature, so
	// the single target is optional (MinTargets 0).
	fightMode := content.Modes[2]
	if len(fightMode.Targets) != 1 || fightMode.Targets[0].MinTargets != 0 || fightMode.Targets[0].MaxTargets != 1 {
		t.Fatalf("Fight targets = %+v, want one optional target", fightMode.Targets)
	}
	if _, ok := fightMode.Sequence[0].Primitive.(game.Fight); !ok {
		t.Fatalf("Fight primitive = %T, want game.Fight", fightMode.Sequence[0].Primitive)
	}

	if len(face.StaticAbilities) != 1 || face.StaticAbilities[0].VarName != "game.HasteStaticBody" {
		t.Fatalf("static abilities = %+v, want one Haste static", face.StaticAbilities)
	}
}

package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSagaChapterFlavorNames verifies that a Final Fantasy "Summon:" Saga
// whose chapters carry a flavor-name prefix ("I — Gungnir — Destroy ...") lowers
// every chapter as if the flavor name were absent: the parser strips the
// Title-Case proper name so the effect body classifies normally. It exercises
// all three of Summon: Primal Odin's chapter effects end to end — the
// controller-restricted destroy, the self ability grant, and the draw paired
// with symmetric life loss.
func TestLowerSagaChapterFlavorNames(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Summon: Primal Odin",
		Layout:   "saga",
		TypeLine: "Enchantment Creature — Saga Knight",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Gungnir — Destroy target creature an opponent controls.\n" +
			"II — Zantetsuken — This creature gains \"Whenever this creature deals combat damage to a player, that player loses the game.\"\n" +
			"III — Hall of Sorrow — Draw two cards. Each player loses 2 life.",
		Power:     new("5"),
		Toughness: new("3"),
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("got %d chapter abilities, want 3", len(face.ChapterAbilities))
	}
	if !slices.Equal(face.ChapterAbilities[0].Chapters, []int{1}) ||
		!slices.Equal(face.ChapterAbilities[1].Chapters, []int{2}) ||
		!slices.Equal(face.ChapterAbilities[2].Chapters, []int{3}) {
		t.Fatalf("chapter numbers = %v, %v, %v",
			face.ChapterAbilities[0].Chapters,
			face.ChapterAbilities[1].Chapters,
			face.ChapterAbilities[2].Chapters)
	}

	if _, ok := face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("chapter I primitive = %T, want game.Destroy",
			face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}

	thirdSeq := face.ChapterAbilities[2].Content.Modes[0].Sequence
	if len(thirdSeq) != 2 {
		t.Fatalf("chapter III sequence length = %d, want 2", len(thirdSeq))
	}
	draw, ok := thirdSeq[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("chapter III primitive[0] = %T, want game.Draw", thirdSeq[0].Primitive)
	}
	if draw.Amount != game.Fixed(2) {
		t.Fatalf("chapter III draw amount = %#v, want 2", draw.Amount)
	}
	loss, ok := thirdSeq[1].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("chapter III primitive[1] = %T, want game.LoseLife", thirdSeq[1].Primitive)
	}
	if loss.Amount != game.Fixed(2) {
		t.Fatalf("chapter III life loss amount = %#v, want 2", loss.Amount)
	}
	if loss.PlayerGroup.Kind == game.PlayerGroupReferenceNone {
		t.Error("chapter III life loss PlayerGroup = none, want each player")
	}
}

// TestLowerChapterGrantsQuotedTriggeredAbility verifies that a chapter granting
// the source permanent a quoted triggered ability ("This creature gains
// \"Whenever this creature deals combat damage to a player, that player loses
// the game.\"") lowers to a permanent ApplyContinuous that adds the recursively
// lowered triggered ability at LayerAbility, and that the conferred ability's
// body lowers the inner "that player loses the game" to a PlayerLosesGame on the
// triggering player.
func TestLowerChapterGrantsQuotedTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Grant Saga",
		Layout:   "saga",
		TypeLine: "Enchantment Creature — Saga Knight",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II, III — This creature gains \"Whenever this creature deals combat damage to a player, that player loses the game.\"",
		Power:     new("5"),
		Toughness: new("3"),
	})
	if len(face.ChapterAbilities) != 1 {
		t.Fatalf("got %d chapter abilities, want 1", len(face.ChapterAbilities))
	}
	grant, ok := face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous",
			face.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if grant.Duration != game.DurationPermanent {
		t.Fatalf("grant duration = %#v, want DurationPermanent", grant.Duration)
	}
	if len(grant.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(grant.ContinuousEffects))
	}
	effect := grant.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("continuous layer = %#v, want LayerAbility", effect.Layer)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("added abilities = %d, want 1", len(effect.AddAbilities))
	}
	granted, ok := effect.AddAbilities[0].(*game.TriggeredAbility)
	if !ok {
		t.Fatalf("granted ability = %T, want *game.TriggeredAbility", effect.AddAbilities[0])
	}
	if granted.Trigger.Pattern.Event != game.EventDamageDealt || !granted.Trigger.Pattern.RequireCombatDamage {
		t.Fatalf("granted trigger pattern = %#v, want combat damage dealt", granted.Trigger.Pattern)
	}
	if _, ok := granted.Content.Modes[0].Sequence[0].Primitive.(game.PlayerLosesGame); !ok {
		t.Fatalf("granted body primitive = %T, want game.PlayerLosesGame",
			granted.Content.Modes[0].Sequence[0].Primitive)
	}
}

// TestLowerLoseGameControllerEnterTrigger verifies that the exact controller
// effect "you lose the game" lowers to a single PlayerLosesGame instruction
// scoped to the ability's controller, mirroring the win-game enter trigger.
func TestLowerLoseGameControllerEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Lose Game",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment enters, you lose the game.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(seq))
	}
	lose, ok := seq[0].Primitive.(game.PlayerLosesGame)
	if !ok {
		t.Fatalf("instruction[0] = %#v, want PlayerLosesGame", seq[0].Primitive)
	}
	if lose.Player.Kind() == game.PlayerReferenceNone {
		t.Error("PlayerLosesGame.Player = none, want the controller reference")
	}
}

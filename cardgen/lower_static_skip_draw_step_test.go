package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSkipDrawStepStatic proves that the fixed player-rule phrase "Skip
// your draw step." lowers to the shared SkipDrawStepStaticBody, carrying the
// RuleEffectSkipDrawStep rule effect scoped to the controller.
func TestLowerSkipDrawStepStatic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Skip Draw Tester",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Skip your draw step.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if !reflect.DeepEqual(body, game.SkipDrawStepStaticBody) {
		t.Fatalf("body = %#v, want SkipDrawStepStaticBody", body)
	}
	if len(body.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", body.RuleEffects)
	}
	effect := body.RuleEffects[0]
	if effect.Kind != game.RuleEffectSkipDrawStep {
		t.Fatalf("rule effect kind = %v, want RuleEffectSkipDrawStep", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.AffectedSource || effect.AffectedAttached {
		t.Fatalf("rule effect must be player-scoped: %#v", effect)
	}
}

// TestGenerateYawgmothsBargain proves the full Necropotence-cluster pay-life
// draw engine lowers end to end: the "Skip your draw step." static renders to
// SkipDrawStepStaticBody alongside its pay-life activated draw.
func TestGenerateYawgmothsBargain(t *testing.T) {
	t.Parallel()
	generatedSourceContains(t, &ScryfallCard{
		Name:       "Yawgmoth's Bargain",
		Layout:     "normal",
		ManaCost:   "{3}{B}{B}",
		TypeLine:   "Enchantment",
		OracleText: "Skip your draw step.\nPay 1 life: Draw a card.",
	}, []string{"game.SkipDrawStepStaticBody"})
}

// TestGenerateGriselbrandSpelledDrawCount proves that a fixed draw whose count
// is spelled out beyond four ("Draw seven cards.") lowers exactly, rather than
// being rejected by a conservative spelled-cardinal ceiling.
func TestGenerateGriselbrandSpelledDrawCount(t *testing.T) {
	t.Parallel()
	power, toughness := "7", "7"
	generatedSourceContains(t, &ScryfallCard{
		Name:       "Griselbrand",
		Layout:     "normal",
		ManaCost:   "{4}{B}{B}{B}{B}",
		TypeLine:   "Legendary Creature — Demon",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: "Flying, lifelink\nPay 7 life: Draw seven cards.",
	}, []string{"game.Draw{", "game.Fixed(7)"})
}

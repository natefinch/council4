package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerNoMaximumHandSizeStatic proves that the fixed player-rule phrase
// "You have no maximum hand size." lowers to the shared
// NoMaximumHandSizeStaticBody, carrying the RuleEffectNoMaximumHandSize rule
// effect scoped to the controller.
func TestLowerNoMaximumHandSizeStatic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Reliquary Tester",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "You have no maximum hand size.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if !reflect.DeepEqual(body, game.NoMaximumHandSizeStaticBody) {
		t.Fatalf("body = %#v, want NoMaximumHandSizeStaticBody", body)
	}
	if len(body.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", body.RuleEffects)
	}
	effect := body.RuleEffects[0]
	if effect.Kind != game.RuleEffectNoMaximumHandSize {
		t.Fatalf("rule effect kind = %v, want RuleEffectNoMaximumHandSize", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.AffectedSource || effect.AffectedAttached {
		t.Fatalf("rule effect must be player-scoped: %#v", effect)
	}
}

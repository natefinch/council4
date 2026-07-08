package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerLivingMetalKeyword proves the Universes-Beyond Transformers "Living
// metal" keyword lowers to the shared self static that adds the creature card
// type during the controller's turn, reusing the SourceControllerTurn condition
// and the type continuous layer, and renders through the reusable body variable.
func TestLowerLivingMetalKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vehicle",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact — Vehicle",
		ManaCost:   "{2}{R}",
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want one Living metal static", len(face.StaticAbilities))
	}
	lowered := face.StaticAbilities[0]
	if lowered.VarName != "game.LivingMetalStaticBody" {
		t.Fatalf("var name = %q, want game.LivingMetalStaticBody", lowered.VarName)
	}
	if !reflect.DeepEqual(lowered.Body, game.LivingMetalStaticBody) {
		t.Fatalf("body = %#v, want game.LivingMetalStaticBody", lowered.Body)
	}
	ability := lowered.Body
	if !ability.Condition.Exists || !ability.Condition.Val.SourceControllerTurn {
		t.Fatalf("condition = %#v, want SourceControllerTurn", ability.Condition)
	}
	if len(ability.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %#v, want one type add", ability.ContinuousEffects)
	}
	effect := ability.ContinuousEffects[0]
	if effect.Layer != game.LayerType || !effect.AffectedSource {
		t.Fatalf("effect = %#v, want AffectedSource type-layer add", effect)
	}
	if len(effect.AddTypes) != 1 || effect.AddTypes[0] != types.Creature {
		t.Fatalf("added types = %#v, want [Creature]", effect.AddTypes)
	}
}

package action

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestPass(t *testing.T) {
	got := Pass()
	if got.Kind != ActionPass {
		t.Fatalf("Pass() kind = %v, want %v", got.Kind, ActionPass)
	}
}

func TestPlayLand(t *testing.T) {
	cardID := id.ID(42)

	got := PlayLand(cardID)
	if got.Kind != ActionPlayLand {
		t.Fatalf("PlayLand() kind = %v, want %v", got.Kind, ActionPlayLand)
	}
	if got.PlayLand.CardID != cardID {
		t.Fatalf("PlayLand() card ID = %v, want %v", got.PlayLand.CardID, cardID)
	}
}

func TestCastSpellPreservesTargets(t *testing.T) {
	cardID := id.ID(42)
	targets := []game.Target{
		game.PlayerTarget(game.Player2),
		game.PermanentTarget(id.ID(99)),
	}

	got := CastSpell(cardID, targets, 0, []int{1})
	if got.Kind != ActionCastSpell {
		t.Fatalf("CastSpell() kind = %v, want %v", got.Kind, ActionCastSpell)
	}
	if got.CastSpell.CardID != cardID {
		t.Fatalf("CastSpell() card ID = %v, want %v", got.CastSpell.CardID, cardID)
	}
	if !slices.Equal(got.CastSpell.Targets, targets) {
		t.Fatalf("CastSpell() targets = %+v, want %+v", got.CastSpell.Targets, targets)
	}
}

func TestActivateAbilityPreservesTargets(t *testing.T) {
	sourceID := id.ID(42)
	targets := []game.Target{
		game.PlayerTarget(game.Player3),
		game.PermanentTarget(id.ID(100)),
	}

	got := ActivateAbility(sourceID, 2, targets, 0)
	if got.Kind != ActionActivateAbility {
		t.Fatalf("ActivateAbility() kind = %v, want %v", got.Kind, ActionActivateAbility)
	}
	if got.ActivateAbility.SourceID != sourceID {
		t.Fatalf("ActivateAbility() source ID = %v, want %v", got.ActivateAbility.SourceID, sourceID)
	}
	if !slices.Equal(got.ActivateAbility.Targets, targets) {
		t.Fatalf("ActivateAbility() targets = %+v, want %+v", got.ActivateAbility.Targets, targets)
	}
}

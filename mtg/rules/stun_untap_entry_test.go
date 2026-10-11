package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestStunUntapEntryIsNotAnUntapAttempt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Untapped Entry", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersUntappedGroupReplacement(
			"Lands you control enter untapped.", game.TriggerControllerYou, types.Land,
		)},
	}})
	start := len(g.Events)
	entry := game.EntersTappedReplacement("This land enters tapped with two stun counters.")
	entry.Replacement.EntersWithCounters = []game.CounterPlacement{{Kind: counter.Stun, Amount: 2}}
	target := addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Stunned Entry", Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{entry},
	}})
	if target.Tapped || target.Counters.Get(counter.Stun) != 2 {
		t.Fatalf("entry = tapped %v, stun %d; want false, 2", target.Tapped, target.Counters.Get(counter.Stun))
	}
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventPermanentUntapped {
			t.Fatal("setting untapped entry state emitted an actual untap event")
		}
	}
}

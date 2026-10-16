package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestWarpWorldNativeBodyReplaysThroughJSON(t *testing.T) {
	// This fixture isolates body/replay fidelity; printed payment is covered in rules.
	warp := authored.WarpWorld.New()
	warp.ManaCost = opt.Val(cost.Mana{})
	creature := &game.CardDef{CardFace: game.CardFace{
		Name: "Replay creature", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 3}),
	}}
	enchantment := &game.CardDef{CardFace: game.CardFace{
		Name: "Replay enchantment", Types: []types.Card{types.Enchantment},
	}}
	cfg := Config{Seed: 3288, Games: 1, TurnLimit: 4, NewAgents: genericAgents}
	for player := range cfg.Configs {
		for range 6 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, warp)
		}
		for range 18 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, creature)
		}
		for range 6 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, enchantment)
		}
	}
	result, record := RecordGame(cfg, 0)
	resolved := false
	for _, turn := range result.Turns {
		for _, entry := range turn.Entries {
			resolved = resolved || entry.Resolve.SourceName == "Warp World"
		}
	}
	if !resolved || result.Execution.InvalidActions != 0 || result.Execution.InvalidChoices != 0 {
		t.Fatal("recorded game did not execute a valid native Warp World")
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var restored ReplayRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if replayed := Replay(cfg.Configs, restored); !reflect.DeepEqual(replayed, result) {
		t.Fatal("native staged movement and choices diverged after JSON replay")
	}
}

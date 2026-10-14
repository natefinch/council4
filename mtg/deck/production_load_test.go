package deck

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

func TestAllProvingCardsAndLegacyDeckReloadKeepProductionEvidence(t *testing.T) {
	var text strings.Builder
	_, _ = text.WriteString("40 Forest\n")
	for _, entry := range authored.Cards {
		_, _ = text.WriteString("1 " + entry.Name + "\n")
	}
	list, err := Parse(strings.NewReader(text.String()))
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	var restored Decklist
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	var inputs [game.NumPlayers]PlayerInput
	for p := range inputs {
		inputs[p] = PlayerInput{Name: "mixed reload fixture", Decklist: list}
	}
	first := Load(inputs, game.Player1, cards.NewDefaultRegistry())
	for p := range inputs {
		inputs[p].Decklist = &restored
	}
	registry := cards.NewDefaultRegistry()
	reloaded := Load(inputs, game.Player1, registry)
	// A small fixture is not a legal 100-card Commander deck. Its explicit
	// legality errors must survive reload rather than blocking lookup evidence.
	if len(first.Unresolved) != 0 || len(reloaded.Unresolved) != 0 ||
		!reflect.DeepEqual(first.Support, reloaded.Support) ||
		!reflect.DeepEqual(first.Legality, reloaded.Legality) || len(reloaded.Legality) == 0 {
		t.Fatal("mixed deck reload lost identity, support or legality information")
	}
	authoredEvidence := 0
	for _, support := range reloaded.Support {
		if support.Implementation == "authored" && support.Loadable && support.BehaviorVerified {
			authoredEvidence++
		}
	}
	if authoredEvidence != game.NumPlayers*len(authored.Cards) {
		t.Fatal("deck consumers lost a proving-card admission")
	}
	engine := rules.NewEngineWithPrograms(nil, registry.Programs())
	if err := engine.ValidatePrograms(reloaded.Configs); err != nil {
		t.Fatal(err)
	}
}

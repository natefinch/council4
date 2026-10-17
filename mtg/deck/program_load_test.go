package deck

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestMissingProgramIsADeckLoadingProblem(t *testing.T) {
	t.Parallel()
	entry := cardset.Entry{Name: "Missing", OracleID: "oracle-id",
		Verification: "mtg/deck/program_load_test.go: missing-program test fixture",
		New: func() *game.CardDef {
			return &game.CardDef{CardFace: game.CardFace{Name: "Missing", Types: []types.Card{types.Sorcery},
				SpellAbility: opt.Val(game.Mode{Program: game.ProgramReference{ID: "missing", Version: "1"}}.Ability())}}
		}}
	registry, err := cards.NewProgramRegistry([]cardset.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}
	list, err := Parse(strings.NewReader("1 Missing"))
	if err != nil {
		t.Fatal(err)
	}
	var inputs [game.NumPlayers]PlayerInput
	inputs[0] = PlayerInput{Name: "Player", Decklist: list}
	loaded := Load(inputs, game.Player1, registry)
	if loaded.OK() || len(loaded.Unresolved) != 1 || loaded.Unresolved[0].Reason == "" {
		t.Fatal("missing program was hidden or admitted")
	}
}

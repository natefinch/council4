package report

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
	"github.com/natefinch/council4/opt"
)

var updateGolden = flag.Bool("update", false, "update the golden report files in testdata")

const (
	goldenForest    id.ID = 1
	goldenBolt      id.ID = 2
	goldenBomb      id.ID = 3
	goldenGrizzly   id.ID = 4
	goldenGrizzlyID id.ID = 100
)

func goldenCards() map[id.ID]rules.CardInfo {
	return map[id.ID]rules.CardInfo{
		goldenForest:  {Name: "Forest", Owner: game.Player1, ManaValue: 0, Types: []types.Card{types.Land}},
		goldenBolt:    {Name: "Lightning Bolt", Owner: game.Player1, ManaValue: 1, Types: []types.Card{types.Instant}},
		goldenBomb:    {Name: "Expensive Bomb", Owner: game.Player1, ManaValue: 6, Types: []types.Card{types.Sorcery}},
		goldenGrizzly: {Name: "Grizzly Bears", Owner: game.Player1, ManaValue: 2, Types: []types.Card{types.Creature}},
	}
}

func goldenLand() rules.ActionLog {
	return rules.ActionLog{Player: game.Player1, Action: action.PlayLand(goldenForest)}
}

func goldenCast(cardID id.ID) rules.ActionLog {
	return rules.ActionLog{Player: game.Player1, Action: action.CastSpell(cardID, nil, 0, nil)}
}

// goldenFixture is a fixed, fully deterministic five-game simulation (a win, a
// loss, another win, a draw, and a failed game) for the deck under test at seat
// Player1. It exercises every report section.
func goldenFixture() sim.SimulationResult {
	cards := goldenCards()

	win := rules.GameResult{
		HasWinner: true, Winner: game.Player1, TurnCount: 3, Cards: cards,
		EliminationOrder: []game.PlayerID{game.Player2, game.Player3, game.Player4},
		Turns: []rules.TurnLog{
			{TurnNumber: 1, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenLand()}},
			{TurnNumber: 2, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenLand(), goldenCast(goldenBolt)},
				CombatDamage: []rules.CombatDamageLog{{Controller: game.Player1, DefendingPlayer: game.Player2, Damage: 4}}},
			{TurnNumber: 3, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenCast(goldenGrizzly)},
				CombatDamage: []rules.CombatDamageLog{{Controller: game.Player1, DefendingPlayer: game.Player2, Damage: 4}}},
		},
		Events: []game.Event{
			{Kind: game.EventCardDrawn, CardID: goldenBolt},
			{Kind: game.EventSpellCast, CardID: goldenBolt, ManaValue: opt.Val(1)},
			{Kind: game.EventSpellResolved, CardID: goldenBolt},
			{Kind: game.EventCardDrawn, CardID: goldenGrizzly},
			{Kind: game.EventSpellCast, CardID: goldenGrizzly, ManaValue: opt.Val(2)},
			{Kind: game.EventPermanentEnteredBattlefield, PermanentID: goldenGrizzlyID, CardID: goldenGrizzly},
			{Kind: game.EventCardDrawn, CardID: goldenBomb},
			{Kind: game.EventObjectBecameTarget, Controller: game.Player2, Target: game.PlayerTarget(game.Player1)},
			{Kind: game.EventObjectBecameTarget, Controller: game.Player2, Target: game.PermanentTarget(goldenGrizzlyID)},
		},
	}
	win.EndState.Players[game.Player1] = rules.PlayerEndState{Life: 40, Hand: []id.ID{goldenBomb}, LibrarySize: 20, CommanderCasts: 1}

	loss := rules.GameResult{
		HasWinner: true, Winner: game.Player2, TurnCount: 6, Cards: cards,
		EliminationOrder: []game.PlayerID{game.Player1, game.Player3, game.Player4},
		Turns: []rules.TurnLog{
			{TurnNumber: 1, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenLand()}},
			{TurnNumber: 3, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenCast(goldenBolt)}},
		},
		Events: []game.Event{
			{Kind: game.EventCardDrawn, CardID: goldenBomb},
			{Kind: game.EventCardDrawn, CardID: goldenBolt},
			{Kind: game.EventSpellCast, CardID: goldenBolt, ManaValue: opt.Val(1)},
		},
	}
	loss.EndState.Players[game.Player1] = rules.PlayerEndState{Life: 0, Eliminated: true, LibrarySize: 10, CommanderCasts: 0}

	secondWin := rules.GameResult{
		HasWinner: true, Winner: game.Player1, TurnCount: 5, Cards: cards,
		EliminationOrder: []game.PlayerID{game.Player3, game.Player4, game.Player2},
		Turns: []rules.TurnLog{
			{TurnNumber: 2, ActivePlayer: game.Player1, Actions: []rules.ActionLog{goldenCast(goldenBolt)},
				CombatDamage: []rules.CombatDamageLog{{Controller: game.Player1, DefendingPlayer: game.Player3, Damage: 6}}},
		},
		Events: []game.Event{
			{Kind: game.EventSpellCast, CardID: goldenBolt, ManaValue: opt.Val(1)},
		},
	}
	secondWin.EndState.Players[game.Player1] = rules.PlayerEndState{Life: 30, LibrarySize: 15, CommanderCasts: 2}

	draw := rules.GameResult{
		HasWinner: false, TurnCount: 1000, Cards: cards,
		EliminationOrder: []game.PlayerID{game.Player2, game.Player3},
	}
	draw.EndState.Players[game.Player1] = rules.PlayerEndState{Life: 12, LibrarySize: 0, CommanderCasts: 0}

	failed := rules.GameResult{}

	return sim.SimulationResult{
		Games:      []rules.GameResult{win, loss, secondWin, draw, failed},
		Seeds:      []uint64{11, 22, 33, 44, 55},
		GameCount:  5,
		MasterSeed: 7,
		Failures:   []sim.GameFailure{{Index: 4, Seed: 55, Reason: "unsupported card: Mystery Card"}},
	}
}

func goldenOptions() Options {
	return Options{
		TestedSeat: game.Player1,
		DeckNames:  [game.NumPlayers]string{"Tested Deck", "Rival A", "Rival B", "Rival C"},
	}
}

func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGolden {
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatalf("update golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run with -update to create): %v", name, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden %s drift.\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
	}
}

func TestGoldenReportText(t *testing.T) {
	report := Generate(goldenFixture(), goldenOptions())
	var out bytes.Buffer
	if err := report.WriteText(&out); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	checkGolden(t, "report.txt", out.Bytes())
}

func TestGoldenReportJSON(t *testing.T) {
	report := Generate(goldenFixture(), goldenOptions())
	var out bytes.Buffer
	if err := report.WriteJSON(&out); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	checkGolden(t, "report.json", out.Bytes())
}

// TestGoldenReportIsDeterministic guards against unstable ordering: generating
// the report twice must produce byte-identical output.
func TestGoldenReportIsDeterministic(t *testing.T) {
	first := Generate(goldenFixture(), goldenOptions())
	second := Generate(goldenFixture(), goldenOptions())

	var firstText, secondText, firstJSON, secondJSON bytes.Buffer
	_ = first.WriteText(&firstText)
	_ = second.WriteText(&secondText)
	_ = first.WriteJSON(&firstJSON)
	_ = second.WriteJSON(&secondJSON)

	if !bytes.Equal(firstText.Bytes(), secondText.Bytes()) {
		t.Error("text report is not deterministic across two generations")
	}
	if !bytes.Equal(firstJSON.Bytes(), secondJSON.Bytes()) {
		t.Error("JSON report is not deterministic across two generations")
	}
}

package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledMovePermanentLibrarySequences(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       string
		owner      game.PlayerID
		bottom     bool
		token      bool
		drawMoved  bool
		draw       bool
		lifePlayer game.PlayerID
		lifeDelta  int
	}{
		{"bottom controller life", "Put target creature on the bottom of its owner's library. You lose 2 life.", game.Player2, true, false, false, false, game.Player1, -2},
		{"top controller life", "Put target creature on top of its owner's library. You lose 2 life.", game.Player2, false, false, false, false, game.Player1, -2},
		{"top then draw", "Put target creature on top of its owner's library. Draw a card.", game.Player1, false, false, true, true, game.Player1, 0},
		{"bottom then draw", "Put target creature on the bottom of its owner's library. Draw a card.", game.Player1, true, false, false, true, game.Player1, 0},
		{"draw before bottom", "Draw a card. Put target creature on the bottom of its owner's library.", game.Player2, true, false, false, true, game.Player1, 0},
		{"controller LKI", "Put target creature on the bottom of its owner's library. Its controller loses 2 life.", game.Player2, true, false, false, false, game.Player3, -2},
		{"owner LKI", "Put target creature on top of its owner's library. Its owner gains 2 life.", game.Player2, false, false, false, false, game.Player2, 2},
		{"toughness LKI", "Put target creature on the bottom of its owner's library. Its controller gains life equal to its toughness.", game.Player2, true, false, false, false, game.Player3, 3},
		{"owner toughness LKI", "Put target creature on top of its owner's library. Its owner gains life equal to its toughness.", game.Player2, false, false, false, false, game.Player2, 3},
		{"token toughness LKI", "Put target creature on the bottom of its owner's library. Its controller gains life equal to its toughness.", game.Player2, true, true, false, false, game.Player3, 3},
		{"token", "Put target creature on the bottom of its owner's library. Its controller loses 2 life.", game.Player2, true, true, false, false, game.Player3, -2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
				Name: "Test Library Move", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "Tap target creature. " + tc.body,
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
			}

			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			prior := addCombatCreaturePermanentWithPower(g, game.Player4, 7)
			moved := addCombatCreaturePermanentWithPower(g, tc.owner, 3)
			moved.Controller = game.Player3
			moved.Token = tc.token
			if tc.token {
				moved.TokenDef = g.CardInstances[moved.CardInstanceID].Def
			}
			var existing [game.NumPlayers]id.ID
			for player := range g.Players {
				existing[player] = addCardToLibrary(g, game.PlayerID(player), &game.CardDef{CardFace: game.CardFace{Name: "Existing Card"}})
			}
			spell := addCardToHand(g, game.Player1, defs[0])
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
				Targets:      []game.Target{game.PermanentTarget(prior.ObjectID), game.PermanentTarget(moved.ObjectID)},
				TargetCounts: []int{1, 1},
			})
			clone := g.Clone()
			engine := NewEngine(nil)
			engine.resolveTopOfStack(clone, &TurnLog{})
			if _, ok := permanentByObjectID(g, moved.ObjectID); !ok || g.Players[tc.owner].Library.Contains(moved.CardInstanceID) {
				t.Fatal("resolving the clone mutated the original")
			}
			g = clone
			if permanent, ok := permanentByObjectID(g, prior.ObjectID); !ok || !permanent.Tapped {
				t.Fatal("earlier target was moved instead of tapped")
			}
			if _, ok := permanentByObjectID(g, moved.ObjectID); ok {
				t.Fatal("library target remained on battlefield")
			}
			owner := g.Players[tc.owner]
			switch {
			case tc.token:
				engine.applyStateBasedActions(g)
				for _, player := range g.Players {
					if player.Library.Contains(moved.ObjectID) || player.Hand.Contains(moved.ObjectID) ||
						player.Library.Contains(moved.CardInstanceID) || player.Hand.Contains(moved.CardInstanceID) {
						t.Fatal("token became a library or hand card")
					}
				}
			case tc.drawMoved:
				if !owner.Hand.Contains(moved.CardInstanceID) || owner.Library.Contains(moved.CardInstanceID) {
					t.Fatal("top move did not precede drawing the moved card")
				}
			default:
				card, ok := owner.Library.Top()
				if tc.bottom {
					card, ok = owner.Library.Bottom()
				}
				if !ok || card != moved.CardInstanceID {
					t.Fatalf("library placement = %v (ok=%v), want moved card %v", card, ok, moved.CardInstanceID)
				}
				if got, ok := cardZone(g, moved.CardInstanceID); !ok || got != zone.Library {
					t.Fatalf("moved card zone = %v (ok=%v), want library", got, ok)
				}
			}
			for player, state := range g.Players {
				wantLife := 40
				if game.PlayerID(player) == tc.lifePlayer {
					wantLife += tc.lifeDelta
				}
				if state.Life != wantLife {
					t.Fatalf("player %d life = %d, want %d", player, state.Life, wantLife)
				}
				if game.PlayerID(player) != tc.owner && state.Library.Contains(moved.CardInstanceID) {
					t.Fatalf("moved card entered nonowner player %d's library", player)
				}
			}
			if tc.draw && !tc.drawMoved && !g.Players[game.Player1].Hand.Contains(existing[game.Player1]) {
				t.Fatal("draw rider did not draw the controller's preexisting top card")
			}
		})
	}
}

func TestCompiledMovePermanentConditionalLifeRider(t *testing.T) {
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Test Conditional Library Move", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Put target creature on the bottom of its owner's library. If you control three or more artifacts, you lose 2 life.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
	}
	for _, count := range []int{2, 3} {
		t.Run(strconv.Itoa(count)+" artifacts", func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			moved := addCombatCreaturePermanent(g, game.Player2)
			for range count {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Artifact", Types: []types.Card{types.Artifact}}})
			}
			spell := addCardToHand(g, game.Player1, defs[0])
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(moved.ObjectID)}, TargetCounts: []int{1},
			})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Library.Contains(moved.CardInstanceID) {
				t.Fatal("unconditional move was incorrectly gated")
			}
			wantLife := 40
			if count == 3 {
				wantLife -= 2
			}
			if got := g.Players[game.Player1].Life; got != wantLife {
				t.Fatalf("controller life = %d, want %d", got, wantLife)
			}
		})
	}
}

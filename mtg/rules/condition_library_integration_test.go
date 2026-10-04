package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledUnlessLibraryComposition(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "The Spot's Portal", Layout: "normal", TypeLine: "Instant", ManaCost: "{2}{B}",
		OracleText: "Put target creature on the bottom of its owner's library. You lose 2 life unless you control a Villain.",
	})
	for _, tc := range []struct {
		name            string
		movedController game.PlayerID
		movedVillain    bool
		otherVillain    bool
		otherController game.PlayerID
		wantLifeLost    int
	}{
		{name: "no Villain", movedController: game.Player2, wantLifeLost: 2},
		{name: "own remaining Villain", movedController: game.Player2, otherVillain: true, otherController: game.Player1},
		{name: "opponent Villain", movedController: game.Player2, otherVillain: true, otherController: game.Player2, wantLifeLost: 2},
		{name: "own Villain moved first", movedController: game.Player1, movedVillain: true, wantLifeLost: 2},
		{name: "opponent Villain moved", movedController: game.Player2, movedVillain: true, wantLifeLost: 2},
		{name: "own second Villain remains", movedController: game.Player1, movedVillain: true, otherVillain: true, otherController: game.Player1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			subtypes := []types.Sub{types.Rogue}
			if tc.movedVillain {
				subtypes = []types.Sub{types.Villain}
			}
			moved := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
				Name: "Moved creature", Types: []types.Card{types.Creature}, Subtypes: subtypes,
			}})
			moved.Controller = tc.movedController
			if tc.otherVillain {
				addCombatPermanent(g, tc.otherController, &game.CardDef{CardFace: game.CardFace{
					Name: "Remaining Villain", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Villain},
				}})
			}
			addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Existing library card"}})
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
				Targets: []game.Target{game.PermanentTarget(moved.ObjectID)}, TargetCounts: []int{1},
			})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if _, ok := permanentByObjectID(g, moved.ObjectID); ok {
				t.Fatal("unconditional library move did not resolve")
			}
			if bottom, ok := g.Players[game.Player3].Library.Bottom(); !ok || bottom != moved.CardInstanceID {
				t.Fatal("target did not reach its owner's library bottom")
			}
			for player, state := range g.Players {
				want := 40
				if game.PlayerID(player) == game.Player1 {
					want -= tc.wantLifeLost
				}
				if state.Life != want {
					t.Fatalf("player %d life = %d, want %d", player, state.Life, want)
				}
				if game.PlayerID(player) != game.Player3 && state.Library.Contains(moved.CardInstanceID) {
					t.Fatalf("target reached nonowner player %d's library", player)
				}
			}
		})
	}
}

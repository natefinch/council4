package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

// The original runtime counterexample exiled two chosen cards but created one
// token. An unmodeled actual-card iterator must refuse before execution.
func TestXCardExileUnmodeledActualCountRefused(t *testing.T) {
	card := &cardgen.ScryfallCard{
		Name: "Unmodeled Iterator", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{X}{B}",
		OracleText: "Exile X target creature cards from your graveyard. For each creature card exiled this way, create a 2/2 black Zombie creature token.",
	}
	defs, diagnostics, err := cardgen.CompileCardDefs(card)
	if err != nil || len(defs) != 0 || len(diagnostics) == 0 {
		t.Fatalf("unmodeled actual-card iterator admitted: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
	}
}

func TestCompiledLeadingTokenCountPreservesTypedSelection(t *testing.T) {
	def := compileTargetCountSpell(t, "For each token you control, create a 1/1 white Rabbit creature token.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 2 {
		addCreaturePermanent(g, game.Player1)
		token := addCreaturePermanent(g, game.Player1)
		token.Token, token.TokenDef = true, g.CardInstances[token.CardInstanceID].Def
	}
	opponentToken := addCreaturePermanent(g, game.Player2)
	opponentToken.Token, opponentToken.TokenDef = true, g.CardInstances[opponentToken.CardInstanceID].Def
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spell)
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	tokens := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			tokens++
		}
	}
	if tokens != 4 {
		t.Fatalf("controlled tokens=%d, want two originals and two created; non-tokens and opponents must not count", tokens)
	}
}

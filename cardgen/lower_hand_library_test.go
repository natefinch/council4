package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestHandLibraryClauseCompositions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text, path string
		index, count     int
		bottom, x        bool
	}{
		{"standalone top", "Put a card from your hand on top of your library.", "SpellAbility.Val", 0, 1, false, false},
		{"standalone bottom", "Put five cards from your hand on the bottom of your library.", "SpellAbility.Val", 0, 5, true, false},
		{"life then put", "You gain 2 life, then put two cards from your hand on top of your library.", "SpellAbility.Val", 1, 2, false, false},
		{"intervening rider", "Draw three cards. You gain 2 life. Put X cards from your hand on the bottom of your library in any order.", "SpellAbility.Val", 2, 0, true, true},
		{"draw X then put X", "Draw X cards, then put X cards from your hand on top of your library.", "SpellAbility.Val", 1, 0, false, true},
		{"activated X", "{X}: Put X cards from your hand on top of your library.", "ActivatedAbilities[0].Content", 0, 0, false, true},
		{"trigger standalone", "At the beginning of your upkeep, put a card from your hand on the bottom of your library.", "TriggeredAbilities[0].Content", 0, 1, true, false},
		{"optional", "You may put a card from your hand on top of your library.", "SpellAbility.Val", 0, 1, false, false},
		{"target then independent put", "Destroy target creature. Put a card from your hand on top of your library.", "SpellAbility.Val", 1, 1, false, false},
		{"optional after draw", "Draw a card. You may put a card from your hand on top of your library.", "SpellAbility.Val", 1, 1, false, false},
		{"conditioned group", "If you have no cards in hand, draw two cards, then put a card from your hand on top of your library.", "SpellAbility.Val", 1, 1, false, false},
		{"resolving cost consequence", "You may discard a card. If you do, put a card from your hand on top of your library.", "SpellAbility.Val", 1, 1, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			typeLine := "Instant"
			if tt.path != "SpellAbility.Val" {
				typeLine = "Artifact"
			}
			card := &ScryfallCard{Name: "Hand Library Test", Layout: "normal", TypeLine: typeLine, ManaCost: "{X}{U}", OracleText: tt.text}
			path := fmt.Sprintf("%s.Modes[0].Sequence[%d].Primitive.(game.MoveCard)", tt.path, tt.index)
			assertCardPaths(t, card, path+".Player.kind = game.PlayerReferenceController",
				path+".FromZone = zone.Hand", path+".Destination = zone.Library")
			if tt.x {
				assertCardPaths(t, card, path+".Amount.dynamic.Kind = game.DynamicAmountX")
			} else {
				assertCardPaths(t, card, fmt.Sprintf("%s.Amount.fixed = %d", path, tt.count))
			}
			if tt.bottom {
				assertCardPaths(t, card, path+".DestinationBottom = true")
			} else {
				assertCardPathsAbsent(t, card, path+".DestinationBottom")
			}
			if tt.name == "optional" {
				assertCardPaths(t, card, tt.path+".Modes[0].Sequence[0].Optional = true")
			}
			assertCardPathsAbsent(t, card, path+".Card", path+".Group", path+".Player.targetIndex")
		})
	}
}

func TestHandLibraryCompiledParameterSpans(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"", "You gain 2 life. "} {
		source := prefix + "Put two cards from an opponent's hand on top of your library."
		_, diagnostics, err := CompileCardDefs(&ScryfallCard{
			Name: "Hand Library Parameter", Layout: "normal", TypeLine: "Instant", OracleText: source,
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, diagnostic := range diagnostics {
			if diagnostic.Summary == "unsupported hand-library source owner" {
				got := source[diagnostic.Span.Start.Offset:diagnostic.Span.End.Offset]
				if got != "from an opponent's hand" {
					t.Fatalf("diagnostic span=%q", got)
				}
				found = true
			}
		}
		if !found {
			t.Fatalf("missing source-owner diagnostic: %+v", diagnostics)
		}
	}
}

func TestHandLibraryClauseRefusals(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ text, reason string }{
		{"Put zero cards from your hand on top of your library.", "hand-library amount"},
		{"Put up to X cards from your hand on top of your library.", "hand-library amount"},
		{"Put any number of cards from your hand on top of your library.", "hand-library amount"},
		{"Put X cards from your hand on top of your library, where X is the number of cards in your hand.", "hand-library amount"},
		{"Put two cards from your hand on top of your library in a random order.", "hand-library library order"},
		{"Put two cards from your hand on top of your library in the same order.", "hand-library library order"},
		{"Put two cards from an opponent's hand on top of your library.", "hand-library source owner"},
		{"Put two cards from your hand on top of an opponent's library.", "hand-library destination owner"},
		{"Put two revealed cards from your hand on top of your library.", "hand-library card selection"},
		{"Draw a card. Put it from your hand on top of your library.", ""},
		{"Put two cards from your hand on top of your library and shuffle it.", ""},
		{"Put two cards from your hand and graveyard on top of your library.", ""},
		{"You may put two cards from your hand on top of your library. If you do, you gain 2 life.", "hand-library resolving cost"},
		{"You may put X cards from your hand on top of your library. If you do, you gain 2 life.", "hand-library resolving cost"},
		{"You may put a card from your hand on top of your library. When you do, you gain 2 life.", "hand-library resolving cost"},
		{"Put two cards from your hand on top of your library. If you don't, you lose 2 life.", "hand-library resolving cost"},
		{"You may put two cards from your hand on top of your library. If you do, you gain 2 life. Otherwise, you lose 2 life.", "hand-library resolving cost"},
		{"You may put X cards from your hand on top of your library. If you don't, you lose 2 life.", "hand-library resolving cost"},
		{"You may put two cards from your hand on top of your library. Otherwise, you lose 2 life.", ""},
		{"You lose 2 life unless you put two cards from your hand on top of your library.", ""},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			for _, prefix := range []string{"", "You gain 2 life. "} {
				assertCardUnsupported(t, &ScryfallCard{Name: "Hand Library Near Miss", Layout: "normal", TypeLine: "Instant",
					ManaCost: "{X}{U}", OracleText: prefix + tt.text}, tt.reason)
			}
		})
	}
}

func TestHandLibraryLoweringIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse("Put X cards from your hand on the bottom of your library.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}

	document.Abilities[0].Sentences[0].Effects[0].Text = "not Oracle text"
	document.Abilities[0].Sentences[0].Effects[0].Tokens = nil
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Effects[0].Text = "not Oracle text"
	lowered, diagnostic, handled := lowerHandLibraryPutClause(contentCtx{content: content})
	if diagnostic != nil || !handled {
		t.Fatalf("diagnostic=%v handled=%v", diagnostic, handled)
	}
	move, ok := lowered.Modes[0].Sequence[0].Primitive.(game.MoveCard)
	if !ok || !move.DestinationBottom || !move.Amount.IsDynamic() ||
		move.Amount.DynamicAmount().Val.Kind != game.DynamicAmountX {
		t.Fatalf("move = %+v", move)
	}
}

func TestHandLibraryAmountAdapterDomain(t *testing.T) {
	t.Parallel()
	for _, amount := range []compiler.CompiledAmount{
		{Known: true, Value: 0},
		{Known: true, Value: -1},
		{},
		{VariableX: true, RangeKnown: true},
		{AnyNumber: true},
		{VariableX: true, Multiplier: 2},
		{VariableX: true, Addend: 1},
		{Known: true, Value: 2, VariableX: true},
		{DynamicKind: compiler.DynamicAmountSourcePower},
		{DynamicKind: compiler.DynamicAmountTriggeringEventAmount},
		{VariableX: true, DynamicForm: compiler.DynamicAmountWhereX},
	} {
		t.Run(fmt.Sprintf("%+v", amount), func(t *testing.T) {
			t.Parallel()
			_, diagnostic, handled := lowerHandLibraryPutClause(contentCtx{content: compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{
					Exact: true, Context: parser.EffectContextController, FromZone: zone.Hand, ToZone: zone.Library,
					HandLibraryPut: parser.HandLibraryPutSyntax{Present: true}, Amount: amount,
				}},
			}})
			if !handled || diagnostic == nil || diagnostic.Summary != "unsupported hand-library amount" {
				t.Fatalf("amount=%+v diagnostic=%v handled=%v", amount, diagnostic, handled)
			}
		})
	}
}

package oracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

type cachedCard struct {
	Name       string       `json:"name"`
	OracleText string       `json:"oracle_text"`
	CardFaces  []cachedFace `json:"card_faces"`
}

type cachedFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

func TestLexerExamples(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		want   string
	}{
		"mana ability": {
			source: "{T}: Add {G}.",
			want:   "symbol({T}) colon(:) word(Add) symbol({G}) period(.) EOF()",
		},
		"loyalty": {
			source: "−2: Target creature gets +1/+0.",
			want:   "minus(−) integer(2) colon(:) word(Target) word(creature) word(gets) plus(+) integer(1) slash(/) plus(+) integer(0) period(.) EOF()",
		},
		"modal": {
			source: "Choose one —\n• Search your library.\n• Fight.",
			want:   "word(Choose) word(one) em dash(—) newline(\\n) bullet(•) word(Search) word(your) word(library) period(.) newline(\\n) bullet(•) word(Fight) period(.) EOF()",
		},
		"keyword reminder": {
			source: "First strike (This creature can't be blocked.)",
			want:   "word(First) word(strike) left parenthesis(() word(This) word(creature) word(can't) word(be) word(blocked) period(.) right parenthesis()) EOF()",
		},
		"ability word": {
			source: "Formidable — Whenever you attack, draw a card.",
			want:   "word(Formidable) em dash(—) word(Whenever) word(you) word(attack) comma(,) word(draw) word(a) word(card) period(.) EOF()",
		},
		"quoted ability": {
			source: `Equipped creature has "{2}: This creature gets +1/+0."`,
			want:   "word(Equipped) word(creature) word(has) quote(\") symbol({2}) colon(:) word(This) word(creature) word(gets) plus(+) integer(1) slash(/) plus(+) integer(0) period(.) quote(\") EOF()",
		},
		"class level": {
			source: "{1}{G}: Level 2",
			want:   "symbol({1}) symbol({G}) colon(:) word(Level) integer(2) EOF()",
		},
		"hyphenated word": {
			source: "Fire-Lit non-Human",
			want:   "word(Fire-Lit) word(non-Human) EOF()",
		},
		"extended punctuation": {
			source: "opponents' [−1] Minsc & Boo = friends",
			want:   "word(opponents) apostrophe(') left bracket([) minus(−) integer(1) right bracket(]) word(Minsc) ampersand(&) word(Boo) glyph(=) word(friends) EOF()",
		},
		"crlf": {
			source: "Flying\r\nHaste",
			want:   "word(Flying) newline(\\r\\n) word(Haste) EOF()",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tokenSummary(test.source); got != test.want {
				t.Fatalf("tokens:\n got: %s\nwant: %s", got, test.want)
			}
		})
	}
}

func TestLexerPositions(t *testing.T) {
	t.Parallel()
	lexer := NewLexer("Flying\n{T}: Add {G}.")
	want := []Token{
		{Kind: Word, Text: "Flying", Span: Span{Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 6, Line: 1, Column: 7}}},
		{Kind: Newline, Text: "\n", Span: Span{Start: Position{Offset: 6, Line: 1, Column: 7}, End: Position{Offset: 7, Line: 2, Column: 1}}},
		{Kind: Symbol, Text: "{T}", Span: Span{Start: Position{Offset: 7, Line: 2, Column: 1}, End: Position{Offset: 10, Line: 2, Column: 4}}},
	}
	for i, expected := range want {
		if got := lexer.Next(); got != expected {
			t.Fatalf("token %d: got %#v, want %#v", i, got, expected)
		}
	}
}

func TestLexerInvalidInput(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"invalid UTF-8":  string([]byte{0xff}),
		"NUL":            "\x00",
		"midstream BOM":  "A\ufeffB",
		"unclosed brace": "{T",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var found bool
			lexer := NewLexer(source)
			for {
				token := lexer.Next()
				found = found || token.Kind == Invalid
				if token.Kind == EOF {
					break
				}
			}
			if !found {
				t.Fatalf("expected invalid token for %q", source)
			}
		})
	}
}

func TestInvalidReason(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		want   string
	}{
		"invalid UTF-8": {source: string([]byte{0xff}), want: "invalid UTF-8 encoding"},
		"NUL":           {source: "\x00", want: "NUL is not valid in Oracle text"},
		"midstream BOM": {source: "A\uFEFF", want: "a UTF-8 BOM is only valid at the start of Oracle text"},
		"unclosed":      {source: "{T", want: "unclosed braced symbol"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lexer := NewLexer(test.source)
			token := lexer.Next()
			if token.Kind != Invalid {
				token = lexer.Next()
			}
			if got := InvalidReason(token); got != test.want {
				t.Fatalf("reason = %q, want %q", got, test.want)
			}
		})
	}
	if got := InvalidReason(Token{Kind: Word}); got != "" {
		t.Fatalf("word reason = %q", got)
	}
}

func TestLexerLeadingBOM(t *testing.T) {
	t.Parallel()
	lexer := NewLexer("\uFEFFFlying")
	if got := lexer.Next(); got.Kind != Word || got.Text != "Flying" || got.Span.Start.Offset != 3 {
		t.Fatalf("first token = %#v", got)
	}
}

func TestLexerScryfallCache(t *testing.T) {
	t.Parallel()
	cache := filepath.Join("..", "..", ".cardwork", "deck", "cache", "scryfall")
	paths, err := filepath.Glob(filepath.Join(cache, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("local Scryfall cache is not present")
	}

	var texts int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var card cachedCard
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatalf("%s: %v", path, err)
		}

		check := func(name, source string) {
			t.Helper()
			if source == "" {
				return
			}
			texts++
			assertValidTokenStream(t, name, source)
		}
		check(card.Name, card.OracleText)
		for _, face := range card.CardFaces {
			check(card.Name+" / "+face.Name, face.OracleText)
		}
	}
	if texts != 59 {
		t.Fatalf("checked %d non-empty Oracle texts, want 59", texts)
	}
}

func FuzzLexer(f *testing.F) {
	for _, source := range []string{
		"",
		"{T}: Add {G}.",
		"Choose one —\n• Draw a card.",
		string([]byte{0xff, 0x00}),
		"Equipped creature has \"{2}: This creature gets +1/+0.\"",
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		lexer := NewLexer(source)
		lastEnd := 0
		for count := 0; ; count++ {
			if count > len(source)+1 {
				t.Fatal("lexer did not terminate")
			}
			token := lexer.Next()
			if token.Span.Start.Offset < lastEnd ||
				token.Span.End.Offset < token.Span.Start.Offset ||
				token.Span.End.Offset > len(source) {
				t.Fatalf("invalid span %#v after byte %d", token.Span, lastEnd)
			}
			if token.Kind == EOF {
				if token.Span.Start.Offset != len(source) || token.Span.End.Offset != len(source) {
					t.Fatalf("EOF span = %#v, source length %d", token.Span, len(source))
				}
				return
			}
			if token.Span.Start.Offset == token.Span.End.Offset {
				t.Fatalf("non-EOF token did not consume input: %#v", token)
			}
			if token.Text != source[token.Span.Start.Offset:token.Span.End.Offset] {
				t.Fatalf("token text %q does not match its span", token.Text)
			}
			lastEnd = token.Span.End.Offset
		}
	})
}

func assertValidTokenStream(t *testing.T, name, source string) {
	t.Helper()
	lexer := NewLexer(source)
	lastEnd := 0
	for {
		token := lexer.Next()
		if token.Kind == Invalid {
			r, _ := utf8.DecodeRuneInString(token.Text)
			t.Fatalf("%s: invalid token %q (%U) at %#v", name, token.Text, r, token.Span)
		}
		if token.Span.Start.Offset < lastEnd || token.Span.End.Offset > len(source) {
			t.Fatalf("%s: invalid token span %#v", name, token.Span)
		}
		if token.Kind == EOF {
			return
		}
		lastEnd = token.Span.End.Offset
	}
}

func tokenSummary(source string) string {
	lexer := NewLexer(source)
	var tokens []string
	for {
		token := lexer.Next()
		text := strings.NewReplacer("\r", `\r`, "\n", `\n`).Replace(token.Text)
		tokens = append(tokens, fmt.Sprintf("%s(%s)", token.Kind, text))
		if token.Kind == EOF {
			return strings.Join(tokens, " ")
		}
	}
}

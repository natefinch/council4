package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// renderTestCards are representative cards exercising every lowered ability
// category through the full typed pipeline and deterministic renderer.
var renderTestCards = []*ScryfallCard{
	{
		Name:       "Render Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	},
	{
		Name:       "Render Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	},
	{
		Name:       "Render Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		Colors:     []string{"R"},
		OracleText: "Render Bolt deals 3 damage to any target.",
	},
}

func generateExecutable(t *testing.T, card *ScryfallCard) string {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(card, "cards")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource(%q): %v", card.Name, err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("GenerateExecutableCardSource(%q) diagnostics: %#v", card.Name, diagnostics)
	}
	return source
}

func TestRenderDeterministic(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			first := generateExecutable(t, card)
			for i := range 5 {
				again := generateExecutable(t, card)
				if again != first {
					t.Fatalf("render not deterministic on iteration %d", i)
				}
			}
		})
	}
}

func TestRenderParses(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			source := generateExecutable(t, card)
			if _, err := parser.ParseFile(token.NewFileSet(), "card.go", source, parser.AllErrors); err != nil {
				t.Fatalf("generated source does not parse: %v\n%s", err, source)
			}
		})
	}
}

func TestRenderNoTODO(t *testing.T) {
	t.Parallel()
	for _, card := range renderTestCards {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			source := generateExecutable(t, card)
			if strings.Contains(source, "TODO") {
				t.Fatalf("executable source unexpectedly contains TODO:\n%s", source)
			}
		})
	}
}

func TestRenderImportsDeterministicOrder(t *testing.T) {
	t.Parallel()
	source := generateExecutable(t, renderTestCards[1])
	start := strings.Index(source, "import (")
	if start < 0 {
		t.Fatalf("no import block found:\n%s", source)
	}
	end := strings.Index(source[start:], ")")
	block := source[start : start+end]
	var paths []string
	for line := range strings.SplitSeq(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"`) {
			paths = append(paths, line)
		}
	}
	for i := 1; i < len(paths); i++ {
		if paths[i-1] > paths[i] {
			t.Fatalf("imports not sorted: %q before %q", paths[i-1], paths[i])
		}
	}
}

// TestRenderUnsupportedReplacementErrors verifies the renderer returns an error
// (rather than silently omitting a field) when a CardDef contains a typed value
// the renderer cannot spell, here a non-EntersTapped replacement ability.
func TestRenderUnsupportedReplacementErrors(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Test",
			Types: []types.Card{types.Creature},
			ReplacementAbilities: []game.ReplacementAbility{
				{
					Text: "unsupported",
					Replacement: game.ReplacementEffect{
						EntersTapped: false,
						Condition:    opt.Val(game.Condition{Text: "some condition"}),
					},
				},
			},
		},
	}
	card := &ScryfallCard{Name: "Test", Layout: "normal", TypeLine: "Creature"}
	_, err := Renderer{}.RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err == nil {
		t.Fatal("expected error for unsupported replacement ability, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("error should mention 'unsupported', got: %v", err)
	}
}

// TestRenderHintDivergenceErrors verifies the renderer refuses to use a
// static-ability VarName hint whose recorded body diverges from the validated
// CardDef value, returning a divergence error instead of emitting a wrong var.
func TestRenderHintDivergenceErrors(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Test Bear", Layout: "normal", TypeLine: "Creature — Bear",
		OracleText: "Flying", Power: new("2"), Toughness: new("2"),
	}
	faceAbilities, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("lowering diagnostics: %v", diagnostics)
	}
	defs, err := assembleCardDefs(card, faceAbilities)
	if err != nil {
		t.Fatalf("assembleCardDefs: %v", err)
	}
	hints := []faceRenderHints{{
		StaticVarNames: []staticVarHint{{
			VarName: "game.FlyingStaticBody",
			Body:    game.VigilanceStaticBody,
		}},
	}}
	_, err = Renderer{}.RenderCardSource(card, defs, hints, "cards")
	if err == nil {
		t.Fatal("expected error for hint body divergence, got nil")
	}
	if !strings.Contains(err.Error(), "divergence") {
		t.Fatalf("error should mention 'divergence', got: %v", err)
	}
}

package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
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

func TestRenderUsesEquipMechanicTemplate(t *testing.T) {
	card := &ScryfallCard{Name: "Test Equipment", Layout: "normal", TypeLine: "Artifact — Equipment"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               card.Name,
		Types:              []types.Card{types.Artifact},
		Subtypes:           []types.Sub{types.Equipment},
		ActivatedAbilities: []game.ActivatedAbility{game.EquipActivatedAbility(cost.Mana{cost.O(2)})},
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.EquipActivatedAbility(cost.Mana{cost.O(2)})") {
		t.Fatalf("source does not use Equip template:\n%s", source)
	}
}

func TestRenderUsesEnchantMechanicTemplate(t *testing.T) {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "creature",
		Allow:      game.TargetAllowPermanent,
		Predicate: game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
		},
	}
	card := &ScryfallCard{Name: "Test Aura", Layout: "normal", TypeLine: "Enchantment — Aura"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:            card.Name,
		Types:           []types.Card{types.Enchantment},
		Subtypes:        []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&target)},
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.EnchantStaticAbility(&game.TargetSpec{") {
		t.Fatalf("source does not use Enchant template:\n%s", source)
	}
}

func TestRenderTargetPredicateQualifiers(t *testing.T) {
	ctx := newRenderCtx()
	lit, ok, err := (Renderer{}).renderTargetPredicate(ctx, game.TargetPredicate{
		PermanentTypes: []types.Card{types.Creature},
		Controller:     game.ControllerYou,
		Tapped:         game.TriTrue,
		CombatState:    game.CombatStateAttacking,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("renderTargetPredicate() did not render qualified predicate")
	}
	for _, want := range []string{
		"Controller: game.ControllerYou",
		"Tapped: game.TriTrue",
		"CombatState: game.CombatStateAttacking",
	} {
		if !strings.Contains(lit, want) {
			t.Fatalf("predicate literal %q does not contain %q", lit, want)
		}
	}
}

func TestRenderUsesProtectionMechanicTemplate(t *testing.T) {
	card := &ScryfallCard{Name: "Test Bear", Layout: "normal", TypeLine: "Creature — Bear"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  card.Name,
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			game.ProtectionFromColorsStaticAbility(color.Black, color.Red),
		},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}

	source, err := (Renderer{}).RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "game.ProtectionFromColorsStaticAbility(color.Black, color.Red)") {
		t.Fatalf("source does not use Protection template:\n%s", source)
	}
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

func TestRenderUnsupportedAbilityLayerFieldsErrors(t *testing.T) {
	t.Parallel()
	tests := map[string]game.ContinuousEffect{
		"unsupported field": {
			Layer:          game.LayerAbility,
			Group:          game.BattlefieldGroup(game.Selection{}),
			RemoveKeywords: []game.Keyword{game.Flying},
		},
		"PT field in ability layer": {
			Layer:      game.LayerAbility,
			Group:      game.BattlefieldGroup(game.Selection{}),
			PowerDelta: 1,
		},
		"keyword field in PT layer": {
			Layer:       game.LayerPowerToughnessModify,
			Group:       game.BattlefieldGroup(game.Selection{}),
			AddKeywords: []game.Keyword{game.Flying},
		},
	}
	for name, effect := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			def := &game.CardDef{
				CardFace: game.CardFace{
					Name:  "Test",
					Types: []types.Card{types.Enchantment},
					StaticAbilities: []game.StaticAbility{{
						ContinuousEffects: []game.ContinuousEffect{effect},
					}},
				},
			}
			card := &ScryfallCard{Name: "Test", Layout: "normal", TypeLine: "Enchantment"}
			_, err := Renderer{}.RenderCardSource(card, []*game.CardDef{def}, []faceRenderHints{{}}, "cards")
			if err == nil {
				t.Fatal("expected error for incompatible continuous-effect fields")
			}
		})
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

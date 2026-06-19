package cardgen

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRenderSelection(t *testing.T) {
	t.Parallel()
	r := Renderer{}
	tests := []struct {
		name      string
		selection game.Selection
		wantErr   bool
		wantParts []string
	}{
		{
			name:      "empty selection",
			selection: game.Selection{},
			wantParts: []string{"game.Selection{}"},
		},
		{
			name:      "required types",
			selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			wantParts: []string{"RequiredTypes:", "types.Creature"},
		},
		{
			name:      "required types any",
			selection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
			wantParts: []string{"RequiredTypesAny:", "types.Artifact", "types.Creature"},
		},
		{
			name:      "excluded types",
			selection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
			wantParts: []string{"ExcludedTypes:", "types.Land"},
		},
		{
			name:      "supertypes",
			selection: game.Selection{Supertypes: []types.Super{types.Legendary}},
			wantParts: []string{"Supertypes:", "types.Legendary"},
		},
		{
			name:      "excluded supertype",
			selection: game.Selection{ExcludedSupertype: types.Basic},
			wantParts: []string{"ExcludedSupertype: types.Basic"},
		},
		{
			name:      "subtypes any",
			selection: game.Selection{SubtypesAny: []types.Sub{"Goblin"}},
			wantParts: []string{"SubtypesAny:"},
		},
		{
			name:      "colors any",
			selection: game.Selection{ColorsAny: []color.Color{color.Red, color.Green}},
			wantParts: []string{"ColorsAny:", "color.Red", "color.Green"},
		},
		{
			name:      "excluded colors",
			selection: game.Selection{ExcludedColors: []color.Color{color.Blue}},
			wantParts: []string{"ExcludedColors:", "color.Blue"},
		},
		{
			name:      "controller you",
			selection: game.Selection{Controller: game.ControllerYou},
			wantParts: []string{"Controller: game.ControllerYou"},
		},
		{
			name:      "controller opponent",
			selection: game.Selection{Controller: game.ControllerOpponent},
			wantParts: []string{"game.ControllerOpponent"},
		},
		{
			name:      "controller not you",
			selection: game.Selection{Controller: game.ControllerNotYou},
			wantParts: []string{"game.ControllerNotYou"},
		},
		{
			name:      "player relation",
			selection: game.Selection{Player: game.PlayerOpponent},
			wantParts: []string{"Player:"},
		},
		{
			name:      "tapped true",
			selection: game.Selection{Tapped: game.TriTrue},
			wantParts: []string{"Tapped: game.TriTrue"},
		},
		{
			name:      "tapped false",
			selection: game.Selection{Tapped: game.TriFalse},
			wantParts: []string{"Tapped: game.TriFalse"},
		},
		{
			name:      "combat attacking",
			selection: game.Selection{CombatState: game.CombatStateAttacking},
			wantParts: []string{"CombatState: game.CombatStateAttacking"},
		},
		{
			name:      "combat blocking",
			selection: game.Selection{CombatState: game.CombatStateBlocking},
			wantParts: []string{"game.CombatStateBlocking"},
		},
		{
			name:      "combat attacking or blocking",
			selection: game.Selection{CombatState: game.CombatStateAttackingOrBlocking},
			wantParts: []string{"game.CombatStateAttackingOrBlocking"},
		},
		{
			name:      "keyword flying",
			selection: game.Selection{Keyword: game.Flying},
			wantParts: []string{"Keyword: game.Flying"},
		},
		{
			name:      "excluded keyword deathtouch",
			selection: game.Selection{ExcludedKeyword: game.Deathtouch},
			wantParts: []string{"ExcludedKeyword: game.Deathtouch"},
		},
		{
			name:      "mana value equal",
			selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: 3})},
			wantParts: []string{"ManaValue:", "compare.Equal", "Value: 3"},
		},
		{
			name:      "power less or equal",
			selection: game.Selection{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},
			wantParts: []string{"Power:", "compare.LessOrEqual"},
		},
		{
			name:      "toughness greater than",
			selection: game.Selection{Toughness: opt.Val(compare.Int{Op: compare.GreaterThan, Value: 0})},
			wantParts: []string{"Toughness:", "compare.GreaterThan"},
		},
		{
			name:      "exclude source",
			selection: game.Selection{ExcludeSource: true},
			wantParts: []string{"ExcludeSource: true"},
		},
		{
			name:      "non token",
			selection: game.Selection{NonToken: true},
			wantParts: []string{"NonToken: true"},
		},
		{
			name: "combined selection",
			selection: game.Selection{
				RequiredTypesAny: []types.Card{types.Creature},
				ColorsAny:        []color.Color{color.White},
				Controller:       game.ControllerYou,
				Tapped:           game.TriTrue,
				ExcludeSource:    true,
				NonToken:         true,
			},
			wantParts: []string{
				"RequiredTypesAny:", "types.Creature",
				"ColorsAny:", "color.White",
				"Controller: game.ControllerYou",
				"Tapped: game.TriTrue",
				"ExcludeSource: true",
				"NonToken: true",
			},
		},
		{
			name:      "unknown controller relation",
			selection: game.Selection{Controller: game.ControllerRelation(999)},
			wantErr:   true,
		},
		{
			name:      "unknown tri-state",
			selection: game.Selection{Tapped: game.TriState(999)},
			wantErr:   true,
		},
		{
			name:      "unknown combat state",
			selection: game.Selection{CombatState: game.CombatStateFilter(999)},
			wantErr:   true,
		},
		{
			name:      "unknown keyword",
			selection: game.Selection{Keyword: game.Keyword(9999)},
			wantErr:   true,
		},
		{
			name:      "unknown excluded keyword",
			selection: game.Selection{ExcludedKeyword: game.Keyword(9999)},
			wantErr:   true,
		},
		{
			name:      "unknown color in ColorsAny",
			selection: game.Selection{ColorsAny: []color.Color{"Purple"}},
			wantErr:   true,
		},
		{
			name:      "unknown color in ExcludedColors",
			selection: game.Selection{ExcludedColors: []color.Color{"Gold"}},
			wantErr:   true,
		},
		{
			name:      "unknown compare op in ManaValue",
			selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Op(999), Value: 1})},
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := newRenderCtx()
			got, err := r.renderSelection(ctx, tc.selection)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("renderSelection(%v): want error, got %q", tc.selection, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("renderSelection(%v): unexpected error: %v", tc.selection, err)
			}
			for _, part := range tc.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("renderSelection output missing %q; got:\n%s", part, got)
				}
			}
		})
	}
}

func TestRenderColorValueError(t *testing.T) {
	t.Parallel()
	_, err := colorValueToLiteral("Purple")
	if err == nil {
		t.Fatal("colorValueToLiteral(Purple): want error, got nil")
	}
}

func TestRenderManaSymbolError(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	_, err := renderManaSymbol(ctx, cost.Symbol{Kind: cost.SymbolKind(999)})
	if err == nil {
		t.Fatal("renderManaSymbol(unknown kind): want error, got nil")
	}
}

func TestRenderKeywordAllValues(t *testing.T) {
	t.Parallel()
	keywords := []game.Keyword{
		game.KeywordNone,
		game.Deathtouch,
		game.Defender,
		game.DoubleStrike,
		game.FirstStrike,
		game.Flash,
		game.Flying,
		game.Haste,
		game.Hexproof,
		game.Indestructible,
		game.Lifelink,
		game.Menace,
		game.Protection,
		game.Reach,
		game.Shroud,
		game.Trample,
		game.Vigilance,
		game.Ward,
		game.SplitSecond,
		game.Equip,
		game.Enchant,
		game.Cycling,
		game.Flashback,
		game.Kicker,
		game.Madness,
		game.Morph,
		game.Disguise,
		game.Convoke,
		game.Delve,
		game.Suspend,
		game.Storm,
		game.Cascade,
		game.Prowess,
		game.Mutate,
		game.Companion,
		game.Ninjutsu,
		game.Escape,
		game.Foretell,
		game.Craft,
		game.Discover,
		game.Eternalize,
		game.Affinity,
		game.Improvise,
		game.Emerge,
		game.Undying,
		game.Persist,
		game.Wither,
		game.Infect,
		game.Toxic,
		game.Annihilator,
		game.Exalted,
		game.ReadAhead,
		game.Horsemanship,
	}
	for i, kw := range keywords {
		t.Run(fmt.Sprintf("keyword_%d", i), func(t *testing.T) {
			t.Parallel()
			got, err := renderKeyword(kw)
			if err != nil {
				t.Fatalf("renderKeyword(%d): unexpected error: %v", kw, err)
			}
			if !strings.HasPrefix(got, "game.") {
				t.Errorf("renderKeyword(%d) = %q, want game.Xxx prefix", kw, got)
			}
		})
	}
}

func TestRenderCompareOpAllValues(t *testing.T) {
	t.Parallel()
	ops := []compare.Op{
		compare.Any,
		compare.Equal,
		compare.LessOrEqual,
		compare.GreaterOrEqual,
		compare.LessThan,
		compare.GreaterThan,
	}
	for _, op := range ops {
		t.Run("op", func(t *testing.T) {
			t.Parallel()
			got, err := renderCompareOp(op)
			if err != nil {
				t.Fatalf("renderCompareOp(%v): unexpected error: %v", op, err)
			}
			if !strings.HasPrefix(got, "compare.") {
				t.Errorf("renderCompareOp(%v) = %q, want compare.Xxx prefix", op, got)
			}
		})
	}
}

func TestRenderCompareOpUnknown(t *testing.T) {
	t.Parallel()
	_, err := renderCompareOp(compare.Op(9999))
	if err == nil {
		t.Fatal("renderCompareOp(unknown): want error, got nil")
	}
}

func TestRenderPlayerRelation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		relation game.PlayerRelation
		want     string
	}{
		{game.PlayerAny, "game.PlayerAny"},
		{game.PlayerYou, "game.PlayerYou"},
		{game.PlayerOpponent, "game.PlayerOpponent"},
		{game.PlayerNotYou, "game.PlayerNotYou"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got, err := renderPlayerRelation(tc.relation)
			if err != nil {
				t.Fatalf("renderPlayerRelation(%d): unexpected error: %v", tc.relation, err)
			}
			if got != tc.want {
				t.Errorf("renderPlayerRelation(%d) = %q, want %q", tc.relation, got, tc.want)
			}
		})
	}
}

func TestRenderPlayerRelationUnknown(t *testing.T) {
	t.Parallel()
	_, err := renderPlayerRelation(game.PlayerRelation(9999))
	if err == nil {
		t.Fatal("renderPlayerRelation(unknown): want error, got nil")
	}
}

func TestRenderSelectionUnknownPlayerRelation(t *testing.T) {
	t.Parallel()
	r := Renderer{}
	ctx := newRenderCtx()
	_, err := r.renderSelection(ctx, game.Selection{Player: game.PlayerRelation(9999)})
	if err == nil {
		t.Fatal("renderSelection with unknown PlayerRelation: want error, got nil")
	}
}

func TestCardTypeLiteralUnknown(t *testing.T) {
	t.Parallel()
	_, err := cardTypeLiteral(types.Card("Conspiracy"))
	// Conspiracy is known; test a truly unknown type
	if err != nil {
		t.Fatalf("cardTypeLiteral(Conspiracy): unexpected error: %v", err)
	}
	_, err = cardTypeLiteral(types.Card("NotARealType"))
	if err == nil {
		t.Fatal("cardTypeLiteral(unknown): want error, got nil")
	}
}

func TestRenderTypesCardSliceUnknown(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	_, err := renderTypesCardSlice(ctx, []types.Card{"NotARealType"})
	if err == nil {
		t.Fatal("renderTypesCardSlice(unknown type): want error, got nil")
	}
}

func TestRenderSelectionUnknownCardType(t *testing.T) {
	t.Parallel()
	r := Renderer{}
	ctx := newRenderCtx()
	_, err := r.renderSelection(ctx, game.Selection{RequiredTypes: []types.Card{"NotARealType"}})
	if err == nil {
		t.Fatal("renderSelection with unknown card type: want error, got nil")
	}
}

func TestSupertypeLiteralUnknown(t *testing.T) {
	t.Parallel()
	_, err := supertypeLiteral(types.Super("NotAReal"))
	if err == nil {
		t.Fatal("supertypeLiteral(unknown): want error, got nil")
	}
}

func TestRenderSelectionUnknownSupertype(t *testing.T) {
	t.Parallel()
	r := Renderer{}
	ctx := newRenderCtx()
	_, err := r.renderSelection(ctx, game.Selection{Supertypes: []types.Super{"NotAReal"}})
	if err == nil {
		t.Fatal("renderSelection with unknown supertype: want error, got nil")
	}
}

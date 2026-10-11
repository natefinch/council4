package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const ownedLandSearch = "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle."

func TestOwnedSearchPositions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text string
		kinds      []game.PrimitiveKind
		search     int
	}{
		{"alone", ownedLandSearch, []game.PrimitiveKind{game.PrimitiveSearch}, 0},
		{"last", "Draw a card. " + ownedLandSearch, []game.PrimitiveKind{game.PrimitiveDraw, game.PrimitiveSearch}, 1},
		{"first", ownedLandSearch + " Scry 1.", []game.PrimitiveKind{game.PrimitiveSearch, game.PrimitiveScry}, 0},
		{"middle", "Draw a card. " + ownedLandSearch + " Scry 1.", []game.PrimitiveKind{game.PrimitiveDraw, game.PrimitiveSearch, game.PrimitiveScry}, 1},
		{"life riders", "You gain 2 life. " + ownedLandSearch + " You lose 1 life.", []game.PrimitiveKind{game.PrimitiveGainLife, game.PrimitiveSearch, game.PrimitiveLoseLife}, 1},
		{"target riders", "Tap target creature. " + ownedLandSearch + " Untap target land.", []game.PrimitiveKind{game.PrimitiveTap, game.PrimitiveSearch, game.PrimitiveUntap}, 1},
		{"token rider", "Draw a card. " + ownedLandSearch + " Create a 1/1 white Soldier creature token.", []game.PrimitiveKind{game.PrimitiveDraw, game.PrimitiveSearch, game.PrimitiveCreateToken}, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Owned Search", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v %#v; definitions=%d", err, diagnostics, len(defs))
			}
			sequence := defs[0].SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != len(tt.kinds) {
				t.Fatalf("instructions=%d, want %d", len(sequence), len(tt.kinds))
			}
			for i, kind := range tt.kinds {
				if sequence[i].Primitive.Kind() != kind {
					t.Fatalf("instruction %d: %v, want %v", i, sequence[i].Primitive.Kind(), kind)
				}
			}
			prefix := fmt.Sprintf("SpellAbility.Val.Modes[0].Sequence[%d].Primitive.(game.Search)", tt.search)
			assertCardPaths(t, card,
				prefix+".Amount.fixed = 1",
				prefix+".Spec.SourceZone = zone.Library",
				prefix+".Spec.Destination = zone.Battlefield",
				prefix+".Spec.EntersTapped = true",
				prefix+".Spec.Filter.RequiredTypes[0] = types.Land",
				prefix+".Spec.Filter.Supertypes[0] = types.Basic",
			)
		})
	}
}

func TestOwnedSearchProductsAndConditions(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text string
		paths      []string
	}{
		{"linked after rider", ownedLandSearch + " You gain 2 life. Untap that land.", []string{
			`SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Search).PublishLinked = "sequence-effect-0-product"`,
			`SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.Untap).Object.linkID = "sequence-effect-0-product"`,
		}},
		{"nonzero producer", "Tap target creature. " + ownedLandSearch + " You gain 2 life. Untap that land. Untap target land.", []string{
			`SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.Search).PublishLinked = "sequence-effect-1-product"`,
			`SpellAbility.Val.Modes[0].Sequence[3].Primitive.(game.Untap).Object.linkID = "sequence-effect-1-product"`,
			"SpellAbility.Val.Modes[0].Sequence[4].Primitive.(game.Untap).Object.targetIndex = 1",
		}},
		{"independent products", ownedLandSearch + " You gain 2 life. Untap that land. " + ownedLandSearch + " Scry 1. Untap that land.", []string{
			`SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Search).PublishLinked = "sequence-effect-0-product"`,
			`SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.Untap).Object.linkID = "sequence-effect-0-product"`,
			`SpellAbility.Val.Modes[0].Sequence[3].Primitive.(game.Search).PublishLinked = "sequence-effect-5-product"`,
			`SpellAbility.Val.Modes[0].Sequence[5].Primitive.(game.Untap).Object.linkID = "sequence-effect-5-product"`,
		}},
		{"group condition", "If you have no cards in hand, " + ownedLandSearch + " If you have no cards in hand, draw a card.", []string{
			"SpellAbility.Val.Modes[0].Sequence[0].Condition.Val.Condition.Val.ControllerHandEmpty = true",
			"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.ControllerHandEmpty = true",
		}},
		{"conditional linked consumer", ownedLandSearch + " Then if you control four or more lands, untap that land.", []string{
			`SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Search).PublishLinked = "sequence-effect-0-product"`,
			"SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.MinCount = 4",
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{Name: "Search Product", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}, tt.paths...)
		})
	}
}

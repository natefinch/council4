package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRenderLookAtLibraryTopPrimitive(t *testing.T) {
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.LookAtLibraryTop{
		Player:        game.ControllerReference(),
		PublishLinked: game.LinkedKey("chosen-type-top"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.LookAtLibraryTop",
		"Player: game.ControllerReference()",
		`PublishLinked: game.LinkedKey("chosen-type-top")`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered LookAtLibraryTop missing %q:\n%s", want, rendered)
		}
	}
	assertParsesAsGoExpr(t, rendered)
}

func TestRenderRevealOfLinkedCard(t *testing.T) {
	rendered, err := (Renderer{}).renderPrimitive(newRenderCtx(), game.Reveal{
		Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "chosen-type-top"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.Reveal",
		"game.CardReferenceLinked",
		`LinkID: "chosen-type-top"`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered Reveal missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Player:") || strings.Contains(rendered, "Amount:") {
		t.Fatalf("card-form Reveal rendered player/amount fields:\n%s", rendered)
	}
	assertParsesAsGoExpr(t, rendered)
}

func TestRenderCardConditionChosenSubtype(t *testing.T) {
	rendered, err := (Renderer{}).renderCardCondition(newRenderCtx(), game.CardCondition{
		Card:              game.CardReference{Kind: game.CardReferenceLinked, LinkID: "chosen-type-top"},
		Types:             []types.Card{types.Creature},
		ChosenSubtypeFrom: game.EntryTypeChoiceKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "ChosenSubtypeFrom: game.EntryTypeChoiceKey") {
		t.Fatalf("rendered CardCondition missing chosen-subtype provenance:\n%s", rendered)
	}
	assertParsesAsGoExpr(t, rendered)
}

func TestRenderCostModifierChosenSubtype(t *testing.T) {
	rendered, err := (Renderer{}).renderCostModifier(newRenderCtx(), game.CostModifier{
		Kind:                         game.CostModifierSpell,
		MatchCardType:                true,
		CardType:                     types.Creature,
		ChosenSubtypeFromEntryChoice: true,
		GenericReduction:             1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "ChosenSubtypeFromEntryChoice: true") {
		t.Fatalf("rendered CostModifier missing chosen-subtype provenance:\n%s", rendered)
	}
	assertParsesAsGoExpr(t, rendered)
}

func TestRenderCostModifierColorDisjunction(t *testing.T) {
	rendered, err := (Renderer{}).renderCostModifier(newRenderCtx(), game.CostModifier{
		Kind:             game.CostModifierSpell,
		GenericReduction: 1,
		MatchColors:      []color.Color{color.Red, color.Green},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "MatchColors: []color.Color{color.Red, color.Green}") {
		t.Fatalf("rendered CostModifier missing color disjunction:\n%s", rendered)
	}
	assertParsesAsGoExpr(t, rendered)
}

func assertParsesAsGoExpr(t *testing.T, rendered string) {
	t.Helper()
	if _, err := parser.ParseExprFrom(token.NewFileSet(), "", rendered, 0); err != nil {
		t.Fatalf("rendered output is not a valid Go expression: %v\n%s", err, rendered)
	}
}

package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const anchorOracle = "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land."

func TestLowerFabledPassageEndToEnd(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Fabled Passage",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: anchorOracle,
	}
	face := lowerSingleFace(t, card)
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ActivationCondition.Exists {
		t.Fatalf("activation condition = %#v, want resolving condition only", ability.ActivationCondition)
	}
	if len(ability.AdditionalCosts) != 2 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalTap ||
		ability.AdditionalCosts[1].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("additional costs = %#v, want tap then sacrifice source", ability.AdditionalCosts)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want search then conditional untap", mode.Sequence)
	}
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.Search", mode.Sequence[0].Primitive)
	}
	if search.Amount.Value() != 1 ||
		search.PublishLinked == "" ||
		search.Spec.SourceZone != zone.Library ||
		search.Spec.Destination != zone.Battlefield ||
		!search.Spec.EntersTapped ||
		!search.Spec.CardType.Exists ||
		search.Spec.CardType.Val != types.Land {
		t.Fatalf("search = %#v", search)
	}
	untap, ok := mode.Sequence[1].Primitive.(game.Untap)
	if !ok || untap.Object.Kind() != game.ObjectReferenceLinkedObject {
		t.Fatalf("sequence[1] = %#v, want linked untap", mode.Sequence[1])
	}
	if untap.Object.LinkID() != string(search.PublishLinked) {
		t.Fatalf("untap link = %q, search link = %q", untap.Object.LinkID(), search.PublishLinked)
	}
	if !mode.Sequence[1].Condition.Exists ||
		!mode.Sequence[1].Condition.Val.Condition.Exists ||
		!mode.Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Exists ||
		mode.Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.MinCount != 4 {
		t.Fatalf("untap condition = %#v, want controller controls at least four lands", mode.Sequence[1].Condition)
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("generate source: err=%v diagnostics=%#v", err, diagnostics)
	}
	for _, want := range []string{
		`PublishLinked: game.LinkedKey("searched-land-1")`,
		`game.LinkedObjectReference("searched-land-1")`,
		"MinCount:  4",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLinkedSearchUntapBoundariesFailClosed(t *testing.T) {
	t.Parallel()
	tests := []string{
		"{T}, Sacrifice this land: Search your library for a basic land card, put it into your hand, then shuffle. Then if you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic creature card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for two basic land cards, put them onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield, then shuffle. Then if you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. If you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Before you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control exactly four lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or fewer lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more basic lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a land card with mana value 2 or less, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap target land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that basic land.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that creature.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Boundary Test",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: oracleText,
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported diagnostic")
			}
		})
	}
}

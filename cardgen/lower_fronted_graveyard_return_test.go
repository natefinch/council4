package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func scrapTrawlerCard() *ScryfallCard {
	power, toughness := "3", "2"
	return &ScryfallCard{
		Name:     "Scrap Trawler",
		Layout:   "normal",
		ManaCost: "{3}",
		TypeLine: "Artifact Creature — Construct",
		OracleText: "Whenever this creature dies or another artifact you control is put into a graveyard " +
			"from the battlefield, return to your hand target artifact card in your graveyard with lesser mana value.",
		Power:     &power,
		Toughness: &toughness,
	}
}

// TestLowerScrapTrawlerFrontedReturnCarriesEventRelativeManaValueBound proves the
// fronted-destination hand return "return to your hand target artifact card in
// your graveyard with lesser mana value" lowers, under Scrap Trawler's
// self-or-another-artifact battlefield-to-graveyard union trigger, to a
// graveyard-card target whose Selection carries the event-relative
// ManaValueLessThanEventPermanent bound alongside the artifact filter and "your"
// owner, returning the target from the graveyard to its owner's hand.
func TestLowerScrapTrawlerFrontedReturnCarriesEventRelativeManaValueBound(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, scrapTrawlerCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventZoneChanged || !pattern.SubjectSelectionOrSelf {
		t.Fatalf("trigger pattern = %#v, want self-or-subject zone change", pattern)
	}
	if !pattern.MatchFromZone || pattern.FromZone != zone.Battlefield ||
		!pattern.MatchToZone || pattern.ToZone != zone.Graveyard {
		t.Fatalf("trigger zones = %#v, want battlefield to graveyard", pattern)
	}
	if !slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("subject selection types = %v, want [Artifact]", pattern.SubjectSelection.RequiredTypes)
	}

	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target zone/allow = %#v, want graveyard card", target)
	}
	sel := target.Selection.Val
	if !sel.ManaValueLessThanEventPermanent {
		t.Fatal("target selection must carry ManaValueLessThanEventPermanent")
	}
	if !slices.Equal(sel.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("required types = %v, want [Artifact]", sel.RequiredTypes)
	}
	if sel.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", sel.Controller)
	}

	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %#v, want MoveCard", mode.Sequence[0].Primitive)
	}
	if move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("move = from %v to %v, want graveyard to hand", move.FromZone, move.Destination)
	}
}

// TestGenerateScrapTrawlerRendersEventRelativeManaValueBound proves the generated
// source renders the ManaValueLessThanEventPermanent field for the fronted hand
// return, exercising the compiler field copy and the render_target projection end
// to end.
func TestGenerateScrapTrawlerRendersEventRelativeManaValueBound(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(scrapTrawlerCard(), "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "ManaValueLessThanEventPermanent: true") {
		t.Fatalf("generated source missing ManaValueLessThanEventPermanent:\n%s", source)
	}
}

// TestLowerFrontedEqualOrLesserGraveyardReturnFailsClosed proves the fronted hand
// return with the ≤ wording "with equal or lesser mana value", which the strict
// event-relative bound must not absorb, stays unsupported rather than silently
// lowering to the strict bound or dropping the qualifier.
func TestLowerFrontedEqualOrLesserGraveyardReturnFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Fronted Equal Or Lesser Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return to your hand target artifact card in your graveyard with equal or lesser mana value.",
	})
}

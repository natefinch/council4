package cardgen

import (
	"strings"
	"testing"
)

func TestSharedTargetCountAdapters(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want []string
	}{
		{"Destroy X target artifacts.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Sequence[0].Primitive.(game.Destroy).Object.kind = game.ObjectReferenceAllTargetPermanents",
		}},
		{"Tap X target creatures you don't control.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Targets[0].Selection.Val.Controller = game.ControllerNotYou",
			"Modes[0].Sequence[0].Primitive.(game.Tap)",
		}},
		{"Untap X target lands.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Sequence[19].Primitive.(game.Untap).Object.targetIndex = 19",
		}},
		{"Return X target nonland permanents to their owners' hands.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Sequence[0].Primitive.(game.MovePermanent).Destination = zone.Hand",
		}},
		{"Exile X target cards from a single graveyard.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Targets[0].SameGraveyard = true",
			"Modes[0].Sequence[19].Primitive.(game.MoveCard).Destination = zone.Exile",
		}},
		{"Destroy any number of target artifacts and/or enchantments.", []string{
			"Modes[0].Targets[0].MaxTargets = 99",
			"Modes[0].Sequence[0].Primitive.(game.Destroy).Object.kind = game.ObjectReferenceAllTargetPermanents",
		}},
		{"Destroy X target artifacts. Draw X cards.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[1].Primitive.(game.Draw)",
		}},
		{"Destroy X target artifacts. You gain 2 life. Draw X cards.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[1].Primitive.(game.GainLife)",
			"Modes[0].Sequence[2].Primitive.(game.Draw)",
		}},
		{"Destroy X target artifacts. Create X 2/2 red Gremlin creature tokens.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[1].Primitive.(game.CreateToken)",
		}},
		{"Tap X target creatures. Scry 1.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Tap)",
			"Modes[0].Sequence[1].Primitive.(game.Scry).Amount.fixed = 1",
		}},
		{"Count Parameters deals 4 damage to each of X targets.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Sequence[0].Primitive.(game.Damage).EachTarget = true",
		}},
		{"Target player draws a card. Destroy X target artifacts.", []string{
			"Modes[0].Targets[1].CountEqualsX = true",
			"Modes[0].Sequence[1].Primitive.(game.Destroy).Object.targetIndex = 1",
		}},
		{"If you control an artifact, destroy X target creatures.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[0].Condition.Exists = true",
		}},
	} {
		t.Run(tc.body, func(t *testing.T) {
			t.Parallel()
			card := targetCountTestCard(tc.body)
			assertCardPaths(t, card, tc.want...)
			assertCardPathsAbsent(t, card, "Modes[1]", "Targets[2]")
		})
	}
}

func TestSharedTargetCountsFailClosed(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body   string
		reason string
	}{
		{"Destroy up to X target artifacts.", "up-to-X"},
		{"Exile up to X target cards from a single graveyard.", "up-to-X"},
		{"Destroy X target artifacts. Target player draws a card.", "target-spec identity"},
		{"Exile any number of target artifacts. Target player draws a card.", "target-spec identity"},
		{"Destroy two target artifacts. Tap X target creatures.", "target-spec identity"},
		{"Destroy up to one target artifact. Tap X target creatures.", "target-spec identity"},
		{"Destroy X target artifacts, where X is the number of creatures you control.", "derived X"},
		{"Destroy twice X target artifacts.", "unsupported"},
		{"Destroy X target artifacts that entered this turn.", "unsupported"},
		{"Untap any number of target lands.", "unsupported"},
		{"Return X target creatures you own to your hand.", "unsupported"},
		{"Destroy X target artifacts. Draw a card for each artifact destroyed this way.", "unsupported"},
		{"X target players each mill a card.", "unsupported"},
		{"Exile X target creature cards from your graveyard. Then for each creature card exiled this way, create a 2/2 black Zombie creature token.", "unsupported"},
		{"Choose two —\n• Destroy X target artifacts.\n• Target player draws a card.", "target-spec identity"},
		{"Choose two —\n• Target player draws a card.\n• Destroy any number of target artifacts.", "target-spec identity"},
		{"Choose two —\n• Target player draws a card.\n• Exile any number of target artifacts.", "target-spec identity"},
	} {
		t.Run(tc.body, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, targetCountTestCard(tc.body), tc.reason)
		})
	}
}

func targetCountTestCard(body string) *ScryfallCard {
	return &ScryfallCard{
		Name: "Count Parameters", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{X}{B}",
		OracleText: body,
	}
}

func TestTargetCountDiagnosticUsesParserParameterSpan(t *testing.T) {
	t.Parallel()
	body := "Draw a card. Destroy up to X target artifacts."
	_, diagnostics, err := CompileCardDefs(targetCountTestCard(body))
	if err != nil {
		t.Fatal(err)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Summary != "unsupported target count parameter" {
			continue
		}
		start, end := diagnostic.Span.Start.Offset, diagnostic.Span.End.Offset
		if start != strings.Index(body, "up to X") || body[start:end] != "up to X" {
			t.Fatalf("parameter span=%+v text=%q", diagnostic.Span, body[start:end])
		}
		return
	}
	t.Fatalf("missing target-count diagnostic: %+v", diagnostics)
}

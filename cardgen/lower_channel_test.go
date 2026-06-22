package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerChannelActivationFromHand(t *testing.T) {
	t.Parallel()

	const oracleText = "Channel — {1}{G}, Discard this card: Destroy target artifact, enchantment, or nonbasic land an opponent controls. That player may search their library for a land card with a basic land type, put it onto the battlefield, then shuffle. This ability costs {1} less to activate for each legendary creature you control."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Channel Card",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: oracleText,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}

	body := face.ActivatedAbilities[0]
	if game.BodyFunctionZone(&body) != zone.Hand {
		t.Fatalf("function zone = %v, want hand", game.BodyFunctionZone(&body))
	}
	if len(body.AdditionalCosts) != 1 ||
		body.AdditionalCosts[0].Kind != cost.AdditionalDiscard ||
		body.AdditionalCosts[0].Text != "Discard this card" {
		t.Fatalf("additional costs = %#v, want discard this card", body.AdditionalCosts)
	}
	targets := game.BodyTargets(&body)
	if len(targets) != 1 || !targets[0].Selection.Exists {
		t.Fatalf("targets = %#v, want one Selection target", targets)
	}
	selection := targets[0].Selection.Val
	if selection.Controller != game.ControllerOpponent || len(selection.AnyOf) != 3 {
		t.Fatalf("selection = %#v, want opponent-controlled three-way union", selection)
	}
	if selection.AnyOf[2].ExcludedSupertype != types.Basic {
		t.Fatalf("land alternative = %#v, want nonbasic", selection.AnyOf[2])
	}
	if len(body.CostModifiers) != 1 ||
		body.CostModifiers[0].PerObjectReduction != 1 ||
		body.CostModifiers[0].CountSelection.Controller != game.ControllerYou {
		t.Fatalf("cost modifiers = %#v, want {1} per controlled legendary creature", body.CostModifiers)
	}
	sequence := body.Content.Modes[0].Sequence
	if len(sequence) != 2 || !sequence[1].Optional {
		t.Fatalf("sequence = %#v, want removal then optional search", sequence)
	}
	search, ok := sequence[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want search", sequence[1].Primitive)
	}
	if len(search.Spec.Filter.Supertypes) != 0 ||
		len(search.Spec.Filter.SubtypesAny) != 5 ||
		search.Player != game.ObjectControllerReference(game.TargetPermanentReference(0)) {
		t.Fatalf("search = %#v, want affected controller's basic-land-type search", search)
	}
}

func TestLowerBoseijuChannelNearMissesFailClosed(t *testing.T) {
	t.Parallel()

	for _, oracleText := range []string{
		"Channel — {1}{G}, Discard a card: Draw a card.",
		"Channel — {1}{G}, Discard this card: Destroy target artifact, creature, or nonbasic land an opponent controls. That player may search their library for a land card with a basic land type, put it onto the battlefield, then shuffle.",
		"Channel — {1}{G}, Discard this card: Destroy target artifact, enchantment, or nonbasic land an opponent controls. That player may search their library for a basic land card, put it onto the battlefield, then shuffle.",
		"Channel — {1}{G}, Discard this card: Destroy target artifact, enchantment, or nonbasic land an opponent controls. Its owner may search their library for a land card with a basic land type, put it onto the battlefield, then shuffle.",
		"Channel — {1}{G}, Discard this card: Destroy target artifact, enchantment, or nonbasic land an opponent controls. That player may search their library for a land card with a basic land type, put it onto the battlefield, then shuffle. You gain 2 life.",
	} {
		faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Channel Near Miss",
			Layout:     "normal",
			TypeLine:   "Legendary Land",
			OracleText: oracleText,
		})
		lowered := false
		for i := range faces {
			lowered = lowered || len(faces[i].ActivatedAbilities) != 0
		}
		if len(diagnostics) == 0 || lowered {
			t.Errorf("near miss unexpectedly lowered: %q; faces=%#v diagnostics=%#v", oracleText, faces, diagnostics)
		}
	}
}

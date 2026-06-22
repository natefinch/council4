package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// searchInstruction returns the lone game.Search primitive an ability content's
// single mode produces, with that mode's targets.
func searchInstruction(t *testing.T, content game.AbilityContent) (game.Search, []game.TargetSpec) {
	t.Helper()
	if content.IsModal() || len(content.Modes) != 1 {
		t.Fatalf("content = %#v, want one non-modal mode", content)
	}
	mode := content.Modes[0]
	for i := range mode.Sequence {
		if search, ok := mode.Sequence[i].Primitive.(game.Search); ok {
			return search, append(append([]game.TargetSpec(nil), content.SharedTargets...), mode.Targets...)
		}
	}
	t.Fatalf("mode sequence = %#v, want a game.Search primitive", mode.Sequence)
	return game.Search{}, nil
}

// TestLowerFertilidTargetPlayerSearch verifies Fertilid: an activated ability
// whose target player searches their own library for a basic land and puts it
// onto the battlefield tapped. The searcher is the chosen target player, so the
// found land enters under that player's control without a separate Controller
// reference.
func TestLowerFertilidTargetPlayerSearch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Fertilid",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental Beast",
		ManaCost:   "{2}{G}",
		OracleText: "This creature enters with two +1/+1 counters on it.\n{1}{G}, Remove a +1/+1 counter from this creature: Target player searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || len(faces[0].ActivatedAbilities) != 1 {
		t.Fatalf("faces = %#v, want one face with one activated ability", faces)
	}
	search, targets := searchInstruction(t, faces[0].ActivatedAbilities[0].Content)
	if search.Player != game.TargetPlayerReference(0) {
		t.Errorf("search Player = %#v, want TargetPlayerReference(0)", search.Player)
	}
	if search.Controller.Exists {
		t.Errorf("search Controller = %#v, want unset (the searcher controls the land)", search.Controller)
	}
	if len(targets) != 1 || targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %#v, want one player target", targets)
	}
	want := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		EntersTapped: true,
		Filter: game.Selection{
			RequiredTypes: []types.Card{types.Land},
			Supertypes:    []types.Super{types.Basic},
		},
	}
	if !reflect.DeepEqual(search.Spec, want) {
		t.Errorf("search Spec = %#v, want %#v", search.Spec, want)
	}
}

// TestLowerYavimayaDryadEntersUnderTargetControl verifies Yavimaya Dryad: the
// controller searches their own library for a Forest, but the found permanent
// enters under a chosen target player's control. The searcher stays the
// controller; Search.Controller binds the new controller to the target player.
func TestLowerYavimayaDryadEntersUnderTargetControl(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Yavimaya Dryad",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		ManaCost:   "{2}{G}",
		OracleText: "Forestwalk\nWhen this creature enters, you may search your library for a Forest card, put it onto the battlefield tapped under target player's control, then shuffle.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || len(faces[0].TriggeredAbilities) != 1 {
		t.Fatalf("faces = %#v, want one face with one triggered ability", faces)
	}
	search, targets := searchInstruction(t, faces[0].TriggeredAbilities[0].Content)
	if search.Player != game.ControllerReference() {
		t.Errorf("search Player = %#v, want ControllerReference (the controller searches)", search.Player)
	}
	if search.Controller != opt.Val(game.TargetPlayerReference(0)) {
		t.Errorf("search Controller = %#v, want TargetPlayerReference(0)", search.Controller)
	}
	if len(targets) != 1 || targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %#v, want one player target", targets)
	}
	if !search.Spec.EntersTapped || search.Spec.Destination != zone.Battlefield {
		t.Errorf("search Spec = %#v, want a tapped battlefield destination", search.Spec)
	}
}

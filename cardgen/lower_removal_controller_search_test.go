package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestLowerRemovalThenControllerSearch verifies the Path to Exile rider: a
// targeted removal spell whose affected permanent's controller may fetch a basic
// land lowers to the removal instruction followed by an Optional game.Search
// whose Player and OptionalActor both name the removal target's controller, so
// the affected player — not the spell's controller — decides whether to search.
func TestLowerRemovalThenControllerSearch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Path to Exile",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{W}",
		OracleText: "Exile target creature. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || !faces[0].SpellAbility.Exists {
		t.Fatalf("faces = %#v", faces)
	}
	content := faces[0].SpellAbility.Val
	if content.IsModal() || len(content.Modes) != 1 {
		t.Fatalf("content = %#v, want one non-modal mode", content)
	}
	mode := content.Modes[0]
	if len(content.SharedTargets)+len(mode.Targets) != 1 {
		t.Fatalf("targets = shared %#v mode %#v, want one", content.SharedTargets, mode.Targets)
	}
	seq := mode.Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}

	if seq[0].Optional || seq[0].OptionalActor.Exists {
		t.Errorf("removal instruction optionality = (%v, %v), want mandatory", seq[0].Optional, seq[0].OptionalActor.Exists)
	}
	exile, ok := seq[0].Primitive.(game.Exile)
	if !ok {
		t.Fatalf("first primitive = %#v, want game.Exile", seq[0].Primitive)
	}
	if exile.Object != game.TargetPermanentReference(0) {
		t.Errorf("exile object = %#v, want TargetPermanentReference(0)", exile.Object)
	}

	searcher := game.ObjectControllerReference(game.TargetPermanentReference(0))
	if !seq[1].Optional {
		t.Error("search instruction Optional = false, want true")
	}
	if seq[1].OptionalActor != opt.Val(searcher) {
		t.Errorf("search OptionalActor = %#v, want the target's controller", seq[1].OptionalActor)
	}
	search, ok := seq[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want game.Search", seq[1].Primitive)
	}
	if search.Player != searcher {
		t.Errorf("search Player = %#v, want the target's controller", search.Player)
	}
	want := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		CardType:     opt.Val(types.Land),
		Supertype:    opt.Val(types.Basic),
		EntersTapped: true,
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Errorf("spec = %+v, want %+v", search.Spec, want)
	}
}

// TestLowerRemovalThenControllerSearchDestroy verifies the same rider on a
// Destroy removal (Assassin's Trophy shape), where the fetched land enters
// untapped, exercising the EntersTapped=false branch of the shared spec builder.
func TestLowerRemovalThenControllerSearchDestroy(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Assassin's Trophy",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{B}{G}",
		OracleText: "Destroy target permanent an opponent controls. Its controller may search their library for a basic land card, put it onto the battlefield, then shuffle.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	if _, ok := seq[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %#v, want game.Destroy", seq[0].Primitive)
	}
	searcher := game.ObjectControllerReference(game.TargetPermanentReference(0))
	if !seq[1].Optional || seq[1].OptionalActor != opt.Val(searcher) {
		t.Errorf("search optionality = (%v, %#v), want Optional with the target's controller", seq[1].Optional, seq[1].OptionalActor)
	}
	search, ok := seq[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("second primitive = %#v, want game.Search", seq[1].Primitive)
	}
	if search.Spec.EntersTapped {
		t.Error("spec EntersTapped = true, want false for an untapped fetch")
	}
}

// TestLowerRemovalThenControllerSearchFailsClosed confirms the lowerer rejects
// near-miss shapes rather than producing a silently-wrong sequence. Each case
// keeps the "Its controller may ..." rider but breaks one structural requirement.
func TestLowerRemovalThenControllerSearchFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
	}{
		{
			// The rider is not a library search, so there is no basic-land fetch
			// to model; the optional draw must stay unsupported.
			name:   "non-search rider",
			oracle: "Exile target creature. Its controller may draw a card.",
		},
		{
			// The leading effect is not removal of the search subject; tapping
			// leaves the target on the battlefield and is outside the shape.
			name:   "non-removal leading effect",
			oracle: "Tap target creature. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
		{
			// The searcher is the spell's controller, not the target's controller,
			// so the rider is an ordinary self-tutor, not this shape.
			name:   "controller searches own library",
			oracle: "Exile target creature. You may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		},
		{
			// A color filter the runtime SearchSpec cannot express must keep the
			// whole body unsupported.
			name:   "unsupported search filter",
			oracle: "Exile target creature. Its controller may search their library for a green creature card, put it onto the battlefield, then shuffle.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Spell",
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{W}",
				OracleText: tc.oracle,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected an unsupported diagnostic, got faces = %#v", faces)
			}
		})
	}
}

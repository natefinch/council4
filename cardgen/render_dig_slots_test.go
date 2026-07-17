package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestRenderDigSlotsEmitsRoutingAndImpulseGrant proves that a Dig carrying
// ordered destination slots renders each slot's count, destination, bottom
// placement, and the exile slot's impulse play grant, and that it requests the
// zone and opt imports. This is the Expressive Iteration shape: look at three,
// one to hand (the primary Take), one to the bottom of the library, one exiled
// and playable this turn.
func TestRenderDigSlotsEmitsRoutingAndImpulseGrant(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Slots: []game.DigSlot{
			{Count: game.Fixed(1), Destination: zone.Library, Bottom: true},
			{Count: game.Fixed(1), Destination: zone.Exile, Play: opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn})},
		},
	})
	if err != nil {
		t.Fatalf("renderPrimitive() error = %v", err)
	}
	for _, want := range []string{
		"game.Dig{",
		"Slots: []game.DigSlot{",
		"Count: game.Fixed(1),",
		"Destination: zone.Library,",
		"Bottom: true,",
		"Destination: zone.Exile,",
		"Play: opt.Val(game.ImpulsePlayGrant{",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered dig slots missing %q:\n%s", want, rendered)
		}
	}
	if _, ok := ctx.imports[importZone]; !ok {
		t.Fatal("dig slots did not request the zone import")
	}
	if _, ok := ctx.imports[importOpt]; !ok {
		t.Fatal("dig slots did not request the opt import")
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered output is not valid Go: %v\n%s", err, rendered)
	}
}

// TestRenderDigSlotsCastGrantEmitsCast proves the exile slot's cast-only grant
// renders the Cast flag so a generated card can distinguish "you may cast" from
// "you may play".
func TestRenderDigSlotsCastGrantEmitsCast(t *testing.T) {
	t.Parallel()
	ctx := newRenderCtx()
	rendered, err := (Renderer{}).renderPrimitive(ctx, game.Dig{
		Player: game.ControllerReference(),
		Look:   game.Fixed(1),
		Take:   game.Fixed(0),
		Slots: []game.DigSlot{
			{Count: game.Fixed(1), Destination: zone.Exile, Play: opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn, Cast: true})},
		},
	})
	if err != nil {
		t.Fatalf("renderPrimitive() error = %v", err)
	}
	if !strings.Contains(rendered, "Cast: true,") {
		t.Fatalf("rendered cast grant missing Cast flag:\n%s", rendered)
	}
	src := "package p\nvar _ = " + rendered
	if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
		t.Fatalf("rendered output is not valid Go: %v\n%s", err, rendered)
	}
}

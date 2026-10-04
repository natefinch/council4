package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRenderResultObjectSelectionGate(t *testing.T) {
	instruction := game.Instruction{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		ResultGate: opt.Val(game.InstructionResultGate{
			Key: "exile", Succeeded: game.TriTrue,
			ObjectSelection: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Pirate}}),
			Negate:          true,
			CardOnly:        true,
		}),
	}
	for _, test := range []struct {
		name, subtype string
		render        func(Renderer, *renderCtx, game.Instruction) (string, error)
	}{
		{"ordered instruction", `types.Sub("Pirate")`, func(r Renderer, ctx *renderCtx, instruction game.Instruction) (string, error) {
			return r.renderInstruction(ctx, &instruction)
		}},
		{"generated struct", "types.Pirate", func(r Renderer, ctx *renderCtx, instruction game.Instruction) (string, error) {
			return r.renderGameInstruction(ctx, instruction)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newRenderCtx()
			source, err := test.render(Renderer{}, ctx, instruction)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"ObjectSelection: opt.Val(game.Selection{", "SubtypesAny: []types.Sub{" + test.subtype + "}", "Succeeded: game.TriTrue", "Negate: true", "CardOnly: true"} {
				if !strings.Contains(source, want) {
					t.Fatalf("rendered gate lost %q:\n%s", want, source)
				}
			}
			if _, ok := ctx.imports[importTypes]; !ok {
				t.Fatal("result selection did not request its subtype import")
			}
		})
	}
}

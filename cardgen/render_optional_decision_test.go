package cardgen

import (
	"go/format"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRenderOptionalDecisionEnvelope(t *testing.T) {
	instruction := game.Instruction{
		Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		Optional:  true, PublishOptionalDecision: "inner", OptionalDecisionGate: "outer",
		PublishResult: "actual",
	}
	for _, tt := range []struct {
		name   string
		render func(Renderer, *renderCtx, game.Instruction) (string, error)
	}{
		{"ordered", func(r Renderer, ctx *renderCtx, instruction game.Instruction) (string, error) {
			return r.renderInstruction(ctx, &instruction)
		}},
		{"generated", func(r Renderer, ctx *renderCtx, instruction game.Instruction) (string, error) {
			return r.renderGameInstruction(ctx, instruction)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source, err := tt.render(Renderer{}, newRenderCtx(), instruction)
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"PublishOptionalDecision:", "OptionalDecisionGate:", "PublishResult:", `"inner"`, `"outer"`, `"actual"`} {
				if !strings.Contains(source, field) {
					t.Fatalf("render lost %q:\n%s", field, source)
				}
			}
			if tt.name == "ordered" {
				source = "game.Instruction" + source
			}
			if _, err := format.Source([]byte("package probe\nvar _ = " + source)); err != nil {
				t.Fatalf("invalid generated expression: %v", err)
			}
		})
	}
}

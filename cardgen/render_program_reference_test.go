package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRendererPreservesProgramReferenceData(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{CardFace: game.CardFace{Name: "Program", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: game.ProgramReference{ID: "test/body", Version: "1"}}.Ability())}}
	source, err := (Renderer{}).RenderCardSource(&ScryfallCard{Name: "Program", Layout: "normal"},
		[]*game.CardDef{def}, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	source = strings.Join(strings.Fields(source), " ")
	for _, fragment := range []string{"Program:", "game.ProgramReference", `ID: "test/body"`, `Version: "1"`} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("binding %q was lost in source:\n%s", fragment, source)
		}
	}
}

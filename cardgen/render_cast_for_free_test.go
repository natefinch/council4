package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestRenderCastForFreeImportsZone(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Free Cast",
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.CastForFree{
					Player: game.ControllerReference(),
					Zone:   zone.Hand,
				},
			}},
		}.Ability()),
	}}
	source, err := (Renderer{}).RenderCardSource(
		&ScryfallCard{Name: "Free Cast", Layout: "normal"},
		[]*game.CardDef{def},
		nil,
		"testcards",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, `"github.com/natefinch/council4/mtg/game/zone"`) {
		t.Fatalf("rendered source lacks zone import:\n%s", source)
	}
	if !strings.Contains(source, "zone.Hand") {
		t.Fatalf("rendered source lacks hand zone:\n%s", source)
	}
}

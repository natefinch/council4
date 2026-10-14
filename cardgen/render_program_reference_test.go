package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

func TestRendererPreservesReplacementProgramAndRevealPartition(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{CardFace: game.CardFace{Name: "Replacement program", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{{
			Text: "replace",
			Replacement: game.ReplacementEffect{
				Description: "replace", MatchEvent: game.EventCardDrawn,
				ControllerFilter: game.TriggerControllerYou, Duration: game.DurationPermanent,
				Program:  game.ProgramReference{ID: "test/replacement", Version: "1"},
				Optional: true, SelfReplacement: true,
			},
		}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.RevealUntil{
			Player: game.ControllerReference(), Destination: zone.Hand,
			MatchToDestinationRestChosenBottom: true,
		}}}}.Ability()),
	}}
	source, err := (Renderer{}).RenderCardSource(&ScryfallCard{Name: def.Name, Layout: "normal"}, []*game.CardDef{def}, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	source = strings.Join(strings.Fields(source), " ")
	for _, fragment := range []string{
		"Program:", "game.ProgramReference", `ID: "test/replacement"`, `Version: "1"`,
		"Optional: true", "SelfReplacement: true", "MatchToDestinationRestChosenBottom: true",
		"MatchEvent: game.EventCardDrawn", "ControllerFilter: game.TriggerControllerYou",
	} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("replacement data %q was lost in source:\n%s", fragment, source)
		}
	}
}

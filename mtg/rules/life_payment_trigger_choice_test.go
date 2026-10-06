package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledSingletonLifePaymentTriggerChoice(t *testing.T) {
	for _, tc := range []struct {
		name        string
		text        string
		life        int
		accept      bool
		forbidden   bool
		wantLife    int
		wantDraws   int
		wantPrompts []string
	}{
		{name: "accepted", life: 20, accept: true, wantLife: 18, wantPrompts: []string{"Pay resolution cost?"}},
		{name: "declined", life: 20, wantLife: 20, wantPrompts: []string{"Pay resolution cost?"}},
		{name: "insufficient", life: 1, accept: true, wantLife: 1},
		{name: "forbidden", life: 20, accept: true, forbidden: true, wantLife: 20},
		{name: "ordinary optional accept", text: "At the beginning of your upkeep, you may draw a card.", life: 20, accept: true, wantLife: 20, wantDraws: 1, wantPrompts: []string{"Apply optional triggered ability?"}},
		{name: "ordinary optional decline", text: "At the beginning of your upkeep, you may draw a card.", life: 20, wantLife: 20, wantPrompts: []string{"Apply optional triggered ability?"}},
		{name: "independent rider accept", text: "At the beginning of your upkeep, you may pay 2 life. You may draw a card.", life: 20, accept: true, wantLife: 18, wantDraws: 1, wantPrompts: []string{"Pay resolution cost?", "Apply optional effect?"}},
		{name: "independent rider decline", text: "At the beginning of your upkeep, you may pay 2 life. You may draw a card.", life: 20, wantLife: 20, wantPrompts: []string{"Pay resolution cost?", "Apply optional effect?"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			text := tc.text
			if text == "" {
				text = "At the beginning of your upkeep, you may pay 2 life."
			}
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Singleton Trigger Choice Runtime", Layout: "normal", TypeLine: "Enchantment", OracleText: text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = tc.life
			addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 2, 3))
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
					ID: g.IDGen.Next(), Kind: game.RuleEffectLifeTotalCantChange, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
				})
			}
			controller := &libraryPaymentAgent{accept: tc.accept}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			NewEngine(nil).resolveTriggeredAbilityBodyWithChoices(g, obj, def, &def.TriggeredAbilities[0],
				[game.NumPlayers]PlayerAgent{game.Player1: controller}, &TurnLog{})
			if g.Players[game.Player1].Life != tc.wantLife || g.Players[game.Player1].Hand.Size() != tc.wantDraws {
				t.Errorf("life=%d draws=%d, want %d/%d", g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size(), tc.wantLife, tc.wantDraws)
			}
			if len(controller.prompts) != len(tc.wantPrompts) {
				t.Errorf("prompts=%#v, want %v", controller.prompts, tc.wantPrompts)
			} else {
				for i, request := range controller.prompts {
					if request.Prompt != tc.wantPrompts[i] || request.Player != game.Player1 {
						t.Errorf("prompt[%d]=%#v, want controller prompt %q", i, request, tc.wantPrompts[i])
					}
				}
			}
			lifeEvents, drawEvents := 0, 0
			for _, event := range g.Events {
				if event.Kind == game.EventLifeLost {
					lifeEvents++
					if event.Player != game.Player1 || event.Amount != 2 {
						t.Errorf("wrong payment life event: %#v", event)
					}
				}
				if event.Kind == game.EventCardDrawn {
					drawEvents++
				}
			}
			wantLifeEvents := 0
			if tc.life-tc.wantLife == 2 {
				wantLifeEvents = 1
			}
			if lifeEvents != wantLifeEvents || drawEvents != tc.wantDraws {
				t.Errorf("life/draw events=%d/%d, want %d/%d", lifeEvents, drawEvents, wantLifeEvents, tc.wantDraws)
			}
		})
	}
}

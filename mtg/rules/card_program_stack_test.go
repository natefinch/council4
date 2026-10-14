package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
)

func TestProgramActivatedTriggeredAndChapterResolveOnStack(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name string
		kind game.StackObjectKind
		saga bool
	}{
		{name: "activated", kind: game.StackActivatedAbility},
		{name: "triggered", kind: game.StackTriggeredAbility},
		{name: "chapter", kind: game.StackTriggeredAbility, saga: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			ref := game.ProgramReference{ID: "test/" + scenario.name, Version: "1"}
			catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
				Reference: ref,
				Resolve: func(ctx cardprogram.Context) error {
					info := ctx.View().Source()
					if info.CardID != source.CardInstanceID || info.ObjectID != source.ObjectID || info.Controller != game.Player1 {
						t.Fatal("program source identity/controller changed")
					}
					_, err := ctx.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()})
					return err
				},
			}})
			if err != nil {
				t.Fatal(err)
			}
			content := game.Mode{Program: ref}.Ability()
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: scenario.kind, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Controller: game.Player1, SagaChapter: scenario.saga,
			}
			if scenario.kind == game.StackActivatedAbility {
				obj.InlineActivated = &game.ActivatedAbility{Content: content}
			} else {
				obj.InlineTrigger = &game.TriggeredAbility{Content: content}
			}
			g.Stack.Push(obj)
			log := TurnLog{}
			NewEngineWithPrograms(nil, catalog).resolveTopOfStack(g, &log)
			if !g.Stack.IsEmpty() || !g.Players[game.Player1].Hand.Contains(drawn) || len(log.Draws) != 1 {
				t.Fatalf("stack body did not resolve normally: draws=%d", len(log.Draws))
			}
		})
	}
}

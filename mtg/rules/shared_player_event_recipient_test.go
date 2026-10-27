package rules

import (
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledSharedPlayerRecipientCopiedEventAfterSourceLeaves(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text          string
		life, library int
	}{
		{"Whenever a creature dies, that creature's controller loses 2 life.", 36, 5},
		{"Whenever a creature dies, that creature's controller mills two cards.", 40, 1},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			def := legacyCard(t, carddata.ScryfallCard{
				Name: "Copied Event Recipient", Layout: "normal", TypeLine: "Enchantment", OracleText: tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, def)
			subject := addCombatCreaturePermanent(g, game.Player2)
			subject.Controller = game.Player3
			for player := range g.Players {
				for range 5 {
					addCardToLibrary(g, game.PlayerID(player), evidenceCard("Event library", 1))
				}
			}
			resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, SourceID: subject.ObjectID},
				game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Graveyard}, nil)
			var event game.Event
			for _, emitted := range g.Events {
				if emitted.Kind == game.EventPermanentDied && emitted.PermanentID == subject.ObjectID {
					event = emitted
				}
			}
			if event.Kind != game.EventPermanentDied {
				t.Fatal("actual subject death did not emit its event")
			}
			original := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
				Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &def.TriggeredAbilities[0], HasTriggerEvent: true, TriggerEvent: event}
			g.Stack.Push(original)
			resolveInstruction(engine, g, &game.StackObject{Controller: game.Player4,
				Targets: []game.Target{game.StackObjectTarget(original.ID)}},
				game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, nil)
			copied, ok := g.Stack.Peek()
			if !ok || !copied.Copy || copied.Controller != game.Player4 ||
				!copied.HasTriggerEvent || copied.TriggerEvent.PermanentID != subject.ObjectID {
				t.Fatal("real stack copy lost the original event subject or distinct resolving controller")
			}
			resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID},
				game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile}, nil)
			engine.resolveTopOfStack(g, &TurnLog{})
			engine.resolveTopOfStack(g, &TurnLog{})
			for player, state := range g.Players {
				life, library := 40, 5
				if game.PlayerID(player) == game.Player3 {
					life, library = tt.life, tt.library
				}
				if state.Life != life || state.Library.Size() != library {
					t.Fatalf("player %d life/library=%d/%d, want %d/%d after original+copy",
						player, state.Life, state.Library.Size(), life, library)
				}
			}
		})
	}
}

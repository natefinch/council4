package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestObservedLibraryPlacementPreservesIncarnation(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		decoy := addCardToLibrary(g, game.Player3, vanillaCreature("Decoy", 1, 1))
		cardID := addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
		obj := &game.StackObject{
			ID: g.IDGen.Next(), Controller: game.Player1,
			Targets: []game.Target{game.PlayerTarget(game.Player3)},
		}
		resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		resolver.resolveInstruction(&game.Instruction{Primitive: game.LookAtLibraryTop{
			Player: game.TargetPlayerReference(0), PublishLinked: "observed",
		}})
		card := g.CardInstances[cardID]
		version, events := card.ZoneVersion, len(g.Events)
		resolver.resolveInstruction(&game.Instruction{
			Primitive: game.MoveCard{
				Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
				FromZone: zone.Library, Destination: zone.Library, DestinationBottom: bottom,
			},
			PublishResult: "placed",
		})
		top, _ := g.Players[game.Player3].Library.Top()
		want := cardID
		if bottom {
			want = decoy
		}
		if top != want || card.ZoneVersion != version || len(g.Events) != events {
			t.Fatalf("placement changed identity or emitted a zone change: top=%d want=%d version=%d events=%d",
				top, want, card.ZoneVersion, len(g.Events)-events)
		}
		if actual, _, ok := resolveCardReference(g, obj, game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"}); !ok || actual != cardID {
			t.Fatal("same-library placement substituted the new top or lost the actual card")
		}
	}
}

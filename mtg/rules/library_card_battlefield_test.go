package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardConsumerPublishesActualBattlefieldIncarnation(t *testing.T) {
	for _, reveal := range []bool{false, true} {
		for _, tapped := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := addCardToLibrary(g, game.Player1, vanillaCreature("Decoy", 1, 1))
			observed := addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
			}
			var producer game.Primitive = game.LookAtLibraryTop{
				Player: game.TargetPlayerReference(1), PublishLinked: "observed",
			}
			if reveal {
				producer = game.Reveal{Player: game.TargetPlayerReference(1), Amount: game.Fixed(1), PublishLinked: "observed"}
			}
			resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: producer})
			version := g.CardInstances[observed].ZoneVersion
			newTop := addCardToLibrary(g, game.Player3, vanillaCreature("Later Top", 4, 5))
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.PutOnBattlefield{
					Source: game.LinkedBattlefieldSource("observed"), Recipient: opt.Val(game.ControllerReference()),
					EntryTapped: tapped, PublishLinked: "entered",
				},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				})}),
			})
			refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "entered"))
			if len(refs) != 1 || refs[0].CardID != observed || refs[0].ObjectID == 0 {
				t.Fatalf("consumer did not publish actual new permanent: %#v", refs)
			}
			permanent, ok := permanentByObjectID(g, refs[0].ObjectID)
			if !ok || permanent.Owner != game.Player3 || permanent.Controller != game.Player1 || permanent.Tapped != tapped {
				t.Fatalf("consumer lost owner/controller/entry state: %#v", permanent)
			}
			if g.CardInstances[observed].ZoneVersion == version ||
				g.Players[game.Player3].Library.Contains(observed) ||
				!g.Players[game.Player3].Library.Contains(newTop) || !g.Players[game.Player1].Library.Contains(decoy) {
				t.Fatal("consumer did not move the exact observed incarnation")
			}
			if _, ok := resolveObjectReference(g, obj, game.LinkedObjectReference("observed")); ok {
				t.Fatal("old library observation silently became the new battlefield object")
			}
			if _, ok := resolveObjectReference(g, obj, game.LinkedObjectReference("entered")); !ok {
				t.Fatal("actual new battlefield identity is unavailable")
			}
		}
	}
}

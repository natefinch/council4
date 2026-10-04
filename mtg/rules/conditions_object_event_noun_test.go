package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestObjectMatchesEventLandNotTargetCreatureLand(t *testing.T) {
	t.Parallel()
	for _, eventIsland := range []bool{false, true} {
		t.Run(fmt.Sprintf("event-island=%t", eventIsland), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			eventSubtype, targetSubtype := types.Forest, types.Island
			if eventIsland {
				eventSubtype, targetSubtype = types.Island, types.Forest
			}
			eventLand := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Event Land", Types: []types.Card{types.Land}, Subtypes: []types.Sub{eventSubtype},
			}})
			target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Target Creature Land", Types: []types.Card{types.Creature, types.Land},
				Subtypes: []types.Sub{targetSubtype},
			}})
			obj := &game.StackObject{
				Controller: game.Player1, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: eventLand.ObjectID},
				Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
			}
			condition := opt.Val(game.Condition{
				Object:        opt.Val(game.EventPermanentReference()),
				ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Island}}),
			})
			ctx := conditionContext{controller: game.Player1, obj: obj, event: &obj.TriggerEvent}
			if got := conditionSatisfied(g, ctx, condition); got != eventIsland {
				t.Fatalf("event-land gate = %v, want %v", got, eventIsland)
			}
			condition.Val.Object = opt.Val(game.TargetPermanentReference(0))
			if got := conditionSatisfied(g, ctx, condition); got == eventIsland {
				t.Fatal("cross-case did not distinguish event from creature-land target")
			}
		})
	}
}

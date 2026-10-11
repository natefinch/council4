package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledExactXUntapUsesBoundedSlots(t *testing.T) {
	for _, x := range []int{0, 1, 3} {
		t.Run(strconv.Itoa(x), func(t *testing.T) {
			def := compileTargetCountSpell(t, "Untap X target lands.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var targets []game.Target
			var chosen []*game.Permanent
			for range x {
				land := addLandPermanent(g, game.Player2, "Chosen Land")
				land.Tapped = true
				chosen = append(chosen, land)
				targets = append(targets, game.PermanentTarget(land.ObjectID))
			}
			untargeted := addLandPermanent(g, game.Player2, "Untargeted Land")
			untargeted.Tapped = true
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
				XValue: x, Targets: targets, TargetCounts: []int{x},
			})
			if x == 3 {
				if _, destroyed := destroyPermanentInBatch(g, chosen[1].ObjectID, 0, false); !destroyed {
					t.Fatal("could not make the middle selected target illegal")
				}
			}
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			for _, land := range chosen {
				if _, exists := permanentByObjectID(g, land.ObjectID); exists && land.Tapped {
					t.Fatal("a legal selected land was not untapped")
				}
			}
			if !untargeted.Tapped {
				t.Fatal("an illegal or unfilled target slot untapped another land")
			}
		})
	}
}

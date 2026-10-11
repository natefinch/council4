package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledHalfPlayerAmountUsesCurrentSameRecipient(t *testing.T) {
	for _, library := range []bool{false, true} {
		for _, roundUp := range []bool{false, true} {
			for _, inherited := range []bool{false, true} {
				rounding := "down"
				if roundUp {
					rounding = "up"
				}
				body := "Target opponent loses half their life, rounded " + rounding + "."
				if library {
					body = "Target opponent mills half their library, rounded " + rounding + "."
				}
				if inherited {
					body = "Target opponent loses 2 life. Scry 1. That player loses half their life, rounded " + rounding + "."
					if library {
						body = "Target opponent loses 2 life. Scry 1. That player mills half their library, rounded " + rounding + "."
					}
				}
				def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Half Current Player", Layout: "normal",
					TypeLine: "Instant", OracleText: body})
				for _, initial := range []int{7, 8} {
					t.Run(fmt.Sprintf("library%v-up%v-inherited%v-%d", library, roundUp, inherited, initial), func(t *testing.T) {
						g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
						g.Players[game.Player3].Life = initial
						for player := range g.Players {
							for range initial {
								addCardToLibrary(g, game.PlayerID(player), evidenceCard("Half library", 1))
							}
						}
						obj := &game.StackObject{Kind: game.StackSpell, Controller: game.Player1,
							Targets: []game.Target{game.PlayerTarget(game.Player3)}, TargetCounts: []int{1}}
						NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
						current := initial
						if inherited && !library {
							current -= 2
						}
						half := current / 2
						if roundUp {
							half = (current + 1) / 2
						}
						wantLife, wantLibrary := initial-half, initial
						if inherited {
							wantLife -= 2
						}
						if library {
							wantLife, wantLibrary = initial, initial-half
							if inherited {
								wantLife -= 2
							}
						}
						if g.Players[game.Player3].Life != wantLife || g.Players[game.Player3].Library.Size() != wantLibrary {
							t.Fatalf("recipient life/library=%d/%d want%d/%d", g.Players[game.Player3].Life,
								g.Players[game.Player3].Library.Size(), wantLife, wantLibrary)
						}
						for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player4} {
							if g.Players[player].Life != 40 || g.Players[player].Library.Size() != initial {
								t.Fatal("half quantity read or changed another controller/owner/player")
							}
						}
					})
				}
			}
		}
	}
}

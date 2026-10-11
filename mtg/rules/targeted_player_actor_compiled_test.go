package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledCourtActorChosenSetAndCurrentMonarch(t *testing.T) {
	t.Parallel()
	const action = "any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead."
	for _, prefix := range []string{"", "Counter target spell. "} {
		for _, count := range []int{0, 1, 2} {
			for _, monarch := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/monarch=%v", prefix, count, monarch), func(t *testing.T) {
					t.Parallel()
					card := &cardgen.ScryfallCard{Name: "Court of Cunning", Layout: "normal", TypeLine: "Enchantment",
						ManaCost: "{1}{U}{U}", OracleText: "When this enchantment enters, you become the monarch.\nAt the beginning of your upkeep, " + prefix + action + " You gain 1 life."}
					defs, diagnostics, err := cardgen.CompileCardDefs(card)
					if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
						t.Fatalf("compile Court actor: %v %+v", err, diagnostics)
					}
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player1, defs[0])
					for player := range g.Players {
						for range 12 {
							addCardToLibrary(g, game.PlayerID(player), &game.CardDef{CardFace: game.CardFace{Name: "Library card"}})
						}
					}
					var targets []game.Target
					var counts []int
					if prefix == "Counter target spell. " {
						stackTarget := addStackSpell(g, game.Player4, "Nonplayer prefix", []types.Card{types.Sorcery})
						targets = append(targets, game.StackObjectTarget(stackTarget.ID))
						counts = append(counts, 1)
					}
					for i := range count {
						targets = append(targets, game.PlayerTarget(game.PlayerID(i+1)))
					}
					counts = append(counts, count)
					ability := &defs[0].TriggeredAbilities[1]
					obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
						Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
						InlineTrigger: ability, Targets: targets, TargetCounts: counts}
					g.Stack.Push(obj)
					g.Players[game.Player1].IsMonarch = monarch
					NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
					amount := 2
					if monarch {
						amount = 10
					}
					for player := range g.Players {
						want := 12
						if player > 0 && player <= count {
							want -= amount
						}
						if g.Players[player].Library.Size() != want {
							t.Fatalf("player%d library=%d want%d: chosen actor set borrowed a neighbor", player, g.Players[player].Library.Size(), want)
						}
					}
					if g.Players[game.Player1].Life != 41 {
						t.Fatal("legal zero-count/group action suppressed its independent rider")
					}
				})
			}
		}
	}
}

func TestCompiledCourtActorPartialAndFullTargetFizzle(t *testing.T) {
	t.Parallel()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Court of Cunning", Layout: "normal", TypeLine: "Enchantment", ManaCost: "{1}{U}{U}",
		OracleText: "When this enchantment enters, you become the monarch.\nAt the beginning of your upkeep, any number of target players each mill two cards. If you're the monarch, each of those players mills ten cards instead. You gain 1 life.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile Court actor: %v %+v", err, diagnostics)
	}
	for _, allGone := range []bool{false, true} {
		t.Run(fmt.Sprintf("allgone=%v", allGone), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, defs[0])
			for player := range g.Players {
				for range 12 {
					addCardToLibrary(g, game.PlayerID(player), &game.CardDef{CardFace: game.CardFace{Name: "Library card"}})
				}
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
				Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				InlineTrigger: &defs[0].TriggeredAbilities[1],
				Targets:       []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
				TargetCounts:  []int{2}}
			g.Stack.Push(obj)
			g.Players[game.Player2].Eliminated = true
			g.Players[game.Player3].Eliminated = allGone
			g.Players[game.Player1].IsMonarch = true
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			wantLibrary, wantLife := 2, 41
			if allGone {
				wantLibrary, wantLife = 12, 40
			}
			if g.Players[game.Player3].Library.Size() != wantLibrary ||
				g.Players[game.Player2].Library.Size() != 12 || g.Players[game.Player4].Library.Size() != 12 ||
				g.Players[game.Player1].Life != wantLife {
				t.Fatal("partially stale actor rebound to a neighbor, or full fizzle applied an independent rider")
			}
		})
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestResultThisWaySpecializedSacrificeFlows(t *testing.T) {
	for _, test := range []struct {
		name, tail string
		hand, life int
	}{
		{"scaled rewards", "you gain X life and draw X cards, where X is that creature's power.", 2, 42},
		{"linked return", "return that card to the battlefield under its owner's control with three +1/+1 counters on it.", 0, 40},
		{"search", "search your library for a basic land card, reveal it, put it into your hand, then shuffle.", 1, 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
				Name: "Sacrifice Filter Probe", Layout: "normal", TypeLine: "Artifact",
				OracleText: "When this artifact enters, you may sacrifice a creature. If a Saproling was sacrificed this way, " + test.tail,
			})
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("compile = %v, %v", diagnostics, err)
			}
			content := defs[0].TriggeredAbilities[0].Content
			for i := 1; i < len(content.Modes[0].Sequence); i++ {
				gate := content.Modes[0].Sequence[i].ResultGate
				if !gate.Exists || !gate.Val.ObjectSelection.Exists {
					t.Fatalf("specialized instruction %d dropped its typed result selection", i)
				}
			}
			for _, matching := range []bool{false, true} {
				t.Run(map[bool]string{false: "nonmatching", true: "matching"}[matching], func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player1, defs[0])
					subtype := types.Sub("Elf")
					if matching {
						subtype = "Saproling"
					}
					candidate := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
						Types: []types.Card{types.Creature}, Subtypes: []types.Sub{subtype},
						Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
					}})
					for range 3 {
						addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
							Types: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic},
						}})
					}
					obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
						SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					wantHand, wantLife := 0, 40
					if matching {
						wantHand, wantLife = test.hand, test.life
					}
					if g.Players[game.Player1].Hand.Size() != wantHand || g.Players[game.Player1].Life != wantLife {
						t.Fatalf("hand/life = %d/%d, want %d/%d", g.Players[game.Player1].Hand.Size(), g.Players[game.Player1].Life, wantHand, wantLife)
					}
					wantBattlefield := 1
					if matching && test.name == "linked return" {
						wantBattlefield = 2
					}
					if len(g.Battlefield) != wantBattlefield {
						t.Fatalf("battlefield = %d, want %d", len(g.Battlefield), wantBattlefield)
					}
					if matching && test.name == "linked return" {
						returned := g.Battlefield[1]
						if returned.CardInstanceID != candidate.CardInstanceID || returned.ObjectID == candidate.ObjectID ||
							returned.Counters.Get(counter.PlusOnePlusOne) != 3 {
							t.Fatal("matching linked return lost its original card or entry counters")
						}
					}
				})
			}
		})
	}
}

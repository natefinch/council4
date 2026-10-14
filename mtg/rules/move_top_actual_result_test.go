package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestMoveTopOfLibraryPublishesActualMovement(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		cards, requested                    int
		redirect, group, unchanged, missing bool
	}{
		{name: "empty", requested: 3},
		{name: "partial", cards: 1, requested: 3},
		{name: "redirected", cards: 1, requested: 1, redirect: true},
		{name: "group empty", requested: 2, group: true},
		{name: "group partial", cards: 1, requested: 3, group: true},
		{name: "group redirected", cards: 1, requested: 1, group: true, redirect: true},
		{name: "zero", cards: 1},
		{name: "replacement retains library zone", cards: 1, requested: 1, unchanged: true},
		{name: "missing target player", cards: 1, requested: 1, missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tc.redirect {
				graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false)
			}
			for range tc.cards {
				def := vanillaCreature("Moved", 1, 1)
				if tc.unchanged {
					def = selfLibraryReplacementCardDef()
				}
				addCardToLibrary(g, game.Player2, def)
			}
			primitive := game.MoveTopOfLibrary{Player: game.TargetPlayerReference(0), Amount: game.Fixed(tc.requested), Destination: zone.Graveyard}
			if tc.group {
				primitive.Player = game.PlayerReference{}
				primitive.PlayerGroup = game.OpponentsReference()
			}
			obj := &game.StackObject{Controller: game.Player1}
			if !tc.missing {
				obj.Targets = []game.Target{game.PlayerTarget(game.Player2)}
			}
			engine := NewEngine(nil)
			instruction := &game.Instruction{Primitive: primitive, PublishResult: "move"}
			engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			result := obj.ResolutionResults["move"]
			want := tc.cards > 0 && tc.requested > 0 && !tc.missing && !tc.unchanged
			if result.Succeeded != want {
				t.Fatalf("Succeeded=%t want actual movement=%t", result.Succeeded, want)
			}
			wantMembers := 0
			if want {
				wantMembers = min(tc.cards, tc.requested)
			}
			if result.Amount != wantMembers {
				t.Fatalf("actual mill Amount=%d want=%d", result.Amount, wantMembers)
			}
			snapshots, available := obj.ResolutionResultObjects["move"]
			if !available {
				t.Fatal("complete acted mill batch lost known-empty or public member availability")
			}
			if len(snapshots) != wantMembers {
				t.Fatalf("public milled members=%d want=%d", len(snapshots), wantMembers)
			}
			if tc.redirect {
				if !want || g.Players[game.Player2].Exile.Size() != 1 || len(snapshots) != 1 ||
					!resultObjectsMatchSelection(g, obj, snapshots, game.Selection{RequiredTypes: []types.Card{types.Creature}}) {
					t.Fatal("successful mill omitted the creature moved to public exile")
				}
			}
			if want {
				engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if obj.ResolutionResults["move"].Succeeded || len(obj.ResolutionResultObjects["move"]) != 0 {
					t.Fatal("empty repeat reused an earlier movement receipt")
				}
			}
		})
	}
}

func TestMoveTopOfLibraryExileAndLegacyMovementReceipts(t *testing.T) {
	for _, destination := range []zone.Type{zone.Graveyard, zone.Exile} {
		for _, commander := range []bool{false, true} {
			t.Run(destination.String()+map[bool]string{false: "/ordinary", true: "/commander"}[commander], func(t *testing.T) {
				newGame := func() *game.Game {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{
						game.Player2: {Commander: commanderDef("Moved Commander", color.Green)},
					})
					if commander {
						cardID := g.Players[game.Player2].CommanderInstanceID
						g.Players[game.Player2].CommandZone.Remove(cardID)
						g.Players[game.Player2].Library.Add(cardID)
					} else {
						addCardToLibrary(g, game.Player2, vanillaCreature("Moved", 1, 1))
					}
					return g
				}
				g := newGame()
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PlayerTarget(game.Player2)}}
				engine := NewEngine(nil)
				instruction := &game.Instruction{
					Primitive:     game.MoveTopOfLibrary{Player: game.TargetPlayerReference(0), Amount: game.Fixed(3), Destination: destination},
					PublishResult: "move",
				}
				engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				result := obj.ResolutionResults["move"]
				wantMembers := 1
				if commander {
					if cards, _ := destinationZone(g, game.Player2, destination); !cards.Contains(g.Players[game.Player2].CommanderInstanceID) {
						t.Fatal("commander did not arrive before its SBA")
					}
					engine.applyStateBasedActions(g)
					if g.Players[game.Player2].CommandZone.Size() != 1 {
						t.Fatal("commander did not reach replacement destination")
					}
				}
				wantAmount := 3
				if destination == zone.Graveyard {
					wantAmount = 1
				}
				if !result.Succeeded || result.Amount != wantAmount || len(obj.ResolutionResultObjects["move"]) != wantMembers {
					t.Fatal("actual mill receipt/public members or independent exile receipt/destination members changed")
				}
				engine.resolveInstructionWithChoices(g, obj, instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if obj.ResolutionResults["move"].Succeeded {
					t.Fatal("empty exile/mill repeat succeeded")
				}
				legacy := newGame()
				cards := millCards(legacy, game.Player2, 3)
				if destination == zone.Exile {
					legacy = newGame()
					cards = exileTopOfLibraryCards(legacy, game.Player2, 3, opt.V[counter.Kind]{}, game.Player1, false)
				}
				legacyMembers := 1
				if len(cards) != legacyMembers {
					t.Fatal("legacy wrapper changed destination-filtered IDs")
				}
			})
		}
	}
}

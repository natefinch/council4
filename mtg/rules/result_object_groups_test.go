package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestResultObjectDiscardGroups(t *testing.T) {
	for _, test := range []struct {
		name    string
		player  game.PlayerID
		discard game.Discard
	}{
		{"chosen", game.Player1, game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
		{"random", game.Player1, game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(2), AtRandom: true}},
		{"entire hand", game.Player1, game.Discard{Player: game.ControllerReference(), EntireHand: true}},
		{"player group", game.Player2, game.Discard{PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(2)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for _, cardType := range []types.Card{types.Creature, types.Land} {
				addCardToHand(g, test.player, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{cardType}}})
			}
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: test.discard, PublishResult: "discard",
			}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			snapshots := obj.ResolutionResultObjects["discard"]
			if len(snapshots) != 2 || !resultObjectsMatchSelection(g, obj, snapshots, game.Selection{RequiredTypes: []types.Card{types.Land}}) {
				t.Fatalf("actual discarded members = %#v", snapshots)
			}
		})
	}
}

func TestResultObjectDestroyGroupIgnoresPreventedMemberAndKeepsToken(t *testing.T) {
	for _, tokenPirate := range []bool{false, true} {
		t.Run(map[bool]string{false: "only prevented Pirate", true: "destroyed Pirate token"}[tokenPirate], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			protected := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
			card, _ := g.GetCardInstance(protected.CardInstanceID)
			card.Def.Subtypes = []types.Sub{types.Pirate}
			subtype := types.Sub("Goblin")
			if tokenPirate {
				subtype = types.Pirate
			}
			token := &game.Permanent{
				ObjectID: g.IDGen.Next(), Owner: game.Player1, Controller: game.Player1, Token: true,
				TokenDef: &game.CardDef{CardFace: game.CardFace{
					Types: []types.Card{types.Creature}, Subtypes: []types.Sub{subtype},
				}},
			}
			g.Battlefield = append(g.Battlefield, token)
			obj := &game.StackObject{Controller: game.Player1}
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive:     game.Destroy{Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})},
				PublishResult: "destroy",
			}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			snapshots := obj.ResolutionResultObjects["destroy"]
			if len(snapshots) != 1 || snapshots[0].ObjectID != token.ObjectID {
				t.Fatalf("actual destroyed members = %#v", snapshots)
			}
			if got := resultObjectsMatchSelection(g, obj, snapshots, game.Selection{SubtypesAny: []types.Sub{types.Pirate}}); got != tokenPirate {
				t.Fatalf("Pirate match = %t, want %t", got, tokenPirate)
			}
		})
	}
}

func TestResultObjectMillGroupExcludesRedirectedMembers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Land)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive:     game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(2), Destination: zone.Graveyard},
		PublishResult: "mill",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	snapshots := obj.ResolutionResultObjects["mill"]
	if g.Players[game.Player1].Exile.Size() != 1 || len(snapshots) != 1 {
		t.Fatalf("redirected setup or actual milled members = %#v", snapshots)
	}
	if resultObjectsMatchSelection(g, obj, snapshots, game.Selection{RequiredTypes: []types.Card{types.Land}}) {
		t.Fatal("redirected land matched as an actual milled card")
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLocalProductFrameSeparatesNestedReceiptsObjectsAndScalars(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	obj.ResolvedAmounts = map[string]int{"local": 9, "cost": 3}
	obj.ResolvedExcessDamage = map[string]int{"local": 4}
	obj.ResolutionResults = map[string]game.InstructionResolutionResult{"local": {Accepted: true, Succeeded: true}}
	obj.ResolutionResultObjects = map[string][]game.ObjectSnapshot{"local": {{CardID: id.ID(91)}}}
	address := linkedObjectSourceKey(g, obj, "local-link")
	g.LinkedObjects = map[game.LinkedObjectKey][]game.LinkedObjectRef{
		address: {{CardID: id.ID(92)}},
		{SourceID: address.SourceID, LinkID: "persistent"}: {{CardID: id.ID(93)}},
	}
	sequence := []game.Instruction{{LocalProducts: game.LocalProducts{
		Results: []game.ResultKey{"local"}, Links: []game.LinkedKey{"local-link"},
	}}}
	restore := enterLocalProductFrame(g, obj, sequence)
	localAddress := linkedObjectSourceKey(g, obj, "local-link")
	if _, exists := obj.ResolvedAmounts["local"]; exists || len(g.LinkedObjects[localAddress]) != 0 {
		t.Fatal("frame inherited an outer scalar or linked subject")
	}
	if _, exists := obj.ResolutionResults["local"]; exists || len(obj.ResolutionResultObjects["local"]) != 0 {
		t.Fatal("frame inherited an outer result or result object")
	}
	obj.ResolvedAmounts["local"] = 0
	obj.ResolutionResults["local"] = game.InstructionResolutionResult{Accepted: true}
	obj.ResolutionResultObjects["local"] = []game.ObjectSnapshot{{CardID: id.ID(94)}}
	g.LinkedObjects[localAddress] = []game.LinkedObjectRef{{CardID: id.ID(95)}}
	nestedRestore := enterLocalProductFrame(g, obj, sequence)
	nestedAddress := linkedObjectSourceKey(g, obj, "local-link")
	if _, exists := obj.ResolvedAmounts["local"]; exists || len(g.LinkedObjects[nestedAddress]) != 0 || nestedAddress == localAddress {
		t.Fatal("recursive frame inherited the previous invocation")
	}
	nestedRestore()
	if value, exists := obj.ResolvedAmounts["local"]; !exists || value != 0 || g.LinkedObjects[localAddress][0].CardID != 95 {
		t.Fatal("recursive frame failed to restore known zero or exact outer subject")
	}
	restore()
	if obj.ResolvedAmounts["local"] != 9 || obj.ResolvedExcessDamage["local"] != 4 ||
		!obj.ResolutionResults["local"].Succeeded || obj.ResolutionResultObjects["local"][0].CardID != 91 ||
		g.LinkedObjects[address][0].CardID != 92 || obj.ResolvedAmounts["cost"] != 3 ||
		g.LinkedObjects[game.LinkedObjectKey{SourceID: address.SourceID, LinkID: "persistent"}][0].CardID != 93 {
		t.Fatal("frame lost parent products or persistent facts")
	}
}

func TestLocalProductModesDoNotConsumePreviousOrPersistentSubjects(t *testing.T) {
	for _, selected := range [][]int{{0, 1}, {0, 1, 0, 1}} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player1].Life = 20
		observed := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 2, 3))
		obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, ChosenModes: selected}
		condition := opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
			Object:        opt.Val(game.LinkedObjectReference("local-link")),
			ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		})})
		content := game.AbilityContent{MinModes: 2, MaxModes: 4, Modes: []game.Mode{
			{Sequence: []game.Instruction{
				{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "local-link"},
					LocalProducts: game.LocalProducts{Links: []game.LinkedKey{"local-link"}}},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, Condition: condition},
			}},
			{Sequence: []game.Instruction{
				{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "local-link"},
					ConditionGate: "absent", LocalProducts: game.LocalProducts{Links: []game.LinkedKey{"local-link"}}},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(10)}, Condition: condition},
			}},
		}}
		persistent := linkedObjectSourceKey(g, obj, "local-link")
		g.LinkedObjects = map[game.LinkedObjectKey][]game.LinkedObjectRef{persistent: {{CardID: observed}}}
		NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if got, want := g.Players[game.Player1].Life, 20+len(selected)/2; got != want {
			t.Fatalf("life=%d, want %d: a skipped mode consumed a previous frame", got, want)
		}

		if len(g.LinkedObjects) != 1 || g.LinkedObjects[persistent][0].CardID != observed {
			t.Fatal("local mode escaped into persistent link storage")
		}
	}
}

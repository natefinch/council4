package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

var chosenNameFreshPrefixes = []string{
	"Manifest the top card of your library.",
	"Bolster 1.",
	"Choose a creature type.",
}

func TestChosenNameFreshPrefixProtocolCannotBorrowPermanentName(t *testing.T) {
	choice := chosenNameRuntimeDef(t, "Choose a land you control. "+chosenNameRuntimeSearch)
	for _, prefix := range chosenNameFreshPrefixes {
		t.Run(prefix, func(t *testing.T) {
			def := chosenNameRuntimeDef(t, prefix)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, vanillaCreature("Forest", 1, 1))
			addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			engine := NewEngine(nil)
			log := &TurnLog{}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
			engine.resolveInstructionSequence(g, obj, def.SpellAbility.Val.Modes[0].Sequence, agents, log)
			if _, available := linkedResolutionChoice(obj, string(game.ResolutionChosenPermanentChoiceKey)); available {
				t.Fatal("implicit prefix published the permanent-name choice namespace")
			}
			before, prompts := g.Players[game.Player1].Library.Size(), len(log.Choices)
			// Before public fresh-frame admission, prove the actual standalone
			// prefix cannot supply a stale name to the existing compiled protocol.
			engine.resolveInstructionSequence(g, obj, choice.SpellAbility.Val.Modes[0].Sequence, agents, log)
			if g.Players[game.Player1].Library.Size() != before || len(log.Choices) != prompts {
				t.Fatal("unavailable land choice borrowed an earlier implicit object or type")
			}
		})
	}
}

type freshChosenNameAgent struct {
	name     string
	count    int
	requests []game.ChoiceRequest
}

func (*freshChosenNameAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *freshChosenNameAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind != game.ChoiceSearch && request.Prompt != "Choose a permanent" {
		return []int{request.Options[0].Index}
	}
	count := a.count
	if request.Prompt == "Choose a permanent" {
		count = 1
	}
	var selected []int
	for _, option := range request.Options {
		if option.Label == a.name && len(selected) < count {
			selected = append(selected, option.Index)
		}
	}
	return selected
}

func TestCompiledChosenNameFreshPrefixActualNamesAndCounts(t *testing.T) {
	for _, prefix := range chosenNameFreshPrefixes {
		for _, name := range []string{"Forest", "Island"} {
			for _, count := range []int{0, 1, 2} {
				t.Run(fmt.Sprintf("%s/%s/%d", prefix, name, count), func(t *testing.T) {
					def := chosenNameRuntimeDef(t, prefix+" Choose a land you control. "+chosenNameRuntimeSearch)
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					land := addCombatPermanent(g, game.Player2, basicNamedLand(name))
					land.Controller = game.Player1
					other := "Island"
					if name == other {
						other = "Forest"
					}
					addCombatPermanent(g, game.Player1, basicNamedLand(other))
					addCombatPermanent(g, game.Player1, vanillaCreature(other, 1, 1))
					addCombatPermanent(g, game.Player3, basicNamedLand(name))
					for range 2 {
						addCardToLibrary(g, game.Player1, basicNamedLand(name))
					}
					unmatched := addCardToLibrary(g, game.Player1, basicNamedLand(other))
					if prefix == chosenNameFreshPrefixes[0] {
						addCardToLibrary(g, game.Player1, vanillaCreature(other, 1, 1))
					}
					agent := &freshChosenNameAgent{name: name, count: count}
					obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
					NewEngine(nil).resolveInstructionSequence(g, obj, def.SpellAbility.Val.Modes[0].Sequence,
						[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
					choice, available := linkedResolutionChoice(obj, string(game.ResolutionChosenPermanentChoiceKey))
					if !available || choice.Kind != game.ResolutionChoicePermanent ||
						choice.PermanentID != land.ObjectID || choice.CardName != name {
						t.Fatal("fresh choice borrowed the prefix product, noun, or physical owner")
					}
					if g.Players[game.Player1].Library.Size() != 3-count ||
						!g.Players[game.Player1].Library.Contains(unmatched) {
						t.Fatal("fresh selected name/count did not restrict actual library moves")
					}
					arrivals, searches, choices := 0, 0, 0
					for _, permanent := range g.Battlefield {
						if permanent.Tapped && permanentEffectiveName(g, permanent) == name {
							arrivals++
							if permanent.Owner != game.Player1 || permanent.Controller != game.Player1 {
								t.Fatal("search arrival borrowed prefix or chosen-permanent ownership")
							}
						}
					}
					for _, request := range agent.requests {
						if request.Kind == game.ChoiceSearch {
							searches++
						}
						if request.Prompt == "Choose a permanent" {
							choices++
						}
						if request.Player != game.Player1 {
							t.Fatal("fresh choice/search was offered to another actor")
						}
					}
					if arrivals != count || choices != 1 || searches != 1 {
						t.Fatalf("arrivals/choices/searches=%d/%d/%d", arrivals, choices, searches)
					}
				})
			}
		}
	}
}

func TestCompiledChosenNameFreshPrefixUnavailableLandDoesNotBorrow(t *testing.T) {
	for _, prefix := range chosenNameFreshPrefixes {
		t.Run(prefix, func(t *testing.T) {
			def := chosenNameRuntimeDef(t, prefix+" Choose a land you control. "+chosenNameRuntimeSearch)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, vanillaCreature("Forest", 1, 1))
			addCombatPermanent(g, game.Player2, basicNamedLand("Forest"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
			addCardToLibrary(g, game.Player1, basicNamedLand("Island"))
			before := 2
			if prefix == chosenNameFreshPrefixes[0] {
				before--
			}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			agent := &freshChosenNameAgent{name: "Forest", count: 2}
			NewEngine(nil).resolveInstructionSequence(g, obj, def.SpellAbility.Val.Modes[0].Sequence,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if _, available := linkedResolutionChoice(obj, string(game.ResolutionChosenPermanentChoiceKey)); available ||
				g.Players[game.Player1].Library.Size() != before {
				t.Fatal("failed fresh declaration borrowed implicit prefix name/key")
			}
			for _, request := range agent.requests {
				if request.Kind == game.ChoiceSearch || request.Prompt == "Choose a permanent" {
					t.Fatal("no available land manufactured a fresh choice/search prompt")
				}
			}
		})
	}
}

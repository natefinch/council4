package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Memory Sluice Oracle ID a30b659e-ae2b-4736-9b4f-a9805905bf75, printing
// revision 5fd93980-55f6-43a0-87db-5fa1203aa7b6 (SHM 170), cost {U/B}, with
// exact Oracle text: "Target player mills four cards. Conspire (As you cast
// this spell, you may tap two untapped creatures you control that share a
// color with it. When you do, copy it and you may choose a new target for the
// copy.)" Scryfall rulings URI
// https://api.scryfall.com/cards/5fd93980-55f6-43a0-87db-5fa1203aa7b6/rulings
// returned {"object":"list","has_more":false,"data":[]} on 2026-11-19.
// Official Wizards Comprehensive Rules effective 2026-11-06, confirmed by
// https://magic.wizards.com/en/rules; TXT:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source-review/MagicCompRules-20261106.txt).
// Read 105.2; 601.2b/c/f-i; 608.2b/c/n; 701.17a-b; 702.78a; and
// 707.10/.10c. These govern the spell's colors and hybrid/additional costs,
// its target and cast-time Conspire trigger, copied targets and retargeting,
// target revalidation, resolution to its owner's graveyard, and milling four
// cards or as many as possible from a shorter library.
func TestMemorySluiceConspireMillsDifferentTargetPlayers(t *testing.T) {
	const oracleID = "a30b659e-ae2b-4736-9b4f-a9805905bf75"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Memory Sluice by Oracle ID: %v", err)
	}
	if def.Name != "Memory Sluice" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	spellID := addCardToHand(g, game.Player1, def)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Black)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Black)
	playerTwoLibrary := memorySluiceLibrary(g, game.Player2)
	playerThreeLibrary := memorySluiceLibrary(g, game.Player3)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer &&
					option.Targets[0].PlayerID == game.Player3 {
					return []int{option.Index}
				}
			}
			t.Fatal("Player 3 was not offered as a legal new target for the copy")
		}
		return chooseConspireCost(req)
	})
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil),
		agents, log) {
		t.Fatal("could not cast Memory Sluice targeting Player 2")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 1 ||
		!g.Players[game.Player1].ManaPool.IsEmpty() || len(original.Targets) != 1 ||
		original.Targets[0].PlayerID != game.Player2 {
		t.Fatalf("Memory Sluice did not pay its hybrid and Conspire costs: %+v", original)
	}
	for _, creature := range []*game.Permanent{first, second} {
		if permanent, _ := permanentByObjectID(g, creature.ObjectID); !permanent.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !isConspireTrigger(trigger) {
		t.Fatal("the top stack object is not Memory Sluice's Conspire trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.CastCopy || len(copy.Targets) != 1 ||
		copy.Targets[0].PlayerID != game.Player3 || original.Targets[0].PlayerID != game.Player2 {
		t.Fatalf("Conspire did not copy Memory Sluice with its new target: %+v", copy)
	}

	engine.resolveTopOfStack(g, log)
	assertMemorySluiceMilledFour(t, g, game.Player3, playerThreeLibrary)
	engine.resolveTopOfStack(g, log)
	assertMemorySluiceMilledFour(t, g, game.Player2, playerTwoLibrary)
	if !g.Players[game.Player1].Graveyard.Contains(spellID) || g.Stack.Size() != 0 {
		t.Fatal("the original Memory Sluice did not finish in its owner's graveyard")
	}
}

func memorySluiceLibrary(g *game.Game, player game.PlayerID) []id.ID {
	library := make([]id.ID, 0, 5)
	for range 5 {
		library = append(library, addCardToLibrary(g, player, &game.CardDef{
			CardFace: game.CardFace{Name: "Memory Sluice library card"},
		}))
	}
	return library
}

func assertMemorySluiceMilledFour(t *testing.T, g *game.Game, player game.PlayerID, library []id.ID) {
	t.Helper()
	milled := 0
	for _, cardID := range library {
		if g.Players[player].Graveyard.Contains(cardID) {
			milled++
		} else if !g.Players[player].Library.Contains(cardID) {
			t.Fatalf("player %d's library card %d left its library without reaching their graveyard", player, cardID)
		}
	}
	if milled != 4 || g.Players[player].Library.Size() != 1 {
		t.Fatalf("player %d had %d cards milled and %d remaining, want 4 and 1",
			player, milled, g.Players[player].Library.Size())
	}
}

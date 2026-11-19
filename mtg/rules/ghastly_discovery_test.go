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

// Source review: Ghastly Discovery, Oracle ID
// 1ca875f0-50a7-4332-8643-c2b4bbda18b1, pinned printing revision
// 9dbd510a-d71f-4b0c-bc6c-e403f632cbdb (SHM 39), from pinned corpus
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Oracle text: "Draw two cards, then discard a card. Conspire (As you cast
// this spell, you may tap two untapped creatures you control that share a
// color with it. When you do, copy it.)" Its Scryfall rulings endpoint,
// https://api.scryfall.com/cards/9dbd510a-d71f-4b0c-bc6c-e403f632cbdb/rulings,
// returned an empty list when retrieved 2026-11-19.
// Official Wizards Comprehensive Rules, effective 2026-11-06, confirmed by
// https://magic.wizards.com/en/rules; TXT:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (retrieved copy .cardwork/source-review/MagicCompRules-20261106.txt).
// Read 105.2-3 (color), 121.1-2 (individual draws), 307.1 (sorcery timing),
// 601.2a-i (casting and cost payment), 608.2c-e (ordered resolution and
// resolution-time choices), 701.9a-b (discard and its chooser), 702.78a
// (Conspire's additional cost and cast trigger), and 707.10 (copies repeat
// spell characteristics but not choices made on resolution). The spell draws
// two before its controller chooses one card to discard on each resolution;
// paying Conspire copies the spell without casting the copy.
func TestGhastlyDiscoveryDrawsThenDiscardsWithConspire(t *testing.T) {
	const oracleID = "1ca875f0-50a7-4332-8643-c2b4bbda18b1"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Ghastly Discovery by Oracle ID: %v", err)
	}
	if def.Name != "Ghastly Discovery" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	firstDiscard := addCardToHand(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "First discard"},
	})
	secondDiscard := addCardToHand(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Second discard"},
	})
	spellID := addCardToHand(g, game.Player1, def)
	drawn := make([]id.ID, 0, 4)
	drawnNames := []string{"Drawn A", "Drawn B", "Drawn C", "Drawn D"}
	for _, name := range drawnNames {
		drawn = append(drawn, addCardToLibrary(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: name},
		}))
	}
	firstCreature := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	secondCreature := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	discardNames := []string{"First discard", "Second discard"}
	discardChoices := 0
	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose spell cost" {
			return chooseConspireCost(req)
		}
		hasCards := false
		for _, option := range req.Options {
			hasCards = hasCards || (option.Card.Exists && len(option.Targets) == 0)
		}
		if hasCards {
			if discardChoices >= len(discardNames) {
				t.Fatal("Ghastly Discovery asked for an unexpected extra discard")
			}
			drawnOptions := 0
			for _, name := range drawnNames {
				for _, option := range req.Options {
					if option.Card.Exists && len(option.Targets) == 0 && option.Card.Val.Name == name {
						drawnOptions++
					}
				}
			}
			if drawnOptions != 2*(discardChoices+1) {
				t.Fatalf("discard choice included %d drawn cards after resolution %d, want %d",
					drawnOptions, discardChoices+1, 2*(discardChoices+1))
			}
			want := discardNames[discardChoices]
			for _, option := range req.Options {
				if option.Card.Exists && len(option.Targets) == 0 && option.Card.Val.Name == want {
					discardChoices++
					return []int{option.Index}
				}
			}
			t.Fatalf("expected %q in discard choices: %+v", want, req.Options)
		}
		return req.DefaultSelection
	})
	log := &TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1,
		action.CastSpell(spellID, nil, 0, nil), agents, log) {
		t.Fatal("could not cast Ghastly Discovery")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ConspirePaid != 1 || original.ManaSpentToCast != 3 ||
		!g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("expected Ghastly Discovery's mana and Conspire costs to be paid: %+v", original)
	}
	for _, creature := range []*game.Permanent{firstCreature, secondCreature} {
		if permanent, ok := permanentByObjectID(g, creature.ObjectID); !ok || !permanent.Tapped {
			t.Fatalf("Conspire did not tap creature %d", creature.ObjectID)
		}
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("paid Conspire did not put its cast trigger on the stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !isConspireTrigger(trigger) {
		t.Fatal("the top stack object is not Ghastly Discovery's Conspire trigger")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.CastCopy || copy.ConspirePaid != 1 {
		t.Fatalf("Conspire did not put a spell copy above the original: %+v", copy)
	}

	engine.resolveTopOfStackWithChoices(g, agents, log)
	remaining, ok := g.Stack.Peek()
	if !g.Players[game.Player1].Graveyard.Contains(firstDiscard) ||
		g.Players[game.Player1].Graveyard.Contains(secondDiscard) ||
		g.Players[game.Player1].Library.Size() != 2 || !ok || remaining.ID != original.ID {
		t.Fatal("the copy did not draw two cards, then discard its own resolution choice")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)

	player := g.Players[game.Player1]
	if discardChoices != 2 || !player.Graveyard.Contains(firstDiscard) ||
		!player.Graveyard.Contains(secondDiscard) || !player.Graveyard.Contains(spellID) ||
		player.Graveyard.Size() != 3 || player.Library.Size() != 0 || player.Hand.Size() != 4 {
		t.Fatalf("hand/discards/library = %d/%d/%d; want two independent discards and four draws",
			player.Hand.Size(), player.Graveyard.Size(), player.Library.Size())
	}
	for _, cardID := range drawn {
		if !player.Hand.Contains(cardID) {
			t.Fatalf("Ghastly Discovery failed to leave drawn card %d in hand", cardID)
		}
	}
}

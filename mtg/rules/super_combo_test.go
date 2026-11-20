package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Super Combo Oracle ID 68555673-2833-459a-b415-e2805f837778, printing
// 52bed0d4-3bcf-4770-b9c9-2a1d0666d220 (TMC 34), {1}{G}, Sorcery:
// “Replicate {2} (When you cast this spell, copy it for each time you paid
// its replicate cost. You may choose new targets for the copies.) Target
// creature you control deals damage equal to its power to target creature an
// opponent controls.” The exact printing's Scryfall card and rulings endpoints
// were retrieved 2026-11-19:
// https://api.scryfall.com/cards/52bed0d4-3bcf-4770-b9c9-2a1d0666d220 and
// https://api.scryfall.com/cards/52bed0d4-3bcf-4770-b9c9-2a1d0666d220/rulings.
// The two WotC rulings (2026-01-27) say copies are individually counterable
// and a copy deals no damage if either target is illegal on resolution.
// Read the official Comprehensive Rules, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Sections 601.2b/c/f-i govern optional additional-cost announcements,
// target choices, paying the total cost, and the cast event; 603.2-4 and
// 702.56a-b govern the Replicate cast trigger and its paid count; 707.10/.10c
// govern the copy and optional new targets; 608.2b/h govern legal targets and
// determining the creature's power as damage resolves, with 113.7a supplying
// last known information if the creature leaves; 120.1 defines damage dealt by
// the chosen creature as its source.
func TestSuperComboReplicateScenario(t *testing.T) {
	const oracleID = "68555673-2833-459a-b415-e2805f837778"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Super Combo by Oracle ID: %v", err)
	}
	if def.Name != "Super Combo" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	source := addCombatPermanent(g, game.Player1, vanillaCreature("Power source", 2, 2))
	target := addCombatPermanent(g, game.Player2, vanillaCreature("Opponent target", 7, 7))
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	setMainPhasePriority(g, game.Player1)

	agents := conspireTestAgent(func(request game.ChoiceRequest) []int {
		if request.Prompt == "Choose new targets" {
			for _, option := range request.Options {
				if option.Label == "Keep current targets" {
					return []int{option.Index}
				}
			}
			t.Fatal("Replicate copy did not offer to keep the announced targets")
		}
		return request.DefaultSelection
	})
	log := &TurnLog{}
	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID: cardID, SourceZone: zone.Hand,
		Targets: []game.Target{
			game.PermanentTarget(source.ObjectID),
			game.PermanentTarget(target.ObjectID),
		},
		ReplicateCounts: []int{2},
	})
	if !engine.applyActionWithChoices(g, game.Player1, cast, agents, log) {
		t.Fatal("could not cast Super Combo while paying Replicate twice")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 6 ||
		len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 2 {
		t.Fatalf("cast did not pay {1}{G} and Replicate {2} twice: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) || g.Stack.Size() != 2 {
		t.Fatal("two paid Replicate instances did not put the cast trigger above the original")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if g.Stack.Size() != 3 {
		t.Fatalf("Replicate created %d total stack objects, want the original and two copies", g.Stack.Size())
	}
	for range 3 {
		engine.resolveTopOfStackWithChoices(g, agents, log)
	}

	damageEvents := 0
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt {
			continue
		}
		damageEvents++
		if event.Amount != 2 || event.SourceObjectID != source.ObjectID {
			t.Fatalf("Super Combo dealt damage with the wrong source or amount: %+v", event)
		}
	}
	if damageEvents != 3 || target.MarkedDamage != 6 {
		t.Fatalf("damage events=%d, target marked damage=%d; want three 2-damage resolutions", damageEvents, target.MarkedDamage)
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved Super Combo did not leave the stack for its owner's graveyard")
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 9cc64a91-82d2-4be1-812f-531863e99801, printing/revision
// bf9f85b3-eb8c-43de-aaac-8a887e9c58da (Guildpact #26), {U}, Instant.
// Exact text: “Replicate {U} (When you cast this spell, copy it for each time
// you paid its replicate cost. You may choose new targets for the copies.) Tap
// target permanent.” On 2026-11-19, the Scryfall card endpoint
// https://api.scryfall.com/cards/bf9f85b3-eb8c-43de-aaac-8a887e9c58da confirmed
// that text; its rulings endpoint
// https://api.scryfall.com/cards/bf9f85b3-eb8c-43de-aaac-8a887e9c58da/rulings
// returned no rulings.
// Read the official Wizards Comprehensive Rules effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Sections 115.1a and 601.2b/c/f-i require choosing the permanent target and
// announcing/paying Replicate with the spell's other costs before it becomes
// cast; 603.2-3 put the paid Replicate trigger on the stack; 608.2b rechecks
// target legality; 701.26a taps only an untapped permanent; 702.56a creates a
// copy per Replicate payment and allows new targets; 707.10/.10c make those
// copies not cast and permit changing their targets to legal ones.
func TestGigadrowseReplicateTapsTargetPermanent(t *testing.T) {
	const oracleID = "9cc64a91-82d2-4be1-812f-531863e99801"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Gigadrowse by Oracle ID: %v", err)
	}
	if def.Name != "Gigadrowse" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addCardToHand(g, game.Player1, def)
	target := addCreaturePermanent(g, game.Player2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	setMainPhasePriority(g, game.Player1)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID:          card,
		SourceZone:      zone.Hand,
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
		ReplicateCounts: []int{2},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("could not cast Gigadrowse paying Replicate twice")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 3 || g.Players[game.Player1].ManaPool.Total() != 0 ||
		len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 2 {
		t.Fatalf("expected {U} plus two {U} Replicate payments: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("two Replicate payments did not trigger above Gigadrowse")
	}

	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if option.Label == "Keep current targets" {
					return []int{option.Index}
				}
			}
			t.Fatal("Replicate copy did not offer to keep its target")
		}
		return req.DefaultSelection
	})
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if g.Stack.Size() != 3 {
		t.Fatalf("two Replicate payments made %d stack objects, want two copies and Gigadrowse", g.Stack.Size())
	}
	for range 3 {
		engine.resolveTopOfStackWithChoices(g, agents, log)
	}
	permanent, ok := permanentByObjectID(g, target.ObjectID)
	if !ok || !permanent.Tapped {
		t.Fatal("Gigadrowse and its Replicate copies did not tap the target permanent")
	}
	if !g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("resolved Gigadrowse did not leave the stack for its owner's graveyard")
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID fe725b27-49ef-41bc-91f9-9cd9871ad863, printing d6dcff21-5900-43c4-a38b-cdc19c704ce4
// (Guildpact #75), {R}, Sorcery. Exact text: “Replicate {R} (When you cast
// this spell, copy it for each time you paid its replicate cost. You may
// choose new targets for the copies.) Destroy target artifact.” Live Scryfall
// endpoints retrieved 2026-11-20:
// https://api.scryfall.com/cards/d6dcff21-5900-43c4-a38b-cdc19c704ce4 and
// https://api.scryfall.com/cards/d6dcff21-5900-43c4-a38b-cdc19c704ce4/rulings.
// The card endpoint confirms the pinned text. Its two WotC rulings, both
// published 2024-01-12, say copies are made even if the original has left the
// stack and that replicate copies are put on the stack, not cast.
//
// Read the official Wizards Comprehensive Rules effective November 6, 2026:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Sections 601.2b/c/f-i require announcing and paying Replicate with the spell's
// other costs and choosing its target; 603.2-3 and 113.7a require its cast
// trigger and independent resolution; 608.2b rechecks target legality;
// 701.8a-c define destruction; 702.56a-b require a copy for each payment; and
// 707.10 requires spell copies to retain cast choices while permitting the
// Replicate trigger to choose new targets.
func TestAuthoredShatteringSpreeScenario(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("fe725b27-49ef-41bc-91f9-9cd9871ad863")
	if err != nil {
		t.Fatalf("lookup Shattering Spree by Oracle ID: %v", err)
	}
	if def.Name != "Shattering Spree" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	card := addCardToHand(g, game.Player1, def)
	originalTarget := addNamedTypedPermanent(g, game.Player2, "Original artifact", types.Artifact)
	firstCopyTarget := addNamedTypedPermanent(g, game.Player2, "First copy artifact", types.Artifact)
	secondCopyTarget := addNamedTypedPermanent(g, game.Player2, "Second copy artifact", types.Artifact)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	setMainPhasePriority(g, game.Player1)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID: card, SourceZone: zone.Hand,
		Targets:         []game.Target{game.PermanentTarget(originalTarget.ObjectID)},
		ReplicateCounts: []int{2},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("could not cast Shattering Spree paying Replicate twice")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 3 || g.Players[game.Player1].ManaPool.Total() != 0 ||
		len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 2 {
		t.Fatalf("expected {R} plus two {R} Replicate payments: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("two Replicate payments did not trigger above Shattering Spree")
	}

	newTargets := []game.ObjectID{firstCopyTarget.ObjectID, secondCopyTarget.ObjectID}
	nextTarget := 0
	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			if nextTarget >= len(newTargets) {
				t.Fatal("asked for more Replicate target choices than copies")
			}
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == newTargets[nextTarget] {
					nextTarget++
					return []int{option.Index}
				}
			}
			t.Fatalf("Replicate copy was not offered target %d", newTargets[nextTarget])
		}
		return req.DefaultSelection
	})
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if nextTarget != 2 || g.Stack.Size() != 3 {
		t.Fatal("Replicate did not create two copies with new target choices")
	}
	copiedTargets := make(map[game.ObjectID]bool, len(newTargets))
	for _, object := range g.Stack.Objects() {
		if !object.Copy {
			continue
		}
		if object.CastCopy || len(object.Targets) != 1 {
			t.Fatalf("Replicate produced an invalid spell copy: %+v", object)
		}
		copiedTargets[object.Targets[0].PermanentID] = true
	}
	for _, target := range newTargets {
		if !copiedTargets[target] {
			t.Fatalf("Replicate copy did not target artifact %d", target)
		}
	}

	for range 3 {
		engine.resolveTopOfStack(g, log)
	}
	for _, target := range []*game.Permanent{originalTarget, firstCopyTarget, secondCopyTarget} {
		if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
			t.Fatalf("Shattering Spree did not destroy artifact %q", g.CardInstances[target.CardInstanceID].Def.Name)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("resolved Shattering Spree did not leave the stack for its owner's graveyard")
	}
}

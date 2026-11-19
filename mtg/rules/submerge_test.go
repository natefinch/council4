package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: Oracle ID 99427ebe-c00d-4206-84ca-9764f6e952c6, Nemesis #48,
// printing revision d2741fe4-37fe-427f-ae85-5107991d4eee; pinned corpus
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle verified 2026-11-19 from
// https://api.scryfall.com/cards/d2741fe4-37fe-427f-ae85-5107991d4eee:
// {4}{U} Instant. "If an opponent controls a Forest and you control an Island,
// you may cast this spell without paying its mana cost. Put target creature on
// top of its owner's library."
// Rulings retrieved 2026-11-19 from
// https://api.scryfall.com/cards/d2741fe4-37fe-427f-ae85-5107991d4eee/rulings;
// the endpoint returned an empty list.
// Official Wizards Comprehensive Rules effective 2026-09-25 (the pinned rules
// edition), https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx.
// Read 108.3, 109.5, 118.9, 305.6, 400.7, 401.2/401.4, 601.2b/c/e-h, and
// 608.2b-c: the prospective caster's controlled Island and any opponent's
// Forest enable the optional alternate cost during casting; the creature is
// targeted while casting and revalidated on resolution; the effect moves that
// creature to its owner's library, not its controller's.
func TestAuthoredSubmergeConditionalAlternativeCostPutsTargetOnOwnerLibrary(t *testing.T) {
	const oracleID = "99427ebe-c00d-4206-84ca-9764f6e952c6"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Submerge by Oracle ID: %v", err)
	}
	if def.Name != "Submerge" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	spellID := addCardToHand(g, game.Player1, def)
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	island.Tapped = true
	forest := addBasicLandPermanent(g, game.Player3, types.Forest)
	forest.Tapped = true
	target := addCreaturePermanent(g, game.Player2)
	target.Owner = game.Player3
	g.CardInstances[target.CardInstanceID].Owner = game.Player3
	libraryCard := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Existing library card"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Submerge was not offered with an opponent's Forest and the caster's Island")
	}
	choseAlternative := false
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(request game.ChoiceRequest) []int {
			if request.Prompt != "Choose spell cost" {
				return request.DefaultSelection
			}
			hasNormal, hasAlternative := false, false
			alternativeIndex := -1
			for _, option := range request.Options {
				switch option.Label {
				case "Normal cost":
					hasNormal = true
				case "Cast without paying its mana cost":
					hasAlternative = true
					alternativeIndex = option.Index
				}
			}
			if !hasNormal || !hasAlternative {
				t.Fatalf("Submerge cost choices = %+v, want normal and conditional alternative costs", request.Options)
			}
			choseAlternative = true
			return []int{alternativeIndex}
		},
	}}
	if !engine.applyActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}) {
		t.Fatal("could not cast Submerge using its conditional alternative cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !choseAlternative || spell.ManaSpentToCast != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.C) != 4 ||
		g.Players[game.Player1].ManaPool.Amount(mana.U) != 1 {
		t.Fatalf("Submerge did not use the chosen free alternative cost: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("Submerge left its target on the battlefield")
	}
	if top, ok := g.Players[game.Player3].Library.Top(); !ok || top != target.CardInstanceID {
		t.Fatalf("target on owner's library top = %d (%t), want owner %d card %d",
			top, ok, game.Player3, target.CardInstanceID)
	}
	if !g.Players[game.Player3].Library.Contains(libraryCard) ||
		g.Players[game.Player2].Library.Contains(target.CardInstanceID) {
		t.Fatal("Submerge did not preserve the library card and use the target's owner rather than controller")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Submerge did not reach its owner's graveyard")
	}
}

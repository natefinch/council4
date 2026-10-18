package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	ghostVacuumOracleID                        = "4ef583ab-e0a0-4e4b-bd65-af766038b1fd"
	ghostVacuumProducer game.LinkedProducerKey = ghostVacuumOracleID + "/activated/exile"
)

func ghostVacuumProductionFixture(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(ghostVacuumOracleID)
	if err != nil || def == nil {
		t.Fatalf("Ghost Vacuum exact registry entry: %v", err)
	}
	if problems := game.ValidateCardDef(def); len(problems) != 0 {
		t.Fatalf("Ghost Vacuum production declaration: %+v", problems)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	source := addCombatPermanent(g, game.Player1, def)
	return g, NewEngineWithPrograms(nil, registry.Programs()), source
}

func activateProductionGhostVacuumExile(t *testing.T, g *game.Game, e *Engine, source *game.Permanent, cardID id.ID) {
	t.Helper()
	source.Tapped = false
	if !e.applyAction(g, source.Controller, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{currentCardTarget(t, g, cardID)}, 0)) {
		t.Fatal("Ghost Vacuum could not activate its targeted graveyard exile")
	}
	e.resolveTopOfStack(g, &TurnLog{})
}

func TestGhostVacuumProductionFilteredReturnAndEntryReplacements(t *testing.T) {
	g, e, source := ghostVacuumProductionFixture(t)
	creatureArtifact := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Vacuum Elf", Types: []types.Card{types.Artifact, types.Creature},
		Subtypes: []types.Sub{types.Elf}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 5}),
	}})
	creature := addCardToGraveyard(g, game.Player3, vanillaCreature("Vacuum Beast", 6, 7))
	noncreature := addCardToGraveyard(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
		Name: "Linked noncreature", Types: []types.Card{types.Artifact},
	}})
	unlinked := addCardInstance(g, game.Player2, vanillaCreature("Unlinked exile", 2, 2))
	g.Players[game.Player2].Exile.Add(unlinked)
	for _, cardID := range []id.ID{creatureArtifact, creature, noncreature} {
		activateProductionGhostVacuumExile(t, g, e, source, cardID)
	}

	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventPermanentEnteredBattlefield,
		EntersTapped: true, EntersTappedOthers: true,
		EntersTappedSelection: &game.Selection{SubtypesAny: []types.Sub{types.Spirit},
			Power: opt.Val(compare.Int{Op: compare.Equal, Value: 1})},
	})
	addReplacementPermanent(t, g, game.Player1, doublingSeasonCounterReplacementCardDef())
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	source.Tapped = false
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("Ghost Vacuum could not pay for its sorcery-speed return activation")
	}
	if permanentByCardID(g, source.CardInstanceID) != nil {
		t.Fatal("Ghost Vacuum's sacrifice cost did not remove its source")
	}
	e.resolveTopOfStack(g, &TurnLog{})

	for _, cardID := range []id.ID{creatureArtifact, creature} {
		permanent := assertVacuumSpirit(t, g, cardID, game.Player1, 2)
		if !permanent.Tapped {
			t.Fatalf("entry replacement did not see the returned Spirit characteristics for card %v", cardID)
		}
	}
	first := permanentByCardID(g, creatureArtifact)
	if !permanentHasType(g, first, types.Artifact) || !permanentHasSubtype(g, first, types.Elf) {
		t.Fatal("Spirit subtype replaced the original artifact creature types")
	}
	if !g.Players[game.Player4].Exile.Contains(noncreature) || !g.Players[game.Player2].Exile.Contains(unlinked) {
		t.Fatal("return included a linked noncreature or an unlinked exile")
	}
	var batch game.ObjectID
	arrivals := 0
	for _, event := range g.Events {
		if event.Kind != game.EventPermanentEnteredBattlefield ||
			(event.CardID != creatureArtifact && event.CardID != creature) {
			continue
		}
		arrivals++
		if event.SimultaneousID == 0 || batch != 0 && batch != event.SimultaneousID {
			t.Fatal("Ghost Vacuum's creature returns were not one simultaneous entry")
		}
		batch = event.SimultaneousID
	}
	if arrivals != 2 {
		t.Fatalf("Ghost Vacuum creature arrivals = %d, want 2", arrivals)
	}
}

func TestGhostVacuumProductionRejectsStaleGraveyardTarget(t *testing.T) {
	g, e, source := ghostVacuumProductionFixture(t)
	cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Stale target", 4, 4))
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{currentCardTarget(t, g, cardID)}, 0)) {
		t.Fatal("Ghost Vacuum could not target the card in its graveyard")
	}
	moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Hand)
	moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Graveyard)
	e.resolveTopOfStack(g, &TurnLog{})
	key := game.LinkedObjectKey{SourceID: source.ObjectID, LinkID: string(ghostVacuumProducer), Producer: ghostVacuumProducer}
	if !g.Players[game.Player2].Graveyard.Contains(cardID) || len(linkedObjects(g, key)) != 0 {
		t.Fatal("the returned-to-graveyard target was exiled or linked as its old incarnation")
	}
}

func TestGhostVacuumProductionUsesSorceryTimingAndActivationCosts(t *testing.T) {
	g, e, source := ghostVacuumProductionFixture(t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	g.Turn.Phase = game.PhaseBeginning
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("Ghost Vacuum activated its sorcery-only return outside a main phase")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	source.Tapped = true
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("Ghost Vacuum ignored its tap cost")
	}
	source.Tapped = false
	g.Players[game.Player1].ManaPool.Empty()
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("Ghost Vacuum ignored its six-mana cost")
	}
	if permanentByCardID(g, source.CardInstanceID) == nil || g.Stack.Size() != 0 {
		t.Fatal("an illegal return activation paid the sacrifice cost or reached the stack")
	}
}

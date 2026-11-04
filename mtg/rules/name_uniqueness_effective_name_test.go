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

// Source review (#3616): Guardian Project RVR 146 printing
// 2c523ef7-d6c4-4887-9a4e-0c2bc6f63676 and Yenna, Redtooth Regent WOE 219
// printing e635d461-254a-434e-8e5d-dea61dd8ca4f, Oracle text from the pinned
// corpus. https://api.scryfall.com/cards/2c523ef7-d6c4-4887-9a4e-0c2bc6f63676/rulings
// returned four 2019-01-25 rulings: the name check happens on entry and on
// resolution, a same card that returned or hit the graveyard shares the name,
// and "A face-down creature has no name, so it can't share a name with
// anything." Yenna's five 2023-09-01 rulings cover the copy only. CR 2026-09-25
// (verified official TXT copy) read: 201.2a (same name means a name in common;
// a nameless object shares no name), 201.2c (only "different name" wording
// excludes nameless objects), 603.4 (intervening if), 613.1a and 707.2 (a copy's
// copiable name replaces its printed name), 708.2/708.2a (a face-down permanent
// has no name). Both checks therefore compare current effective names of the
// other permanents, never the printed name of the card beneath a copy.
const (
	nameUniquenessGuardianProjectOracleID = "4f9e07ae-6341-4b46-9f77-f17ab659d266"
	nameUniquenessYennaOracleID           = "0adede77-5181-4190-8066-4797cefe0189"
	nameUniquenessCloneOracleID           = "42226b87-0746-4ebf-9fd0-108d508462af"
	nameUniquenessGrizzlyBearsOracleID    = "14c8f55d-d177-4c25-a931-ebeb9e6062a0"
	nameUniquenessCourserOracleID         = "46779609-4fa7-4fd2-b5b4-7d4d749339e6"
)

func nameUniquenessCard(t *testing.T, registry *cards.Registry, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		// Legacy name-only registrations carry no Oracle identity; require
		// the printed name to be unambiguous instead.
		matches, lookupErr := registry.LookupAllChecked(name)
		if lookupErr != nil || len(matches) != 1 {
			t.Fatalf("default-registry lookup of %s failed: %v; by name: %d matches, %v", name, err, len(matches), lookupErr)
		}
		def = matches[0]
	}
	if def.Name != name {
		t.Fatalf("oracle %s resolved to %q, want %q", oracleID, def.Name, name)
	}
	return def
}

// resolveCloneCopying resolves a default-registry Clone spell for controller
// while exactly one creature is on the battlefield, so it enters as a copy of
// that creature.
func resolveCloneCopying(t *testing.T, g *game.Game, e *Engine, clone *game.CardDef, controller game.PlayerID, wantName string) *game.Permanent {
	t.Helper()
	addImplementationSpellToStack(g, controller, clone, nil)
	var agents [game.NumPlayers]PlayerAgent
	agents[controller] = &choiceOnlyAgent{choices: [][]int{{1}}}
	e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	copied := g.Battlefield[len(g.Battlefield)-1]
	card, ok := g.GetCardInstance(copied.CardInstanceID)
	if !ok || card.Def.Name != clone.Name || permanentEffectiveName(g, copied) != wantName {
		t.Fatalf("Clone entered as %q, want a physical Clone copying %q", permanentEffectiveName(g, copied), wantName)
	}
	return copied
}

type guardianProjectScenario struct {
	g     *game.Game
	e     *Engine
	bears *game.CardDef
	clone *game.CardDef
}

func newGuardianProjectScenario(t *testing.T) guardianProjectScenario {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	s := guardianProjectScenario{
		g:     game.NewGame([game.NumPlayers]game.PlayerConfig{}),
		e:     NewEngine(nil),
		bears: nameUniquenessCard(t, registry, nameUniquenessGrizzlyBearsOracleID, "Grizzly Bears"),
		clone: nameUniquenessCard(t, registry, nameUniquenessCloneOracleID, "Clone"),
	}
	addCombatPermanent(s.g, game.Player1, nameUniquenessCard(t, registry, nameUniquenessGuardianProjectOracleID, "Guardian Project"))
	for range 4 {
		addCardToLibrary(s.g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Guardian Project library card"}})
	}
	return s
}

// settle puts every pending trigger on the stack and resolves the stack, and
// reports how many cards Player1 drew.
func (s guardianProjectScenario) settle(t *testing.T) int {
	t.Helper()
	before := s.g.Players[game.Player1].Hand.Size()
	for range 8 {
		s.e.putTriggeredAbilitiesOnStack(s.g)
		if s.g.Stack.Size() == 0 {
			return s.g.Players[game.Player1].Hand.Size() - before
		}
		s.e.resolveTopOfStack(s.g, &TurnLog{})
	}
	t.Fatal("stack did not settle")
	return 0
}

// castBears resolves a Grizzly Bears spell for Player1 and reports how many
// cards Guardian Project drew for its entry.
func (s guardianProjectScenario) castBears(t *testing.T) int {
	t.Helper()
	s.g.TriggerEventCursor = len(s.g.Events)
	addImplementationSpellToStack(s.g, game.Player1, s.bears, nil)
	s.e.resolveTopOfStack(s.g, &TurnLog{})
	return s.settle(t)
}

// TestGuardianProjectComparesCopiedCreatureEffectiveName proves Guardian
// Project's intervening condition reads another creature's current effective
// name: a Clone that entered as Grizzly Bears shares the name of an entering
// Grizzly Bears even though its printed name is Clone (CR 707.2, 201.2a).
func TestGuardianProjectComparesCopiedCreatureEffectiveName(t *testing.T) {
	s := newGuardianProjectScenario(t)
	original := addCombatPermanent(s.g, game.Player1, s.bears)
	resolveCloneCopying(t, s.g, s.e, s.clone, game.Player1, "Grizzly Bears")
	s.g.TriggerEventCursor = len(s.g.Events)
	if !movePermanentToZone(s.g, original, zone.Exile) {
		t.Fatal("could not exile the original Grizzly Bears")
	}
	s.settle(t)
	if drew := s.castBears(t); drew != 0 {
		t.Fatalf("drew %d card(s) for a Grizzly Bears sharing a name with Clone-as-Grizzly Bears", drew)
	}
}

// TestGuardianProjectDrawsWhenCopiedCreatureHasAnotherName is the positive
// counterpart: a Clone that copied a different creature has that creature's
// name, so the entering Grizzly Bears is the first of its name.
func TestGuardianProjectDrawsWhenCopiedCreatureHasAnotherName(t *testing.T) {
	s := newGuardianProjectScenario(t)
	other := addCombatPermanent(s.g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Copied Blueprint", Types: []types.Card{types.Creature},
	}})
	resolveCloneCopying(t, s.g, s.e, s.clone, game.Player1, "Copied Blueprint")
	if !movePermanentToZone(s.g, other, zone.Exile) {
		t.Fatal("could not exile the copied blueprint")
	}
	s.settle(t)
	if drew := s.castBears(t); drew != 1 {
		t.Fatalf("drew %d card(s) for the first Grizzly Bears, want 1", drew)
	}
}

// TestGuardianProjectIgnoresFaceDownCreaturesName proves a face-down creature
// has no name (CR 708.2a, Guardian Project ruling): a face-down Grizzly Bears
// card does not stop a face-up Grizzly Bears from drawing, and its hidden
// printed name is not consulted.
func TestGuardianProjectIgnoresFaceDownCreaturesName(t *testing.T) {
	s := newGuardianProjectScenario(t)
	addFaceDownPermanent(s.g, game.Player1, s.bears, game.FaceDownManifest)
	if drew := s.castBears(t); drew != 1 {
		t.Fatalf("drew %d card(s) beside a face-down Grizzly Bears, want 1", drew)
	}
}

// TestGuardianProjectDrawsForNamelessEnteringCreature proves a creature that
// enters face down has no name and so shares none (CR 201.2a), even with a
// face-up Grizzly Bears that is the same card beneath.
func TestGuardianProjectDrawsForNamelessEnteringCreature(t *testing.T) {
	s := newGuardianProjectScenario(t)
	addCombatPermanent(s.g, game.Player1, s.bears)
	addCardToLibrary(s.g, game.Player1, s.bears)
	s.g.TriggerEventCursor = len(s.g.Events)
	manifested, ok := s.e.manifestTopCard(s.g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}, game.Player1, game.FaceDownManifest)
	if !ok || !manifested.FaceDown {
		t.Fatal("manifest did not put a face-down creature onto the battlefield")
	}
	if drew := s.settle(t); drew != 1 {
		t.Fatalf("drew %d card(s) for a nameless entering creature, want 1", drew)
	}
}

// TestGuardianProjectGraveyardComparisonReadsCardNames preserves the separate
// graveyard comparison: a Clone that died as a copy is a card named Clone in
// the graveyard, while a Grizzly Bears card there still blocks the draw.
func TestGuardianProjectGraveyardComparisonReadsCardNames(t *testing.T) {
	s := newGuardianProjectScenario(t)
	original := addCombatPermanent(s.g, game.Player1, s.bears)
	copied := resolveCloneCopying(t, s.g, s.e, s.clone, game.Player1, "Grizzly Bears")
	if !movePermanentToZone(s.g, original, zone.Exile) || !movePermanentToZone(s.g, copied, zone.Graveyard) {
		t.Fatal("could not clear the battlefield Grizzly Bears")
	}
	s.g.TriggerEventCursor = len(s.g.Events)
	if drew := s.castBears(t); drew != 1 {
		t.Fatalf("drew %d card(s) with only a Clone card in the graveyard, want 1", drew)
	}

	s = newGuardianProjectScenario(t)
	addCardToGraveyard(s.g, game.Player1, s.bears)
	if drew := s.castBears(t); drew != 0 {
		t.Fatalf("drew %d card(s) with a Grizzly Bears card in the graveyard", drew)
	}
}

// TestYennaTargetComparesCopiedPermanentEffectiveName proves Yenna's "target
// enchantment you control that doesn't have the same name as another permanent
// you control" compares current effective names: a Clone copying Courser of
// Kruphix makes both the original and the copy illegal targets, and once the
// original leaves, the copy is a legal target whose activation creates a token
// copy. Opponents' permanents and a face-down Courser card don't count.
func TestYennaTargetComparesCopiedPermanentEffectiveName(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	courserDef := nameUniquenessCard(t, registry, nameUniquenessCourserOracleID, "Courser of Kruphix")
	cloneDef := nameUniquenessCard(t, registry, nameUniquenessCloneOracleID, "Clone")
	yennaDef := nameUniquenessCard(t, registry, nameUniquenessYennaOracleID, "Yenna, Redtooth Regent")

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	courser := addCombatPermanent(g, game.Player1, courserDef)
	clone := resolveCloneCopying(t, g, e, cloneDef, game.Player1, "Courser of Kruphix")
	if !permanentHasType(g, clone, types.Enchantment) {
		t.Fatal("Clone-as-Courser is not an enchantment")
	}
	yenna := addCombatPermanent(g, game.Player1, yennaDef)
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhasePrecombatMain, game.StepNone, game.Player1
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	activate := func(target *game.Permanent) action.Action {
		return action.ActivateAbility(yenna.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)
	}
	legal := e.legalActions(g, game.Player1)
	if containsAction(legal, activate(courser)) {
		t.Fatal("Yenna may target Courser of Kruphix while Clone-as-Courser shares its name")
	}
	if containsAction(legal, activate(clone)) {
		t.Fatal("Yenna may target Clone-as-Courser while the original Courser shares its name")
	}

	if !movePermanentToZone(g, courser, zone.Exile) {
		t.Fatal("could not exile the original Courser of Kruphix")
	}
	addCombatPermanent(g, game.Player2, courserDef)
	addFaceDownPermanent(g, game.Player1, courserDef, game.FaceDownManifest)
	act := activate(clone)
	if !containsAction(e.legalActions(g, game.Player1), act) {
		t.Fatal("Yenna cannot target the only Courser-named permanent its controller controls")
	}
	if !e.applyAction(g, game.Player1, act) {
		t.Fatal("Yenna activation targeting Clone-as-Courser failed")
	}
	before := len(g.Battlefield)
	e.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != before+1 {
		t.Fatalf("battlefield grew by %d, want one token copy", len(g.Battlefield)-before)
	}
	token := g.Battlefield[len(g.Battlefield)-1]
	if !token.Token || token.Controller != game.Player1 || permanentEffectiveName(g, token) != "Courser of Kruphix" {
		t.Fatalf("Yenna created %q (token %v), want a Courser of Kruphix token copy", permanentEffectiveName(g, token), token.Token)
	}
	subject := selectionSubject{kind: subjectPermanent, g: g, permanent: clone}
	if subject.nameUniqueAmongControlled() {
		t.Fatal("Clone-as-Courser is name-unique beside its Courser token copy")
	}
}

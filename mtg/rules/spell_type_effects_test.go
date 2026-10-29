package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned corpus printings Leyline of Transformation
// 4bd941ca-f3d2-44c1-8df3-851362f6b848 (three rulings, 2024-09-20), Arcane
// Adaptation bf3edaaf-cf63-4e17-94ae-9d9991d9fb5f (two rulings, 2017-09-29 and
// 2018-01-19), Maskwood Nexus 1246c42d-57c0-4cba-959a-15ad89d8a50b (six
// rulings, 2021-02-05), Biotransference 75a787a3-5838-4121-bbaf-745d22c4b0bc
// (one ruling, 2022-10-07) and Encroaching Mycosynth
// 65a2fcc9-2317-48a1-a5eb-234fb3300364 (two rulings, 2023-02-04; "Permanent
// spells you control become artifact spells in addition to their other
// types"), each from https://api.scryfall.com/cards/<id>/rulings. Scenario
// cards Reflections of Littjara, Gravel Slinger, Urza's Incubator, Grizzly
// Bears, Vanquisher's Banner, Lightning Bolt and Annul use their pinned Oracle
// text. Mana scenarios use pinned Mishra's Workshop
// aac0c8df-f01d-4178-8d66-ee603f814d24 (four rulings, 2004-10-04), Cavern of
// Souls 3aad15a2-8a1b-4460-9b06-e85863081878 (four rulings, 2018-12-07 and
// 2023-11-10), Path of Ancestry b1aaa7b0-1cac-4a92-b880-7ef1ac00618f (seven
// rulings, 2020-11-10 and 2023-07-28; the commander's types are checked right
// after the spell is cast), Universal Automaton, Marwyn, the Nurturer, Mistcutter
// Hydra 3c844ed7-8ed2-4a39-9134-e14e476ab0c4 ({X}{G}) and Farmer Cotton
// 25b49eaa-13e7-4f4e-b0dc-56ebd28a4a22 ({X}{G}{W}), with
// CR 106.6 (mana spending restrictions and spend triggers), 601.2h (paying the
// total cost) and 702.73a (Changeling). CR 2026-09-25 (official MagicCompRules 20260925.txt, sha256
// 8d860e45...070ca) read: 611.3a-b (a static ability's effect is not locked in
// and applies while its source is on the battlefield), 613.1/613.1d/613.7/613.7a
// (type-changing effects in layer 4 by timestamp), 604.3 (CDAs function in
// every zone), 601.2f/601.2i/601.3 (cost determination, cast triggers and cast
// legality read the spell's characteristics), 702.103b (a bestowed spell is an
// Aura spell, not a creature spell), 708.2/708.2a (a face-down spell is a 2/2
// creature spell with no subtypes) and 707.2/707.10 (a copy's copiable values
// exclude continuous effects; its own controller's effects then apply).

func spellTypeScenario(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return g, NewEngine(nil)
}

func addChosenTypePermanent(t *testing.T, g *game.Game, controller game.PlayerID, name string, chosen types.Sub) *game.Permanent {
	t.Helper()
	permanent := addCombatPermanent(g, controller, productionChoiceCard(t, name))
	permanent.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: chosen},
	}
	return permanent
}

func castSpellTypeTestCard(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, name string, targets []game.Target, pool ...mana.Color) *game.StackObject {
	t.Helper()
	cardID := addCardToHand(g, player, productionChoiceCard(t, name))
	for _, c := range pool {
		g.Players[player].ManaPool.Add(c, 1)
	}
	if !engine.applyAction(g, player, action.CastSpell(cardID, targets, 0, nil)) {
		t.Fatalf("%s cast failed", name)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID {
		t.Fatalf("%s is not on top of the stack", name)
	}
	return obj
}

func spellCharacteristics(t *testing.T, g *game.Game, obj *game.StackObject) *game.CardDef {
	t.Helper()
	def, ok := stackSpellCharacteristicsDef(g, obj)
	if !ok {
		t.Fatal("stack object has no spell characteristics")
	}
	return def
}

func removePermanentForSpellTypeTest(t *testing.T, g *game.Game, engine *Engine, permanent *game.Permanent) {
	t.Helper()
	addEffectSpellToStack(g, game.Player2, game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(permanent.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("setup effect did not remove the permanent")
	}
}

// TestLeylineOfTransformationCreatureSpellTriggersReflectionsAndCopyKeepsType
// is the #3069 acceptance scenario: with Leyline of Transformation (Goblin) and
// Reflections of Littjara (Goblin), casting Gravel Slinger (Human Soldier) is
// casting a Goblin spell, so Reflections copies it, and the copy, also a
// creature spell its controller controls, is a Goblin spell too.
func TestLeylineOfTransformationCreatureSpellTriggersReflectionsAndCopyKeepsType(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Goblin)
	addChosenTypePermanent(t, g, game.Player1, "Reflections of Littjara", types.Goblin)

	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Gravel Slinger", nil, mana.W, mana.C, mana.C, mana.C)
	cast := lastSpellCastEvent(t, g)
	if !slices.Contains(cast.CardSubtypes, types.Goblin) || !slices.Contains(cast.CardSubtypes, types.Human) {
		t.Fatalf("cast event subtypes = %v, want Goblin added to Human Soldier", cast.CardSubtypes)
	}
	if def := spellCharacteristics(t, g, spell); !def.HasSubtype(types.Goblin) || !def.HasSubtype(types.Soldier) {
		t.Fatalf("spell subtypes = %v, want Goblin added", def.Subtypes)
	}

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Reflections of Littjara did not trigger for the Goblin spell")
	}
	// Clone mid-stack: the clone sees the same live type line and resolves
	// the same copy.
	clone := g.Clone()
	for _, state := range []*game.Game{g, clone} {
		engine.resolveTopOfStack(state, &TurnLog{})
		copyObj, ok := state.Stack.Peek()
		if !ok || !copyObj.Copy || copyObj.Controller != game.Player1 {
			t.Fatalf("top of stack = %+v, want Player1's copy of Gravel Slinger", copyObj)
		}
		if def := spellCharacteristics(t, state, copyObj); !def.HasSubtype(types.Goblin) {
			t.Fatalf("copy subtypes = %v, want Goblin", def.Subtypes)
		}
		assertEvent(t, state.Events, game.EventSpellCopied, func(event game.Event) bool {
			return slices.Contains(event.CardSubtypes, types.Goblin)
		})
		if engine.putTriggeredAbilitiesOnStackWithChoices(state, simPassPolicies(), &TurnLog{}) {
			t.Fatal("the uncast copy triggered Reflections of Littjara")
		}
	}
}

// TestReflectionsIgnoresNonGoblinSpellWithoutLeyline proves the trigger above
// comes from Leyline: without it Gravel Slinger is not a Goblin spell.
func TestReflectionsIgnoresNonGoblinSpellWithoutLeyline(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Reflections of Littjara", types.Goblin)
	castSpellTypeTestCard(t, g, engine, game.Player1, "Gravel Slinger", nil, mana.W, mana.C, mana.C, mana.C)
	if slices.Contains(lastSpellCastEvent(t, g).CardSubtypes, types.Goblin) {
		t.Fatal("Gravel Slinger was cast as a Goblin spell without Leyline")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Reflections of Littjara triggered for a non-Goblin spell")
	}
}

// TestLeylineOfTransformationDoesNotAffectOpponentSpells proves "creature spells
// you control" is scoped to the effect controller: Player2's Gravel Slinger is
// not a Goblin while Player1 controls Leyline, so Player2's Reflections does not
// trigger.
func TestLeylineOfTransformationDoesNotAffectOpponentSpells(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Goblin)
	addChosenTypePermanent(t, g, game.Player2, "Reflections of Littjara", types.Goblin)
	setMainPhasePriority(g, game.Player2)

	spell := castSpellTypeTestCard(t, g, engine, game.Player2, "Gravel Slinger", nil, mana.W, mana.C, mana.C, mana.C)
	if spellCharacteristics(t, g, spell).HasSubtype(types.Goblin) {
		t.Fatal("Player1's Leyline made Player2's spell a Goblin")
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Player2's Reflections triggered for a non-Goblin spell")
	}
}

// TestLeylineOfTransformationSpellTypeIsNotLockedIn proves the effect follows its
// source (CR 611.3a-b): once Leyline leaves, the spell already on the stack and
// later spells are no longer Goblins.
func TestLeylineOfTransformationSpellTypeIsNotLockedIn(t *testing.T) {
	g, engine := spellTypeScenario(t)
	leyline := addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Goblin)
	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	if !spellCharacteristics(t, g, spell).HasSubtype(types.Goblin) {
		t.Fatal("Grizzly Bears spell is not a Goblin under Leyline")
	}
	removePermanentForSpellTypeTest(t, g, engine, leyline)
	if spellCharacteristics(t, g, spell).HasSubtype(types.Goblin) {
		t.Fatal("Grizzly Bears spell stayed a Goblin after Leyline left")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("Grizzly Bears did not resolve")
	}
	castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	if slices.Contains(lastSpellCastEvent(t, g).CardSubtypes, types.Goblin) {
		t.Fatal("a later spell was cast as a Goblin after Leyline left")
	}
}

// TestArcaneAdaptationChosenTypeReducesCreatureSpellCost proves the cost
// determination (CR 601.2f) reads the spell's effective subtypes: with Arcane
// Adaptation and Urza's Incubator both naming Elf, Grizzly Bears ({1}{G}) costs
// {G}; without Arcane Adaptation {G} is not enough.
func TestArcaneAdaptationChosenTypeReducesCreatureSpellCost(t *testing.T) {
	for _, withAdaptation := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		addChosenTypePermanent(t, g, game.Player1, "Urza's Incubator", types.Elf)
		if withAdaptation {
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Elf)
		}
		cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		cast := engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil))
		if cast != withAdaptation {
			t.Fatalf("with Arcane Adaptation=%v: casting Grizzly Bears for {G} = %v", withAdaptation, cast)
		}
		if cast && g.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("reduced cast did not spend the {G}")
		}
	}
}

// TestArcaneAdaptationDoesNotAffectNoncreatureSpells proves the spell clause is
// limited to creature spells.
func TestArcaneAdaptationDoesNotAffectNoncreatureSpells(t *testing.T) {
	g, _ := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Elf)
	bolt := productionChoiceCard(t, "Lightning Bolt")
	if line := spellTypeLineFor(g, game.Player1, bolt, false); slices.Contains(line.subtypes, types.Elf) {
		t.Fatalf("Lightning Bolt spell subtypes = %v, want no Elf", line.subtypes)
	}
	bears := productionChoiceCard(t, "Grizzly Bears")
	if line := spellTypeLineFor(g, game.Player1, bears, false); !slices.Contains(line.subtypes, types.Elf) {
		t.Fatalf("Grizzly Bears spell subtypes = %v, want Elf", line.subtypes)
	}
}

// TestBiotransferenceCreatureSpellIsArtifactSpell proves a creature spell is an
// artifact spell under Biotransference, so its own "whenever you cast an
// artifact spell" trigger fires; an instant does not.
func TestBiotransferenceCreatureSpellIsArtifactSpell(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	lifeBefore := g.Players[game.Player1].Life

	castSpellTypeTestCard(t, g, engine, game.Player1, "Lightning Bolt", []game.Target{game.PlayerTarget(game.Player2)}, mana.R)
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Biotransference triggered for an instant spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	if def := spellCharacteristics(t, g, spell); !def.HasType(types.Artifact) || !def.HasType(types.Creature) {
		t.Fatalf("Grizzly Bears spell types = %v, want artifact creature", def.Types)
	}
	if !slices.Contains(lastSpellCastEvent(t, g).CardTypes, types.Artifact) {
		t.Fatal("cast event does not carry the artifact type")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Biotransference did not trigger for the artifact creature spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != lifeBefore-1 {
		t.Fatalf("life = %d, want %d", got, lifeBefore-1)
	}
	tokens := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == game.Player1 && permanent.TokenDef != nil && permanent.TokenDef.Name == "Necron Warrior" {
			tokens++
		}
	}
	if tokens != 1 {
		t.Fatalf("Necron Warrior tokens = %d, want 1", tokens)
	}
}

// TestMaskwoodNexusCreatureSpellTriggersChosenTypeCastTrigger proves Maskwood
// Nexus makes a creature spell every creature type, so Vanquisher's Banner
// (Dragon) draws a card for Grizzly Bears.
func TestMaskwoodNexusCreatureSpellTriggersChosenTypeCastTrigger(t *testing.T) {
	for _, withNexus := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		if withNexus {
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Maskwood Nexus"))
		}
		addChosenTypePermanent(t, g, game.Player1, "Vanquisher's Banner", types.Dragon)
		addCardToLibraryNamed(g, game.Player1, "Library card")
		castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
		handBefore := g.Players[game.Player1].Hand.Size()
		triggered := engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
		if triggered != withNexus {
			t.Fatalf("with Maskwood Nexus=%v: Vanquisher's Banner triggered = %v", withNexus, triggered)
		}
		if !triggered {
			continue
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player1].Hand.Size(); got != handBefore+1 {
			t.Fatalf("hand size = %d, want %d", got, handBefore+1)
		}
	}
}

// TestEncroachingMycosynthPermanentSpellIsArtifactTarget proves targeting reads
// the effective type: Annul ("Counter target artifact or enchantment spell")
// can target Player1's Grizzly Bears spell under Player1's Encroaching
// Mycosynth but not Player2's Grizzly Bears spell.
func TestEncroachingMycosynthPermanentSpellIsArtifactTarget(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Encroaching Mycosynth"))
	ownSpell := castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.G, mana.C)
	if !spellCharacteristics(t, g, ownSpell).HasType(types.Artifact) {
		t.Fatal("Player1's Grizzly Bears spell is not an artifact under Mycosynth")
	}

	g.Turn.PriorityPlayer = game.Player2
	annul := castSpellTypeTestCard(t, g, engine, game.Player2, "Annul", []game.Target{game.StackObjectTarget(ownSpell.ID)}, mana.U)
	if len(annul.Targets) != 1 || annul.Targets[0] != game.StackObjectTarget(ownSpell.ID) {
		t.Fatalf("Annul targets = %v", annul.Targets)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("Annul did not counter the artifact creature spell")
	}

	setMainPhasePriority(g, game.Player2)
	opponentSpell := castSpellTypeTestCard(t, g, engine, game.Player2, "Grizzly Bears", nil, mana.G, mana.C)
	if spellCharacteristics(t, g, opponentSpell).HasType(types.Artifact) {
		t.Fatal("Player1's Mycosynth made Player2's spell an artifact")
	}
	g.Turn.PriorityPlayer = game.Player1
	annulID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Annul"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	if engine.applyAction(g, game.Player1, action.CastSpell(annulID, []game.Target{game.StackObjectTarget(opponentSpell.ID)}, 0, nil)) {
		t.Fatal("Annul targeted Player2's nonartifact creature spell")
	}
}

// TestSpellTypeEffectsOnBestowedSpell proves the selection matches the spell's
// characteristics after the bestow transform (CR 702.103b): a bestowed spell is
// an Aura enchantment spell, so Leyline's creature-spell clause does not apply
// but Mycosynth's permanent-spell clause does.
func TestSpellTypeEffectsOnBestowedSpell(t *testing.T) {
	g, engine := setupBestowMain(t)
	addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Goblin)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Encroaching Mycosynth"))
	obj, _ := castBestowTester(t, g, engine)
	def := spellCharacteristics(t, g, obj)
	if def.HasType(types.Creature) || !def.HasSubtype(types.Aura) {
		t.Fatalf("bestowed spell = %v %v, want noncreature Aura", def.Types, def.Subtypes)
	}
	if def.HasSubtype(types.Goblin) {
		t.Fatal("Leyline's creature-spell clause applied to a bestowed Aura spell")
	}
	if !def.HasType(types.Artifact) {
		t.Fatal("Mycosynth's permanent-spell clause did not apply to a bestowed Aura spell")
	}
	if cast := lastSpellCastEvent(t, g); slices.Contains(cast.CardSubtypes, types.Goblin) || !slices.Contains(cast.CardTypes, types.Artifact) {
		t.Fatalf("bestowed cast event = %v %v", cast.CardTypes, cast.CardSubtypes)
	}
}

// TestLeylineOfTransformationFaceDownSpellGetsChosenType proves a face-down
// spell (a 2/2 creature spell with no subtypes, CR 708.2a) is still a creature
// spell its controller controls, so it becomes a Goblin, and the cast event
// carries that without revealing the hidden card's subtypes.
func TestLeylineOfTransformationFaceDownSpellGetsChosenType(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Goblin)
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Gravel Slinger"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	obj := castFaceDownPaying(t, engine, g, cardID, game.FaceDownMorph, 3)
	def := spellCharacteristics(t, g, obj)
	if !def.HasSubtype(types.Goblin) || def.HasSubtype(types.Human) || def.HasSubtype(types.Soldier) {
		t.Fatalf("face-down spell subtypes = %v, want only Goblin", def.Subtypes)
	}
	cast := lastSpellCastEvent(t, g)
	if !slices.Equal(cast.CardSubtypes, []types.Sub{types.Goblin}) {
		t.Fatalf("face-down cast event subtypes = %v, want [Goblin]", cast.CardSubtypes)
	}
}

// TestBiotransferenceAppliesToSpellCastWithoutPayingFromExile covers the
// free-cast route: a creature card cast from exile without paying its mana cost
// is an artifact spell its controller controls and triggers Biotransference.
func TestBiotransferenceAppliesToSpellCastWithoutPayingFromExile(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: productionChoiceCard(t, "Grizzly Bears"), Owner: game.Player1}
	g.Players[game.Player1].Exile.Add(cardID)

	if !engine.castFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Exile, false, simPassPolicies(), &TurnLog{}) {
		t.Fatal("free cast from exile failed")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.SourceID != cardID || !spellCharacteristics(t, g, spell).HasType(types.Artifact) {
		t.Fatal("free-cast Grizzly Bears is not an artifact spell on the stack")
	}
	if !slices.Contains(lastSpellCastEvent(t, g).CardTypes, types.Artifact) {
		t.Fatal("free-cast event does not carry the artifact type")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Biotransference did not trigger for the free-cast artifact creature spell")
	}
}

func activateManaAbilityForSpellTypeTest(t *testing.T, g *game.Game, engine *Engine, permanent *game.Permanent, name string, index int) {
	t.Helper()
	def := productionChoiceCard(t, name)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(permanent.ObjectID, def.ManaAbilityIndex(index), nil, 0)) {
		t.Fatalf("%s mana ability failed", name)
	}
}

// TestRestrictedManaPaysForEffectiveSpellType proves restricted mana (CR
// 106.6) is tested against the spell's effective characteristics while its
// costs are paid (CR 601.2h), in both the legal-action listing and the cast:
// Mishra's Workshop mana pays for a creature spell Biotransference makes an
// artifact spell, Cavern of Souls (Goblin) mana pays for a creature spell
// Leyline of Transformation makes a Goblin and that spell can't be countered,
// and Cavern mana pays for a Changeling spell, which is every creature type by
// its characteristic-defining ability (CR 702.73a, 604.3). Without the
// type-changing permanent the restricted mana cannot pay.
func TestRestrictedManaPaysForEffectiveSpellType(t *testing.T) {
	cases := []struct {
		name         string
		typeChanger  string
		manaSource   string
		manaAbility  int
		spell        string
		extra        []mana.Color
		uncounterble bool
		alwaysLegal  bool
	}{
		{name: "Biotransference with Mishra's Workshop", typeChanger: "Biotransference", manaSource: "Mishra's Workshop", spell: "Grizzly Bears", extra: []mana.Color{mana.G}},
		{name: "Leyline with Cavern of Souls", typeChanger: "Leyline of Transformation", manaSource: "Cavern of Souls", manaAbility: 1, spell: "Gravel Slinger", extra: []mana.Color{mana.C, mana.C, mana.C}, uncounterble: true},
		{name: "Changeling with Cavern of Souls", manaSource: "Cavern of Souls", manaAbility: 1, spell: "Universal Automaton", uncounterble: true, alwaysLegal: true},
	}
	for _, tc := range cases {
		for _, withChanger := range []bool{false, true} {
			if tc.alwaysLegal && !withChanger {
				continue
			}
			t.Run(tc.name+map[bool]string{false: " absent", true: " present"}[withChanger], func(t *testing.T) {
				g, engine := spellTypeScenario(t)
				if withChanger && tc.typeChanger != "" {
					addChosenTypePermanent(t, g, game.Player1, tc.typeChanger, types.Goblin)
				}
				source := addChosenTypePermanent(t, g, game.Player1, tc.manaSource, types.Goblin)
				activateManaAbilityForSpellTypeTest(t, g, engine, source, tc.manaSource, tc.manaAbility)
				if len(g.Players[game.Player1].ManaRiders) == 0 {
					t.Fatalf("%s produced no restricted mana", tc.manaSource)
				}
				for _, c := range tc.extra {
					g.Players[game.Player1].ManaPool.Add(c, 1)
				}
				cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, tc.spell))
				cast := action.CastSpell(cardID, nil, 0, nil)
				listed := containsAction(engine.legalActions(g, game.Player1), cast)
				applied := engine.applyAction(g, game.Player1, cast)
				if listed != withChanger || applied != withChanger {
					t.Fatalf("cast listed=%v applied=%v, want %v", listed, applied, withChanger)
				}
				if !applied {
					return
				}
				obj, ok := g.Stack.Peek()
				if !ok || obj.SourceID != cardID {
					t.Fatalf("%s is not on top of the stack", tc.spell)
				}
				if tc.uncounterble && stackSpellCanBeCountered(g, obj) {
					t.Fatal("spell cast with Cavern of Souls mana can be countered")
				}
			})
		}
	}
}

// TestRestrictedManaBoundsXForEffectiveSpellType proves the X values offered
// for a cast (CR 601.2b, 107.3) are bounded by what restricted mana (CR 106.6)
// can pay for the spell's effective characteristics (CR 601.2h), exactly as the
// cast pays: Mishra's Workshop's three mana pays X=3 for Mistcutter Hydra made
// an artifact spell by Biotransference, and Cavern of Souls (Elf) mana pays the
// white symbol of Farmer Cotton made an Elf by Arcane Adaptation. Without the
// type-changing permanent neither X cast is listed or castable.
func TestRestrictedManaBoundsXForEffectiveSpellType(t *testing.T) {
	cases := []struct {
		name        string
		typeChanger string
		manaSource  string
		manaAbility int
		spell       string
		x           int
	}{
		{name: "Biotransference with Mishra's Workshop", typeChanger: "Biotransference", manaSource: "Mishra's Workshop", spell: "Mistcutter Hydra", x: 3},
		{name: "Arcane Adaptation with Cavern of Souls", typeChanger: "Arcane Adaptation", manaSource: "Cavern of Souls", manaAbility: 1, spell: "Farmer Cotton", x: 0},
	}
	for _, tc := range cases {
		for _, withChanger := range []bool{false, true} {
			t.Run(tc.name+map[bool]string{false: " absent", true: " present"}[withChanger], func(t *testing.T) {
				g, engine := spellTypeScenario(t)
				if withChanger {
					addChosenTypePermanent(t, g, game.Player1, tc.typeChanger, types.Elf)
				}
				source := addChosenTypePermanent(t, g, game.Player1, tc.manaSource, types.Elf)
				activateManaAbilityForSpellTypeTest(t, g, engine, source, tc.manaSource, tc.manaAbility)
				if len(g.Players[game.Player1].ManaRiders) == 0 {
					t.Fatalf("%s produced no restricted mana", tc.manaSource)
				}
				g.Players[game.Player1].ManaPool.Add(mana.G, 1)
				cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, tc.spell))
				cast := action.CastSpell(cardID, nil, tc.x, nil)
				listed := containsAction(engine.legalActions(g, game.Player1), cast)
				applied := engine.applyAction(g, game.Player1, cast)
				if listed != withChanger || applied != withChanger {
					t.Fatalf("X=%d cast listed=%v applied=%v, want %v", tc.x, listed, applied, withChanger)
				}
				if !applied {
					return
				}
				obj, ok := g.Stack.Peek()
				if !ok || obj.SourceID != cardID || obj.XValue != tc.x {
					t.Fatalf("%s with X=%d is not on top of the stack", tc.spell, tc.x)
				}
				if player := g.Players[game.Player1]; len(player.ManaRiders) != 0 || player.ManaPool.Total() != 0 {
					t.Fatalf("restricted mana not spent: riders=%d pool=%d", len(player.ManaRiders), player.ManaPool.Total())
				}
			})
		}
	}
}

// TestBestowedSpellCannotSpendCreatureOnlyMana proves the planner reads a
// bestowed spell's effective characteristics (CR 702.103b): mana spendable only
// on creature spells cannot pay its bestow cost, while it still pays the same
// card's normal creature-spell cost.
func TestBestowedSpellCannotSpendCreatureOnlyMana(t *testing.T) {
	g, engine := setupBestowMain(t)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.G, 1)
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: mana.G},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:   game.ManaSpendCastCreatureSpell,
			Restriction: game.ManaSpendRestrictedToCondition,
		},
	})
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	spellID := addCardToHand(g, game.Player1, bestowTestCreature())
	bestow := action.CastBestowSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if containsAction(engine.legalActions(g, game.Player1), bestow) || engine.applyAction(g, game.Player1, bestow) {
		t.Fatal("creature-only mana paid for a bestowed Aura spell")
	}
	player.ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("creature-only mana could not pay for the normally cast creature spell")
	}
	if player.ManaPool.Amount(mana.G) != 0 || len(player.ManaRiders) != 0 {
		t.Fatalf("creature-only mana not spent: pool G=%d riders=%d", player.ManaPool.Amount(mana.G), len(player.ManaRiders))
	}
}

// TestManaSpendTriggerSeesEffectiveSpellType proves a mana-spend trigger (CR
// 106.6) reads the cast spell's effective characteristics: Path of Ancestry
// mana spent on Grizzly Bears, made an Elf by Arcane Adaptation, shares a
// creature type with an Elf commander and fires the scry; without Arcane
// Adaptation the Bear does not.
func TestManaSpendTriggerSeesEffectiveSpellType(t *testing.T) {
	for _, withAdaptation := range []bool{false, true} {
		t.Run(map[bool]string{false: "without Arcane Adaptation", true: "with Arcane Adaptation"}[withAdaptation], func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			addCommanderToCommandZone(g, game.Player1, productionChoiceCard(t, "Marwyn, the Nurturer"))
			if withAdaptation {
				addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Elf)
			}
			path := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Path of Ancestry"))
			activateManaAbilityForSpellTypeTest(t, g, engine, path, "Path of Ancestry", 0)
			if riders := g.Players[game.Player1].ManaRiders; len(riders) != 1 || riders[0].Unit.Color != mana.G {
				t.Fatalf("Path of Ancestry mana = %v, want one green rider", riders)
			}
			castSpellTypeTestCard(t, g, engine, game.Player1, "Grizzly Bears", nil, mana.C)
			if fired := firedSpendRiderCount(g); (fired == 1) != withAdaptation {
				t.Fatalf("scry triggers = %d, want fired=%v", fired, withAdaptation)
			}
		})
	}
}

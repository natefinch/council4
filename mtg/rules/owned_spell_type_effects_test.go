package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3530 (apply owned-card type effects to
// opponent-controlled creature spells). Pinned corpus Oracle text: Arcane
// Adaptation (bf3edaaf-cf63-4e17-94ae-9d9991d9fb5f; rulings 2017-09-29 and
// 2018-01-19 from https://api.scryfall.com/cards/bf3edaaf-cf63-4e17-94ae-9d9991d9fb5f/rulings)
// and Conspiracy (rulings 2004-10-04 through 2017-09-29; "Affected spells,
// cards, and creatures lose all of their other creature types"): "Creatures you
// control are the chosen type [in addition to their other types]. The same is
// true for creature spells you control and creature cards you own that aren't
// on the battlefield." Scenario cards Court of Locthwain, Raging Goblin (Goblin
// Berserker), Grizzly Bears, Reflections of Littjara, Urza's Incubator, Cavern
// of Souls and Aethersnatch use their pinned Oracle text. CR 2026-09-25
// (official MagicCompRules 20260925.txt): 108.3 (owner of a card), 112.1 (a
// spell is a card on the stack), 112.1a and 707.10 (a copy of a spell is a
// spell with no card; it is owned and controlled by the player who put it on
// the stack), 601.2a (the card becomes a spell as it is moved to the stack, and
// continuous effects that modify it begin then) and 613.7/613.7a (effects in
// one layer apply in timestamp order). So an owner's "cards you own that aren't
// on the battlefield" clause applies to their card while an opponent casts and
// controls it, independently of the controller's "spells you control" clause,
// and not to a copy.

// castOpponentCourtCard has Player1 cast Player2's def from Player2's exile
// through Court of Locthwain. When free, Player1 is the monarch and casts it
// without paying its mana cost; otherwise Player1 pays with pool.
func castOpponentCourtCard(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, free bool, pool ...mana.Color) (id.ID, *game.StackObject) {
	t.Helper()
	g.Players[game.Player1].IsMonarch = free
	cardID := driveLocthwainUpkeepWithCard(t, engine, g, def)
	setSorcerySpeedTurn(g, game.Player1)
	for _, c := range pool {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("Player1 could not cast Player2's %s from exile", def.Name)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Controller != game.Player1 || obj.Copy {
		t.Fatalf("top of stack = %+v, want Player1's spell of Player2's %s", obj, def.Name)
	}
	return cardID, obj
}

func assertSpellSubtypes(t *testing.T, what string, got []types.Sub, want ...types.Sub) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("%s subtypes = %v, want %v", what, got, want)
	}
}

// TestArcaneAdaptationAppliesToOwnedCardOpponentCasts is the #3530
// reproduction: Player2 controls Arcane Adaptation (Elf) and Player1 casts
// Player2's Raging Goblin through Court of Locthwain. The spell is Player2's
// card, so it is an Elf in addition to Goblin Berserker, on the stack and in
// its cast event, which triggers Player1's Reflections of Littjara (Elf). The
// copy Reflections creates is not a card, so Player2's clause doesn't reach it.
// A clone sees the same live type line, and once Arcane Adaptation leaves the
// spell is no longer an Elf.
func TestArcaneAdaptationAppliesToOwnedCardOpponentCasts(t *testing.T) {
	g, engine := spellTypeScenario(t)
	adaptation := addChosenTypePermanent(t, g, game.Player2, "Arcane Adaptation", types.Elf)
	addChosenTypePermanent(t, g, game.Player1, "Reflections of Littjara", types.Elf)
	_, spell := castOpponentCourtCard(t, g, engine, productionChoiceCard(t, "Raging Goblin"), true)

	assertSpellSubtypes(t, "Player2's Raging Goblin cast by Player1", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker, types.Elf)
	assertSpellSubtypes(t, "cast event", lastSpellCastEvent(t, g).CardSubtypes, types.Goblin, types.Berserker, types.Elf)
	assertSpellSubtypes(t, "cloned spell", spellCharacteristics(t, g.Clone(), spell).Subtypes, types.Goblin, types.Berserker, types.Elf)

	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Reflections of Littjara did not trigger for the Elf spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy || copyObj.Controller != game.Player1 {
		t.Fatalf("top of stack = %+v, want Player1's copy", copyObj)
	}
	assertSpellSubtypes(t, "uncarded copy", spellCharacteristics(t, g, copyObj).Subtypes, types.Goblin, types.Berserker)
	g.Stack.RemoveByID(copyObj.ID)

	removePermanentForSpellTypeTest(t, g, engine, adaptation)
	assertSpellSubtypes(t, "spell after Arcane Adaptation left", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker)
}

// TestConspiracyOwnerAndControllerClausesApplyInTimestampOrder proves the
// owner's and the controller's clauses are ordered together by timestamp (CR
// 613.7): with Player2's Conspiracy (Elf) alone Player2's Raging Goblin cast by
// Player1 is only an Elf; with Player1's Conspiracy (Warrior) as well, the later
// Conspiracy wins, and when it leaves the earlier one applies again.
func TestConspiracyOwnerAndControllerClausesApplyInTimestampOrder(t *testing.T) {
	t.Run("owner only", func(t *testing.T) {
		g, engine := spellTypeScenario(t)
		addConspiracy(t, g, game.Player2, types.Elf)
		_, spell := castOpponentCourtCard(t, g, engine, productionChoiceCard(t, "Raging Goblin"), true)
		assertSpellSubtypes(t, "spell", spellCharacteristics(t, g, spell).Subtypes, types.Elf)
		assertSpellSubtypes(t, "cast event", lastSpellCastEvent(t, g).CardSubtypes, types.Elf)
	})
	for _, ownerFirst := range []bool{true, false} {
		name := map[bool]string{true: "owner earlier", false: "controller earlier"}[ownerFirst]
		t.Run(name, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			var owner, controller *game.Permanent
			if ownerFirst {
				owner = addConspiracy(t, g, game.Player2, types.Elf)
				controller = addConspiracy(t, g, game.Player1, types.Warrior)
			} else {
				controller = addConspiracy(t, g, game.Player1, types.Warrior)
				owner = addConspiracy(t, g, game.Player2, types.Elf)
			}
			_, spell := castOpponentCourtCard(t, g, engine, productionChoiceCard(t, "Raging Goblin"), true)
			later, earlier, laterType, earlierType := controller, owner, types.Warrior, types.Elf
			if !ownerFirst {
				later, earlier, laterType, earlierType = owner, controller, types.Elf, types.Warrior
			}
			assertSpellSubtypes(t, "spell", spellCharacteristics(t, g, spell).Subtypes, laterType)
			assertSpellSubtypes(t, "cast event", lastSpellCastEvent(t, g).CardSubtypes, laterType)
			removePermanentForSpellTypeTest(t, g, engine, later)
			assertSpellSubtypes(t, "spell after the later Conspiracy left", spellCharacteristics(t, g, spell).Subtypes, earlierType)
			removePermanentForSpellTypeTest(t, g, engine, earlier)
			assertSpellSubtypes(t, "spell after both left", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker)
		})
	}
}

// TestArcaneAdaptationReachesOwnSpellOnce proves an effect that reaches a spell
// through both clauses (its controller owns it) applies once.
func TestArcaneAdaptationReachesOwnSpellOnce(t *testing.T) {
	g, engine := spellTypeScenario(t)
	addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Elf)
	spell := castSpellTypeTestCard(t, g, engine, game.Player1, "Raging Goblin", nil, mana.R)
	assertSpellSubtypes(t, "own spell", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker, types.Elf)
	assertSpellSubtypes(t, "cast event", lastSpellCastEvent(t, g).CardSubtypes, types.Goblin, types.Berserker, types.Elf)
}

// TestOwnedCardClauseFollowsStolenSpell proves the clauses track the spell's
// current controller and its fixed owner: after Player1 gains control of
// Player2's Raging Goblin spell with Aethersnatch, Player2's Arcane Adaptation
// (Elf) still applies through ownership and Player1's (Warrior) now applies
// through control.
func TestOwnedCardClauseFollowsStolenSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := stolenCasterEngine()
	addChosenTypePermanent(t, g, game.Player2, "Arcane Adaptation", types.Elf)
	addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Warrior)
	setMainPhasePriority(g, game.Player2)
	spell := castSpellTypeTestCard(t, g, engine, game.Player2, "Raging Goblin", nil, mana.R)
	assertSpellSubtypes(t, "Player2's own spell", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker, types.Elf)

	snatchSpell(t, engine, g, game.Player1, spell)
	assertSpellSubtypes(t, "stolen spell", spellCharacteristics(t, g, spell).Subtypes, types.Goblin, types.Berserker, types.Elf, types.Warrior)
}

// TestOwnedCardClauseDeterminesOpponentCastCost proves cost determination
// (CR 601.2f) and restricted mana (CR 106.6, 601.2h) read the owner's clause:
// Player1's Urza's Incubator (Elf) reduces Player2's Grizzly Bears to {G} only
// while Player2's Arcane Adaptation (Elf) makes it an Elf, and Player1's Cavern
// of Souls (Elf) mana pays for it only then, making it uncounterable.
func TestOwnedCardClauseDeterminesOpponentCastCost(t *testing.T) {
	for _, withAdaptation := range []bool{true, false} {
		suffix := map[bool]string{true: " with Arcane Adaptation", false: " without Arcane Adaptation"}[withAdaptation]
		t.Run("Urza's Incubator"+suffix, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			if withAdaptation {
				addChosenTypePermanent(t, g, game.Player2, "Arcane Adaptation", types.Elf)
			}
			addChosenTypePermanent(t, g, game.Player1, "Urza's Incubator", types.Elf)
			cardID := driveLocthwainUpkeepWithCard(t, engine, g, productionChoiceCard(t, "Grizzly Bears"))
			setSorcerySpeedTurn(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			listed := containsCastFromExile(engine.legalActions(g, game.Player1), cardID)
			cast := engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil))
			if listed != withAdaptation || cast != withAdaptation {
				t.Fatalf("cast for {G} listed=%v applied=%v, want %v", listed, cast, withAdaptation)
			}
		})
		t.Run("Cavern of Souls"+suffix, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			if withAdaptation {
				addChosenTypePermanent(t, g, game.Player2, "Arcane Adaptation", types.Elf)
			}
			cardID := driveLocthwainUpkeepWithCard(t, engine, g, productionChoiceCard(t, "Grizzly Bears"))
			setSorcerySpeedTurn(g, game.Player1)
			cavern := addChosenTypePermanent(t, g, game.Player1, "Cavern of Souls", types.Elf)
			activateManaAbilityForSpellTypeTest(t, g, engine, cavern, "Cavern of Souls", 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			listed := containsCastFromExile(engine.legalActions(g, game.Player1), cardID)
			cast := engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil))
			if listed != withAdaptation || cast != withAdaptation {
				t.Fatalf("cast with Cavern mana listed=%v applied=%v, want %v", listed, cast, withAdaptation)
			}
			if !cast {
				return
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.SourceID != cardID || stackSpellCanBeCountered(g, obj) {
				t.Fatal("Grizzly Bears cast with Cavern of Souls mana is not an uncounterable spell on the stack")
			}
		})
	}
}

// TestOwnedCardClauseOnResolutionCasts covers the resolution cast routes:
// Player1 casting Player2's exiled Raging Goblin itself, with or without paying
// its mana cost, casts Player2's card, an Elf under Player2's Arcane Adaptation;
// casting a copy of that card (CR 707.12) casts no card, so it is not.
func TestOwnedCardClauseOnResolutionCasts(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *Engine, id.ID) {
		g, engine := spellTypeScenario(t)
		addChosenTypePermanent(t, g, game.Player2, "Arcane Adaptation", types.Elf)
		cardID := g.IDGen.Next()
		g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: productionChoiceCard(t, "Raging Goblin"), Owner: game.Player2}
		g.Players[game.Player2].Exile.Add(cardID)
		return g, engine, cardID
	}
	for _, tc := range []struct {
		name   string
		cast   func(*game.Game, *Engine, id.ID) bool
		copied bool
	}{
		{name: "without paying", cast: func(g *game.Game, e *Engine, cardID id.ID) bool {
			return e.castFreeSpellFromSource(g, g.Players[game.Player2], game.Player1, cardID, zone.Exile, false, simPassPolicies(), &TurnLog{})
		}},
		{name: "paying", cast: func(g *game.Game, e *Engine, cardID id.ID) bool {
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			return e.castPaidSpellFromSource(g, g.Players[game.Player2], game.Player1, cardID, zone.Exile, simPassPolicies(), &TurnLog{})
		}},
		{name: "copy of the card", copied: true, cast: func(g *game.Game, e *Engine, cardID id.ID) bool {
			return e.castFreeCopyOfCard(g, game.Player1, cardID, simPassPolicies(), &TurnLog{})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine, cardID := setup(t)
			if !tc.cast(g, engine, cardID) {
				t.Fatal("cast failed")
			}
			spell, ok := g.Stack.Peek()
			if !ok || spell.Controller != game.Player1 || spell.Copy != tc.copied {
				t.Fatalf("top of stack = %+v, want Player1's spell (copy=%v)", spell, tc.copied)
			}
			want := []types.Sub{types.Goblin, types.Berserker, types.Elf}
			if tc.copied {
				want = want[:2]
			}
			assertSpellSubtypes(t, "spell", spellCharacteristics(t, g, spell).Subtypes, want...)
			assertSpellSubtypes(t, "cast event", lastSpellCastEvent(t, g).CardSubtypes, want...)
		})
	}
}

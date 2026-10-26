package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3345. Default-registry cards checked against the
// pinned corpus: War Behemoth 652109b9-d607-42b6-945d-0c0dd5bba89c (Morph
// {4}{W}; nine 2014-09-20 rulings: {3} alternative cost, colorless 2/2 with no
// name/subtypes/abilities), Leyline of Anticipation
// f57bdaa1-ce8a-4103-8598-fee751e65a53 ("You may cast spells as though they had
// flash."), Vedalken Orrery 44067327-6fe3-4222-b591-6d14d0e360d7 (2004 ruling:
// any nonland card at instant timing), Alley Assailant
// edf238c9-61de-4f3a-b82f-05af46e5e81b (Disguise {4}{B}{B}), Branchsnap Lorian
// 52118ff1-ad76-4b97-9fdc-6adfe80140f8 (green, Morph {G}), Yeva, Nature's
// Herald 67ff9ba3-cd2b-4970-b4da-22104b10a4a4 (green creature spells as though
// flash), Grand Abolisher ee793ed2-7d59-4640-8868-ad486600df2c and Rule of Law
// a1f4e79b-b103-4380-afa0-61a2b1773c9e. CR 2026-09-25 (official text from
// media.wizards.com): 702.37c and 708.4 apply effects and prohibitions to the
// face-down 2/2 creature characteristics, not the face-up card's; 601.3 and
// 601.3b allow as-though-flash casts; 702.8a flash; 702.168a disguise;
// 702.61a split second; 101.2 "can't" beats permission.

func faceDownTimingGame(t *testing.T, active game.PlayerID, hand string, color mana.Color) (*game.Game, *Engine, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, hand))
	g.Players[game.Player1].ManaPool.Add(color, 3)
	g.Turn.ActivePlayer = active
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, engine, cardID
}

func faceDownCastOffered(engine *Engine, g *game.Game, cardID id.ID, kind game.FaceDownKind) bool {
	for _, act := range engine.legalActions(g, game.Player1) {
		if payload, ok := act.CastFaceDownPayload(); ok && payload.CardID == cardID && payload.FaceDownKind == kind &&
			(payload.SourceZone == zone.None || payload.SourceZone == zone.Hand) {
			return true
		}
	}
	return false
}

func assertFaceDownCastRejected(t *testing.T, engine *Engine, g *game.Game, cardID id.ID, kind game.FaceDownKind) {
	t.Helper()
	stackSize := g.Stack.Size()
	if faceDownCastOffered(engine, g, cardID, kind) {
		t.Fatal("face-down cast offered when it should be illegal")
	}
	if engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, kind)) {
		t.Fatal("face-down cast executed when it should be illegal")
	}
	if !g.Players[game.Player1].Hand.Contains(cardID) || g.Stack.Size() != stackSize {
		t.Fatal("rejected face-down cast moved the card")
	}
}

func castFaceDownAndResolve(t *testing.T, engine *Engine, g *game.Game, cardID id.ID, kind game.FaceDownKind) *game.Permanent {
	t.Helper()
	if !faceDownCastOffered(engine, g, cardID, kind) {
		t.Fatal("face-down cast not offered")
	}
	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(cardID, game.FaceFront, kind)) {
		t.Fatal("face-down cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.FaceDown || obj.SourceID != cardID || obj.FaceDownKind != kind {
		t.Fatalf("stack top = %+v, want face-down spell", obj)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana left after face-down cast = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			if !permanent.FaceDown || permanent.FaceDownKind != kind || !faceDownIsTwoTwo(g, permanent) {
				t.Fatalf("resolved permanent = %+v, want face-down 2/2", permanent)
			}
			return permanent
		}
	}
	t.Fatal("face-down spell did not resolve onto the battlefield")
	return nil
}

func TestLeylineOfAnticipationLetsMorphBeCastFaceDownOnOpponentsTurn(t *testing.T) {
	g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Leyline of Anticipation"))

	castFaceDownAndResolve(t, engine, g, behemoth, game.FaceDownMorph)
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == behemoth && event.Controller == game.Player1 && event.FromZone == zone.Hand
	})
}

func TestVedalkenOrreryLetsDisguiseBeCastFaceDownDuringCombat(t *testing.T) {
	g, engine, assailant := faceDownTimingGame(t, game.Player2, "Alley Assailant", mana.B)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Vedalken Orrery"))

	castFaceDownAndResolve(t, engine, g, assailant, game.FaceDownDisguise)
}

func TestMorphCannotBeCastFaceDownOnOpponentsTurnWithoutFlashPermission(t *testing.T) {
	g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)

	// An opponent's Leyline grants the permission only to its controller.
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Leyline of Anticipation"))
	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
}

func TestSplitSecondBlocksFlashPermittedFaceDownCast(t *testing.T) {
	g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Leyline of Anticipation"))
	splitSecondID := g.IDGen.Next()
	g.CardInstances[splitSecondID] = &game.CardInstance{
		ID: splitSecondID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Split Second Spell",
			Types:           []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{game.SplitSecondStaticBody}},
		},
		Owner: game.Player2,
	}
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: splitSecondID, Controller: game.Player2})

	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only the split second spell", g.Stack.Size())
	}
}

// TestFaceDownTimingUsesFaceDownCharacteristics proves CR 702.37c/708.4: the
// hidden card's own flash and green color do not let it be cast face down at
// instant timing, while the face-up cast still uses them.
func TestFaceDownTimingUsesFaceDownCharacteristics(t *testing.T) {
	g, engine, lorian := faceDownTimingGame(t, game.Player2, "Branchsnap Lorian", mana.G)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Yeva, Nature's Herald"))
	lorianDef := g.CardInstances[lorian].Def
	if !canCastAtCurrentTiming(g, game.Player1, lorianDef) {
		t.Fatal("Yeva did not grant the green creature card flash timing face up")
	}
	assertFaceDownCastRejected(t, engine, g, lorian, game.FaceDownMorph)

	flashMorph := morphCreature(cost.Mana{cost.G})
	flashMorph.StaticAbilities = append(flashMorph.StaticAbilities, game.FlashStaticBody)
	flashID := addCardToHand(g, game.Player1, flashMorph)
	if !canCastAtCurrentTiming(g, game.Player1, flashMorph) {
		t.Fatal("flash card could not be cast face up at instant timing")
	}
	assertFaceDownCastRejected(t, engine, g, flashID, game.FaceDownMorph)
}

func TestGrandAbolisherProhibitsFlashPermittedFaceDownCast(t *testing.T) {
	g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Leyline of Anticipation"))
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Grand Abolisher"))

	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
}

func TestCantCastFromHandProhibitsFaceDownCast(t *testing.T) {
	g, engine, behemoth := faceDownTimingGame(t, game.Player1, "War Behemoth", mana.W)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Hand Lock",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:              game.RuleEffectCantCastFromZones,
			AffectedPlayer:    game.PlayerOpponent,
			CantCastFromZones: []zone.Type{zone.Hand},
		}}}},
	}})

	assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)
}

func TestRuleOfLawLimitsFaceDownCastsFromHand(t *testing.T) {
	g, engine, first := faceDownTimingGame(t, game.Player1, "War Behemoth", mana.W)
	second := addCardToHand(g, game.Player1, productionChoiceCard(t, "War Behemoth"))
	g.Players[game.Player1].ManaPool.Add(mana.W, 3)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Rule of Law"))

	if !faceDownCastOffered(engine, g, second, game.FaceDownMorph) {
		t.Fatal("first face-down cast of the turn not offered under Rule of Law")
	}
	if !engine.applyAction(g, game.Player1, actionBuild.castFaceDown(first, game.FaceFront, game.FaceDownMorph)) {
		t.Fatal("first face-down cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	assertFaceDownCastRejected(t, engine, g, second, game.FaceDownMorph)
}

func TestNextSpellFlashPermissionUsesFaceDownCharacteristics(t *testing.T) {
	installNextSpellFlash := func(g *game.Game, subtypes ...types.Sub) {
		source := addCombatPermanent(g, game.Player1, &game.CardDef{})
		resolveInstruction(NewEngine(nil), g, &game.StackObject{
			Controller: game.Player1, Kind: game.StackActivatedAbility,
			SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		}, game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:                   game.RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer:         game.PlayerYou,
				AppliesToNextSpellOnly: true,
				SpellSubtypes:          subtypes,
			}},
			Duration: game.DurationThisTurn,
		}, &TurnLog{})
		if len(g.RuleEffects) != 1 {
			t.Fatalf("next-spell flash effects = %#v, want one", g.RuleEffects)
		}
	}

	t.Run("unfiltered permission enables and is consumed", func(t *testing.T) {
		g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
		installNextSpellFlash(g)
		castFaceDownAndResolve(t, engine, g, behemoth, game.FaceDownMorph)
		if len(g.RuleEffects) != 0 {
			t.Fatalf("face-down cast did not consume next-spell flash permission: %#v", g.RuleEffects)
		}
	})

	t.Run("subtype permission ignores the hidden Beast", func(t *testing.T) {
		g, engine, behemoth := faceDownTimingGame(t, game.Player2, "War Behemoth", mana.W)
		installNextSpellFlash(g, types.Beast)
		assertFaceDownCastRejected(t, engine, g, behemoth, game.FaceDownMorph)

		g.Turn.ActivePlayer = game.Player1
		castFaceDownAndResolve(t, engine, g, behemoth, game.FaceDownMorph)
		if len(g.RuleEffects) != 1 {
			t.Fatal("face-down spell consumed a Beast-only flash permission through its hidden subtype")
		}
	})
}

func faceDownIsTwoTwo(g *game.Game, permanent *game.Permanent) bool {
	toughness, ok := effectiveToughness(g, permanent)
	return ok && toughness == 2 && effectivePower(g, permanent) == 2
}

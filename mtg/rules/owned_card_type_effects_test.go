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

// Source review: issue #3524 (sub-issue of #3069), "The same is true for ...
// creature cards you own that aren't on the battlefield" on pinned Arcane
// Adaptation bf3edaaf-cf63-4e17-94ae-9d9991d9fb5f, Leyline of Transformation
// 4bd941ca-f3d2-44c1-8df3-851362f6b848, Maskwood Nexus
// 1246c42d-57c0-4cba-959a-15ad89d8a50b, Biotransference
// 75a787a3-5838-4121-bbaf-745d22c4b0bc and Encroaching Mycosynth
// 65a2fcc9-2317-48a1-a5eb-234fb3300364 ("nonland permanent cards you own that
// aren't on the battlefield"). Rulings from
// https://api.scryfall.com/cards/<id>/rulings: Encroaching Mycosynth 2023-02-04
// ("nonland permanent cards you own in other zones become artifact cards"),
// Maskwood Nexus 2021-02-05 (its creature cards in every other zone are every
// creature type) and Arcane Adaptation 2017-09-29/2018-01-19. Scenario cards
// use their pinned Oracle text: Goblin Matron
// 8ee0fdd4-1fb5-4160-b408-2c51ef4f85d0 (search for a Goblin card), Herald's
// Horn 4bfa71d4-4125-4404-9c24-979401bae592 (upkeep look at the top card; a
// creature card of the chosen type may be revealed and put into hand), Tarmogoyf
// 8be5ee47-8f8a-4e3c-b1a1-9ca0e1fdec7f (card types among cards in all
// graveyards), Buried Ruin 7ab1a074-9d92-4125-9200-84cd035f8f53 (target artifact
// card from your graveyard), Realmwalker 52733829-27b7-40ec-a897-cabab35b3729
// (cast creature spells of the chosen type from the top of your library),
// Grizzly Bears 409f9b88-f03e-40b6-9883-68c14c37c0de and Changeling Outcast.
// CR 2026-09-25 (official MagicCompRules 20260925.txt, sha256
// 8d860e45...070ca) read: 108.3 (a card's owner), 400.7 (a card changing zones
// is a new object), 601.3/601.3e (cast permissions consider the characteristics
// the spell would have, including continuous effects that would apply to it),
// 603.10a (leaves-the-graveyard look-back), 604.3/604.3a (CDAs function in
// every zone), 611.3a-b (a static ability's effect is not locked in and applies
// while its source is on the battlefield), 613.1/613.1d/613.7/613.7a (type
// changing effects in layer 4 in timestamp order), 701.23a (search), 702.73a
// (Changeling) and 709.4 (a split card in a non-stack zone has both halves'
// characteristics).

func addProductionCardToLibrary(t *testing.T, g *game.Game, owner game.PlayerID, name string) id.ID {
	t.Helper()
	return addCardToLibrary(g, owner, productionChoiceCard(t, name))
}

func addProductionCardToGraveyard(t *testing.T, g *game.Game, owner game.PlayerID, name string) id.ID {
	t.Helper()
	return addCardToGraveyard(g, owner, productionChoiceCard(t, name))
}

// resolveGoblinMatronSearch casts Goblin Matron, resolves it, and resolves its
// enters trigger with an agent that accepts the search and picks wanted.
func resolveGoblinMatronSearch(t *testing.T, g *game.Game, engine *Engine, wanted string) {
	t.Helper()
	castSpellTypeTestCard(t, g, engine, game.Player1, "Goblin Matron", nil, mana.R, mana.C, mana.C)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.putTriggeredAbilitiesOnStack(g)
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want Goblin Matron's enters trigger", trigger)
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: wanted}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("Goblin Matron's trigger did not resolve")
	}
}

// TestArcaneAdaptationLetsGoblinMatronFindCreatureCardInLibrary proves a
// creature card in its owner's library has the chosen type: Goblin Matron's
// search for a Goblin card finds Grizzly Bears under Arcane Adaptation (Goblin)
// and cannot find it without, or under an opponent's Arcane Adaptation.
func TestArcaneAdaptationLetsGoblinMatronFindCreatureCardInLibrary(t *testing.T) {
	for _, tc := range []struct {
		name       string
		controller game.PlayerID
		withSource bool
		want       bool
	}{
		{name: "own Arcane Adaptation", controller: game.Player1, withSource: true, want: true},
		{name: "no Arcane Adaptation", withSource: false, want: false},
		{name: "opponent's Arcane Adaptation", controller: game.Player2, withSource: true, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			if tc.withSource {
				addChosenTypePermanent(t, g, tc.controller, "Arcane Adaptation", types.Goblin)
			}
			bears := addProductionCardToLibrary(t, g, game.Player1, "Grizzly Bears")

			resolveGoblinMatronSearch(t, g, engine, "Grizzly Bears")

			if got := g.Players[game.Player1].Hand.Contains(bears); got != tc.want {
				t.Fatalf("Grizzly Bears in hand = %v, want %v", got, tc.want)
			}
			if got := g.Players[game.Player1].Library.Contains(bears); got == tc.want {
				t.Fatalf("Grizzly Bears in library = %v, want %v", got, !tc.want)
			}
		})
	}
}

// TestGoblinMatronFindsChangelingCardInLibrary proves a Changeling card is every
// creature type in the library (CR 604.3, 702.73a) with no other effect.
func TestGoblinMatronFindsChangelingCardInLibrary(t *testing.T) {
	g, engine := spellTypeScenario(t)
	outcast := addProductionCardToLibrary(t, g, game.Player1, "Changeling Outcast")

	resolveGoblinMatronSearch(t, g, engine, "Changeling Outcast")

	if !g.Players[game.Player1].Hand.Contains(outcast) {
		t.Fatal("Goblin Matron did not find the Changeling card")
	}
}

// TestLeylineOfTransformationHeraldsHornRevealsCreatureCardOnTop proves the
// owned-card clause applies to a looked-at library card: under Leyline of
// Transformation (Elf), Herald's Horn (Elf) may reveal Grizzly Bears from the
// top of the library and put it into hand. The clause also holds on a clone.
func TestLeylineOfTransformationHeraldsHornRevealsCreatureCardOnTop(t *testing.T) {
	for _, withLeyline := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		addChosenTypePermanent(t, g, game.Player1, "Herald's Horn", types.Elf)
		if withLeyline {
			addChosenTypePermanent(t, g, game.Player1, "Leyline of Transformation", types.Elf)
		}
		bears := addProductionCardToLibrary(t, g, game.Player1, "Grizzly Bears")
		g.Turn.Step = game.StepUpkeep
		emitBeginningOfStepEvent(g, game.StepUpkeep)
		engine.putTriggeredAbilitiesOnStack(g)
		trigger, ok := g.Stack.Peek()
		if !ok || trigger.Kind != game.StackTriggeredAbility {
			t.Fatalf("top of stack = %+v, want Herald's Horn's upkeep trigger", trigger)
		}
		states := []*game.Game{g}
		if withLeyline {
			states = append(states, g.Clone())
		}
		for _, state := range states {
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &lookChoiceAgent{}}
			engine.resolveTopOfStackWithChoices(state, agents, &TurnLog{})
			if got := state.Players[game.Player1].Hand.Contains(bears); got != withLeyline {
				t.Fatalf("with Leyline %v: Grizzly Bears in hand = %v", withLeyline, got)
			}
			if got := eventRevealedCard(state, bears, trigger.ID); got != withLeyline {
				t.Fatalf("with Leyline %v: Grizzly Bears revealed = %v", withLeyline, got)
			}
		}
	}
}

// TestBiotransferenceGraveyardCreatureCardCountsForTarmogoyf proves graveyard
// card-type counting reads the owned-card clause: Player1's Grizzly Bears is an
// artifact creature card under Player1's Biotransference, so Tarmogoyf counts
// two card types; Player2's Bears stays a creature card. The effect is not
// locked in (CR 611.3a): removing Biotransference reverts the count.
func TestBiotransferenceGraveyardCreatureCardCountsForTarmogoyf(t *testing.T) {
	g, engine := spellTypeScenario(t)
	goyf := addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Tarmogoyf"))
	biotransference := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	opponentBears := addProductionCardToGraveyard(t, g, game.Player2, "Grizzly Bears")
	assertGoyf := func(state *game.Game, power, toughness int, context string) {
		t.Helper()
		gotT, _ := effectiveToughness(state, goyf)
		if gotP := effectivePower(state, goyf); gotP != power || gotT != toughness {
			t.Fatalf("%s: Tarmogoyf = %d/%d, want %d/%d", context, gotP, gotT, power, toughness)
		}
	}
	assertGoyf(g, 1, 2, "only Player2's Grizzly Bears")
	if graveyardCardHasType(g, g.CardInstances[opponentBears], types.Artifact) {
		t.Fatal("Player1's Biotransference made Player2's graveyard card an artifact")
	}

	ownBears := addProductionCardToGraveyard(t, g, game.Player1, "Grizzly Bears")
	assertGoyf(g, 2, 3, "Player1's Grizzly Bears under Biotransference")
	if !graveyardCardHasType(g, g.CardInstances[ownBears], types.Artifact) {
		t.Fatal("Player1's graveyard creature card is not an artifact card")
	}
	assertGoyf(g.Clone(), 2, 3, "clone")

	removePermanentForSpellTypeTest(t, g, engine, biotransference)
	// Graveyards now hold creature, enchantment (Biotransference) and sorcery
	// (the removal spell) cards; the artifact type is gone.
	assertGoyf(g, 3, 4, "after Biotransference left the battlefield")
	if graveyardCardHasType(g, g.CardInstances[ownBears], types.Artifact) {
		t.Fatal("Grizzly Bears is still an artifact after Biotransference left the battlefield")
	}
}

// TestEncroachingMycosynthLetsBuriedRuinTargetCreatureCard proves a targeting
// selection over a graveyard reads the owned-card clause: under Encroaching
// Mycosynth, Buried Ruin can return Player1's Grizzly Bears (a nonland permanent
// card, so an artifact card) but not Player1's land card.
func TestEncroachingMycosynthLetsBuriedRuinTargetCreatureCard(t *testing.T) {
	for _, withMycosynth := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		if withMycosynth {
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Encroaching Mycosynth"))
		}
		ruin := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Buried Ruin"))
		bears := addProductionCardToGraveyard(t, g, game.Player1, "Grizzly Bears")
		forest := addProductionCardToGraveyard(t, g, game.Player1, "Forest")
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)

		landClone := g.Clone()
		if engine.applyAction(landClone, game.Player1, action.ActivateAbility(ruin.ObjectID, 0, []game.Target{game.CardTarget(forest)}, 0)) {
			t.Fatalf("with Mycosynth %v: Buried Ruin targeted a land card", withMycosynth)
		}

		activated := engine.applyAction(g, game.Player1, action.ActivateAbility(ruin.ObjectID, 0, []game.Target{game.CardTarget(bears)}, 0))
		if activated != withMycosynth {
			t.Fatalf("with Mycosynth %v: Buried Ruin activation targeting Grizzly Bears = %v", withMycosynth, activated)
		}
		if !activated {
			continue
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(bears) {
			t.Fatal("Buried Ruin did not return Grizzly Bears to hand")
		}
	}
}

// TestRealmwalkerCastsCreatureCardMadeChosenTypeByArcaneAdaptation proves the
// cast permission reads the characteristics the spell would have (CR 601.3e):
// with Realmwalker (Goblin) and Arcane Adaptation (Goblin), Grizzly Bears on
// top of the library is castable and resolves; without Arcane Adaptation it is
// not offered.
func TestRealmwalkerCastsCreatureCardMadeChosenTypeByArcaneAdaptation(t *testing.T) {
	for _, withAdaptation := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		addChosenTypePermanent(t, g, game.Player1, "Realmwalker", types.Goblin)
		if withAdaptation {
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Goblin)
		}
		bears := addProductionCardToLibrary(t, g, game.Player1, "Grizzly Bears")
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)

		var cast action.Action
		found := false
		for _, act := range engine.legalCastActions(g, game.Player1) {
			if payload, ok := act.CastSpellPayload(); ok && payload.CardID == bears && payload.SourceZone == zone.Library {
				cast, found = act, true
				break
			}
		}
		if found != withAdaptation {
			t.Fatalf("with Arcane Adaptation %v: library cast of Grizzly Bears offered = %v", withAdaptation, found)
		}
		if !found {
			continue
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("casting Grizzly Bears from the top of the library failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !slices.ContainsFunc(g.Battlefield, func(p *game.Permanent) bool { return p.CardInstanceID == bears }) {
			t.Fatal("Grizzly Bears cast from the library did not enter the battlefield")
		}
	}
}

// TestMaskwoodNexusOwnedCardsInHandAndExileAreEveryCreatureType proves the
// owned-card clause covers the hand and exile and only creature cards, and that
// a card on the stack or battlefield is not an owned card outside the
// battlefield.
func TestMaskwoodNexusOwnedCardsInHandAndExileAreEveryCreatureType(t *testing.T) {
	g, _ := spellTypeScenario(t)
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Maskwood Nexus"))
	inHand := addCardToHand(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	exiled := g.IDGen.Next()
	g.CardInstances[exiled] = &game.CardInstance{ID: exiled, Def: productionChoiceCard(t, "Grizzly Bears"), Owner: game.Player1}
	g.Players[game.Player1].Exile.Add(exiled)
	bolt := addCardToHand(g, game.Player1, productionChoiceCard(t, "Lightning Bolt"))
	opponents := addCardToHand(g, game.Player2, productionChoiceCard(t, "Grizzly Bears"))

	elf := game.Selection{SubtypesAny: []types.Sub{types.Elf}}
	for _, tc := range []struct {
		name string
		card id.ID
		want bool
	}{
		{name: "creature card in hand", card: inHand, want: true},
		{name: "creature card in exile", card: exiled, want: true},
		{name: "instant card in hand", card: bolt, want: false},
		{name: "opponent's creature card", card: opponents, want: false},
	} {
		subject := selectionSubject{kind: subjectCard, g: g, card: g.CardInstances[tc.card]}
		if got := matchSelection(&subject, &elf); got != tc.want {
			t.Fatalf("%s is an Elf card = %v, want %v", tc.name, got, tc.want)
		}
	}

	bearsPermanent := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
	subject := selectionSubject{kind: subjectCard, g: g, card: g.CardInstances[bearsPermanent.CardInstanceID]}
	if matchSelection(&subject, &elf) {
		t.Fatal("the owned-card clause applied to a card on the battlefield")
	}
}

// Source review (repair): pinned Gilt-Leaf Palace
// cc4bbbb5-4218-4b2a-9ca2-de4d5f5dda19 ("As this land enters, you may reveal an
// Elf card from your hand. If you don't, this land enters tapped."), Neurok
// Prodigy fb179448-5e1f-4ef3-858c-ba7a9ea05a78 ("Discard an artifact card:
// Return this creature to its owner's hand."), Rust Harvester
// 7ce58765-d2f4-4fbf-8635-580c9400ff2e ("{2}, {T}, Exile an artifact card from
// your graveyard: Put a +1/+1 counter on this creature, then it deals damage
// equal to its power to any target."), Deathcap Cultivator
// 17f9339a-d327-4239-a0d8-12f7ff58982b (delirium deathtouch; 2016-04-08
// rulings: card types, not cards, are counted), Leafcrown Dryad
// 8202e426-ad91-4d2e-9373-7a829b58fff5 (Bestow {3}{G}; 2013-09-15 ruling: on the
// stack a bestow spell is either a creature spell or an Aura spell, never both)
// and Future Sight. Realmwalker 2021-02-05 rulings: you still pay all costs and
// may pay available alternative costs. CR 2026-09-25: 118.3 (a player can't pay
// a cost without the resources), 601.2b/601.2h (costs), 601.3e (alternative
// characteristics and the continuous effects that would apply to them decide
// whether a card may be cast), 701.20a (reveal), 702.103a-b (a spell cast
// bestowed becomes an Aura enchantment spell) and 613.1d.

// TestArcaneAdaptationElfCardRevealedForGiltLeafPalace proves a reveal cost
// reads the real card in hand: under Arcane Adaptation (Elf), Grizzly Bears is
// an Elf card, so Gilt-Leaf Palace reveals it and enters untapped; without the
// enchantment it enters tapped.
func TestArcaneAdaptationElfCardRevealedForGiltLeafPalace(t *testing.T) {
	for _, withAdaptation := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		if withAdaptation {
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Elf)
		}
		bears := addCardToHand(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		land := addCardToHand(g, game.Player1, productionChoiceCard(t, "Gilt-Leaf Palace"))
		if !engine.applyPlayLand(g, game.Player1, land) {
			t.Fatal("could not play Gilt-Leaf Palace")
		}
		palace := permanentByCardID(g, land)
		if palace == nil {
			t.Fatal("Gilt-Leaf Palace is not on the battlefield")
		}
		revealed := eventRevealedCardFromZone(g, game.Player1, land, bears, zone.Hand)
		if revealed != withAdaptation || palace.Tapped == withAdaptation {
			t.Fatalf("with Arcane Adaptation %v: revealed Grizzly Bears = %v, Gilt-Leaf Palace tapped = %v",
				withAdaptation, revealed, palace.Tapped)
		}
	}
}

// TestBiotransferenceArtifactCardCosts proves discard and graveyard-exile costs
// read the real card: under Biotransference, Grizzly Bears in hand is an
// artifact card Neurok Prodigy can discard, and Grizzly Bears in the graveyard
// is an artifact card Rust Harvester can exile. Without Biotransference neither
// cost can be paid.
func TestBiotransferenceArtifactCardCosts(t *testing.T) {
	for _, withBiotransference := range []bool{true, false} {
		g, engine := spellTypeScenario(t)
		if withBiotransference {
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
		}
		prodigy := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Neurok Prodigy"))
		inHand := addCardToHand(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		discarded := engine.applyAction(g, game.Player1, action.ActivateAbility(prodigy.ObjectID, 0, nil, 0))
		if discarded != withBiotransference {
			t.Fatalf("with Biotransference %v: Neurok Prodigy discard activation = %v", withBiotransference, discarded)
		}
		if discarded && !g.Players[game.Player1].Graveyard.Contains(inHand) {
			t.Fatal("Neurok Prodigy's cost did not discard Grizzly Bears")
		}

		harvester := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Rust Harvester"))
		inGraveyard := addProductionCardToGraveyard(t, g, game.Player1, "Grizzly Bears")
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		exiled := engine.applyAction(g, game.Player1, action.ActivateAbility(harvester.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0))
		if exiled != withBiotransference {
			t.Fatalf("with Biotransference %v: Rust Harvester exile activation = %v", withBiotransference, exiled)
		}
		if exiled && !g.Players[game.Player1].Exile.Contains(inGraveyard) && !g.Players[game.Player1].Exile.Contains(inHand) {
			t.Fatal("Rust Harvester's cost did not exile a Grizzly Bears card")
		}
	}
}

// TestBiotransferenceCountsTowardDelirium proves delirium counts the types
// type-changing effects give graveyard cards: Grizzly Bears, Lightning Bolt and
// Forest are four card types under Biotransference (creature and artifact,
// instant, land), so Deathcap Cultivator has deathtouch; without the
// enchantment they are three.
func TestBiotransferenceCountsTowardDelirium(t *testing.T) {
	for _, withBiotransference := range []bool{true, false} {
		g, _ := spellTypeScenario(t)
		cultivator := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Deathcap Cultivator"))
		if withBiotransference {
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
		}
		for _, name := range []string{"Grizzly Bears", "Lightning Bolt", "Forest"} {
			addProductionCardToGraveyard(t, g, game.Player1, name)
		}
		if got := hasKeyword(g, cultivator, game.Deathtouch); got != withBiotransference {
			t.Fatalf("with Biotransference %v: Deathcap Cultivator deathtouch = %v (graveyard card types %d)",
				withBiotransference, got, controllerGraveyardCardTypeCount(g, game.Player1))
		}
	}
}

// TestRealmwalkerPermissionReadsSelectedCastMode proves a cast permission's
// type filter reads the spell as cast in the selected mode (CR 601.3e,
// 702.103b): with Realmwalker and Arcane Adaptation (Goblin), Leafcrown Dryad
// on top of the library may be cast as a Goblin creature spell, but its bestowed
// cast is an Aura spell, not a creature spell, so it is neither offered nor
// accepted. Future Sight, which permits any spell, offers and accepts the
// bestowed cast.
func TestRealmwalkerPermissionReadsSelectedCastMode(t *testing.T) {
	for _, futureSight := range []bool{false, true} {
		g, engine := spellTypeScenario(t)
		if futureSight {
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Future Sight"))
		} else {
			addChosenTypePermanent(t, g, game.Player1, "Realmwalker", types.Goblin)
			addChosenTypePermanent(t, g, game.Player1, "Arcane Adaptation", types.Goblin)
		}
		bears := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Grizzly Bears"))
		dryad := addProductionCardToLibrary(t, g, game.Player1, "Leafcrown Dryad")
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 3)

		normal, bestowed := false, false
		for _, act := range engine.legalCastActions(g, game.Player1) {
			if payload, ok := act.CastSpellPayload(); ok && payload.CardID == dryad && payload.SourceZone == zone.Library {
				if payload.Bestowed {
					bestowed = true
				} else {
					normal = true
				}
			}
		}
		if !normal || bestowed != futureSight {
			t.Fatalf("with Future Sight %v: library casts offered normal = %v, bestowed = %v", futureSight, normal, bestowed)
		}
		bestow := actionBuild.castBestowSpell(dryad, zone.Library, game.FaceFront, []game.Target{game.PermanentTarget(bears.ObjectID)}, 0, nil)
		if got := engine.applyAction(g.Clone(), game.Player1, bestow); got != futureSight {
			t.Fatalf("with Future Sight %v: bestowed library cast accepted = %v", futureSight, got)
		}
		if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(dryad, zone.Library, game.FaceFront, nil, 0, nil)) {
			t.Fatalf("with Future Sight %v: creature-spell library cast of Leafcrown Dryad failed", futureSight)
		}
	}
}

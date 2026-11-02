package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3528, controller graveyard card-type reads must use the
// effective types of cards in the graveyard. Pinned Oracle text: Unholy Heat
// 4e879386-b1f8-4f2a-9820-6e1291746f88 ("Unholy Heat deals 2 damage to target
// creature or planeswalker. Delirium — Unholy Heat deals 6 damage instead if
// there are four or more card types among cards in your graveyard."; 2021-06-18
// rulings: it checks your graveyard as it resolves, when it isn't in the
// graveyard yet, and counts card types such as artifact, creature, instant and
// land), Deathcap Cultivator 17f9339a-d327-4239-a0d8-12f7ff58982b (delirium
// deathtouch), Biotransference 75a787a3-5838-4121-bbaf-745d22c4b0bc ("...
// creature cards you own that aren't on the battlefield" are artifacts too),
// Reclusive Taxidermist 10edf37a-ee35-491d-b83a-39035f7df65a (+3/+2 with four or
// more creature cards in your graveyard), Ghitu Lavarunner
// 8940d76d-c09f-4d72-a037-439c70ee8d9d (+1/+0 and haste with two or more instant
// and/or sorcery cards in your graveyard), Guardian Project
// 2c523ef7-d6c4-4887-9a4e-0c2bc6f63676 ("if it doesn't have the same name as
// another creature you control or a creature card in your graveyard"), Grizzly
// Bears, Lightning Bolt, Forest and Colossal Dreadmaw
// 8059c52b-5d25-4052-b48a-e9e219a7a546 (6/6). CR 2026-09-25: 108.3 (owner),
// 205.2a (card types), 611.3a (a static ability's effect applies while its
// source is on the battlefield), 613.1d (type-changing effects in layer 4) and
// 709.4 (a split card has both halves' characteristics outside the stack). No
// registered card adds creature, instant or sorcery to cards outside the
// battlefield, so the per-type counts use a synthetic owned-card type changer.

// TestBiotransferenceDeliriumUnholyHeatResolution proves the resolving-spell
// condition route reads effective graveyard types: with Grizzly Bears,
// Lightning Bolt and Forest in its controller's graveyard, Unholy Heat deals 6
// to Colossal Dreadmaw under its controller's Biotransference (four types) and
// 2 without it or under an opponent's Biotransference, which only changes cards
// that opponent owns.
func TestBiotransferenceDeliriumUnholyHeatResolution(t *testing.T) {
	for _, tc := range []struct {
		name       string
		enchantOwn bool
		enchantOpp bool
		wantDamage int
	}{
		{name: "no Biotransference", wantDamage: 2},
		{name: "own Biotransference", enchantOwn: true, wantDamage: 6},
		{name: "opponent's Biotransference", enchantOpp: true, wantDamage: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := spellTypeScenario(t)
			if tc.enchantOwn {
				addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
			}
			if tc.enchantOpp {
				addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Biotransference"))
			}
			for _, name := range []string{"Grizzly Bears", "Lightning Bolt", "Forest"} {
				addProductionCardToGraveyard(t, g, game.Player1, name)
			}
			dreadmaw := addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Colossal Dreadmaw"))
			castSpellTypeTestCard(t, g, engine, game.Player1, "Unholy Heat",
				[]game.Target{game.PermanentTarget(dreadmaw.ObjectID)}, mana.R)
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := g.Stack.Peek(); ok {
				t.Fatal("Unholy Heat did not resolve")
			}
			if dreadmaw.MarkedDamage != tc.wantDamage {
				t.Fatalf("Colossal Dreadmaw damage = %d, want %d (graveyard card types %d)",
					dreadmaw.MarkedDamage, tc.wantDamage, controllerGraveyardCardTypeCount(g, game.Player1))
			}
		})
	}
}

// TestDeathcapCultivatorDeliriumIgnoresOpponentBiotransference proves the
// static-condition route follows the type changer's controller: an opponent's
// Biotransference does not make Grizzly Bears in Deathcap Cultivator's
// controller's graveyard an artifact card, and the bonus ends when its
// controller's Biotransference is exiled.
func TestDeathcapCultivatorDeliriumIgnoresOpponentBiotransference(t *testing.T) {
	g, engine := spellTypeScenario(t)
	cultivator := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Deathcap Cultivator"))
	addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Biotransference"))
	for _, name := range []string{"Grizzly Bears", "Lightning Bolt", "Forest"} {
		addProductionCardToGraveyard(t, g, game.Player1, name)
	}
	if hasKeyword(g, cultivator, game.Deathtouch) {
		t.Fatal("an opponent's Biotransference gave Deathcap Cultivator delirium")
	}
	own := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	if !hasKeyword(g, cultivator, game.Deathtouch) {
		t.Fatal("its controller's Biotransference did not give Deathcap Cultivator delirium")
	}
	addEffectSpellToStack(g, game.Player2, game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Exile,
	}, []game.Target{game.PermanentTarget(own.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, own.ObjectID); ok {
		t.Fatal("setup effect did not exile Biotransference")
	}
	if hasKeyword(g, cultivator, game.Deathtouch) {
		t.Fatal("Deathcap Cultivator kept delirium after Biotransference left the battlefield")
	}
}

// ownedCardTypeAdder is a synthetic enchantment: "<from> permanents you control
// are <add> in addition to their other types. The same is true for <from> cards
// you own that aren't on the battlefield."
func ownedCardTypeAdder(from, add types.Card) *game.CardDef {
	selection := game.Selection{RequiredTypes: []types.Card{from}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Owned Card Type Adder",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Duration:           game.DurationPermanent,
				Group:              game.ObjectControlledGroup(game.SourcePermanentReference(), selection),
				Layer:              game.LayerType,
				OwnedCardSelection: &selection,
				AddTypes:           []types.Card{add},
			}},
		}},
	}}
}

// TestGraveyardCreatureCardCountReadsEffectiveTypes proves the "creature cards
// in your graveyard" count reads effective types: four Forest cards are creature
// cards while their owner controls a land-to-creature owned-card type changer,
// so Reclusive Taxidermist is 4/4; without it, or under an opponent's, it is
// 1/2.
func TestGraveyardCreatureCardCountReadsEffectiveTypes(t *testing.T) {
	for _, owner := range []opt.V[game.PlayerID]{{}, opt.Val(game.Player1), opt.Val(game.Player2)} {
		g, _ := spellTypeScenario(t)
		if owner.Exists {
			addCombatPermanent(g, owner.Val, ownedCardTypeAdder(types.Land, types.Creature))
		}
		taxidermist := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Reclusive Taxidermist"))
		for range 4 {
			addProductionCardToGraveyard(t, g, game.Player1, "Forest")
		}
		toughness, _ := effectiveToughness(g, taxidermist)
		power := effectivePower(g, taxidermist)
		pumped := owner.Exists && owner.Val == game.Player1
		if pumped != (power == 4 && toughness == 4) || !pumped && (power != 1 || toughness != 2) {
			t.Fatalf("type changer %+v: Reclusive Taxidermist is %d/%d", owner, power, toughness)
		}
	}
}

// TestGraveyardInstantOrSorceryCountReadsEffectiveTypes proves spell mastery's
// "instant and/or sorcery cards in your graveyard" count reads effective types:
// two Forest cards are instant cards under a land-to-instant owned-card type
// changer, so Ghitu Lavarunner gets +1/+0 and haste; without it it does not.
func TestGraveyardInstantOrSorceryCountReadsEffectiveTypes(t *testing.T) {
	for _, withChanger := range []bool{false, true} {
		g, _ := spellTypeScenario(t)
		if withChanger {
			addCombatPermanent(g, game.Player1, ownedCardTypeAdder(types.Land, types.Instant))
		}
		lavarunner := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Ghitu Lavarunner"))
		for range 2 {
			addProductionCardToGraveyard(t, g, game.Player1, "Forest")
		}
		power := effectivePower(g, lavarunner)
		haste := hasKeyword(g, lavarunner, game.Haste)
		if haste != withChanger || (withChanger && power != 2) || (!withChanger && power != 1) {
			t.Fatalf("type changer %v: Ghitu Lavarunner power %d haste %v", withChanger, power, haste)
		}
	}
}

// TestGuardianProjectNameCheckReadsEffectiveGraveyardCreatures proves Guardian
// Project's "a creature card in your graveyard" name check reads effective
// types: a noncreature artifact card named Relic in the graveyard blocks the
// draw for an entering Relic creature only while an artifact-to-creature
// owned-card type changer makes it a creature card.
func TestGuardianProjectNameCheckReadsEffectiveGraveyardCreatures(t *testing.T) {
	condition := opt.Val(game.Condition{EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true})
	for _, withChanger := range []bool{false, true} {
		g, _ := spellTypeScenario(t)
		if withChanger {
			addCombatPermanent(g, game.Player1, ownedCardTypeAdder(types.Artifact, types.Creature))
		}
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}})
		entering := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Creature}}})
		ctx := conditionContext{controller: game.Player1, event: &game.Event{PermanentID: entering.ObjectID}}
		if unique := conditionSatisfied(g, ctx, condition); unique == withChanger {
			t.Fatalf("type changer %v: name unique = %v", withChanger, unique)
		}
	}
}

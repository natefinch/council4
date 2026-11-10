package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned Fortune Teller's Talent Oracle ID 1b430f67-3686-4452-8594-b060f1a5a04e,
// printing a1d43877-20ab-4e84-a597-4b5e03a6bf90 from corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; all 11 Scryfall rulings retrieved
// 2026-11-09 (https://api.scryfall.com/cards/a1d43877-20ab-4e84-a597-4b5e03a6bf90/rulings)
// were reviewed. They require normal timing and full costs for top-of-library plays,
// and limit the level-3 reduction to generic mana. CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (local SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 305.1, 307.5, 401.2-4, 601.2, and 717.2a/3 govern land/spell timing, private
// library information, casting costs, and which Class abilities are active.
func TestFortuneTellersTalentOwnText(t *testing.T) {
	g, engine, talent := castClassForTest(t, "1b430f67-3686-4452-8594-b060f1a5a04e", mana.U, 0)
	gateGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	gateEngine := NewEngine(nil)
	enterClassTestPermanent(t, gateGame, game.Player1, g.CardInstances[talent.CardInstanceID].Def)
	gateLand := addCardToLibrary(gateGame, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Land before a spell",
		Types: []types.Card{types.Land},
	}})
	setMainPhasePriority(gateGame, game.Player1)
	if gateEngine.applyAction(gateGame, game.Player1, action.PlayLandFaceFromZone(gateLand, zone.Library, game.FaceFront)) {
		t.Fatal("played a land from the top without having cast a spell this turn")
	}

	topSpell := addCardToLibrary(g, game.Player1, classCardSorcery("Top library spell", cost.Mana{cost.O(1), cost.U}))
	topLand := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Top library land",
		Types: []types.Card{types.Land},
	}})
	if got, ok := g.Players[game.Player1].Library.Top(); !ok || got != topLand {
		t.Fatalf("library top = %d, want land %d", got, topLand)
	}
	if got, ok := NewObservation(g, game.Player1).LibraryTopLookable(game.Player1); !ok || got.CardInstanceID != topLand {
		t.Fatalf("owner's top-card view = %+v, %v; want private view of land %d", got, ok, topLand)
	}
	opponent := NewObservation(g, game.Player2)
	if _, ok := opponent.LibraryTopLookable(game.Player1); ok {
		t.Fatal("opponent can privately look at Fortune Teller's Talent's controller's library")
	}
	if _, ok := opponent.LibraryTopRevealed(game.Player1); ok {
		t.Fatal("Fortune Teller's Talent revealed the top card to an opponent")
	}

	handSpell := addCardToHand(g, game.Player1, classTestInstant())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(handSpell, nil, 0, nil)) {
		t.Fatal("could not cast a spell from hand to enable Fortune Teller's Talent")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	activateClassLevelForTest(t, g, engine, talent, 0)
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	g.Turn.PriorityPlayer = game.Player1
	if engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(topLand, zone.Library, game.FaceFront)) {
		t.Fatal("Fortune Teller's Talent let its controller play a land outside a main phase")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(topLand, zone.Library, game.FaceFront)) {
		t.Fatal("could not play the top land at level 2 after casting a spell")
	}
	if classTestPermanentForCard(g, g.CardInstances[topLand]) == nil {
		t.Fatal("top land did not enter the battlefield")
	}

	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	g.Turn.PriorityPlayer = game.Player1
	if engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topSpell, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Fortune Teller's Talent let its controller cast a sorcery outside a legal timing window")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topSpell, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("could not cast the top sorcery at its normal timing and full cost")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting the top sorcery = %d, want its full {1}{U} cost paid", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(topSpell) {
		t.Fatal("top sorcery did not resolve to its owner's graveyard")
	}

	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	activateClassLevelForTest(t, g, engine, talent, 1)
	handCopy := addCardToHand(g, game.Player1, classCardSorcery("Hand spell", cost.Mana{cost.O(2), cost.U}))
	discountedTop := addCardToLibrary(g, game.Player1, classCardSorcery("Discounted top spell", cost.Mana{cost.O(2), cost.U}))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpell(handCopy, nil, 0, nil)) {
		t.Fatal("Fortune Teller's Talent reduced the cost of a spell cast from hand")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(discountedTop, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("level-3 cost reduction did not allow casting a top spell for its colored mana")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after the discounted cast = %d, want the colored mana spent", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
}

// Source review: pinned Ranger Class Oracle ID d1586808-efc2-469a-8384-ba4acf231e69,
// printing 7ca392ca-3219-4694-9a74-aa079c76b91e from corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; all 12 Scryfall rulings retrieved
// 2026-11-09 (https://api.scryfall.com/cards/7ca392ca-3219-4694-9a74-aa079c76b91e/rulings)
// were reviewed. The attack trigger is once per combat, and casting from the top
// retains normal timing and costs. CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (local SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2, 603.4, and 717.2a/3 govern casting, the intervening level gate, and Class
// abilities; the ruling constrains the one-or-more attack trigger and top casting.
func TestRangerClassOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "d1586808-efc2-469a-8384-ba4acf231e69", mana.G, 1)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Ranger Class's enters ability did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	var wolf *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentEffectiveName(g, permanent) == "Wolf" {
			wolf = permanent
			break
		}
	}
	if wolf == nil || effectivePower(g, wolf) != 2 || effectiveToughnessMust(t, g, wolf) != 2 {
		t.Fatalf("Ranger Class token = %+v, want a 2/2 Wolf", wolf)
	}

	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	activateClassLevelForTest(t, g, engine, class, 0)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	declareClassAttackers(t, g, engine, first, second)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Ranger Class's level-2 attack ability did not trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("attack-trigger stack size = %d, want one trigger for both attackers", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if first.Counters.Get(counter.PlusOnePlusOne)+second.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatalf("attacking-creature counters = %d/%d, want one counter on one legal attacker",
			first.Counters.Get(counter.PlusOnePlusOne), second.Counters.Get(counter.PlusOnePlusOne))
	}

	g.Combat = nil
	topCreature := addCardToLibrary(g, game.Player1, classCardCreature("Top library creature", cost.Mana{cost.O(1), cost.G}))
	setMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topCreature, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("cast a creature from the top before Ranger Class reached level 3")
	}
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	activateClassLevelForTest(t, g, engine, class, 1)
	if got, ok := NewObservation(g, game.Player1).LibraryTopLookable(game.Player1); !ok || got.CardInstanceID != topCreature {
		t.Fatalf("level-3 owner's top-card view = %+v, %v; want creature %d", got, ok, topCreature)
	}
	if _, ok := NewObservation(g, game.Player2).LibraryTopLookable(game.Player1); ok {
		t.Fatal("Ranger Class let an opponent privately look at its controller's library")
	}

	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	g.Turn.PriorityPlayer = game.Player1
	if engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topCreature, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Ranger Class bypassed normal creature-spell timing")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(topCreature, zone.Library, game.FaceFront, nil, 0, nil)) {
		t.Fatal("could not cast the top creature during a legal main-phase window")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting the top creature = %d, want its full {1}{G} cost paid", got)
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if classTestPermanentForCard(g, g.CardInstances[topCreature]) == nil {
		t.Fatal("top creature did not resolve to the battlefield")
	}
}

// Source review: pinned Warlock Class Oracle ID 766c2c67-d1ee-44ff-b9bb-355361646ef1,
// printing b7faf899-96b7-454e-b634-6684c2d72f26 from corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; all 9 Scryfall rulings retrieved
// 2026-11-09 (https://api.scryfall.com/cards/b7faf899-96b7-454e-b634-6684c2d72f26/rulings)
// were reviewed. The level-1 ability sees a creature death earlier in the turn even
// before Warlock Class entered; the level-3 ability uses each opponent's own loss,
// with damage counting as life loss. CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (local SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 120.1-3, 603.4, and 717.2a/3 govern damage/life loss, the intervening condition,
// and level-gated abilities.
func TestWarlockClassOwnText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	deadCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	if !movePermanentToZone(g, deadCreature, zone.Graveyard) {
		t.Fatal("could not make a creature die before Warlock Class entered")
	}
	class := castClassForTestInGame(t, g, engine, "766c2c67-d1ee-44ff-b9bb-355361646ef1", mana.B, 0)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Warlock Class did not see the earlier creature death at the beginning of the end step")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("level-1 end-step stack size = %d, want one trigger", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[player].Life; got != 39 {
			t.Fatalf("opponent %v life after the level-1 trigger = %d, want 39", player, got)
		}
	}

	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine = NewEngine(nil)
	class = castClassForTestInGame(t, g, engine, "766c2c67-d1ee-44ff-b9bb-355361646ef1", mana.B, 0)
	alpha := addCardToLibrary(g, game.Player1, classCardSorcery("Dig Alpha", cost.Mana{}))
	middle := addCardToLibrary(g, game.Player1, classCardSorcery("Dig Middle", cost.Mana{}))
	omega := addCardToLibrary(g, game.Player1, classCardSorcery("Dig Omega", cost.Mana{}))
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	activateClassLevelForTest(t, g, engine, class, 0)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Warlock Class's level-2 Dig ability did not trigger")
	}
	digAgent := &fourClassScenarioAgent{digName: "Dig Middle"}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: digAgent}, &TurnLog{})
	if digAgent.digChoice != middle || !g.Players[game.Player1].Hand.Contains(middle) {
		t.Fatalf("Dig choice = %d; wanted %d in hand", digAgent.digChoice, middle)
	}
	for _, card := range []id.ID{alpha, omega} {
		if !g.Players[game.Player1].Graveyard.Contains(card) {
			t.Fatalf("unselected Dig card %d did not go to the graveyard", card)
		}
	}

	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	activateClassLevelForTest(t, g, engine, class, 1)
	damageSource := addColoredSourceCard(g, game.Player1, color.Green)
	if dealt := dealPlayerDamage(g, damageSource, 0, game.Player1, game.Player2, 3, false); dealt != 3 {
		t.Fatalf("pre-end-step damage to Player2 = %d, want 3", dealt)
	}
	resolveClassScenarioLifeLoss(g, engine, game.Player2, 1)
	resolveClassScenarioLifeLoss(g, engine, game.Player3, 2)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Warlock Class's level-3 end-step ability did not trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("no-death level-3 end-step stack size = %d, want only the level-3 trigger", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 32 {
		t.Fatalf("Player2 life = %d, want 32 after losing 4 earlier and 4 to Warlock Class", got)
	}
	if got := g.Players[game.Player3].Life; got != 36 {
		t.Fatalf("Player3 life = %d, want 36 after losing 2 earlier and 2 to Warlock Class", got)
	}
	if got := g.Players[game.Player4].Life; got != 40 {
		t.Fatalf("Player4 life = %d, want 40 when they lost no life earlier", got)
	}
}

// Source review: pinned Artist's Talent Oracle ID ea734863-5793-49d6-b10e-109432aec0a3,
// printing 8b9e51d9-189b-4dd6-87cb-628ea6373e81 from corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4; all 10 Scryfall rulings retrieved
// 2026-11-09 (https://api.scryfall.com/cards/8b9e51d9-189b-4dd6-87cb-628ea6373e81/rulings)
// were reviewed. The level-2 reduction affects generic mana only; the level-3 damage
// stays from the original source and the affected player orders it with prevention,
// with no added damage if all is prevented. CR 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (local SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2, 614.1, 616.1e-f, 717.2a/3, and 120.1-3 govern cost application, replacement
// order, Class levels, and damage.
func TestArtistsTalentOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "ea734863-5793-49d6-b10e-109432aec0a3", mana.R, 1)
	discard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discard for Artist's Talent"}})
	draw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Artist's Talent draw"}})
	firstSpell := addCardToHand(g, game.Player1, classTestInstant())
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(firstSpell, nil, 0, nil)) {
		t.Fatal("could not cast the noncreature spell that triggers Artist's Talent")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Artist's Talent did not trigger for its controller's noncreature spell")
	}
	accept := &fourClassScenarioAgent{acceptOptional: true}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: accept}, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(discard) || !g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("accepting Artist's Talent's optional discard did not discard and draw")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	activateClassLevelForTest(t, g, engine, class, 0)
	kept := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Keep when declining"}})
	remainingDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Must not draw on decline"}})
	reducedCostSpell := addCardToHand(g, game.Player1, classCardSorcery("Reduced noncreature spell", cost.Mana{cost.O(2), cost.U}))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(reducedCostSpell, nil, 0, nil)) {
		t.Fatal("Artist's Talent did not reduce a noncreature spell's generic cost by one")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Artist's Talent did not trigger for the level-2 noncreature spell")
	}
	decline := &fourClassScenarioAgent{acceptOptional: false}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: decline}, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(kept) || !g.Players[game.Player1].Library.Contains(remainingDraw) {
		t.Fatal("declining Artist's Talent's optional discard changed hand or library")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})

	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	activateClassLevelForTest(t, g, engine, class, 1)
	source := addColoredSourceCard(g, game.Player1, color.Red)
	if dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player2, 3, false); dealt != 5 {
		t.Fatalf("noncombat damage to an opponent = %d, want 5", dealt)
	}
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	if dealt := dealPermanentDamage(g, source, 0, game.Player1, opponentCreature, 2, false); dealt != 4 {
		t.Fatalf("noncombat damage to an opponent's permanent = %d, want 4", dealt)
	}
	if dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("damage to its controller = %d, want unmodified 3", dealt)
	}
	opponentSource := addColoredSourceCard(g, game.Player2, color.Red)
	if dealt := dealPlayerDamage(g, opponentSource, 0, game.Player2, game.Player3, 2, false); dealt != 2 {
		t.Fatalf("opponent-controlled source damage = %d, want unmodified 2", dealt)
	}
	if dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player3, 2, true); dealt != 2 {
		t.Fatalf("combat damage to an opponent = %d, want unmodified 2", dealt)
	}

	g.PreventionShields = append(g.PreventionShields, regPlayerShield(g.IDGen.Next(), game.Player2, 3))
	if dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player2, 3, false); dealt != 0 {
		t.Fatalf("damage after the default prevention-first order = %d, want 0", dealt)
	}
	if len(g.ReplacementDecisions) == 0 || !g.ReplacementDecisions[len(g.ReplacementDecisions)-1].UsedFallback {
		t.Fatalf("default prevention/replacement order was not recorded as a fallback: %+v", g.ReplacementDecisions)
	}

	g.PreventionShields = append(g.PreventionShields, regPlayerShield(g.IDGen.Next(), game.Player2, 3))
	orderingAgent := &fourClassScenarioAgent{replacementPreference: "plus 2"}
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player2: orderingAgent}, &TurnLog{})
	dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player2, 3, false)
	g.ClearChoiceContext()
	if dealt != 2 {
		t.Fatalf("damage after choosing Artist's Talent before prevention = %d, want 2", dealt)
	}
	if orderingAgent.replacementLabel == "" || !strings.Contains(strings.ToLower(orderingAgent.replacementLabel), "plus 2") {
		t.Fatalf("replacement order choice = %q, want Artist's Talent's +2 effect", orderingAgent.replacementLabel)
	}
}

type fourClassScenarioAgent struct {
	simPassPolicy
	acceptOptional        bool
	digName               string
	digChoice             id.ID
	replacementPreference string
	replacementLabel      string
}

func (a *fourClassScenarioAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		if a.acceptOptional {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceDig:
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.Name == a.digName {
				a.digChoice = option.Card.Val.CardID
				return []int{option.Index}
			}
		}
	case game.ChoiceReplacement:
		for _, option := range request.Options {
			if strings.Contains(strings.ToLower(option.Label), a.replacementPreference) {
				a.replacementLabel = option.Label
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func classCardSorcery(name string, manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

func classCardCreature(name string, manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(manaCost),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func resolveClassScenarioLifeLoss(g *game.Game, engine *Engine, player game.PlayerID, amount int) {
	addEffectSpellToStack(g, game.Player1, game.LoseLife{
		Amount: game.Fixed(amount),
		Player: game.TargetPlayerReference(0),
	}, []game.Target{game.PlayerTarget(player)})
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
}

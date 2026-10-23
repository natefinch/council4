package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Oracle corpus: /Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/oracle-cards.json,
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Shattering Pulse Oracle 77953b7a-813e-45f5-ad7f-ab72a79a0112, printing
// 89d3b846-6071-4d65-86ba-da08c4bd0aa1; rulings URI
// https://api.scryfall.com/cards/89d3b846-6071-4d65-86ba-da08c4bd0aa1/rulings.
// Whispers of the Muse Oracle a4a8af4e-8686-45ba-94fd-ccf6cc8adaae, printing
// 5852a174-eef6-4c06-abcd-fd90f4b8a188; rulings URI
// https://api.scryfall.com/cards/5852a174-eef6-4c06-abcd-fd90f4b8a188/rulings.
// Disturbed Burial Oracle cdbaa21d-5896-4b8b-9a08-f0839ccfddfc, printing
// 06254b6c-eb22-4ec9-9420-74e9ee15e072; rulings URI
// https://api.scryfall.com/cards/06254b6c-eb22-4ec9-9420-74e9ee15e072/rulings.
// All three full printing-specific responses were retrieved 2026-10-23 and
// contain no rulings; each is retained at
// /Users/natefinch/.copilot/session-state/25e78302-d27a-418d-9506-e2571ee328eb/files/sources/{shattering-pulse,whispers,disturbed-burial}-rulings.json
// (SHA-256 6959514f6cea4012b6cceb5fe573de3433c3927da3183236a64ab122bc02ef10).
// Official Comprehensive Rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy /Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/high-review-rules.txt,
// effective 2026-09-25, SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read 118.8a-d and 601.2b-h (additional costs, target announcement, locked
// total, mana abilities, atomic payment); 307.1 (sorcery timing); 608.2b-e
// (target legality and resolution); 400.7 (new zone objects); 702.27a
// (buyback's stack ability and successful-resolution replacement); 707.10/a
// (copies retain decisions but aren't casts or physical cards); 701.8a-c
// (destruction, indestructible/shield/regeneration boundaries); 614.1/4,
// 615.1, and 616.1e-g (replacement/prevention event processing and order);
// and 121.1-4 (individual draws, replacements, and empty-library loss at the
// next priority opportunity).

func buybackValueCase(t *testing.T, oracleID string) buybackInteractionCase {
	t.Helper()
	for _, tc := range buybackInteractionCases() {
		if tc.oracleID == oracleID {
			return tc
		}
	}
	t.Fatalf("buyback interaction case missing Oracle ID %s", oracleID)
	return buybackInteractionCase{}
}

func TestBuybackValueShatteringPulseResolutionAndReplacements(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	tc := buybackValueCase(t, authored.ShatteringPulse.OracleID)
	for _, scenario := range []string{"indestructible", "shield", "regeneration", "graveyard redirected"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			targetDef := &game.CardDef{CardFace: game.CardFace{
				Name: "Pulse artifact", Types: []types.Card{types.Artifact},
			}}
			if scenario == "indestructible" {
				targetDef.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
			}
			target := addCombatPermanent(g, game.Player2, targetDef)
			switch scenario {
			case "shield":
				target.Counters.Add(counter.Shield, 1)
			case "regeneration":
				target.RegenerationShields = 1
				target.MarkedDamage = 1
			case "graveyard redirected":
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard,
					ReplaceToZone: zone.Exile, AffectedObjectID: target.ObjectID,
					Duration: game.DurationPermanent,
				})
			}
			sourceID, spell := castProductionBuyback(t, g, engine, registry, tc,
				game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
			engine.resolveTopOfStack(g, &TurnLog{})
			if !spell.BuybackPaid || !g.Players[game.Player1].Hand.Contains(sourceID) ||
				g.Players[game.Player1].Graveyard.Contains(sourceID) {
				t.Fatal("legal Pulse resolution did not apply its paid buyback replacement")
			}
			switch scenario {
			case "indestructible":
				if _, live := permanentByObjectID(g, target.ObjectID); !live {
					t.Fatal("Pulse destroyed an indestructible artifact")
				}
			case "shield":
				if _, live := permanentByObjectID(g, target.ObjectID); !live ||
					target.Counters.Get(counter.Shield) != 0 {
					t.Fatal("Pulse failed to consume the shield instead of destroying its artifact")
				}
			case "regeneration":
				if _, live := permanentByObjectID(g, target.ObjectID); !live ||
					target.RegenerationShields != 0 || !target.Tapped || target.MarkedDamage != 0 {
					t.Fatal("Pulse failed to apply the regeneration replacement")
				}
			case "graveyard redirected":
				if _, live := permanentByObjectID(g, target.ObjectID); live ||
					!g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
					t.Fatal("Pulse bypassed the artifact's graveyard replacement")
				}
			}
		})
	}

	t.Run("artifact type is rechecked at resolution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Artifact before type effect", Types: []types.Card{types.Artifact, types.Creature},
		}})
		sourceID, spell := castProductionBuyback(t, g, engine, registry, tc,
			game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType,
			SetTypes: []types.Card{types.Land}, Duration: game.DurationPermanent,
		})
		engine.resolveTopOfStack(g, &TurnLog{})
		if !spell.BuybackPaid || !g.Players[game.Player1].Graveyard.Contains(sourceID) ||
			g.Players[game.Player1].Hand.Contains(sourceID) ||
			permanentByCardID(g, target.CardInstanceID) == nil {
			t.Fatal("all-illegal Pulse did not use its normal destination or changed the target")
		}
	})

	t.Run("retaining artifact type keeps the target legal", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Artifact that gains a land type", Types: []types.Card{types.Artifact, types.Creature},
		}})
		sourceID, _ := castProductionBuyback(t, g, engine, registry, tc,
			game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType,
			SetTypes: []types.Card{types.Artifact, types.Land}, Duration: game.DurationPermanent,
		})
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
			!g.Players[game.Player1].Hand.Contains(sourceID) {
			t.Fatal("retaining artifact type did not keep the target legal and return paid Pulse")
		}
	})
}

func TestBuybackValueShatteringPulseCounteredAndIllegalTargets(t *testing.T) {
	t.Parallel()
	t.Run("nonartifact is rejected atomically by live and Simulator", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		tc := buybackValueCase(t, authored.ShatteringPulse.OracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Not an artifact", Types: []types.Card{types.Enchantment},
		}})
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
			game.PhaseCombat, game.StepDeclareBlockers, game.Player1
		act := action.CastSpell(sourceID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
		before := g.Clone()
		manaBefore := g.Players[game.Player1].ManaPool.Total()
		if engine.applyAction(g, game.Player1, act) {
			t.Fatal("Shattering Pulse accepted a nonartifact target")
		}
		if branch, accepted := engine.Simulator().Apply(before, game.Player1, act, simPassPolicies()); accepted || branch != nil {
			t.Fatal("Simulator accepted the nonartifact target")
		}
		if !g.Players[game.Player1].Hand.Contains(sourceID) ||
			g.Players[game.Player1].ManaPool.Total() != manaBefore || g.Stack.Size() != 0 {
			t.Fatal("illegal target consumed Pulse, mana, or stack state")
		}
	})

	t.Run("countered paid Pulse goes to its graveyard", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		tc := buybackValueCase(t, authored.ShatteringPulse.OracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
			Name: "Countered artifact", Types: []types.Card{types.Artifact},
		}})
		sourceID, pulse := castProductionBuyback(t, g, engine, registry, tc,
			game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
		counterspell, err := registry.LookupChecked("Counterspell")
		if err != nil || counterspell == nil {
			t.Fatalf("production Counterspell: %v", err)
		}
		counterID := addCardToHand(g, game.Player2, counterspell)
		g.Players[game.Player2].ManaPool.Add(mana.U, 2)
		g.Turn.PriorityPlayer = game.Player2
		if !engine.applyAction(g, game.Player2,
			action.CastSpell(counterID, []game.Target{game.StackObjectTarget(pulse.ID)}, 0, nil)) {
			t.Fatal("real Counterspell could not counter paid Pulse")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Graveyard.Contains(sourceID) ||
			g.Players[game.Player1].Hand.Contains(sourceID) ||
			permanentByCardID(g, target.CardInstanceID) == nil {
			t.Fatal("countered Pulse returned to hand or destroyed its target")
		}
	})
}

func TestBuybackValueShatteringPulseDeferredManaAndClone(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	tc := buybackValueCase(t, authored.ShatteringPulse.OracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mountainDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Payment Mountain", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Mountain},
	}}
	mountains := make([]*game.Permanent, 4)
	for i := range mountains {
		mountains[i] = addCombatPermanent(g, game.Player1, mountainDef)
	}
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Deferred-payment artifact", Types: []types.Card{types.Artifact},
	}})
	sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	act := action.CastSpell(sourceID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	agents := buybackInteractionChoiceAgent(t, game.Player1, true)
	before := g.Clone()
	simulated, accepted := engine.Simulator().Apply(before, game.Player1, act, agents)
	if !accepted || simulated == nil {
		t.Fatal("Simulator rejected Pulse with available deferred mana production")
	}
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("real Pulse cast did not activate mana abilities to pay buyback")
	}
	spell, _ := g.Stack.Peek()
	simulatedSpell, _ := simulated.Stack.Peek()
	if !spell.BuybackPaid || spell.ManaSpentToCast != 5 ||
		!simulatedSpell.BuybackPaid || simulatedSpell.ManaSpentToCast != spell.ManaSpentToCast ||
		g.Players[game.Player1].ManaPool.Total() != 0 ||
		simulated.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("deferred mana production did not pay Pulse's exact base and buyback costs")
	}
	for _, mountain := range mountains {
		if !mountain.Tapped {
			t.Fatal("a required Mountain mana ability was not activated during payment")
		}
	}
	clone := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(clone, &TurnLog{})
	simulated = engine.Simulator().ResolvePriority(simulated, simPassPolicies())
	for name, result := range map[string]*game.Game{"live": g, "clone": clone, "Simulator": simulated} {
		if !result.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
			!result.Players[game.Player1].Hand.Contains(sourceID) ||
			result.Stack.Size() != 0 {
			t.Errorf("%s did not destroy the artifact and return paid Pulse", name)
		}
	}
}

func TestBuybackValueWhispersDrawControllerAndClone(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	tc := buybackValueCase(t, authored.WhispersOfTheMuse.OracleID)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	controllerCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Controller's top card"}})
	opponentCard := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent's top card"}})
	sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
	addBuybackInteractionMana(g, tc, true, true)
	setMainPhasePriority(g, game.Player1)
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	act := action.CastSpell(sourceID, nil, 0, nil)
	agents := buybackInteractionChoiceAgent(t, game.Player1, true)
	before := g.Clone()
	simulated, accepted := engine.Simulator().Apply(before, game.Player1, act, agents)
	if !accepted || simulated == nil {
		t.Fatal("Simulator rejected targetless paid Whispers")
	}
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("real Whispers cast failed")
	}
	spell, _ := g.Stack.Peek()
	simulatedSpell, _ := simulated.Stack.Peek()
	if spell == nil || !spell.BuybackPaid || simulatedSpell == nil ||
		!simulatedSpell.BuybackPaid || simulatedSpell.ManaSpentToCast != spell.ManaSpentToCast {
		t.Fatal("Whispers did not retain its exact paid buyback announcement")
	}
	clone := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(clone, &TurnLog{})
	simulated = engine.Simulator().ResolvePriority(simulated, simPassPolicies())
	for name, result := range map[string]*game.Game{"live": g, "clone": clone, "Simulator": simulated} {
		if !result.Players[game.Player1].Hand.Contains(sourceID) ||
			!result.Players[game.Player1].Hand.Contains(controllerCard) ||
			result.Players[game.Player1].Library.Contains(controllerCard) ||
			!result.Players[game.Player2].Library.Contains(opponentCard) ||
			result.Players[game.Player2].Hand.Contains(opponentCard) ||
			result.Stack.Size() != 0 {
			t.Errorf("%s drew for the wrong player or lost Whispers' successful buyback", name)
		}
	}
	if countEventsOfKind(g.Events, game.EventCardDrawn) != 1 ||
		countEventsOfKind(g.Events, game.EventCardRevealed) != 0 {
		t.Fatal("ordinary private draw emitted the wrong public event")
	}
	var drawEvent game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCardDrawn {
			drawEvent = event
			break
		}
	}
	if drawEvent.Player != game.Player1 || drawEvent.Controller != game.Player1 || drawEvent.CardID != controllerCard {
		t.Fatal("Whispers did not draw exactly the controller's top card")
	}
	opponentView := observe(g, game.Player2)
	if len(opponentView.HandOf(game.Player1)) != 0 ||
		len(opponentView.LibraryLookable(game.Player1)) != 0 {
		t.Fatal("Whispers exposed the controller's private hand or library")
	}
	if _, revealed := opponentView.LibraryTopRevealed(game.Player1); revealed {
		t.Fatal("Whispers revealed the controller's library top")
	}
}

func TestBuybackValueWhispersReplacementAndEmptyLibrary(t *testing.T) {
	t.Parallel()
	t.Run("replacement draw uses its chosen card and still buys back", func(t *testing.T) {
		g, engine, registry := abundanceGame(t)
		library := abundanceLibrary(g, "LandTop", "NonlandChoice")
		tc := buybackValueCase(t, authored.WhispersOfTheMuse.OracleID)
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
			game.PhaseCombat, game.StepDeclareBlockers, game.Player1
		if !engine.applyActionWithChoices(g, game.Player1,
			action.CastSpell(sourceID, nil, 0, nil),
			buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
			t.Fatal("Whispers with Abundance replacement failed to cast")
		}
		chooser := &abundanceChoices{kinds: []int{1}}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: chooser}, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(sourceID) ||
			!g.Players[game.Player1].Hand.Contains(library[1]) ||
			g.Players[game.Player1].Hand.Contains(library[0]) ||
			!g.Players[game.Player1].Library.Contains(library[0]) ||
			len(abundanceEvents(g, game.EventCardDrawn)) != 0 ||
			!slices.Equal(abundanceEvents(g, game.EventCardRevealed), library) {
			t.Fatalf("Whispers bypassed the replacement draw or lost its successful buyback: hand=%v library=%v revealed=%v drawn=%v source=%v",
				g.Players[game.Player1].Hand.All(), g.Players[game.Player1].Library.All(),
				abundanceEvents(g, game.EventCardRevealed), abundanceEvents(g, game.EventCardDrawn), sourceID)
		}
	})

	t.Run("empty library loses at the next state-based-action opportunity", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		tc := buybackValueCase(t, authored.WhispersOfTheMuse.OracleID)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
			game.PhaseCombat, game.StepDeclareBlockers, game.Player1
		if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(sourceID, nil, 0, nil),
			buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
			t.Fatal("empty-library Whispers failed to cast")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.FailedDraws[game.Player1] || g.Players[game.Player1].Eliminated ||
			!g.Players[game.Player1].Hand.Contains(sourceID) ||
			g.Players[game.Player1].Graveyard.Contains(sourceID) || !g.Stack.IsEmpty() {
			t.Fatal("empty draw prevented spell resolution/buyback or caused loss before priority")
		}
		engine.applyStateBasedActions(g)
		if !g.Players[game.Player1].Eliminated {
			t.Fatal("failed empty-library draw did not lose at the next priority opportunity")
		}
	})
}

func TestBuybackValueDisturbedBurialOwnCreatureTargetsAndTiming(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	tc := buybackValueCase(t, authored.DisturbedBurial.OracleID)
	for _, scenario := range []string{"opponent graveyard", "own noncreature"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var targetID id.ID
			if scenario == "opponent graveyard" {
				targetID = addCardToGraveyard(g, game.Player2, vanillaCreature("Opponent creature", 2, 2))
			} else {
				targetID = addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Own artifact card", Types: []types.Card{types.Artifact},
				}})
			}
			sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
			addBuybackInteractionMana(g, tc, true, true)
			setMainPhasePriority(g, game.Player1)
			act := action.CastSpell(sourceID, []game.Target{currentCardTarget(t, g, targetID)}, 0, nil)
			before := g.Clone()
			manaBefore := g.Players[game.Player1].ManaPool.Total()
			if engine.applyAction(g, game.Player1, act) {
				t.Fatal("Disturbed Burial accepted a target outside its own creature cards")
			}
			if branch, accepted := engine.Simulator().Apply(before, game.Player1, act, simPassPolicies()); accepted || branch != nil {
				t.Fatal("Simulator accepted an illegal Disturbed Burial target")
			}
			targetOwner := game.Player1
			if scenario == "opponent graveyard" {
				targetOwner = game.Player2
			}
			if !g.Players[game.Player1].Hand.Contains(sourceID) ||
				g.Players[game.Player1].ManaPool.Total() != manaBefore || !g.Players[targetOwner].Graveyard.Contains(targetID) {
				t.Fatal("illegal Burial target consumed resources or moved a card")
			}
		})
	}

	t.Run("sorcery timing and legal paid resolution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		targetID := addCardToGraveyard(g, game.Player1, vanillaCreature("Own creature", 2, 2))
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		target := currentCardTarget(t, g, targetID)
		act := action.CastSpell(sourceID, []game.Target{target}, 0, nil)
		setMainPhasePriority(g, game.Player1)
		g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
		manaBefore := g.Players[game.Player1].ManaPool.Total()
		if engine.applyAction(g, game.Player1, act) {
			t.Fatal("Disturbed Burial was cast as an instant")
		}
		if !g.Players[game.Player1].Hand.Contains(sourceID) ||
			g.Players[game.Player1].ManaPool.Total() != manaBefore {
			t.Fatal("invalid sorcery timing consumed the card or mana")
		}
		setMainPhasePriority(g, game.Player1)
		before := g.Clone()
		agents := buybackInteractionChoiceAgent(t, game.Player1, true)
		simulated, accepted := engine.Simulator().Apply(before, game.Player1, act, agents)
		if !accepted || simulated == nil {
			t.Fatal("Simulator rejected a legal sorcery-speed Burial")
		}
		if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
			t.Fatal("Disturbed Burial did not cast in its controller's main phase")
		}
		clone := g.Clone()
		engine.resolveTopOfStack(g, &TurnLog{})
		engine.resolveTopOfStack(clone, &TurnLog{})
		simulated = engine.Simulator().ResolvePriority(simulated, simPassPolicies())
		for name, result := range map[string]*game.Game{"live": g, "clone": clone, "Simulator": simulated} {
			if !result.Players[game.Player1].Hand.Contains(targetID) ||
				!result.Players[game.Player1].Hand.Contains(sourceID) ||
				result.Stack.Size() != 0 {
				t.Errorf("%s failed to return the own creature and paid Burial", name)
			}
		}
	})
}

func TestBuybackValueDisturbedBurialIncarnationsAndReplacement(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	tc := buybackValueCase(t, authored.DisturbedBurial.OracleID)
	t.Run("leave and return creates a new illegal target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		targetID := addCardToGraveyard(g, game.Player1, vanillaCreature("Incarnation target", 2, 2))
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		if !engine.applyActionWithChoices(g, game.Player1,
			action.CastSpell(sourceID, []game.Target{currentCardTarget(t, g, targetID)}, 0, nil),
			buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
			t.Fatal("valid own graveyard creature could not be targeted")
		}
		oldVersion := g.CardInstances[targetID].ZoneVersion
		if !moveCardBetweenZones(g, game.Player1, targetID, zone.Graveyard, zone.Exile) ||
			!moveCardBetweenZones(g, game.Player1, targetID, zone.Exile, zone.Graveyard) {
			t.Fatal("fixture could not remove and return the target card")
		}
		if g.CardInstances[targetID].ZoneVersion == oldVersion {
			t.Fatal("target did not acquire a new zone incarnation")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Graveyard.Contains(targetID) ||
			!g.Players[game.Player1].Graveyard.Contains(sourceID) ||
			g.Players[game.Player1].Hand.Contains(sourceID) {
			t.Fatal("Burial followed the old card incarnation or applied buyback to an all-illegal spell")
		}
	})

	t.Run("creature type is rechecked at resolution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		targetID := addCardToGraveyard(g, game.Player1, vanillaCreature("Changing target", 2, 2))
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		if !engine.applyActionWithChoices(g, game.Player1,
			action.CastSpell(sourceID, []game.Target{currentCardTarget(t, g, targetID)}, 0, nil),
			buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
			t.Fatal("valid creature target could not be cast")
		}
		g.CardInstances[targetID].Def.CardFace.Types = []types.Card{types.Artifact}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Graveyard.Contains(targetID) ||
			!g.Players[game.Player1].Graveyard.Contains(sourceID) ||
			g.Players[game.Player1].Hand.Contains(sourceID) {
			t.Fatal("Burial ignored the changed target type or returned after failing to resolve")
		}
	})

	t.Run("hand replacement redirects only the returned target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		targetID := addCardToGraveyard(g, game.Player1, vanillaCreature("Redirected target", 2, 2))
		sourceID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Graveyard,
			MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
			AffectedCardID: targetID, Duration: game.DurationPermanent,
		})
		addBuybackInteractionMana(g, tc, true, true)
		setMainPhasePriority(g, game.Player1)
		if !engine.applyActionWithChoices(g, game.Player1,
			action.CastSpell(sourceID, []game.Target{currentCardTarget(t, g, targetID)}, 0, nil),
			buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
			t.Fatal("valid replacement-bound target could not be cast")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Exile.Contains(targetID) ||
			g.Players[game.Player1].Hand.Contains(targetID) ||
			!g.Players[game.Player1].Hand.Contains(sourceID) {
			t.Fatal("hand replacement or independent buyback destination was bypassed")
		}
	})
}

func TestBuybackValueDisturbedBurialCopyRetargetsForController(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	playerOneCreature := addCardToGraveyard(g, game.Player1, vanillaCreature("Player one's creature", 2, 2))
	playerTwoCreature := addCardToGraveyard(g, game.Player2, vanillaCreature("Player two's creature", 2, 2))
	tc := buybackValueCase(t, authored.DisturbedBurial.OracleID)
	burialID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, tc))
	addBuybackInteractionMana(g, tc, true, true)
	setMainPhasePriority(g, game.Player1)
	burialAction := action.CastSpell(burialID,
		[]game.Target{currentCardTarget(t, g, playerOneCreature)}, 0, nil)
	if !engine.applyActionWithChoices(g, game.Player1, burialAction,
		buybackInteractionChoiceAgent(t, game.Player1, true), &TurnLog{}) {
		t.Fatal("original paid Burial could not target its controller's creature")
	}
	burial, _ := g.Stack.Peek()
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player2
	reiterate, err := registry.LookupOracleID(authored.Reiterate.OracleID)
	if err != nil || reiterate == nil {
		t.Fatalf("production Reiterate: %v", err)
	}
	reiterateID := addCardToHand(g, game.Player2, reiterate)
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	g.Players[game.Player2].ManaPool.Add(mana.R, 2)
	if !engine.applyActionWithChoices(g, game.Player2,
		action.CastSpell(reiterateID, []game.Target{game.StackObjectTarget(burial.ID)}, 0, nil),
		buybackInteractionChoiceAgent(t, game.Player2, false), &TurnLog{}) {
		t.Fatal("Player Two could not cast Reiterate targeting Burial")
	}
	newTarget := currentCardTarget(t, g, playerTwoCreature)
	copyChoices := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		return chooseAuthoredTargets(t, request, newTarget)
	}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: copyChoices}, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.Controller != game.Player2 || !copied.BuybackPaid ||
		len(copied.Targets) != 1 || copied.Targets[0].CardID != playerTwoCreature {
		t.Fatal("copy did not retarget Burial to its new controller's own graveyard")
	}
	clone := g.Clone()
	for _, result := range []*game.Game{g, clone} {
		engine.resolveTopOfStack(result, &TurnLog{})
		if !result.Players[game.Player2].Hand.Contains(playerTwoCreature) ||
			result.Players[game.Player1].Hand.Contains(playerTwoCreature) ||
			result.Players[game.Player1].Graveyard.Contains(burialID) ||
			result.Players[game.Player1].Hand.Contains(burialID) {
			t.Fatal("retargeted copy moved the original physical Burial or returned the wrong player's card")
		}
		engine.resolveTopOfStack(result, &TurnLog{})
		if !result.Players[game.Player1].Hand.Contains(playerOneCreature) ||
			!result.Players[game.Player1].Hand.Contains(burialID) ||
			result.Players[game.Player2].Hand.Contains(playerOneCreature) ||
			!result.Players[game.Player2].Hand.Contains(playerTwoCreature) {
			t.Fatal("original paid Burial or copied controller's target resolved to the wrong hand")
		}
	}
	if !g.Players[game.Player2].Graveyard.Contains(reiterateID) {
		t.Fatal("normal Reiterate did not use its ordinary graveyard destination")
	}
}

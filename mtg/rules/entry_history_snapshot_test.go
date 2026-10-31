package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Fixer, Techno Terror (MSC 656 printing
// 3f888843-4bdb-438d-8fb9-968edb2713ac, Oracle
// a47b71df-9357-4118-90c0-f86fca6d362b) "Activate only if an artifact entered
// the battlefield under your control this turn"; its Scryfall rulings list was
// empty. Éowyn, Shieldmaiden (LTC 1 printing
// dd4a00ff-2206-4e12-a0ab-61ed82c9e6c5, Oracle
// 98a2a642-a30a-43e7-b47f-4776c22d1cb0) "if another Human entered the
// battlefield under your control this turn"; its 2023-06-16 ruling requires the
// Human to have entered before the beginning of combat. The Armory Mice
// (Celebration) ruling says entered permanents "don't need to remain on the
// battlefield or under your control" because these abilities check past events.
// CR 2026-09-25 from the verified official copy: a permanent's controller is the
// player under whose control it entered (110.2), and trigger events see objects
// as they exist immediately after the event (603.6a, 603.10). Later control,
// type, face-up or zone changes do not rewrite what entered.
const (
	entryHistoryFixerAbilityIndex  = 0
	entryHistoryEowynTriggerIndex  = 0
	entryHistoryFaceUpMorphCostOne = 1
)

// entryHistoryRegistryCard loads a legacy name-only registry entry; the exact
// Oracle identities are recorded in the source review above.
func entryHistoryRegistryCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	defs, err := cards.NewDefaultRegistry().LookupAllChecked(name)
	if err != nil {
		t.Fatalf("%s default-registry lookup failed: %v", name, err)
	}
	if len(defs) != 1 || defs[0].Name != name {
		t.Fatalf("%s lookup returned %d definitions", name, len(defs))
	}
	return defs[0]
}

func entryHistoryGame(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fixer := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Fixer, Techno Terror"))
	eowyn := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, NewEngine(nil), fixer, eowyn
}

func enterEntryHistoryCard(t *testing.T, g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, controller, def)
	card, _ := g.GetCardInstance(cardID)
	g.Players[controller].Hand.Remove(cardID)
	permanent, ok := createCardPermanent(g, card, controller, zone.Hand)
	if !ok {
		t.Fatalf("%s did not enter the battlefield", def.Name)
	}
	return permanent
}

func gainEntryHistoryControl(g *game.Game, permanent *game.Permanent, controller game.PlayerID) {
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: permanent.ObjectID,
		Layer: game.LayerControl, NewController: opt.Val(controller), Duration: game.DurationPermanent,
	})
}

func changeEntryHistoryTypes(g *game.Game, permanent *game.Permanent, effect game.ContinuousEffect) {
	effect.ID = g.IDGen.Next()
	effect.Timestamp = game.Timestamp(g.IDGen.Next())
	effect.AffectedObjectID = permanent.ObjectID
	effect.Layer = game.LayerType
	effect.Duration = game.DurationPermanent
	g.ContinuousEffects = append(g.ContinuousEffects, effect)
}

func fixerEntryHistoryAllowed(t *testing.T, g *game.Game, engine *Engine, fixer *game.Permanent) bool {
	t.Helper()
	def, ok := permanentCardDef(g, fixer)
	if !ok {
		t.Fatal("Fixer has no card definition")
	}
	ability := def.ActivatedAbilities[entryHistoryFixerAbilityIndex]
	condition := activationConditionSatisfied(g, game.Player1, fixer, ability.ActivationCondition)
	listed := containsAction(engine.legalActions(g, game.Player1),
		action.ActivateAbility(fixer.ObjectID, entryHistoryFixerAbilityIndex, nil, 0))
	if condition != listed {
		t.Fatalf("Fixer activation condition = %t but legal action listed = %t", condition, listed)
	}
	return condition
}

func eowynTriggersAtCombat(t *testing.T, g *game.Game, eowyn *game.Permanent) bool {
	t.Helper()
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	event := g.Events[len(g.Events)-1]
	if event.Kind != game.EventBeginningOfStep || !event.TriggeredAbilitiesCaptured {
		t.Fatalf("beginning of combat event = %+v, want captured beginning-of-step event", event)
	}
	for _, triggered := range event.TriggeredAbilities {
		if triggered.SourceID == eowyn.ObjectID && triggered.AbilityIndex == entryHistoryEowynTriggerIndex {
			return true
		}
	}
	return false
}

func entryHistoryArtifact() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Entry History Relic", Types: []types.Card{types.Artifact}}}
}

func entryHistoryCreature(name string, subtypes ...types.Sub) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  subtypes,
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
}

func TestEntryHistoryUsesControllerUnderWhichPermanentEntered(t *testing.T) {
	t.Parallel()
	g, engine, fixer, eowyn := entryHistoryGame(t)

	stolenArtifact := enterEntryHistoryCard(t, g, game.Player2, entryHistoryArtifact())
	stolenHuman := enterEntryHistoryCard(t, g, game.Player2, entryHistoryCreature("Entry History Squire", types.Human))
	gainEntryHistoryControl(g, stolenArtifact, game.Player1)
	gainEntryHistoryControl(g, stolenHuman, game.Player1)
	if effectiveController(g, stolenArtifact) != game.Player1 || effectiveController(g, stolenHuman) != game.Player1 {
		t.Fatal("control-changing effects did not give Player1 the entered permanents")
	}
	if fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("artifact that entered under Player2's control satisfied Fixer after Player1 gained control of it")
	}
	if eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Human that entered under Player2's control triggered Éowyn after Player1 gained control of it")
	}

	ownArtifact := enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
	ownHuman := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Page", types.Human))
	gainEntryHistoryControl(g, ownArtifact, game.Player2)
	gainEntryHistoryControl(g, ownHuman, game.Player2)
	if effectiveController(g, ownArtifact) != game.Player2 || effectiveController(g, ownHuman) != game.Player2 {
		t.Fatal("control-changing effects did not give Player2 the entered permanents")
	}
	if !fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("artifact that entered under Player1's control stopped satisfying Fixer after Player2 gained control of it")
	}
	if !eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Human that entered under Player1's control stopped triggering Éowyn after Player2 gained control of it")
	}
}

func TestEntryHistoryUsesCharacteristicsPermanentEnteredWith(t *testing.T) {
	t.Parallel()
	g, engine, fixer, eowyn := entryHistoryGame(t)

	land := enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Entry History Wastes", Types: []types.Card{types.Land},
	}})
	wolf := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Wolf", types.Wolf))
	changeEntryHistoryTypes(g, land, game.ContinuousEffect{AddTypes: []types.Card{types.Artifact, types.Creature}})
	changeEntryHistoryTypes(g, wolf, game.ContinuousEffect{AddSubtypes: []types.Sub{types.Human}})
	if !permanentHasType(g, land, types.Artifact) || !permanentHasSubtype(g, wolf, types.Human) {
		t.Fatal("type-changing effects did not apply to the entered permanents")
	}
	if fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("land that later became an artifact satisfied Fixer's artifact-entered history")
	}
	if eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Wolf that later became a Human triggered Éowyn's Human-entered history")
	}

	artifact := enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
	human := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Squire", types.Human))
	changeEntryHistoryTypes(g, artifact, game.ContinuousEffect{SetTypes: []types.Card{types.Land}})
	changeEntryHistoryTypes(g, human, game.ContinuousEffect{SetSubtypes: []types.Sub{types.Wolf}})
	if permanentHasType(g, artifact, types.Artifact) || permanentHasSubtype(g, human, types.Human) {
		t.Fatal("type-setting effects did not remove the entered characteristics")
	}
	if !fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("artifact that later stopped being an artifact no longer satisfied Fixer")
	}
	if !eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Human that later stopped being a Human no longer triggered Éowyn")
	}
}

func TestEntryHistoryCountsPermanentAfterItLeaves(t *testing.T) {
	t.Parallel()
	g, engine, fixer, eowyn := entryHistoryGame(t)

	land := enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Entry History Wastes", Types: []types.Card{types.Land},
	}})
	wolf := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Wolf", types.Wolf))
	changeEntryHistoryTypes(g, land, game.ContinuousEffect{AddTypes: []types.Card{types.Artifact}})
	changeEntryHistoryTypes(g, wolf, game.ContinuousEffect{AddSubtypes: []types.Sub{types.Human}})
	if !movePermanentToZone(g, land, zone.Graveyard) || !movePermanentToZone(g, wolf, zone.Graveyard) {
		t.Fatal("type-changed permanents did not leave the battlefield")
	}
	if fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("land that became an artifact before leaving satisfied Fixer from last-known information")
	}
	if eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Wolf that became a Human before leaving triggered Éowyn from last-known information")
	}

	artifact := enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
	human := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Squire", types.Human))
	gainEntryHistoryControl(g, artifact, game.Player2)
	gainEntryHistoryControl(g, human, game.Player2)
	if !movePermanentToZone(g, artifact, zone.Graveyard) || !movePermanentToZone(g, human, zone.Graveyard) {
		t.Fatal("entered permanents did not leave the battlefield")
	}
	if !fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("artifact that entered under Player1's control and then left no longer satisfied Fixer")
	}
	if !eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Human that entered under Player1's control and then left no longer triggered Éowyn")
	}
}

func TestEntryHistoryFaceDownEntryKeepsFaceDownCharacteristics(t *testing.T) {
	t.Parallel()
	g, engine, fixer, eowyn := entryHistoryGame(t)
	pt := game.PT{Value: 3}
	hidden := addFaceDownPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Entry History Automaton",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.MorphKeyword{Cost: cost.Mana{cost.O(entryHistoryFaceUpMorphCostOne)}}},
		}},
	}}, game.FaceDownMorph)
	entry := g.Events[len(g.Events)-1]
	if entry.Kind != game.EventPermanentEnteredBattlefield || entry.PermanentID != hidden.ObjectID ||
		entry.EnteredPermanent == nil || !entry.EnteredPermanent.FaceDown ||
		!slices.Equal(entry.EnteredPermanent.Types, []types.Card{types.Creature}) || len(entry.EnteredPermanent.Subtypes) != 0 {
		t.Fatalf("face-down entry event = %+v, want a face-down 2/2 creature snapshot", entry)
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, entryHistoryFaceUpMorphCostOne)
	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(hidden.ObjectID)) {
		t.Fatal("turn face-up action failed")
	}
	if hidden.FaceDown || !permanentHasType(g, hidden, types.Artifact) || !permanentHasSubtype(g, hidden, types.Human) {
		t.Fatal("turned-up permanent did not reveal its artifact Human characteristics")
	}
	if fixerEntryHistoryAllowed(t, g, engine, fixer) {
		t.Fatal("face-down entry satisfied Fixer with the artifact type it revealed later")
	}
	if eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("face-down entry triggered Éowyn with the Human subtype it revealed later")
	}
}

func TestEntryHistorySnapshotDoesNotChangeLiveEnterTriggers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	watcher := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Entry History Watcher",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever an artifact you control enters, draw a card.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
			}},
		}},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Entry History Foundry",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer:    game.LayerType,
			AddTypes: []types.Card{types.Artifact},
			Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
		}}}},
	}})

	land := enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Entry History Wastes", Types: []types.Card{types.Land},
	}})
	entry := g.Events[len(g.Events)-1]
	if entry.Kind != game.EventPermanentEnteredBattlefield || entry.PermanentID != land.ObjectID || entry.EnteredPermanent == nil {
		t.Fatalf("last event = %+v, want the land's captured entry event", entry)
	}
	if entry.EnteredPermanent.Controller != game.Player1 || !slices.Contains(entry.EnteredPermanent.Types, types.Artifact) {
		t.Fatalf("entry snapshot = %+v, want Player1 artifact land as it entered", entry.EnteredPermanent)
	}
	triggered := false
	for _, ability := range entry.TriggeredAbilities {
		triggered = triggered || ability.SourceID == watcher.ObjectID
	}
	if !triggered {
		t.Fatal("land that entered as an artifact did not trigger the live artifact-enters watcher")
	}
}

func entryHistoryEnteredEvent(t *testing.T, g *game.Game, permanent *game.Permanent) game.Event {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
			if event.EnteredPermanent == nil {
				t.Fatalf("entry event for %d has no entry snapshot", permanent.ObjectID)
			}
			return event
		}
	}
	t.Fatalf("no entry event for %d", permanent.ObjectID)
	return game.Event{}
}

func TestEntryHistoryControllerAppliesToEverySubjectNamingEnteredPermanent(t *testing.T) {
	t.Parallel()
	for _, subject := range []game.TriggerSubjectObject{game.TriggerSubjectDefault, game.TriggerSubjectPermanent} {
		g, _, fixer, _ := entryHistoryGame(t)
		pattern := game.TriggerPattern{
			Event:            game.EventPermanentEnteredBattlefield,
			Controller:       game.TriggerControllerYou,
			Subject:          subject,
			SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
		}

		ownArtifact := enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		stolenArtifact := enterEntryHistoryCard(t, g, game.Player2, entryHistoryArtifact())
		land := enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Entry History Wastes", Types: []types.Card{types.Land},
		}})
		gainEntryHistoryControl(g, ownArtifact, game.Player2)
		gainEntryHistoryControl(g, stolenArtifact, game.Player1)
		changeEntryHistoryTypes(g, land, game.ContinuousEffect{AddTypes: []types.Card{types.Artifact}})
		changeEntryHistoryTypes(g, ownArtifact, game.ContinuousEffect{SetTypes: []types.Card{types.Land}})

		ownEntry := entryHistoryEnteredEvent(t, g, ownArtifact)
		stolenEntry := entryHistoryEnteredEvent(t, g, stolenArtifact)
		landEntry := entryHistoryEnteredEvent(t, g, land)
		if !triggerMatchesEventForController(g, fixer, game.Player1, &pattern, ownEntry) {
			t.Errorf("subject %d: artifact that entered under Player1 stopped matching Player1 after Player2 gained control and it became a land", subject)
		}
		if triggerMatchesEventForController(g, fixer, game.Player2, &pattern, ownEntry) {
			t.Errorf("subject %d: Player2 matched Player1's artifact entry after gaining control", subject)
		}
		if !triggerMatchesEventForController(g, fixer, game.Player2, &pattern, stolenEntry) {
			t.Errorf("subject %d: artifact that entered under Player2 stopped matching Player2 after Player1 gained control", subject)
		}
		if triggerMatchesEventForController(g, fixer, game.Player1, &pattern, stolenEntry) {
			t.Errorf("subject %d: Player1 matched Player2's artifact entry after gaining control", subject)
		}
		if triggerMatchesEventForController(g, fixer, game.Player1, &pattern, landEntry) {
			t.Errorf("subject %d: land that later became an artifact matched an artifact entry", subject)
		}

		if !movePermanentToZone(g, ownArtifact, zone.Graveyard) {
			t.Fatal("entered artifact did not leave the battlefield")
		}
		if !triggerMatchesEventForController(g, fixer, game.Player1, &pattern, ownEntry) ||
			triggerMatchesEventForController(g, fixer, game.Player2, &pattern, ownEntry) {
			t.Errorf("subject %d: departed artifact entry did not keep its entry controller", subject)
		}
	}
}

// Source review: Grand Ball Guest (WOE 134 printing
// d6e75228-16af-42b0-8441-ed253a660cc9, Oracle
// 0b0fdb28-f7b1-4c30-9f2c-14a9074da22d), 2/2 "Celebration — This creature gets
// +1/+1 and has trample as long as two or more nonland permanents entered the
// battlefield under your control this turn." Its 2023-09-01 rulings say
// Celebration checks past events, not the current game state. Mentor of the Meek
// (HOC 170 printing 05b6a101-4d12-4f50-8f02-5c778e08b149, Oracle
// b9f4f96b-6e54-4fe6-8df7-623e0fc72409) "Whenever another creature you control
// with power 2 or less enters, you may pay {1}. If you do, draw a card." CR
// 603.6a checks every permanent, newcomers included, against each event that
// puts one or more permanents onto the battlefield; CR 603.10 uses objects and
// continuous effects as they exist immediately after the event. Immediately
// after the entry, that entry and any simultaneous entries have happened, so a
// Celebration creature entering as the second nonland permanent is already 3/3.
func entryHistoryCelebrationGuest(name string) *game.CardDef {
	def := entryHistoryCreature(name, types.Human, types.Peasant)
	def.Power = opt.Val(game.PT{Value: 2})
	def.Toughness = opt.Val(game.PT{Value: 2})
	def.StaticAbilities = []game.StaticAbility{{
		Condition: opt.Val(game.Condition{EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				SubjectSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
			},
			Window:   game.EventHistoryCurrentTurn,
			MinCount: 2,
		})}),
		ContinuousEffects: []game.ContinuousEffect{{AffectedSource: true, Layer: game.LayerPowerToughnessModify, PowerDelta: 1, ToughnessDelta: 1}},
	}}
	return def
}

func entryHistoryCelebrationGame(t *testing.T) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mentor := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Mentor of the Meek"))
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	return g, mentor
}

func entryHistoryEnteredPower(t *testing.T, g *game.Game, permanent *game.Permanent) int {
	t.Helper()
	snapshot := entryHistoryEnteredEvent(t, g, permanent).EnteredPermanent
	if !snapshot.Power.Exists {
		t.Fatalf("entry snapshot for %d has no power", permanent.ObjectID)
	}
	return snapshot.Power.Val
}

func entryHistoryMentorTriggered(t *testing.T, g *game.Game, mentor, entered *game.Permanent) bool {
	t.Helper()
	for _, trigger := range entryHistoryEnteredEvent(t, g, entered).TriggeredAbilities {
		if trigger.SourceID == mentor.ObjectID {
			return true
		}
	}
	return false
}

func entryHistoryPowerEnteredThisTurn(g *game.Game, mentor *game.Permanent, power int) bool {
	return conditionEventHistorySatisfied(g, conditionContext{controller: game.Player1, source: mentor}, &game.EventHistoryCondition{
		Pattern: game.TriggerPattern{
			Event:            game.EventPermanentEnteredBattlefield,
			Controller:       game.TriggerControllerYou,
			SubjectSelection: game.Selection{Power: opt.Val(compare.Int{Op: compare.Equal, Value: power})},
		},
		Window: game.EventHistoryCurrentTurn,
	})
}

func TestEntryHistorySnapshotSeesItsOwnEntry(t *testing.T) {
	t.Parallel()
	t.Run("second nonland entry", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry History Guest"))
		if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
			t.Fatalf("Guest entering as the second nonland permanent had entry power %d, want 3", got)
		}
		if entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek triggered for a Guest that entered with power 3")
		}
		if !movePermanentToZone(g, guest, zone.Graveyard) {
			t.Fatal("Guest did not leave the battlefield")
		}
		if !entryHistoryPowerEnteredThisTurn(g, mentor, 3) || entryHistoryPowerEnteredThisTurn(g, mentor, 2) {
			t.Fatal("entry history did not record the departed Guest as a 3-power entry")
		}
	})
	t.Run("first nonland entry", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Entry History Wastes", Types: []types.Card{types.Land},
		}})
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry History Guest"))
		if got := entryHistoryEnteredPower(t, g, guest); got != 2 {
			t.Fatalf("Guest entering after only a land had entry power %d, want 2", got)
		}
		if !entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek did not trigger for a Guest that entered with power 2")
		}
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		if got := effectivePermanentValues(g, guest).power; got != 3 {
			t.Fatalf("Guest power after a second nonland entry = %d, want 3", got)
		}
		if !entryHistoryPowerEnteredThisTurn(g, mentor, 2) || entryHistoryPowerEnteredThisTurn(g, mentor, 3) {
			t.Fatal("Guest's later Celebration bonus rewrote its 2-power entry")
		}
	})
	t.Run("opponent entry does not count", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		enterEntryHistoryCard(t, g, game.Player2, entryHistoryArtifact())
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry History Guest"))
		if got := entryHistoryEnteredPower(t, g, guest); got != 2 {
			t.Fatalf("Guest entering after an opponent's artifact had entry power %d, want 2", got)
		}
		if !entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek did not trigger for a Guest that entered with power 2")
		}
	})
}

func TestEntryHistorySnapshotSeesSimultaneousEntries(t *testing.T) {
	t.Parallel()
	t.Run("token batch", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		e := NewEngine(nil)
		guests, ok := createTokenPermanentsCollectingWithChoices(e, g, game.Player1, entryHistoryCelebrationGuest("Entry History Guest Token"),
			2, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !ok || len(guests) != 2 {
			t.Fatalf("created %d Guest tokens, want 2", len(guests))
		}
		for i, guest := range guests {
			if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
				t.Errorf("simultaneous Guest token %d had entry power %d, want 3", i, got)
			}
			if entryHistoryMentorTriggered(t, g, mentor, guest) {
				t.Errorf("Mentor of the Meek triggered for simultaneous Guest token %d", i)
			}
		}
	})
	t.Run("card batch", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		e := NewEngine(nil)
		var entries []preparedCardPermanentEntry
		for _, def := range []*game.CardDef{entryHistoryCelebrationGuest("Entry History Guest"), entryHistoryArtifact()} {
			cardID := addCardToHand(g, game.Player1, def)
			card := g.CardInstances[cardID]
			g.Players[game.Player1].Hand.Remove(cardID)
			entry, ok := prepareCardPermanentFaceForSimultaneousEntry(e, g, card, game.Player1, zone.Hand, game.FaceFront, nil,
				permanentCreationOptions{SimultaneousID: g.IDGen.Next()}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !ok {
				t.Fatalf("%s was not prepared for simultaneous entry", def.Name)
			}
			entries = append(entries, entry)
		}
		commitSimultaneousCardPermanentEntries(g, entries)
		guest := entries[0].permanent
		if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
			t.Fatalf("Guest entering simultaneously with an artifact had entry power %d, want 3", got)
		}
		if entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek triggered for a Guest that entered simultaneously with an artifact")
		}
		if len(g.PendingEntryEvents()) != 0 {
			t.Fatalf("pending entry overlay leaked after the batch: %d events", len(g.PendingEntryEvents()))
		}
	})
}

// creaturesAreLands adds a layer-4 effect making every creature a land in
// addition to its other types, so an entering Guest is a land creature (CR
// 613.1d, 205.1b).
func creaturesAreLands(g *game.Game) {
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), Layer: game.LayerType,
		Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		AddTypes: []types.Card{types.Land}, Duration: game.DurationPermanent,
	})
}

// A Celebration creature's own pending entry is matched with the
// characteristics it entered with, including layer-4 type changes already
// applied before its layer-7c bonus (CR 603.10, 611.3c, 613.1). A land creature
// is not a nonland permanent, so it does not count toward its own Celebration.
func TestEntryHistorySnapshotSeesEntryTimeTypeChanges(t *testing.T) {
	t.Parallel()
	t.Run("single entry", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		creaturesAreLands(g)
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry History Guest"))
		if !permanentHasType(g, guest, types.Land) {
			t.Fatal("Guest is not a land creature")
		}
		if got := entryHistoryEnteredPower(t, g, guest); got != 2 {
			t.Fatalf("land-creature Guest entering after one artifact had entry power %d, want 2", got)
		}
		if !entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek did not trigger for a Guest that entered with power 2")
		}
		if got := effectivePermanentValues(g, guest).power; got != 2 {
			t.Fatalf("land-creature Guest power = %d, want 2", got)
		}
		if !movePermanentToZone(g, guest, zone.Graveyard) {
			t.Fatal("Guest did not leave the battlefield")
		}
		if !entryHistoryPowerEnteredThisTurn(g, mentor, 2) || entryHistoryPowerEnteredThisTurn(g, mentor, 3) {
			t.Fatal("entry history did not record the departed Guest as a 2-power entry")
		}
	})
	t.Run("simultaneous entries", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		creaturesAreLands(g)
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		guests, ok := createTokenPermanentsCollectingWithChoices(NewEngine(nil), g, game.Player1,
			entryHistoryCelebrationGuest("Entry History Guest Token"), 2, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !ok || len(guests) != 2 {
			t.Fatalf("created %d Guest tokens, want 2", len(guests))
		}
		for i, guest := range guests {
			if got := entryHistoryEnteredPower(t, g, guest); got != 2 {
				t.Errorf("simultaneous land-creature Guest token %d had entry power %d, want 2", i, got)
			}
			if !entryHistoryMentorTriggered(t, g, mentor, guest) {
				t.Errorf("Mentor of the Meek did not trigger for simultaneous Guest token %d", i)
			}
		}
		if len(g.PendingEntryEvents()) != 0 {
			t.Fatalf("pending entry overlay leaked after the batch: %d events", len(g.PendingEntryEvents()))
		}
	})
	t.Run("type change on the other entrant", func(t *testing.T) {
		t.Parallel()
		g, mentor := entryHistoryCelebrationGame(t)
		creaturesAreLands(g)
		e := NewEngine(nil)
		var entries []preparedCardPermanentEntry
		for _, def := range []*game.CardDef{entryHistoryCelebrationGuest("Entry History Guest"), entryHistoryArtifact(), entryHistoryArtifact()} {
			cardID := addCardToHand(g, game.Player1, def)
			card := g.CardInstances[cardID]
			g.Players[game.Player1].Hand.Remove(cardID)
			entry, ok := prepareCardPermanentFaceForSimultaneousEntry(e, g, card, game.Player1, zone.Hand, game.FaceFront, nil,
				permanentCreationOptions{SimultaneousID: g.IDGen.Next()}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !ok {
				t.Fatalf("%s was not prepared for simultaneous entry", def.Name)
			}
			entries = append(entries, entry)
		}
		commitSimultaneousCardPermanentEntries(g, entries)
		guest := entries[0].permanent
		if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
			t.Fatalf("land-creature Guest entering with two artifacts had entry power %d, want 3", got)
		}
		if entryHistoryMentorTriggered(t, g, mentor, guest) {
			t.Fatal("Mentor of the Meek triggered for a Guest that entered with power 3")
		}
	})
}

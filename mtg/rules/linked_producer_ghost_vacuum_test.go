package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const vacuumProducer game.LinkedProducerKey = "test/ghost-vacuum/exile"

func ghostVacuumReturn() game.PutOnBattlefield {
	return game.PutOnBattlefield{
		Source:    game.OriginalSourceLinkedBattlefieldSource(vacuumProducer),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Recipient: opt.Val(game.ControllerReference()),
		EntryCounters: []game.CounterPlacement{{
			Kind: counter.Flying, Amount: 1,
		}},
		ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Spirit}, Duration: game.DurationPermanent},
			{Layer: game.LayerPowerToughnessSet, SetPower: opt.Val(game.PT{Value: 1}),
				SetToughness: opt.Val(game.PT{Value: 1}), Duration: game.DurationPermanent},
		},
	}
}

// This named fixture exercises shared operations, not a registered card implementation.
func ghostVacuumFixture(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	exile := game.ProgramReference{ID: "test/ghost-vacuum/exile", Version: "1"}
	release := game.ProgramReference{ID: "test/ghost-vacuum/return", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{
		{Reference: exile, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(game.MoveCard{
				Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard,
				Destination: zone.Exile, PublishLinkedProducer: vacuumProducer,
			})
			return err
		}},
		{Reference: release, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(ghostVacuumReturn())
			return err
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publish := game.Mode{
		Program: exile,
		Targets: []game.TargetSpec{{Allow: game.TargetAllowCard, TargetZone: zone.Graveyard,
			MinTargets: 1, MaxTargets: 1}},
	}.Ability()
	publish.LinkedProducers = []game.LinkedProducerKey{vacuumProducer}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Ghost Vacuum", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{
			{AdditionalCosts: cost.Tap, ZoneOfFunction: zone.Battlefield, Content: publish},
			{ManaCost: opt.Val(cost.Mana{cost.O(6)}), ZoneOfFunction: zone.Battlefield,
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}, {Kind: cost.AdditionalSacrificeSource}},
				Timing:          game.SorceryOnly, Content: game.Mode{Program: release}.Ability()},
		},
	}}
	if problems := game.ValidateCardDef(def); len(problems) != 0 {
		t.Fatalf("Ghost Vacuum shared-operation fixture: %+v", problems)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	source := addCombatPermanent(g, game.Player1, def)
	return g, NewEngineWithPrograms(nil, catalog), source
}

func activateVacuumExile(t *testing.T, g *game.Game, e *Engine, source *game.Permanent, cardID id.ID) {
	t.Helper()
	source.Tapped = false
	if !e.applyAction(g, source.Controller, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{currentCardTarget(t, g, cardID)}, 0)) {
		t.Fatal("graveyard exile activation failed")
	}
	e.resolveTopOfStack(g, &TurnLog{})
}

func activateVacuumReturn(t *testing.T, g *game.Game, e *Engine, source *game.Permanent) {
	t.Helper()
	source.Tapped = false
	g.Players[source.Controller].ManaPool.Add(mana.C, 6)
	if !e.applyAction(g, source.Controller, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("paid sacrifice activation failed")
	}
	if permanentByCardID(g, source.CardInstanceID) != nil {
		t.Fatal("activation did not sacrifice its original source")
	}
}

func assertVacuumSpirit(t *testing.T, g *game.Game, cardID id.ID, controller game.PlayerID, flying int) *game.Permanent {
	t.Helper()
	permanent := permanentByCardID(g, cardID)
	if permanent == nil || permanent.Controller != controller ||
		permanent.Counters.Get(counter.Flying) != flying || !hasKeyword(g, permanent, game.Flying) ||
		!permanentHasSubtype(g, permanent, types.Spirit) || effectivePower(g, permanent) != 1 {
		t.Fatalf("returned spirit = %+v, want controller %v, 1/1 Spirit with %d flying counters", permanent, controller, flying)
	}
	if toughness, ok := effectiveToughness(g, permanent); !ok || toughness != 1 {
		t.Fatalf("returned toughness=%d/%t, want 1/true", toughness, ok)
	}
	return permanent
}

func TestLinkedProducerGhostVacuumFilteredSimultaneousModifiedReturn(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	a := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Vacuum Elf", Types: []types.Card{types.Artifact, types.Creature},
		Subtypes: []types.Sub{types.Elf}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 5}),
	}})
	b := addCardToGraveyard(g, game.Player3, vanillaCreature("Vacuum Beast", 6, 7))
	observer := game.TriggeredAbility{
		Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield,
			SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Spirit},
				Power: opt.Val(compare.Int{Op: compare.Equal, Value: 1}), Keyword: game.Flying},
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(), Amount: game.Fixed(1),
		}}}}.Ability(),
	}
	g.CardInstances[a].Def.TriggeredAbilities = []game.TriggeredAbility{observer}
	g.CardInstances[b].Def.TriggeredAbilities = []game.TriggeredAbility{observer}
	noncreature := addCardToGraveyard(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
		Name: "Linked noncreature", Types: []types.Card{types.Artifact},
	}})
	unlinked := addCardInstance(g, game.Player2, vanillaCreature("Unlinked exile", 2, 2))
	g.Players[game.Player2].Exile.Add(unlinked)
	for _, cardID := range []id.ID{a, b, noncreature} {
		activateVacuumExile(t, g, e, source, cardID)
	}
	activateVacuumReturn(t, g, e, source)
	e.resolveTopOfStack(g, &TurnLog{})
	first := assertVacuumSpirit(t, g, a, game.Player1, 1)
	assertVacuumSpirit(t, g, b, game.Player1, 1)
	if !permanentHasType(g, first, types.Artifact) || !permanentHasSubtype(g, first, types.Elf) {
		t.Fatal("Spirit modification replaced the card's original types")
	}
	if !g.Players[game.Player4].Exile.Contains(noncreature) || !g.Players[game.Player2].Exile.Contains(unlinked) {
		t.Fatal("return included a noncreature or unlinked exile")
	}
	var batch game.ObjectID
	arrivals := 0
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield && (event.CardID == a || event.CardID == b) {
			arrivals++
			if event.SimultaneousID == 0 || batch != 0 && batch != event.SimultaneousID {
				t.Fatal("filtered returns did not share a simultaneous entry")
			}
			batch = event.SimultaneousID
		}
	}
	if arrivals != 2 {
		t.Fatalf("arrivals=%d, want exactly two creature cards", arrivals)
	}
	beforeLife := g.Players[game.Player1].Life
	if !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 4 {
		t.Fatalf("newcomers did not both see both modified arrivals: stack=%d", g.Stack.Size())
	}
	for g.Stack.Size() > 0 {
		e.resolveTopOfStack(g, &TurnLog{})
	}
	if g.Players[game.Player1].Life != beforeLife+4 {
		t.Fatal("entry triggers did not resolve with flying and 1/1 Spirit characteristics already present")
	}
	clone := g.Clone()
	cloned := permanentByCardID(clone, a)
	cloned.Counters.Add(counter.Flying, 1)
	clone.ContinuousEffects[0].AddSubtypes[0] = types.Zombie
	if first.Counters.Get(counter.Flying) != 1 || !permanentHasSubtype(g, first, types.Spirit) {
		t.Fatal("post-return clone aliased original counters or continuous effects")
	}
}

func TestLinkedProducerGhostVacuumEntryReplacementsSeeSpiritAndBasePower(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	a := addCardToGraveyard(g, game.Player2, vanillaCreature("Large captive", 7, 7))
	b := addCardToGraveyard(g, game.Player3, vanillaCreature("Other large captive", 5, 5))
	for _, cardID := range []id.ID{a, b} {
		activateVacuumExile(t, g, e, source, cardID)
	}
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventPermanentEnteredBattlefield,
		EntersTapped: true, EntersTappedOthers: true,
		EntersTappedSelection: &game.Selection{SubtypesAny: []types.Sub{types.Spirit},
			Power: opt.Val(compare.Int{Op: compare.Equal, Value: 1})},
	})
	addReplacementPermanent(t, g, game.Player1, doublingSeasonCounterReplacementCardDef())
	activateVacuumReturn(t, g, e, source)
	e.resolveTopOfStack(g, &TurnLog{})
	for _, cardID := range []id.ID{a, b} {
		if permanent := assertVacuumSpirit(t, g, cardID, game.Player1, 2); !permanent.Tapped {
			t.Fatal("entry replacement did not see simultaneous 1/1 Spirit characteristics")
		}
	}
	if len(g.ContinuousEffects) != 4 {
		t.Fatalf("initial continuous effects=%d, want two per entrant without duplicates", len(g.ContinuousEffects))
	}
}

func TestLinkedProducerGhostVacuumRejectsReplacedAndLaterExileIncarnations(t *testing.T) {
	for _, scenario := range []string{"replaced exile", "moved and reexiled", "illegal target", "replaced return", "replaced entry"} {
		t.Run(scenario, func(t *testing.T) {
			g, e, source := ghostVacuumFixture(t)
			cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Unavailable captive", 4, 4))
			if scenario == "replaced exile" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Hand,
				})
			}
			if scenario == "illegal target" {
				if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
					[]game.Target{currentCardTarget(t, g, cardID)}, 0)) {
					t.Fatal("targeted activation failed")
				}
				moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Hand)
				moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Graveyard)
				e.resolveTopOfStack(g, &TurnLog{})
			} else {
				activateVacuumExile(t, g, e, source, cardID)
			}
			if scenario == "moved and reexiled" {
				moveCardBetweenZones(g, game.Player2, cardID, zone.Exile, zone.Hand)
				moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Exile)
			}
			if scenario == "replaced return" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Exile,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Graveyard,
				})
			}
			if scenario == "replaced entry" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventPermanentEnteredBattlefield, EntersTapped: true,
					ReplaceToZone: zone.Graveyard,
				})
			}
			activateVacuumReturn(t, g, e, source)
			e.resolveTopOfStack(g, &TurnLog{})
			if permanentByCardID(g, cardID) != nil || len(g.ContinuousEffects) != 0 {
				t.Fatal("an unavailable captive entered or acquired orphaned initial effects")
			}
			if (scenario == "replaced return" || scenario == "replaced entry") && !g.Players[game.Player2].Graveyard.Contains(cardID) {
				t.Fatal("return ignored its replacement destination")
			}
			key := game.LinkedObjectKey{SourceID: source.ObjectID, LinkID: string(vacuumProducer), Producer: vacuumProducer}
			if scenario == "replaced exile" || scenario == "illegal target" {
				if len(linkedObjects(g, key)) != 0 {
					t.Fatal("an unsuccessful graveyard exile published a link")
				}
			}
		})
	}
}

func TestLinkedProducerGhostVacuumBlinkCopiedActivationsAndCloneIsolation(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	oldCard := addCardToGraveyard(g, game.Player2, vanillaCreature("Old source captive", 4, 4))
	newCard := addCardToGraveyard(g, game.Player3, vanillaCreature("New source captive", 5, 5))
	activateVacuumExile(t, g, e, source, oldCard)
	activateVacuumReturn(t, g, e, source)
	oldActivation, _ := g.Stack.Peek()
	copyActivation := game.NewStackObjectCopy(oldActivation, g.IDGen.Next())
	copyActivation.Controller = game.Player4
	g.Stack.Push(copyActivation)
	// Reentry is test setup while both old-source returns remain queued.
	e.resolveInstructionWithChoices(g, oldActivation, &game.Instruction{Primitive: game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	reentered := permanentByCardID(g, source.CardInstanceID)
	if reentered == nil || reentered.ObjectID == source.ObjectID {
		t.Fatal("source did not become a new occurrence")
	}
	activateVacuumExile(t, g, e, reentered, newCard)
	g.Stack.Push(linkedProducerObject(g, reentered, linkedProducerBody("", ghostVacuumReturn())))
	e.resolveTopOfStack(g, &TurnLog{})
	assertVacuumSpirit(t, g, newCard, game.Player1, 1)
	if !g.Players[game.Player2].Exile.Contains(oldCard) {
		t.Fatal("new source occurrence reclaimed the old captive")
	}
	clone := g.Clone()
	e.resolveTopOfStack(clone, &TurnLog{})
	assertVacuumSpirit(t, clone, oldCard, game.Player4, 1)
	e.resolveTopOfStack(clone, &TurnLog{})
	if countCardPermanentsControlledBy(clone, game.Player4, oldCard) != 1 ||
		!g.Players[game.Player2].Exile.Contains(oldCard) {
		t.Fatal("resolving copies duplicated an arrival or mutated the original clone")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	e.resolveTopOfStack(g, &TurnLog{})
	assertVacuumSpirit(t, g, oldCard, game.Player4, 1)
	if !reflect.DeepEqual(g.Events, clone.Events) {
		t.Fatal("independent clones did not replay copied return activations deterministically")
	}
}

func TestLinkedProducerGhostVacuumCopiedExilesAfterDepartureAndCloneIsolation(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	a := addCardToGraveyard(g, game.Player2, vanillaCreature("First copied exile", 4, 4))
	b := addCardToGraveyard(g, game.Player3, vanillaCreature("Second copied exile", 5, 5))
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
		[]game.Target{currentCardTarget(t, g, a)}, 0)) {
		t.Fatal("publication activation failed")
	}
	original, _ := g.Stack.Peek()
	copied := game.NewStackObjectCopy(original, g.IDGen.Next())
	copied.Controller = game.Player4
	copied.Targets = []game.Target{currentCardTarget(t, g, b)}
	g.Stack.Push(copied)
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("original source did not leave")
	}
	clone := g.Clone()
	e.resolveTopOfStack(clone, &TurnLog{})
	e.resolveTopOfStack(clone, &TurnLog{})
	key := game.LinkedObjectKey{SourceID: source.ObjectID, LinkID: string(vacuumProducer), Producer: vacuumProducer}
	if len(linkedObjects(g, key)) != 0 || len(linkedObjects(clone, key)) != 2 {
		t.Fatal("copied publication lost its original source or mutated the independent clone")
	}
	clone.Stack.Push(linkedProducerObject(clone, source, linkedProducerBody("", ghostVacuumReturn())))
	e.resolveTopOfStack(clone, &TurnLog{})
	for _, cardID := range []id.ID{a, b} {
		assertVacuumSpirit(t, clone, cardID, game.Player1, 1)
		if !g.Players[g.CardInstances[cardID].Owner].Graveyard.Contains(cardID) {
			t.Fatal("cloned publication or return changed the original graveyard")
		}
	}
}

func TestLinkedProducerFilteredModifiedUnionDeduplicatesArrivals(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	a := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	b := addCombatCreaturePermanentWithPower(g, game.Player3, 5)
	c := addCombatPermanent(g, game.Player4, &game.CardDef{CardFace: game.CardFace{
		Name: "Noncreature in sibling channel", Types: []types.Card{types.Artifact},
	}})
	resolveLinkedProducer(g, source, "test/a", a)
	resolveLinkedProducer(g, source, "test/b", b)
	resolveLinkedProducer(g, source, "test/b", c)
	put := ghostVacuumReturn()
	put.Source = game.BattlefieldSource{}
	put.Sources = []game.BattlefieldSource{game.OriginalSourceLinkedBattlefieldSource("test/a"),
		game.OriginalSourceLinkedBattlefieldSource("test/a"), game.OriginalSourceLinkedBattlefieldSource("test/b")}
	obj := linkedProducerObject(g, source, linkedProducerBody("", put))
	g.Stack.Push(obj)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	for _, cardID := range []id.ID{a.CardInstanceID, b.CardInstanceID} {
		assertVacuumSpirit(t, g, cardID, game.Player1, 1)
		if countCardPermanentsControlledBy(g, game.Player1, cardID) != 1 {
			t.Fatal("producer union duplicated a linked arrival")
		}
	}
	if !g.Players[c.Owner].Exile.Contains(c.CardInstanceID) {
		t.Fatal("a sibling channel's noncreature bypassed the candidate filter")
	}
}

func TestLinkedProducerGhostVacuumMalformedReturnFailsBeforeMutation(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	cardID := addCardToGraveyard(g, game.Player2, vanillaCreature("Preflight captive", 4, 4))
	activateVacuumExile(t, g, e, source, cardID)
	put := ghostVacuumReturn()
	put.Selection.ExcludedTypes = []types.Card{types.Creature}
	obj := linkedProducerObject(g, source, linkedProducerBody("", put))
	ctx := &programContext{engine: e, game: g, obj: obj, active: true, log: &TurnLog{}}
	beforeEvents := len(g.Events)
	if _, err := ctx.Execute(put); err == nil {
		t.Fatal("malformed authored return did not report an explicit preflight error")
	}
	if !g.Players[game.Player2].Exile.Contains(cardID) || len(g.Events) != beforeEvents || len(g.ContinuousEffects) != 0 {
		t.Fatal("malformed return mutated the game before preflight")
	}
}

func TestLinkedProducerGhostVacuumUsesOrdinaryCostsAndSorceryTiming(t *testing.T) {
	g, e, source := ghostVacuumFixture(t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	g.Turn.Phase = game.PhaseBeginning
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("sorcery-only return activated outside a main phase")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	source.Tapped = true
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("return ignored its tap cost")
	}
	source.Tapped = false
	g.Players[game.Player1].ManaPool.Empty()
	if e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 1, nil, 0)) {
		t.Fatal("return ignored its six-mana cost")
	}
	if permanentByCardID(g, source.CardInstanceID) == nil || g.Stack.Size() != 0 {
		t.Fatal("failed activation paid sacrifice or put an ability on the stack")
	}
}

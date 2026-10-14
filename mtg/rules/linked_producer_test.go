package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const prisonProducer game.LinkedProducerKey = "prison/exile"

func linkedProducerBody(producer game.LinkedProducerKey, primitive game.Primitive) game.AbilityContent {
	body := game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability()
	if producer != "" {
		body.LinkedProducers = []game.LinkedProducerKey{producer}
	}
	return body
}

func linkedProducerObject(g *game.Game, source *game.Permanent, body game.AbilityContent, targets ...game.Target) *game.StackObject {
	return &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: source.ObjectID,
		SourceCardID: source.CardInstanceID, Controller: source.Controller, Targets: targets,
		InlineActivated: &game.ActivatedAbility{Content: body},
	}
}

func resolveLinkedProducer(g *game.Game, source *game.Permanent, producer game.LinkedProducerKey, captive *game.Permanent) {
	body := linkedProducerBody(producer, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinkedProducer: producer,
	})
	g.Stack.Push(linkedProducerObject(g, source, body, game.PermanentTarget(captive.ObjectID)))
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
}

func linkedProducerReturnBody(producers ...game.LinkedProducerKey) game.AbilityContent {
	put := game.PutOnBattlefield{}
	if len(producers) == 1 {
		put.Source = game.OriginalSourceLinkedBattlefieldSource(producers[0])
	} else {
		for _, producer := range producers {
			put.Sources = append(put.Sources, game.OriginalSourceLinkedBattlefieldSource(producer))
		}
	}
	return linkedProducerBody("", put)
}

func TestLinkedProducerAccumulatesAndKeepsIndependentOriginalSources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	a := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	b := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	c := addCombatCreaturePermanentWithPower(g, game.Player4, 2)
	a.Controller = game.Player4
	resolveLinkedProducer(g, first, prisonProducer, a)
	resolveLinkedProducer(g, first, prisonProducer, b)
	resolveLinkedProducer(g, second, prisonProducer, c)
	g.Stack.Push(linkedProducerObject(g, first, linkedProducerReturnBody(prisonProducer)))
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if countCardPermanentsControlledBy(g, a.Owner, a.CardInstanceID) != 1 ||
		countCardPermanentsControlledBy(g, b.Owner, b.CardInstanceID) != 1 ||
		!g.Players[c.Owner].Exile.Contains(c.CardInstanceID) {
		t.Fatal("return lost accumulated cards, used former controllers, or crossed source occurrences")
	}
	var batch game.ObjectID
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentEnteredBattlefield &&
			(event.CardID == a.CardInstanceID || event.CardID == b.CardInstanceID) {
			if event.SimultaneousID == 0 || (batch != 0 && batch != event.SimultaneousID) {
				t.Fatal("linked returns were not simultaneous")
			}
			batch = event.SimultaneousID
		}
	}
	if batch == 0 {
		t.Fatal("linked return did not emit entry events")
	}
}

func TestLinkedProducerBlinkCannotReclaimButQueuedCopyAndCloneCanReturnOldSet(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	oldCaptive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveLinkedProducer(g, source, prisonProducer, oldCaptive)
	oldReturn := linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer))
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("source did not die")
	}
	engine := NewEngine(nil)
	engine.resolveInstructionWithChoices(g, oldReturn, &game.Instruction{Primitive: game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	var reentered *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == source.CardInstanceID {
			reentered = permanent
		}
	}
	if reentered == nil || reentered.ObjectID == source.ObjectID {
		t.Fatal("source did not reenter with fresh object identity")
	}
	newCaptive := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	resolveLinkedProducer(g, reentered, prisonProducer, newCaptive)
	g.Stack.Push(linkedProducerObject(g, reentered, linkedProducerReturnBody(prisonProducer)))
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[oldCaptive.Owner].Exile.Contains(oldCaptive.CardInstanceID) {
		t.Fatal("new source reclaimed the old occurrence's captive")
	}
	g.Stack.Push(oldReturn)
	copied := game.NewStackObjectCopy(oldReturn, g.IDGen.Next())
	g.Stack.Push(copied)
	clone := g.Clone()
	engine.resolveTopOfStack(clone, &TurnLog{})
	if countCardPermanentsControlledBy(clone, oldCaptive.Owner, oldCaptive.CardInstanceID) != 1 ||
		!g.Players[oldCaptive.Owner].Exile.Contains(oldCaptive.CardInstanceID) {
		t.Fatal("queued return copy lost original identity or clone aliased linked state")
	}
	engine.resolveTopOfStack(clone, &TurnLog{})
	if countCardPermanentsControlledBy(clone, oldCaptive.Owner, oldCaptive.CardInstanceID) != 1 {
		t.Fatal("second queued return duplicated a captive")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestLinkedProducerAuthoredMoveCardAfterDeparture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	cardID := addCardInstance(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Captive", Types: []types.Card{types.Creature},
	}})
	g.Players[game.Player2].Graveyard.Add(cardID)
	ref := game.ProgramReference{ID: "test/linked-exile", Version: "1"}
	called := false
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{Reference: ref, Resolve: func(ctx cardprogram.Context) error {
		called = true
		_, err := ctx.Execute(game.MoveCard{
			Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard,
			Destination: zone.Exile, PublishLinkedProducer: prisonProducer,
		})
		return err
	}}})
	if err != nil {
		t.Fatal(err)
	}
	body := game.Mode{Program: ref}.Ability()
	body.Modes[0].Targets = []game.TargetSpec{{Allow: game.TargetAllowCard, MinTargets: 1, MaxTargets: 1, TargetZone: zone.Graveyard}}
	body.LinkedProducers = []game.LinkedProducerKey{prisonProducer}
	obj := linkedProducerObject(g, source, body, game.CardTargetWithZoneVersion(cardID, 0))
	movePermanentToZone(g, source, zone.Hand)
	g.Stack.Push(obj)
	engine := NewEngineWithPrograms(nil, catalog)
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if !g.Players[game.Player2].Exile.Contains(cardID) {
		t.Fatalf("authored exile failed: called=%t targets=%+v log=%+v links=%+v", called, obj.Targets, log.Resolves, g.LinkedObjects)
	}
	g.Stack.Push(linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer)))
	engine.resolveTopOfStack(g, &TurnLog{})
	if countCardPermanentsControlledBy(g, game.Player2, cardID) != 1 {
		t.Fatal("authored card-only captive or departed original source was lost")
	}
}

// This fixture proves the pinned Helvault link topology in the shared Engine;
// it does not install or enable the still-refused legacy card implementation.
func TestLinkedProducerHelvaultTwoProducersAndQueuedDeathReturn(t *testing.T) {
	const yours game.LinkedProducerKey = "helvault/yours"
	const others game.LinkedProducerKey = "helvault/others"
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Helvault", Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventZoneChanged, Source: game.TriggerSourceSelf,
				MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard,
			}},
			Content: linkedProducerReturnBody(yours, others),
		}},
	}}
	for _, producer := range []game.LinkedProducerKey{yours, others} {
		content := linkedProducerBody(producer, game.MovePermanent{
			Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinkedProducer: producer,
		})
		content.Modes[0].Targets = []game.TargetSpec{{Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}}
		def.ActivatedAbilities = append(def.ActivatedAbilities, game.ActivatedAbility{Content: content})
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("Helvault shared-operation fixture invalid: %+v", issues)
	}
	source := addCombatPermanent(g, game.Player1, def)
	a := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	a.Controller = game.Player1
	b := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	c := addCombatCreaturePermanentWithPower(g, game.Player4, 2)
	resolveLinkedProducer(g, source, yours, a)
	resolveLinkedProducer(g, source, others, b)
	resolveLinkedProducer(g, source, others, c)
	// A consumer selecting just one producer cannot see its sibling's products.
	copyGame := g.Clone()
	copyGame.Stack.Push(linkedProducerObject(copyGame, source, linkedProducerReturnBody(yours)))
	NewEngine(nil).resolveTopOfStack(copyGame, &TurnLog{})
	if countCardPermanentsControlledBy(copyGame, a.Owner, a.CardInstanceID) != 1 ||
		!copyGame.Players[b.Owner].Exile.Contains(b.CardInstanceID) {
		t.Fatal("separate producer channels flattened")
	}
	engine := NewEngine(nil)
	// A blink without death creates no return trigger and cannot reclaim captives.
	if !movePermanentToZone(g, source, zone.Hand) {
		t.Fatal("source blink failed")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Helvault returned captives on a non-graveyard departure")
	}
	obj := linkedProducerObject(g, source, linkedProducerReturnBody(yours, others))
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	reentered := permanentByCardID(g, source.CardInstanceID)
	g.Stack.Push(linkedProducerObject(g, reentered, linkedProducerReturnBody(yours, others)))
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[a.Owner].Exile.Contains(a.CardInstanceID) || !g.Players[b.Owner].Exile.Contains(b.CardInstanceID) {
		t.Fatal("blinked Helvault reclaimed older captives")
	}
	newCaptive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveLinkedProducer(g, reentered, yours, newCaptive)
	movePermanentToZone(g, reentered, zone.Graveyard)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("actual Helvault death trigger not queued")
	}
	queued, _ := g.Stack.Peek()
	if queued.SourceID != reentered.ObjectID {
		t.Fatal("death trigger did not capture the original source occurrence")
	}
	engine.resolveInstructionWithChoices(g, queued, &game.Instruction{Primitive: game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if countCardPermanentsControlledBy(g, newCaptive.Owner, newCaptive.CardInstanceID) != 1 ||
		!g.Players[a.Owner].Exile.Contains(a.CardInstanceID) {
		t.Fatal("queued death returned the wrong original-source set")
	}
	// Resolve a return captured from the first source occurrence: both declared
	// producers and every accumulated card return, under their respective owners.
	g.Stack.Push(obj)
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, captive := range []*game.Permanent{a, b, c} {
		if countCardPermanentsControlledBy(g, captive.Owner, captive.CardInstanceID) != 1 {
			t.Fatal("multi-producer owner-controlled return missed a captive")
		}
	}
}

func TestLinkedProducerCaptiveDepartureAndReentryIsNotReclaimed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveLinkedProducer(g, source, prisonProducer, captive)
	if !moveCardBetweenZones(g, captive.Owner, captive.CardInstanceID, zone.Exile, zone.Hand) ||
		!moveCardBetweenZones(g, captive.Owner, captive.CardInstanceID, zone.Hand, zone.Exile) {
		t.Fatal("captive did not leave and reenter exile")
	}
	g.Stack.Push(linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer)))
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[captive.Owner].Exile.Contains(captive.CardInstanceID) ||
		permanentByCardID(g, captive.CardInstanceID) != nil {
		t.Fatal("return reclaimed a later exile incarnation")
	}
}

func TestLinkedProducerMovementReplacementsAndAuraAdmission(t *testing.T) {
	for _, scenario := range []string{"diverted exile", "diverted return", "legal Aura", "Aura no host", "Aura retry"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Prison", Types: []types.Card{types.Artifact},
			}})
			def := vanillaCreature("Captive", 2, 2)
			if scenario == "legal Aura" || scenario == "Aura no host" || scenario == "Aura retry" {
				def = &game.CardDef{CardFace: game.CardFace{Name: "Captive Aura", Types: []types.Card{types.Enchantment},
					Subtypes: []types.Sub{types.Aura}, StaticAbilities: []game.StaticAbility{
						game.EnchantStaticAbility(&game.TargetSpec{Allow: game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}),
					},
				}}
			}
			captive := addCombatPermanent(g, game.Player2, def)
			if scenario == "diverted exile" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Graveyard,
				})
			}
			resolveLinkedProducer(g, source, prisonProducer, captive)
			key := originalLinkedProducerKey(g, linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer)), prisonProducer)
			if scenario == "diverted exile" {
				if len(linkedObjects(g, key)) != 0 || !g.Players[captive.Owner].Graveyard.Contains(captive.CardInstanceID) {
					t.Fatal("replaced exile published an intended rather than actual product")
				}
				return
			}
			if scenario == "diverted return" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Exile,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Graveyard,
				})
			}
			var host *game.Permanent
			if scenario == "legal Aura" {
				host = addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
					Name: "Shrouded Host", Types: []types.Card{types.Creature},
					StaticAbilities: []game.StaticAbility{game.ShroudStaticBody, game.HexproofStaticBody},
				}})
			}
			put := game.PutOnBattlefield{Source: game.OriginalSourceLinkedBattlefieldSource(prisonProducer), PublishLinked: "arrival"}
			body := linkedProducerBody("", put)
			body.Modes[0].Sequence[0].PublishResult = "arrival-result"
			obj := linkedProducerObject(g, source, body)
			g.Stack.Push(obj)
			engine := NewEngine(nil)
			engine.resolveTopOfStack(g, &TurnLog{})
			failed := scenario != "legal Aura"
			receipt := obj.ResolutionResults["arrival-result"]
			arrival := linkedObjects(g, linkedObjectSourceKey(g, obj, "arrival"))
			if receipt.Succeeded == failed || (len(arrival) == 0) != failed ||
				(len(obj.ResolutionResultObjects["arrival-result"]) == 0) != failed {
				t.Fatal("entry outcome or successful products disagreed with actual arrival")
			}
			if scenario == "diverted return" && !g.Players[captive.Owner].Graveyard.Contains(captive.CardInstanceID) {
				t.Fatal("entry replacement did not use actual destination")
			}
			if scenario == "legal Aura" {
				entered := permanentByCardID(g, captive.CardInstanceID)
				if entered == nil || entered.AttachedTo != opt.Val(host.ObjectID) || entered.Controller != captive.Owner {
					t.Fatal("Aura did not legally enter under its owner without targeting a shrouded host")
				}
			}
			if scenario == "Aura retry" {
				addCombatCreaturePermanentWithPower(g, game.Player3, 2)
				g.Stack.Push(game.NewStackObjectCopy(obj, g.IDGen.Next()))
				engine.resolveTopOfStack(g, &TurnLog{})
				if permanentByCardID(g, captive.CardInstanceID) == nil {
					t.Fatal("failed Aura admission consumed the CR 607 link before a resolving copy retried")
				}
			}
		})
	}
}

func TestLinkedProducerLocalFramesCannotShadowPersistentChannel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	body := linkedProducerBody(prisonProducer, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinkedProducer: prisonProducer,
	})
	obj := linkedProducerObject(g, source, body, game.PermanentTarget(captive.ObjectID))
	restore := enterLocalProductFrame(g, obj, []game.Instruction{{LocalProducts: game.LocalProducts{
		Links: []game.LinkedKey{game.LinkedKey(prisonProducer)},
	}}})
	local := obj.LocalLinkedProducts[game.LinkedKey(prisonProducer)]
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, body, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.LinkedObjects[local]) != 0 {
		t.Fatal("local-product frame intercepted a declared persistent producer")
	}
	restore()
	g.Stack.Push(linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer)))
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, captive.CardInstanceID) == nil || len(obj.LocalLinkedProducts) != 0 {
		t.Fatal("persistent channel was lost during local-product restoration")
	}
}

func TestLinkedProducerCopiedPublicationAfterSourceSelfExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	body := linkedProducerBody(prisonProducer, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinkedProducer: prisonProducer,
	})
	original := linkedProducerObject(g, source, body, game.PermanentTarget(source.ObjectID))
	copied := game.NewStackObjectCopy(original, g.IDGen.Next())
	copied.Targets = []game.Target{game.PermanentTarget(captive.ObjectID)}
	secondCopy := game.NewStackObjectCopy(original, g.IDGen.Next())
	secondCopy.Targets = []game.Target{game.PermanentTarget(other.ObjectID)}
	engine := NewEngine(nil)
	g.Stack.Push(original)
	engine.resolveTopOfStack(g, &TurnLog{})
	g.Stack.Push(copied)
	engine.resolveTopOfStack(g, &TurnLog{})
	g.Stack.Push(secondCopy)
	engine.resolveTopOfStack(g, &TurnLog{})
	g.Stack.Push(linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer)))
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, permanent := range []*game.Permanent{source, captive, other} {
		if countCardPermanentsControlledBy(g, permanent.Owner, permanent.CardInstanceID) != 1 {
			t.Fatal("self-departure or copied producer overwrote the original accumulated set")
		}
	}
}

func TestLinkedProducerAuthoredOwnershipAndArrivalOutcomes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	captive := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := linkedProducerObject(g, source, linkedProducerReturnBody(prisonProducer), game.PermanentTarget(captive.ObjectID))
	ctx := &programContext{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}, active: true,
		specs: []game.TargetSpec{{Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}}}
	exile := game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinkedProducer: prisonProducer}
	if _, err := ctx.Execute(exile); err == nil || permanentByCardID(g, captive.CardInstanceID) == nil {
		t.Fatal("undeclared authored producer mutated the game")
	}
	obj.ResolvingLinkedProducers = []game.LinkedProducerKey{prisonProducer}
	if outcome, err := ctx.Execute(exile); err != nil || !outcome.Succeeded {
		t.Fatalf("declared exile outcome=%+v err=%v", outcome, err)
	}
	put := game.PutOnBattlefield{Source: game.OriginalSourceLinkedBattlefieldSource(prisonProducer)}
	outcome, err := ctx.Execute(put)
	if err != nil || !outcome.Succeeded || len(outcome.Objects) != 1 ||
		outcome.Objects[0].CardID != captive.CardInstanceID || outcome.Objects[0].ObjectID == captive.ObjectID {
		t.Fatalf("actual authored arrival outcome=%+v err=%v", outcome, err)
	}
	outcome, err = ctx.Execute(put)
	if err != nil || outcome.Succeeded || len(outcome.Objects) != 0 {
		t.Fatalf("repeat return claimed nonexistent arrivals: outcome=%+v err=%v", outcome, err)
	}
}

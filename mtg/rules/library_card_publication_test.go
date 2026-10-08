package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardPublicationUsesActualObservation(t *testing.T) {
	for _, reveal := range []bool{false, true} {
		t.Run(map[bool]string{false: "look", true: "reveal"}[reveal], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := addCardToLibrary(g, game.Player1, vanillaCreature("Decoy", 1, 1))
			observed := addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			g.CardInstances[observed].ZoneVersion = 7
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
			}
			var primitive game.Primitive = game.LookAtLibraryTop{
				Player: game.TargetPlayerReference(1), PublishLinked: "observation",
			}
			if reveal {
				primitive = game.Reveal{
					Amount: game.Fixed(1), Player: game.TargetPlayerReference(1), PublishLinked: "observation",
				}
			}
			observer, owner := &lookChoiceAgent{}, &lookChoiceAgent{}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1], agents[game.Player3] = observer, owner
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
				Primitive: primitive, PublishResult: "observation-result",
			}, agents, &TurnLog{})
			refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "observation"))
			if len(refs) != 1 || refs[0].CardID != observed || refs[0].ObjectID != 0 ||
				refs[0].CardZoneVersion != 7 || !refs[0].CardZoneVersionSet {
				t.Fatalf("published = %#v, want observed card %d at version 7", refs, observed)
			}
			if !g.Players[game.Player1].Library.Contains(decoy) || !g.Players[game.Player3].Library.Contains(observed) {
				t.Fatal("observation moved a card")
			}
			if eventRevealedCard(g, observed, obj.ID) != reveal {
				t.Fatalf("public reveal event does not match reveal=%v", reveal)
			}
			if !reveal && (len(observer.requests) != 1 || len(owner.requests) != 0 ||
				observer.requests[0].Player != game.Player1 || observer.requests[0].Subject.Val.CardID != observed) {
				t.Fatalf("private observation went to wrong player: observer=%#v owner=%#v", observer.requests, owner.requests)
			}
		})
	}
}

func TestLibraryCardPublisherClearsOnlyItsOwnUnavailableProduct(t *testing.T) {
	for _, reveal := range []bool{false, true} {
		for _, failure := range []string{"empty", "unavailable player", "declined", "condition skipped"} {
			t.Run(map[bool]string{false: "look/", true: "reveal/"}[reveal]+failure, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				old := addCardToLibrary(g, game.Player1, vanillaCreature("Old", 1, 1))
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
				key := linkedObjectSourceKey(g, obj, "observation")
				rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: old, CardZoneVersion: 1})
				otherKey := linkedObjectSourceKey(g, obj, "independent")
				rememberLinkedObject(g, otherKey, game.LinkedObjectRef{CardID: old, CardZoneVersion: 1})
				player := game.ControllerReference()
				instr := game.Instruction{}
				switch failure {
				case "empty":
					g.Players[game.Player1].Library.Remove(old)
				case "unavailable player":
					player = game.TargetPlayerReference(0)
				case "declined":
					instr.Optional = true
				case "condition skipped":
					instr.Condition = opt.Val(game.EffectCondition{
						Object: game.LinkedObjectReference("missing"), PermanentType: opt.Val(types.Creature),
					})
				default:
					t.Fatalf("unknown publication failure %q", failure)
				}
				instr.Primitive = game.LookAtLibraryTop{Player: player, PublishLinked: "observation"}
				if reveal {
					instr.Primitive = game.Reveal{Amount: game.Fixed(1), Player: player, PublishLinked: "observation"}
				}
				var agents [game.NumPlayers]PlayerAgent
				agents[game.Player1] = &declineChoiceAgent{}
				NewEngine(nil).resolveInstructionWithChoices(g, obj, &instr, agents, &TurnLog{})
				if refs := linkedObjects(g, key); len(refs) != 0 {
					t.Fatalf("unavailable publisher retained old product: %#v", refs)
				}
				if refs := linkedObjects(g, otherKey); len(refs) != 1 || refs[0].CardID != old {
					t.Fatalf("publisher cleared an independent observation: %#v", refs)
				}
			})
		}
	}
}

func TestExplicitRevealPublishesSameObservedCardVersion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 3))
	g.CardInstances[cardID].ZoneVersion = 9
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	rememberLinkedObject(g, linkedObjectSourceKey(g, obj, "looked"), game.LinkedObjectRef{
		CardID: cardID, CardZoneVersion: 9,
	})
	primitive := game.Reveal{
		Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "looked"}, PublishLinked: "revealed",
	}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	res := handleReveal(resolver, primitive)
	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "revealed"))
	if !res.succeeded || res.amount != 1 || len(refs) != 1 || refs[0].CardID != cardID ||
		refs[0].CardZoneVersion != 9 || !refs[0].CardZoneVersionSet {
		t.Fatalf("explicit reveal = %#v, product = %#v", res, refs)
	}
	g.Players[game.Player2].Library.Remove(cardID)
	g.Players[game.Player2].Hand.Add(cardID)
	g.CardInstances[cardID].ZoneVersion++
	if res := handleReveal(resolver, primitive); res.succeeded || res.amount != 0 {
		t.Fatalf("unavailable library reveal succeeded: %#v", res)
	}
	if refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "revealed")); len(refs) != 0 {
		t.Fatalf("failed explicit reveal retained product: %#v", refs)
	}
}

func TestRevealReportsActualCardinality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToLibrary(g, game.Player1, vanillaCreature("Only", 1, 1))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	res := handleReveal(resolver, game.Reveal{Amount: game.Fixed(3), Player: game.ControllerReference(), PublishLinked: "revealed"})
	if !res.succeeded || res.amount != 1 {
		t.Fatalf("actual reveal outcome = %#v, want one revealed card", res)
	}
	if refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "revealed")); len(refs) != 1 {
		t.Fatalf("actual revealed membership = %#v", refs)
	}
}

func TestLibraryCardObservationDoesNotFollowReincarnation(t *testing.T) {
	for _, version := range []uint64{0, 7} {
		for _, reveal := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/reveal-%v", version, reveal), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				decoy := addCardToLibrary(g, game.Player2, vanillaCreature("Decoy", 4, 5))
				cardID := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 3))
				g.CardInstances[cardID].ZoneVersion = version
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Controller: game.Player1, Targets: []game.Target{game.PlayerTarget(game.Player2)},
				}
				resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				var producer game.Primitive = game.LookAtLibraryTop{
					Player: game.TargetPlayerReference(0), PublishLinked: "observed",
				}
				if reveal {
					producer = game.Reveal{
						Player: game.TargetPlayerReference(0), Amount: game.Fixed(1), PublishLinked: "observed",
					}
				}
				resolver.resolveInstruction(&game.Instruction{Primitive: producer})
				linked := game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"}
				if id, from, ok := resolveCardReference(g, obj, linked); !ok || id != cardID || from != zone.Library {
					t.Fatalf("initial observation = %d %v %v", id, from, ok)
				}
				if !moveCardBetweenZonesWithPlacement(g, game.Player2, cardID, zone.Library, zone.Hand, false) ||
					!moveCardBetweenZonesWithPlacement(g, game.Player2, cardID, zone.Hand, zone.Library, false) {
					t.Fatal("could not leave and return the observed card")
				}
				if id, _, ok := resolveCardReference(g, obj, linked); ok {
					t.Fatalf("old observation followed card %d into a new incarnation", id)
				}
				if res := handleReveal(resolver, game.Reveal{Card: linked, PublishLinked: "later-reveal"}); res.succeeded {
					t.Fatal("stale observation revealed a new incarnation")
				}
				if res := handleMoveCard(resolver, game.MoveCard{
					Card: linked, FromZone: zone.Library, Destination: zone.Hand,
				}); res.succeeded {
					t.Fatal("stale observation moved a new incarnation")
				}
				if !g.Players[game.Player2].Library.Contains(cardID) || !g.Players[game.Player2].Library.Contains(decoy) {
					t.Fatal("stale consumers moved the observed incarnation or new top card")
				}
			})
		}
	}
}

func TestLibraryCardSameKeyRevealEnvelope(t *testing.T) {
	for _, outcome := range []string{"success", "declined", "condition skipped", "result skipped"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToLibrary(g, game.Player2, vanillaCreature("Observed", 2, 3))
			g.CardInstances[cardID].ZoneVersion = 5
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			key := linkedObjectSourceKey(g, obj, "observed")
			rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: cardID, CardZoneVersion: 5})
			independent := linkedObjectSourceKey(g, obj, "independent")
			rememberLinkedObject(g, independent, game.LinkedObjectRef{CardID: cardID, CardZoneVersion: 5})
			instr := game.Instruction{
				Primitive: game.Reveal{
					Card:          game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
					PublishLinked: "observed",
				},
				CardCondition: opt.Val(game.CardSelection{
					Card:      game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
					Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				}),
				PublishResult: "reveal-result",
			}
			switch outcome {
			case "success":
			case "declined":
				instr.Optional = true
			case "condition skipped":
				instr.CardCondition.Val.Selection.RequiredTypes = []types.Card{types.Land}
			case "result skipped":
				instr.ResultGate = opt.Val(game.InstructionResultGate{Key: "missing", Succeeded: game.TriTrue})
			default:
				t.Fatalf("unknown reveal outcome %q", outcome)
			}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = &declineChoiceAgent{}
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &instr, agents, &TurnLog{})
			refs := linkedObjects(g, key)
			if outcome == "success" {
				if len(refs) != 1 || refs[0].CardID != cardID || refs[0].CardZoneVersion != 5 {
					t.Fatalf("same-key reveal lost its exact input: %#v", refs)
				}
			} else if len(refs) != 0 {
				t.Fatalf("skipped same-key reveal retained stale publication: %#v", refs)
			}
			if eventRevealedCard(g, cardID, obj.ID) != (outcome == "success") {
				t.Fatal("public reveal did not follow the actual action outcome")
			}
			if refs := linkedObjects(g, independent); len(refs) != 1 || refs[0].CardID != cardID {
				t.Fatalf("same-key reveal cleared an independent product: %#v", refs)
			}
		})
	}
}

func TestLibraryCardRepeatedPublicationAndClonePreserveIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	next := addCardToLibrary(g, game.Player1, vanillaCreature("Next", 1, 1))
	first := addCardToLibrary(g, game.Player1, vanillaCreature("First", 2, 2))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	instr := &game.Instruction{
		Primitive:     game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "observed"},
		PublishResult: "observed-result",
	}
	resolver.resolveInstruction(instr)
	key := linkedObjectSourceKey(g, obj, "observed")
	refs := linkedObjects(g, key)
	if len(refs) != 1 || refs[0].CardID != first || refs[0].CardZoneVersion != 0 || !refs[0].CardZoneVersionSet {
		t.Fatalf("initial zero-version publication = %#v", refs)
	}
	cloned := g.Clone()
	clonedRefs := linkedObjects(cloned, key)
	if len(clonedRefs) != 1 || clonedRefs[0].CardID != first ||
		clonedRefs[0].CardZoneVersion != 0 || !clonedRefs[0].CardZoneVersionSet {
		t.Fatalf("clone dropped observed version availability: %#v", clonedRefs)
	}
	if !moveCardBetweenZonesWithPlacement(g, game.Player1, first, zone.Library, zone.Hand, false) {
		t.Fatal("could not move first observation")
	}
	resolver.resolveInstruction(instr)
	refs = linkedObjects(g, key)
	if len(refs) != 1 || refs[0].CardID != next || !refs[0].CardZoneVersionSet {
		t.Fatalf("repeated resolution accumulated or retained old products: %#v", refs)
	}
	if !moveCardBetweenZonesWithPlacement(g, game.Player1, next, zone.Library, zone.Hand, false) {
		t.Fatal("could not empty library")
	}
	resolver.resolveInstruction(instr)
	if refs := linkedObjects(g, key); len(refs) != 0 {
		t.Fatalf("empty repeated resolution retained a product: %#v", refs)
	}
	result := obj.ResolutionResults["observed-result"]
	if !result.Accepted || result.Succeeded || result.Amount != 0 {
		t.Fatalf("empty observation invented success/cardinality: %#v", result)
	}
	if refs := linkedObjects(cloned, key); len(refs) != 1 || refs[0].CardID != first ||
		refs[0].CardZoneVersion != 0 || !refs[0].CardZoneVersionSet ||
		!cloned.Players[game.Player1].Library.Contains(first) {
		t.Fatalf("later resolutions changed cloned observation: %#v", refs)
	}
}

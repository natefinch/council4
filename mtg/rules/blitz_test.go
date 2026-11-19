package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: #4005, six pinned Oracle printings and their Scryfall rulings
// retrieved 2026-11-19: Riveteers Decoy (2d21701b-6a99-4e1c-82b2-0219f170b09f),
// Riveteers Requisitioner (8ab5152d-767a-47d5-acfc-f439810aaeea), Night Clubber
// (59fd3f8a-a13b-4a0a-bb27-b9246949ea7b), Plasma Jockey
// (741ed811-cbcd-4178-a874-dc79d2a5d6f1), Caldaia Strongarm
// (58b431d2-33c1-49d8-b45d-690c174bb456), and Mayhem Patrol
// (50162cdd-ba30-48df-93ff-197c7f4a2913). Each has five 2022-04-29 rulings:
// ordinary timing/countering, no required attack, draw on any death, sacrifice
// only the original object, and no consequences on a copy of the permanent.
// Official CR 2026-09-25 (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 118.9, 601.2b/f–i, 702.152a (alternative payment); 400.7d, 611.3b–c,
// 613.1f, 302.6, 702.10b (static, immediate haste); 700.4, 603.3a/10a
// (dies trigger looks back); 513.1, 603.7a–d, 701.21a (one delayed sacrifice,
// bound to the permanent and controlled by the resolving spell's controller);
// 707.2/10/12 and 111.13 (permanent copies do not inherit payment, stack
// copies do, casting a copy pays its own costs). These are Engine fixtures;
// no card is implemented or registered here.

func blitzFixture() *game.CardDef {
	def := vanillaCreature("Blitz fixture", 3, 1)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.G})
	def.AlternativeCosts = []cost.Alternative{cost.BlitzAlternative(cost.Mana{cost.O(3), cost.G})}
	def.StaticAbilities = []game.StaticAbility{game.BlitzStaticAbility()}
	return def
}

func blitzAgents(label string) [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: label}}
}

func castBlitzFixture(t *testing.T, e *Engine, g *game.Game, cast action.Action) *game.StackObject {
	t.Helper()
	if !e.applyActionWithChoices(g, game.Player1, cast, blitzAgents("Blitz"), &TurnLog{}) {
		t.Fatal("casting for Blitz failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Blitzed || obj.Dashed || obj.Evoked || spellCastEvents(g, obj.ID) != 1 {
		t.Fatalf("cast did not preserve the distinct Blitz payment and emit exactly one cast: %+v", obj)
	}
	return obj
}

func requireBlitzEntry(t *testing.T, g *game.Game, permanent *game.Permanent, paid, cast bool) {
	t.Helper()
	if permanent == nil {
		t.Fatal("spell did not become a permanent")
	}
	entry, ok := enteredEventFor(g, permanent)
	if !ok || entry.EnterBlitzed != paid || entry.EnterWasCast != cast ||
		permanent.Blitzed != paid || hasKeyword(g, permanent, game.Haste) != paid {
		t.Fatalf("entry for object %d: event=%+v; expected Blitz/haste=%t, cast=%t", permanent.ObjectID, entry, paid, cast)
	}
}

func blitzEndStep(e *Engine, g *game.Game) {
	g.Turn.Phase, g.Turn.Step = game.PhaseEnding, game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func TestBlitzPaidCastVersusNormal(t *testing.T) {
	for _, tt := range []struct {
		name  string
		label string
		mana  int
		paid  bool
	}{
		{"paid", "Blitz", 3, true},
		{"declined", "Normal cost", 3, false},
		{"unaffordable", "Normal cost", 1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, g := NewEngine(nil), castCopyProvenanceGame()
			card := addCardToHand(g, game.Player1, blitzFixture())
			draw := addCardToLibraryNamed(g, game.Player1, "Blitz draw")
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, tt.mana)
			agents := blitzAgents(tt.label)
			if !e.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("cast failed")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.Blitzed != tt.paid {
				t.Fatal("wrong committed payment")
			}
			e.resolveTopOfStack(g, &TurnLog{})
			permanent := permanentForCard(g, card)
			requireBlitzEntry(t, g, permanent, tt.paid, true)
			if slices.Contains(eligibleAttackers(g, game.Player1), permanent) != tt.paid {
				t.Fatal("only the Blitz creature should be able to attack immediately")
			}
			e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			if g.Stack.Size() != 0 {
				t.Fatal("Blitz must not put an entry trigger on the stack")
			}
			blitzEndStep(e, g)
			if (g.Stack.Size() == 1) != tt.paid || permanentForCard(g, card) == nil {
				t.Fatal("only a Blitz payment schedules a respondable end-step sacrifice")
			}
			if !tt.paid {
				if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
					t.Fatal("could not destroy normally cast creature")
				}
				e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			}
			resolveStackWithTriggers(e, g, agents)
			if !g.Players[game.Player1].Graveyard.Contains(card) || g.Players[game.Player1].Hand.Contains(draw) != tt.paid {
				t.Fatal("expected the creature to die, with a draw only for Blitz")
			}
		})
	}
}

func TestBlitzCombinesWithAdditionalCosts(t *testing.T) {
	e, g := NewEngine(nil), castCopyProvenanceGame()
	def := blitzFixture()
	// Use unrelated wording to show payment is recognized from the typed
	// alternative and not the display label.
	def.AlternativeCosts[0].Label = "Alternate fixture cost"
	def.AdditionalCosts = []cost.Additional{
		{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
		{Kind: cost.AdditionalDiscard, Amount: 1, ChoiceGroup: 1},
	}
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{
			Cost: cost.Mana{cost.U},
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalPayLife, Amount: 3, ChoiceGroup: 1},
				{Kind: cost.AdditionalDiscard, Amount: 1, ChoiceGroup: 1},
			},
		}},
	})
	card := addCardToHand(g, game.Player1, def)
	p := g.Players[game.Player1]
	life := p.Life
	fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C, mana.U)
	replay := g.Clone()
	log := &TurnLog{}
	cast := action.CastKickedSpell(card, nil, 0, nil)
	if !e.applyActionWithChoices(g, game.Player1, cast, blitzAgents("Alternate fixture cost"), log) {
		t.Fatal("could not combine Blitz with the additional mana and nonmana costs")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Blitzed || !obj.KickerPaid || p.Life != life-5 || p.ManaPool.Total() != 0 {
		t.Fatalf("want Blitz and Kicker, all mana spent, and both choice groups paid; life=%d, mana=%d", p.Life, p.ManaPool.Total())
	}
	if len(log.Choices) == 0 || !e.applyActionWithChoices(replay, game.Player1, cast, replayAgentsFromTurnLog(log), &TurnLog{}) {
		t.Fatal("could not replay recorded additional and alternative cost choices")
	}
	if obj, ok := replay.Stack.Peek(); !ok || !obj.Blitzed || !obj.KickerPaid ||
		replay.Players[game.Player1].Life != life-5 || replay.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("payment replay diverged")
	}
}

func TestBlitzPaidExileCastDuringEndStep(t *testing.T) {
	e, g := NewEngine(nil), castCopyProvenanceGame()
	card := addCardToExile(g, game.Player1, blitzFixture())
	addCardToLibraryNamed(g, game.Player1, "Blitz draw")
	fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
	blitzEndStep(e, g)
	// A resolving effect can give permission to cast a card from exile
	// immediately, paying its costs instead of making it a free cast.
	if !e.castPaidTargetedSpell(g, game.Player1, card, zone.Exile, blitzAgents("Blitz"), &TurnLog{}) {
		t.Fatal("failed to cast for Blitz from exile")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Blitzed || spellCastEvents(g, obj.ID) != 1 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("exile cast did not commit Blitz payment: %+v", obj)
	}
	resolveStackWithTriggers(e, g, [game.NumPlayers]PlayerAgent{})
	requireBlitzEntry(t, g, permanentForCard(g, card), true, true)
	// It is the next end step, even if it belongs to another player.
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player2
	blitzEndStep(e, g)
	resolveStackWithTriggers(e, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Graveyard.Contains(card) || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("next end step did not sacrifice and draw after a late Blitz cast")
	}
}

func TestBlitzAbilityLossDoesNotCancelSacrifice(t *testing.T) {
	e, g := NewEngine(nil), castCopyProvenanceGame()
	card := addCardToHand(g, game.Player1, blitzFixture())
	draw := addCardToLibraryNamed(g, game.Player1, "Blitz draw")
	fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
	castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
	e.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, card)
	requireBlitzEntry(t, g, permanent, true, true)
	if !applyTypedContinuousEffects(g, &game.StackObject{Controller: game.Player2}, permanent, []game.ContinuousEffect{{
		Layer: game.LayerAbility, RemoveAllAbilities: true,
	}}, game.DurationPermanent) || hasKeyword(g, permanent, game.Haste) {
		t.Fatal("ability loss failed to remove Blitz's haste")
	}
	blitzEndStep(e, g)
	resolveStackWithTriggers(e, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Graveyard.Contains(card) || g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("ability loss should prevent the draw but not the delayed sacrifice")
	}
}

func TestBlitzPermanentCopyDoesNotInheritPayment(t *testing.T) {
	e, g := NewEngine(nil), castCopyProvenanceGame()
	card := addCardToHand(g, game.Player1, blitzFixture())
	draw := addCardToLibraryNamed(g, game.Player1, "Blitz draw")
	fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
	castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
	e.resolveTopOfStack(g, &TurnLog{})
	original := permanentForCard(g, card)
	def, ok := permanentTokenCopyDef(g, original)
	if !ok {
		t.Fatal("could not copy the Blitz permanent's copiable values")
	}
	copy, _ := createTokenPermanent(g, game.Player1, def)
	requireBlitzEntry(t, g, copy, false, false)
	if _, ok := destroyPermanent(g, copy.ObjectID); !ok {
		t.Fatal("could not destroy the token copy")
	}
	if e.putTriggeredAbilitiesOnStack(g) || g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("the permanent copy inherited the original's dies ability")
	}
	blitzEndStep(e, g)
	if g.Stack.Size() != 1 {
		t.Fatal("the original should be the only scheduled sacrifice")
	}
	resolveStackWithTriggers(e, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("copying the permanent removed the original's dies ability")
	}
}

func TestBlitzEarlyDeathAndNewObject(t *testing.T) {
	e, g := NewEngine(nil), castCopyProvenanceGame()
	card := addCardToHand(g, game.Player1, blitzFixture())
	draw := addCardToLibraryNamed(g, game.Player2, "Blitz draw")
	fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
	castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
	e.resolveTopOfStack(g, &TurnLog{})
	original := permanentForCard(g, card)
	if !applyTypedContinuousEffects(g, &game.StackObject{Controller: game.Player2}, original, []game.ContinuousEffect{{
		Layer: game.LayerControl, NewController: opt.Val(game.Player2),
	}}, game.DurationPermanent) {
		t.Fatal("could not change control of the blitzed permanent")
	}
	if _, ok := destroyPermanent(g, original.ObjectID); !ok {
		t.Fatal("could not destroy blitzed permanent")
	}
	e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if obj, ok := g.Stack.Peek(); !ok || obj.Controller != game.Player2 {
		t.Fatal("the dies ability did not trigger for the permanent's last controller")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Hand.Contains(draw) {
		t.Fatal("destroying the blitzed creature did not draw for its controller")
	}
	instance, _ := g.GetCardInstance(card)
	returned, _ := createCardPermanent(g, instance, game.Player1, zone.Graveyard)
	requireBlitzEntry(t, g, returned, false, false)
	if returned.ObjectID == original.ObjectID {
		t.Fatal("reentry did not produce a new object")
	}
	blitzEndStep(e, g)
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Controller != game.Player1 {
		t.Fatal("delayed sacrifice lost the resolving spell's controller")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentForCard(g, card) != returned {
		t.Fatal("the stale sacrifice followed the card to its new object")
	}
	blitzEndStep(e, g)
	if g.Stack.Size() != 0 {
		t.Fatal("the delayed sacrifice triggered a second time")
	}
}

func TestBlitzStackCopiesAndCloneReplay(t *testing.T) {
	for _, route := range []string{"direct", "storm"} {
		t.Run(route, func(t *testing.T) {
			e, g := permanentCopyEngine(t), castCopyProvenanceGame()
			def := blitzFixture()
			if route == "storm" {
				def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{KeywordAbilities: game.SimpleKeywords(game.Storm)})
				prior := addCardToHand(g, game.Player1, simpleGainLifeInstant("Prior cast"))
				if !e.applyAction(g, game.Player1, action.CastSpell(prior, nil, 0, nil)) {
					t.Fatal("could not set storm count with a real cast")
				}
				e.resolveTopOfStack(g, &TurnLog{})
			}
			card := addCardToHand(g, game.Player1, def)
			fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
			original := castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
			var copied *game.StackObject
			if route == "direct" {
				copied = copyPermanentSpell(t, e, g, original, game.Player1)
			} else {
				e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				e.resolveTopOfStack(g, &TurnLog{})
				copied, _ = g.Stack.Peek()
			}
			if copied == nil || !copied.Copy || !copied.Blitzed || spellCastEvents(g, copied.ID) != 0 {
				t.Fatal("stack copy did not retain the choice, or was counted as a cast")
			}
			stackBranch := g.Clone()
			for _, branch := range []*game.Game{g, stackBranch} {
				e.resolveTopOfStack(branch, &TurnLog{})
				token := newestTokenNamed(branch, def.Name)
				requireBlitzEntry(t, branch, token, true, false)
				e.resolveTopOfStack(branch, &TurnLog{})
				requireBlitzEntry(t, branch, permanentForCard(branch, card), true, true)
			}
			battlefieldBranch := g.Clone()
			for _, branch := range []*game.Game{g, stackBranch, battlefieldBranch} {
				token := newestTokenNamed(branch, def.Name)
				blitzEndStep(e, branch)
				if branch.Stack.Size() != 2 {
					t.Fatal("each resolving Blitz spell must schedule its own sacrifice")
				}
				addCardToLibraryNamed(branch, game.Player1, "Draw one")
				addCardToLibraryNamed(branch, game.Player1, "Draw two")
				resolveStackWithTriggers(e, branch, [game.NumPlayers]PlayerAgent{})
				if _, ok := permanentByObjectID(branch, token.ObjectID); ok ||
					!branch.Players[game.Player1].Graveyard.Contains(card) || branch.Players[game.Player1].Hand.Size() != 2 {
					t.Fatal("cloned or replayed sacrifices/draws diverged")
				}
			}
		})
	}
}

func TestBlitzCounteredOrFreeCastCopy(t *testing.T) {
	t.Run("countered", func(t *testing.T) {
		e, g := NewEngine(nil), castCopyProvenanceGame()
		card := addCardToHand(g, game.Player1, blitzFixture())
		fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
		obj := castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
		if !counterStackObject(g, obj.ID) || !g.Players[game.Player1].Graveyard.Contains(card) {
			t.Fatal("could not counter the Blitz spell")
		}
		blitzEndStep(e, g)
		if g.Stack.Size() != 0 || permanentForCard(g, card) != nil {
			t.Fatal("countered spell created a Blitz permanent or sacrifice trigger")
		}
	})
	t.Run("entry replaced", func(t *testing.T) {
		e, g := NewEngine(nil), castCopyProvenanceGame()
		def := blitzFixture()
		def.ReplacementAbilities = []game.ReplacementAbility{game.EntersUnlessPaidElseZoneReplacement(
			"Exile this instead unless you pay {1}.",
			game.ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.O(1)})}, zone.Exile,
		)}
		card := addCardToHand(g, game.Player1, def)
		fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
		castBlitzFixture(t, e, g, action.CastSpell(card, nil, 0, nil))
		e.resolveTopOfStack(g, &TurnLog{})
		blitzEndStep(e, g)
		if !g.Players[game.Player1].Exile.Contains(card) || g.Stack.Size() != 0 || permanentForCard(g, card) != nil {
			t.Fatal("replaced entry must not make a permanent or schedule a sacrifice")
		}
	})
	t.Run("copy then cast without paying mana cost", func(t *testing.T) {
		e, g := NewEngine(nil), castCopyProvenanceGame()
		def := blitzFixture()
		card := addCardToExile(g, game.Player1, def)
		fuseAddMana(g, game.Player1, mana.G, mana.C, mana.C, mana.C)
		obj := castCopyOfExiledCard(t, e, g, card)
		if obj.Blitzed || spellCastEvents(g, obj.ID) != 1 || g.Players[game.Player1].ManaPool.Total() != 4 {
			t.Fatal("free cast is a different alternative, and must count as a cast")
		}
		e.resolveTopOfStack(g, &TurnLog{})
		requireBlitzEntry(t, g, newestTokenNamed(g, def.Name), false, true)
		blitzEndStep(e, g)
		if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(card) {
			t.Fatal("free cast scheduled a sacrifice or moved the original card")
		}
	})
}

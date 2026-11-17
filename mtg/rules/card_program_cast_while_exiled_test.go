package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Psychic Theft (66953dfd-d908-41d3-abb4-309c6a165dc3), {1}{U} Sorcery:
// "Target player reveals their hand. You choose an instant or sorcery card from
// it and exile that card. You may cast that card for as long as it remains
// exiled. At the beginning of the next end step, if you haven't cast the card,
// return it to its owner's hand." Rulings: 2014-02-01 targeting yourself
// reveals your whole hand; 2004-10-04 the card is revealed while exiled; it can
// be cast during the end step after the trigger is put on the stack, before it
// resolves; it can be cast only once: once announced and put on the stack it is
// no longer exiled and Psychic Theft's effect ends. CR 2026-11-06 verified copy:
// 117.1a and 307.1 (casting timing still applies), 400.7 (a card that changes
// zones is a new object), 601.2a (casting moves the card to the stack),
// 601.3 (a permission to cast doesn't change timing or costs), 603.4 (an
// intervening "if" is checked when the trigger event occurs and again on
// resolution), 603.7b-c (a delayed trigger triggers once and still triggers if
// its object left, doing nothing), 611.2b (an effect lasting "for as long as" a
// card remains exiled). The test-local program proves the shared operations;
// the card stays unregistered.
var psychicTheftTestReference = game.ProgramReference{ID: "test/psychic-theft", Version: "1"}

func psychicTheftReturnTrigger() game.DelayedTriggerDef {
	return game.DelayedTriggerDef{
		Timing:                           game.DelayedAtBeginningOfNextEndStep,
		Text:                             "At the beginning of the next end step, if you haven't cast the card, return it to its owner's hand.",
		InterveningIfCapturedCardNotCast: true,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card: game.CardReference{Kind: game.CardReferenceCaptured}, FromZone: zone.Exile, Destination: zone.Hand,
		}}}}.Ability(),
	}
}

func psychicTheftTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: psychicTheftTestReference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Max: 1,
			Filter: game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
		})
		if err != nil || len(revealed.Chosen) == 0 {
			return err
		}
		exiled, err := ctx.MoveCards(cardprogram.CardPlacement{
			Cards: revealed.Chosen, Destination: zone.Exile, CastWhileExiled: true,
		})
		if err != nil || len(exiled) != 1 || exiled[0].Zone != zone.Exile {
			return err
		}
		_, err = ctx.DelayCardTrigger(cardprogram.CardDelayedTrigger{Card: exiled[0], Trigger: psychicTheftReturnTrigger()})
		return err
	}}
}

func psychicTheftTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Psychic Theft", Types: []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
		SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer,
		}}, Program: psychicTheftTestReference}.Ability()),
	}}
}

func psychicTheftInstant(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}), Types: []types.Card{types.Instant},
	}}
}

// resolvePsychicTheft resolves Psychic Theft targeting victim, with its caster
// choosing the named card, and returns that card.
func resolvePsychicTheft(t *testing.T, engine *Engine, g *game.Game, caster, victim game.PlayerID, choose string) id.ID {
	t.Helper()
	addImplementationSpellToStack(g, caster, psychicTheftTestDef(), []game.Target{game.PlayerTarget(victim)})
	var chosen id.ID
	chooser := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		offered := optionCardNames(t, req.Options)
		if req.Player != caster || req.MinChoices != 1 || req.MaxChoices != 1 {
			t.Fatalf("the instant-or-sorcery choice was not one mandatory choice by the caster: %+v", req)
		}
		for _, cardID := range g.Players[victim].Hand.All() {
			card, _ := g.GetCardInstance(cardID)
			matches := card.Def.HasType(types.Instant) || card.Def.HasType(types.Sorcery)
			if slices.Contains(offered, card.Def.Name) != matches {
				t.Fatalf("offered %v: %q offered state does not match being an instant or sorcery", offered, card.Def.Name)
			}
			if card.Def.Name == choose {
				chosen = cardID
			}
		}
		return []int{slices.Index(offered, choose)}
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[caster] = chooser
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	requireCardInZone(t, g, chosen, zone.Exile)
	return chosen
}

func castWhileExiledPermissions(g *game.Game, cardID id.ID) []game.RuleEffect {
	var permissions []game.RuleEffect
	for _, effect := range g.RuleEffects {
		if effect.Kind == game.RuleEffectCastFromZone && effect.AffectedCardID == cardID {
			permissions = append(permissions, effect)
		}
	}
	return permissions
}

func setEndStepPriority(g *game.Game, player game.PlayerID) {
	g.Turn.ActivePlayer = player
	g.Turn.Phase = game.PhaseEnding
	g.Turn.Step = game.StepEnd
	g.Turn.PriorityPlayer = player
}

// TestPsychicTheftCastsStolenSorceryOnceFromOpponentsExile plays Psychic
// Theft's text against an opponent: the caster may cast the exiled sorcery from
// that opponent's exile at sorcery timing, paying its cost, and only once. The
// cast card is countered, and no return triggers at the next end step. Clone
// branches leave the card uncast, so it returns to its owner's hand, or move it
// out of exile, so the trigger still triggers and does nothing (CR 603.7c).
func TestPsychicTheftCastsStolenSorceryOnceFromOpponentsExile(t *testing.T) {
	t.Parallel()
	engine := revealHandEngine(t, psychicTheftTestProgram())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player2, revealHandLand())
	addCardToHand(g, game.Player2, vanillaCreature("Grizzly Bears", 2, 2))
	addCardToHand(g, game.Player2, psychicTheftInstant("Opt", 1))
	divination := addCardToHand(g, game.Player2, evidenceCard("Divination", 3))
	stolen := resolvePsychicTheft(t, engine, g, game.Player1, game.Player2, "Divination")
	if stolen != divination || !g.Players[game.Player2].Exile.Contains(divination) ||
		g.Players[game.Player2].Exile.IsFaceDown(divination) {
		t.Fatal("the chosen sorcery was not exiled face up from its owner's hand")
	}
	permissions := castWhileExiledPermissions(g, divination)
	if len(permissions) != 1 || permissions[0].Controller != game.Player1 || permissions[0].CastFromZone != zone.Exile ||
		!slices.Contains(foreignExileCastableCards(g, game.Player1), divination) {
		t.Fatalf("the caster did not get a permission to cast the exiled card: %+v", permissions)
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedCardID != divination ||
		g.DelayedTriggers[0].CapturedCardZoneVersion != g.CardInstances[divination].ZoneVersion {
		t.Fatal("the end-step return was not bound to the exiled card")
	}
	cast := action.CastSpellFromZone(divination, zone.Exile, nil, 0, nil)

	// Uncast: the owner, who has no permission, can't cast it, and it returns
	// to their hand at the next end step.
	uncast := g.Clone()
	setMainPhasePriority(uncast, game.Player2)
	uncast.Players[game.Player2].ManaPool.Add(mana.C, 3)
	if containsAction(engine.Simulator().LegalActions(uncast, game.Player2), cast) {
		t.Fatal("the exiled card's owner may cast it")
	}
	setEndStepPriority(uncast, game.Player1)
	emitBeginningOfStepEvent(uncast, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(uncast) || uncast.Stack.Size() != 1 || len(uncast.DelayedTriggers) != 0 {
		t.Fatal("the return did not trigger at the next end step")
	}
	engine.resolveTopOfStack(uncast, &TurnLog{})
	requireCardInZone(t, uncast, divination, zone.Hand)
	if !uncast.Players[game.Player2].Hand.Contains(divination) || len(castWhileExiledPermissions(uncast, divination)) != 0 {
		t.Fatal("the uncast card was not returned to its owner's hand without a lingering permission")
	}

	// Moved out of exile by another effect: the permission ends and the
	// return still triggers, but finds nothing (CR 400.7, 603.7c).
	moved := g.Clone()
	moveCardBetweenZones(moved, game.Player2, divination, zone.Exile, zone.Graveyard)
	if len(castWhileExiledPermissions(moved, divination)) != 0 || slices.Contains(foreignExileCastableCards(moved, game.Player1), divination) {
		t.Fatal("the permission outlived the card leaving exile")
	}
	setEndStepPriority(moved, game.Player1)
	emitBeginningOfStepEvent(moved, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(moved) || moved.Stack.Size() != 1 {
		t.Fatal("the return did not trigger after the card left exile uncast")
	}
	log := &TurnLog{}
	engine.resolveTopOfStack(moved, log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "resolved" {
		t.Fatalf("resolves = %+v, want the return to resolve and do nothing", log.Resolves)
	}
	requireCardInZone(t, moved, divination, zone.Graveyard)

	// A sorcery keeps sorcery timing (CR 117.1a, 601.3).
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: addCardToHand(g, game.Player3, evidenceCard("Pending", 1)), Controller: game.Player3})
	if containsAction(engine.Simulator().LegalActions(g, game.Player1), cast) {
		t.Fatal("the exiled sorcery could be cast while the stack was not empty")
	}
	g.Stack.Pop()
	if !containsAction(engine.Simulator().LegalActions(g, game.Player1), cast) {
		t.Fatal("the caster may not cast the exiled sorcery at sorcery timing")
	}
	start := len(g.Events)
	obj := castFromGraveyardOrFail(t, g, engine, cast)
	if obj.Controller != game.Player1 || obj.SourceID != divination || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("the stolen sorcery was not cast by the caster paying its cost: %+v", obj)
	}
	castEvents := 0
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventSpellCast && event.CardID == divination && event.Controller == game.Player1 {
			castEvents++
		}
	}
	if castEvents != 1 {
		t.Fatalf("saw %d cast events for the stolen sorcery, want 1", castEvents)
	}
	if len(castWhileExiledPermissions(g, divination)) != 0 || slices.Contains(foreignExileCastableCards(g, game.Player1), divination) {
		t.Fatal("the permission survived the card being cast")
	}
	if !counterStackObject(g, obj.ID) {
		t.Fatal("could not counter the stolen sorcery")
	}
	requireCardInZone(t, g, divination, zone.Graveyard)
	if !g.Players[game.Player2].Graveyard.Contains(divination) {
		t.Fatal("the countered card did not go to its owner's graveyard")
	}

	// The card was cast, so the return doesn't trigger (CR 603.4).
	setEndStepPriority(g, game.Player1)
	emitBeginningOfStepEvent(g, game.StepEnd)
	if engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 || len(g.DelayedTriggers) != 0 {
		t.Fatal("the return triggered even though the caster cast the card")
	}
	requireCardInZone(t, g, divination, zone.Graveyard)
}

// TestPsychicTheftEndStepResponseCastFromOwnExile plays Psychic Theft
// targeting its caster, who may reveal their own hand. The instant is cast from
// the caster's own exile in response to the end-step return, which then fails
// its intervening "if" on resolution (CR 603.4, Psychic Theft ruling).
func TestPsychicTheftEndStepResponseCastFromOwnExile(t *testing.T) {
	t.Parallel()
	engine := revealHandEngine(t, psychicTheftTestProgram())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, evidenceCard("Divination", 3))
	addCardToHand(g, game.Player1, psychicTheftInstant("Opt", 1))
	stolen := resolvePsychicTheft(t, engine, g, game.Player1, game.Player1, "Opt")
	if !g.Players[game.Player1].Exile.Contains(stolen) || len(castWhileExiledPermissions(g, stolen)) != 1 {
		t.Fatal("the caster's own instant was not exiled with a permission")
	}

	setEndStepPriority(g, game.Player1)
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("the return did not trigger at the next end step")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	cast := action.CastSpellFromZone(stolen, zone.Exile, nil, 0, nil)
	if !containsAction(engine.Simulator().LegalActions(g, game.Player1), cast) {
		t.Fatal("the exiled instant could not be cast in response to the return")
	}
	if spell := castFromGraveyardOrFail(t, g, engine, cast); spell.SourceID != stolen || g.Stack.Size() != 2 {
		t.Fatal("the instant was not cast above the return")
	}
	log := &TurnLog{}
	engine.resolveTopOfStack(g, log)
	requireCardInZone(t, g, stolen, zone.Graveyard)
	engine.resolveTopOfStack(g, log)
	if len(log.Resolves) != 2 || log.Resolves[1].Result != "intervening if false" {
		t.Fatalf("resolves = %+v, want the return removed by its intervening if", log.Resolves)
	}
	requireCardInZone(t, g, stolen, zone.Graveyard)
}

// TestPsychicTheftCastWhileExiledAndCardTriggerValidation rejects a cast
// permission outside exile, a captured-card cast condition without a captured
// card, and card triggers bound to handles this invocation didn't mint or to
// another capture.
func TestPsychicTheftCastWhileExiledAndCardTriggerValidation(t *testing.T) {
	t.Parallel()
	condition := psychicTheftReturnTrigger()
	if err := game.ValidatePrimitive(game.CreateDelayedTrigger{Trigger: condition}, nil); err == nil ||
		!strings.Contains(err.Error(), "requires a CapturedCard") {
		t.Fatalf("error = %v, want a missing captured card error", err)
	}
	reveal := func(ctx cardprogram.Context) (cardprogram.HandRevealResult, error) {
		return ctx.RevealHand(cardprogram.HandReveal{Player: game.TargetPlayerReference(0), Max: 1})
	}
	tests := []struct {
		name    string
		resolve func(cardprogram.Context) error
		want    string
	}{
		{
			name: "cast permission outside exile",
			resolve: func(ctx cardprogram.Context) error {
				revealed, err := reveal(ctx)
				if err != nil {
					return err
				}
				_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Graveyard, CastWhileExiled: true})
				return err
			},
			want: "requires exile",
		},
		{
			name: "stale hand handle",
			resolve: func(ctx cardprogram.Context) error {
				revealed, err := reveal(ctx)
				if err != nil {
					return err
				}
				if _, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Exile}); err != nil {
					return err
				}
				_, err = ctx.DelayCardTrigger(cardprogram.CardDelayedTrigger{Card: revealed.Chosen[0], Trigger: psychicTheftReturnTrigger()})
				return err
			},
			want: "handle from this invocation",
		},
		{
			name: "foreign handle",
			resolve: func(ctx cardprogram.Context) error {
				revealed, err := reveal(ctx)
				if err != nil {
					return err
				}
				foreign := revealed.Chosen[0]
				foreign.Identity.CardID++
				_, err = ctx.DelayCardTrigger(cardprogram.CardDelayedTrigger{Card: foreign, Trigger: psychicTheftReturnTrigger()})
				return err
			},
			want: "handle from this invocation",
		},
		{
			name: "other capture",
			resolve: func(ctx cardprogram.Context) error {
				revealed, err := reveal(ctx)
				if err != nil {
					return err
				}
				trigger := psychicTheftReturnTrigger()
				trigger.CapturedCard = opt.Val(game.LinkedObjectReference("other"))
				_, err = ctx.DelayCardTrigger(cardprogram.CardDelayedTrigger{Card: revealed.Chosen[0], Trigger: trigger})
				return err
			},
			want: "bound only to its card handle",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			engine := revealHandEngine(t, cardprogram.Program{Reference: psychicTheftTestReference, Resolve: test.resolve})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToHand(g, game.Player2, psychicTheftInstant("Opt", 1))
			addImplementationSpellToStack(g, game.Player1, psychicTheftTestDef(), []game.Target{game.PlayerTarget(game.Player2)})
			defer func() {
				failure, ok := recover().(UnsupportedError)
				if !ok || !strings.Contains(failure.Reason, test.want) {
					t.Fatalf("failure = %+v, want an unsupported error containing %q", failure, test.want)
				}
				if len(g.DelayedTriggers) != 0 {
					t.Fatal("a rejected card trigger was scheduled")
				}
			}()
			engine.resolveTopOfStack(g, &TurnLog{})
		})
	}
}

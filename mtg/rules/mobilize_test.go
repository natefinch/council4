package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// mobilizeCreatureDef is a plain creature carrying the reusable Mobilize
// triggered body for the given amount, so attacking it creates and links its
// tapped-and-attacking Warrior tokens.
func mobilizeCreatureDef(amount game.MobilizeAmount) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:               "Mobilize Captain",
		Types:              []types.Card{types.Creature},
		Subtypes:           []types.Sub{types.Soldier},
		Power:              opt.Val(game.PT{Value: 2}),
		Toughness:          opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{game.MobilizeTriggeredBody(amount)},
	}}
}

// plainWarriorDef is an unrelated 1/1 red Warrior creature (not a token) used to
// prove Mobilize's next-end-step sacrifice only touches its own linked tokens.
func plainWarriorDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Unrelated Warrior",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Warrior},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func mobilizeTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

// resolveMobilizeAttack sets up combat with the given source attacking Player2,
// resolves the source's Mobilize triggered body, and returns the created tokens
// and the scheduled delayed trigger.
func resolveMobilizeAttack(t *testing.T, g *game.Game, engine *Engine, source *game.Permanent, amount game.MobilizeAmount) ([]*game.Permanent, game.DelayedTrigger) {
	t.Helper()
	before := map[id.ID]bool{}
	for _, token := range mobilizeTokens(g) {
		before[token.ObjectID] = true
	}
	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceObjectID: source.ObjectID,
			Player:         game.Player2,
		},
	}
	log := TurnLog{}
	engine.resolveAbilityContentWithChoices(g, obj, game.MobilizeTriggeredBody(amount).Content, [game.NumPlayers]PlayerAgent{}, &log)

	var created []*game.Permanent
	for _, token := range mobilizeTokens(g) {
		if !before[token.ObjectID] {
			created = append(created, token)
		}
	}
	if len(g.DelayedTriggers) == 0 {
		t.Fatal("mobilize scheduled no delayed trigger")
	}
	return created, g.DelayedTriggers[len(g.DelayedTriggers)-1]
}

func fireDelayedTrigger(engine *Engine, g *game.Game, delayed game.DelayedTrigger) {
	obj := &game.StackObject{
		Controller:        delayed.Controller,
		SourceID:          delayed.SourceObjectID,
		SourceCardID:      delayed.SourceID,
		CapturedObjectIDs: append([]id.ID(nil), delayed.CapturedObjectIDs...),
	}
	engine.resolveAbilityContentWithChoices(g, obj, delayed.Ability.Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// A fixed Mobilize attack creates N tapped-and-attacking 1/1 red Warrior tokens
// joining the source's attack, then sacrifices exactly those tokens when the
// paired next-end-step delayed trigger fires (CR 702.169).
func TestMobilizeCreatesTappedAttackingTokensAndSacrificesAtNextEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 2}))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	tokens, delayed := resolveMobilizeAttack(t, g, engine, attacker, game.MobilizeAmount{Fixed: 2})

	if len(tokens) != 2 {
		t.Fatalf("created %d tokens, want 2 (Mobilize 2)", len(tokens))
	}
	for _, token := range tokens {
		if !token.Token {
			t.Fatalf("token %d is not a token", token.ObjectID)
		}
		if !token.Tapped {
			t.Fatalf("token %d entered untapped, want tapped", token.ObjectID)
		}
		def, ok := permanentFaceDef(g, token)
		if !ok || def.Name != "Warrior" {
			t.Fatalf("token %d is not a Warrior token: %+v", token.ObjectID, def)
		}
		if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Warrior {
			t.Fatalf("token %d subtypes = %v, want [Warrior]", token.ObjectID, def.Subtypes)
		}
		if len(def.Colors) != 1 || def.Colors[0] != color.Red {
			t.Fatalf("token %d colors = %v, want [Red]", token.ObjectID, def.Colors)
		}
		if !def.Power.Exists || def.Power.Val.Value != 1 || !def.Toughness.Exists || def.Toughness.Val.Value != 1 {
			t.Fatalf("token %d is not 1/1: %+v/%+v", token.ObjectID, def.Power, def.Toughness)
		}
		declaration, ok := attackerDeclarationFor(g, token.ObjectID)
		if !ok {
			t.Fatalf("token %d was not declared attacking", token.ObjectID)
		}
		if declaration.Target.Player != game.Player2 {
			t.Fatalf("token %d attacks Player%d, want Player2 (same as source)", token.ObjectID, declaration.Target.Player)
		}
	}

	if delayed.Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("delayed trigger timing = %v, want DelayedAtBeginningOfNextEndStep", delayed.Timing)
	}
	if len(delayed.CapturedObjectIDs) != 2 {
		t.Fatalf("captured %d objects, want the 2 created tokens", len(delayed.CapturedObjectIDs))
	}

	// The tokens are not sacrificed at end of combat: they remain on the
	// battlefield until the next-end-step trigger fires.
	for _, token := range tokens {
		if _, ok := permanentByObjectID(g, token.ObjectID); !ok {
			t.Fatalf("token %d left the battlefield before the end step", token.ObjectID)
		}
	}

	fireDelayedTrigger(engine, g, delayed)

	for _, token := range tokens {
		if _, ok := permanentByObjectID(g, token.ObjectID); ok {
			t.Fatalf("token %d survived the next end step, want sacrificed", token.ObjectID)
		}
	}
}

// The next-end-step sacrifice touches only the linked created tokens: an
// unrelated Warrior and a token that left and re-entered as a new object both
// survive, because the delayed trigger captured the original token ObjectIDs.
func TestMobilizeSacrificesOnlyLinkedTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 2}))
	unrelated := addCombatPermanent(g, game.Player1, plainWarriorDef())
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	tokens, delayed := resolveMobilizeAttack(t, g, engine, attacker, game.MobilizeAmount{Fixed: 2})
	if len(tokens) != 2 {
		t.Fatalf("created %d tokens, want 2", len(tokens))
	}

	// Simulate one token leaving and re-entering as a brand-new object: remove
	// the original permanent and add a replacement with a fresh ObjectID. The
	// delayed trigger captured the original ID, so the new object is not linked.
	original := tokens[0]
	removePermanentFromBattlefield(g, original.ObjectID)
	reentered := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: original.CardInstanceID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Token:          true,
	}
	g.Battlefield = append(g.Battlefield, reentered)

	fireDelayedTrigger(engine, g, delayed)

	if _, ok := permanentByObjectID(g, tokens[1].ObjectID); ok {
		t.Fatalf("linked token %d survived, want sacrificed", tokens[1].ObjectID)
	}
	if _, ok := permanentByObjectID(g, reentered.ObjectID); !ok {
		t.Fatalf("re-entered token %d was sacrificed, want survives (new object not linked)", reentered.ObjectID)
	}
	if _, ok := permanentByObjectID(g, unrelated.ObjectID); !ok {
		t.Fatalf("unrelated Warrior %d was sacrificed, want survives", unrelated.ObjectID)
	}
}

// Two separate Mobilize resolutions (e.g. multiple attacks or extra combats)
// each schedule an independent next-end-step trigger capturing only its own
// tokens, so firing one leaves the other resolution's tokens untouched.
func TestMobilizeMultipleTriggersLinkIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 1}))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	firstTokens, firstDelayed := resolveMobilizeAttack(t, g, engine, attacker, game.MobilizeAmount{Fixed: 1})
	secondTokens, secondDelayed := resolveMobilizeAttack(t, g, engine, attacker, game.MobilizeAmount{Fixed: 1})

	if len(firstTokens) != 1 || len(secondTokens) != 1 {
		t.Fatalf("created %d and %d tokens, want 1 each", len(firstTokens), len(secondTokens))
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("scheduled %d delayed triggers, want 2 independent triggers", len(g.DelayedTriggers))
	}
	if firstDelayed.CapturedObjectIDs[0] == secondDelayed.CapturedObjectIDs[0] {
		t.Fatalf("both triggers captured the same token %d, want disjoint sets", firstDelayed.CapturedObjectIDs[0])
	}

	// Firing only the first trigger sacrifices only the first resolution's token.
	fireDelayedTrigger(engine, g, firstDelayed)
	if _, ok := permanentByObjectID(g, firstTokens[0].ObjectID); ok {
		t.Fatalf("first token %d survived its own trigger", firstTokens[0].ObjectID)
	}
	if _, ok := permanentByObjectID(g, secondTokens[0].ObjectID); !ok {
		t.Fatalf("second token %d was sacrificed by the first trigger, want independent", secondTokens[0].ObjectID)
	}

	// Firing the second trigger then sacrifices the second resolution's token.
	fireDelayedTrigger(engine, g, secondDelayed)
	if _, ok := permanentByObjectID(g, secondTokens[0].ObjectID); ok {
		t.Fatalf("second token %d survived its own trigger", secondTokens[0].ObjectID)
	}
}

// Dynamic "Mobilize X, where X is the number of creature cards in your
// graveyard" creates one token per creature card in the controller's graveyard.
func TestMobilizeDynamicGraveyardCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	amount := game.MobilizeAmount{Dynamic: game.MobilizeDynamicCreatureCardsInGraveyard}
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(amount))
	for range 3 {
		addCardToGraveyard(g, game.Player1, plainWarriorDef())
	}
	// A noncreature card in the graveyard must not be counted.
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Some Instant",
		Types: []types.Card{types.Instant},
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	tokens, _ := resolveMobilizeAttack(t, g, engine, attacker, amount)
	if len(tokens) != 3 {
		t.Fatalf("created %d tokens, want 3 (creature cards in graveyard)", len(tokens))
	}
}

// When the source has left combat after its attack trigger fired, the tokens
// still join the attack against the defending player recorded on the trigger
// event (CR 702.169b fallback).
func TestMobilizeTokensAttackTriggerEventDefenderWhenSourceLeftCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 1}))
	// Source is no longer among the declared attackers (it left combat), so the
	// live-declaration lookup fails and the trigger-event defender is used.
	g.Combat = &game.CombatState{AttackersDeclared: true}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	tokens, _ := resolveMobilizeAttack(t, g, engine, attacker, game.MobilizeAmount{Fixed: 1})
	if len(tokens) != 1 {
		t.Fatalf("created %d tokens, want 1", len(tokens))
	}
	declaration, ok := attackerDeclarationFor(g, tokens[0].ObjectID)
	if !ok {
		t.Fatalf("token %d was not declared attacking", tokens[0].ObjectID)
	}
	if declaration.Target.Player != game.Player2 {
		t.Fatalf("token attacks Player%d, want Player2 (trigger-event defender)", declaration.Target.Player)
	}
}

// The intrinsic Mobilize creature's attacks trigger matches on
// EventAttackerDeclared for itself regardless of which player it attacks, and
// never on another creature's attack (TriggerSourceSelf).
func TestMobilizeTriggerFiresWhenAttackerDeclared(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 1}))
	body := game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 1})
	pattern := &body.Trigger.Pattern

	for _, defender := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		event := game.Event{
			Kind:         game.EventAttackerDeclared,
			Controller:   game.Player1,
			PermanentID:  attacker.ObjectID,
			AttackTarget: game.AttackTarget{Player: defender},
		}
		if !triggerMatchesEvent(g, attacker, pattern, event) {
			t.Fatalf("mobilize did not trigger when attacking Player%d", defender)
		}
	}

	other := addCombatPermanent(g, game.Player1, mobilizeCreatureDef(game.MobilizeAmount{Fixed: 1}))
	event := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player1,
		PermanentID:  other.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player2},
	}
	if triggerMatchesEvent(g, attacker, pattern, event) {
		t.Fatal("mobilize wrongly triggered on another creature's attack (TriggerSourceSelf)")
	}
}

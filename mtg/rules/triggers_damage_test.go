package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDamageTriggerGoesOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:           game.EventDamageDealt,
		Player:          game.TriggerPlayerOpponent,
		DamageRecipient: game.DamageRecipientPlayer,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	dealPlayerDamage(g, 0, 0, game.Player2, game.Player2, 1, false)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want damage trigger to draw one card", got)
	}
}

func TestCombatDamageTriggerRequiresCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Source:              game.TriggerSourceSelf,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  source.ObjectID,
		Controller:      game.Player1,
		Player:          game.Player2,
		DamageRecipient: game.DamageRecipientPlayer,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("combat-damage trigger matched non-combat damage")
	}
	event.CombatDamage = true
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("combat-damage trigger did not match combat damage")
	}
}

func TestDamageSourceSubjectDoesNotMatchDamageRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	recipient := addCombatCreaturePermanent(g, game.Player1)
	source := addCombatCreaturePermanent(g, game.Player2)
	pattern := &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		Source:               game.TriggerSourceSelf,
		Subject:              game.TriggerSubjectDamageSource,
		DamageRecipient:      game.DamageRecipientPermanent,
		DamageRecipientTypes: []types.Card{types.Creature},
		RequireCombatDamage:  true,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        source.CardInstanceID,
		SourceObjectID:  source.ObjectID,
		Controller:      game.Player2,
		CardID:          recipient.CardInstanceID,
		PermanentID:     recipient.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		CombatDamage:    true,
	}
	if triggerMatchesEvent(g, recipient, pattern, event) {
		t.Fatal("damage-source trigger matched the damage recipient")
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("damage-source trigger did not match the damage source")
	}
	nonCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Attacked Battle",
		Types: []types.Card{types.Battle},
	}})
	event.PermanentID = nonCreature.ObjectID
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("damage-to-creature trigger matched a noncreature permanent recipient")
	}
}

func TestDamageRecipientSubjectDoesNotMatchDamageSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	recipient := addCombatCreaturePermanent(g, game.Player1)
	source := addCombatCreaturePermanent(g, game.Player2)
	pattern := &game.TriggerPattern{
		Event:           game.EventDamageDealt,
		Source:          game.TriggerSourceSelf,
		Subject:         game.TriggerSubjectPermanent,
		DamageRecipient: game.DamageRecipientPermanent,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        source.CardInstanceID,
		SourceObjectID:  source.ObjectID,
		Controller:      game.Player2,
		CardID:          recipient.CardInstanceID,
		PermanentID:     recipient.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("damage-recipient trigger matched the damage source")
	}
	if !triggerMatchesEvent(g, recipient, pattern, event) {
		t.Fatal("damage-recipient trigger did not match the damage recipient")
	}
}

func TestDamageRecipientTriggerUsesLKIAfterRecipientDies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	recipient := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:           game.EventDamageDealt,
		Source:          game.TriggerSourceSelf,
		Subject:         game.TriggerSubjectPermanent,
		DamageRecipient: game.DamageRecipientPermanent,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	source := addCombatCreaturePermanent(g, game.Player2)

	dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player2, recipient, 1, false)
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, recipient.ObjectID); ok {
		t.Fatal("recipient survived damage; test requires LKI trigger source")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("damage-recipient trigger from dead recipient was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want dead recipient trigger to draw one card", got)
	}
}

func TestCombatDamageSourceTriggerUsesLKIAfterSourceDies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	attacker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventDamageDealt,
		Source:               game.TriggerSourceSelf,
		Subject:              game.TriggerSubjectDamageSource,
		DamageRecipient:      game.DamageRecipientPermanent,
		DamageRecipientTypes: []types.Card{types.Creature},
		RequireCombatDamage:  true,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		Blockers: []game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}},
		BlockedAttackers: map[id.ID]bool{attacker.ObjectID: true},
		BlockerOrder:     map[id.ID][]id.ID{attacker.ObjectID: {blocker.ObjectID}},
	}

	combatEngine{}.resolveDamagePass(g, normalCombatDamage, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("attacker survived combat damage; test requires LKI trigger source")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger from dead source was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want dead source trigger to draw one card", got)
	}
}

func TestAttachedPermanentDamageSourceTriggerMatchesDealer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dealer := addCombatCreaturePermanent(g, game.Player1)
	other := addCombatCreaturePermanent(g, game.Player1)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
	}})
	if !attachPermanent(g, equipment, dealer) {
		t.Fatal("attachPermanent failed")
	}
	pattern := &game.TriggerPattern{
		Event:               game.EventDamageDealt,
		Source:              game.TriggerSourceAttachedPermanent,
		Subject:             game.TriggerSubjectDamageSource,
		DamageRecipient:     game.DamageRecipientPlayer,
		RequireCombatDamage: true,
	}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        dealer.CardInstanceID,
		SourceObjectID:  dealer.ObjectID,
		Controller:      game.Player1,
		Player:          game.Player2,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    true,
	}
	if !triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached permanent trigger did not match damage dealt by attached creature")
	}
	event.SourceID = other.CardInstanceID
	event.SourceObjectID = other.ObjectID
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached permanent trigger matched damage dealt by unattached creature")
	}
	event.SourceID = dealer.CardInstanceID
	event.SourceObjectID = dealer.ObjectID
	event.CombatDamage = false
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("combat damage trigger matched noncombat damage")
	}
}

func TestAttachedPermanentDamageSourceTriggerUsesLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceAttachedPermanent,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
				RequireCombatDamage:  true,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	if !attachPermanent(g, equipment, attacker) {
		t.Fatal("attachPermanent failed")
	}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		Blockers: []game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}},
		BlockedAttackers: map[id.ID]bool{attacker.ObjectID: true},
		BlockerOrder:     map[id.ID][]id.ID{attacker.ObjectID: {blocker.ObjectID}},
	}

	combatEngine{}.resolveDamagePass(g, normalCombatDamage, &TurnLog{})
	engine.applyStateBasedActions(g)
	if equipment.AttachedTo.Exists {
		t.Fatal("equipment remained attached after damage source died")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attached permanent damage-source trigger was not put on stack from LKI")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want attached permanent trigger to draw one card", got)
	}
}

func TestAuraDamageSourceTriggerUsesLKIAfterAuraLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Curious Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			game.EnchantStaticAbility(&game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Predicate: game.TargetPredicate{
					PermanentTypes: []types.Card{types.Creature},
				},
			}),
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				Subject:         game.TriggerSubjectDamageSource,
				Player:          game.TriggerPlayerOpponent,
				DamageRecipient: game.DamageRecipientPlayer,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}})
	dealer := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	if !attachPermanent(g, aura, dealer) {
		t.Fatal("attachPermanent failed")
	}

	dealPlayerDamage(g, dealer.CardInstanceID, dealer.ObjectID, game.Player1, game.Player2, 1, true)
	dealer.MarkedDamage = 1
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, dealer.ObjectID); ok {
		t.Fatal("damage source remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Aura remained on battlefield after enchanted creature died")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("departed Aura damage-source trigger was not put on stack from LKI")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want departed Aura trigger to draw one card", got)
	}
}

func TestBecomesBlockedTriggerFiresOnceForMultipleBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	attacker := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventAttackerBecameBlocked,
		Source: game.TriggerSourceSelf,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	firstBlocker := addCombatCreaturePermanent(g, game.Player2)
	secondBlocker := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	declare := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: firstBlocker.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: secondBlocker.ObjectID, Blocking: attacker.ObjectID},
	}))

	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatal("applyDeclareBlockers() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("becomes-blocked trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one becomes-blocked trigger", got)
	}
}

func TestDrawTriggerChoosesDeterministicLegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})

	if _, ok := engine.drawCard(g, game.Player1, false); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 39 {
		t.Fatalf("player 2 life = %d, want deterministic opponent target to lose 1", got)
	}
}

func TestTriggerTargetChoiceCanBeMadeByAgent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCardDrawn,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0)}}}, []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player1, false); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("draw trigger was not put on stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("player 2 life = %d, want agent to choose another target", got)
	}
	if got := g.Players[game.Player3].Life; got != 39 {
		t.Fatalf("player 3 life = %d, want chosen target to lose 1", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceTarget || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want recorded target choice without fallback", log.Choices)
	}
}

func TestOptionalTriggeredAbilityChoiceHappensOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Triggering Drawn"}})
	addOptionalTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventCardDrawn,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	log := TurnLog{}

	if _, ok := engine.drawCard(g, game.Player2, false); !ok {
		t.Fatal("drawCard() = false, want true")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("optional trigger was not put on stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want optional trigger on stack before may choice", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("player 1 hand size = %d, want optional trigger declined on resolution", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceMay || log.Choices[0].Selected[0] != 0 {
		t.Fatalf("choices = %+v, want no may choice recorded", log.Choices)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "declined" {
		t.Fatalf("resolves = %+v, want declined optional trigger", log.Resolves)
	}
}

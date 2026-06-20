package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestTriggerPatternCanRequireStackSpellTargetingSelf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventObjectBecameTarget,
		Source:               game.TriggerSourceSelf,
		MatchStackObjectKind: true,
		StackObjectKind:      game.StackSpell,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-target trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want spell-target trigger to draw one card", got)
	}
}

func TestTriggerPatternStackSpellDoesNotMatchAbilityTargetingSelf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventObjectBecameTarget,
		Source:               game.TriggerSourceSelf,
		MatchStackObjectKind: true,
		StackObjectKind:      game.StackSpell,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	ability := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	g.Stack.Push(ability)

	emitTargetEvents(g, ability)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-target trigger matched an activated ability")
	}
}

func TestTriggerPatternControlledCreatureAttackMatchesOnlyIntendedSubject(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:      game.EventAttackerDeclared,
		Controller: game.TriggerControllerYou,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	event := game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: attacker.ObjectID,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controlled-creature attack pattern did not match intended attacker")
	}

	opponent := addCombatCreaturePermanent(g, game.Player2)
	event.Controller = game.Player2
	event.PermanentID = opponent.ObjectID
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controlled-creature attack pattern matched opponent's attacker")
	}

	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Attacking Land",
		Types: []types.Card{types.Land},
	}})
	event.Controller = game.Player1
	event.PermanentID = land.ObjectID
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controlled-creature attack pattern matched noncreature attacker")
	}
}

func TestTriggerPatternSeparatesTargetSubjectAndCauseControllers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	subject := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:           game.EventObjectBecameTarget,
		Controller:      game.TriggerControllerYou,
		CauseController: game.TriggerControllerOpponent,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	event := game.Event{
		Kind:        game.EventObjectBecameTarget,
		Controller:  game.Player2,
		PermanentID: subject.ObjectID,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("opponent-controlled cause targeting controlled creature did not match")
	}
	event.Controller = game.Player1
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controller's own cause matched opponent-cause relation")
	}
	event.Controller = game.Player2
	event.PermanentID = addCombatCreaturePermanent(g, game.Player2).ObjectID
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("opponent-controlled subject matched controlled-subject relation")
	}
}

func TestTriggerPatternActivatedAbilityMatchesActorSourceAndManaFilter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	abilitySource := addCombatCreaturePermanent(g, game.Player2)
	pattern := &game.TriggerPattern{
		Event:              game.EventAbilityActivated,
		Player:             game.TriggerPlayerOpponent,
		ExcludeManaAbility: true,
		SubjectSelection: game.Selection{
			RequiredTypesAny: []types.Card{types.Creature, types.Land},
		},
	}
	event := game.Event{
		Kind:        game.EventAbilityActivated,
		Player:      game.Player2,
		Controller:  game.Player2,
		PermanentID: abilitySource.ObjectID,
	}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("opponent nonmana creature ability did not match")
	}
	unrestricted := *pattern
	unrestricted.ExcludeManaAbility = false
	if triggerMatchesEvent(g, source, &unrestricted, event) {
		t.Fatal("unrestricted ability-activated pattern matched an incomplete runtime event stream")
	}
	event.ManaAbility = true
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("mana ability matched nonmana ability pattern")
	}
	event.ManaAbility = false
	event.Player = game.Player1
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("controller ability matched opponent ability pattern")
	}
}

func TestTriggerPatternAttachedCreatureBlocksMatchesOnlyAttachment(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blocker := addCombatCreaturePermanent(g, game.Player1)
	other := addCombatCreaturePermanent(g, game.Player1)
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Equipment",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
	}})
	if !attachPermanent(g, equipment, blocker) {
		t.Fatal("attachPermanent failed")
	}
	pattern := &game.TriggerPattern{
		Event:  game.EventBlockerDeclared,
		Source: game.TriggerSourceAttachedPermanent,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		},
	}
	event := game.Event{
		Kind:        game.EventBlockerDeclared,
		Controller:  game.Player1,
		PermanentID: blocker.ObjectID,
	}
	if !triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached-creature block pattern did not match equipped blocker")
	}
	event.PermanentID = other.ObjectID
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached-creature block pattern matched an unattached blocker")
	}
	event.Kind = game.EventAttackerDeclared
	event.PermanentID = blocker.ObjectID
	if triggerMatchesEvent(g, equipment, pattern, event) {
		t.Fatal("attached-creature block pattern matched adjacent attack event")
	}
}

func TestCombatTriggerPatternTypedSelectionsAndRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	attacker := addCombatCreaturePermanent(g, game.Player1, game.Flying)
	blocker := addCombatCreaturePermanent(g, game.Player2)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Defending Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})

	t.Run("attack recipient", func(t *testing.T) {
		event := game.Event{
			Kind:        game.EventAttackerDeclared,
			Controller:  game.Player1,
			Player:      game.Player2,
			PermanentID: attacker.ObjectID,
			AttackTarget: game.AttackTarget{
				Player:         game.Player2,
				PlaneswalkerID: planeswalker.ObjectID,
			},
		}
		pattern := game.TriggerPattern{
			Event:           game.EventAttackerDeclared,
			AttackRecipient: game.AttackRecipientPlaneswalker,
			Player:          game.TriggerPlayerOpponent,
			AttackRecipientSelection: game.Selection{
				RequiredTypes: []types.Card{types.Planeswalker},
				Controller:    game.ControllerNotYou,
			},
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("planeswalker recipient pattern did not match")
		}
		pattern.AttackRecipient = game.AttackRecipientPlayer
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("player recipient pattern matched a planeswalker attack")
		}
	})

	t.Run("related blocker subject", func(t *testing.T) {
		pattern := game.TriggerPattern{
			Event: game.EventBlockerDeclared,
			RelatedSubjectSelection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Keyword:       game.Flying,
			},
		}
		event := game.Event{
			Kind:               game.EventBlockerDeclared,
			Controller:         game.Player2,
			PermanentID:        blocker.ObjectID,
			RelatedPermanentID: attacker.ObjectID,
			BlockedAttackerID:  attacker.ObjectID,
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("related flying attacker pattern did not match")
		}
		event.RelatedPermanentID = blocker.ObjectID
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("related flying attacker pattern matched a nonflying creature")
		}
	})

	t.Run("damage source and recipient", func(t *testing.T) {
		pattern := game.TriggerPattern{
			Event:      game.EventDamageDealt,
			Controller: game.TriggerControllerYou,
			Player:     game.TriggerPlayerOpponent,
			Subject:    game.TriggerSubjectDamageSource,
			DamageSourceSelection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			},
			DamageRecipient: game.DamageRecipientPermanent,
			DamageRecipientSelection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerNotYou,
			},
			RequireNonCombatDamage: true,
		}
		event := game.Event{
			Kind:            game.EventDamageDealt,
			Controller:      game.Player1,
			Player:          game.Player2,
			SourceObjectID:  attacker.ObjectID,
			PermanentID:     blocker.ObjectID,
			DamageRecipient: game.DamageRecipientPermanent,
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("typed damage source and recipient pattern did not match")
		}
		pattern.Controller = game.TriggerControllerAny
		event.SourceObjectID = planeswalker.ObjectID
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("creature damage-source Selection matched a planeswalker")
		}
		event.SourceObjectID = attacker.ObjectID
		event.CombatDamage = true
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("noncombat pattern matched combat damage")
		}
	})

	t.Run("player or permanent damage union", func(t *testing.T) {
		pattern := game.TriggerPattern{
			Event:           game.EventDamageDealt,
			Player:          game.TriggerPlayerOpponent,
			DamageRecipient: game.DamageRecipientPlayer | game.DamageRecipientPermanent,
			DamageRecipientSelection: game.Selection{
				RequiredTypes: []types.Card{types.Planeswalker},
			},
		}
		event := game.Event{
			Kind:            game.EventDamageDealt,
			Player:          game.Player2,
			DamageRecipient: game.DamageRecipientPlayer,
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("player branch of player-or-planeswalker damage pattern did not match")
		}
	})

	t.Run("ability source damage recipient", func(t *testing.T) {
		pattern := game.TriggerPattern{
			Event:                   game.EventDamageDealt,
			DamageRecipient:         game.DamageRecipientPermanent,
			DamageRecipientIsSource: true,
		}
		event := game.Event{
			Kind:            game.EventDamageDealt,
			PermanentID:     source.ObjectID,
			CardID:          source.CardInstanceID,
			DamageRecipient: game.DamageRecipientPermanent,
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("ability-source damage recipient did not match")
		}
		event.PermanentID = blocker.ObjectID
		event.CardID = blocker.CardInstanceID
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("ability-source damage recipient matched another permanent")
		}
	})

	t.Run("attached permanent controller step", func(t *testing.T) {
		source.AttachedTo = opt.Val(attacker.ObjectID)
		pattern := game.TriggerPattern{
			Event: game.EventBeginningOfStep,
			Step:  game.StepUpkeep,
			StepPlayerSourceAttachedSelection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			},
		}
		event := game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player1}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("attached permanent controller step did not match")
		}
		event.Player = game.Player2
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("attached permanent controller step matched another player's step")
		}
	})

	t.Run("player event ordinal this turn", func(t *testing.T) {
		pattern := game.TriggerPattern{
			Event:                      game.EventCardDrawn,
			Player:                     game.TriggerPlayerYou,
			PlayerEventOrdinalThisTurn: 2,
		}
		event := game.Event{
			Kind:                       game.EventCardDrawn,
			Player:                     game.Player1,
			PlayerEventOrdinalThisTurn: 2,
		}
		if !triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("second player event did not match")
		}
		event.PlayerEventOrdinalThisTurn = 1
		if triggerMatchesEvent(g, source, &pattern, event) {
			t.Fatal("first player event matched second-event pattern")
		}
	})
}

// TestUnionAttackBecameTargetTriggerFiresOnAttackAndSpellTarget exercises the
// event-union trigger "Whenever this creature attacks or becomes the target of
// a spell": the self-scoped ability must fire on the attack event and on
// becoming the target of a spell, but the stack-object filter (which is scoped
// to the became-target event) must not reject the attack constituent, and must
// still reject becoming the target of an activated ability.
func TestUnionAttackBecameTargetTriggerFiresOnAttackAndSpellTarget(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:                game.EventObjectBecameTarget,
		UnionEvent:           game.EventAttackerDeclared,
		Source:               game.TriggerSourceSelf,
		MatchStackObjectKind: true,
		StackObjectKind:      game.StackSpell,
	}
	source := addTriggeredPermanent(g, game.Player1, pattern,
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("attack constituent did not fire despite the spell-target filter")
	}

	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(source.ObjectID)},
	}
	g.Stack.Push(spell)
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:          game.EventObjectBecameTarget,
		Controller:    game.Player2,
		PermanentID:   source.ObjectID,
		StackObjectID: spell.ID,
	}) {
		t.Fatal("became-target-of-a-spell constituent did not fire")
	}

	ability := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(source.ObjectID)},
	}
	g.Stack.Push(ability)
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:          game.EventObjectBecameTarget,
		Controller:    game.Player2,
		PermanentID:   source.ObjectID,
		StackObjectID: ability.ID,
	}) {
		t.Fatal("spell-only filter wrongly matched becoming the target of an ability")
	}
}

// TestUnionEnterDiesTriggerFiresOnEnterAndDeath covers the event-union trigger
// "When this creature enters or dies": the self-scoped ability must fire on the
// enters-the-battlefield event and on the dies event, sharing one subject.
func TestUnionEnterDiesTriggerFiresOnEnterAndDeath(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pattern := &game.TriggerPattern{
		Event:      game.EventPermanentEnteredBattlefield,
		UnionEvent: game.EventPermanentDied,
		Source:     game.TriggerSourceSelf,
	}
	source := addTriggeredPermanent(g, game.Player1, pattern,
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("enter constituent did not fire")
	}
	if !triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventPermanentDied,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("dies constituent did not fire")
	}
	if triggerMatchesEvent(g, source, pattern, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}) {
		t.Fatal("unrelated attack event wrongly matched the enter-or-dies union")
	}
}

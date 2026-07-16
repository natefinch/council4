package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func questingBeastRulesDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Questing Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind: game.RuleEffectCombatDamageCantBePrevented,
				AffectedSelection: game.Selection{
					Controller:    game.ControllerYou,
					RequiredTypes: []types.Card{types.Creature},
				},
			}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:               game.EventDamageDealt,
					Source:              game.TriggerSourceSelf,
					Subject:             game.TriggerSubjectDamageSource,
					Player:              game.TriggerPlayerOpponent,
					DamageRecipient:     game.DamageRecipientPlayer,
					RequireCombatDamage: true,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target planeswalker that player controls",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes:           []types.Card{types.Planeswalker},
						ControlledByEventPlayer: true,
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:       game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventDamage}),
					Recipient:    game.AnyTargetDamageRecipient(0),
					DamageSource: opt.Val(game.EventPermanentReference()),
				}}},
			}.Ability(),
		}},
	}}
}

func TestQuestingBeastBlockerRestrictionUsesLivePower(t *testing.T) {
	t.Run("pump above threshold allows block", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{
			Kind:  game.BlockerRestrictionPowerLessOrEqual,
			Power: 2,
		})
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		blocker.Counters.Add(counter.PlusOnePlusOne, 1)
		g.Turn.Phase = game.PhaseCombat
		g.Turn.Step = game.StepDeclareBlockers
		g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}}}

		payload := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}}))
		if !NewEngine(nil).applyDeclareBlockers(g, game.Player2, payload) {
			t.Fatal("live power 3 blocker was rejected")
		}
	})

	t.Run("shrink to threshold prohibits block", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{
			Kind:  game.BlockerRestrictionPowerLessOrEqual,
			Power: 2,
		})
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		blocker.Counters.Add(counter.MinusOneMinusOne, 1)
		g.Turn.Phase = game.PhaseCombat
		g.Turn.Step = game.StepDeclareBlockers
		g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}}}

		payload := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}}))
		if NewEngine(nil).applyDeclareBlockers(g, game.Player2, payload) {
			t.Fatal("live power 2 blocker was allowed")
		}
	})
}

func TestQuestingBeastCombatDamageCantBePreventedButCanBeReplaced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	addReplacementPermanent(t, g, game.Player3, damagePreventionCardDef(&game.DamagePreventionSpec{Amount: 20}))

	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{
		All:        true,
		CombatOnly: true,
		Global:     true,
	}, nil)
	resolveInstruction(engine, g, &game.StackObject{
		Controller: game.Player3,
		Targets:    []game.Target{game.PermanentTarget(source.ObjectID)},
	}, game.PreventDamage{
		Object:   game.TargetPermanentReference(0),
		All:      true,
		BySource: true,
	}, nil)

	if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player3, 3, true); dealt != 6 {
		t.Fatalf("unpreventable replaced combat damage = %d, want 6", dealt)
	}
	if dealt := dealPlayerDamage(g, other.CardInstanceID, other.ObjectID, game.Player2, game.Player3, 3, true); dealt != 0 {
		t.Fatalf("other creature combat damage = %d, want prevented", dealt)
	}
	if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player3, 3, false); dealt != 0 {
		t.Fatalf("noncombat damage = %d, want prevented", dealt)
	}
}

func TestQuestingBeastUnpreventableCombatDamageDoesNotConsumeShieldCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	target.Counters.Add(counter.Shield, 1)

	if dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, target, 3, true); dealt != 3 {
		t.Fatalf("combat damage = %d, want 3", dealt)
	}

	if target.Counters.Get(counter.Shield) != 1 {
		t.Fatalf("shield counters = %d, want 1", target.Counters.Get(counter.Shield))
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
}

func TestQuestingBeastCombatDamageBypassesProtectionPrevention(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	protected := addProtectionFromTypesPermanent(g, game.Player2, types.Creature)

	if dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, protected, 3, true); dealt != 3 {
		t.Fatalf("combat damage through protection = %d, want 3", dealt)
	}
	protected.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, protected, 3, false); dealt != 0 {
		t.Fatalf("noncombat damage through protection = %d, want prevented", dealt)
	}
}

func TestQuestingBeastUnpreventableCombatDamageOrdersRedirectionWithReplacement(t *testing.T) {
	for _, test := range []struct {
		name   string
		prefer string
		want   int
	}{
		{name: "redirect first", prefer: "redirection", want: 3},
		{name: "multiply first", prefer: "filtered", want: 6},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCombatPermanent(g, game.Player1, questingBeastRulesDef())
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
			addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
				Multiplier:                  2,
				RecipientOpponent:           true,
				RecipientOpponentPlayerOnly: true,
				Controller:                  game.TriggerControllerYou,
			}))
			redirect := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
			g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
				Kind:           game.RuleEffectRedirectDamageToSource,
				Controller:     game.Player2,
				AffectedPlayer: game.PlayerYou,
				SourceObjectID: redirect.ObjectID,
				Duration:       game.DurationPermanent,
			})
			engine := NewEngine(nil)
			engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{
				game.Player2: replacementChoosingAgent{prefer: test.prefer},
			}, &TurnLog{})
			defer g.ClearChoiceContext()

			life := g.Players[game.Player2].Life
			if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 3, true); dealt != test.want {
				t.Fatalf("damage dealt = %d, want %d", dealt, test.want)
			}
			if g.Players[game.Player2].Life != life {
				t.Fatal("redirected damage changed player life")
			}
			if redirect.MarkedDamage != test.want {
				t.Fatalf("redirect target marked damage = %d, want %d", redirect.MarkedDamage, test.want)
			}
		})
	}
}

func TestQuestingBeastPreventionProhibitionFollowsLiveController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	questing := addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	player1Creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	player2Creature := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{
		All:        true,
		CombatOnly: true,
		Global:     true,
	}, nil)

	if dealt := dealPlayerDamage(g, player1Creature.CardInstanceID, player1Creature.ObjectID, game.Player1, game.Player3, 3, true); dealt != 3 {
		t.Fatalf("Player1 creature damage before control change = %d, want 3", dealt)
	}
	questing.Controller = game.Player2
	if dealt := dealPlayerDamage(g, player1Creature.CardInstanceID, player1Creature.ObjectID, game.Player1, game.Player3, 3, true); dealt != 0 {
		t.Fatalf("Player1 creature damage after control change = %d, want prevented", dealt)
	}
	if dealt := dealPlayerDamage(g, player2Creature.CardInstanceID, player2Creature.ObjectID, game.Player2, game.Player3, 3, true); dealt != 3 {
		t.Fatalf("Player2 creature damage after control change = %d, want 3", dealt)
	}
}

func TestQuestingBeastDamageTriggersCorrelateAmountPlayerAndTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	questing := addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	player2Walker := addLoyaltyPlaneswalker(g, game.Player2, 20)
	player3Walker := addLoyaltyPlaneswalker(g, game.Player3, 20)

	dealPlayerDamage(g, questing.CardInstanceID, questing.ObjectID, game.Player1, game.Player2, 3, true)
	dealPlayerDamage(g, questing.CardInstanceID, questing.ObjectID, game.Player1, game.Player3, 5, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage triggers did not reach the stack")
	}
	objects := g.Stack.Objects()
	if len(objects) != 2 {
		t.Fatalf("stack objects = %d, want 2", len(objects))
	}
	for _, object := range objects {
		if len(object.Targets) != 1 || object.Targets[0].Kind != game.TargetPermanent {
			t.Fatalf("targets = %#v", object.Targets)
		}
		switch object.TriggerEvent.Player {
		case game.Player2:
			if object.TriggerEvent.Amount != 3 || object.Targets[0].PermanentID != player2Walker.ObjectID {
				t.Fatalf("Player2 trigger = %#v", object)
			}
		case game.Player3:
			if object.TriggerEvent.Amount != 5 || object.Targets[0].PermanentID != player3Walker.ObjectID {
				t.Fatalf("Player3 trigger = %#v", object)
			}
		default:
			t.Fatalf("unexpected damaged player %v", object.TriggerEvent.Player)
		}
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if got := player2Walker.Counters.Get(counter.Loyalty); got != 17 {
		t.Fatalf("Player2 planeswalker loyalty = %d, want 17", got)
	}
	if got := player3Walker.Counters.Get(counter.Loyalty); got != 15 {
		t.Fatalf("Player3 planeswalker loyalty = %d, want 15", got)
	}
}

func TestQuestingBeastDamageTargetMustRemainControlledByDamagedPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	questing := addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	walker := addLoyaltyPlaneswalker(g, game.Player2, 10)

	dealPlayerDamage(g, questing.CardInstanceID, questing.ObjectID, game.Player1, game.Player2, 4, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger did not reach the stack")
	}
	walker.Controller = game.Player3
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := walker.Counters.Get(counter.Loyalty); got != 10 {
		t.Fatalf("illegal target loyalty = %d, want unchanged 10", got)
	}
}

func TestQuestingBeastTriggeredDamageUsesLiveSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	questing := addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	walker := addLoyaltyPlaneswalker(g, game.Player2, 10)

	dealPlayerDamage(g, questing.CardInstanceID, questing.ObjectID, game.Player1, game.Player2, 4, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger did not reach the stack")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Controller != game.Player1 {
		t.Fatalf("trigger controller = %#v, want event-time controller Player1", top)
	}
	questing.Controller = game.Player3
	engine.resolveTopOfStack(g, &TurnLog{})
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.PermanentID == walker.ObjectID &&
			event.SourceID == questing.CardInstanceID &&
			event.SourceObjectID == questing.ObjectID &&
			event.Controller == game.Player3 &&
			event.Amount == 4 &&
			!event.CombatDamage
	})
}

func TestQuestingBeastTriggerControllerSnapshotsWhenDamageIsDealt(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	questing := addCombatPermanent(g, game.Player1, questingBeastRulesDef())
	addLoyaltyPlaneswalker(g, game.Player2, 10)

	dealPlayerDamage(g, questing.CardInstanceID, questing.ObjectID, game.Player1, game.Player2, 4, true)
	questing.Controller = game.Player3
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger did not reach the stack")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Controller != game.Player1 {
		t.Fatalf("trigger controller = %#v, want event-time controller Player1", top)
	}
}

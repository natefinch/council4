package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func testVirtuousRoleDef() *game.CardDef {
	count := game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Enchantment},
			Controller:    game.ControllerYou,
		}),
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Virtuous Role",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura, types.Role},
		StaticAbilities: []game.StaticAbility{
			game.EnchantStaticAbility(&game.TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
				}),
			}),
			{ContinuousEffects: []game.ContinuousEffect{{
				Layer:                 game.LayerPowerToughnessModify,
				Group:                 game.AttachedObjectGroup(game.SourcePermanentReference()),
				PowerDeltaDynamic:     opt.Val(count),
				ToughnessDeltaDynamic: opt.Val(count),
			}}},
		},
	}}
}

func addRoleTestCreature(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func addTestEnchantment(g *game.Game, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Test Enchantment", Types: []types.Card{types.Enchantment}}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func TestCreatedAuraTokenEntersAttached(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addRoleTestCreature(g, game.Player1, "Target")
	created, ok := createTokenPermanentsCollectingAttachedWithChoices(
		NewEngine(nil), g, game.Player1, testVirtuousRoleDef(), 1, false, target,
		[game.NumPlayers]PlayerAgent{}, nil,
	)
	if !ok || len(created) != 1 {
		t.Fatalf("created = %d, ok = %v", len(created), ok)
	}
	role := created[0]
	if role.Owner != game.Player1 || role.Controller != game.Player1 ||
		!role.AttachedTo.Exists || role.AttachedTo.Val != target.ObjectID ||
		!permanentIDsContain(target.Attachments, role.ObjectID) {
		t.Fatalf("created Role = %#v, target attachments = %v", role, target.Attachments)
	}
}

func TestRoleRuleReplacementAndCoexistence(t *testing.T) {
	t.Run("same controller keeps newest token and old token ceases", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addRoleTestCreature(g, game.Player1, "Target")
		oldRole, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
		newRole, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
		attachPermanent(g, oldRole, target)
		attachPermanent(g, newRole, target)

		_, deaths := NewEngine(nil).applyStateBasedActionsWithDeaths(g)
		if _, ok := permanentByObjectID(g, oldRole.ObjectID); ok {
			t.Fatal("older Role remained on the battlefield")
		}
		if _, ok := permanentByObjectID(g, newRole.ObjectID); !ok {
			t.Fatal("newer Role left the battlefield")
		}
		if !deathReasonFound(deaths, oldRole.ObjectID, PermanentDeathReasonRoleRule) {
			t.Fatalf("deaths = %#v", deaths)
		}
		if len(g.Players[game.Player1].Graveyard.All()) != 0 {
			t.Fatal("Role token did not cease to exist after entering the graveyard")
		}
	})

	t.Run("different controllers coexist until control changes", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addRoleTestCreature(g, game.Player1, "Target")
		first, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
		second, _ := createTokenPermanent(g, game.Player2, testVirtuousRoleDef())
		attachPermanent(g, first, target)
		attachPermanent(g, second, target)
		NewEngine(nil).applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, first.ObjectID); !ok {
			t.Fatal("first player's Role did not coexist")
		}
		if _, ok := permanentByObjectID(g, second.ObjectID); !ok {
			t.Fatal("second player's Role did not coexist")
		}

		second.Controller = game.Player1
		NewEngine(nil).applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, first.ObjectID); ok {
			t.Fatal("older Role remained after the newer Role changed controllers")
		}
		if _, ok := permanentByObjectID(g, second.ObjectID); !ok {
			t.Fatal("newer Role did not remain after control change")
		}
	})

	t.Run("attachment change forms a new competing group", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		firstTarget := addRoleTestCreature(g, game.Player1, "First")
		secondTarget := addRoleTestCreature(g, game.Player1, "Second")
		first, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
		second, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
		attachPermanent(g, first, firstTarget)
		attachPermanent(g, second, secondTarget)
		NewEngine(nil).applyStateBasedActions(g)

		attachPermanent(g, first, secondTarget)
		NewEngine(nil).applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, first.ObjectID); ok {
			t.Fatal("older moved Role remained beside newer Role")
		}
		if _, ok := permanentByObjectID(g, second.ObjectID); !ok {
			t.Fatal("newer Role did not remain")
		}
	})
}

func TestVirtuousRoleDynamicPumpTracksControlAndAttachment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addRoleTestCreature(g, game.Player1, "First")
	second := addRoleTestCreature(g, game.Player1, "Second")
	role, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
	attachPermanent(g, role, first)

	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("power with Role alone = %d, want 3", got)
	}
	addTestEnchantment(g, game.Player1)
	if got := effectivePower(g, first); got != 4 {
		t.Fatalf("power with two enchantments = %d, want 4", got)
	}

	role.Controller = game.Player2
	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("power after Role control change = %d, want 3", got)
	}
	addTestEnchantment(g, game.Player2)
	if got := effectivePower(g, first); got != 4 {
		t.Fatalf("power after new controller gains enchantment = %d, want 4", got)
	}

	attachPermanent(g, role, second)
	if got := effectivePower(g, first); got != 2 {
		t.Fatalf("old attachment power = %d, want 2", got)
	}
	if got := effectivePower(g, second); got != 4 {
		t.Fatalf("new attachment power = %d, want 4", got)
	}
}

func TestEnchantedCombatDamageTriggerUsesEventSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ellivere := addRoleTestCreature(g, game.Player1, "Ellivere")
	damageSource := addRoleTestCreature(g, game.Player1, "Enchanted Attacker")
	role, _ := createTokenPermanent(g, game.Player1, testVirtuousRoleDef())
	attachPermanent(g, role, damageSource)
	pattern := game.TriggerPattern{
		Event:                    game.EventDamageDealt,
		Controller:               game.TriggerControllerYou,
		Subject:                  game.TriggerSubjectDamageSource,
		DamageSourceSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
		DamageSourceWasEnchanted: true,
		DamageRecipient:          game.DamageRecipientPlayer,
		RequireCombatDamage:      true,
	}
	g.CardInstances[ellivere.CardInstanceID].Def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: pattern,
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}}
	event := game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  damageSource.ObjectID,
		Controller:      game.Player1,
		Player:          game.Player2,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    true,
		Amount:          2,
	}
	emitEvent(g, event)
	recorded := g.Events[len(g.Events)-1]
	if !recorded.DamageSourceWasEnchanted {
		t.Fatal("damage event did not snapshot enchanted state")
	}
	if len(recorded.TriggeredAbilities) != 1 ||
		recorded.TriggeredAbilities[0].Controller != game.Player1 {
		t.Fatalf("captured triggers = %#v", recorded.TriggeredAbilities)
	}
	event.Player = game.Player3
	emitEvent(g, event)
	secondRecorded := g.Events[len(g.Events)-1]
	if len(secondRecorded.TriggeredAbilities) != 1 {
		t.Fatalf("second player-damage event captured %d triggers, want 1", len(secondRecorded.TriggeredAbilities))
	}

	damageSource.Controller = game.Player2
	detachPermanent(g, role)
	if !triggerMatchesEvent(g, ellivere, &pattern, recorded) {
		t.Fatal("event-time enchanted/control snapshot did not match after live state changed")
	}
	ellivere.Controller = game.Player2
	if recorded.TriggeredAbilities[0].Controller != game.Player1 {
		t.Fatal("captured trigger controller changed with its source")
	}
	recorded.DamageSourceWasEnchanted = false
	if triggerMatchesEvent(g, ellivere, &pattern, recorded) {
		t.Fatal("unenchanted damage event matched enchanted-source trigger")
	}
}

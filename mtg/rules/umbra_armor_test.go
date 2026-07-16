package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func umbraArmorAura(g *game.Game, controller game.PlayerID, target *game.Permanent, name string, extra ...game.StaticAbility) *game.Permanent {
	abilities := []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{
			game.EnchantKeyword{Target: game.TargetSpec{
				Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				}),
			}},
		},
	}, game.UmbraArmorStaticBody}
	abilities = append(abilities, extra...)
	aura := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Enchantment},
		Subtypes:        []types.Sub{types.Aura},
		StaticAbilities: abilities,
	}})
	if !attachPermanent(g, aura, target) {
		panic("test Umbra Aura did not attach")
	}
	return aura
}

func umbraGrantedAttackAbility() game.StaticAbility {
	granted := &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Untap{Group: game.PlayerControlledGroup(
				game.ControllerReference(),
				game.Selection{RequiredTypes: []types.Card{types.Land}},
			)},
		}}}.Ability(),
	}
	return game.StaticAbility{ContinuousEffects: []game.ContinuousEffect{{
		Layer:        game.LayerAbility,
		Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
		AddAbilities: []game.Ability{granted},
	}}}
}

func tappedLand(g *game.Game, controller game.PlayerID) *game.Permanent {
	land := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Land",
		Types: []types.Card{types.Land},
	}})
	land.Tapped = true
	return land
}

func TestUmbraGrantedAttackAbilityUsesCreatureControllerAndSurvivesAuraLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	aura := umbraArmorAura(g, game.Player1, creature, "Bear Umbra", umbraGrantedAttackAbility())
	auraControllersLand := tappedLand(g, game.Player1)
	creatureControllersLand := tappedLand(g, game.Player2)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: creature.ObjectID,
		Target:   game.AttackTarget{Player: game.Player3},
	}}}

	emitEvent(g, game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player2,
		PermanentID:  creature.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player3},
	})
	creature.Controller = game.Player3
	newControllerLand := tappedLand(g, game.Player3)
	if !movePermanentToZone(g, aura, zone.Graveyard) {
		t.Fatal("failed to remove Aura after attack trigger")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("granted attack trigger was not captured")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if creatureControllersLand.Tapped {
		t.Fatal("granted ability did not untap the attacking creature controller's land")
	}
	if !auraControllersLand.Tapped {
		t.Fatal("granted ability untapped the Aura controller's land")
	}
	if !newControllerLand.Tapped {
		t.Fatal("granted ability followed the creature's post-trigger control change")
	}
}

func TestUmbraGrantedAttackAbilityStopsWhenAuraLeavesBeforeAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := umbraArmorAura(g, game.Player1, creature, "Bear Umbra", umbraGrantedAttackAbility())
	if !movePermanentToZone(g, aura, zone.Graveyard) {
		t.Fatal("failed to remove Aura")
	}
	emitEvent(g, game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: creature.ObjectID,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("granted attack trigger fired after Aura left")
	}
}

func TestUmbraArmorReplacesLethalDamageStateBasedDestruction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := umbraArmorAura(g, game.Player2, creature, "Hyena Umbra")
	aura.Controller = game.Player3
	creature.MarkedDamage = 2
	creature.MarkedDeathtouchDamage = true

	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("creature left battlefield after Umbra armor replacement")
	}
	if creature.MarkedDamage != 0 || creature.MarkedDeathtouchDamage {
		t.Fatalf("marked damage = %d, deathtouch = %v; want cleared", creature.MarkedDamage, creature.MarkedDeathtouchDamage)
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Umbra Aura remained after replacing lethal destruction")
	}
	if !g.Players[aura.Owner].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("Umbra Aura was not destroyed into its owner's graveyard")
	}
}

func TestUmbraArmorStateBasedReplacementRepeatsAfterAuraLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreatureWithPowerToughness(g, game.Player1, 0, 0)
	aura := umbraArmorAura(g, game.Player1, creature, "Hyena Umbra", game.StaticAbility{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:          game.LayerPowerToughnessModify,
			Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
			PowerDelta:     1,
			ToughnessDelta: 1,
		}},
	})
	creature.MarkedDamage = 1

	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Umbra Aura remained after replacing lethal destruction")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("creature remained at zero toughness after the Umbra Aura left")
	}
}

func TestStateBasedDestructionPlansGrantedUmbraArmorBeforeSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	sourceAura := umbraArmorAura(g, game.Player1, first, "Umbra Granter", game.StaticAbility{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.ObjectControlledGroup(
				game.SourcePermanentReference(),
				game.Selection{SubtypesAny: []types.Sub{types.Aura}},
			),
			AddKeywords: []game.Keyword{game.UmbraArmor},
		}},
	})
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	grantedAura := addAuraPermanent(g, game.Player1)
	if !attachPermanent(g, grantedAura, second) {
		t.Fatal("plain Aura did not attach")
	}
	first.MarkedDamage = 2
	second.MarkedDamage = 2

	engine.applyStateBasedActions(g)

	for _, creature := range []*game.Permanent{first, second} {
		if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
			t.Fatalf("creature %d was destroyed despite pre-event Umbra armor", creature.ObjectID)
		}
		if creature.MarkedDamage != 0 {
			t.Fatalf("creature %d marked damage = %d, want zero", creature.ObjectID, creature.MarkedDamage)
		}
	}
	for _, aura := range []*game.Permanent{sourceAura, grantedAura} {
		if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
			t.Fatalf("Umbra Aura %d remained after replacing destruction", aura.ObjectID)
		}
	}
}

func TestUmbraArmorReplacesDestroySpellWithNoMarkedDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := umbraArmorAura(g, game.Player2, creature, "Boar Umbra")
	addEffectSpellToStack(g, game.Player3, game.Destroy{
		Object: game.TargetPermanentReference(0),
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("destroy spell destroyed Umbra-armored creature")
	}
	if creature.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want zero", creature.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Umbra Aura survived ordinary replacement destruction")
	}
}

func TestUmbraArmorAuraDestructionSharesMassDestroyBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saved := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := umbraArmorAura(g, game.Player1, saved, "Bear Umbra")
	destroyed := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addEffectSpellToStack(g, game.Player3, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	var auraBatch, creatureBatch game.ObjectID
	for _, event := range g.Events {
		if event.Kind != game.EventPermanentDied {
			continue
		}
		switch event.PermanentID {
		case aura.ObjectID:
			auraBatch = event.SimultaneousID
		case destroyed.ObjectID:
			creatureBatch = event.SimultaneousID
		default:
			continue
		}
	}
	if auraBatch == 0 || creatureBatch == 0 {
		t.Fatalf("death batches: Aura=%d creature=%d, want both nonzero", auraBatch, creatureBatch)
	}
	if auraBatch != creatureBatch {
		t.Fatalf("death batches: Aura=%d creature=%d, want one simultaneous batch", auraBatch, creatureBatch)
	}
	if _, ok := permanentByObjectID(g, saved.ObjectID); !ok {
		t.Fatal("Umbra-armored creature was destroyed by the mass destroy")
	}
}

func TestMassDestroyPlansAllReplacementsBeforeUmbraAuraLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saved := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	aura := umbraArmorAura(g, game.Player1, saved, "Protective Umbra", game.StaticAbility{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.ObjectControlledGroup(
				game.SourcePermanentReference(),
				game.Selection{RequiredTypes: []types.Card{types.Creature}},
			),
			AddKeywords: []game.Keyword{game.Indestructible},
		}},
	})
	protected := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addEffectSpellToStack(g, game.Player3, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Umbra Aura remained after replacing the attached creature's destruction")
	}
	if _, ok := permanentByObjectID(g, saved.ObjectID); !ok {
		t.Fatal("Umbra-armored creature was destroyed")
	}
	if _, ok := permanentByObjectID(g, protected.ObjectID); !ok {
		t.Fatal("creature indestructible in the pre-event state was destroyed after the Aura left")
	}
}

func TestMultipleUmbraArmorsUseAffectedCreatureControllerChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player3
	first := umbraArmorAura(g, game.Player1, creature, "First Umbra")
	second := umbraArmorAura(g, game.Player2, creature, "Second Umbra")
	agents := [game.NumPlayers]PlayerAgent{
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := &TurnLog{}
	engine.setReplacementChoiceContext(g, agents, log)
	defer g.ClearChoiceContext()

	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("creature was destroyed through Umbra armor")
	}
	if _, ok := permanentByObjectID(g, first.ObjectID); !ok {
		t.Fatal("unchosen Umbra Aura was destroyed")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("chosen Umbra Aura remained")
	}
	if len(g.ReplacementDecisions) != 1 ||
		g.ReplacementDecisions[0].Player != game.Player3 ||
		!slices.Equal(g.ReplacementDecisions[0].Selected, []int{1}) {
		t.Fatalf("replacement decisions = %+v, want Player3 choosing second Aura", g.ReplacementDecisions)
	}
}

func TestUmbraArmorOrdersWithShieldCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.Shield, 1)
	creature.RegenerationShields = 1
	aura := umbraArmorAura(g, game.Player2, creature, "Spider Umbra")
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{2}}},
	}
	engine.setReplacementChoiceContext(g, agents, &TurnLog{})
	defer g.ClearChoiceContext()

	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("first destroy was not replaced")
	}
	if creature.Counters.Get(counter.Shield) != 1 {
		t.Fatal("shield counter was consumed when Umbra armor was chosen")
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("chosen Umbra Aura remained")
	}
	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("second destroy was not replaced by shield counter")
	}
	if creature.Counters.Get(counter.Shield) != 0 {
		t.Fatal("shield counter remained after replacing second destroy")
	}
	if creature.RegenerationShields != 1 {
		t.Fatal("regeneration shield was consumed before it was chosen")
	}
	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("third destroy was not replaced by regeneration")
	}
	if creature.RegenerationShields != 0 {
		t.Fatal("regeneration shield remained after replacing third destroy")
	}
}

func TestUmbraArmorCanBeGrantedByAContinuousEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := addAuraPermanent(g, game.Player1)
	if !attachPermanent(g, aura, creature) {
		t.Fatal("plain Aura did not attach")
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Umbra Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{SubtypesAny: []types.Sub{types.Aura}},
				),
				AddKeywords: []game.Keyword{game.UmbraArmor},
			}},
		}},
	}})

	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("creature was destroyed through continuously granted Umbra armor")
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("Aura with granted Umbra armor remained")
	}
}

func TestUmbraArmorDoesNotReplaceSacrificeExileOrZeroToughness(t *testing.T) {
	t.Run("sacrifice", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		umbraArmorAura(g, game.Player1, creature, "Sacrifice Umbra")
		addEffectSpellToStack(g, game.Player1, game.Sacrifice{
			Object: game.TargetPermanentReference(0),
		}, []game.Target{game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
			t.Fatal("sacrificed creature remained")
		}
	})
	t.Run("exile", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		umbraArmorAura(g, game.Player1, creature, "Exile Umbra")
		addEffectSpellToStack(g, game.Player1, game.Exile{
			Object: game.TargetPermanentReference(0),
		}, []game.Target{game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
			t.Fatal("exiled creature remained")
		}
	})
	t.Run("zero toughness", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		zero := opt.Val(game.PT{Value: 0})
		creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Zero Toughness", Types: []types.Card{types.Creature}, Power: zero, Toughness: zero,
		}})
		aura := umbraArmorAura(g, game.Player1, creature, "Zero Umbra")
		engine.applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
			t.Fatal("zero-toughness creature remained")
		}
		if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
			t.Fatal("Aura remained attached after zero-toughness creature left")
		}
	})
}

func TestUmbraArmorDoesNotApplyAfterAuraRemovalOrToIndestructibleCreature(t *testing.T) {
	t.Run("Aura removed", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		aura := umbraArmorAura(g, game.Player1, creature, "Removed Umbra")
		movePermanentToZone(g, aura, zone.Graveyard)
		if _, destroyed := destroyPermanent(g, creature.ObjectID); !destroyed {
			t.Fatal("destroy was replaced after Aura left")
		}
	})
	t.Run("indestructible creature", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
		aura := umbraArmorAura(g, game.Player1, creature, "Indestructible Creature Umbra")
		creature.MarkedDamage = 10
		NewEngine(nil).applyStateBasedActions(g)
		if _, ok := permanentByObjectID(g, aura.ObjectID); !ok {
			t.Fatal("Umbra armor was used for an indestructible creature")
		}
		if creature.MarkedDamage != 10 {
			t.Fatalf("marked damage = %d, want retained on indestructible creature", creature.MarkedDamage)
		}
	})
}

func TestUmbraArmorStillReplacesWhenAuraDestructionIsPrevented(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Game, *game.Permanent)
		check func(*testing.T, *game.Permanent)
	}{
		{
			name: "indestructible Aura",
			setup: func(g *game.Game, aura *game.Permanent) {
				card, ok := g.GetCardInstance(aura.CardInstanceID)
				if !ok {
					panic("test Aura card missing")
				}
				card.Def.StaticAbilities = append(card.Def.StaticAbilities, game.IndestructibleStaticBody)
			},
		},
		{
			name: "shielded Aura",
			setup: func(_ *game.Game, aura *game.Permanent) {
				aura.Counters.Add(counter.Shield, 1)
			},
			check: func(t *testing.T, aura *game.Permanent) {
				if aura.Counters.Get(counter.Shield) != 0 {
					t.Fatal("Aura shield counter was not consumed")
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			aura := umbraArmorAura(g, game.Player1, creature, test.name)
			test.setup(g, aura)
			creature.MarkedDamage = 1
			if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
				t.Fatal("creature was destroyed")
			}
			if _, ok := permanentByObjectID(g, aura.ObjectID); !ok {
				t.Fatal("Aura left despite its destruction being replaced/prevented")
			}
			if creature.MarkedDamage != 0 {
				t.Fatalf("creature marked damage = %d, want cleared", creature.MarkedDamage)
			}
			if test.check != nil {
				test.check(t, aura)
			}
		})
	}
}

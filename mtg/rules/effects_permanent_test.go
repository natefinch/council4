package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDestroyEffectMovesPermanentToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("destroyed permanent remained on battlefield")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("destroyed card was not in owner's graveyard")
	}
}

func TestDestroyEffectPreventRegenerationBypassesShieldButHonorsIndestructibleAndCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	regenerated := addCreaturePermanent(g, game.Player2)
	regenerated.RegenerationShields = 1
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Object:              game.TargetPermanentReference(0),
		PreventRegeneration: true,
	}, []game.Target{game.PermanentTarget(regenerated.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, regenerated.ObjectID); ok {
		t.Fatal("regeneration shield wrongly saved a permanent from a can't-be-regenerated destroy")
	}
	if !g.Players[game.Player2].Graveyard.Contains(regenerated.CardInstanceID) {
		t.Fatal("destroyed card was not in owner's graveyard")
	}
}

func TestDestroyEffectPreventRegenerationStillHonorsIndestructibleAndShieldCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	indestructible := addCombatCreaturePermanent(g, game.Player2, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Object:              game.TargetPermanentReference(0),
		PreventRegeneration: true,
	}, []game.Target{game.PermanentTarget(indestructible.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, indestructible.ObjectID); !ok {
		t.Fatal("indestructible creature was destroyed by a can't-be-regenerated destroy")
	}

	shielded := addCombatCreaturePermanent(g, game.Player2)
	shielded.Counters.Add(counter.Shield, 1)
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Object:              game.TargetPermanentReference(0),
		PreventRegeneration: true,
	}, []game.Target{game.PermanentTarget(shielded.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, shielded.ObjectID); !ok || shielded.Counters.Get(counter.Shield) != 0 {
		t.Fatal("shield counter did not replace a can't-be-regenerated destroy")
	}
}

func TestMassDestroyPreventRegenerationIgnoresRegenerationShields(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	regenerated := addCombatCreaturePermanent(g, game.Player2)
	regenerated.RegenerationShields = 1
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		PreventRegeneration: true,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, regenerated.ObjectID); ok {
		t.Fatal("regeneration shield wrongly saved a creature from a mass can't-be-regenerated destroy")
	}
}

func TestExileAndBounceEffectsMovePermanentsToOwnerZones(t *testing.T) {
	tests := []struct {
		name      string
		primitive game.Primitive
	}{
		{name: "exile", primitive: game.Exile{Object: game.TargetPermanentReference(0)}},
		{name: "bounce", primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreaturePermanent(g, game.Player2)
			addEffectSpellToStack(g, game.Player1, tt.primitive, []game.Target{game.PermanentTarget(target.ObjectID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("moved permanent remained on battlefield")
			}
			var z *zone.Zone
			switch tt.name {
			case "exile":
				z = &g.Players[game.Player2].Exile
			case "bounce":
				z = &g.Players[game.Player2].Hand
			default:
			}
			if z == nil || !z.Contains(target.CardInstanceID) {
				t.Fatalf("card was not moved to expected zone for %s", tt.name)
			}
		})
	}
}

func TestSacrificeEffectMovesControllerPermanentThroughGraveyardIgnoringIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Sacrifice{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("sacrificed permanent remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("sacrificed permanent did not move to graveyard")
	}
	assertEvent(t, g.Events, game.EventPermanentSacrificed, func(event game.Event) bool {
		return event.Controller == game.Player1 &&
			event.Player == game.Player1 &&
			event.PermanentID == target.ObjectID &&
			event.CardID == target.CardInstanceID
	})
}

func TestTapAndUntapEffectsChangeTappedState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Tap{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped {
		t.Fatal("tap effect did not tap permanent")
	}

	addEffectSpellToStack(g, game.Player1, game.Untap{Object: game.TargetPermanentReference(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.Tapped {
		t.Fatal("untap effect did not untap permanent")
	}
}

func TestMassTapTapsOnlyGroupAndUntapSequenceUntapsControllers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCreaturePermanent(g, game.Player1)
	theirs := addCreaturePermanent(g, game.Player2)
	// Start both creatures tapped so the untap clause has visible work to do.
	mine.Tapped = true
	theirs.Tapped = true

	// "Tap all creatures your opponents control and untap all creatures you
	// control." resolves as a two-instruction sequence; the order must hold.
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Tap{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerOpponent,
			}),
		}},
		{Primitive: game.Untap{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}),
		}},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !theirs.Tapped {
		t.Fatal("opponent's creature was not tapped by the mass tap clause")
	}
	if mine.Tapped {
		t.Fatal("controller's creature was not untapped by the mass untap clause")
	}
}

func TestDamageToPermanentEffectCanCauseLethalSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addEffectSpellToStack(g, game.Player1, game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 1 {
		t.Fatalf("deaths = %d, want 1", len(deaths))
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("lethally damaged permanent remained on battlefield")
	}
}

func TestMassDestroyCreaturesUsesSnapshotAndRespectsIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	creature2 := addCreaturePermanent(g, game.Player2)
	phased := addCreaturePermanent(g, game.Player2)
	phased.PhasedOut = true
	indestructible := addCombatCreaturePermanent(g, game.Player3, game.Indestructible)
	shielded := addCombatCreaturePermanent(g, game.Player3)
	shielded.Counters.Add(counter.Shield, 1)
	regenerated := addCombatCreaturePermanent(g, game.Player4)
	regenerated.RegenerationShields = 1
	artifact := addCombatPermanent(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, creature1.ObjectID); ok {
		t.Fatal("first creature survived mass destroy")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("second creature survived mass destroy")
	}
	if _, ok := permanentByObjectID(g, phased.ObjectID); !ok || !phased.PhasedOut {
		t.Fatal("mass destroy affected phased-out creature")
	}
	if _, ok := permanentByObjectID(g, indestructible.ObjectID); !ok {
		t.Fatal("indestructible creature did not survive mass destroy")
	}
	if _, ok := permanentByObjectID(g, shielded.ObjectID); !ok || shielded.Counters.Get(counter.Shield) != 0 {
		t.Fatal("shield counter did not replace mass destroy")
	}
	if _, ok := permanentByObjectID(g, regenerated.ObjectID); !ok || regenerated.RegenerationShields != 0 {
		t.Fatal("regeneration shield did not replace mass destroy")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); !ok {
		t.Fatal("noncreature artifact did not survive mass destroy")
	}
}

func TestMassDestroyDeathsShareOneOrMoreTriggerBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                 game.EventPermanentDied,
		RequirePermanentTypes: []types.Card{types.Creature},
		OneOrMore:             true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-or-more death trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one trigger for simultaneous mass destroy", got)
	}
}

func TestMassDestroyNonlandPermanentsLeavesLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}},
	})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Aura",
		Types: []types.Card{types.Enchantment}},
	})
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, land.ObjectID); !ok {
		t.Fatal("land did not survive nonland permanent wipe")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("artifact survived nonland permanent wipe")
	}
	if _, ok := permanentByObjectID(g, enchantment.ObjectID); ok {
		t.Fatal("enchantment survived nonland permanent wipe")
	}
}

func TestMassDestroySubtypeLeavesOtherPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	island := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Island}},
	})
	forest := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest}},
	})
	creature := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Island}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, island.ObjectID); ok {
		t.Fatal("Island survived Destroy all Islands")
	}
	if _, ok := permanentByObjectID(g, forest.ObjectID); !ok {
		t.Fatal("Forest did not survive Destroy all Islands")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("creature did not survive Destroy all Islands")
	}
}

func TestMassDestroyNonbasicLandsLeavesBasics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	basic := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}, Subtypes: []types.Sub{types.Island}},
	})
	nonbasic := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wasteland",
		Types: []types.Card{types.Land}},
	})
	creature := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes:     []types.Card{types.Land},
			ExcludedSupertype: types.Basic,
		}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, basic.ObjectID); !ok {
		t.Fatal("basic land did not survive Destroy all nonbasic lands")
	}
	if _, ok := permanentByObjectID(g, nonbasic.ObjectID); ok {
		t.Fatal("nonbasic land survived Destroy all nonbasic lands")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("creature did not survive Destroy all nonbasic lands")
	}
}

func TestDualTargetBounceReturnsBothTargetsToOwnersHands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCreaturePermanent(g, game.Player1)
	theirs := addCreaturePermanent(g, game.Player2)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.Bounce{Object: game.TargetPermanentReference(1)}},
	}, []game.Target{
		game.PermanentTarget(mine.ObjectID),
		game.PermanentTarget(theirs.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, mine.ObjectID); ok {
		t.Fatal("first dual-bounce target remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, theirs.ObjectID); ok {
		t.Fatal("second dual-bounce target remained on battlefield")
	}
	if !g.Players[game.Player1].Hand.Contains(mine.CardInstanceID) {
		t.Fatal("first dual-bounce target was not returned to its owner's hand")
	}
	if !g.Players[game.Player2].Hand.Contains(theirs.CardInstanceID) {
		t.Fatal("second dual-bounce target was not returned to its owner's hand")
	}
}

func TestMassBounceCreaturesReturnsOnlyCreaturesToOwnersHands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	creature2 := addCreaturePermanent(g, game.Player2)
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}},
	})
	addEffectSpellToStack(g, game.Player1, game.Bounce{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, creature1.ObjectID); ok {
		t.Fatal("Player1 creature remained on battlefield after mass bounce")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("Player2 creature remained on battlefield after mass bounce")
	}
	if _, ok := permanentByObjectID(g, land.ObjectID); !ok {
		t.Fatal("land was bounced by a creatures-only mass bounce")
	}
	if !g.Players[game.Player1].Hand.Contains(creature1.CardInstanceID) {
		t.Fatal("Player1 creature was not returned to its owner's hand")
	}
	if !g.Players[game.Player2].Hand.Contains(creature2.CardInstanceID) {
		t.Fatal("Player2 creature was not returned to its owner's hand")
	}
}

func TestMassBounceYouControlReturnsOnlyControllersPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCreaturePermanent(g, game.Player1)
	theirs := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Bounce{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, mine.ObjectID); ok {
		t.Fatal("controller's creature remained on battlefield after self-control mass bounce")
	}
	if _, ok := permanentByObjectID(g, theirs.ObjectID); !ok {
		t.Fatal("opponent's creature was bounced by a 'you control' mass bounce")
	}
	if !g.Players[game.Player1].Hand.Contains(mine.CardInstanceID) {
		t.Fatal("controller's creature was not returned to its owner's hand")
	}
}

func TestControlledChoiceBounceAutoChoosesWhenEligibleCountLEAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCreaturePermanent(g, game.Player1)
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
	})
	theirs := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Bounce{
		ControlledChoice: true,
		Amount:           game.Fixed(1),
		Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
	}, nil)

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no prompt when eligible count <= amount", log.Choices)
	}
	if _, ok := permanentByObjectID(g, mine.ObjectID); ok {
		t.Fatal("controller's only eligible creature was not bounced")
	}
	if !g.Players[game.Player1].Hand.Contains(mine.CardInstanceID) {
		t.Fatal("bounced creature was not returned to its owner's hand")
	}
	if _, ok := permanentByObjectID(g, land.ObjectID); !ok {
		t.Fatal("non-matching land was bounced")
	}
	if _, ok := permanentByObjectID(g, theirs.ObjectID); !ok {
		t.Fatal("opponent's creature was bounced by a 'you control' choice bounce")
	}
}

func TestControlledChoiceBounceAsksControllerToChooseWhenExcessEligible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	creature2 := addCreaturePermanent(g, game.Player1)
	addEffectSpellToStack(g, game.Player1, game.Bounce{
		ControlledChoice: true,
		Amount:           game.Fixed(1),
		Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceResolution {
		t.Fatalf("choices = %+v, want one ChoiceResolution prompt", log.Choices)
	}
	if len(log.Choices[0].Request.Options) != 2 {
		t.Fatalf("options = %d, want 2 (one per eligible permanent)", len(log.Choices[0].Request.Options))
	}
	// Agent chose index 1 (creature2).
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("chosen creature remained on battlefield")
	}
	if !g.Players[game.Player1].Hand.Contains(creature2.CardInstanceID) {
		t.Fatal("chosen creature was not returned to its owner's hand")
	}
	if _, ok := permanentByObjectID(g, creature1.ObjectID); !ok {
		t.Fatal("unchosen creature was bounced")
	}
}

func TestControlledChoiceBounceAnotherExcludesSourceAndUsesOwnerHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCreaturePermanent(g, game.Player1)
	// A creature controlled by Player1 but owned by Player2 must still go to
	// its OWNER's hand when bounced.
	borrowed := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Borrowed Beast",
		Types: []types.Card{types.Creature}},
	})
	borrowed.Owner = game.Player2
	g.CardInstances[borrowed.CardInstanceID].Owner = game.Player2

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		SourceID:   source.ObjectID,
		Controller: game.Player1,
	}
	log := TurnLog{}
	resolveInstruction(engine, g, obj, game.Bounce{
		ControlledChoice: true,
		Amount:           game.Fixed(1),
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
			ExcludeSource: true,
		}),
	}, &log)

	if _, ok := permanentByObjectID(g, source.ObjectID); !ok {
		t.Fatal("source permanent was bounced despite ExcludeSource")
	}
	if _, ok := permanentByObjectID(g, borrowed.ObjectID); ok {
		t.Fatal("the other controlled creature was not bounced")
	}
	if !g.Players[game.Player2].Hand.Contains(borrowed.CardInstanceID) {
		t.Fatal("bounced creature was not returned to its OWNER's hand")
	}
}

func TestSelectorOtherCreaturesDefendingPlayerControlsUsesTriggerRecipientController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	damagedBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	otherDefenderCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	defenderArtifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Defender Relic",
		Types: []types.Card{types.Artifact}},
	})
	attackerCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventDamageDealt,
			PermanentID: damagedBlocker.ObjectID,
		},
	}

	resolveInstruction(engine, g, obj, game.Destroy{
		Group: game.ObjectControlledGroupExcluding(
			game.EventPermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.EventPermanentReference(),
		),
	}, &TurnLog{})

	if _, ok := permanentByObjectID(g, otherDefenderCreature.ObjectID); ok {
		t.Fatal("other defender creature survived selector destroy")
	}
	for _, permanent := range []*game.Permanent{damagedBlocker, defenderArtifact, attackerCreature} {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
			t.Fatalf("selector destroyed permanent %v unexpectedly", permanent.ObjectID)
		}
	}
}

func TestMassDamageDeathsAreLoggedTogetherBySBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	artifact := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Fixed(3),
			Recipient: game.GroupDamageRecipient(
				game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 2 {
		t.Fatalf("deaths = %d, want 2", len(deaths))
	}
	if _, ok := permanentByObjectID(g, creature1.ObjectID); ok {
		t.Fatal("first damaged creature survived SBA")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("second damaged creature survived SBA")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); !ok {
		t.Fatal("noncreature artifact was affected by creature mass damage")
	}
}

func TestTemporaryPTModifierChangesCombatDamageAndLethalThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(3),
		ToughnessDelta: game.Fixed(3),
		Duration:       game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: creature.ObjectID},
		},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if blocker.MarkedDamage != 5 {
		t.Fatalf("blocker marked damage = %d, want 5", blocker.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived pumped combat damage")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("pumped creature died despite increased toughness")
	}
	if len(deaths) != 1 || deaths[0].Permanent != blocker.ObjectID {
		t.Fatalf("deaths = %+v, want blocker death only", deaths)
	}
}

func TestTemporaryPTModifiersStackDeterministically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, primitive := range []game.Primitive{
		game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(1), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn},
		game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(-2), ToughnessDelta: game.Fixed(-1), Duration: game.DurationUntilEndOfTurn},
	} {
		addEffectSpellToStack(g, game.Player1, primitive, []game.Target{game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 3 {
		t.Fatalf("effective toughness = %d ok=%v, want 3 true", got, ok)
	}
}

func TestAddCounterEffectAddsCountersToTargetPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Object:      game.TargetPermanentReference(0),
		Amount:      game.Fixed(3),
		CounterKind: counter.PlusOnePlusOne,
	}, []game.Target{game.PermanentTarget(artifact.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := artifact.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
}

func TestAddCounterEffectFansOutAcrossChosenTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureDef := &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})},
	}
	first := addCombatPermanent(g, game.Player1, creatureDef)
	second := addCombatPermanent(g, game.Player1, creatureDef)

	instructions := []game.Instruction{
		{Primitive: game.AddCounter{Object: game.TargetPermanentReference(0), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne}},
		{Primitive: game.AddCounter{Object: game.TargetPermanentReference(1), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne}},
	}
	addInstructionSpellToStackForController(g, game.Player1, instructions, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := first.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("first +1/+1 counters = %d, want 1", got)
	}
	if got := second.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("second +1/+1 counters = %d, want 1", got)
	}
}

func TestAddCounterEffectSkipsUnchosenOptionalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureDef := &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})},
	}
	chosen := addCombatPermanent(g, game.Player1, creatureDef)
	untargeted := addCombatPermanent(g, game.Player1, creatureDef)

	// An "up to two" placement may choose a single target; the second
	// per-target instruction resolves against an unchosen index and must no-op
	// rather than panic.
	instructions := []game.Instruction{
		{Primitive: game.AddCounter{Object: game.TargetPermanentReference(0), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne}},
		{Primitive: game.AddCounter{Object: game.TargetPermanentReference(1), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne}},
	}
	addInstructionSpellToStackForController(g, game.Player1, instructions, []game.Target{
		game.PermanentTarget(chosen.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := chosen.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("chosen +1/+1 counters = %d, want 1", got)
	}
	if got := untargeted.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("untargeted +1/+1 counters = %d, want 0", got)
	}
}

func TestAddCounterEffectAddsCountersToGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureDef := &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})},
	}
	mine1 := addCombatPermanent(g, game.Player1, creatureDef)
	mine2 := addCombatPermanent(g, game.Player1, creatureDef)
	theirs := addCombatPermanent(g, game.Player2, creatureDef)

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
		Amount:      game.Fixed(1),
		CounterKind: counter.PlusOnePlusOne,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := mine1.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("mine1 +1/+1 counters = %d, want 1", got)
	}
	if got := mine2.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("mine2 +1/+1 counters = %d, want 1", got)
	}
	if got := theirs.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent creature +1/+1 counters = %d, want 0 (you-control filter)", got)
	}
}

// A keyword counter placed on a filtered group grants the keyword to each
// matching permanent, exercising the group counter placement that lowers from
// the wording "Put a deathtouch counter on each creature you control".
func TestAddKeywordCounterEffectAddsKeywordToGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureDef := &game.CardDef{CardFace: game.CardFace{Name: "Bear",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})},
	}
	mine := addCombatPermanent(g, game.Player1, creatureDef)
	theirs := addCombatPermanent(g, game.Player2, creatureDef)

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
		Amount:      game.Fixed(1),
		CounterKind: counter.Deathtouch,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := mine.Counters.Get(counter.Deathtouch); got != 1 {
		t.Fatalf("mine deathtouch counters = %d, want 1", got)
	}
	if !hasKeyword(g, mine, game.Deathtouch) {
		t.Fatal("controlled creature should gain deathtouch from its deathtouch counter")
	}
	if got := theirs.Counters.Get(counter.Deathtouch); got != 0 {
		t.Fatalf("opponent creature deathtouch counters = %d, want 0 (you-control filter)", got)
	}
	if hasKeyword(g, theirs, game.Deathtouch) {
		t.Fatal("opponent creature should not gain deathtouch")
	}
}

func TestMoveCountersEffectMovesCountersBetweenTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source Relic",
		Types: []types.Card{types.Artifact}},
	})
	destination := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Destination Relic",
		Types: []types.Card{types.Artifact}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	source.Counters.Add(counter.Charge, 1)
	addEffectSpellToStack(g, game.Player1, game.MoveCounters{
		Object: game.TargetPermanentReference(1),
		Source: game.CounterSourceSpec{
			Kind:   game.CounterSourceTarget,
			Object: game.TargetPermanentReference(0),
		},
	}, []game.Target{
		game.PermanentTarget(source.ObjectID),
		game.PermanentTarget(destination.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source +1/+1 counters = %d, want 0", got)
	}
	if got := source.Counters.Get(counter.Charge); got != 0 {
		t.Fatalf("source charge counters = %d, want 0", got)
	}
	if got := destination.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("destination +1/+1 counters = %d, want 2", got)
	}
	if got := destination.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("destination charge counters = %d, want 1", got)
	}
}

func TestConditionalContinuousEffectAnimatesNonCreatureArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	zero := game.PT{Value: 0}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:       game.LayerType,
					AddTypes:    []types.Card{types.Creature},
					AddSubtypes: []types.Sub{types.Robot},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					SetPower:     opt.Val(zero),
					SetToughness: opt.Val(zero),
				},
			},
		},
		Condition: opt.Val(game.EffectCondition{Text: "it isn't a creature", Object: game.TargetPermanentReference(0), PermanentType: opt.Val(types.Creature), Negate: true}),
	}}, []game.Target{game.PermanentTarget(artifact.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !permanentHasType(g, artifact, types.Creature) {
		t.Fatal("noncreature artifact did not become a creature")
	}
	if !permanentHasSubtype(g, artifact, types.Robot) {
		t.Fatal("noncreature artifact did not gain Robot subtype")
	}
	if got := effectivePower(g, artifact); got != 0 {
		t.Fatalf("effective power = %d, want 0", got)
	}
	if got, ok := effectiveToughness(g, artifact); !ok || got != 0 {
		t.Fatalf("effective toughness = %d ok=%v, want 0 true", got, ok)
	}
}

func TestConditionalContinuousEffectSkipsCreatureArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	two := game.PT{Value: 2}
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(two),
		Toughness: opt.Val(two)},
	})
	zero := game.PT{Value: 0}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:       game.LayerType,
					AddTypes:    []types.Card{types.Creature},
					AddSubtypes: []types.Sub{types.Robot},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					SetPower:     opt.Val(zero),
					SetToughness: opt.Val(zero),
				},
			},
		},
		Condition: opt.Val(game.EffectCondition{Text: "it isn't a creature", Object: game.TargetPermanentReference(0), PermanentType: opt.Val(types.Creature), Negate: true}),
	}}, []game.Target{game.PermanentTarget(artifactCreature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentHasSubtype(g, artifactCreature, types.Robot) {
		t.Fatal("creature artifact incorrectly gained Robot subtype")
	}
	if got := effectivePower(g, artifactCreature); got != 2 {
		t.Fatalf("effective power = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, artifactCreature); !ok || got != 2 {
		t.Fatalf("effective toughness = %d ok=%v, want 2 true", got, ok)
	}
}

func TestTemporaryPTModifierExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.TemporaryPowerModifier = 3
	creature.TemporaryToughnessModifier = 3

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.TemporaryPowerModifier != 0 || creature.TemporaryToughnessModifier != 0 {
		t.Fatalf("temporary modifiers = +%d/+%d, want 0/0", creature.TemporaryPowerModifier, creature.TemporaryToughnessModifier)
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want 2", got)
	}
}

func TestCreateTokenEffectCreatesTokenPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}
	addEffectSpellToStack(g, game.Player1, game.CreateToken{Amount: game.Fixed(2), Source: game.TokenDef(token)}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	for _, permanent := range tokens {
		if permanent.TokenDef != token {
			t.Fatalf("token def = %p, want %p", permanent.TokenDef, token)
		}
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 {
			t.Fatalf("token owner/controller = %v/%v, want %v", permanent.Owner, permanent.Controller, game.Player1)
		}
		if !permanent.SummoningSick {
			t.Fatal("token did not enter summoning sick")
		}
	}
}

func TestCreateTokenEffectDynamicCountMatchesControllerLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 3
	engine := NewEngine(nil)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Pegasus",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}
	addEffectSpellToStack(g, game.Player1, game.CreateToken{
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountControllerLife, Multiplier: 1}),
		Source: game.TokenDef(token),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	tokens := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens++
		}
	}
	if tokens != 3 {
		t.Fatalf("tokens = %d, want 3 (controller life total)", tokens)
	}
}

func TestCreateTokenEffectEntryTappedEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Zombie Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	}
	addEffectSpellToStack(g, game.Player1, game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(token), EntryTapped: true}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	found := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			found++
			if !permanent.Tapped {
				t.Fatal("token Tapped = false, want true")
			}
		}
	}
	if found != 1 {
		t.Fatalf("tokens = %d, want 1", found)
	}
}

func TestCreateTokenPermanentAppliesReplacementAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Modified Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedReplacement("This token enters tapped."),
			game.EntersWithCountersReplacement("This token enters with a +1/+1 counter.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
		}},
	}

	permanent, ok := createTokenPermanent(g, game.Player1, token)

	if !ok {
		t.Fatal("token was not created")
	}
	if !permanent.Tapped {
		t.Fatal("token did not enter tapped")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
}

func TestCreateTokenCanCopySourceCardWithModifications(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Fanatic Source",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			ManaCost:  opt.Val(cost.Mana{cost.G}),
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4})},
		},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     sourceID,
		SourceCardID: sourceID,
		Controller:   game.Player1,
		AbilityIndex: 0,
	})
	g.CardInstances[sourceID].Def.ActivatedAbilities = []game.ActivatedAbility{
		game.EternalizeActivatedBody(cost.Mana{cost.O(0)}, types.Snake, types.Druid),
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			token = permanent
			break
		}
	}
	if token == nil || token.TokenDef == nil {
		t.Fatal("copy token was not created")
	}
	if token.TokenDef.ManaCost.Exists || token.TokenDef.ManaValue() != 0 {
		t.Fatalf("token mana cost/value = %+v/%d, want no cost and mana value 0", token.TokenDef.ManaCost, token.TokenDef.ManaValue())
	}
	if got := token.TokenDef.Subtypes; !slices.Equal(got, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token subtypes = %+v, want Zombie Snake Druid", got)
	}
	if got := token.TokenDef.Colors; !slices.Equal(got, []color.Color{color.Black}) {
		t.Fatalf("token colors = %+v, want black", got)
	}
	if got := effectivePower(g, token); got != 4 {
		t.Fatalf("token power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, token); !ok || got != 4 {
		t.Fatalf("token toughness = %d ok=%v, want 4 true", got, ok)
	}
}

func TestCopyCardDefPreservesCategorizedAbilitiesWithoutDuplication(t *testing.T) {
	source := &game.CardDef{CardFace: game.CardFace{
		Name: "Categorized Source",
		StaticAbilities: []game.StaticAbility{{
			Text:             "Flying",
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flying}},
		}},
	}}

	copied := copyCardDef(source)

	if copied.AbilityCount() != 1 {
		t.Fatalf("copied abilities = %d, want one categorized ability without duplication", copied.AbilityCount())
	}
	if !copied.HasKeyword(game.Flying) {
		t.Fatal("copied categorized keyword ability was not preserved")
	}
}

func TestClearCardFaceAbilitiesClearsCategorizedAbilities(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		StaticAbilities: []game.StaticAbility{{
			Text:             "Flying",
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flying}},
		}},
	}}
	face := card.CardFace

	clearCardFaceAbilities(&face)

	if face.AbilityCount() != 0 {
		t.Fatalf("abilities = %d, want categorized abilities cleared", face.AbilityCount())
	}
	face.StaticAbilities = []game.StaticAbility{game.FlyingStaticBody}
	if !face.HasKeyword(game.Flying) {
		t.Fatal("ability cache remained stale after clearing and adding a categorized ability")
	}
}

func TestTokenCanBlockTakeCombatDamageAndDie(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 2}
	token, ok := createTokenPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Bear Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt)},
	})
	if !ok {
		t.Fatal("token was not created")
	}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: token.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("lethally damaged token remained on battlefield")
	}
	if g.Players[game.Player2].Graveyard.Contains(token.ObjectID) {
		t.Fatal("dead token did not cease to exist from graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != token.ObjectID || deaths[0].TokenName != "Bear Token" {
		t.Fatalf("death logs = %+v, want readable token death", deaths)
	}
}

func TestMultiTargetPumpAppliesToEachChosenTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(2), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn}},
		{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(1), PowerDelta: game.Fixed(2), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn}},
	}, []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, creature := range []*game.Permanent{first, second} {
		if got := effectivePower(g, creature); got != 4 {
			t.Fatalf("effective power = %d, want 4", got)
		}
		if got, ok := effectiveToughness(g, creature); !ok || got != 4 {
			t.Fatalf("effective toughness = %d ok=%v, want 4 true", got, ok)
		}
	}
}

func TestMultiTargetPumpNoOpsOnDeclinedSlot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	untouched := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	// An "up to two" spell where only one target was chosen leaves the second
	// slot's target reference unresolved; that ModifyPT must no-op.
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(2), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn}},
		{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(1), PowerDelta: game.Fixed(2), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn}},
	}, []game.Target{game.PermanentTarget(chosen.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, chosen); got != 4 {
		t.Fatalf("chosen effective power = %d, want 4", got)
	}
	if got := effectivePower(g, untouched); got != 2 {
		t.Fatalf("untouched effective power = %d, want 2", got)
	}
}

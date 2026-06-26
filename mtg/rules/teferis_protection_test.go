package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestPlayerProtectionRuleEffectsAndDuration(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{
			{Kind: game.RuleEffectLifeTotalCantChange, AffectedPlayer: game.PlayerYou},
			{
				Kind:           game.RuleEffectPlayerProtection,
				AffectedPlayer: game.PlayerYou,
				Protection:     game.ProtectionKeyword{Everything: true},
			},
		},
		Duration: game.DurationUntilYourNextTurn,
	}, nil)

	startLife := g.Players[game.Player1].Life
	startEvents := len(g.Events)
	if gainLife(g, game.Player1, 5) != 0 || loseLife(g, game.Player1, 5) != 0 {
		t.Fatal("immutable life total reported a life change")
	}
	if g.Players[game.Player1].Life != startLife || len(g.Events) != startEvents {
		t.Fatalf("blocked life change mutated life/events: life=%d events=%d", g.Players[game.Player1].Life, len(g.Events))
	}
	manaCost := cost.Mana{cost.PhyrexianMana(mana.G)}
	if payTestGenericCostWithPreferences(g, game.Player1, &manaCost, &payment.Preferences{
		PhyrexianLifeChoices: []bool{true},
	}) {
		t.Fatal("life payment succeeded while life total could not change")
	}
	emptyCost := cost.Mana{}
	lifeCostRequest := payment.GenericRequest{
		PlayerID: game.Player1,
		Cost:     &emptyCost,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalPayLife,
			Amount: 2,
		}},
	}
	if paymentOrch.canPayGenericCost(g, lifeCostRequest) ||
		paymentOrch.payGenericCost(g, lifeCostRequest) {
		t.Fatal("additional life cost remained legal while life total could not change")
	}
	if !targetProtectedFromSource(g, game.Player2, nil, 0, game.PlayerTarget(game.Player1)) {
		t.Fatal("protected player remained targetable")
	}
	if got := applyDamageModifications(g, damageEvent{
		controller: game.Player2,
		player:     game.Player1,
		amount:     7,
	}); got != 0 {
		t.Fatalf("damage after protection = %d, want 0", got)
	}
	for _, effect := range g.RuleEffects {
		if effect.ExpiresFor != game.Player1 {
			t.Fatalf("ExpiresFor = %v, want Player1", effect.ExpiresFor)
		}
	}

	g.Turn.ActivePlayer = game.Player2
	g.Turn.TurnNumber++
	expireTurnStartDurations(g)
	if len(g.RuleEffects) != 2 {
		t.Fatalf("effects expired on opponent turn: %+v", g.RuleEffects)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber++
	expireTurnStartDurations(g)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("effects did not expire at controller's next turn: %+v", g.RuleEffects)
	}
	if loseLife(g, game.Player1, 1) != 1 || g.Players[game.Player1].Life != startLife-1 {
		t.Fatal("life total remained immutable after duration expired")
	}
}

func TestGroupPhaseOutPreservesAttachmentsTokensAndPhaseInTiming(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := makeAuraAttachedTo(g, game.Player2, creature, "Opponent Aura")
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Controller: game.Player1,
		Owner:      game.Player1,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:  "Token",
			Types: []types.Card{types.Creature},
		}},
	}
	g.Battlefield = append(g.Battlefield, token)
	creature.Tapped = true
	token.Tapped = true
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: creature.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	startEvents := len(g.Events)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.PhaseOut{
		Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
	}, nil)

	for _, permanent := range []*game.Permanent{creature, aura, token} {
		if !permanent.PhasedOut || !permanent.PhaseInScheduled || permanent.PhasedOutFor != game.Player1 {
			t.Fatalf("permanent %d phase state = %+v, want phased out for Player1", permanent.ObjectID, permanent)
		}
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != creature.ObjectID ||
		len(creature.Attachments) != 1 || creature.Attachments[0] != aura.ObjectID {
		t.Fatal("phasing changed attachment links")
	}
	if len(g.Battlefield) != 3 {
		t.Fatalf("battlefield length = %d, want tokens and cards preserved", len(g.Battlefield))
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatal("phased-out creature remained in combat")
	}
	phasedEvents := 0
	for _, event := range g.Events[startEvents:] {
		if event.Kind == game.EventZoneChanged {
			t.Fatal("phasing emitted a zone-change event")
		}
		if event.Kind == game.EventPermanentPhasedOut {
			phasedEvents++
		}
	}
	if phasedEvents != 3 {
		t.Fatalf("phase-out events = %d, want 3", phasedEvents)
	}

	creature.Controller = game.Player2
	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, permanent := range []*game.Permanent{creature, aura, token} {
		if permanent.PhasedOut {
			t.Fatalf("permanent %d did not phase in on captured player's untap", permanent.ObjectID)
		}
	}
	if token.Tapped {
		t.Fatal("phased-in permanent controlled by active player did not untap")
	}
	if !creature.Tapped {
		t.Fatal("phased-in permanent controlled by another player untapped")
	}
}

func TestAttachmentStateBasedActionsIgnorePhasedOutAura(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := makeAuraAttachedTo(g, game.Player2, creature, "Opponent Aura")

	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.PhaseOut{
		Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
	}, nil)
	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, aura.ObjectID); !ok {
		t.Fatal("state-based actions moved phased-out Aura from the battlefield")
	}
	if !aura.PhasedOut || !creature.PhasedOut {
		t.Fatal("Aura and enchanted permanent should remain phased out")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != creature.ObjectID ||
		len(creature.Attachments) != 1 || creature.Attachments[0] != aura.ObjectID {
		t.Fatal("state-based actions changed phased-out attachment links")
	}
}

func TestPhasingPreservesEffectiveSourceSnapshot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	equipment := addEquipmentWithPTBuff(g, game.Player1, 3, 0)
	if !attachPermanent(g, equipment, creature) {
		t.Fatal("attachPermanent(equipment, creature) = false")
	}
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power before phasing = %d, want 5", got)
	}

	if !phaseOutPermanentTree(g, creature, game.Player1, make(map[game.ObjectID]bool)) {
		t.Fatal("phaseOutPermanentTree() = false")
	}
	resolved, ok := resolveSourcePermanentOrLastKnown(g, creature.ObjectID)
	if !ok || !resolved.snapshot.Power.Exists || resolved.snapshot.Power.Val != 5 {
		t.Fatalf("phased source snapshot power = %v, want 5", resolved.snapshot.Power)
	}
}

func TestPermanentTriggersWhenItPhasesOut(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventPermanentPhasedOut,
		Source: game.TriggerSourceSelf,
	}, nil, nil)

	if !phaseOutPermanentTree(g, source, game.Player1, make(map[game.ObjectID]bool)) {
		t.Fatal("phaseOutPermanentTree() = false")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self phase-out trigger was not put on stack")
	}
}

func TestPhaseOutAttachmentInheritsHostScheduleWhenSelectedFirst(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	host := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := makeAuraAttachedTo(g, game.Player2, host, "Opponent Aura")

	if !phaseOutPermanentTrees(g, []phaseOutRoot{
		{permanent: aura, phaseInFor: game.Player2},
		{permanent: host, phaseInFor: game.Player1},
	}) {
		t.Fatal("phaseOutPermanentTrees() = false")
	}

	for _, permanent := range []*game.Permanent{host, aura} {
		if !permanent.PhasedOut || permanent.PhasedOutFor != game.Player1 {
			t.Fatalf("permanent %d phase state = %+v, want phased out for host controller Player1", permanent.ObjectID, permanent)
		}
	}
}

func TestGroupPhaseOutCapturesAllRootsBeforeMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addAnthemPermanent(g, game.Player1)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.PhaseOut{
		Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
	}, nil)

	resolved, ok := resolveSourcePermanentOrLastKnown(g, target.ObjectID)
	if !ok || !resolved.snapshot.Power.Exists || resolved.snapshot.Power.Val != 3 {
		t.Fatalf("later root snapshot power = %v, want 3 from active anthem", resolved.snapshot.Power)
	}
	if !source.PhasedOut || !target.PhasedOut {
		t.Fatal("group roots did not phase out")
	}
}

func TestGroupPhaseOutCapturesCrossRootTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:       game.EventPermanentPhasedOut,
		Source:      game.TriggerSourceAny,
		ExcludeSelf: true,
		OneOrMore:   true,
	}, nil, nil)
	firstTarget := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	secondTarget := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.PhaseOut{
		Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
	}, nil)

	if !source.PhasedOut || !firstTarget.PhasedOut || !secondTarget.PhasedOut {
		t.Fatal("group roots did not phase out")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("phase-out trigger from an earlier group root was not captured")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("triggered abilities on stack = %d, want 1", g.Stack.Size())
	}
}

func TestPhasingOutAttackTargetClearsAttackDeclaration(t *testing.T) {
	for _, test := range []struct {
		name   string
		target func(*game.Game) (*game.Permanent, game.AttackTarget)
	}{
		{
			name: "planeswalker",
			target: func(g *game.Game) (*game.Permanent, game.AttackTarget) {
				permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name:    "Target Planeswalker",
					Types:   []types.Card{types.Planeswalker},
					Loyalty: opt.Val(5),
				}})
				return permanent, game.AttackTarget{Player: game.Player2, PlaneswalkerID: permanent.ObjectID}
			},
		},
		{
			name: "battle",
			target: func(g *game.Game) (*game.Permanent, game.AttackTarget) {
				permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name:    "Target Battle",
					Types:   []types.Card{types.Battle},
					Defense: opt.Val(5),
				}})
				return permanent, game.AttackTarget{Player: game.Player2, BattleID: permanent.ObjectID}
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
			target, attackTarget := test.target(g)
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
				Attacker: attacker.ObjectID,
				Target:   attackTarget,
			}}}
			startingLife := g.Players[game.Player2].Life

			if !phaseOutPermanentTree(g, target, game.Player2, make(map[game.ObjectID]bool)) {
				t.Fatal("phaseOutPermanentTree() = false")
			}
			if len(g.Combat.Attackers) != 1 || !g.Combat.Attackers[0].Target.NoTarget {
				t.Fatalf("attack declarations after target phased out = %+v, want attacker attacking nothing", g.Combat.Attackers)
			}
			NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
			if g.Players[game.Player2].Life != startingLife {
				t.Fatalf("defending player life = %d, want %d", g.Players[game.Player2].Life, startingLife)
			}
		})
	}
}

func TestBlockedAttackerStillDamagesBlockerAfterAttackTargetPhasesOut(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Target Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(5),
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target: game.AttackTarget{
				Player:         game.Player2,
				PlaneswalkerID: target.ObjectID,
			},
		}},
		Blockers: []game.BlockDeclaration{{
			Blocker:  blocker.ObjectID,
			Blocking: attacker.ObjectID,
		}},
		BlockedAttackers: map[id.ID]bool{attacker.ObjectID: true},
	}

	if !phaseOutPermanentTree(g, target, game.Player2, make(map[game.ObjectID]bool)) {
		t.Fatal("phaseOutPermanentTree() = false")
	}
	if len(g.Combat.Attackers) != 1 || !g.Combat.Attackers[0].Target.NoTarget {
		t.Fatalf("attack declarations after target phased out = %+v, want blocked attacker attacking nothing", g.Combat.Attackers)
	}
	if len(g.Combat.Blockers) != 1 {
		t.Fatalf("block declarations after target phased out = %+v, want blocker preserved", g.Combat.Blockers)
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if attacker.MarkedDamage != 2 {
		t.Fatalf("attacker marked damage = %d, want 2", attacker.MarkedDamage)
	}
	if blocker.MarkedDamage != 3 {
		t.Fatalf("blocker marked damage = %d, want 3", blocker.MarkedDamage)
	}
}

func TestAttackTargetLookupRejectsPhasedOutPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Target Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(5),
	}})
	target.PhasedOut = true

	if permanent, ok := attackTargetPermanent(g, game.AttackTarget{
		Player:         game.Player2,
		PlaneswalkerID: target.ObjectID,
	}); ok || permanent != nil {
		t.Fatalf("attackTargetPermanent() = (%v, %v), want nil and false", permanent, ok)
	}
}

func TestResolvingSourceSpellExilesItselfButCopyDoesNot(t *testing.T) {
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:  "Self Exile",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Exile{SourceSpell: true},
		}}}.Ability()),
	}}

	t.Run("card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToHand(g, game.Player1, spell)
		g.Players[game.Player1].Hand.Remove(cardID)
		g.Stack.Push(&game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackSpell,
			SourceID:   cardID,
			Controller: game.Player1,
		})
		NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !g.Players[game.Player1].Exile.Contains(cardID) ||
			g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("self-exiling spell did not move from stack to exile")
		}
		resolved := false
		for _, event := range g.Events {
			resolved = resolved || event.Kind == game.EventSpellResolved
		}
		if !resolved {
			t.Fatal("self-exiling spell did not emit EventSpellResolved")
		}
	})

	t.Run("copy", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		cardID := addCardToHand(g, game.Player1, spell)
		g.Stack.Push(&game.StackObject{
			ID:         g.IDGen.Next(),
			Kind:       game.StackSpell,
			SourceID:   cardID,
			Controller: game.Player1,
			Copy:       true,
		})
		NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(cardID) ||
			g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("resolving spell copy moved the represented card")
		}
	})
}

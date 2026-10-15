package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestInfectPlayerDamageUsesPoisonInLiveAndSimulatorCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Infect)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies := simPassPolicies()
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}}
	resolved, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
	if !ok || resolved.Players[game.Player2].Life != 40 || resolved.Players[game.Player2].PoisonCounters != 3 {
		t.Fatal("Simulator Infect combat did not preserve life and give three poison")
	}
	policies[game.Player1] = simultaneousCombatPolicy{attacker: attacker.ObjectID, defender: game.Player2}
	live := g.Clone()
	NewEngine(nil).runCombatPhase(live, policies, &TurnLog{})
	if live.Players[game.Player2].Life != 40 || live.Players[game.Player2].PoisonCounters != 3 ||
		g.Players[game.Player2].PoisonCounters != 0 {
		t.Fatal("live Infect or Simulator clone independence disagreed")
	}
}

func TestInfectPlayerNoncombatAbilityAndLifelink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Infect damage ability", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Infect, game.Lifelink)}},
		ActivatedAbilities: []game.ActivatedAbility{{Content: game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount: game.Fixed(3), Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
			}}},
		}.Ability()}},
	}})
	engine := NewEngine(nil)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 0)) {
		t.Fatal("could not activate the real noncombat damage ability")
	}
	resolved := engine.Simulator().ResolvePriority(g, simPassPolicies())
	if resolved.Players[game.Player2].Life != 40 || resolved.Players[game.Player2].PoisonCounters != 3 ||
		resolved.Players[game.Player1].Life != 43 {
		t.Fatalf("noncombat life/poison/source life = %d/%d/%d, want 40/3/43",
			resolved.Players[game.Player2].Life, resolved.Players[game.Player2].PoisonCounters, resolved.Players[game.Player1].Life)
	}
}

func TestInfectPlayerPoisonUsesCounterReplacementNotLifeLossRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Infect)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectDamageDoesntCauseLifeLoss, Controller: game.Player2, AffectedPlayer: game.PlayerYou,
	})
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventCountersAdded, MatchCounterKind: true,
		CounterKindFilter: counter.Poison, CounterMultiplier: 2,
	})
	markPlayerCombatDamage(g, source, game.Player2, 3, &TurnLog{})
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player2].PoisonCounters != 6 {
		t.Fatal("life-loss prohibition suppressed Infect or poison bypassed counter replacements")
	}
	damageFacts, counterFacts := 0, 0
	for _, event := range g.Events {
		if event.Kind == game.EventDamageDealt && event.Player == game.Player2 {
			damageFacts++
			if event.Amount != 3 || event.SourceObjectID != source.ObjectID {
				t.Fatal("Infect changed the actual damage event or source attribution")
			}
		}
		if event.Kind == game.EventCountersAdded && event.Player == game.Player2 && event.CounterKind == counter.Poison {
			counterFacts++
			if event.Amount != 6 || event.Controller != game.Player1 {
				t.Fatal("poison placement lost its replacement-adjusted amount or source controller")
			}
		}
		if event.Kind == game.EventLifeLost && event.Player == game.Player2 {
			t.Fatal("Infect emitted a life-loss result")
		}
	}
	if damageFacts != 1 || counterFacts != 1 || len(g.ReplacementDecisions) != 0 {
		t.Fatal("Infect duplicated facts or introduced a fake replacement ordering choice")
	}
}

func TestInfectPlayerActualDamagePreservesLifelinkToxicAndCommander(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Infect Toxic Commander", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			{KeywordAbilities: game.SimpleKeywords(game.Infect, game.Lifelink)},
			{KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: 2}}},
		},
	}}
	source := addCombatPermanent(g, game.Player1, def)
	g.CommanderIDs[source.CardInstanceID] = true
	g.PreventionShields = []game.PreventionShield{{ID: g.IDGen.Next(), Player: game.Player2, Amount: 1}}
	markPlayerCombatDamage(g, source, game.Player2, 3, &TurnLog{})
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player2].PoisonCounters != 4 ||
		g.Players[game.Player1].Life != 42 || g.Players[game.Player2].CommanderDamage[source.CardInstanceID] != 2 {
		t.Fatal("Infect/Toxic/lifelink/commander results did not all use the actual two damage")
	}
}

func TestInfectPlayerPreventionAndRedirectUseActualRecipient(t *testing.T) {
	for _, redirect := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Infect)
		var protector *game.Permanent
		if redirect {
			protector = addCombatCreaturePermanentWithPower(g, game.Player2, 5)
			g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
				Kind: game.RuleEffectRedirectDamageToSource, Controller: game.Player2,
				AffectedPlayer: game.PlayerYou, SourceObjectID: protector.ObjectID,
			})
		} else {
			g.PreventionShields = []game.PreventionShield{{ID: g.IDGen.Next(), Player: game.Player2, Amount: 3}}
		}
		markPlayerCombatDamage(g, source, game.Player2, 3, &TurnLog{})
		if g.Players[game.Player2].Life != 40 || g.Players[game.Player2].PoisonCounters != 0 {
			t.Fatal("prevented/redirected damage poisoned the original player")
		}
		if redirect && (protector.MarkedDamage != 0 || protector.Counters.Get(counter.MinusOneMinusOne) != 3) {
			t.Fatal("redirected Infect damage did not apply the creature result")
		}
	}
}

func TestInfectPlayerOrdinaryAndWitherDamageStillLoseLife(t *testing.T) {
	for _, keyword := range []game.Keyword{game.KeywordNone, game.Wither} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatCreaturePermanentWithPower(g, game.Player1, 3, keyword)
		markPlayerCombatDamage(g, source, game.Player2, 3, &TurnLog{})
		if g.Players[game.Player2].Life != 37 || g.Players[game.Player2].PoisonCounters != 0 {
			t.Fatal("ordinary/Wither-only player damage incorrectly used Infect results")
		}
	}
}

func TestInfectPlayerGrantedLostAndReenteredSourceIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerAbility, AffectedObjectID: source.ObjectID, AddKeywords: []game.Keyword{game.Infect},
	})
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 2, false)
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("could not record old source LKI")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	g.Players[game.Player1].Graveyard.Remove(card.ID)
	current, entered := createCardPermanent(g, card, game.Player1, zone.Graveyard)
	if !entered {
		t.Fatal("could not reenter the physical card")
	}
	dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 2, false)
	dealPlayerDamage(g, current.CardInstanceID, current.ObjectID, game.Player1, game.Player2, 2, false)
	if g.Players[game.Player2].PoisonCounters != 4 || g.Players[game.Player2].Life != 38 {
		t.Fatal("old-source LKI or new incarnation's lost Infect was misattributed")
	}
}

func TestInfectPlayerTenPoisonUsesExistingSBALoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 10, game.Infect)
	markPlayerCombatDamage(g, source, game.Player2, 10, &TurnLog{})
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player2].Eliminated {
		t.Fatal("Infect directly changed life or eliminated before the existing SBA")
	}
	if lossReason(g, g.Players[game.Player2]) != LossReasonPoisonCounters {
		t.Fatal("ten poison did not select the existing poison loss reason")
	}
	NewEngine(nil).applyStateBasedActions(g)
	if !g.Players[game.Player2].Eliminated {
		t.Fatal("the existing poison SBA did not eliminate the player")
	}
}

func TestInfectPlayerNonpermanentSpellUsesExactStackSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Infect damage spell", Types: []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Infect, game.Lifelink)}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount: game.Fixed(3), Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
			}}},
		}.Ability()),
	}})
	engine := NewEngine(nil)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("could not cast the real nonpermanent damage source")
	}
	resolved := engine.Simulator().ResolvePriority(g, simPassPolicies())
	if resolved.Players[game.Player2].Life != 40 || resolved.Players[game.Player2].PoisonCounters != 3 ||
		resolved.Players[game.Player1].Life != 43 {
		t.Fatal("nonpermanent spell source lost Infect/lifelink while leaving the stack")
	}
}

func TestInfectPlayerGroupDamageLifelinkIsOneGainPerSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Infect, game.Lifelink)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventLifeGained, Controller: game.Player1,
		ControllerFilter: game.TriggerControllerYou, LifeGainAddend: 1,
	})
	resolveInstruction(NewEngine(nil), g, &game.StackObject{
		Kind: game.StackActivatedAbility, Controller: game.Player1,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
	}, game.Damage{
		Amount: game.Fixed(1), Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
	}, &TurnLog{})
	if g.Players[game.Player1].Life != 44 {
		t.Fatalf("one source's simultaneous lifelink gain = %d, want damage3 + replacement1",
			g.Players[game.Player1].Life-40)
	}
	gains := 0
	for _, event := range g.Events {
		if event.Kind == game.EventLifeGained {
			gains++
		}
	}
	if gains != 1 {
		t.Fatalf("life-gain facts = %d, want one per source", gains)
	}
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if g.Players[player].PoisonCounters != 1 || g.Players[player].Life != 40 {
			t.Fatal("group Infect results changed while aggregating lifelink")
		}
	}
}

func TestInfectPlayerMultiTargetLifelinkIsOneGain(t *testing.T) {
	for _, divided := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		source := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Infect, game.Lifelink)
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), MatchEvent: game.EventLifeGained, Controller: game.Player1,
			ControllerFilter: game.TriggerControllerYou, LifeGainAddend: 1,
		})
		amount := 1
		if divided {
			amount = 3
		}
		resolveInstruction(NewEngine(nil), g, &game.StackObject{
			Kind: game.StackActivatedAbility, Controller: game.Player1,
			SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
			Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3), game.PlayerTarget(game.Player4)},
		}, game.Damage{
			Amount: game.Fixed(amount), Recipient: game.AnyTargetDamageRecipient(0),
			Divided: divided, EachTarget: !divided,
		}, &TurnLog{})
		if g.Players[game.Player1].Life != 44 {
			t.Fatalf("multi-target lifelink gain = %d, want one gain of four (divided=%v)",
				g.Players[game.Player1].Life-40, divided)
		}
	}
}

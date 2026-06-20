package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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
	if got := applyDamagePrevention(g, damageEvent{
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

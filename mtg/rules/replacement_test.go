package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestShieldCounterPreventsDamageBeforeMutationAndEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == target.ObjectID &&
			event.Amount == 3 &&
			event.DamageRecipient == game.DamageRecipientPermanent
	})
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestShieldCounterReplacesDestroyBeforeZoneChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target.Counters.Add(counter.Shield, 1)

	removed, ok := destroyPermanent(g, target.ObjectID)

	if ok || removed != nil {
		t.Fatalf("destroyPermanent() = %+v, %v, want nil, false for replaced destroy", removed, ok)
	}
	if permanentByObjectID(g, target.ObjectID) == nil {
		t.Fatal("shield-replaced permanent left the battlefield")
	}
	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want 0", target.Counters.Get(counter.Shield))
	}
	if g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("shield-replaced permanent moved to graveyard")
	}
	assertEvent(t, g.Events, game.EventDestroyReplaced, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID &&
			event.FromZone == game.ZoneBattlefield &&
			event.ToZone == game.ZoneGraveyard
	})
	assertNoEvent(t, g.Events, game.EventPermanentDied, func(event game.GameEvent) bool {
		return event.PermanentID == target.ObjectID
	})
}

func TestPreventedCombatDamageDoesNotGrantLifelinkOrMarkDeathtouch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Deathtouch, game.Lifelink)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker.Counters.Add(counter.Shield, 1)
	g.Players[game.Player1].Life = 40
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("lifelink controller life = %d, want 40", g.Players[game.Player1].Life)
	}
	if blocker.MarkedDamage != 0 || blocker.MarkedDeathtouchDamage {
		t.Fatalf("blocker damage = %d deathtouch = %v, want no marked damage", blocker.MarkedDamage, blocker.MarkedDeathtouchDamage)
	}
}

func TestProtectionFromColorPreventsDamageAndTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, mana.Red)
	protected := addProtectionFromColorPermanent(g, game.Player2, mana.Red)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, protected, 2, false)

	if dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0", dealt)
	}
	if protected.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", protected.MarkedDamage)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.GameEvent) bool {
		return event.SourceID == sourceID &&
			event.PermanentID == protected.ObjectID &&
			event.Amount == 2
	})

	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:   "Red Strike",
		Types:  []game.CardType{game.TypeInstant},
		Colors: []mana.Color{mana.Red},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				},
				Effects: []game.Effect{{Type: game.EffectDamage, Amount: 1, TargetIndex: 0}},
			},
		},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if engine.canCastSpell(g, game.Player1, spellID, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil) {
		t.Fatal("red spell could target a permanent with protection from red")
	}
}

func addColoredSourceCard(g *game.Game, owner game.PlayerID, color mana.Color) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{
			Name:   "Colored Source",
			Types:  []game.CardType{game.TypeInstant},
			Colors: []mana.Color{color},
		},
		Owner: owner,
	}
	return cardID
}

func addProtectionFromColorPermanent(g *game.Game, controller game.PlayerID, color mana.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:      "Protected Creature",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &pt,
		Toughness: &pt,
		Abilities: []game.AbilityDef{
			{
				Kind:                 game.StaticAbility,
				Keywords:             []game.Keyword{game.Protection},
				ProtectionFromColors: []mana.Color{color},
			},
		},
	})
}

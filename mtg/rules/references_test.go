package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestRecipientReferenceUsesDestroyedTargetControllerLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Borrowed Permanent",
		Types: []types.Card{types.Artifact}},
	})
	target.Controller = game.Player3
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	token := &game.CardDef{CardFace: game.CardFace{Name: "Beast",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Destroy{Object: game.TargetPermanentReference(0)}, &log)
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(token),
		Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
	}, &log)

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target permanent remained on battlefield")
	}
	if got := countControlledTokensNamed(g, game.Player3, types.Beast); got != 1 {
		t.Fatalf("Player3 Beast tokens = %d, want 1", got)
	}
	if got := countControlledTokensNamed(g, game.Player1, types.Beast); got != 0 {
		t.Fatalf("spell controller Beast tokens = %d, want 0", got)
	}
	if got := countControlledTokensNamed(g, game.Player2, types.Beast); got != 0 {
		t.Fatalf("target owner Beast tokens = %d, want 0", got)
	}
}

func TestDamageRecipientReferenceHitsDestroyedTargetControllerLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Borrowed Land",
		Types: []types.Card{types.Land}},
	})
	target.Controller = game.Player3
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Destroy{Object: game.TargetPermanentReference(0)}, &log)
	resolveInstruction(engine, g, obj, game.Damage{
		Amount:    game.Fixed(2),
		Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
	}, &log)

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target permanent remained on battlefield")
	}
	if got := g.Players[game.Player3].Life; got != 38 {
		t.Fatalf("destroyed permanent controller life = %d, want 38", got)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("spell controller life = %d, want 40 (unharmed)", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("target owner life = %d, want 40 (unharmed)", got)
	}
}

func TestDamageSourceReferenceAppliesCreatureDamageKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Venomous Healer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Deathtouch, game.Lifelink),
		}}},
	})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5})},
	})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets: []game.Target{
			game.PermanentTarget(source.ObjectID),
			game.PermanentTarget(target.ObjectID),
		},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(1),
		DamageSource: opt.Val(game.TargetPermanentReference(0)),
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:   game.DynamicAmountTargetPower,
			Object: game.TargetPermanentReference(0),
		}),
	}, &log)

	if got := target.MarkedDamage; got != 2 {
		t.Fatalf("marked damage = %d, want 2", got)
	}
	if !target.MarkedDeathtouchDamage {
		t.Fatal("target was not marked with deathtouch damage")
	}
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("Player1 life = %d, want 42 from lifelink", got)
	}
}

func TestEventPermanentDamageSourceUsesLastKnownKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Dying Devil",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Deathtouch, game.Lifelink),
		}},
	}})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	sourceID := source.ObjectID
	source.MarkedDamage = 2
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, sourceID); ok {
		t.Fatal("damage source remained on battlefield")
	}
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: sourceID},
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.EventPermanentReference()),
		Amount:       game.Fixed(2),
	}, &log)

	if got := target.MarkedDamage; got != 2 {
		t.Fatalf("marked damage = %d, want 2", got)
	}
	if !target.MarkedDeathtouchDamage {
		t.Fatal("target was not marked with last-known deathtouch damage")
	}
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("Player1 life = %d, want 42 from last-known lifelink", got)
	}
}

func TestEventPermanentModifyPTResolvesLiveSubject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanent(g, game.Player1)
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: permanent.ObjectID},
	}

	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.EventPermanentReference(),
		PowerDelta:     game.Fixed(2),
		ToughnessDelta: game.Fixed(0),
		Duration:       game.DurationUntilEndOfTurn,
	}, nil)

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("continuous effects = %d, want event-subject P/T modification", len(g.ContinuousEffects))
	}
}

func TestEventPermanentDamageSourceUsesKeywordsFromSimultaneousDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	granter := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Dying Mentor",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				AddKeywords: []game.Keyword{game.Deathtouch, game.Lifelink},
			}},
		}},
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Dying Devil",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	granter.MarkedDamage = 1
	source.MarkedDamage = 1
	sourceID := source.ObjectID

	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, sourceID); ok {
		t.Fatal("damage source remained on battlefield")
	}
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: sourceID},
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	log := TurnLog{}
	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.EventPermanentReference()),
		Amount:       game.Fixed(2),
	}, &log)

	if !target.MarkedDeathtouchDamage {
		t.Fatal("target was not marked with simultaneously granted deathtouch damage")
	}
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("Player1 life = %d, want 42 from simultaneously granted lifelink", got)
	}
}

func TestEventPermanentDamageSourceUsesKeywordsFromGroupDestroy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	granter := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Dying Mentor",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
				AddKeywords: []game.Keyword{game.Deathtouch, game.Lifelink},
			}},
		}},
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Dying Devil",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	sourceID := source.ObjectID
	log := TurnLog{}
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}, &log)

	if _, ok := permanentByObjectID(g, granter.ObjectID); ok {
		t.Fatal("keyword granter remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, sourceID); ok {
		t.Fatal("damage source remained on battlefield")
	}
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: sourceID},
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	resolveInstruction(engine, g, obj, game.Damage{
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.EventPermanentReference()),
		Amount:       game.Fixed(2),
	}, &log)

	if !target.MarkedDeathtouchDamage {
		t.Fatal("target was not marked with group-destroyed source's deathtouch damage")
	}
	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("Player1 life = %d, want 42 from group-destroyed source's lifelink", got)
	}
}

func TestLegacyTokenCreationStillUsesSpellController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Legacy Token", Types: []types.Card{types.Creature}}}
	obj := &game.StackObject{Controller: game.Player1}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenDef(token),
	}, &log)

	if got := countControlledTokensNamed(g, game.Player1, "Legacy Token"); got != 1 {
		t.Fatalf("Player1 legacy tokens = %d, want 1", got)
	}
}

func TestEventPlayerReferenceUsesZeroValuedEventPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:   game.EventCardDrawn,
			Player: game.Player1,
		},
	}

	player, ok := newReferenceResolver(g, obj).player(game.EventPlayerReference())
	if !ok || player != game.Player1 {
		t.Fatalf("event player = %v (%v), want Player1", player, ok)
	}
}

func TestEventPlayerReferenceUsesSpellController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:       game.EventSpellCast,
			Controller: game.Player3,
		},
	}

	player, ok := newReferenceResolver(g, obj).player(game.EventPlayerReference())
	if !ok || player != game.Player3 {
		t.Fatalf("event player = %v (%v), want Player3", player, ok)
	}
}

func TestEventPlayerReferenceRejectsEventWithoutPlayerSubject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:            game.EventDamageDealt,
			Controller:      game.Player2,
			DamageRecipient: game.DamageRecipientPermanent,
		},
	}

	if player, ok := newReferenceResolver(g, obj).player(game.EventPlayerReference()); ok {
		t.Fatalf("event player = %v, want unresolved", player)
	}
}

func countControlledTokensNamed(g *game.Game, controller game.PlayerID, name types.Sub) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.Controller != controller || permanent.TokenDef == nil || permanent.TokenDef.Name != string(name) {
			continue
		}
		count++
	}
	return count
}

// TestDamageControllerRecipientHitsSpellController covers the "deals N damage to
// you" lowering, where the recipient is the controller of the resolving spell.
// Only the controller loses life; opponents are unharmed.
func TestDamageControllerRecipientHitsSpellController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{Controller: game.Player1}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.PlayerDamageRecipient(game.ControllerReference()),
	}, &log)

	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("spell controller life = %d, want 37", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("opponent life = %d, want 40 (unharmed)", got)
	}
}

// TestSelfDamageRiderHitsTargetAndController covers the "deals A damage to
// <target> and B damage to you" lowering: the chosen target and the controller
// are each damaged by their own instruction, with the controller taking only the
// rider amount.
func TestSelfDamageRiderHitsTargetAndController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.Damage{
		Amount:    game.Fixed(4),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, &log)
	resolveInstruction(engine, g, obj, game.Damage{
		Amount:    game.Fixed(2),
		Recipient: game.PlayerDamageRecipient(game.ControllerReference()),
	}, &log)

	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("target player life = %d, want 36", got)
	}
	if got := g.Players[game.Player1].Life; got != 38 {
		t.Fatalf("spell controller life = %d, want 38 (rider only)", got)
	}
}

// TestDefendingPlayerReferenceSacrificesOnAttack covers Annihilator: a creature
// with an attack-triggered ability whose effect makes the defending player
// sacrifice permanents resolves the reference to the attack's defending player.
func TestDefendingPlayerReferenceSacrificesOnAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, Player: game.Player2},
	}
	log := TurnLog{}

	resolveInstruction(engine, g, obj, game.SacrificePermanents{
		Player: game.DefendingPlayerReference(),
		Amount: game.Fixed(2),
	}, &log)

	if _, ok := permanentByObjectID(g, first.ObjectID); ok {
		t.Fatal("first defending permanent was not sacrificed")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("second defending permanent was not sacrificed")
	}
	if !g.Players[game.Player2].Graveyard.Contains(first.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(second.CardInstanceID) {
		t.Fatal("sacrificed permanents were not moved to defending player's graveyard")
	}
}

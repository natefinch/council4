package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #2347 ("When that creature dies this turn" delayed
// triggers, blocking Grim Javelineer and 18 other cards). Oracle text from the
// pinned corpus 6da9ced0…; Scryfall rulings retained under .cardwork/gap-2347
// for Searing Blood (damage goes to the creature's controller as it last
// existed on the battlefield, however it died), Cryptek and Reincarnation (the
// resolution only creates the delayed trigger), Burn Away and Virulent Wound
// (the trigger can't fire on a later turn). CR 2026-09-25 (verified official
// copy): 603.7a (created on resolution), 603.7b ("this turn" is a stated
// duration), 603.7c (a returned card is a new object), 603.7d-e (source and
// controller of the delayed trigger), 603.10a (dies triggers look back),
// 608.2b (an all-illegal-target spell creates nothing), 608.2h/113.7a (last
// known information) and 700.4 (dies). These scenarios are test-local
// fixtures; no card is registered.

// diesThisTurnTrigger is the shared engine shape for "When that creature dies
// this turn, <content>" bound to the instruction's target 0.
func diesThisTurnTrigger(content ...game.Instruction) game.Instruction {
	return game.Instruction{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:               game.EventPermanentDied,
			DyingObjectCaptured: true,
			SubjectSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
		}),
		Window:              game.DelayedWindowThisTurn,
		CapturedDyingObject: opt.Val(game.TargetPermanentReference(0)),
		Content:             game.Mode{Sequence: content}.Ability(),
	}}}
}

func targetCreatureSpec(constraint string, selection game.Selection) game.TargetSpec {
	selection.RequiredTypesAny = []types.Card{types.Creature}
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Constraint: constraint,
		Allow: game.TargetAllowPermanent, Selection: opt.Val(selection),
	}
}

// searingBloodShape is Searing Blood as authored Engine operations: the 3
// damage is the delayed trigger's content, not part of the spell's resolution.
func searingBloodShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Searing Blood (delayed dies fixture)",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{targetCreatureSpec("target creature", game.Selection{})},
			Sequence: []game.Instruction{
				{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
				diesThisTurnTrigger(game.Instruction{Primitive: game.Damage{
					Amount:    game.Fixed(3),
					Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.EventPermanentReference())),
				}}),
			},
		}.Ability()),
	}}
}

func castDiesThisTurnSpell(t *testing.T, g *game.Game, engine *Engine, def *game.CardDef, target *game.Permanent) (spellCard, stackID game.ObjectID) {
	t.Helper()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("fixture validation issues: %v", issues)
	}
	spellCard = addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PermanentTarget(target.ObjectID)})
	obj, _ := g.Stack.Peek()
	obj.TargetCounts = []int{1}
	stackID = obj.ID
	engine.resolveTopOfStack(g, &TurnLog{})
	return spellCard, stackID
}

// fireDiesThisTurn runs state-based actions, then puts every ready trigger on
// the stack and resolves exactly one delayed trigger, returning its stack
// object as it existed before resolution.
func fireDiesThisTurn(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent) game.StackObject {
	t.Helper()
	engine.applyStateBasedActions(g)
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) || g.Stack.Size() != 1 {
		t.Fatalf("delayed dies trigger did not fire alone: stack size %d", g.Stack.Size())
	}
	obj, _ := g.Stack.Peek()
	fired := *obj
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after delayed trigger = %d, want 0", g.Stack.Size())
	}
	return fired
}

func assertNoDelayedDiesFire(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	engine.applyStateBasedActions(g)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 0 {
		t.Fatalf("delayed dies trigger fired unexpectedly: stack size %d", g.Stack.Size())
	}
}

// TestDelayedDiesThisTurnSpellDamagesLastKnownController covers the spell
// route (Searing Blood): resolution deals only the 2 damage and schedules the
// trigger; the 3 damage comes from the delayed trigger, from the spell card as
// its source, to whoever controlled the creature when it died rather than its
// owner.
func TestDelayedDiesThisTurnSpellDamagesLastKnownController(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	victim.Controller = game.Player3

	spellCard, stackID := castDiesThisTurnSpell(t, g, engine, searingBloodShape(), victim)

	if victim.MarkedDamage != 2 {
		t.Fatalf("victim marked damage = %d, want 2", victim.MarkedDamage)
	}
	for _, player := range []game.PlayerID{game.Player2, game.Player3} {
		if life := g.Players[player].Life; life != 40 {
			t.Fatalf("player %v life after resolution = %d, want 40 (no immediate damage)", player, life)
		}
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].BoundDyingObjectID != victim.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one bound to the victim", g.DelayedTriggers)
	}

	fired := fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

	if fired.Controller != game.Player1 || fired.SourceID != stackID || fired.SourceCardID != spellCard {
		t.Fatalf("delayed trigger controller/source/card = %v/%v/%v, want %v/spell %v/card %v", fired.Controller, fired.SourceID, fired.SourceCardID, game.Player1, stackID, spellCard)
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not die to its owner's graveyard")
	}
	if got := g.Players[game.Player3].Life; got != 37 {
		t.Fatalf("last-known controller life = %d, want 37", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("owner life = %d, want 40", got)
	}
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Player == game.Player3 && event.Amount == 3 && event.SourceID == spellCard
	})
	// "This turn" is a stated duration (CR 603.7b), so the trigger stays until
	// cleanup, but the dead object can never die again and nothing refires.
	assertNoDelayedDiesFire(t, g, engine)
	expireCleanupDurations(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived cleanup: %+v", g.DelayedTriggers)
	}
}

// surveilToGraveyardAgent puts surveiled cards into the graveyard and takes
// the default for every other choice, including the attack trigger's target.
type surveilToGraveyardAgent struct{ monarchNoChoiceAgent }

func (surveilToGraveyardAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceSurveil {
		return []int{1}
	}
	return request.DefaultSelection
}

// TestDelayedDiesThisTurnAttackTriggerSurveils covers the triggered-ability
// route through real attack declaration (Grim Javelineer): the attack trigger
// targets the attacker, pumps it, and its death later that turn makes the
// trigger's controller surveil 1.
func TestDelayedDiesThisTurnAttackTriggerSurveils(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	javelineer := &game.CardDef{CardFace: game.CardFace{
		Name:  "Grim Javelineer (delayed dies fixture)",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventAttackerDeclared, Controller: game.TriggerControllerYou, OneOrMore: true,
			}},
			Content: game.Mode{
				Targets: []game.TargetSpec{targetCreatureSpec("target attacking creature", game.Selection{CombatState: game.CombatStateAttacking})},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(1), Duration: game.DurationUntilEndOfTurn}},
					diesThisTurnTrigger(game.Instruction{Primitive: game.Surveil{Amount: game.Fixed(1), Player: game.ControllerReference()}}),
				},
			}.Ability(),
		}},
	}}
	if issues := game.ValidateCardDef(javelineer); len(issues) != 0 {
		t.Fatalf("fixture validation issues: %v", issues)
	}
	source := addCombatPermanent(g, game.Player1, javelineer)
	top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})

	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declarations, _ := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}).DeclareAttackersPayload()
	if !engine.applyDeclareAttackers(g, game.Player1, declarations) {
		t.Fatal("attack declaration failed")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: surveilToGraveyardAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) || g.Stack.Size() != 1 {
		t.Fatalf("attack trigger stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if got := effectivePower(g, source); got != 3 {
		t.Fatalf("attacker power = %d, want 3", got)
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].BoundDyingObjectID != source.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one bound to the attacker", g.DelayedTriggers)
	}

	source.MarkedDamage = 2
	fired := fireDiesThisTurn(t, g, engine, agents)

	if fired.Controller != game.Player1 {
		t.Fatalf("delayed trigger controller = %v, want %v", fired.Controller, game.Player1)
	}
	if !g.Players[game.Player1].Graveyard.Contains(top) || g.Players[game.Player1].Library.Contains(top) {
		t.Fatal("surveil 1 did not put the chosen top card into the graveyard")
	}
	assertEvent(t, g.Events, game.EventSurveil, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 1
	})
}

// TestDelayedDiesThisTurnActivatedAbilities covers the activated-ability route
// through real activation (Kelsien's experience counter and Cryptek's return of
// the dead card under the ability controller's control, tapped).
func TestDelayedDiesThisTurnActivatedAbilities(t *testing.T) {
	activated := func(name string, targets game.TargetSpec, seq ...game.Instruction) *game.CardDef {
		def := &game.CardDef{CardFace: game.CardFace{
			Name: name, Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{{
				AdditionalCosts: cost.Tap, ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{Targets: []game.TargetSpec{targets}, Sequence: seq}.Ability(),
			}},
		}}
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("%s validation issues: %v", name, issues)
		}
		return def
	}
	activate := func(t *testing.T, g *game.Game, e *Engine, source, target *game.Permanent) {
		t.Helper()
		if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0)) {
			t.Fatal("activation failed")
		}
		e.resolveTopOfStack(g, &TurnLog{})
	}

	t.Run("Kelsien experience counter", func(t *testing.T) {
		g, engine := setupBestowMain(t)
		source := addCombatPermanent(g, game.Player1, activated("Kelsien (delayed dies fixture)",
			targetCreatureSpec("target creature you don't control", game.Selection{Controller: game.ControllerOpponent}),
			game.Instruction{Primitive: game.Damage{Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0), DamageSource: opt.Val(game.SourcePermanentReference())}},
			diesThisTurnTrigger(game.Instruction{Primitive: game.AddPlayerCounter{
				Amount: game.Fixed(1), Player: game.ControllerReference(), CounterKind: counter.Experience,
			}}),
		))
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 1)

		activate(t, g, engine, source, victim)
		if g.Players[game.Player1].ExperienceCounters != 0 {
			t.Fatal("experience counter was given before the creature died")
		}
		fired := fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

		if fired.Controller != game.Player1 || fired.SourceID != source.ObjectID {
			t.Fatalf("delayed trigger controller/source = %v/%v, want %v/%v", fired.Controller, fired.SourceID, game.Player1, source.ObjectID)
		}
		if got := g.Players[game.Player1].ExperienceCounters; got != 1 {
			t.Fatalf("experience counters = %d, want 1", got)
		}
	})

	t.Run("Cryptek returns the dead card under your control tapped", func(t *testing.T) {
		g, engine := setupBestowMain(t)
		source := addCombatPermanent(g, game.Player1, activated("Cryptek (delayed dies fixture)",
			targetCreatureSpec("another target artifact creature you control", game.Selection{Controller: game.ControllerYou}),
			diesThisTurnTrigger(game.Instruction{Primitive: game.PutOnBattlefield{
				Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
				Recipient:   opt.Val(game.ControllerReference()),
				EntryTapped: true,
			}}),
		))
		// A creature Player1 controls but Player2 owns dies to Player2's
		// graveyard and must still return under Player1's control.
		borrowed := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		borrowed.Controller = game.Player1

		activate(t, g, engine, source, borrowed)
		borrowed.MarkedDamage = 2
		fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

		if g.Players[game.Player2].Graveyard.Contains(borrowed.CardInstanceID) {
			t.Fatal("dead card stayed in its owner's graveyard")
		}
		returned := permanentByCardID(g, borrowed.CardInstanceID)
		if returned == nil {
			t.Fatal("dead card did not return to the battlefield")
		}
		if returned.ObjectID == borrowed.ObjectID {
			t.Fatal("returned permanent kept its old object identity")
		}
		if returned.Controller != game.Player1 || returned.Owner != game.Player2 || !returned.Tapped {
			t.Fatalf("returned permanent controller/owner/tapped = %v/%v/%v, want %v/%v/true", returned.Controller, returned.Owner, returned.Tapped, game.Player1, game.Player2)
		}
	})
}

var diesThisTurnPoisonProgram = game.ProgramReference{ID: "fixture/dies-this-turn-poison", Version: "1"}

// TestDelayedDiesThisTurnProgramBodyPoisonsLastKnownController covers the
// authored Go resolving-body route: the program executes CreateDelayedTrigger
// with the same captured-dying binding (Virulent Wound's poison counter to the
// creature's last-known controller).
func TestDelayedDiesThisTurnProgramBodyPoisonsLastKnownController(t *testing.T) {
	entry := cardset.Entry{
		OracleID: "fixture/dies-this-turn-poison", Name: "Delayed dies poison fixture",
		Verification: "mtg/rules/delayed_dies_this_turn_test.go: program delayed dies fixture",
		New: func() *game.CardDef {
			return &game.CardDef{CardFace: game.CardFace{
				Name: "Delayed dies poison fixture", Types: []types.Card{types.Artifact},
				ActivatedAbilities: []game.ActivatedAbility{{
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{targetCreatureSpec("target creature", game.Selection{})},
						Program: diesThisTurnPoisonProgram,
					}.Ability(),
				}},
			}}
		},
		Programs: []cardprogram.Program{{Reference: diesThisTurnPoisonProgram, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(diesThisTurnTrigger(game.Instruction{Primitive: game.AddPlayerCounter{
				Amount:      game.Fixed(1),
				Player:      game.ObjectControllerReference(game.EventPermanentReference()),
				CounterKind: counter.Poison,
			}}).Primitive)
			return err
		}}},
	}
	registry, err := cards.NewProgramRegistry([]cardset.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}
	def, err := registry.LookupOracleID("fixture/dies-this-turn-poison")
	if err != nil {
		t.Fatal(err)
	}
	g, _ := setupBestowMain(t)
	engine := NewEngineWithPrograms(nil, registry.Programs())
	source := addCombatPermanent(g, game.Player1, def)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	victim.Controller = game.Player4

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(victim.ObjectID)}, 0)) {
		t.Fatal("program fixture did not activate")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].BoundDyingObjectID != victim.ObjectID {
		t.Fatalf("delayed triggers = %+v, want one bound to the victim", g.DelayedTriggers)
	}

	victim.MarkedDamage = 2
	fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

	if got := g.Players[game.Player4].PoisonCounters; got != 1 {
		t.Fatalf("last-known controller poison = %d, want 1", got)
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 0 {
		t.Fatalf("owner poison = %d, want 0", got)
	}
}

// TestDelayedDiesThisTurnReadsLastKnownPower proves "its power" in the
// delayed content reads the creature as it last existed, including a pump
// that ended when it left (End-Blaze Epiphany's power-sized amount).
func TestDelayedDiesThisTurnReadsLastKnownPower(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Last-known power fixture", Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{targetCreatureSpec("target creature", game.Selection{})},
			Sequence: []game.Instruction{
				{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(3), Duration: game.DurationUntilEndOfTurn}},
				diesThisTurnTrigger(game.Instruction{Primitive: game.GainLife{
					Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountObjectPower, Object: game.EventPermanentReference()}),
					Player: game.ControllerReference(),
				}}),
			},
		}.Ability()),
	}}
	castDiesThisTurnSpell(t, g, engine, def, victim)

	victim.MarkedDamage = 2
	fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

	if got := g.Players[game.Player1].Life; got != 45 {
		t.Fatalf("life = %d, want 45 from the dead creature's last-known power 5", got)
	}
}

// TestDelayedDiesThisTurnExilesLastKnownControllersGraveyard covers Burn
// Away: the delayed trigger resolves after the creature is already in the
// graveyard, so the dead card is exiled along with the rest of that graveyard.
func TestDelayedDiesThisTurnExilesLastKnownControllersGraveyard(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	earlier := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Earlier"}})
	mine := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mine"}})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Burn Away (delayed dies fixture)", Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{targetCreatureSpec("target creature", game.Selection{})},
			Sequence: []game.Instruction{
				{Primitive: game.Damage{Amount: game.Fixed(6), Recipient: game.AnyTargetDamageRecipient(0)}},
				diesThisTurnTrigger(game.Instruction{Primitive: game.MoveCard{
					Player:      game.ObjectControllerReference(game.EventPermanentReference()),
					FromZone:    zone.Graveyard,
					Destination: zone.Exile,
				}}),
			},
		}.Ability()),
	}}
	castDiesThisTurnSpell(t, g, engine, def, victim)
	if !g.Players[game.Player2].Graveyard.Contains(earlier) {
		t.Fatal("graveyard was exiled before the creature died")
	}

	fireDiesThisTurn(t, g, engine, [game.NumPlayers]PlayerAgent{game.Player1: monarchNoChoiceAgent{}})

	if size := g.Players[game.Player2].Graveyard.Size(); size != 0 {
		t.Fatalf("Player2 graveyard size = %d, want 0", size)
	}
	for _, card := range []id.ID{earlier, victim.CardInstanceID} {
		if !g.Players[game.Player2].Exile.Contains(card) {
			t.Fatalf("card %v was not exiled from the dead creature's controller's graveyard", card)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(mine) {
		t.Fatal("another player's graveyard was exiled")
	}
}

// TestDelayedDiesThisTurnOnlyMatchesThatCreatureDyingThisTurn proves the
// binding and window: another creature dying, the bound creature leaving
// without dying, a later turn, and an illegal target at resolution all fail to
// fire the trigger.
func TestDelayedDiesThisTurnOnlyMatchesThatCreatureDyingThisTurn(t *testing.T) {
	t.Run("another creature dies", func(t *testing.T) {
		g := mainPhaseGame(game.Player1)
		engine := NewEngine(nil)
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
		bystander := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
		castDiesThisTurnSpell(t, g, engine, searingBloodShape(), victim)

		bystander.MarkedDamage = 1
		assertNoDelayedDiesFire(t, g, engine)
		if _, ok := permanentByObjectID(g, bystander.ObjectID); ok {
			t.Fatal("bystander did not die")
		}
		if len(g.DelayedTriggers) != 1 || g.Players[game.Player2].Life != 40 {
			t.Fatalf("delayed triggers/life = %d/%d, want still-waiting/40", len(g.DelayedTriggers), g.Players[game.Player2].Life)
		}
	})

	t.Run("bound creature is exiled", func(t *testing.T) {
		g := mainPhaseGame(game.Player1)
		engine := NewEngine(nil)
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
		castDiesThisTurnSpell(t, g, engine, searingBloodShape(), victim)

		if !movePermanentToZone(g, victim, zone.Exile) {
			t.Fatal("exile failed")
		}
		assertNoDelayedDiesFire(t, g, engine)
		if g.Players[game.Player2].Life != 40 {
			t.Fatalf("Player2 life = %d, want 40", g.Players[game.Player2].Life)
		}
	})

	t.Run("bound creature dies on a later turn", func(t *testing.T) {
		g := mainPhaseGame(game.Player1)
		engine := NewEngine(nil)
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
		castDiesThisTurnSpell(t, g, engine, searingBloodShape(), victim)

		expireCleanupDurations(g)
		if len(g.DelayedTriggers) != 0 {
			t.Fatalf("this-turn delayed trigger survived cleanup: %+v", g.DelayedTriggers)
		}
		victim.MarkedDamage = 5
		assertNoDelayedDiesFire(t, g, engine)
		if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
			t.Fatal("victim did not die")
		}
		if g.Players[game.Player2].Life != 40 {
			t.Fatalf("Player2 life = %d, want 40", g.Players[game.Player2].Life)
		}
	})

	t.Run("target illegal at resolution", func(t *testing.T) {
		g := mainPhaseGame(game.Player1)
		engine := NewEngine(nil)
		victim := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
		other := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
		addImplementationSpellToStack(g, game.Player1, searingBloodShape(), []game.Target{game.PermanentTarget(victim.ObjectID)})
		obj, _ := g.Stack.Peek()
		obj.TargetCounts = []int{1}
		if !movePermanentToZone(g, victim, zone.Graveyard) {
			t.Fatal("removing target failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if len(g.DelayedTriggers) != 0 {
			t.Fatalf("illegal-target spell scheduled %+v", g.DelayedTriggers)
		}
		other.MarkedDamage = 1
		assertNoDelayedDiesFire(t, g, engine)
		if g.Players[game.Player2].Life != 40 {
			t.Fatalf("Player2 life = %d, want 40", g.Players[game.Player2].Life)
		}
	})
}

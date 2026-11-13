package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: one-shot "a source of your choice" damage redirection
// (#3806). Comprehensive Rules effective 2026-11-06 (local verified copy
// .cardwork/MagicCompRules-20261106.txt):
//   - 614.9: a redirection effect replaces damage dealt to one battle,
//     creature, planeswalker or player with the same damage dealt to another;
//     if the destination permanent has left the battlefield or is no longer a
//     battle, creature or planeswalker, or a player has left the game, the
//     effect does nothing.
//   - 609.7a/609.7b: the source is chosen as the effect is created; a shield
//     that replaces no damage is not used up.
//   - 614.5, 616.2: a replacement modifies an event once, and the redirected
//     event can then be modified by effects applying to its new recipient.
//   - 615.12: unpreventable damage can still be redirected.
// Scryfall Oracle (pinned corpus 6da9ced0...):
//   - Beacon of Destiny a1ed2274 "{T}: The next time a source of your choice
//     would deal damage to you this turn, that damage is dealt to this
//     creature instead." General's Regalia cceba6f3 "{3}: ... to target
//     creature you control instead." Nova Pentacle db121503 "{3}, {T}: ... to
//     target creature of an opponent's choice instead." with its 2009-10-01
//     rulings: an opponent chooses the target on activation, the source is
//     chosen on resolution, and the damage isn't redirected to a target no
//     longer a creature or on the battlefield.
//   - Jade Monolith 1e105ab7 "{1}: The next time a source of your choice would
//     deal damage to target creature this turn, that source deals that damage
//     to you instead."
//   - Reflect Damage 21565b0d "The next time a source of your choice would
//     deal damage this turn, that damage is dealt to that source's controller
//     instead."
//   - Opal-Eye, Konda's Yojimbo e4acac65 "{T}: The next time a source of your
//     choice would deal damage this turn, that damage is dealt to Opal-Eye
//     instead. {1}{W}: Prevent the next 1 damage that would be dealt to
//     Opal-Eye this turn."
//   - Shaman en-Kor 4d9d5dbb second ability "{1}{W}: The next time a source of
//     your choice would deal damage to target creature this turn, that damage
//     is dealt to this creature instead." (2013-07-01 ruling: redirected combat
//     damage is still combat damage). Its first, fixed-amount ability is not
//     this capability.
//   - Eye for an Eye 22647b1a "The next time a source of your choice would
//     deal damage to you this turn, instead that source deals that much damage
//     to you and Eye for an Eye deals that much damage to that source's
//     controller." (2009-02-01 ruling: the damage to you is still dealt by the
//     original source; the other damage is dealt by Eye for an Eye.)

// redirectChoiceAgent picks sourceCard for the shield's source choice, the
// option whose target is target for a target choice, and the default
// otherwise.
func redirectChoiceAgent(t *testing.T, sourceCard id.ID, target game.Target) *controlScript {
	t.Helper()
	sources := sourceChoiceAgent(t, sourceCard, nil)
	return &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceTarget {
			for _, option := range request.Options {
				if len(option.Targets) == 1 && option.Targets[0] == target {
					return []int{option.Index}
				}
			}
			t.Fatalf("target %+v was not offered in %+v", target, request.Options)
		}
		return sources.choice(obs, request)
	}}
}

func redirectAbility(text string, manaCost opt.V[cost.Mana], additional []cost.Additional, targets []game.TargetSpec, prim game.Primitive) game.ActivatedAbility {
	return game.ActivatedAbility{
		Text: text, ManaCost: manaCost, AdditionalCosts: additional, ZoneOfFunction: zone.Battlefield,
		Content: game.Mode{Targets: targets, Sequence: []game.Instruction{{Primitive: prim}}}.Ability(),
	}
}

func redirectTestCard(name string, cardTypes []types.Card, power, toughness int, abilities ...game.ActivatedAbility) *game.CardDef {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: cardTypes, Colors: []color.Color{color.White}, ActivatedAbilities: abilities,
	}}
	if power > 0 || toughness > 0 {
		def.Power, def.Toughness = opt.Val(game.PT{Value: power}), opt.Val(game.PT{Value: toughness})
	}
	return def
}

func redirectCreatureTarget(constraint string, controller game.ControllerRelation, chooser game.TargetChooser) game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Constraint: constraint, Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: controller}),
		Chooser:   chooser,
	}
}

func boltOnStack(g *game.Game, controller game.PlayerID, target game.Target) id.ID {
	card := addEffectSpellToStack(g, controller, game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)},
		[]game.Target{target})
	def := g.CardInstances[card].Def
	def.Name, def.Types, def.Colors = "Lightning Bolt", []types.Card{types.Instant}, []color.Color{color.Red}
	return card
}

func activateRedirect(t *testing.T, engine *Engine, g *game.Game, player game.PlayerID, source *game.Permanent, ability int, targets []game.Target, agents [game.NumPlayers]PlayerAgent) {
	t.Helper()
	setMainPhasePriority(g, player)
	if !engine.applyActionWithChoices(g, player, action.ActivateAbility(source.ObjectID, ability, targets, 0), agents, &TurnLog{}) {
		t.Fatalf("could not activate %s", permanentEffectiveName(g, source))
	}
}

// TestBeaconOfDestinyRedirectsBoltToItself plays Beacon of Destiny: tapping it
// with a Bolt aimed at its controller on the stack chooses the Bolt, whose 3
// damage is then dealt to the 1/3 Beacon instead, which dies, and the player
// loses no life.
func TestBeaconOfDestinyRedirectsBoltToItself(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	beacon := addCombatPermanent(g, game.Player2, redirectTestCard("Beacon of Destiny", []types.Card{types.Creature}, 1, 3,
		redirectAbility("{T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to this creature instead.",
			opt.V[cost.Mana]{}, cost.Tap, nil, game.PreventDamage{
				Player: game.ControllerReference(), All: true, OneShot: true,
				Redirect: game.DamageRedirectToObject, RedirectObject: game.SourcePermanentReference(),
			})))
	bolt := boltOnStack(g, game.Player1, game.PlayerTarget(game.Player2))
	agents := [game.NumPlayers]PlayerAgent{game.Player2: sourceChoiceAgent(t, bolt, nil)}

	activateRedirect(t, engine, g, game.Player2, beacon, 0, nil, agents)
	resolveStackWithTriggers(engine, g, agents)
	if g.Players[game.Player2].Life != 40 || beacon.MarkedDamage != 3 || len(g.PreventionShields) != 0 {
		t.Fatalf("after Bolt: life %d, Beacon damage %d, shields %+v; want 40, 3 and the shield used",
			g.Players[game.Player2].Life, beacon.MarkedDamage, g.PreventionShields)
	}
	engine.applyStateBasedActions(g)
	if _, ok := g.PermanentByID(beacon.ObjectID); ok {
		t.Fatal("Beacon of Destiny survived 3 redirected damage")
	}
}

// TestJadeMonolithRedirectsDamageToTargetCreatureToController plays Jade
// Monolith: an opponent's Bolt at the controller's creature is dealt to the
// controller instead, and the creature survives.
func TestJadeMonolithRedirectsDamageToTargetCreatureToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := bandingTestCreature(g, game.Player2, 2, 2)
	monolith := addCombatPermanent(g, game.Player2, redirectTestCard("Jade Monolith", []types.Card{types.Artifact}, 0, 0,
		redirectAbility("{1}: The next time a source of your choice would deal damage to target creature this turn, that source deals that damage to you instead.",
			opt.Val(cost.Mana{cost.O(1)}), nil,
			[]game.TargetSpec{redirectCreatureTarget("creature", game.ControllerAny, game.TargetChooserController)},
			game.PreventDamage{
				Object: game.TargetPermanentReference(0), All: true, OneShot: true, Redirect: game.DamageRedirectToController,
			})))
	bolt := boltOnStack(g, game.Player1, game.PermanentTarget(creature.ObjectID))
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	target := game.PermanentTarget(creature.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: redirectChoiceAgent(t, bolt, target)}

	activateRedirect(t, engine, g, game.Player2, monolith, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)
	if creature.MarkedDamage != 0 || g.Players[game.Player2].Life != 37 || len(g.PreventionShields) != 0 {
		t.Fatalf("after Bolt: creature damage %d, life %d, shields %+v; want 0, 37 and the shield used",
			creature.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields)
	}
}

// TestReflectDamageSendsLifelinkCombatDamageToItsController plays Reflect
// Damage: the chosen lifelink attacker's combat damage is dealt to its own
// controller instead, as combat damage, and lifelink still gains that life.
func TestReflectDamageSendsLifelinkCombatDamageToItsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := bandingTestCreature(g, game.Player1, 3, 3, game.Lifelink)
	agent := sourceChoiceAgent(t, attacker.CardInstanceID, nil)
	createOneShotShield(t, engine, g, addPreventionSourceSpell(g, game.Player2, "Reflect Damage"), game.PreventDamage{
		Global: true, All: true, OneShot: true, Redirect: game.DamageRedirectToDamageSourceController,
	}, agent)

	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}
	resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
	if g.Players[game.Player2].Life != 40 || len(g.PreventionShields) != 0 {
		t.Fatalf("defender life %d, shields %+v; want 40 and the shield used", g.Players[game.Player2].Life, g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.SourceObjectID == attacker.ObjectID && event.Player == game.Player1 &&
			event.DamageRecipient == game.DamageRecipientPlayer && event.Amount == 3 && event.CombatDamage
	})
	assertEvent(t, g.Events, game.EventLifeGained, func(event game.Event) bool {
		return event.Player == game.Player1 && event.Amount == 3
	})
	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("attacker's controller life = %d, want 40 after losing and gaining 3", g.Players[game.Player1].Life)
	}
}

// TestGeneralsRegaliaRedirectsToTargetCreatureOnlyWhileItRemains plays
// General's Regalia: the chosen source's damage to the controller is dealt to
// the target creature instead; with that creature gone, the damage is dealt to
// the controller and the shield remains unused (CR 614.9, 609.7b).
func TestGeneralsRegaliaRedirectsToTargetCreatureOnlyWhileItRemains(t *testing.T) {
	for _, gone := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		guard := bandingTestCreature(g, game.Player2, 2, 4)
		source := redCreature(g, game.Player1, 3, 3)
		regalia := addCombatPermanent(g, game.Player2, redirectTestCard("General's Regalia", []types.Card{types.Artifact}, 0, 0,
			redirectAbility("{3}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature you control instead.",
				opt.Val(cost.Mana{cost.O(3)}), nil,
				[]game.TargetSpec{redirectCreatureTarget("creature you control", game.ControllerYou, game.TargetChooserController)},
				game.PreventDamage{
					Player: game.ControllerReference(), All: true, OneShot: true,
					Redirect: game.DamageRedirectToObject, RedirectObject: game.TargetPermanentReference(0),
				})))
		g.Players[game.Player2].ManaPool.Add(mana.C, 3)
		target := game.PermanentTarget(guard.ObjectID)
		agents := [game.NumPlayers]PlayerAgent{game.Player2: redirectChoiceAgent(t, source.CardInstanceID, target)}
		activateRedirect(t, engine, g, game.Player2, regalia, 0, []game.Target{target}, agents)
		resolveStackWithTriggers(engine, g, agents)
		if len(g.PreventionShields) != 1 || g.PreventionShields[0].RedirectPermanentID != guard.ObjectID {
			t.Fatalf("shields = %+v, want one redirecting to the target creature", g.PreventionShields)
		}
		if gone {
			removePermanentFromBattlefield(g, guard.ObjectID)
		}

		dealDamageFrom(g, source, game.Player2, 3)
		wantLife, wantDamage, wantShields := 40, 3, 0
		if gone {
			wantLife, wantDamage, wantShields = 37, 0, 1
		}
		if g.Players[game.Player2].Life != wantLife || guard.MarkedDamage != wantDamage || len(g.PreventionShields) != wantShields {
			t.Fatalf("gone %v: life %d, creature damage %d, shields %+v; want %d, %d and %d shields",
				gone, g.Players[game.Player2].Life, guard.MarkedDamage, g.PreventionShields, wantLife, wantDamage, wantShields)
		}
	}
}

// TestNovaPentacleRedirectsToOpponentChosenCreature plays Nova Pentacle: the
// opponent chooses the target creature as the ability is activated, the
// controller chooses the source on resolution, and that source's damage to
// the controller is dealt to the opponent's chosen creature instead.
func TestNovaPentacleRedirectsToOpponentChosenCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pentacle := addCombatPermanent(g, game.Player1, redirectTestCard("Nova Pentacle", []types.Card{types.Artifact}, 0, 0,
		redirectAbility("{3}, {T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature of an opponent's choice instead.",
			opt.Val(cost.Mana{cost.O(3)}), cost.Tap,
			[]game.TargetSpec{redirectCreatureTarget("creature of an opponent's choice", game.ControllerAny, game.TargetChooserOpponent)},
			game.PreventDamage{
				Player: game.ControllerReference(), All: true, OneShot: true,
				Redirect: game.DamageRedirectToObject, RedirectObject: game.TargetPermanentReference(0),
			})))
	placeholder := bandingTestCreature(g, game.Player1, 2, 5)
	chosen := bandingTestCreature(g, game.Player1, 2, 5)
	source := redCreature(g, game.Player2, 3, 3)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: sourceChoiceAgent(t, source.CardInstanceID, nil),
		game.Player2: redirectChoiceAgent(t, source.CardInstanceID, game.PermanentTarget(chosen.ObjectID)),
	}

	activateRedirect(t, engine, g, game.Player1, pentacle, 0, []game.Target{game.PermanentTarget(placeholder.ObjectID)}, agents)
	if obj, ok := g.Stack.Peek(); !ok || len(obj.Targets) != 1 || obj.Targets[0] != game.PermanentTarget(chosen.ObjectID) {
		t.Fatalf("stack top = %+v, want the opponent's chosen creature targeted", obj)
	}
	resolveStackWithTriggers(engine, g, agents)
	dealDamageFrom(g, source, game.Player1, 3)
	if g.Players[game.Player1].Life != 40 || chosen.MarkedDamage != 3 || placeholder.MarkedDamage != 0 || len(g.PreventionShields) != 0 {
		t.Fatalf("life %d, chosen damage %d, other damage %d, shields %+v; want 40, 3, 0 and the shield used",
			g.Players[game.Player1].Life, chosen.MarkedDamage, placeholder.MarkedDamage, g.PreventionShields)
	}
}

// TestOpalEyeRedirectedDamageMeetsItsPreventionShield plays Opal-Eye, Konda's
// Yojimbo: its redirection takes the chosen source's damage to another player,
// and its own "prevent the next 1" shield then applies to the redirected
// damage (CR 616.2).
func TestOpalEyeRedirectedDamageMeetsItsPreventionShield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	opal := addCombatPermanent(g, game.Player2, redirectTestCard("Opal-Eye, Konda's Yojimbo", []types.Card{types.Creature}, 1, 4,
		redirectAbility("{T}: The next time a source of your choice would deal damage this turn, that damage is dealt to Opal-Eye instead.",
			opt.V[cost.Mana]{}, cost.Tap, nil, game.PreventDamage{
				Global: true, All: true, OneShot: true,
				Redirect: game.DamageRedirectToObject, RedirectObject: game.SourcePermanentReference(),
			}),
		redirectAbility("{1}{W}: Prevent the next 1 damage that would be dealt to Opal-Eye this turn.",
			opt.Val(cost.Mana{cost.O(1), cost.W}), nil, nil, game.PreventDamage{
				Amount: game.Fixed(1), Object: game.SourcePermanentReference(),
			})))
	source := redCreature(g, game.Player1, 3, 3)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: sourceChoiceAgent(t, source.CardInstanceID, nil)}
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)

	activateRedirect(t, engine, g, game.Player2, opal, 0, nil, agents)
	resolveStackWithTriggers(engine, g, agents)
	activateRedirect(t, engine, g, game.Player2, opal, 1, nil, agents)
	resolveStackWithTriggers(engine, g, agents)
	if len(g.PreventionShields) != 2 {
		t.Fatalf("shields = %+v, want the redirection and the prevent-1 shield", g.PreventionShields)
	}

	dealDamageFrom(g, source, game.Player3, 3)
	if g.Players[game.Player3].Life != 40 || opal.MarkedDamage != 2 || len(g.PreventionShields) != 0 {
		t.Fatalf("other player life %d, Opal-Eye damage %d, shields %+v; want 40, 2 and both shields used",
			g.Players[game.Player3].Life, opal.MarkedDamage, g.PreventionShields)
	}
}

// TestShamanEnKorTakesCombatDamageAimedAtTargetCreature plays Shaman
// en-Kor's second ability: the chosen attacker's combat damage to the target
// blocker is dealt to the Shaman instead, still as combat damage.
func TestShamanEnKorTakesCombatDamageAimedAtTargetCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shaman := addCombatPermanent(g, game.Player2, redirectTestCard("Shaman en-Kor", []types.Card{types.Creature}, 1, 2,
		redirectAbility("{1}{W}: The next time a source of your choice would deal damage to target creature this turn, that damage is dealt to this creature instead.",
			opt.Val(cost.Mana{cost.O(1), cost.W}), nil,
			[]game.TargetSpec{redirectCreatureTarget("creature", game.ControllerAny, game.TargetChooserController)},
			game.PreventDamage{
				Object: game.TargetPermanentReference(0), All: true, OneShot: true,
				Redirect: game.DamageRedirectToObject, RedirectObject: game.SourcePermanentReference(),
			})))
	attacker := bandingTestCreature(g, game.Player1, 3, 3)
	blocker := bandingTestCreature(g, game.Player2, 2, 2)
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	target := game.PermanentTarget(blocker.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: redirectChoiceAgent(t, attacker.CardInstanceID, target)}
	activateRedirect(t, engine, g, game.Player2, shaman, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)

	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}
	resolveCombatWithChoices(engine, g, agents)
	if blocker.MarkedDamage != 0 || shaman.MarkedDamage != 3 || attacker.MarkedDamage != 2 || len(g.PreventionShields) != 0 {
		t.Fatalf("blocker %d, Shaman %d, attacker %d, shields %+v; want 0, 3, 2 and the shield used",
			blocker.MarkedDamage, shaman.MarkedDamage, attacker.MarkedDamage, g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.SourceObjectID == attacker.ObjectID && event.PermanentID == shaman.ObjectID && event.CombatDamage
	})
}

// TestEyeForAnEyeDealsMatchingDamageSimultaneously plays Eye for an Eye: the
// chosen attacker's 4 combat damage is still dealt to the controller, and in
// the same event Eye for an Eye deals 4 damage to the attacker's controller.
func TestEyeForAnEyeDealsMatchingDamageSimultaneously(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := bandingTestCreature(g, game.Player1, 4, 4)
	eye := addPreventionSourceSpell(g, game.Player2, "Eye for an Eye")
	agent := sourceChoiceAgent(t, attacker.CardInstanceID, nil)
	createOneShotShield(t, engine, g, eye, game.PreventDamage{
		Player: game.ControllerReference(), All: true, OneShot: true,
		Redirect: game.DamageRedirectAlsoToDamageSourceController,
	}, agent)

	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}
	resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
	if g.Players[game.Player2].Life != 36 || g.Players[game.Player1].Life != 36 || len(g.PreventionShields) != 0 {
		t.Fatalf("defender life %d, attacker's controller life %d, shields %+v; want 36, 36 and the shield used",
			g.Players[game.Player2].Life, g.Players[game.Player1].Life, g.PreventionShields)
	}
	var combat, reflected game.Event
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt {
			continue
		}
		switch {
		case event.SourceObjectID == attacker.ObjectID && event.Player == game.Player2:
			combat = event
		case event.SourceID == eye.SourceID && event.Player == game.Player1:
			reflected = event
		}
	}
	if !combat.CombatDamage || combat.Amount != 4 || reflected.Amount != 4 || reflected.CombatDamage ||
		combat.SimultaneousID == 0 || combat.SimultaneousID != reflected.SimultaneousID {
		t.Fatalf("combat damage %+v, Eye for an Eye damage %+v; want 4 combat and 4 noncombat in one simultaneous event", combat, reflected)
	}
}

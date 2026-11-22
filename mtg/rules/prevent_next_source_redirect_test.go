package rules

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/cards/p"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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

// Source review: #4084, pinned Oracle 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Warrior, Outrider, Nomads, Lancers, and Spirit en-Kor share the {0} ability
// for the next 1 damage to this creature, redirected to a creature you control
// this turn. Exact printing rulings: Scryfall IDs afcf1e51-5e9a-4c86-8a14-97d9e919c320,
// de9b2506-ead4-49ff-a733-66f3f13dbe17, 48ed076a-909f-4ef0-b85e-a1935b1c2827,
// e7cd99c2-6d4a-48e8-8848-6fdc0788525d, 8c57ca87-75b0-4058-acd8-5394545d0846.
// These allow repeated activation and self-targeting; the damage retains its
// source and combat status. Outrider's rulings specify choosing the source
// during simultaneous damage, and no redirect to a departed or ineligible
// destination. Full records: .cardwork/sources/en-kor/.
// CR 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 601.2c, 602.2a-b, 608.2b govern choice and validation of the ability target;
// 120.1, 120.4b-c, 614.1a, 614.5, 614.9 and 616.1-2 govern damage identity,
// eligible damage recipients, and order/rechecking of replacements.
func fixedRedirectFixture(t *testing.T, g *game.Game, amount int) *game.Permanent {
	t.Helper()
	def := redirectTestCard("Fixed redirect fixture", []types.Card{types.Creature}, 0, 10,
		redirectAbility("Redirect a fixed amount of damage.", opt.V[cost.Mana]{}, nil,
			[]game.TargetSpec{redirectCreatureTarget("creature you control", game.ControllerYou, game.TargetChooserController)},
			game.PreventDamage{
				Amount: game.Fixed(amount), Object: game.SourcePermanentReference(),
				Redirect: game.DamageRedirectToObject, RedirectObject: game.TargetPermanentReference(0),
			}))
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("fixed-amount definition rejected: %+v", issues)
	}
	return addCombatPermanent(g, game.Player2, def)
}

func installFixedRedirect(t *testing.T, engine *Engine, g *game.Game, from, to *game.Permanent) {
	t.Helper()
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &controlScript{
		choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind == game.ChoicePreventionAllocation {
				t.Fatal("shield installation asked for a damage source")
			}
			return request.DefaultSelection
		},
	}}
	activateRedirect(t, engine, g, game.Player2, from, 0, []game.Target{game.PermanentTarget(to.ObjectID)}, agents)
	resolveStackWithTriggers(engine, g, agents)
}

func TestFixedDamageRedirectSplitsAndExhausts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	from := fixedRedirectFixture(t, g, 1)
	to := bandingTestCreature(g, game.Player2, 0, 10)
	installFixedRedirect(t, engine, g, from, to)
	source := bandingTestCreature(g, game.Player1, 3, 10, game.Lifelink)

	resolvePermanentDamage(engine, g, source, game.Damage{
		Amount: game.Fixed(3), Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)), CantBePrevented: true,
	}, game.PermanentTarget(from.ObjectID))
	if from.MarkedDamage != 2 || to.MarkedDamage != 1 || g.Players[game.Player1].Life != 43 || len(g.PreventionShields) != 0 {
		t.Fatalf("marked damage %d/%d, life %d, shields %+v; want 2/1, 43, none",
			from.MarkedDamage, to.MarkedDamage, g.Players[game.Player1].Life, g.PreventionShields)
	}
	var original, redirected game.Event
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt || event.SourceObjectID != source.ObjectID {
			continue
		}
		switch event.PermanentID {
		case from.ObjectID:
			original = event
		case to.ObjectID:
			redirected = event
		}
	}
	if original.Amount != 2 || redirected.Amount != 1 || original.CombatDamage || redirected.CombatDamage ||
		original.SimultaneousID == 0 || original.SimultaneousID != redirected.SimultaneousID {
		t.Fatalf("damage facts %+v / %+v; want original source, split 2/1 simultaneously", original, redirected)
	}
	assertNoEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.SourceObjectID == source.ObjectID
	})
	boltOnStack(g, game.Player1, game.PermanentTarget(from.ObjectID))
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if from.MarkedDamage != 5 || to.MarkedDamage != 1 {
		t.Fatalf("spent shield affected subsequent damage: %d/%d", from.MarkedDamage, to.MarkedDamage)
	}
}

func TestFixedDamageRedirectChoosesSimultaneousCombatSources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	from := fixedRedirectFixture(t, g, 1)
	to := bandingTestCreature(g, game.Player2, 0, 10)
	installFixedRedirect(t, engine, g, from, to)
	secondTarget := bandingTestCreature(g, game.Player2, 0, 10)
	installFixedRedirect(t, engine, g, from, secondTarget)
	first := bandingTestCreature(g, game.Player1, 3, 10)
	second := bandingTestCreature(g, game.Player1, 1, 10)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: from.ObjectID, Target: game.AttackTarget{Player: game.Player1}}},
		Blockers:  []game.BlockDeclaration{{Blocker: first.ObjectID, Blocking: from.ObjectID}, {Blocker: second.ObjectID, Blocking: from.ObjectID}},
	}
	allocations := 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceReplacement {
			for _, option := range request.Options {
				if strings.Contains(option.Label, fmt.Sprintf("[%d]", secondTarget.ObjectID)) {
					return []int{option.Index}
				}
			}
			t.Fatalf("second shield's destination was not identified: %+v", request.Options)
		}
		if request.Kind != game.ChoicePreventionAllocation {
			return request.DefaultSelection
		}
		if request.Player != game.Player2 || len(request.Options) != 2 || !strings.Contains(request.Prompt, "redirect") ||
			from.MarkedDamage != 0 || to.MarkedDamage != 0 || secondTarget.MarkedDamage != 0 {
			t.Fatalf("unexpected allocation request or damage already committed: %+v", request)
		}
		allocations++
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == second.CardInstanceID {
				return []int{option.Index}
			}
		}
		t.Fatal("second combat source was not offered")
		return nil
	}}
	resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
	if allocations != 1 || from.MarkedDamage != 2 || to.MarkedDamage != 1 || secondTarget.MarkedDamage != 1 {
		t.Fatalf("allocations %d, marked damage %d/%d/%d; want 1, 2/1/1", allocations, from.MarkedDamage, to.MarkedDamage, secondTarget.MarkedDamage)
	}
	for _, want := range []struct {
		source, recipient *game.Permanent
		amount            int
	}{{first, from, 2}, {second, secondTarget, 1}, {first, to, 1}} {
		assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
			return event.PermanentID == want.recipient.ObjectID && event.SourceObjectID == want.source.ObjectID &&
				event.CombatDamage && event.Amount == want.amount
		})
	}
	assertNoEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.SourceObjectID == second.ObjectID && event.PermanentID == from.ObjectID
	})
}

func TestFixedDamageRedirectRepeatedActivationsDealOneOccurrencePerRecipient(t *testing.T) {
	for _, self := range []bool{false, true} {
		t.Run(fmt.Sprintf("self=%v", self), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			from := fixedRedirectFixture(t, g, 1)
			to := from
			if !self {
				to = bandingTestCreature(g, game.Player2, 0, 10)
			}
			installFixedRedirect(t, engine, g, from, to)
			installFixedRedirect(t, engine, g, from, to)
			bolt := boltOnStack(g, game.Player1, game.PermanentTarget(from.ObjectID))
			resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
			wantFrom, wantTo := 1, 2
			if self {
				wantFrom, wantTo = 3, 3
			}
			if from.MarkedDamage != wantFrom || to.MarkedDamage != wantTo || len(g.PreventionShields) != 0 {
				t.Fatalf("damage %d/%d, shields %+v; want %d/%d and both shields used",
					from.MarkedDamage, to.MarkedDamage, g.PreventionShields, wantFrom, wantTo)
			}
			occurrences := 0
			for _, event := range g.Events {
				if event.Kind == game.EventDamageDealt && event.SourceID == bolt && event.PermanentID == to.ObjectID {
					occurrences++
					if event.Amount != wantTo {
						t.Errorf("damage occurrence has amount %d, want %d", event.Amount, wantTo)
					}
				}
			}
			if occurrences != 1 {
				t.Fatalf("damage occurrences at destination = %d, want 1", occurrences)
			}
		})
	}
}

// Source review: pinned Oracle printings Pyroclasm 4391b0af-2f26-4a45-9e2a-5bd8e9838107
// (2 damage to each creature; exact printing's Scryfall rulings list empty)
// and Boros Reckoner 55a2f767-5928-44c5-91b7-2300bade458f (when dealt damage,
// deals that much to any target; exact printing's rulings confirm one trigger
// and one target for simultaneous damage, even when the damage is lethal).
// https://api.scryfall.com/cards/{printing}/rulings
// CR 2026-11-06 §§120.4b-c, 603.2c and 614.9: capture damage triggers from the
// modified damage event; redirection does not create a new damage source.
func TestFixedDamageRedirectJoinsPyroclasmDamageBeforeReckonerTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	from := fixedRedirectFixture(t, g, 1)
	reckoner := addCombatPermanent(g, game.Player2, b.BorosReckoner())
	installFixedRedirect(t, engine, g, from, reckoner)
	setMainPhasePriority(g, game.Player1)
	spell := addCardToHand(g, game.Player1, p.Pyroclasm())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agents := [game.NumPlayers]PlayerAgent{}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spell, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("could not cast Pyroclasm")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if from.MarkedDamage != 1 || reckoner.MarkedDamage != 3 {
		t.Fatalf("Pyroclasm dealt %d/%d damage; want 1/3", from.MarkedDamage, reckoner.MarkedDamage)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	stack := g.Stack.Objects()
	if len(stack) != 1 {
		t.Fatalf("Boros Reckoner put %d triggers on the stack, want one", len(stack))
	}
	trigger := stack[0]
	if trigger.Kind != game.StackTriggeredAbility || trigger.SourceID != reckoner.ObjectID ||
		trigger.TriggerEvent.SourceID != spell || trigger.TriggerEvent.Amount != 3 ||
		len(trigger.Targets) != 1 || trigger.Targets[0] != game.PlayerTarget(game.Player1) {
		t.Fatalf("stack kind %v, source %d, triggering source %d, damage %d, targets %v; want Reckoner's one trigger for 3 from Pyroclasm targeting player 1",
			trigger.Kind, trigger.SourceID, trigger.TriggerEvent.SourceID, trigger.TriggerEvent.Amount, trigger.Targets)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if life := g.Players[game.Player1].Life; life != 37 {
		t.Fatalf("target player's life = %d, want 37", life)
	}
}

func TestFixedDamageRedirectCoalescesAcrossAssignmentsWithoutLosingAttribution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first, second := fixedRedirectFixture(t, g, 1), fixedRedirectFixture(t, g, 1)
	dest := bandingTestCreature(g, game.Player2, 0, 10)
	otherDest := bandingTestCreature(g, game.Player2, 0, 10)
	installFixedRedirect(t, engine, g, first, dest)
	installFixedRedirect(t, engine, g, second, dest)
	source := bandingTestCreature(g, game.Player1, 0, 10)
	otherSource := bandingTestCreature(g, game.Player1, 0, 10)
	batchID := g.IDGen.Next()
	results := resolveDamageBatch(g, []damageEvent{
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, permanent: first, amount: 1},
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, permanent: second, amount: 1},
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, permanent: dest, amount: 2},
		{sourceID: otherSource.CardInstanceID, sourceObjectID: otherSource.ObjectID, controller: game.Player1, permanent: dest, amount: 4},
		{sourceID: source.CardInstanceID, sourceObjectID: source.ObjectID, controller: game.Player1, permanent: otherDest, amount: 2},
	}, batchID, false, nil)
	for index, want := range []struct {
		amount     int
		redirected bool
	}{{1, true}, {1, true}, {2, false}, {4, false}, {2, false}} {
		if got := results[index]; got.amount != want.amount || got.redirected != want.redirected {
			t.Errorf("assignment %d returned amount %d redirected %v; want %d/%v",
				index, got.amount, got.redirected, want.amount, want.redirected)
		}
	}
	if first.MarkedDamage != 0 || second.MarkedDamage != 0 || dest.MarkedDamage != 8 || otherDest.MarkedDamage != 2 {
		t.Fatalf("marked damage %d/%d/%d/%d, want 0/0/8/2", first.MarkedDamage, second.MarkedDamage, dest.MarkedDamage, otherDest.MarkedDamage)
	}
	type occurrence struct{ source, recipient id.ID }
	want := map[occurrence]int{
		{source.ObjectID, otherDest.ObjectID}: 2,
		{source.ObjectID, dest.ObjectID}:      4,
		{otherSource.ObjectID, dest.ObjectID}: 4,
	}
	for _, event := range g.Events {
		if event.Kind != game.EventDamageDealt {
			continue
		}
		key := occurrence{event.SourceObjectID, event.PermanentID}
		if event.SimultaneousID != batchID || event.Amount != want[key] {
			t.Errorf("unexpected occurrence: source %d recipient %d amount %d batch %d",
				key.source, key.recipient, event.Amount, event.SimultaneousID)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing damage occurrences: %v", want)
	}
}

func TestFixedDamageRedirectRechecksDestination(t *testing.T) {
	for _, change := range []string{"left battlefield", "lost creature type", "became a planeswalker"} {
		t.Run(change, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			from := fixedRedirectFixture(t, g, 1)
			to := bandingTestCreature(g, game.Player2, 0, 10)
			installFixedRedirect(t, engine, g, from, to)
			wantFrom, wantShields := 3, 1
			switch change {
			case "left battlefield":
				removePermanentFromBattlefield(g, to.ObjectID)
			default:
				newType := types.Artifact
				if change == "became a planeswalker" {
					newType = types.Planeswalker
					to.Counters.Add(counter.Loyalty, 4)
					wantFrom, wantShields = 2, 0
				}
				if !applyTypedContinuousEffects(g, &game.StackObject{
					Kind: game.StackActivatedAbility, SourceID: from.ObjectID, Controller: game.Player2,
				}, to, []game.ContinuousEffect{{
					Layer: game.LayerType, RemoveTypes: []types.Card{types.Creature}, AddTypes: []types.Card{newType},
				}}, game.DurationUntilEndOfTurn) || permanentHasType(g, to, types.Creature) || !permanentHasType(g, to, newType) {
					t.Fatal("could not change destination's type")
				}
			}
			boltOnStack(g, game.Player1, game.PermanentTarget(from.ObjectID))
			resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
			if from.MarkedDamage != wantFrom || to.MarkedDamage != 0 || len(g.PreventionShields) != wantShields {
				t.Fatalf("damage %d/%d, shields %+v; want %d/0 and %d shields", from.MarkedDamage, to.MarkedDamage, g.PreventionShields, wantFrom, wantShields)
			}
			if change == "became a planeswalker" && to.Counters.Get(counter.Loyalty) != 3 {
				t.Fatal("eligible planeswalker destination did not receive 1 damage")
			}
		})
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: recipient scopes of one-shot "a source of your choice"
// prevention shields (#3805). Comprehensive Rules effective 2026-11-06 (local
// verified copy .cardwork/MagicCompRules-20261106.txt):
//   - 609.7a/609.7b: the source is chosen as the effect is created; a shield
//     that prevents nothing is not used up.
//   - 615.1, 615.6: a prevention shield applies to the next damage event it
//     matches; "the next time ... would deal damage" prevents all of that one
//     event's damage it covers, so the chosen source's simultaneous damage to
//     every in-scope recipient is prevented together and its damage to other
//     recipients in the same event is still dealt.
//   - 616.1: each affected player or affected object's controller applies the
//     shield to their own damage, in APNAP order.
//   - 702.19b: trample's lethal assignment ignores prevention, so the excess
//     assigned to the player is a separate recipient in the same event.
// Scryfall Oracle (pinned corpus 6da9ced0...): Charm Peddler c425fe26 and
// Samite Blessing a524ae69 "... would deal damage to target creature this
// turn, prevent that damage."; Kithkin Armor 35e2730e "Sacrifice this Aura:
// The next time a source of your choice would deal damage to enchanted
// creature this turn, prevent that damage."; Circle of Despair 4ad57fbf,
// Martyr's Cause 72d1c789 and Sanctum Guardian eb7fe094 "... to any target
// ..."; Pilgrim of Justice 959b672e "{W}, Sacrifice this creature: The next
// time a red source of your choice would deal damage this turn, prevent that
// damage." (Pilgrim of Virtue e2f73743: black); Shadowbane 9e9ec33b "... would
// deal damage to you and/or creatures you control this turn, prevent that
// damage." with its 2008-08-01 ruling that only damage to you or a creature
// you control is prevented. Riders (Shadowbane's life gain), redirection and
// static group prevention are outside this scope.

// tramplingCombat has attacker attack Player2 and blocker block it.
func tramplingCombat(attacker, blocker *game.Permanent) *game.CombatState {
	return &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}
}

func resolveCombatWithChoices(engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	engine.setReplacementChoiceContext(g, agents, &TurnLog{})
	defer g.ClearChoiceContext()
	engine.resolveCombatDamage(g, &TurnLog{})
}

// resolvePermanentDamage resolves one damage instruction of an ability of
// source, so source deals the damage as a single damage event.
func resolvePermanentDamage(engine *Engine, g *game.Game, source *game.Permanent, prim game.Damage, targets ...game.Target) {
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: source.ObjectID,
		Controller: source.Controller, Targets: targets,
	}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: prim}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func redCreature(g *game.Game, controller game.PlayerID, power, toughness int, keywords ...game.Keyword) *game.Permanent {
	permanent := bandingTestCreature(g, controller, power, toughness, keywords...)
	g.CardInstances[permanent.CardInstanceID].Def.Colors = []color.Color{color.Red}
	return permanent
}

// TestOneShotTargetCreatureShieldLeavesTrampleExcess plays the Charm Peddler /
// Samite Blessing shape: the shield on the target blocker prevents the chosen
// trampler's combat damage to it, the excess trampled over to the defending
// player is dealt in the same event, the shield is used, and the trampler's
// next damage to the blocker is not prevented.
func TestOneShotTargetCreatureShieldLeavesTrampleExcess(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := bandingTestCreature(g, game.Player1, 5, 5, game.Trample)
	blocker := bandingTestCreature(g, game.Player2, 2, 2)
	_, ability := protectionShieldSource(g, game.Player2)
	ability.Targets = []game.Target{game.PermanentTarget(blocker.ObjectID)}
	agent := sourceChoiceAgent(t, attacker.CardInstanceID, nil)
	createOneShotShield(t, engine, g, ability, game.PreventDamage{
		Object: game.TargetPermanentReference(0), All: true, OneShot: true,
	}, agent)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].PermanentID != blocker.ObjectID {
		t.Fatalf("shields = %+v, want one on the target blocker", g.PreventionShields)
	}

	g.Combat = tramplingCombat(attacker, blocker)
	resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
	if blocker.MarkedDamage != 0 || g.Players[game.Player2].Life != 37 || len(g.PreventionShields) != 0 {
		t.Fatalf("combat: blocker damage %d, defender life %d, shields %+v; want 0, 37 and the shield used",
			blocker.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields)
	}
	if dealt := dealPermanentDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player1, blocker, 1, false); dealt != 1 {
		t.Fatalf("second damage to the blocker = %d, want 1", dealt)
	}
}

// TestOneShotEnchantedCreatureShieldAfterSacrificingAura plays the Kithkin
// Armor shape through a real activation: sacrificing the Aura as the cost
// leaves the shield on the creature it enchanted, which the chosen source's
// next damage then fails to damage.
func TestOneShotEnchantedCreatureShieldAfterSacrificingAura(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchanted := bandingTestCreature(g, game.Player2, 2, 2)
	other := bandingTestCreature(g, game.Player2, 2, 2)
	armor := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Kithkin Armor", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		Colors: []color.Color{color.White},
		StaticAbilities: []game.StaticAbility{game.EnchantStaticAbility(&game.TargetSpec{
			MinTargets: 1, MaxTargets: 1, Constraint: "creature", Allow: game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
		})},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "Sacrifice this Aura: The next time a source of your choice would deal damage to enchanted creature this turn, prevent that damage.",
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this Aura", Amount: 1}},
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PreventDamage{
				Object: game.SourceAttachedPermanentReference(), All: true, OneShot: true,
			}}}}.Ability(),
		}},
	}})
	armor.AttachedTo = opt.Val(enchanted.ObjectID)
	source := redCreature(g, game.Player1, 3, 3)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: sourceChoiceAgent(t, source.CardInstanceID, nil)}

	setMainPhasePriority(g, game.Player2)
	if !engine.applyActionWithChoices(g, game.Player2, action.ActivateAbility(armor.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
		t.Fatal("could not activate Kithkin Armor")
	}
	if _, ok := g.PermanentByID(armor.ObjectID); ok {
		t.Fatal("Kithkin Armor was not sacrificed as the cost")
	}
	resolveStackWithTriggers(engine, g, agents)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].PermanentID != enchanted.ObjectID {
		t.Fatalf("shields = %+v, want one on the formerly enchanted creature", g.PreventionShields)
	}
	if dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, other, 1, false); dealt != 1 {
		t.Fatalf("damage to another creature = %d, want 1", dealt)
	}
	if dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, enchanted, 3, false); dealt != 0 {
		t.Fatalf("chosen source's damage to the enchanted creature = %d, want 0", dealt)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want the shield used", g.PreventionShields)
	}
}

// TestOneShotAnyTargetShieldCoversOnlyItsRecipient plays the Circle of
// Despair / Martyr's Cause / Sanctum Guardian shape once with a player and
// once with a creature: the chosen trampler's combat damage is prevented only
// to the chosen target, and its damage to the other recipient is dealt.
func TestOneShotAnyTargetShieldCoversOnlyItsRecipient(t *testing.T) {
	for _, targetPlayer := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		attacker := bandingTestCreature(g, game.Player1, 5, 5, game.Trample)
		blocker := bandingTestCreature(g, game.Player2, 2, 2)
		_, ability := protectionShieldSource(g, game.Player2)
		ability.Targets = []game.Target{game.PermanentTarget(blocker.ObjectID)}
		wantBlocker, wantLife := 0, 37
		if targetPlayer {
			ability.Targets = []game.Target{game.PlayerTarget(game.Player2)}
			wantBlocker, wantLife = 2, 40
		}
		agent := sourceChoiceAgent(t, attacker.CardInstanceID, nil)
		createOneShotShield(t, engine, g, ability, game.PreventDamage{
			AnyTarget: game.AnyTargetDamageRecipient(0), All: true, OneShot: true,
		}, agent)

		g.Combat = tramplingCombat(attacker, blocker)
		resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
		if blocker.MarkedDamage != wantBlocker || g.Players[game.Player2].Life != wantLife || len(g.PreventionShields) != 0 {
			t.Fatalf("target player %v: blocker damage %d, life %d, shields %+v; want %d, %d and the shield used",
				targetPlayer, blocker.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields, wantBlocker, wantLife)
		}
	}
}

// TestOneShotUnrestrictedShieldPreventsWholeEvent plays the Pilgrim of
// Justice shape: the chosen red source's "each creature" damage to creatures
// of several players is all prevented as one use (each controller applying
// the shield to its own creature), its next damage event is dealt, and a
// second shield prevents the same source's trampling combat damage to both
// the blocker and the defending player.
func TestOneShotUnrestrictedShieldPreventsWholeEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := redCreature(g, game.Player1, 5, 5, game.Trample)
	creatures := []*game.Permanent{
		bandingTestCreature(g, game.Player1, 4, 4),
		bandingTestCreature(g, game.Player2, 4, 4),
		bandingTestCreature(g, game.Player3, 4, 4),
	}
	pilgrim := game.PreventDamage{Global: true, All: true, OneShot: true, SourceColors: []color.Color{color.Red}}
	agent := sourceChoiceAgent(t, source.CardInstanceID, nil)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, pilgrim, agent)

	eachCreature := game.Damage{Amount: game.Fixed(2), Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	}))}
	resolvePermanentDamage(engine, g, source, eachCreature)
	for _, creature := range creatures {
		if creature.MarkedDamage != 0 {
			t.Fatalf("creature of player %d marked %d, want the whole event prevented", creature.Controller, creature.MarkedDamage)
		}
	}
	if source.MarkedDamage != 0 || len(g.PreventionShields) != 0 {
		t.Fatalf("source damage %d, shields %+v; want 0 and the shield used once", source.MarkedDamage, g.PreventionShields)
	}
	resolvePermanentDamage(engine, g, source, eachCreature)
	if creatures[1].MarkedDamage != 2 {
		t.Fatalf("next event marked %d, want 2", creatures[1].MarkedDamage)
	}

	createOneShotShield(t, engine, g, ability, pilgrim, agent)
	blocker := creatures[1]
	g.Combat = tramplingCombat(source, blocker)
	resolveCombatWithChoices(engine, g, [game.NumPlayers]PlayerAgent{game.Player2: agent})
	if blocker.MarkedDamage != 2 || g.Players[game.Player2].Life != 40 || len(g.PreventionShields) != 0 {
		t.Fatalf("combat: blocker damage %d, life %d, shields %+v; want 2 (earlier), 40 and the shield used",
			blocker.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields)
	}
}

// TestOneShotPlayerAndCreaturesShieldChecksControlAtDamage plays the
// Shadowbane prevention shape: the chosen source's damage to the controller
// and to the creatures the controller controls as the damage is dealt
// (including one gained after the shield was created) is prevented as one
// use, while its damage to an opponent's creature and another player in the
// same event is dealt.
func TestOneShotPlayerAndCreaturesShieldChecksControlAtDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := redCreature(g, game.Player1, 3, 3)
	own := bandingTestCreature(g, game.Player2, 4, 4)
	stolen := bandingTestCreature(g, game.Player3, 4, 4)
	opponents := bandingTestCreature(g, game.Player3, 4, 4)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, game.PreventDamage{
		Player: game.ControllerReference(), AndControlledCreatures: true, All: true, OneShot: true,
	}, sourceChoiceAgent(t, source.CardInstanceID, nil))
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerControl, AffectedObjectID: stolen.ObjectID,
		NewController: opt.Val(game.Player2), Duration: game.DurationPermanent,
	})

	eachOfThem := game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0), EachTarget: true}
	resolvePermanentDamage(engine, g, source, eachOfThem,
		game.PlayerTarget(game.Player2), game.PermanentTarget(own.ObjectID), game.PermanentTarget(stolen.ObjectID),
		game.PermanentTarget(opponents.ObjectID), game.PlayerTarget(game.Player3))
	if g.Players[game.Player2].Life != 40 || own.MarkedDamage != 0 || stolen.MarkedDamage != 0 {
		t.Fatalf("in scope: life %d, own %d, gained %d; want 40, 0, 0", g.Players[game.Player2].Life, own.MarkedDamage, stolen.MarkedDamage)
	}
	if opponents.MarkedDamage != 2 || g.Players[game.Player3].Life != 38 {
		t.Fatalf("out of scope: opponent's creature %d, other player life %d; want 2, 38", opponents.MarkedDamage, g.Players[game.Player3].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want the shield used by the event", g.PreventionShields)
	}
	resolvePermanentDamage(engine, g, source, eachOfThem, game.PlayerTarget(game.Player2))
	if g.Players[game.Player2].Life != 38 {
		t.Fatalf("next event: life %d, want 38", g.Players[game.Player2].Life)
	}
}

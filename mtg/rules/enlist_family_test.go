package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func enlistFamilyAttack(t *testing.T, g *game.Game, attacker, enlisted *game.Permanent) *Engine {
	t.Helper()
	policy := &chosenTapPolicy{pick: pickPermanentOption(enlisted)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy

	engine := NewEngine(nil)
	declaration := action.DeclareAttackersWithBands([]game.AttackDeclaration{
		attackAt(attacker, game.AttackTarget{Player: game.Player2}),
	}, nil)
	if !engine.applyActionWithChoices(g, game.Player1, declaration, agents, &TurnLog{}) {
		t.Fatal("declaring the Enlist creature as an attacker failed")
	}
	if !enlisted.Tapped {
		t.Fatal("the chosen creature was not tapped to pay the Enlist cost")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("paying the Enlist cost did not create its reflexive trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size after paying Enlist = %d, want its one power trigger", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	return engine
}

func enlistFamilyTargetPlayer(player game.PlayerID) PlayerAgent {
	return enlistFamilyPlayerTargetAgent{player: player}
}

type enlistFamilyPlayerTargetAgent struct {
	simPassPolicy
	player game.PlayerID
}

func (a enlistFamilyPlayerTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			for _, target := range option.Targets {
				if target.Kind == game.TargetPlayer && target.PlayerID == a.player {
					return []int{option.Index}
				}
			}
		}
	}
	return request.DefaultSelection
}

// Source review: Oracle 278db664-4235-40f0-a926-3074a0e038c1, DMU printing
// 084d77df-a899-406f-9d79-e3fa3abeaebc; its Scryfall rulings URI
// https://api.scryfall.com/cards/084d77df-a899-406f-9d79-e3fa3abeaebc/rulings
// returned four WotC rulings dated 2022-09-09. They require the enlist choice
// after attackers tap, a legal untapped nonattacker, and a linked +X/+0 trigger.
// Official CR, effective 2026-11-06, https://magic.wizards.com/en/rules
// (PDF https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 302.6, 508.1a/f-g/j/m, 607.2h, 702.20a-b, and 702.154a-d. Vigilance
// leaves this attacker untapped while paying its separate enlist cost.
func TestEnlistFamilyBenalishFaithbonder(t *testing.T) {
	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.BenalishFaithbonder.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 3)

	enlistFamilyAttack(t, g, attacker, enlisted)

	if attacker.Tapped {
		t.Fatal("Vigilance tapped Benalish Faithbonder while it attacked")
	}
	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("Faithbonder power = %d, want 4 after enlisting the 3-power creature", got)
	}
}

// Source review: Oracle 3a7b1d1c-1724-4208-943b-650a5d773653, DMU printing
// c35bab2f-8464-4968-8c50-8807638d687b; its Scryfall rulings URI
// https://api.scryfall.com/cards/c35bab2f-8464-4968-8c50-8807638d687b/rulings
// returned five WotC rulings dated 2022-09-09, including that the death trigger
// uses the Berserker's last battlefield power. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 302.6, 508.1f-g/j/m, 603.2-603.3, 607.2h, 608.2h, 700.4, 702.154a-d,
// and 115.1a/120.3. The trigger therefore deals 4 after this 1/3 enlists a
// 3-power creature and then dies.
func TestEnlistFamilyBalduvianBerserker(t *testing.T) {
	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.BalduvianBerserker.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	engine := enlistFamilyAttack(t, g, attacker, enlisted)
	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("Berserker power before it dies = %d, want 4", got)
	}

	addEffectSpellToStack(g, game.Player2, game.Damage{
		Amount:    game.Fixed(4),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(attacker.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player1].Graveyard.Contains(attacker.CardInstanceID) {
		t.Fatal("lethal damage did not put Balduvian Berserker into its owner's graveyard")
	}

	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = enlistFamilyTargetPlayer(game.Player2)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Balduvian Berserker's dies ability did not trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size after the dies trigger = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("opponent life after the dies trigger = %d, want 36 from last-known power 4", got)
	}
}

// Source review: Oracle 84d67d87-2696-4c2f-9f10-29d5832bfbb6, DMU printing
// 554c422f-2a1b-47c0-9358-42b85cdbb84e; its Scryfall rulings URI
// https://api.scryfall.com/cards/554c422f-2a1b-47c0-9358-42b85cdbb84e/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 607.2h, 702.19a-b, 702.154a-d, and 510.1. Enlist adds
// power before combat damage; trample assigns lethal damage to the blocker and
// the excess to the defending player.
func TestEnlistFamilyCoalitionWarbrute(t *testing.T) {
	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.CoalitionWarbrute.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	engine := enlistFamilyAttack(t, g, attacker, enlisted)
	if got := effectivePower(g, attacker); got != 7 {
		t.Fatalf("Warbrute power after enlist = %d, want 7", got)
	}

	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	blocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: blocker.ObjectID, Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, blocks) {
		t.Fatal("Warbrute block declaration failed")
	}
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 35 {
		t.Fatalf("defending player life after trample = %d, want 35 after 5 damage to the 2-toughness blocker", got)
	}
}

// Source review: Oracle ad181ee7-756f-4002-9009-023ea6172991, DMU printing
// 25175622-7b68-4916-adf2-3bf2d82e8356; its Scryfall rulings URI
// https://api.scryfall.com/cards/25175622-7b68-4916-adf2-3bf2d82e8356/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 602.2, 607.2h, 608.2b, 701.8, 702.154a-d, and 118.3.
// After enlisting, the already-tapped attacker can pay {1} and sacrifice itself;
// its activated ability still resolves to destroy the targeted artifact.
func TestEnlistFamilyYavimayaSteelcrusher(t *testing.T) {
	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.YavimayaSteelcrusher.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine := enlistFamilyAttack(t, g, attacker, enlisted)
	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("Steelcrusher power after enlist = %d, want 4", got)
	}

	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Artifact",
		Types: []types.Card{types.Artifact},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(attacker.ObjectID, 0,
		[]game.Target{game.PermanentTarget(artifact.ObjectID)}, 0)) {
		t.Fatal("Steelcrusher could not pay {1} and sacrifice itself to target the artifact")
	}
	if !g.Players[game.Player1].Graveyard.Contains(attacker.CardInstanceID) {
		t.Fatal("Steelcrusher's sacrifice cost was not paid")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) {
		t.Fatal("Steelcrusher's ability did not destroy the targeted artifact")
	}
}

// Source review: Oracle c3b395b3-fea0-41d9-a3f2-6c90f89d8bdb, DMU printing
// 15cc59c7-6478-4f84-898a-7e0fed10f004; its Scryfall rulings URI
// https://api.scryfall.com/cards/15cc59c7-6478-4f84-898a-7e0fed10f004/rulings
// returned five WotC rulings dated 2022-09-09, including that reducing a
// blocker's power after it blocks does not unblock this creature. Official CR,
// effective 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 509.1b, 607.2h, 702.154a-d. The power restriction applies
// when blockers are declared; a legal block remains after the blocker shrinks.
func TestEnlistFamilyLinebreakerBaloth(t *testing.T) {
	t.Run("a two-power creature can block a vanilla creature", func(t *testing.T) {
		g := newEnlistCombatGame()
		vanilla := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		engine := NewEngine(nil)
		attack := action.DeclareAttackersWithBands([]game.AttackDeclaration{
			attackAt(vanilla, game.AttackTarget{Player: game.Player2}),
		}, nil)
		if !engine.applyAction(g, game.Player1, attack) {
			t.Fatal("baseline vanilla attack declaration failed")
		}
		g.Turn.Step = game.StepDeclareBlockers
		blocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
			Blocker: blocker.ObjectID, Blocking: vanilla.ObjectID,
		}}))
		if !engine.applyDeclareBlockers(g, game.Player2, blocks) {
			t.Fatal("the same two-power block was not legal without Linebreaker Baloth's restriction")
		}
	})

	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.LinebreakerBaloth.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	engine := enlistFamilyAttack(t, g, attacker, enlisted)
	if got := effectivePower(g, attacker); got != 5 {
		t.Fatalf("Linebreaker Baloth power after enlist = %d, want 5", got)
	}
	weak := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	strong := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	weakBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: weak.ObjectID, Blocking: attacker.ObjectID,
	}}))
	if engine.applyDeclareBlockers(g, game.Player2, weakBlock) {
		t.Fatal("Linebreaker Baloth was blocked by a creature with power 2")
	}
	strongBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: strong.ObjectID, Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, strongBlock) {
		t.Fatal("a creature with power 3 could not block Linebreaker Baloth")
	}

	strong.Counters.Add(counter.MinusOneMinusOne, 1)
	if got := effectivePower(g, strong); got != 2 {
		t.Fatalf("blocking creature power after shrinking = %d, want 2", got)
	}
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("defending player life after a blocker shrank = %d, want 40 because the block remains", got)
	}
}

// Source review: Oracle c45090bd-a589-45a0-8388-025b8e0b3487, DMU printing
// 57746e72-ba9f-4717-a4a2-6c6a039080c3; its Scryfall rulings URI
// https://api.scryfall.com/cards/57746e72-ba9f-4717-a4a2-6c6a039080c3/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 603.2-603.3/603.6, 607.2h, 701.7a, 111.1-111.3, and
// 702.154a-d. The cast creature's entry trigger creates a white 1/1 Soldier;
// on a later turn it can attack and enlist.
func TestEnlistFamilyArgivianCavalier(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, authored.ArgivianCavalier.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Argivian Cavalier for {2}{W} failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Argivian Cavalier's entry trigger did not trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	var cavalier *game.Permanent
	soldiers := 0
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			cavalier = permanent
		}
		if effectiveController(g, permanent) == game.Player1 && permanentHasSubtype(g, permanent, types.Sub("Soldier")) {
			soldiers++
			toughness, _ := effectiveToughness(g, permanent)
			if effectivePower(g, permanent) != 1 || toughness != 1 ||
				!slices.Contains(permanentEffectiveColors(g, permanent), color.White) {
				t.Fatal("Argivian Cavalier's token was not a 1/1 white Soldier")
			}
		}
	}
	if cavalier == nil {
		t.Fatal("Argivian Cavalier did not resolve onto the battlefield")
	}
	if soldiers != 1 {
		t.Fatalf("white Soldier tokens = %d, want 1", soldiers)
	}

	g.Turn.TurnNumber = 2
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	cavalier.SummoningSick = false
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	enlistFamilyAttack(t, g, cavalier, enlisted)
	if got := effectivePower(g, cavalier); got != 5 {
		t.Fatalf("Argivian Cavalier power after enlist = %d, want 5", got)
	}
}

// Source review: Oracle cbac75c6-6006-4025-b893-8828965b5f97, DMU printing
// 41c62c55-7b04-4987-9047-1322c842bb59; its Scryfall rulings URI
// https://api.scryfall.com/cards/41c62c55-7b04-4987-9047-1322c842bb59/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 302.6, 508.1f-g/j/m, 607.2h, and 702.154a-d. After attacking and tapping
// a 2-power nonattacker, the Enlist trigger gives this 2/5 creature +2/+0.
func TestEnlistFamilyBarkweaveCrusher(t *testing.T) {
	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.BarkweaveCrusher.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	enlistFamilyAttack(t, g, attacker, enlisted)

	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("Barkweave Crusher power after enlist = %d, want 4", got)
	}
	toughness, _ := effectiveToughness(g, attacker)
	if got := toughness; got != 5 {
		t.Fatalf("Barkweave Crusher toughness after enlist = %d, want 5", got)
	}
}

// Source review: Oracle dc763f20-73e7-4db7-9e98-6f419b8fba84, DMU printing
// 8b12da45-aff4-406d-83a3-39ecaff6ff56; its Scryfall rulings URI
// https://api.scryfall.com/cards/8b12da45-aff4-406d-83a3-39ecaff6ff56/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 509.1b, 607.2h, 702.9a-b, and 702.154a-d. Flying rejects
// the nonflying block but permits the flying blocker after the enlist trigger.
func TestEnlistFamilyCoalitionSkyknight(t *testing.T) {
	t.Run("a nonflying creature can block a vanilla creature", func(t *testing.T) {
		g := newEnlistCombatGame()
		vanilla := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
		engine := NewEngine(nil)
		attack := action.DeclareAttackersWithBands([]game.AttackDeclaration{
			attackAt(vanilla, game.AttackTarget{Player: game.Player2}),
		}, nil)
		if !engine.applyAction(g, game.Player1, attack) {
			t.Fatal("baseline vanilla attack declaration failed")
		}
		g.Turn.Step = game.StepDeclareBlockers
		blocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
			Blocker: blocker.ObjectID, Blocking: vanilla.ObjectID,
		}}))
		if !engine.applyDeclareBlockers(g, game.Player2, blocks) {
			t.Fatal("the same nonflying block was not legal without flying")
		}
	})

	g := newEnlistCombatGame()
	attacker := addCombatPermanent(g, game.Player1, authored.CoalitionSkyknight.New())
	enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine := enlistFamilyAttack(t, g, attacker, enlisted)
	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("Skyknight power after enlist = %d, want 4", got)
	}

	groundBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flyingBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Flying blocker", Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	groundBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: groundBlocker.ObjectID, Blocking: attacker.ObjectID,
	}}))
	if engine.applyDeclareBlockers(g, game.Player2, groundBlock) {
		t.Fatal("Coalition Skyknight was blocked by a creature without flying or reach")
	}
	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: flyingBlocker.ObjectID, Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("a flying creature could not block Coalition Skyknight")
	}
}

// Source review: Oracle f3b4097f-b84f-4645-81f1-018366e40eeb, DMU printing
// acd5a635-bcf5-4a05-b00b-bf40322823fe; its Scryfall rulings URI
// https://api.scryfall.com/cards/acd5a635-bcf5-4a05-b00b-bf40322823fe/rulings
// returned four WotC rulings dated 2022-09-09. Official CR, effective
// 2026-11-06, https://magic.wizards.com/en/rules (PDF
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf):
// read 508.1f-g/j/m, 607.2h, 702.21a, and 702.154a-d. The opposing targeted
// spell is countered when its controller declines Ward {2}; the Tortoise later
// attacks and gets its enlisted creature's power.
func TestEnlistFamilyHexbaneTortoise(t *testing.T) {
	t.Run("ward counters an opposing spell when its controller declines", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		tortoise := addCombatPermanent(g, game.Player1, authored.HexbaneTortoise.New())
		spellID := addCardToHand(g, game.Player2, targetCreatureInstant())
		for _, color := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
			g.Players[game.Player2].ManaPool.Add(color, 3)
		}
		setMainPhasePriority(g, game.Player2)
		if !engine.applyAction(g, game.Player2, action.CastSpell(spellID,
			[]game.Target{game.PermanentTarget(tortoise.ObjectID)}, 0, nil)) {
			t.Fatal("opponent could not target Hexbane Tortoise")
		}
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("Ward {2} did not trigger for the opposing spell")
		}
		var agents [game.NumPlayers]PlayerAgent
		agents[game.Player2] = &chosenTapPolicy{may: false}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		if g.Stack.Size() != 0 || !g.Players[game.Player2].Graveyard.Contains(spellID) {
			t.Fatal("the opposing spell was not countered after its controller declined Ward {2}")
		}
	})

	t.Run("enlist adds the chosen creature's power", func(t *testing.T) {
		g := newEnlistCombatGame()
		attacker := addCombatPermanent(g, game.Player1, authored.HexbaneTortoise.New())
		enlisted := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		enlistFamilyAttack(t, g, attacker, enlisted)
		if got := effectivePower(g, attacker); got != 6 {
			t.Fatalf("Hexbane Tortoise power after enlist = %d, want 6", got)
		}
	})
}

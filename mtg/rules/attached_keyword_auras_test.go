package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lookupAttachedKeywordAura(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("default registry Oracle ID %s returned %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func attachedKeywordAuraTestCreature(name string, colors ...color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Colors:    colors,
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func castAttachedKeywordAura(
	t *testing.T,
	g *game.Game,
	engine *Engine,
	oracleID, name string,
	caster game.PlayerID,
	manaColor mana.Color,
	manaAmount int,
	target *game.Permanent,
) *game.Permanent {
	t.Helper()
	def := lookupAttachedKeywordAura(t, oracleID, name)
	auraID := addCardToHand(g, caster, def)
	g.Players[caster].ManaPool.Add(manaColor, manaAmount)
	setMainPhasePriority(g, caster)
	if !engine.applyAction(g, caster, action.CastSpell(
		auraID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatalf("%s could not be cast targeting creature %v", name, target.ObjectID)
	}
	if got := g.Players[caster].ManaPool.Amount(manaColor); got != 0 {
		t.Fatalf("%s left %d mana of its casting color, want its full cost paid", name, got)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatalf("%s did not enter the battlefield", name)
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != target.ObjectID {
		t.Fatalf("%s attached to %v, want creature %v", name, aura.AttachedTo, target.ObjectID)
	}
	return aura
}

func declareAttachedAuraTestAttack(
	t *testing.T,
	g *game.Game,
	engine *Engine,
	attackingPlayer game.PlayerID,
	attacker *game.Permanent,
	defender game.PlayerID,
) {
	t.Helper()
	g.Turn.ActivePlayer = attackingPlayer
	g.Turn.PriorityPlayer = attackingPlayer
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackers := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: defender},
	}}))
	if !engine.applyDeclareAttackers(g, attackingPlayer, attackers) {
		t.Fatalf("creature %v could not attack %v", attacker.ObjectID, defender)
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Fear Oracle ID 355bbe9b-59bf-470f-8600-410af4c7fe18, printing revision
// d8fa57eb-e307-4104-a291-c6cbfa235816: {B}{B}, “Enchant creature (Target a
// creature as you cast this. This card enters attached to that creature.) /
// Enchanted creature has fear. (It can't be blocked except by artifact creatures
// and/or black creatures.)” Scryfall rulings
// https://api.scryfall.com/cards/d8fa57eb-e307-4104-a291-c6cbfa235816/rulings
// returned an empty list on 2026-11-10. Official CR effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 303.1 and 303.4a-d, 601.2c/f-h, 608.2b and 608.3b-c, 604.1,
// 611.3a-b, 613.1f, 702.5a-b and 702.36a-b. The Aura targets a creature,
// attaches as it resolves, and continuously grants fear in layer 6; fear
// excludes nonartifact, nonblack blockers.
func TestFearAuraGrantsFearInCombat(t *testing.T) {
	newCombat := func() (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		attacker := addCombatPermanent(g, game.Player1, attachedKeywordAuraTestCreature("Fear target"))
		blocker := addCombatPermanent(g, game.Player2, attachedKeywordAuraTestCreature("Ordinary blocker"))
		return g, engine, attacker, blocker
	}

	baseline, baselineEngine, ordinaryAttacker, ordinaryBlocker := newCombat()
	declareAttachedAuraTestAttack(t, baseline, baselineEngine, game.Player1, ordinaryAttacker, game.Player2)
	baseline.Turn.Step = game.StepDeclareBlockers
	baseline.Turn.PriorityPlayer = game.Player2
	ordinaryBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  ordinaryBlocker.ObjectID,
		Blocking: ordinaryAttacker.ObjectID,
	}}))
	if !baselineEngine.applyDeclareBlockers(baseline, game.Player2, ordinaryBlock) {
		t.Fatal("ordinary creature could not block the same kind of attacker without Fear")
	}

	g, engine, attacker, blocker := newCombat()
	blackBlocker := addCombatPermanent(g, game.Player2, attachedKeywordAuraTestCreature("Black blocker", color.Black))
	castAttachedKeywordAura(t, g, engine, "355bbe9b-59bf-470f-8600-410af4c7fe18", "Fear",
		game.Player1, mana.B, 2, attacker)
	if !hasKeyword(g, attacker, game.Fear) {
		t.Fatal("enchanted creature did not gain fear")
	}
	declareAttachedAuraTestAttack(t, g, engine, game.Player1, attacker, game.Player2)
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player2
	ordinaryBlock = mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: attacker.ObjectID,
	}}))
	if engine.applyDeclareBlockers(g, game.Player2, ordinaryBlock) {
		t.Fatal("Fear allowed a nonartifact, nonblack creature to block")
	}

	blackBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blackBlocker.ObjectID,
		Blocking: attacker.ObjectID,
	}}))
	if !engine.applyDeclareBlockers(g, game.Player2, blackBlock) {
		t.Fatal("Fear did not allow a black creature to block")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Lifelink Oracle ID 4fe8316a-9cff-43e9-a1d5-993a0c9daf3a, printing revision
// a8e207d4-9930-4aff-a7c8-b53bd1b5d566: {W}, “Enchant creature / Enchanted
// creature has lifelink. (Damage dealt by the creature also causes its
// controller to gain that much life.)” Scryfall rulings
// https://api.scryfall.com/cards/a8e207d4-9930-4aff-a7c8-b53bd1b5d566/rulings
// returned the WotC ruling dated 2009-10-01 on 2026-11-10: the enchanted
// creature's controller, not Lifelink's controller, gains the life. Official
// CR effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 303.4a-b/e, 601.2c, 608.2b and 608.3b-c, 604.1, 611.3a-b, 613.1f,
// 702.5a-b and 702.15a-c. The resolving Aura stays controlled by its caster
// while the layer-6 lifelink grant causes the enchanted creature's controller
// to gain life from its damage.
func TestLifelinkAuraGrantsLifeToEnchantedCreatureController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player2, attachedKeywordAuraTestCreature("Lifelink target"))
	aura := castAttachedKeywordAura(t, g, engine, "4fe8316a-9cff-43e9-a1d5-993a0c9daf3a", "Lifelink",
		game.Player1, mana.W, 1, attacker)
	if aura.Controller != game.Player1 || !hasKeyword(g, attacker, game.Lifelink) {
		t.Fatal("Lifelink was not controlled by its caster or did not grant lifelink")
	}

	player1Life := g.Players[game.Player1].Life
	player2Life := g.Players[game.Player2].Life
	player3Life := g.Players[game.Player3].Life
	declareAttachedAuraTestAttack(t, g, engine, game.Player2, attacker, game.Player3)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})

	if got := g.Players[game.Player3].Life; got != player3Life-2 {
		t.Fatalf("defending player life = %d, want %d after 2 combat damage", got, player3Life-2)
	}
	if got := g.Players[game.Player2].Life; got != player2Life+2 {
		t.Fatalf("enchanted creature controller life = %d, want %d after lifelink", got, player2Life+2)
	}
	if got := g.Players[game.Player1].Life; got != player1Life {
		t.Fatalf("Aura controller life = %d, want unchanged %d", got, player1Life)
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Vigilance Oracle ID 70570170-be76-4c56-9151-c4b6e253f462, printing revision
// d24dc353-abc9-430a-a7f5-5da3b38cd411: {W}, “Enchant creature / Enchanted
// creature has vigilance. (Attacking doesn't cause it to tap.)” Scryfall
// rulings https://api.scryfall.com/cards/d24dc353-abc9-430a-a7f5-5da3b38cd411/rulings
// returned an empty list on 2026-11-10. Official CR effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 303.4a-b/e, 601.2c, 608.2b and 608.3b-c, 604.1, 611.3a-b, 613.1f,
// 702.5a-b and 702.20a-c. The Aura targets and attaches to a creature; its
// static layer-6 effect grants vigilance, so declaring that creature as an
// attacker does not tap it.
func TestVigilanceAuraAllowsEnchantedCreatureToAttackUntapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, attachedKeywordAuraTestCreature("Vigilance target"))
	castAttachedKeywordAura(t, g, engine, "70570170-be76-4c56-9151-c4b6e253f462", "Vigilance",
		game.Player1, mana.W, 1, attacker)
	if !hasKeyword(g, attacker, game.Vigilance) {
		t.Fatal("enchanted creature did not gain vigilance")
	}

	declareAttachedAuraTestAttack(t, g, engine, game.Player1, attacker, game.Player2)
	if len(g.Combat.Attackers) != 1 || g.Combat.Attackers[0].Attacker != attacker.ObjectID {
		t.Fatal("enchanted creature was not declared as an attacker")
	}
	if attacker.Tapped {
		t.Fatal("attacking creature with vigilance was tapped")
	}
}

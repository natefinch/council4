package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// applyFlareOfFortitude resolves Flare of Fortitude's spell body: the
// controller-scoped life-total-can't-change player rule and the group grant of
// hexproof and indestructible to the permanents the controller controls, both
// until end of turn.
func applyFlareOfFortitude(engine *Engine, g *game.Game, controller game.PlayerID) {
	obj := &game.StackObject{Controller: controller}
	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{
			{Kind: game.RuleEffectLifeTotalCantChange, AffectedPlayer: game.PlayerYou},
		},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)
	resolveInstruction(engine, g, obj, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			Group:       game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
			AddKeywords: []game.Keyword{game.Hexproof, game.Indestructible},
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)
}

// TestFlareOfFortitudeLifeTotalCantChange proves the until-end-of-turn player
// rule makes the caster's life total immutable to gain, loss, damage, and life
// payment while the effect is active, then mutable again once it expires.
func TestFlareOfFortitudeLifeTotalCantChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	applyFlareOfFortitude(engine, g, game.Player1)

	startLife := g.Players[game.Player1].Life
	if gainLife(g, game.Player1, 5) != 0 || loseLife(g, game.Player1, 5) != 0 {
		t.Fatal("immutable life total reported a life change")
	}
	if g.Players[game.Player1].Life != startLife {
		t.Fatalf("blocked life change mutated life: %d", g.Players[game.Player1].Life)
	}

	// Damage still resolves (the damage event is emitted so poison/commander
	// damage behave by rules) but it causes no life loss.
	beforeEvents := len(g.Events)
	dealt := dealPlayerDamage(g, 0, 0, game.Player2, game.Player1, 4, false)
	if dealt != 4 {
		t.Fatalf("damage dealt = %d, want the full 4 to still be dealt", dealt)
	}
	if g.Players[game.Player1].Life != startLife {
		t.Fatalf("damage changed the immutable life total: %d", g.Players[game.Player1].Life)
	}
	sawDamage := false
	for _, event := range g.Events[beforeEvents:] {
		if event.Kind == game.EventDamageDealt && event.Player == game.Player1 {
			sawDamage = true
		}
	}
	if !sawDamage {
		t.Fatal("no damage event emitted; poison/commander damage would not behave by rules")
	}

	// Paying life as a cost is illegal while the life total can't change.
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
		t.Fatal("life payment remained legal while life total could not change")
	}

	// The opponent's life total is unaffected.
	if loseLife(g, game.Player2, 3) != 3 {
		t.Fatal("opponent life total was wrongly made immutable")
	}

	// After the end-of-turn cleanup the rule expires and life is mutable again.
	expireRuleEffects(g)
	if loseLife(g, game.Player1, 1) != 1 || g.Players[game.Player1].Life != startLife-1 {
		t.Fatal("life total remained immutable after the effect expired")
	}
}

// TestFlareOfFortitudeGroupProtection proves the group grant gives hexproof and
// indestructible to every permanent the caster controls at resolution
// (regardless of type), excludes opponents' permanents, protects against
// opponent targeting and destruction while leaving the caster's own targeting
// legal, snapshots its membership per CR 611.2c ("gain" locks the set of
// affected permanents at resolution), and expires at end of turn.
func TestFlareOfFortitudeGroupProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controlled Artifact",
		Types: []types.Card{types.Artifact},
	}})
	enchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controlled Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	applyFlareOfFortitude(engine, g, game.Player1)

	// Every permanent the caster controls gains both keywords, across all types.
	for _, permanent := range []*game.Permanent{creature, artifact, enchantment} {
		if !hasKeyword(g, permanent, game.Hexproof) || !hasKeyword(g, permanent, game.Indestructible) {
			t.Fatalf("controlled permanent %q did not gain both keywords", permanentName(g, permanent))
		}
	}
	// The opponent's permanent gains neither.
	if hasKeyword(g, opponentCreature, game.Hexproof) || hasKeyword(g, opponentCreature, game.Indestructible) {
		t.Fatal("opponent permanent wrongly gained the protections")
	}

	// Hexproof stops the opponent from targeting a protected permanent while the
	// caster's own spells can still target it.
	if !targetProtectedFromSource(g, game.Player2, nil, 0, game.PermanentTarget(creature.ObjectID)) {
		t.Fatal("opponent could target a hexproof permanent you control")
	}
	if targetProtectedFromSource(g, game.Player1, nil, 0, game.PermanentTarget(creature.ObjectID)) {
		t.Fatal("your own spell could not target your hexproof permanent")
	}

	// Indestructible keeps the permanent on the battlefield when destroyed.
	if _, destroyed := destroyPermanent(g, creature.ObjectID); destroyed {
		t.Fatal("indestructible permanent was destroyed")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("indestructible permanent left the battlefield")
	}

	// The set of affected permanents is fixed at resolution (CR 611.2c). A
	// permanent that later leaves the caster's control keeps the grant, and one
	// that enters the caster's control afterward does not gain it.
	creature.Controller = game.Player2
	if !hasKeyword(g, creature, game.Hexproof) || !hasKeyword(g, creature, game.Indestructible) {
		t.Fatal("snapshotted permanent lost the grant after a control change")
	}
	creature.Controller = game.Player1
	opponentCreature.Controller = game.Player1
	if hasKeyword(g, opponentCreature, game.Hexproof) || hasKeyword(g, opponentCreature, game.Indestructible) {
		t.Fatal("permanent that entered your control after resolution wrongly gained the grant")
	}
	opponentCreature.Controller = game.Player2

	// The grant expires with the turn's cleanup.
	expireCleanupDurations(g)
	if hasKeyword(g, artifact, game.Hexproof) || hasKeyword(g, artifact, game.Indestructible) {
		t.Fatal("protections persisted past end of turn")
	}
}

// permanentName returns a permanent's printed name for test diagnostics.
func permanentName(g *game.Game, permanent *game.Permanent) string {
	if instance, ok := g.CardInstances[permanent.CardInstanceID]; ok && instance.Def != nil {
		return instance.Def.Name
	}
	if permanent.TokenDef != nil {
		return permanent.TokenDef.Name
	}
	return ""
}

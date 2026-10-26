package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Exile cast-route parity (#3405, part of #3355): legal-action enumeration
// lists an exile cast face exactly when some exile cast route authorizes it,
// the same route set cast validation and application use.
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     601.3 (a player can begin to cast a spell only if a rule or effect
//     allows it), 601.2b (alternative costs and the value of X are announced;
//     two alternative costs can't apply to one spell), 118.9a (only one
//     alternative cost applies), 107.3b (a spell cast paying neither its mana
//     cost nor an alternative cost that includes X must choose X=0), 118.14
//     ("mana of any type can be spent" applies only to spells cast that way).
//   - Court of Locthwain, printing a8e1f698-1c04-4f31-aeaa-b796c5eb5070
//     (oracle_id 6106ba65-f07b-41ec-8b59-e052bbde5529), {2}{B}{B}: "When this
//     enchantment enters, you become the monarch. / At the beginning of your
//     upkeep, exile the top card of target opponent's library. You may play
//     that card for as long as it remains exiled, and mana of any type can be
//     spent to cast it. If you're the monarch, until end of turn, you may cast
//     a spell from among cards exiled with this enchantment without paying its
//     mana cost." Ruling 2023-09-01: you pay the costs for an exiled card if
//     you cast it, and may pay its alternative costs instead of its mana cost.
//   - Exsanguinate, printing f11d7311-4066-4a5d-ba28-9857fa707a0b (oracle_id
//     8164b1e8-3350-465e-8a17-75f57d326344), {X}{B}{B} Sorcery: "Each opponent
//     loses X life. You gain life equal to the life lost this way."
//   - Evelyn, the Covetous, printing c0dad61f-36cd-46af-82b7-a02e04efd676
//     (oracle_id 4ab6f615-1f04-4ed2-88b2-49faa8356dd0): "... Once each turn,
//     you may play a card from exile with a collection counter on it if it
//     was exiled by an ability you controlled, and you may spend mana as
//     though it were mana of any color to cast it." No rulings.

// courtMonarchClause returns the registered Court of Locthwain upkeep
// sequence's monarch clause and the link key of the pool its free cast reads.
func courtMonarchClause(t *testing.T) (*game.Instruction, game.LinkedKey) {
	t.Helper()
	clause := locthwainUpkeepSequence(t)[1]
	apply, ok := clause.Primitive.(game.ApplyRule)
	if !ok || len(apply.RuleEffects) != 1 || apply.RuleEffects[0].Kind != game.RuleEffectCastLinkedExileForFree {
		t.Fatalf("Court of Locthwain upkeep clause 2 = %+v, want the linked free-cast ApplyRule", clause.Primitive)
	}
	return &clause, apply.RuleEffects[0].ExiledLinkKey
}

// TestLinkedFreeCastOnlyPermissionIsListedAndApplied puts a card in Court of
// Locthwain's linked pool covered only by the registered monarch clause's
// free-cast permission, with no per-card play grant and no mana. Casting it
// from exile is validated, listed by legal actions (also after cloning) and
// applied through that permission, which is spent (CR 601.3, 118.9a). The card
// rests either in an opponent's exile or the caster's own.
func TestLinkedFreeCastOnlyPermissionIsListedAndApplied(t *testing.T) {
	t.Parallel()
	for name, owner := range map[string]game.PlayerID{"opponent's exile": game.Player2, "own exile": game.Player1} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player1].IsMonarch = true
			obj := locthwainSource(g)
			clause, key := courtMonarchClause(t)
			cardID := addCardToExile(g, owner, opponentSpellDef())
			rememberLinkedObject(g, linkedObjectSourceKey(g, obj, string(key)), game.LinkedObjectRef{CardID: cardID})
			engine.resolveInstructionWithChoices(g, obj, clause, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			if linkedFreeCastPermissions(g) != 1 {
				t.Fatal("precondition: the monarch's linked free-cast permission is missing")
			}
			if canCastFromZoneByRuleEffect(g, game.Player1, cardID, zone.Exile, game.FaceFront) ||
				canCastSpellsFromZoneByRuleEffect(g, game.Player1, cardID, zone.Exile, game.FaceFront) {
				t.Fatal("precondition: another permission covers the pooled card")
			}
			setSorcerySpeedTurn(g, game.Player1)
			if !g.Players[game.Player1].ManaPool.IsEmpty() {
				t.Fatal("precondition: Player1 has mana")
			}

			act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
			if !engine.canCastSpellFaceFromZoneWithKicker(g, game.Player1, cardID, zone.Exile, game.FaceFront, nil, 0, nil, false) {
				t.Fatal("validation rejected the linked free cast")
			}
			if !actionsContain(engine.legalActions(g, game.Player1), act) {
				t.Fatal("legal actions omit the linked free cast")
			}
			if !actionsContain(engine.legalActions(g.Clone(), game.Player1), act) {
				t.Fatal("legal actions of a cloned game omit the linked free cast")
			}
			if !engine.applyAction(g, game.Player1, act) {
				t.Fatal("applying the listed linked free cast failed")
			}
			top, ok := g.Stack.Peek()
			if !ok || top.SourceID != cardID || top.Controller != game.Player1 || top.ManaSpentToCast != 0 {
				t.Fatalf("stack top = %+v, want Player1's free cast of the pooled card", top)
			}
			if linkedFreeCastPermissions(g) != 0 {
				t.Fatal("the linked free cast did not consume its one-shot permission")
			}
		})
	}
}

// TestFreeAndEvelynRoutesBothListedForXSpell exiles registered Exsanguinate
// with Evelyn, the Covetous's collection counter and a per-card "without
// paying its mana cost" permission. The free route affords only X=0
// (CR 107.3b) and Evelyn's any-color route pays a larger X with blue mana, so
// X=0 through X=2 are listed and applied with four blue mana: the X=0 cast
// takes the free route and leaves Evelyn's once-each-turn use unspent, while
// the X=2 cast pays through Evelyn's route and spends it (CR 601.2b, 118.9a).
func TestFreeAndEvelynRoutesBothListedForXSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evelyn := groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
	cardID := collectionExiledCard(g, registryCardDef(t, "Exsanguinate"))
	freePlayPermission(g, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.U, 4)

	assertCastXRange(t, g, engine, game.Player1, cardID, zone.Exile, game.FaceFront, nil, 3, "Exsanguinate under free and Evelyn routes")

	cast := func(xValue int) (*game.Game, *game.StackObject) {
		t.Helper()
		clone := g.Clone()
		castXAndCheckStack(t, clone, engine, game.Player1, cardID, zone.Exile, game.FaceFront, nil, xValue)
		top, _ := clone.Stack.Peek()
		return clone, top
	}
	free, top := cast(0)
	if top.ManaSpentToCast != 0 || free.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
		t.Fatalf("X=0 cast spent %d mana, Evelyn used = %v; want the free route", top.ManaSpentToCast, free.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID])
	}
	paid, top := cast(2)
	if top.ManaSpentToCast != 4 || !paid.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
		t.Fatalf("X=2 cast spent %d mana, Evelyn used = %v; want four blue mana through Evelyn's route", top.ManaSpentToCast, paid.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID])
	}
	if !paid.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("pool = %+v after the X=2 cast, want empty", paid.Players[game.Player1].ManaPool)
	}
}

package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#3404, Engine gap; no card admitted):
//   - Evelyn, the Covetous (oracle 4ab6f615-1f04-4ed2-88b2-49faa8356dd0): "Once
//     each turn, you may play a card from exile with a collection counter on it
//     if it was exiled by an ability you controlled, and you may spend mana as
//     though it were mana of any color to cast it." No Scryfall rulings.
//   - Grolnok, the Omnivore (oracle 53029460-50a0-4934-8fa9-0271517cd5d0): "You
//     may play lands and cast spells from among cards you own in exile with
//     croak counters on them." Rulings 2021-11-19 cover only croak-counter
//     provenance.
//   - CR 116.2a/305.1: playing a land is a special action that puts it onto the
//     battlefield from the zone it was in. CR 601.2b: the caster's earlier
//     choice of how to cast (from which permission) restricts later choices;
//     the same permission choice applies to a land play. Only the permission
//     actually used spends its once-each-turn use.

// playLandFromExile is the land-play action for cardID's front face in exile.
func playLandFromExile(cardID id.ID) action.Action {
	return action.PlayLandFaceFromZone(cardID, zone.Exile, game.FaceFront)
}

// exiledLandOwnedBy puts a basic land owned by owner into owner's exile.
func exiledLandOwnedBy(g *game.Game, owner game.PlayerID, name string) id.ID {
	return addCardToExile(g, owner, &game.CardDef{CardFace: game.CardFace{
		Name: name, Supertypes: []types.Super{types.Basic}, Types: []types.Card{types.Land}}})
}

// collectedSorcery is a {1} sorcery Evelyn's cast permission can authorize.
func collectedSorcery() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Collected Sorcery",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

// assertLandPlayedFromExile applies the land play from exile, then checks the
// land is on the battlefield and counted as the turn's land play.
func assertLandPlayedFromExile(t *testing.T, g *game.Game, engine *Engine, landID id.ID, agents [game.NumPlayers]PlayerAgent) {
	t.Helper()
	played := g.Turn.LandsPlayedThisTurn
	if !engine.applyActionWithChoices(g, game.Player1, playLandFromExile(landID), agents, &TurnLog{}) {
		t.Fatal("playing the land from exile was rejected")
	}
	if g.Players[game.Player1].Exile.Contains(landID) || !permanentWithCardOnBattlefield(g, landID) {
		t.Fatal("the land did not move from exile to the battlefield")
	}
	if g.Turn.LandsPlayedThisTurn != played+1 {
		t.Fatalf("lands played = %d, want %d", g.Turn.LandsPlayedThisTurn, played+1)
	}
}

func permanentWithCardOnBattlefield(g *game.Game, cardID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return true
		}
	}
	return false
}

// castCollectedSorcery casts spellID from exile through Evelyn's cast
// permission with one blue mana, which her any-color rider pays {1} with.
func castCollectedSorcery(g *game.Game, engine *Engine, spellID id.ID) bool {
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	return engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(spellID, zone.Exile, game.FaceFront, nil, 0, nil))
}

// TestRegisteredGrolnokLandPlayLeavesEvelynUnspent puts registered Grolnok, the
// Omnivore and Evelyn, the Covetous under Player1's control in both battlefield
// orders. A land Player1 owns, exiled by Player1's own ability, carries both a
// croak and a collection counter. Grolnok's unlimited permission authorizes
// playing it, so Evelyn's once-each-turn use stays unspent and she can still
// authorize casting a different collection-counter card that turn, which then
// spends her use.
func TestRegisteredGrolnokLandPlayLeavesEvelynUnspent(t *testing.T) {
	t.Parallel()
	for _, evelynFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		var evelyn *game.Permanent
		if evelynFirst {
			evelyn = addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Grolnok, the Omnivore"))
		} else {
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Grolnok, the Omnivore"))
			evelyn = addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
		}
		landID := exiledLandOwnedBy(g, game.Player1, "Forest")
		g.AddExileCounterFromController(landID, counter.Collection, 1, game.Player1)
		g.AddExileCounter(landID, counter.Croak, 1)
		spellID := collectionExiledCard(g, collectedSorcery())
		giveMainPhasePriority(g, game.Player1)

		if routes := playLandFromZoneRoutes(g, game.Player1, landID, zone.Exile); len(routes) != 2 || routes[0].OncePerTurn || !routes[1].OncePerTurn {
			t.Fatalf("evelynFirst=%v: routes = %+v, want Grolnok's unlimited route before Evelyn's", evelynFirst, routes)
		}
		if !canPlayLandFromZoneByRuleEffect(g, game.Player1, landID, zone.Exile) {
			t.Fatalf("evelynFirst=%v: the land is not playable from exile", evelynFirst)
		}
		assertExileCastListed(t, g, engine, playLandFromExile(landID), true)
		assertLandPlayedFromExile(t, g, engine, landID, [game.NumPlayers]PlayerAgent{})
		if g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatalf("evelynFirst=%v: Grolnok's land play spent Evelyn's once-each-turn use", evelynFirst)
		}

		if !canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront, false) {
			t.Fatalf("evelynFirst=%v: Evelyn cannot authorize a cast after Grolnok's land play", evelynFirst)
		}
		if !castCollectedSorcery(g, engine, spellID) {
			t.Fatalf("evelynFirst=%v: casting through Evelyn after Grolnok's land play was rejected", evelynFirst)
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != spellID || obj.SourceZone != zone.Exile {
			t.Fatalf("evelynFirst=%v: stack object = %+v, want the collected sorcery", evelynFirst, obj)
		}
		if !g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatalf("evelynFirst=%v: casting through Evelyn did not spend her use", evelynFirst)
		}
	}
}

// TestEvelynAloneLandPlaySpendsHerUse plays a collection-counter land only
// Evelyn's permission covers. The play spends her once-each-turn use, so a
// second land play (with a land drop to spare) or a cast through her that turn
// is rejected.
func TestEvelynAloneLandPlaySpendsHerUse(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	evelyn := addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
	first := exiledLandOwnedBy(g, game.Player1, "Forest")
	second := exiledLandOwnedBy(g, game.Player1, "Island")
	g.AddExileCounterFromController(first, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(second, counter.Collection, 1, game.Player1)
	spellID := collectionExiledCard(g, collectedSorcery())
	g.Turn.LandsAllowedThisTurn = 2
	giveMainPhasePriority(g, game.Player1)

	assertExileCastListed(t, g, engine, playLandFromExile(first), true)
	assertLandPlayedFromExile(t, g, engine, first, [game.NumPlayers]PlayerAgent{})
	if !g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
		t.Fatal("playing a land only Evelyn covers did not spend her use")
	}
	if !playerCanPlayLand(g, game.Player1) {
		t.Fatal("fixture should leave a land drop to spare")
	}
	assertExileCastListed(t, g, engine, playLandFromExile(second), false)
	if engine.applyAction(g, game.Player1, playLandFromExile(second)) {
		t.Fatal("a second land play through Evelyn was accepted in the same turn")
	}
	if canCastSpellsFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront, false) || castCollectedSorcery(g, engine, spellID) {
		t.Fatal("a cast through Evelyn was accepted after her land play spent the use")
	}
}

// perCardLandPlayGrant is a static ability granting Player1 permission to play
// cardID from exile (the per-card RuleEffectPlayFromZone shape of an impulse
// grant), so its order among battlefield statics can precede or follow Evelyn's.
func perCardLandPlayGrant(g *game.Game, cardID id.ID) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Impulse Grant",
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectPlayFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Exile,
			AffectedCardID: cardID,
		}}}},
	}})
}

// TestPerCardLandPlayPermissionLeavesEvelynUnspent plays a land that both a
// per-card play permission and Evelyn's collection-counter permission cover,
// with the per-card permission as a game rule effect (Court of Locthwain's
// shape, ordered first) and as a battlefield static before and after Evelyn.
// Whatever the effect order, the per-card permission authorizes the play and
// Evelyn's once-each-turn use stays unspent.
func TestPerCardLandPlayPermissionLeavesEvelynUnspent(t *testing.T) {
	t.Parallel()
	for _, grant := range []string{"rule effect", "static before Evelyn", "static after Evelyn"} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		landID := exiledLandOwnedBy(g, game.Player2, "Swamp")
		g.AddExileCounterFromController(landID, counter.Collection, 1, game.Player1)
		var evelyn *game.Permanent
		switch grant {
		case "rule effect":
			evelyn = evelynCovetousPermanent(g, game.Player1)
			g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
				ID:             g.IDGen.Next(),
				Kind:           game.RuleEffectPlayFromZone,
				Controller:     game.Player1,
				AffectedPlayer: game.PlayerYou,
				Duration:       game.DurationPermanent,
				CastFromZone:   zone.Exile,
				AffectedCardID: landID,
				ExpiresFor:     game.Player1,
				SpendAnyMana:   game.ManaSpendAsAnyType,
			})
		case "static before Evelyn":
			perCardLandPlayGrant(g, landID)
			evelyn = evelynCovetousPermanent(g, game.Player1)
		default:
			evelyn = evelynCovetousPermanent(g, game.Player1)
			perCardLandPlayGrant(g, landID)
		}
		giveMainPhasePriority(g, game.Player1)

		if routes := playLandFromZoneRoutes(g, game.Player1, landID, zone.Exile); len(routes) != 2 || routes[0].Kind != game.RuleEffectPlayFromZone {
			t.Fatalf("%s: routes = %+v, want the per-card route first", grant, routes)
		}
		assertExileCastListed(t, g, engine, playLandFromExile(landID), true)
		assertLandPlayedFromExile(t, g, engine, landID, [game.NumPlayers]PlayerAgent{})
		if g.ExilePlayPermissionUsedThisTurn[evelyn.ObjectID] {
			t.Fatalf("%s: the per-card permission's land play spent Evelyn's use", grant)
		}
	}
}

// landRouteAgent answers the land-play permission choice with index and
// records the labels it was offered.
type landRouteAgent struct {
	index   int
	offered *[]string
}

func (landRouteAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a landRouteAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt != "Choose land play permission" {
		return request.DefaultSelection
	}
	for _, option := range request.Options {
		*a.offered = append(*a.offered, option.Label)
	}
	return []int{a.index}
}

// TestLandPlayChoosesAmongLimitedPermissions plays a land two once-each-turn
// permissions cover and no unlimited one does. The player chooses which
// permission to spend; only that source's use is recorded, the other stays
// available, and the recorded use lives in game state so a clone chooses
// independently of the original.
func TestLandPlayChoosesAmongLimitedPermissions(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := evelynCovetousPermanent(g, game.Player1)
	second := evelynCovetousPermanent(g, game.Player1)
	landID := exiledLandOwnedBy(g, game.Player1, "Mountain")
	g.AddExileCounterFromController(landID, counter.Collection, 1, game.Player1)
	giveMainPhasePriority(g, game.Player1)

	if routes := playLandFromZoneRoutes(g, game.Player1, landID, zone.Exile); len(routes) != 2 ||
		routes[0].SourceObjectID != first.ObjectID || routes[1].SourceObjectID != second.ObjectID {
		t.Fatalf("routes = %+v, want one route per once-each-turn source", routes)
	}

	clone := g.Clone()
	var offered []string
	agents := [game.NumPlayers]PlayerAgent{game.Player1: landRouteAgent{index: 1, offered: &offered}}
	assertLandPlayedFromExile(t, clone, engine, landID, agents)
	if len(offered) != 2 {
		t.Fatalf("offered %q, want a choice between the two limited permissions", offered)
	}
	if clone.ExilePlayPermissionUsedThisTurn[first.ObjectID] || !clone.ExilePlayPermissionUsedThisTurn[second.ObjectID] {
		t.Fatalf("used = %v, want only the chosen second permission spent", clone.ExilePlayPermissionUsedThisTurn)
	}
	if len(g.ExilePlayPermissionUsedThisTurn) != 0 || !g.Players[game.Player1].Exile.Contains(landID) {
		t.Fatal("playing the land in the clone changed the original game")
	}

	assertLandPlayedFromExile(t, g, engine, landID, [game.NumPlayers]PlayerAgent{})
	if !g.ExilePlayPermissionUsedThisTurn[first.ObjectID] || g.ExilePlayPermissionUsedThisTurn[second.ObjectID] {
		t.Fatalf("used = %v, want only the default first permission spent", g.ExilePlayPermissionUsedThisTurn)
	}
}

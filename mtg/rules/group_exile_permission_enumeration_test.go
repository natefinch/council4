package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Group play/cast-from-exile permission enumeration (#3357) and the "cards you
// own" ownership rider those permissions need before foreign exile is offered
// (#3358).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     108.3 (owner), 116.2a and 305.1 (playing a land is a special action with
//     the usual timing and land-play limits), 406.3 (exiled cards are face up
//     unless exiled face down), 601.3 (a player may begin to cast a spell only
//     if a rule or effect allows it), 715.3d (other effects that allow a player
//     to cast an Adventurer card may allow casting it as an Adventure).
//   - Evelyn, the Covetous, printing c0dad61f-36cd-46af-82b7-a02e04efd676
//     (oracle_id 4ab6f615-1f04-4ed2-88b2-49faa8356dd0): "Once each turn, you may
//     play a card from exile with a collection counter on it if it was exiled by
//     an ability you controlled ..." Its Scryfall rulings endpoint returned no
//     rulings. The exiled cards include other players' library cards, which rest
//     in their owners' exile.
//   - Grolnok, the Omnivore, printing 7aaba76b-9cec-4c2b-b0eb-8f44201f6422
//     (oracle_id 53029460-50a0-4934-8fa9-0271517cd5d0): "You may play lands and
//     cast spells from among cards you own in exile with croak counters on
//     them." Rulings 2021-11-19: the permission covers all such cards you own,
//     including those exiled by an earlier Grolnok.
//
// Enumeration and direct validation must agree: every exile card a group
// permission reaches is offered for each face it covers, and no other card is.

const (
	groupExileTestCreatureName = "Exiled Bear"
	groupExileTestLandName     = "Exiled Grove"
)

func groupExileRegistryPermanent(t *testing.T, g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupChecked(name)
	if err != nil {
		t.Fatalf("default-registry lookup of %s failed: %v", name, err)
	}
	return addCombatPermanent(g, controller, def)
}

func groupExileCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      groupExileTestCreatureName,
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func groupExileLandDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  groupExileTestLandName,
		Types: []types.Card{types.Land},
	}}
}

func groupExileCast(cardID id.ID, face game.FaceIndex) action.Action {
	return action.CastSpellFaceFromZone(cardID, zone.Exile, face, nil, 0, nil)
}

func groupExilePlayLand(cardID id.ID) action.Action {
	return action.PlayLandFaceFromZone(cardID, zone.Exile, game.FaceFront)
}

// assertGroupExileAction gives player sorcery-speed priority and checks that
// legal-action enumeration and direct application on a clone agree with want.
func assertGroupExileAction(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, act action.Action, want bool, label string) {
	t.Helper()
	giveMainPhasePriority(g, player)
	listed := actionsContain(engine.legalActions(g, player), act)
	applied := engine.applyAction(g.Clone(), player, act)
	if listed != want || applied != want {
		t.Fatalf("%s: player %v listed=%v applied=%v, want %v", label, player, listed, applied, want)
	}
}

// TestGroupExilePermissionEnumeratesForeignExileCards covers Evelyn's group
// permission over cards that rest in another player's exile: the permitted
// spell and land are offered and apply, while the provenance, counter and
// once-per-turn filters exclude the same cards from enumeration as from
// validation, and the owner without the permission gets nothing.
func TestGroupExilePermissionEnumeratesForeignExileCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
	addBasicLandPermanent(g, game.Player1, types.Forest)

	creature := addCardToExile(g, game.Player2, groupExileCreatureDef())
	land := addCardToExile(g, game.Player2, groupExileLandDef())
	exiledByOwner := addCardToExile(g, game.Player2, groupExileCreatureDef())
	noCounter := addCardToExile(g, game.Player2, groupExileCreatureDef())
	g.AddExileCounterFromController(creature, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(land, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(exiledByOwner, counter.Collection, 1, game.Player2)

	assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(creature, game.FaceFront), true, "collection creature in opponent's exile")
	assertGroupExileAction(t, g, engine, game.Player1, groupExilePlayLand(land), true, "collection land in opponent's exile")
	assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(exiledByOwner, game.FaceFront), false, "card exiled by its owner's ability")
	assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(noCounter, game.FaceFront), false, "card without a collection counter")
	assertGroupExileAction(t, g, engine, game.Player2, groupExileCast(creature, game.FaceFront), false, "owner without the permission")
	assertGroupExileAction(t, g, engine, game.Player2, groupExilePlayLand(land), false, "owner without the land permission")

	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, groupExilePlayLand(land)) {
		t.Fatal("playing the enumerated foreign-exile land failed")
	}
	if perm := permanentByCardID(g, land); perm == nil || perm.Controller != game.Player1 {
		t.Fatal("the foreign-exile land did not enter the battlefield")
	}
	if g.Players[game.Player2].Exile.Contains(land) {
		t.Fatal("the played land is still in its owner's exile")
	}
	assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(creature, game.FaceFront), false, "spell after the shared once-per-turn use")

	engine.advanceToNextTurn(g)
	assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(creature, game.FaceFront), true, "spell on the next turn")
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, groupExileCast(creature, game.FaceFront)) {
		t.Fatal("casting the enumerated foreign-exile creature failed")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != creature || top.Controller != game.Player1 {
		t.Fatalf("stack top = %+v, want Player1's %s", top, groupExileTestCreatureName)
	}
}

// TestGroupExilePermissionEnumeratesEveryCoveredFace covers CR 715.3d's
// independent-permission case: an Adventurer card a group cast permission
// reaches, in the caster's own exile or another player's, is offered as both
// its creature face and its Adventure.
func TestGroupExilePermissionEnumeratesEveryCoveredFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}

	own := addCardToExile(g, game.Player1, rosethornAcolyteFixture())
	foreign := addCardToExile(g, game.Player2, rosethornAcolyteFixture())
	g.AddExileCounterFromController(own, counter.Collection, 1, game.Player1)
	g.AddExileCounterFromController(foreign, counter.Collection, 1, game.Player1)

	for _, cardID := range []id.ID{own, foreign} {
		assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(cardID, game.FaceFront), true, "Rosethorn Acolyte creature face")
		assertGroupExileAction(t, g, engine, game.Player1, groupExileCast(cardID, game.FaceAlternate), true, "Seasonal Ritual Adventure face")
	}

	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, groupExileCast(foreign, game.FaceAlternate)) {
		t.Fatal("casting the enumerated foreign Adventure face failed")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != foreign || top.Face != game.FaceAlternate || top.Controller != game.Player1 {
		t.Fatalf("stack top = %+v, want Player1's Seasonal Ritual", top)
	}
}

// TestGroupExilePermissionOwnedCardsOnlyFollowsControl covers Grolnok's "cards
// you own" rider (#3358) on both the land and spell permissions, for
// enumeration and direct validation: the controller reaches only croak cards
// they own, a player who gains control of Grolnok reaches only their own croak
// cards, and the original owner regains use when control returns.
func TestGroupExilePermissionOwnedCardsOnlyFollowsControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grolnok := groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)

	p1Creature := addCardToExile(g, game.Player1, groupExileCreatureDef())
	p1Land := addCardToExile(g, game.Player1, groupExileLandDef())
	p2Creature := addCardToExile(g, game.Player2, groupExileCreatureDef())
	p2Land := addCardToExile(g, game.Player2, groupExileLandDef())
	for _, cardID := range []id.ID{p1Creature, p1Land, p2Creature, p2Land} {
		g.AddExileCounter(cardID, counter.Croak, 1)
	}

	assertReach := func(controller game.PlayerID, label string) {
		t.Helper()
		for _, player := range []game.PlayerID{game.Player1, game.Player2} {
			ownCreature, ownLand, otherCreature, otherLand := p1Creature, p1Land, p2Creature, p2Land
			if player == game.Player2 {
				ownCreature, ownLand, otherCreature, otherLand = p2Creature, p2Land, p1Creature, p1Land
			}
			want := player == controller
			assertGroupExileAction(t, g, engine, player, groupExileCast(ownCreature, game.FaceFront), want, label+": own croak creature")
			assertGroupExileAction(t, g, engine, player, groupExilePlayLand(ownLand), want, label+": own croak land")
			assertGroupExileAction(t, g, engine, player, groupExileCast(otherCreature, game.FaceFront), false, label+": another owner's croak creature")
			assertGroupExileAction(t, g, engine, player, groupExilePlayLand(otherLand), false, label+": another owner's croak land")
		}
	}

	assertReach(game.Player1, "owner controls Grolnok")

	steal := game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: grolnok.ObjectID,
		Layer: game.LayerControl, NewController: opt.Val(game.Player2), Duration: game.DurationPermanent,
	}
	g.ContinuousEffects = append(g.ContinuousEffects, steal)
	if effectiveController(g, grolnok) != game.Player2 {
		t.Fatal("control-changing effect did not give Player2 Grolnok")
	}
	assertReach(game.Player2, "Player2 controls Grolnok")

	g.ContinuousEffects = nil
	if effectiveController(g, grolnok) != game.Player1 {
		t.Fatal("Grolnok did not return to Player1")
	}
	assertReach(game.Player1, "owner regains Grolnok")
}

func groupExileCastX(cardID id.ID, xValue int) action.Action {
	return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, xValue, nil)
}

func groupExileRegistryDef(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupChecked(name)
	if err != nil {
		t.Fatalf("default-registry lookup of %s failed: %v", name, err)
	}
	return def
}

// assertExileXRange checks that exactly the X values below limit are offered
// and directly castable for cardID from exile, and that limit is neither.
func assertExileXRange(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, limit int, label string) {
	t.Helper()
	for x := 0; x <= limit; x++ {
		assertGroupExileAction(t, g, engine, player, groupExileCastX(cardID, x), x < limit, fmt.Sprintf("%s X=%d", label, x))
	}
}

// TestExileAnyManaPermissionEnumeratesXSpells covers CR 107.3a and 609.4b for
// X spells cast from exile: a permission to spend mana as though it were mana
// of any color changes only how the {X}{G} cost may be paid, so the offered X
// values are those the permitted payment affords. With two Islands, Evelyn's
// group permission offers Mistcutter Hydra for X=0 and X=1 from its caster's
// own exile and from another player's exile, as does Court of Locthwain's
// per-card "mana of any type" permission; X=2 is unaffordable and offered by
// neither enumeration nor validation. Grolnok's permission carries no mana
// rider, so the same Hydra stays uncastable with Islands alone and is offered
// for X=0 through X=2 once a Forest pays its {G}.
//
// Source review: Mistcutter Hydra, printing
// 3c844ed7-8ed2-4a39-9134-e14e476ab0c4 (oracle_id
// 30b5b764-478a-46fa-8f95-a1f515925a2d), {X}{G}: "This spell can't be
// countered. / Haste, protection from blue / This creature enters with X +1/+1
// counters on it." Ruling 2013-09-15 concerns only counter attempts. CR 107.3a
// (the caster announces X), 601.2f-h (total cost determined and paid) and
// 609.4b ("as though it were mana of any [type or color]" affects only how the
// cost may be paid) from the Comprehensive Rules cited above.
func TestExileAnyManaPermissionEnumeratesXSpells(t *testing.T) {
	t.Run("Evelyn group permission", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
		addBasicLandPermanent(g, game.Player1, types.Island)
		addBasicLandPermanent(g, game.Player1, types.Island)
		hydra := groupExileRegistryDef(t, "Mistcutter Hydra")
		own := addCardToExile(g, game.Player1, hydra)
		foreign := addCardToExile(g, game.Player2, hydra)
		g.AddExileCounterFromController(own, counter.Collection, 1, game.Player1)
		g.AddExileCounterFromController(foreign, counter.Collection, 1, game.Player1)

		assertExileXRange(t, g, engine, game.Player1, own, 2, "own-exile Hydra")
		assertExileXRange(t, g, engine, game.Player1, foreign, 2, "foreign-exile Hydra")

		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, groupExileCastX(foreign, 1)) {
			t.Fatal("casting the enumerated foreign Hydra for X=1 failed")
		}
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID != foreign || top.Controller != game.Player1 || top.XValue != 1 {
			t.Fatalf("stack top = %+v, want Player1's Mistcutter Hydra with X=1", top)
		}
	})

	t.Run("Court of Locthwain per-card permission", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addBasicLandPermanent(g, game.Player1, types.Island)
		addBasicLandPermanent(g, game.Player1, types.Island)
		topID := addCardToLibrary(g, game.Player2, groupExileRegistryDef(t, "Mistcutter Hydra"))
		obj := &game.StackObject{
			Kind:         game.StackTriggeredAbility,
			Controller:   game.Player1,
			SourceCardID: g.IDGen.Next(),
			SourceID:     g.IDGen.Next(),
			Targets:      []game.Target{game.PlayerTarget(game.Player2)},
		}
		resolveInstruction(engine, g, obj, game.ImpulseExile{
			Player:        game.TargetPlayerReference(0),
			Amount:        game.Fixed(1),
			Duration:      game.DurationPermanent,
			SpendAnyMana:  true,
			PublishLinked: courtOfLocthwainLink,
		}, &TurnLog{})
		if !g.Players[game.Player2].Exile.Contains(topID) {
			t.Fatal("Court's exile did not move the Hydra to its owner's exile")
		}

		assertExileXRange(t, g, engine, game.Player1, topID, 2, "Court-exiled Hydra")
	})

	t.Run("Grolnok permission without a mana rider", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		addBasicLandPermanent(g, game.Player1, types.Island)
		addBasicLandPermanent(g, game.Player1, types.Island)
		hydra := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounter(hydra, counter.Croak, 1)

		assertExileXRange(t, g, engine, game.Player1, hydra, 0, "Hydra with only Islands")

		addBasicLandPermanent(g, game.Player1, types.Forest)
		assertExileXRange(t, g, engine, game.Player1, hydra, 3, "Hydra with a Forest")
	})
}

// assertExileTargetedXRange is assertExileXRange for a spell cast with targets.
func assertExileTargetedXRange(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, targets []game.Target, limit int, label string) {
	t.Helper()
	for x := 0; x <= limit; x++ {
		act := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, targets, x, nil)
		assertGroupExileAction(t, g, engine, player, act, x < limit, fmt.Sprintf("%s X=%d", label, x))
	}
}

// TestExileXCastCostReductionsEnumerate covers CR 601.2f and 118.7a for X
// spells cast from exile: cost reductions lower the total cost the caster must
// pay, so every X the reduced total cost affords is offered, exactly as
// validation and casting accept it, and generic reductions never pay a colored
// requirement. Heartless Summoning's group reduction lets one land cast
// Mistcutter Hydra for X up to 2, from another player's exile under Evelyn and
// from the caster's own exile under Grolnok, while Islands alone still can't
// pay Grolnok's Hydra's {G}. Elderwood Scion's target-gated reduction applies
// only to the Blaze cast that targets it, so one Mountain affords X up to 2
// aimed at the Scion and only X=0 aimed at an opponent.
//
// Source review (Comprehensive Rules and Mistcutter Hydra as cited above):
//   - CR 118.7a: effects that reduce a cost by generic mana affect only its
//     generic component. CR 601.2f: the total cost is the mana or alternative
//     cost plus increases minus all reductions, and can't go below {0}.
//   - Heartless Summoning, printing c50fdb09-e06c-463c-87dd-3c46f4db889a
//     (oracle_id fa1be67e-a07e-42ce-88be-446acd643dc6), {1}{B}: "Creature
//     spells you cast cost {2} less to cast. / Creatures you control get
//     -1/-1." Ruling 2025-01-24: it reduces only the generic mana portion.
//   - Elderwood Scion, printing c16dabf9-6b77-44dd-8fe8-bd3e36a57d7e
//     (oracle_id 99c8647f-e08d-4217-b637-3f8f2bffdaa3), {3}{G}{W}: "Spells you
//     cast that target this creature cost {2} less to cast. / Spells your
//     opponents cast that target this creature cost {2} more to cast." Rulings
//     2012-06-01: the reduction applies even with additional targets, can't
//     reduce colored requirements, and applies to alternative costs.
//   - Blaze, printing b16d0422-61ea-4039-b1c1-e590ace9156e (oracle_id
//     0596920f-9946-42f4-a03b-24aab67f9f1b), {X}{R} Sorcery: "Blaze deals X
//     damage to any target." No rulings.
func TestExileXCastCostReductionsEnumerate(t *testing.T) {
	t.Run("Heartless Summoning under Evelyn", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
		groupExileRegistryPermanent(t, g, game.Player1, "Heartless Summoning")
		addBasicLandPermanent(g, game.Player1, types.Island)
		hydra := addCardToExile(g, game.Player2, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounterFromController(hydra, counter.Collection, 1, game.Player1)

		assertExileXRange(t, g, engine, game.Player1, hydra, 3, "reduced foreign-exile Hydra")

		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, groupExileCastX(hydra, 2)) {
			t.Fatal("casting the enumerated reduced Hydra for X=2 failed")
		}
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID != hydra || top.Controller != game.Player1 || top.XValue != 2 {
			t.Fatalf("stack top = %+v, want Player1's Mistcutter Hydra with X=2", top)
		}
	})

	t.Run("Heartless Summoning under Grolnok", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		groupExileRegistryPermanent(t, g, game.Player1, "Heartless Summoning")
		addBasicLandPermanent(g, game.Player1, types.Island)
		addBasicLandPermanent(g, game.Player1, types.Island)
		hydra := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounter(hydra, counter.Croak, 1)

		assertExileXRange(t, g, engine, game.Player1, hydra, 0, "reduced Hydra with only Islands")

		addBasicLandPermanent(g, game.Player1, types.Forest)
		assertExileXRange(t, g, engine, game.Player1, hydra, 5, "reduced Hydra with a Forest")
	})

	t.Run("Elderwood Scion target-gated reduction", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		scion := groupExileRegistryPermanent(t, g, game.Player1, "Elderwood Scion")
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		blaze := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Blaze"))
		g.AddExileCounter(blaze, counter.Croak, 1)

		atScion := []game.Target{game.PermanentTarget(scion.ObjectID)}
		atOpponent := []game.Target{game.PlayerTarget(game.Player2)}
		assertExileTargetedXRange(t, g, engine, game.Player1, blaze, atScion, 3, "Blaze at Elderwood Scion")
		assertExileTargetedXRange(t, g, engine, game.Player1, blaze, atOpponent, 1, "Blaze at an opponent")

		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(blaze, zone.Exile, game.FaceFront, atScion, 2, nil)) {
			t.Fatal("casting the enumerated reduced Blaze for X=2 failed")
		}
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID != blaze || top.XValue != 2 || len(top.Targets) != 1 || top.Targets[0].PermanentID != scion.ObjectID {
			t.Fatalf("stack top = %+v, want Blaze with X=2 targeting Elderwood Scion", top)
		}
	})
}

// activateRestrictedManaAbility gives player priority and activates source's
// mana ability, leaving its spend-restricted mana in player's pool.
func activateRestrictedManaAbility(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, source *game.Permanent) {
	t.Helper()
	giveMainPhasePriority(g, player)
	if !engine.applyAction(g, player, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatalf("activating permanent %v's mana ability failed", source.ObjectID)
	}
}

// TestExileXCastSpendsSpellRestrictedMana covers CR 106.6 and 601.2g-h for X
// spells cast from exile: mana whose spending is restricted to a kind of spell
// pays any part of a qualifying spell's total cost, so every X it affords is
// offered, with or without a cost reduction, and it never raises the X offered
// for a spell outside its restriction. Humble Naturalist's creature-only mana
// plus one land casts Mistcutter Hydra for X up to 1 under Grolnok and under
// Evelyn, and for X up to 3 with Heartless Summoning. Vodalian Arcanist's
// instant-or-sorcery mana raises Blaze's X but leaves Grolnok's Hydra at X=0.
//
// Source review (Comprehensive Rules, Mistcutter Hydra, Heartless Summoning
// and Blaze as cited above):
//   - CR 106.6: some abilities that produce mana restrict how that mana can be
//     spent. CR 601.2g-h: mana abilities are activated before the total cost is
//     paid, in any combination of mana that pays it.
//   - Humble Naturalist, printing 8f56705d-eb64-4cef-b716-edbeac60bf79
//     (oracle_id ff3a69a7-a244-4958-9e8b-d153942c1ba6), {1}{G}: "{T}: Add one
//     mana of any color. Spend this mana only to cast a creature spell."
//     Ruling 2020-04-17: the mana can be spent on any part of a creature
//     spell's total cost, including additional and alternative costs.
//   - Vodalian Arcanist, printing a52e90c0-b012-4ce5-8462-1e33c7143de5
//     (oracle_id 1d07856b-d9ef-418c-bd06-3888dbcaddf4), {1}{U}: "{T}: Add {C}.
//     Spend this mana only to cast an instant or sorcery spell." No rulings.
func TestExileXCastSpendsSpellRestrictedMana(t *testing.T) {
	t.Run("Humble Naturalist under Grolnok", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		naturalist := groupExileRegistryPermanent(t, g, game.Player1, "Humble Naturalist")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		hydra := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounter(hydra, counter.Croak, 1)
		activateRestrictedManaAbility(t, g, engine, game.Player1, naturalist)

		assertExileXRange(t, g, engine, game.Player1, hydra, 2, "Hydra with Naturalist mana")

		if !engine.applyAction(g, game.Player1, groupExileCastX(hydra, 1)) {
			t.Fatal("casting the enumerated Hydra for X=1 failed")
		}
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID != hydra || top.XValue != 1 || !g.Players[game.Player1].ManaPool.IsEmpty() {
			t.Fatalf("stack top = %+v, pool = %+v; want Hydra with X=1 paid with the Naturalist mana", top, g.Players[game.Player1].ManaPool)
		}
	})

	t.Run("Humble Naturalist and Heartless Summoning under Grolnok", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		groupExileRegistryPermanent(t, g, game.Player1, "Heartless Summoning")
		naturalist := groupExileRegistryPermanent(t, g, game.Player1, "Humble Naturalist")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		hydra := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounter(hydra, counter.Croak, 1)
		activateRestrictedManaAbility(t, g, engine, game.Player1, naturalist)

		assertExileXRange(t, g, engine, game.Player1, hydra, 4, "reduced Hydra with Naturalist mana")
	})

	t.Run("Humble Naturalist under Evelyn", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
		naturalist := groupExileRegistryPermanent(t, g, game.Player1, "Humble Naturalist")
		addBasicLandPermanent(g, game.Player1, types.Island)
		hydra := addCardToExile(g, game.Player2, groupExileRegistryDef(t, "Mistcutter Hydra"))
		g.AddExileCounterFromController(hydra, counter.Collection, 1, game.Player1)
		activateRestrictedManaAbility(t, g, engine, game.Player1, naturalist)

		assertExileXRange(t, g, engine, game.Player1, hydra, 2, "foreign-exile Hydra with Naturalist mana")
	})

	t.Run("Vodalian Arcanist pays only instants and sorceries", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		groupExileRegistryPermanent(t, g, game.Player1, "Grolnok, the Omnivore")
		arcanist := groupExileRegistryPermanent(t, g, game.Player1, "Vodalian Arcanist")
		addBasicLandPermanent(g, game.Player1, types.Forest)
		addBasicLandPermanent(g, game.Player1, types.Mountain)
		hydra := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Mistcutter Hydra"))
		blaze := addCardToExile(g, game.Player1, groupExileRegistryDef(t, "Blaze"))
		g.AddExileCounter(hydra, counter.Croak, 1)
		g.AddExileCounter(blaze, counter.Croak, 1)
		activateRestrictedManaAbility(t, g, engine, game.Player1, arcanist)

		assertExileXRange(t, g, engine, game.Player1, hydra, 2, "Hydra beside Arcanist mana")
		atOpponent := []game.Target{game.PlayerTarget(game.Player2)}
		assertExileTargetedXRange(t, g, engine, game.Player1, blaze, atOpponent, 3, "Blaze with Arcanist mana")
	})
}

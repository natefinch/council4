package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestTokenCreationReplacementDoublesTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 2, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 4 {
		t.Fatalf("created tokens = %d, want 4", got)
	}
}

func TestTokenCreationReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("created tokens after source leaves = %d, want 1", got)
	}
}

func TestTokenCreationEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Zombie", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 2, true, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	created := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Zombie" {
			created++
			if !permanent.Tapped {
				t.Fatal("token Tapped = false, want true")
			}
		}
	}
	if created != 2 {
		t.Fatalf("created tokens = %d, want 2", created)
	}
}

func TestTokenCreationEntersUntappedByDefault(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Zombie", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Zombie" && permanent.Tapped {
			t.Fatal("token Tapped = true, want false")
		}
	}
}

func TestTokenCreationReplacementStacksAndRecordsOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 4 {
		t.Fatalf("created tokens = %d, want 4", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player1 {
		t.Fatalf("replacement decision player = %v, want Player1", got)
	}
}

func TestTokenCreationReplacementDoesNotAffectOpponentTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices() = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("opponent-created tokens = %d, want 1", got)
	}
}

func TestTokenCreationReplacementUsesCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
		Duration:         game.DurationPermanent,
	})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token", Types: []types.Card{types.Creature}}}

	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player1, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player1) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 1 {
		t.Fatalf("old controller tokens = %d, want 1", got)
	}
	if !createTokenPermanentsWithChoices(NewEngine(nil), g, game.Player2, token, 1, false, [game.NumPlayers]PlayerAgent{}, nil) {
		t.Fatal("createTokenPermanentsWithChoices(Player2) = false, want true")
	}
	if got := countTokenPermanentsNamed(g, "Soldier Token"); got != 3 {
		t.Fatalf("tokens after new controller creates one = %d, want 3", got)
	}
}

func TestCounterPlacementReplacementAddsToSpecificCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, hardenedScalesReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
	if !addCountersToPermanent(g, creature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanent(stun) = false, want true")
	}
	if got := creature.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("stun counters = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementDoublesSpecificCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 2) {
		t.Fatal("addCountersToPermanent(+1/+1) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4", got)
	}
	if !addCountersToPermanent(g, creature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanent(stun) = false, want true")
	}
	if got := creature.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("stun counters = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementDoublesETBCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Entering Creature",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersWithCountersReplacement(
				"Entering Creature enters with a +1/+1 counter on it.",
				game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1},
			),
		},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("entering card instance missing")
	}

	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent() = false, want true")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("ETB +1/+1 counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesAllCounterKinds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanent(stun) = false, want true")
	}
	if got := creature.Counters.Get(counter.Stun); got != 2 {
		t.Fatalf("stun counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementUsesPlacingController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})
	controllerCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controller Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanentControlledBy(g, game.Player1, opponentCreature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanentControlledBy(Player1) = false, want true")
	}
	if got := opponentCreature.Counters.Get(counter.Stun); got != 2 {
		t.Fatalf("opponent creature stun counters = %d, want 2", got)
	}
	if !addCountersToPermanentControlledBy(g, game.Player2, controllerCreature, counter.Stun, 1) {
		t.Fatal("addCountersToPermanentControlledBy(Player2) = false, want true")
	}
	if got := controllerCreature.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("controller creature stun counters from opponent = %d, want 1", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesProliferatedCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})
	opponentCreature.Counters.Add(counter.Stun, 1)
	g.Players[game.Player2].PoisonCounters = 1

	if !addProliferatedCounter(g, game.Player1, proliferateTarget{
		permanentID: opponentCreature.ObjectID,
		counters:    []counter.Kind{counter.Stun},
	}, counter.Stun) {
		t.Fatal("addProliferatedCounter(permanent) = false, want true")
	}
	if got := opponentCreature.Counters.Get(counter.Stun); got != 3 {
		t.Fatalf("proliferated stun counters = %d, want 3", got)
	}
	if !addProliferatedCounter(g, game.Player1, proliferateTarget{
		player:   game.Player2,
		counters: []counter.Kind{counter.Poison},
	}, counter.Poison) {
		t.Fatal("addProliferatedCounter(player) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 3 {
		t.Fatalf("proliferated poison counters = %d, want 3", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesToxicCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Toxic Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			{KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: 1}}},
		},
	}})

	markPlayerCombatDamage(g, source, game.Player2, 1, &TurnLog{})
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("toxic poison counters = %d, want 2", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesWitherDamageCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Wither)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 7)

	dealt := dealPermanentDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, target, 3, false)
	if dealt != 3 {
		t.Fatalf("damage dealt = %d, want 3", dealt)
	}
	if got := target.Counters.Get(counter.MinusOneMinusOne); got != 6 {
		t.Fatalf("-1/-1 counters = %d, want 6", got)
	}
}

func TestAnyCounterPlacementReplacementDoublesPlayerCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())

	if !addCountersToPlayerControlledBy(g, game.Player1, g.Players[game.Player2], counter.Poison, 1) {
		t.Fatal("addCountersToPlayerControlledBy(Player1) = false, want true")
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("poison counters = %d, want 2", got)
	}
}

func TestCounterPlacementReplacementOnlyMatchesCreatureRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Noncreature Artifact",
		Types: []types.Card{types.Artifact},
	}})

	if !addCountersToPermanent(g, artifact, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(noncreature) = false, want true")
	}
	if got := artifact.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("noncreature +1/+1 counters = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementUsesRecipientController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Controller Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanentControlledBy(g, game.Player2, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanentControlledBy(opponent) = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters from opponent effect = %d, want 2", got)
	}
}

func TestCounterPlacementReplacementStacksAndRecordsOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("+1/+1 counters = %d, want 4", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player1 {
		t.Fatalf("replacement decision player = %v, want Player1", got)
	}
}

func TestCounterPlacementReplacementExpiresWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Creature",
		Types: []types.Card{types.Creature},
	}})
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}

	if !addCountersToPermanent(g, creature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after source leaves = %d, want 1", got)
	}
}

func TestCounterPlacementReplacementUsesCurrentController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		NewController:    opt.Val(game.Player2),
		Duration:         game.DurationPermanent,
	})
	oldControllerCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Old Creature",
		Types: []types.Card{types.Creature},
	}})
	newControllerCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "New Creature",
		Types: []types.Card{types.Creature},
	}})

	if !addCountersToPermanent(g, oldControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(old controller) = false, want true")
	}
	if got := oldControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("old controller +1/+1 counters = %d, want 1", got)
	}
	if !addCountersToPermanent(g, newControllerCreature, counter.PlusOnePlusOne, 1) {
		t.Fatal("addCountersToPermanent(new controller) = false, want true")
	}
	if got := newControllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("new controller +1/+1 counters = %d, want 2", got)
	}
}

func tokenDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Anointed Procession",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.TokenCreationReplacement(
				"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
				2,
				game.TriggerControllerYou,
			),
		},
	}}
}

func counterDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Branching Evolution",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.CounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
				2,
				0,
				counter.PlusOnePlusOne,
				game.TriggerControllerYou,
			),
		},
	}}
}

func hardenedScalesReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Hardened Scales",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.CounterPlacementReplacement(
				"If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.",
				0,
				1,
				counter.PlusOnePlusOne,
				game.TriggerControllerYou,
			),
		},
	}}
}

func anyCounterDoublingReplacementCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Vorinclex",
		Types: []types.Card{types.Creature},
		ReplacementAbilities: []game.ReplacementAbility{
			game.AnyCounterPlacementReplacement(
				"If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
				2,
				0,
				game.TriggerControllerYou,
			),
		},
	}}
}

func countTokenPermanentsNamed(g *game.Game, name string) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == name {
			count++
		}
	}
	return count
}

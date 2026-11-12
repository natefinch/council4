package rules

import (
	"strings"
	"testing"

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

// Source review: issue #3756 (sub-issue of #2351), cast-permission entry
// counter rider. CR effective 2025-11-14 (official text from media.wizards.com,
// "MagicCompRules 20251114.txt"): 601.2b (the caster announces how a spell is
// cast; earlier choices such as flashback restrict later ones), 122.6/122.6a
// (an object given counters as it enters), 614.1c/614.12 ("enters with"
// replacement effects may come from other sources), 707.10 (a copy of a spell
// is not cast), 122.1h (finality counters). Hundred-Battle Veteran (TDM,
// 9ed6aa5b-4369-42f7-9853-aae1ac9b17e2): "You may cast this card from your
// graveyard. If you do, it enters with a finality counter on it." Rulings
// 2025-04-04 (https://api.scryfall.com/cards/e53adf93-2db5-4087-a2dc-c8f53401d700/rulings):
// all costs and timing still apply; if countered after being cast from the
// graveyard it returns to the graveyard and may be cast this way again.
// Leonardo, Sewer Samurai (TMT, 74e3243b-6707-4769-8767-7cf2ef71009a): "During
// your turn, you may cast creature spells with power or toughness 1 or less
// from your graveyard. If you cast a spell this way, that creature enters with
// a finality counter on it." Rulings 2026-01-27
// (https://api.scryfall.com/cards/0e3a0a26-c163-4f31-bed8-1be52a35feea/rulings):
// all costs and timing rules apply. Neither card is admitted here: the
// test-local shapes mirror only the rider (self and granted graveyard
// permissions) to prove the shared rule; Leonardo's turn and power/toughness
// filters and the Veteran's counter-kinds static stay with their card units.

// hundredBattleVeteranShape mirrors Hundred-Battle Veteran's self graveyard
// cast permission and its finality rider.
func hundredBattleVeteranShape() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Hundred-Battle Veteran shape",
			ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.B}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{{
				Text:           "You may cast this card from your graveyard. If you do, it enters with a finality counter on it.",
				ZoneOfFunction: zone.Graveyard,
				RuleEffects: []game.RuleEffect{{
					Kind:              game.RuleEffectCastFromZone,
					AffectedSource:    true,
					Duration:          game.DurationPermanent,
					AffectedPlayer:    game.PlayerYou,
					CastFromZone:      zone.Graveyard,
					CastEntryCounters: []game.CounterPlacement{{Kind: counter.Finality, Amount: 1}},
				}},
			}},
		},
	}
}

// leonardoShape mirrors Leonardo, Sewer Samurai's granted graveyard permission
// for creature spells and its finality rider.
func leonardoShape() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Leonardo shape",
			ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.W}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{{
				Text: "You may cast creature spells from your graveyard. If you cast a spell this way, that creature enters with a finality counter on it.",
				RuleEffects: []game.RuleEffect{{
					Kind:              game.RuleEffectCastSpellsFromZone,
					Duration:          game.DurationPermanent,
					AffectedPlayer:    game.PlayerYou,
					SpellTypes:        []types.Card{types.Creature},
					CastFromZone:      zone.Graveyard,
					CastEntryCounters: []game.CounterPlacement{{Kind: counter.Finality, Amount: 1}},
				}},
			}},
		},
	}
}

func addAllMana(g *game.Game, playerID game.PlayerID, amount int) {
	for _, symbol := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G} {
		g.Players[playerID].ManaPool.Add(symbol, amount)
	}
}

func castFromGraveyardOrFail(t *testing.T, g *game.Game, engine *Engine, cast action.Action) *game.StackObject {
	t.Helper()
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("cast from graveyard failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("cast spell is not on the stack")
	}
	return obj
}

func finalityOnCard(t *testing.T, g *game.Game, cardID id.ID) int {
	t.Helper()
	permanent := permanentByCardID(g, cardID)
	if permanent == nil {
		t.Fatal("cast permanent is not on the battlefield")
	}
	return permanent.Counters.Get(counter.Finality)
}

func TestCastEntryCounterRiderSelfGraveyardCastEntersWithFinality(t *testing.T) {
	g, engine := finalityMainPhase()
	veteran := addCardToGraveyard(g, game.Player1, hundredBattleVeteranShape())
	addAllMana(g, game.Player1, 4)
	castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, veteran); n != 1 {
		t.Fatalf("graveyard-cast finality counters = %d, want 1", n)
	}
	permanent := permanentByCardID(g, veteran)
	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(permanent.ObjectID)})
	requireFinalityExiled(t, g, permanent)

	g, engine = finalityMainPhase()
	fromHand := addCardToHand(g, game.Player1, hundredBattleVeteranShape())
	addAllMana(g, game.Player1, 4)
	if !engine.applyAction(g, game.Player1, action.CastSpell(fromHand, nil, 0, nil)) {
		t.Fatal("cast from hand failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, fromHand); n != 0 {
		t.Fatalf("hand-cast finality counters = %d, want 0", n)
	}
}

func TestCastEntryCounterRiderGrantedPermissionSurvivesGranterLeaving(t *testing.T) {
	g, engine := finalityMainPhase()
	bears := addCardToGraveyard(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"))
	cast := action.CastSpellFromZone(bears, zone.Graveyard, nil, 0, nil)
	addAllMana(g, game.Player1, 2)
	if containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("creature card was castable from the graveyard without a permission")
	}
	leonardo := addCombatPermanent(g, game.Player1, leonardoShape())
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("granted graveyard permission was not offered as a legal cast")
	}
	castFromGraveyardOrFail(t, g, engine, cast)

	castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
		[]game.Target{game.PermanentTarget(leonardo.ObjectID)})
	if _, ok := permanentByObjectID(g, leonardo.ObjectID); ok {
		t.Fatal("permission source is still on the battlefield")
	}
	setMainPhasePriority(g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, bears); n != 1 {
		t.Fatalf("finality counters after granter left = %d, want 1", n)
	}
}

func TestCastEntryCounterRiderMuldrothaRouteTakesNoCounter(t *testing.T) {
	g, engine, source := muldrothaGame(t)
	addCombatPermanent(g, game.Player1, leonardoShape())
	scopedCard := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
	riderCard := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))

	castFromGraveyardOrFail(t, g, engine, graveyardCast(scopedCard, source, types.Creature))
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, scopedCard); n != 0 || len(g.PermanentTypePlayUsesThisTurn) != 1 {
		t.Fatalf("Muldrotha cast finality counters = %d, uses = %d; want 0 and 1", n, len(g.PermanentTypePlayUsesThisTurn))
	}
	castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(riderCard, zone.Graveyard, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, riderCard); n != 1 || len(g.PermanentTypePlayUsesThisTurn) != 1 {
		t.Fatalf("rider cast finality counters = %d, uses = %d; want 1 and 1", n, len(g.PermanentTypePlayUsesThisTurn))
	}
}

func TestCastEntryCounterRiderCounteredSpellReturnsToGraveyard(t *testing.T) {
	g, engine := finalityMainPhase()
	veteran := addCardToGraveyard(g, game.Player1, hundredBattleVeteranShape())
	addAllMana(g, game.Player1, 8)
	obj := castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
	if !counterStackObject(g, obj.ID) {
		t.Fatal("countering the graveyard-cast spell failed")
	}
	player := g.Players[game.Player1]
	if !player.Graveyard.Contains(veteran) || player.Exile.Contains(veteran) || permanentByCardID(g, veteran) != nil {
		t.Fatal("countered rider spell did not go to its owner's graveyard")
	}
	castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, veteran); n != 1 {
		t.Fatalf("recast finality counters = %d, want 1", n)
	}
}

func TestCastEntryCounterRiderSurvivesCloneAndReplay(t *testing.T) {
	run := func() (*game.Game, *game.Game, id.ID) {
		g, engine := finalityMainPhase()
		veteran := addCardToGraveyard(g, game.Player1, hundredBattleVeteranShape())
		addAllMana(g, game.Player1, 4)
		obj := castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
		branch := g.Clone()
		obj.CastPermissionEntryCounters[0].Amount = 3
		engine.resolveTopOfStack(branch, &TurnLog{})
		engine.resolveTopOfStack(g, &TurnLog{})
		return g, branch, veteran
	}
	original, branch, veteran := run()
	if n := finalityOnCard(t, branch, veteran); n != 1 {
		t.Fatalf("cloned game finality counters = %d, want 1", n)
	}
	if n := finalityOnCard(t, original, veteran); n != 3 {
		t.Fatalf("original game finality counters = %d, want its own rider", n)
	}
	_, replay, replayVeteran := run()
	if finalityOnCard(t, replay, replayVeteran) != 1 || len(replay.Events) != len(branch.Events) {
		t.Fatal("replayed rider cast diverged")
	}
}

func TestCastEntryCounterRiderSpellRoutes(t *testing.T) {
	t.Run("stack copy is not cast", func(t *testing.T) {
		g, engine := finalityMainPhase()
		veteran := addCardToGraveyard(g, game.Player1, hundredBattleVeteranShape())
		addAllMana(g, game.Player1, 4)
		obj := castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
		g.Stack.Push(game.NewStackObjectCopy(obj, g.IDGen.Next()))
		before := len(g.Battlefield)
		engine.resolveTopOfStack(g, &TurnLog{})
		if len(g.Battlefield) != before+1 {
			t.Fatal("spell copy did not resolve to a permanent")
		}
		if token := g.Battlefield[len(g.Battlefield)-1]; token.CardInstanceID == veteran || token.Counters.Get(counter.Finality) != 0 {
			t.Fatal("copy of the rider spell entered with the cast rider's counter")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if n := finalityOnCard(t, g, veteran); n != 1 {
			t.Fatalf("original finality counters = %d, want 1", n)
		}
	})

	t.Run("kicker combines with rider", func(t *testing.T) {
		g, engine := finalityMainPhase()
		def := hundredBattleVeteranShape()
		def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.G}}},
		})
		veteran := addCardToGraveyard(g, game.Player1, def)
		addAllMana(g, game.Player1, 4)
		obj := castFromGraveyardOrFail(t, g, engine, action.CastKickedSpellFromZone(veteran, zone.Graveyard, nil, 0, nil))
		if !obj.KickerPaid {
			t.Fatal("kicker was not paid on the rider route")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if n := finalityOnCard(t, g, veteran); n != 1 {
			t.Fatalf("kicked rider cast finality counters = %d, want 1", n)
		}
	})

	t.Run("flashback cast takes no rider", func(t *testing.T) {
		flashbackCreature := func() *game.CardDef {
			def := permanentWitness(types.Creature)
			def.ManaCost = opt.Val(cost.Mana{cost.G, cost.G})
			def.StaticAbilities = []game.StaticAbility{{
				Text:             "Flashback {B}",
				KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.B}}},
			}}
			return def
		}
		for _, chooseFlashback := range []bool{true, false} {
			g, engine := finalityMainPhase()
			addCombatPermanent(g, game.Player1, leonardoShape())
			card := addCardToGraveyard(g, game.Player1, flashbackCreature())
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			g.Players[game.Player1].ManaPool.Add(mana.G, 2)
			offered := 0
			agent := &muldrothaChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoicePayment {
					return request.DefaultSelection
				}
				offered = len(request.Options)
				for _, option := range request.Options {
					if strings.Contains(option.Label, "lashback") == chooseFlashback {
						return []int{option.Index}
					}
				}
				return request.DefaultSelection
			}}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("cast from graveyard failed")
			}
			obj, _ := g.Stack.Peek()
			if offered < 2 || obj.Flashback != chooseFlashback {
				t.Fatalf("offered %d cast choices, flashback = %v; want both routes and flashback = %v", offered, obj.Flashback, chooseFlashback)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			want := 1
			if chooseFlashback {
				want = 0
			}
			if n := finalityOnCard(t, g, card); n != want {
				t.Fatalf("flashback=%v finality counters = %d, want %d", chooseFlashback, n, want)
			}
		}
	})

	t.Run("mutate with illegal target enters with rider", func(t *testing.T) {
		g, engine := finalityMainPhase()
		addCombatPermanent(g, game.Player1, leonardoShape())
		bears := addCombatPermanent(g, game.Player1, finalityRegistryCard(t, "Grizzly Bears"))
		card := addCardToGraveyard(g, game.Player1, mutateCard())
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		castFromGraveyardOrFail(t, g, engine, action.CastMutateSpellFromZone(card, zone.Graveyard, bears.ObjectID))
		castAndResolveFinalitySpell(t, g, engine, game.Player2, finalityRegistryCard(t, "Murder"), zone.Hand,
			[]game.Target{game.PermanentTarget(bears.ObjectID)})
		setMainPhasePriority(g, game.Player1)
		engine.resolveTopOfStack(g, &TurnLog{})
		if n := finalityOnCard(t, g, card); n != 1 {
			t.Fatalf("mutate spell with illegal target finality counters = %d, want 1", n)
		}
	})
}

// A group graveyard permission authorizes only the casts whose spell it
// matches: a bestowed cast is an Aura spell, not a creature spell (CR
// 702.103b), so it is selected through Muldrotha's enchantment permission
// rather than the creature-spell rider permission, while a creature cast of the
// same card still uses the rider permission without spending Muldrotha.
func TestCastEntryCounterRiderBestowedCastSelectsPermissionsByCastMode(t *testing.T) {
	g, engine, source := muldrothaGame(t)
	addCombatPermanent(g, game.Player1, leonardoShape())
	card := addCardToGraveyard(g, game.Player1, bestowTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	targets := []game.Target{game.PermanentTarget(source.ObjectID)}
	bestow := action.CastBestowSpellFaceFromZone(card, zone.Graveyard, game.FaceFront, targets, 0, nil)
	enchantment := muldrothaPermission(source, types.Enchantment)
	legal := engine.legalActions(g, game.Player1)
	if containsAction(legal, bestow) || engine.castActionPermissionLegal(g, game.Player1, mustCastPayload(t, bestow), game.PlayPermission{}) {
		t.Fatal("bestowed cast was authorized by the creature-spell permission")
	}
	if !containsAction(legal, bestow.WithPlayPermission(enchantment)) {
		t.Fatal("bestowed cast was not offered through Muldrotha's enchantment permission")
	}
	if creatureCast := action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil); !containsAction(legal, creatureCast) {
		t.Fatal("creature cast was not offered through the creature-spell permission")
	}

	obj := castFromGraveyardOrFail(t, g, engine, bestow)
	if !obj.Bestowed || len(obj.CastPermissionEntryCounters) != 0 {
		t.Fatalf("bestowed = %v, rider = %v; want a bestowed spell without the creature permission's rider", obj.Bestowed, obj.CastPermissionEntryCounters)
	}
	if !g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: enchantment, Player: game.Player1}] {
		t.Fatal("bestowed cast did not spend Muldrotha's enchantment permission")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	next := addCardToGraveyard(g, game.Player1, permanentWitness(types.Enchantment))
	if engine.applyAction(g, game.Player1, graveyardCast(next, source, types.Enchantment)) {
		t.Fatal("Muldrotha allowed a second enchantment spell this turn")
	}

	creature := addCardToGraveyard(g, game.Player1, permanentWitness(types.Creature))
	castFromGraveyardOrFail(t, g, engine, action.CastSpellFromZone(creature, zone.Graveyard, nil, 0, nil))
	engine.resolveTopOfStack(g, &TurnLog{})
	if n := finalityOnCard(t, g, creature); n != 1 ||
		g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: muldrothaPermission(source, types.Creature), Player: game.Player1}] {
		t.Fatalf("creature cast finality counters = %d; want 1 through the rider permission without spending Muldrotha", n)
	}
}

func mustCastPayload(t *testing.T, act action.Action) action.CastSpellAction {
	t.Helper()
	cast, ok := act.CastSpellPayload()
	if !ok {
		t.Fatal("action is not a cast")
	}
	return cast
}

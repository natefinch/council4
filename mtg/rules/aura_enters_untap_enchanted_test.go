package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#2119): pinned corpus 6da9ced0 Oracle text for Howl of the
// Hunt (22fe038e-033e-4632-a0d7-116492cded83) and Volition Reins
// (bda496f9-5362-46f2-ad4f-cd96410c3c61). Scryfall rulings retrieved
// 2026-10-31: Howl of the Hunt returned an empty list; Volition Reins returned
// three 2011-01-01 rulings, including that the trigger doesn't trigger at all
// when the enchanted permanent is untapped as the Aura enters. Official CR
// effective 2026-09-25 (media.wizards.com MagicCompRules 20260925.txt) read:
// 303.4b ("enchanted" names the object the Aura is attached to), 303.4e (Aura
// and enchanted-object control are separate), 303.4i (illegal attachment from
// the stack goes to the graveyard), 603.4 (intervening "if" checked on trigger
// and resolution), 603.6a (enters abilities), 608.3c (an Aura spell is put onto
// the battlefield attached to its target), 707.10 (a copied Aura spell becomes
// a token attached the same way), 613.1b (control changes in layer 2), 701.26b
// (untap), 113.7a (last-known information). The Aura is attached as it enters, so the enters trigger checks the
// enchanted permanent. "That creature"/"it" in these effects
// co-refer with the enchanted permanent, not the Aura that entered. Both cards
// are legacy name-only registry entries, so tests look them up by name.
const auraEntersUntapTestHowlCMC = 3

func lookupAuraEntersUntapCard(t *testing.T, name string) *game.CardDef {
	t.Helper()
	defs := cards.NewDefaultRegistry().LookupAll(name)
	if len(defs) != 1 || defs[0].Name != name {
		t.Fatalf("default registry %s lookup = %d defs, want exactly one", name, len(defs))
	}
	return defs[0]
}

// castAuraEntersUntap casts the Aura onto target, resolves it, and reports
// whether its enters trigger went on the stack.
func castAuraEntersUntap(t *testing.T, g *game.Game, engine *Engine, auraID id.ID, target *game.Permanent) (*game.Permanent, bool) {
	t.Helper()
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(auraID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatal("casting the Aura failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	aura, ok := findPermanentByCardID(g, auraID)
	if !ok {
		t.Fatal("Aura did not enter the battlefield")
	}
	if !aura.AttachedTo.Exists || aura.AttachedTo.Val != target.ObjectID {
		t.Fatalf("Aura attached to %v, want %v", aura.AttachedTo, target.ObjectID)
	}
	triggered := engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if triggered {
		top, ok := g.Stack.Peek()
		if !ok || top.SourceID != aura.ObjectID {
			t.Fatal("Aura enters trigger is not on top of the stack")
		}
	}
	return aura, triggered
}

func howlScenario(t *testing.T, subtype types.Sub) (*game.Game, *Engine, id.ID, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	auraID := addCardToHand(g, game.Player1, lookupAuraEntersUntapCard(t, "Howl of the Hunt"))
	g.Players[game.Player1].ManaPool.Add(mana.G, auraEntersUntapTestHowlCMC)
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Howl test creature",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{subtype},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	creature.Tapped = true
	return g, NewEngine(nil), auraID, creature
}

func TestHowlOfTheHuntUntapsEnchantedWolfNotTheAura(t *testing.T) {
	for _, subtype := range []types.Sub{types.Wolf, types.Werewolf} {
		t.Run(string(subtype), func(t *testing.T) {
			g, engine, auraID, creature := howlScenario(t, subtype)
			aura, triggered := castAuraEntersUntap(t, g, engine, auraID, creature)
			if !triggered {
				t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf or Werewolf")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if creature.Tapped {
				t.Fatal("enchanted creature stayed tapped; untap bound to the entering Aura")
			}
			if aura.Tapped {
				t.Fatal("Howl of the Hunt itself became tapped")
			}
			if !g.Stack.IsEmpty() {
				t.Fatal("stack not empty after the trigger resolved")
			}
		})
	}
}

func TestHowlOfTheHuntDoesNotTriggerForOtherCreatures(t *testing.T) {
	g, engine, auraID, creature := howlScenario(t, types.Bear)
	if _, triggered := castAuraEntersUntap(t, g, engine, auraID, creature); triggered {
		t.Fatal("Howl of the Hunt triggered for a non-Wolf, non-Werewolf creature")
	}
	if !creature.Tapped {
		t.Fatal("non-Wolf creature was untapped")
	}
}

func volitionScenario(t *testing.T, tapped bool) (*game.Game, *Engine, id.ID, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	auraID := addCardToHand(g, game.Player1, lookupAuraEntersUntapCard(t, "Volition Reins"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 6)
	land := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Volition test land",
		Types: []types.Card{types.Land},
	}})
	land.Tapped = tapped
	return g, NewEngine(nil), auraID, land
}

func TestVolitionReinsStealsAndUntapsEnchantedPermanent(t *testing.T) {
	g, engine, auraID, land := volitionScenario(t, true)
	aura, triggered := castAuraEntersUntap(t, g, engine, auraID, land)
	if !triggered {
		t.Fatal("Volition Reins did not trigger for a tapped enchanted permanent")
	}
	if got := effectiveController(g, land); got != game.Player1 {
		t.Fatalf("enchanted land controller = %v, want Player1", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if land.Tapped {
		t.Fatal("enchanted permanent stayed tapped; untap bound to the entering Aura")
	}
	if aura.Tapped {
		t.Fatal("Volition Reins itself became tapped")
	}
	if got := effectiveController(g, land); got != game.Player1 {
		t.Fatalf("enchanted land controller after trigger = %v, want Player1", got)
	}
}

func TestVolitionReinsDoesNotTriggerForUntappedPermanent(t *testing.T) {
	g, engine, auraID, land := volitionScenario(t, false)
	if _, triggered := castAuraEntersUntap(t, g, engine, auraID, land); triggered {
		t.Fatal("Volition Reins triggered for an untapped enchanted permanent")
	}
	if got := effectiveController(g, land); got != game.Player1 {
		t.Fatalf("enchanted land controller = %v, want Player1", got)
	}
}

func TestHowlOfTheHuntCopiedSpellEntersAttachedAndUntapsEnchantedWolf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	wolf := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Howl copy test Wolf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	wolf.Tapped = true
	source := &game.CardInstance{ID: g.IDGen.Next(), Def: lookupAuraEntersUntapCard(t, "Howl of the Hunt"), Owner: game.Player1}
	g.CardInstances[source.ID] = source
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     source.ID,
		Controller:   game.Player1,
		Copy:         true,
		Targets:      []game.Target{game.PermanentTarget(wolf.ObjectID)},
		TargetCounts: []int{1},
	})
	engine := NewEngine(nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			token = permanent
		}
	}
	if token == nil || !token.AttachedTo.Exists || token.AttachedTo.Val != wolf.ObjectID {
		t.Fatal("copied Howl of the Hunt did not become a token attached to the Wolf")
	}
	if !permanentIDsContain(wolf.Attachments, token.ObjectID) {
		t.Fatal("Wolf does not record the copied Aura token as an attachment")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("copied Howl of the Hunt did not trigger for its enchanted Wolf")
	}
	if top, ok := g.Stack.Peek(); !ok || top.SourceID != token.ObjectID {
		t.Fatal("copied Aura enters trigger is not on top of the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if wolf.Tapped {
		t.Fatal("enchanted Wolf stayed tapped after the copied Aura's trigger")
	}
}

// An Aura removed in response to its enters trigger still names the permanent
// it last enchanted (CR 113.7a); the intervening "if" is rechecked against that
// permanent's current state (CR 603.4) and the effect untaps it.
func TestAuraEntersUntapUsesLastKnownAttachmentWhenAuraLeaves(t *testing.T) {
	t.Run("Howl of the Hunt", func(t *testing.T) {
		g, engine, auraID, wolf := howlScenario(t, types.Wolf)
		aura, triggered := castAuraEntersUntap(t, g, engine, auraID, wolf)
		if !triggered {
			t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf")
		}
		if !movePermanentToZone(g, aura, zone.Graveyard) {
			t.Fatal("failed to remove Howl of the Hunt in response")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if wolf.Tapped {
			t.Fatal("formerly enchanted Wolf stayed tapped after the Aura left in response")
		}
	})
	t.Run("Volition Reins", func(t *testing.T) {
		g, engine, auraID, land := volitionScenario(t, true)
		aura, triggered := castAuraEntersUntap(t, g, engine, auraID, land)
		if !triggered {
			t.Fatal("Volition Reins did not trigger for a tapped enchanted permanent")
		}
		if !movePermanentToZone(g, aura, zone.Graveyard) {
			t.Fatal("failed to remove Volition Reins in response")
		}
		if got := effectiveController(g, land); got != game.Player2 {
			t.Fatalf("land controller after Volition Reins left = %v, want Player2", got)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if land.Tapped {
			t.Fatal("formerly enchanted permanent stayed tapped after the Aura left in response")
		}
	})
}

func TestAuraEntersUntapRechecksFormerlyEnchantedPermanentWhenAuraLeaves(t *testing.T) {
	g, engine, auraID, wolf := howlScenario(t, types.Wolf)
	aura, triggered := castAuraEntersUntap(t, g, engine, auraID, wolf)
	if !triggered {
		t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf")
	}
	if !movePermanentToZone(g, aura, zone.Graveyard) {
		t.Fatal("failed to remove Howl of the Hunt in response")
	}
	if !movePermanentToZone(g, wolf, zone.Graveyard) {
		t.Fatal("failed to remove the Wolf in response")
	}
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Howl bystander Wolf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	other.Tapped = true
	engine.resolveTopOfStack(g, &TurnLog{})
	if !other.Tapped {
		t.Fatal("trigger untapped a Wolf the Aura never enchanted")
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("stack not empty after the trigger resolved")
	}
}

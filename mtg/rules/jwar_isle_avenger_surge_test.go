package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle ID 2db8fc61-13af-4227-83f9-9f426ab978d3, Oath of the
// Gatewatch #58, printing revision 0a5b059f-2fa3-474f-9c74-6d4021703add;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle retrieved 2026-11-20 from
// https://api.scryfall.com/cards/0a5b059f-2fa3-474f-9c74-6d4021703add:
// "{4}{U} Creature — Sphinx, 3/3. Surge {2}{U} (You may cast this spell for its
// surge cost if you or a teammate has cast another spell this turn.) Flying".
// Rulings retrieved from
// https://api.scryfall.com/cards/0a5b059f-2fa3-474f-9c74-6d4021703add/rulings:
// four WotC rulings dated 2016-01-22; some cards have Surge payment riders,
// a prior spell may be resolved, countered, or (for instants) still on the
// stack, stack copies inherit Surge payment, and Surge does not change mana
// value.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 102.3-4, 806.1 and 903.2 (only the caster in free-for-all);
// 118.9-9d, 202.3, 601.2a-b,f-i and 702.117a (optional alternative, paid
// with applicable additional costs/modifiers, and gated on a completed
// earlier cast this turn; printed mana value is unchanged); 500.1 and 514.2
// (turn boundary); 702.9a-b
// (Flying); 305.1, 602.2, 701.6a, 707.10 and 707.12 (land plays,
// activations and spell copies aren't casts; countering doesn't undo casting,
// and a cast copy of a card is a new cast).
// The named-card scenario uses the registered authored definition; remaining
// Surge edge cases below exercise the shared operation independently.
func TestJwarIsleAvengerSurgeCostAfterCasterCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})

	avengerID := addCardToHand(g, game.Player1, authored.JwarIsleAvenger.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)

	cast := action.CastSpell(avengerID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) ||
		!engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Jwar Isle Avenger was not offered and castable for {2}{U} after the caster's prior spell")
	}
	obj, _ := g.Stack.Peek()
	if obj.AlternativeCost != 1 || obj.ManaSpentToCast != 3 {
		t.Fatalf("did not pay Surge: alternative=%d, mana spent=%d", obj.AlternativeCost, obj.ManaSpentToCast)
	}
	if value, ok := stackObjectManaValue(g, obj); !ok || value != 5 {
		t.Fatalf("stack mana value = %d (%t), want 5", value, ok)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == avengerID && permanent.Controller == game.Player1 && hasKeyword(g, permanent, game.Flying) {
			return
		}
	}
	t.Fatal("Jwar Isle Avenger did not enter under its caster's control with flying")
}

func surgeProbeAlternative() cost.Alternative {
	return cost.Alternative{
		Label: "Surge {2}{U}", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
		Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
	}
}

func surgeProbeInstant() *game.CardDef {
	def := simpleGainLifeInstant("Surge fixture")
	def.ManaCost = opt.Val(cost.Mana{cost.O(4), cost.U})
	def.AlternativeCosts = []cost.Alternative{surgeProbeAlternative()}
	return def
}

func castSurgeSetupSpell(t *testing.T, e *Engine, g *game.Game, caster game.PlayerID) id.ID {
	t.Helper()
	def := simpleGainLifeInstant("Earlier spell")
	def.ManaCost = opt.Val(cost.Mana{})
	card := addCardToHand(g, caster, def)
	if !e.applyAction(g, caster, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("failed to cast the setup spell")
	}
	return card
}

func TestSurgeRequiresCompletedCastByCasterThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	card := addCardToHand(g, game.Player1, surgeProbeInstant())
	cast := action.CastSpell(card, nil, 0, nil)
	fund := func() {
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	}
	check := func(stage string, want bool) {
		t.Helper()
		if got := containsAction(e.legalActions(g, game.Player1), cast); got != want {
			t.Fatalf("%s: Surge offered = %t, want %t", stage, got, want)
		}
	}

	fund()
	check("no prior cast", false)
	if e.applyAction(g, game.Player1, cast) || !g.Players[game.Player1].Hand.Contains(card) || g.Stack.Size() != 0 {
		t.Fatal("rejected announcement committed a cast")
	}
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Island", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Island},
	}})
	if !e.applyAction(g, game.Player1, action.PlayLand(land)) {
		t.Fatal("could not play the setup land")
	}
	check("land play and rejected cast", false)

	setMainPhasePriority(g, game.Player2)
	castSurgeSetupSpell(t, e, g, game.Player2)
	g.Turn.PriorityPlayer = game.Player1
	check("opponent's cast", false)
	opposingSpell, _ := g.Stack.Peek()
	copier := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Copy fixture", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					Allow: game.TargetAllowStackObject, MinTargets: 1, MaxTargets: 1,
					Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
				}},
				Sequence: []game.Instruction{{Primitive: game.CopyStackObject{Object: game.TargetStackObjectReference(0)}}},
			}.Ability(),
		}},
	}})
	if !e.applyAction(g, game.Player1, action.ActivateAbility(copier.ObjectID, 0,
		[]game.Target{game.StackObjectTarget(opposingSpell.ID)}, 0)) {
		t.Fatal("could not activate the copying ability")
	}
	check("activation without casting", false)
	e.resolveTopOfStack(g, &TurnLog{})
	copyObj, _ := g.Stack.Peek()
	if !copyObj.Copy || copyObj.Controller != game.Player1 {
		t.Fatal("the copying ability did not put a spell copy under Player1's control")
	}
	check("stack copy without casting", false)
	e.resolveTopOfStack(g, &TurnLog{})
	own := castSurgeSetupSpell(t, e, g, game.Player1)
	check("own cast still on the stack during an opponent's turn", true)
	obj, _ := g.Stack.Peek()
	if !counterStackObject(g, obj.ID) || !g.Players[game.Player1].Graveyard.Contains(own) {
		t.Fatal("setup spell was not countered")
	}
	check("countered own cast", true)
	e.resolveTopOfStack(g, &TurnLog{})
	e.advanceToNextTurn(g)
	setMainPhasePriority(g, g.Turn.ActivePlayer)
	g.Turn.PriorityPlayer = game.Player1
	g.Players[game.Player1].ManaPool.Empty()
	fund()
	check("next turn", false)
}

func TestSurgeAcceptsFreeCastsAndCastCardCopies(t *testing.T) {
	for _, copyCard := range []bool{false, true} {
		name := "physical spell"
		if copyCard {
			name = "copy of another player's card"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			owner := game.Player1
			if copyCard {
				owner = game.Player2
			}
			exiled := addCardToExile(g, owner, surgeProbeInstant())
			gated := addCardToHand(g, game.Player1, surgeProbeInstant())
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			cast := action.CastSpell(gated, nil, 0, nil)
			if containsAction(e.legalActions(g, game.Player1), cast) {
				t.Fatal("Surge was offered before the first cast")
			}

			var ok bool
			if copyCard {
				ok = castExiledCardCopy(e, g, game.Player1, exiled, nil)
			} else {
				ok = e.castFreeOption(g, g.Players[game.Player1], game.Player1, g.CardInstances[exiled],
					action.CastSpellAction{CardID: exiled, SourceZone: zone.Exile}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			}
			if !ok {
				t.Fatal("effect-granted free cast incorrectly required Surge eligibility")
			}
			obj, _ := g.Stack.Peek()
			if obj.CastCopy != copyCard || obj.AlternativeCost != 0 || obj.ManaSpentToCast != 0 || spellCastEvents(g, obj.ID) != 1 {
				t.Fatalf("wrong free cast branch or event: %+v", obj)
			}
			if !containsAction(e.legalActions(g, game.Player1), cast) || !e.applyAction(g, game.Player1, cast) {
				t.Fatal("genuine free cast did not unlock Surge for the caster")
			}
		})
	}
}

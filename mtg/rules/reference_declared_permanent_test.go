package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledDeclaredPermanentReadbacks(t *testing.T) {
	for _, permanentAction := range []string{
		"Destroy target creature.",
		"Tap target creature.",
		"Return target creature to its owner's hand.",
		"If you control an Island, destroy target creature.",
	} {
		for _, lifecycle := range []string{"ordinary", "real copy", "failed destroy", "invalid permanent", "invalid card"} {
			t.Run(permanentAction+lifecycle, func(t *testing.T) {
				prefix := "Exile target card from a graveyard. "
				extraLife := 1
				if permanentAction == "Return target creature to its owner's hand." {
					prefix += "You gain 1 life. "
					extraLife++
				}
				def := compileCounterTaxCard(t, "Declared Runtime Probe", prefix+permanentAction+
					" You gain 1 life. Its controller loses 2 life. Its owner gains 3 life.", "Sorcery")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				targetDef := vanillaCreature("Borrowed Target", 2, 3)
				if lifecycle == "failed destroy" {
					targetDef.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Indestructible)}}
				}
				target := addCombatPermanent(g, game.Player2, targetDef)
				target.Controller = game.Player3
				card := addCardToGraveyard(g, game.Player4, vanillaCreature("Independent Card", 1, 1))
				source := addCardToHand(g, game.Player4, def)
				g.Players[game.Player4].Hand.Remove(source)
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
					SourceID: source, SourceCardID: source, TargetCounts: []int{1, 1},
					Targets: []game.Target{
						{Kind: game.TargetCard, CardID: card, CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true},
						game.PermanentTarget(target.ObjectID),
					},
				}
				g.Stack.Push(obj)
				engine := NewEngine(nil)
				if lifecycle == "real copy" {
					addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
						[]game.Target{game.StackObjectTarget(obj.ID)})
					engine.resolveTopOfStack(g, &TurnLog{})
					var ok bool
					obj, ok = g.Stack.Peek()
					if !ok || !obj.Copy || len(obj.Targets) != 2 || obj.Targets[1].PermanentID != target.ObjectID {
						t.Fatal("real stack copy lost exact declared target occurrence")
					}
				}
				if lifecycle == "invalid permanent" && !movePermanentToZone(g, target, zone.Exile) {
					t.Fatal("target invalidation failed")
				}
				if lifecycle == "invalid card" && !moveCardBetweenZones(g, game.Player4, card, zone.Graveyard, zone.Hand) {
					t.Fatal("independent card invalidation failed")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				wantOwner, wantController := 43, 38
				if lifecycle == "invalid permanent" {
					wantOwner, wantController = 40, 40
				}
				if g.Players[game.Player1].Life != 40+extraLife ||
					g.Players[game.Player2].Life != wantOwner || g.Players[game.Player3].Life != wantController ||
					g.Players[game.Player4].Life != 40 {
					t.Fatalf("actor/owner/source/card isolation: life=%d/%d/%d/%d",
						g.Players[game.Player1].Life, g.Players[game.Player2].Life,
						g.Players[game.Player3].Life, g.Players[game.Player4].Life)
				}
				if lifecycle != "invalid card" && !g.Players[game.Player4].Exile.Contains(card) {
					t.Fatal("independent CARD action lost its exact target")
				}
				live := permanentForCard(g, target.CardInstanceID)
				if permanentAction == "If you control an Island, destroy target creature." && lifecycle != "invalid permanent" && live == nil {
					t.Fatal("skipped action moved declared target")
				}
				if permanentAction == "Destroy target creature." && lifecycle == "failed destroy" && live == nil {
					t.Fatal("indestructible target was destroyed")
				}
				if lifecycle != "invalid permanent" {
					if permanentAction == "Tap target creature." && (live == nil || !live.Tapped) {
						t.Fatal("Tap did not mutate the exact live target")
					}
					if permanentAction == "Return target creature to its owner's hand." &&
						(live != nil || !g.Players[game.Player2].Hand.Contains(target.CardInstanceID)) {
						t.Fatal("Return did not move the exact card to its physical owner's hand")
					}
					if permanentAction == "Destroy target creature." && lifecycle != "failed destroy" &&
						(live != nil || !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID)) {
						t.Fatal("Destroy did not move the exact target to its physical owner's graveyard")
					}
				}
			})
		}
	}
}

func TestCompiledDeclaredPermanentRawPaidCopyAndClone(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		for _, copied := range []bool{false, true} {
			name := noun + "/original"
			if copied {
				name = noun + "/real copy"
			}
			t.Run(name, func(t *testing.T) {
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Raw Declared Composition", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{G}",
					OracleText: "Exile target card from a graveyard. Return target " + noun +
						" to its owner's hand. Its controller loses 2 life. Its owner gains 3 life.",
				})
				if problems := game.ValidateCardDef(def); len(problems) != 0 {
					t.Fatalf("whole CardDef invalid: %v", problems)
				}
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				prior := addCardToGraveyard(g, game.Player4, vanillaCreature("Independent Card", 1, 1))
				targetDef := vanillaCreature("Borrowed Creature", 2, 2)
				if noun == "creature or Vehicle" {
					targetDef = &game.CardDef{CardFace: game.CardFace{
						Name: "Noncreature Vehicle", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Vehicle},
					}}
				}
				target := addCombatPermanent(g, game.Player2, targetDef)
				target.Controller = game.Player3
				if noun == "creature or Vehicle" && permanentHasType(g, target, types.Creature) {
					t.Fatal("Vehicle witness accidentally used the creature branch")
				}
				addBasicLandPermanent(g, game.Player1, types.Forest)
				spell := addCardToHand(g, game.Player1, def)
				setMainPhasePriority(g, game.Player1)
				engine := NewEngine(nil)
				targets := []game.Target{currentCardTarget(t, g, prior), game.PermanentTarget(target.ObjectID)}
				if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil)) {
					t.Fatal("natural paid raw composition failed to reach stack")
				}
				original, ok := g.Stack.Peek()
				if !ok || original.Copy || len(original.Targets) != 2 || original.Targets[1].PermanentID != target.ObjectID {
					t.Fatal("natural cast lost original permanent occurrence")
				}
				if copied {
					copiedSpell := game.NewStackObjectCopy(original, g.IDGen.Next())
					if !copiedSpell.Copy || copiedSpell.ID == original.ID || !slices.Equal(copiedSpell.Targets, original.Targets) {
						t.Fatal("real spell copy did not retain exact target incarnations")
					}
					g.Stack.Push(copiedSpell)
				}
				clone := g.Clone()
				for _, state := range []*game.Game{g, clone} {
					engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					if !state.Players[game.Player4].Exile.Contains(prior) ||
						!state.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
						permanentForCard(state, target.CardInstanceID) != nil {
						t.Fatal("raw composition lost exact card/permanent physical ownership")
					}
					if copied {
						pending, ok := state.Stack.Peek()
						if !ok || pending.Copy || pending.ID != original.ID {
							t.Fatal("copy consumed or replaced the original spell")
						}
						engine.resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					}
					if state.Stack.Size() != 0 || !state.Players[game.Player1].Graveyard.Contains(spell) {
						t.Fatal("original spell did not finish after its copied targets departed")
					}
					for player, want := range map[game.PlayerID]int{
						game.Player1: 40, game.Player2: 43, game.Player3: 38, game.Player4: 40,
					} {
						if got := state.Players[player].Life; got != want {
							t.Fatalf("player %d life=%d, want=%d after exact raw/copy/clone resolution", player, got, want)
						}
					}
					t.Logf("raw=%s copy=%v originalObject=%d card=%d owner=P2 hand=true originalController=P3 life38 caster=P1 life40 priorCardOwner=P4 exile=true life40",
						noun, copied, target.ObjectID, target.CardInstanceID)
				}
			})
		}
	}
}

func TestCompiledDeclaredPermanentDepartureAndReentry(t *testing.T) {
	for _, permanentAction := range []string{"Destroy target creature.", "Return target creature to its owner's hand."} {
		t.Run(permanentAction, func(t *testing.T) {
			prefix := "Exile target card from a graveyard. "
			first := 2
			destination := zone.Graveyard
			if permanentAction == "Return target creature to its owner's hand." {
				prefix += "You gain 1 life. "
				first++
				destination = zone.Hand
			}
			def := compileCounterTaxCard(t, "Declared Reentry Probe", prefix+permanentAction+
				" Its controller loses 2 life. Its owner gains 3 life.", "Sorcery")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Reentered Target", 2, 3))
			target.Controller = game.Player3
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType, AddTypes: []types.Card{types.Artifact},
			})
			if !permanentHasType(g, target, types.Artifact) {
				t.Fatal("live target did not acquire effective type")
			}
			card := addCardToGraveyard(g, game.Player4, vanillaCreature("Prefix Card", 1, 1))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{
					{Kind: game.TargetCard, CardID: card, CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true},
					game.PermanentTarget(target.ObjectID),
				},
			}
			engine := NewEngine(nil)
			sequence := def.SpellAbility.Val.Modes[0].Sequence
			for _, instruction := range sequence[:first] {
				resolveInstruction(engine, g, obj, instruction.Primitive, &TurnLog{})
			}
			resolved, ok := resolveObjectReference(g, obj, game.TargetPermanentReference(1))
			if !ok || resolved.permanent != nil || resolved.snapshot.Controller != game.Player3 ||
				!slices.Contains(resolved.snapshot.Types, types.Artifact) {
				t.Fatal("departing original target lost effective LKI")
			}
			r := effectResolver{game: g, engine: engine, obj: obj}
			entered, ok := r.putResolvedCardOnBattlefieldValue(g.CardInstances[target.CardInstanceID], destination,
				game.Player4, nil, permanentCreationOptions{})
			if !ok || entered.ObjectID == target.ObjectID || permanentHasType(g, entered, types.Artifact) {
				t.Fatal("reentry did not create a new object with current, independent types")
			}
			for _, instruction := range sequence[first:] {
				resolveInstruction(engine, g, obj, instruction.Primitive, &TurnLog{})
			}
			if g.Players[game.Player2].Life != 43 || g.Players[game.Player3].Life != 38 || g.Players[game.Player4].Life != 40 ||
				permanentForCard(g, entered.CardInstanceID).ObjectID != entered.ObjectID {
				t.Fatal("readbacks followed reached card/current controller instead of original ObjectID")
			}
		})
	}
}

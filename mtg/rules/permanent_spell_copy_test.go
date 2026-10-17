package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func permanentCopyEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: game.ProgramReference{ID: "test/permanent-spell-copy", Version: "1"},
		Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Copy(game.CopyStackObject{Object: game.TargetStackObjectReference(0)}, ctx.View().Source().Controller)
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func copyPermanentSpell(t *testing.T, e *Engine, g *game.Game, original *game.StackObject, controller game.PlayerID) *game.StackObject {
	t.Helper()
	addImplementationSpellToStack(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name: "Copy a permanent spell", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Program: game.ProgramReference{ID: "test/permanent-spell-copy", Version: "1"},
			Targets: []game.TargetSpec{{
				Allow: game.TargetAllowStackObject, Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
				MinTargets: 1, MaxTargets: 1,
			}},
		}.Ability()),
	}}, []game.Target{game.StackObjectTarget(original.ID)})
	e.resolveTopOfStack(g, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.ID == original.ID {
		t.Fatal("authored copy did not produce an independent spell")
	}
	return copied
}

func TestPermanentSpellCopyIsNotTokenCreation(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, replacement := range []string{"Doubling Season", "identity", "cross-type addend"} {
		t.Run(replacement, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := permanentCopyEngine(t)
			if replacement == "Doubling Season" {
				addReplacementPermanent(t, g, game.Player1, registry.Lookup("Doubling Season"))
			} else {
				effect := game.ReplacementEffect{ID: g.IDGen.Next(), MatchEvent: game.EventTokenCreated}
				if replacement == "identity" {
					effect.TokenReplaceDef = treasureTokenCardDef()
				} else {
					effect.TokenAddend, effect.TokenAddendDef = 1, treasureTokenCardDef()
				}
				g.ReplacementEffects = append(g.ReplacementEffects, effect)
			}
			def := vanillaCreature("Copied creature", 2, 2)
			def.ReplacementAbilities = []game.ReplacementAbility{
				game.EntersTappedReplacement("Enter tapped."),
				game.EntersWithCountersReplacement("Enter with three counters.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}),
			}
			addImplementationSpellToStack(g, game.Player1, def, nil)
			original, _ := g.Stack.Peek()
			copied := copyPermanentSpell(t, e, g, original, game.Player1)
			start, before := len(g.Events), len(g.Battlefield)
			branch := g.Clone()
			e.resolveTopOfStack(g, &TurnLog{})
			e.resolveTopOfStack(branch, &TurnLog{})
			for _, result := range []*game.Game{g, branch} {
				if len(result.Battlefield) != before+1 || result.Stack.Size() != 1 {
					t.Fatal("creation replacements changed the copy count or its original")
				}
				token := result.Battlefield[len(result.Battlefield)-1]
				want := 3
				if replacement == "Doubling Season" {
					want = 6
				}
				if !token.Token || permanentEffectiveName(result, token) != def.Name || !token.Tapped ||
					token.Counters.Get(counter.PlusOnePlusOne) != want || token.CardInstanceID != 0 {
					t.Fatalf("copy lost entry characteristics or gained token-creation substitutions: %+v", token)
				}
				arrivals := 0
				for _, event := range result.Events[start:] {
					if event.Kind == game.EventTokenCreated || event.Kind == game.EventSpellCast {
						t.Fatal("resolving a spell copy emitted creation/cast")
					}
					if event.Kind == game.EventPermanentEnteredBattlefield {
						arrivals++
						if event.FromZone != zone.Stack || event.StackObjectID != copied.ID {
							t.Fatal("copy entry lost its stack origin")
						}
					}
				}
				if arrivals != 1 {
					t.Fatalf("copy arrival events = %d, want 1", arrivals)
				}
			}
			start = len(g.Events)
			created, ok := createTokenPermanentsCollectingWithChoices(e, g, game.Player1, def, 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !ok || len(created) == 0 {
				t.Fatal("ordinary token creation failed")
			}
			creationEvents := 0
			for _, event := range g.Events[start:] {
				if event.Kind == game.EventTokenCreated {
					creationEvents++
				}
			}
			want := 2
			if replacement == "identity" {
				want = 1
				if permanentEffectiveName(g, created[0]) != "Treasure" {
					t.Fatal("ordinary token creation missed identity replacement")
				}
			}
			if len(created) != want || creationEvents != want {
				t.Fatal("spell-copy correction changed ordinary token creation")
			}
		})
	}
}

func TestPermanentSpellCopyOnlyFiresEntryTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := permanentCopyEngine(t)
	for _, event := range []game.EventKind{game.EventTokenCreated, game.EventPermanentEnteredBattlefield, game.EventSpellCast} {
		addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{Event: event},
			[]game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}, nil)
	}
	addImplementationSpellToStack(g, game.Player1, vanillaCreature("Entry subject", 2, 2), nil)
	original, _ := g.Stack.Peek()
	copyPermanentSpell(t, e, g, original, game.Player1)
	g.TriggerEventCursor = len(g.Events)
	e.resolveTopOfStack(g, &TurnLog{})
	if !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatalf("entry must fire once, creation/cast never: stack %d", g.Stack.Size())
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 {
		t.Fatal("copy entry trigger did not execute")
	}
}

func TestPermanentSpellCopyChoicesAuraAndFaces(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	t.Run("copied entry choice", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		addReplacementPermanent(t, g, game.Player1, registry.Lookup("Triskelion"))
		addImplementationSpellToStack(g, game.Player1, registry.Lookup("Clone"), nil)
		original, _ := g.Stack.Peek()
		copyPermanentSpell(t, e, g, original, game.Player2)
		e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
			game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
		}, &TurnLog{})
		token := g.Battlefield[len(g.Battlefield)-1]
		if !token.Token || token.Owner != game.Player2 || token.Controller != game.Player2 ||
			token.Counters.Get(counter.PlusOnePlusOne) != 3 || permanentEffectiveName(g, token) != "Triskelion" {
			t.Fatal("spell-copy entry skipped its copy choice or has the original's owner")
		}
	})
	t.Run("Aura", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		host := addCombatCreaturePermanent(g, game.Player1)
		addImplementationSpellToStack(g, game.Player1, registry.Lookup("Rancor"), []game.Target{game.PermanentTarget(host.ObjectID)})
		original, _ := g.Stack.Peek()
		copyPermanentSpell(t, e, g, original, game.Player2)
		e.resolveTopOfStack(g, &TurnLog{})
		aura := g.Battlefield[len(g.Battlefield)-1]
		if !aura.Token || !aura.AttachedTo.Exists || aura.AttachedTo.Val != host.ObjectID {
			t.Fatal("copied Aura did not attach normally")
		}
	})
	for _, face := range []game.FaceIndex{game.FaceFront, game.FaceBack} {
		t.Run(map[game.FaceIndex]string{game.FaceFront: "front", game.FaceBack: "back"}[face], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := permanentCopyEngine(t)
			def := vanillaCreature("Front creature", 2, 2)
			back := vanillaCreature("Back creature", 4, 4)
			def.Layout, def.Back = game.LayoutTransform, opt.Val(back.CardFace)
			addImplementationSpellToStack(g, game.Player1, def, nil)
			original, _ := g.Stack.Peek()
			original.Face = face
			copyPermanentSpell(t, e, g, original, game.Player2)
			e.resolveTopOfStack(g, &TurnLog{})
			token := g.Battlefield[len(g.Battlefield)-1]
			if !token.Token || token.Face != face || !token.TokenDef.Back.Exists ||
				!transformPermanent(g, token) || token.Face == face {
				t.Fatal("spell copy lost a face or could not transform independently")
			}
			if original.Face != face {
				t.Fatal("transforming token changed the original spell")
			}
		})
	}
	t.Run("face down", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		addImplementationSpellToStack(g, game.Player1, registry.Lookup("Triskelion"), nil)
		original, _ := g.Stack.Peek()
		original.FaceDown = true
		copyPermanentSpell(t, e, g, original, game.Player2)
		e.resolveTopOfStack(g, &TurnLog{})
		token := g.Battlefield[len(g.Battlefield)-1]
		if !token.Token || token.FaceDown || effectivePower(g, token) != 2 || len(permanentEffectiveAbilities(g, token)) != 0 ||
			token.Counters.Get(counter.PlusOnePlusOne) != 0 || !original.FaceDown {
			t.Fatal("copy leaked the hidden spell's printed characteristics")
		}
	})
	t.Run("modal back face", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		def := vanillaCreature("Modal front", 2, 2)
		def.ManaCost = opt.Val(cost.Mana{cost.O(2)})
		back := vanillaCreature("Modal back", 7, 7)
		back.ManaCost = opt.Val(cost.Mana{cost.O(7)})
		def.Layout, def.Back = game.LayoutModalDFC, opt.Val(back.CardFace)
		addImplementationSpellToStack(g, game.Player1, def, nil)
		original, _ := g.Stack.Peek()
		original.Face = game.FaceBack
		copyPermanentSpell(t, e, g, original, game.Player2)
		e.resolveTopOfStack(g, &TurnLog{})
		token := g.Battlefield[len(g.Battlefield)-1]
		mv, known := effectivePermanentManaValue(g, token)
		if !known || mv != 7 || token.TokenDef.Layout != game.LayoutModalDFC || !token.TokenDef.Back.Exists {
			t.Fatal("modal copy was treated as a nonmodal transforming token")
		}
	})
	t.Run("illegal mutate target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		addReplacementPermanent(t, g, game.Player1, registry.Lookup("Doubling Season"))
		target := addCombatCreaturePermanent(g, game.Player1)
		addImplementationSpellToStack(g, game.Player1, registry.Lookup("Gemrazer"), nil)
		original, _ := g.Stack.Peek()
		original.Mutate, original.MutateTargetID = true, target.ObjectID
		copyPermanentSpell(t, e, g, original, game.Player1)
		movePermanentToZone(g, target, zone.Graveyard)
		start, before := len(g.Events), len(g.Battlefield)
		e.resolveTopOfStack(g, &TurnLog{})
		if len(g.Battlefield) != before+1 {
			t.Fatal("failed mutate copy used token-creation replacements")
		}
		for _, event := range g.Events[start:] {
			if event.Kind == game.EventTokenCreated {
				t.Fatal("failed mutate copy emitted token creation")
			}
		}
	})
}

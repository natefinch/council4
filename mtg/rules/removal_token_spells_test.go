package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Oracle corpus SHA-256: 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// WotC rulings retrieved 2026-10-14: Gift (2019-06-14), Pongify (2021-03-19),
// Hybridization (2013-01-24, 2013-07-01); Resculpt has no published rulings.
// Official CR 2026-09-25: 111.2–4, 122.1c, 608.2b–c/h, 701.8/13/19,
// 702.12, 614.6/12, 616.1 and 903.9a.
type removalTokenCase struct {
	entry         cardset.Entry
	tokenName     string
	size          int
	colors        []color.Color
	subtypes      []types.Sub
	manaColor     mana.Color
	generic       int
	destination   zone.Type
	canRegenerate bool
	allowed       []types.Card
}

//nolint:gosec // Oracle token names are characteristics, not credentials.
func removalTokenCases() []removalTokenCase {
	return []removalTokenCase{
		{
			entry: authored.GenerousGift, tokenName: "Elephant Token", size: 3,
			colors: []color.Color{color.Green}, subtypes: []types.Sub{types.Elephant},
			manaColor: mana.W, generic: 2, destination: zone.Graveyard, canRegenerate: true,
			allowed: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
		},
		{
			entry: authored.Pongify, tokenName: "Ape Token", size: 3,
			colors: []color.Color{color.Green}, subtypes: []types.Sub{types.Ape},
			manaColor: mana.U, destination: zone.Graveyard, allowed: []types.Card{types.Creature},
		},
		{
			entry: authored.RapidHybridization, tokenName: "Frog Lizard Token", size: 3,
			colors: []color.Color{color.Green}, subtypes: []types.Sub{types.Frog, types.Lizard},
			manaColor: mana.U, destination: zone.Graveyard, allowed: []types.Card{types.Creature},
		},
		{
			entry: authored.Resculpt, tokenName: "Elemental Token", size: 4,
			colors: []color.Color{color.Blue, color.Red}, subtypes: []types.Sub{types.Elemental},
			manaColor: mana.U, generic: 1, destination: zone.Exile, allowed: []types.Card{types.Artifact, types.Creature},
		},
	}
}

func removalTokenEngine(t *testing.T, entry cardset.Entry) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(entry.Programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func removalTokenStack(t *testing.T, tc removalTokenCase, def *game.CardDef) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, def)
	addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.PermanentTarget(target.ObjectID)})
	return g, removalTokenEngine(t, tc.entry), target
}

func assertRemovalTokens(t *testing.T, g *game.Game, tc removalTokenCase, controller game.PlayerID, count int, tapped bool) {
	t.Helper()
	got := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.TokenDef.Name != tc.tokenName {
			continue
		}
		got++
		def := permanent.TokenDef
		if permanent.Controller != controller || permanent.Owner != controller || permanent.Tapped != tapped ||
			!slices.Equal(def.Types, []types.Card{types.Creature}) || !slices.Equal(def.Subtypes, tc.subtypes) ||
			!slices.Equal(def.Colors, tc.colors) || !def.Power.Exists || def.Power.Val.Value != tc.size ||
			!def.Toughness.Exists || def.Toughness.Val.Value != tc.size || def.ManaCost.Exists ||
			len(def.StaticAbilities) != 0 || len(def.TriggeredAbilities) != 0 || len(def.ActivatedAbilities) != 0 {
			t.Fatalf("wrong token recipient, entry state or characteristics: %+v / %+v", permanent, def)
		}
	}
	if got != count {
		t.Fatalf("%s count = %d, want %d", tc.tokenName, got, count)
	}
}

func assertRemovalDestination(t *testing.T, g *game.Game, target *game.Permanent, destination zone.Type) {
	t.Helper()
	if permanentByCardID(g, target.CardInstanceID) != nil {
		t.Fatal("target remained on the battlefield")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == target.CardInstanceID && event.FromZone == zone.Battlefield && event.ToZone == destination
	})
}

// The generated Generous Gift scenario is completed with its real target/body.
func TestAuthoredGenerousGiftScenario(t *testing.T) {
	t.Parallel()
	tc := removalTokenCases()[0]
	g, engine, target := removalTokenStack(t, tc, vanillaCreature("Target", 2, 2))
	engine.resolveTopOfStack(g, &TurnLog{})
	assertRemovalDestination(t, g, target, zone.Graveyard)
	assertRemovalTokens(t, g, tc, game.Player2, 1, false)
}

func TestAuthoredRemovalTokensCastAndTargetTypes(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, targetType := range []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Land, types.Planeswalker} {
				t.Run(string(targetType), func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					def := vanillaCreature("Target", 2, 2)
					def.Types = []types.Card{targetType}
					target := addCombatPermanent(g, game.Player2, def)
					spell := addCardToHand(g, game.Player1, tc.entry.New())
					g.Players[game.Player1].ManaPool.Add(tc.manaColor, 1)
					g.Players[game.Player1].ManaPool.Add(mana.C, tc.generic)
					setMainPhasePriority(g, game.Player1)
					cast := engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil))
					wantCast := slices.Contains(tc.allowed, targetType)
					if cast != wantCast {
						t.Fatalf("cast on %s = %v, want %v", targetType, cast, wantCast)
					}
					if !cast {
						if !g.Stack.IsEmpty() || !g.Players[game.Player1].Hand.Contains(spell) {
							t.Fatal("illegal cast consumed the spell")
						}
						return
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					assertRemovalDestination(t, g, target, tc.destination)
					assertRemovalTokens(t, g, tc, game.Player2, 1, false)
					if !g.Players[game.Player1].Graveyard.Contains(spell) {
						t.Fatal("instant did not finish in its owner's graveyard")
					}
				})
			}
		})
	}
}

func TestAuthoredRemovalTokensProtectedTargets(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, protection := range []struct {
				name    string
				static  []game.StaticAbility
				shield  int
				regen   int
				survive bool
			}{
				{name: "indestructible", static: []game.StaticAbility{game.IndestructibleStaticBody}, survive: tc.destination == zone.Graveyard},
				{name: "shield", shield: 1, survive: tc.destination == zone.Graveyard},
				{name: "regeneration", regen: 1, survive: tc.canRegenerate},
			} {
				t.Run(protection.name, func(t *testing.T) {
					t.Parallel()
					def := vanillaCreature("Target", 2, 2)
					def.StaticAbilities = protection.static
					g, engine, target := removalTokenStack(t, tc, def)
					target.Counters.Add(counter.Shield, protection.shield)
					target.RegenerationShields = protection.regen
					target.MarkedDamage = 1
					engine.resolveTopOfStack(g, &TurnLog{})
					if protection.survive {
						if permanentByCardID(g, target.CardInstanceID) != target {
							t.Fatal("protected target did not survive")
						}
						if protection.shield > 0 && target.Counters.Get(counter.Shield) != 0 {
							t.Fatal("shield was not consumed")
						}
						if protection.regen > 0 && (target.RegenerationShields != 0 || !target.Tapped || target.MarkedDamage != 0) {
							t.Fatal("regeneration did not consume the shield, tap and clear damage")
						}
					} else {
						assertRemovalDestination(t, g, target, tc.destination)
						if protection.regen > 0 && target.RegenerationShields != 1 {
							t.Fatal("prohibited regeneration was applied")
						}
					}
					assertRemovalTokens(t, g, tc, game.Player2, 1, false)
				})
			}
		})
	}
}

func TestAuthoredRemovalTokensAllIllegalTargets(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, change := range []struct {
				name  string
				apply func(*testing.T, *game.Game, *game.Permanent)
			}{
				{"left and returned", func(t *testing.T, g *game.Game, target *game.Permanent) {
					if !movePermanentToZone(g, target, zone.Exile) {
						t.Fatal("fixture could not exile target")
					}
					card, _ := g.GetCardInstance(target.CardInstanceID)
					returned, ok := createCardPermanent(g, card, game.Player2, zone.Exile)
					if !ok || returned.ObjectID == target.ObjectID {
						t.Fatal("fixture did not create a new incarnation")
					}
				}},
				{"hexproof", func(t *testing.T, g *game.Game, target *game.Permanent) {
					card, _ := g.GetCardInstance(target.CardInstanceID)
					def := *card.Def
					def.StaticAbilities = []game.StaticAbility{game.HexproofStaticBody}
					card.Def = &def
				}},
			} {
				t.Run(change.name, func(t *testing.T) {
					t.Parallel()
					g, engine, target := removalTokenStack(t, tc, vanillaCreature("Target", 2, 2))
					obj, _ := g.Stack.Peek()
					spell := obj.SourceID
					change.apply(t, g, target)
					engine.resolveTopOfStack(g, &TurnLog{})
					assertRemovalTokens(t, g, tc, game.Player2, 0, false)
					if !g.Stack.IsEmpty() || !g.Players[game.Player1].Graveyard.Contains(spell) ||
						permanentByCardID(g, target.CardInstanceID) == nil {
						t.Fatal("all-illegal spell resolved or did not leave the stack")
					}
				})
			}
		})
	}
}

func TestAuthoredRemovalTokensChangedTypesAndCopies(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			t.Run("changed type", func(t *testing.T) {
				t.Parallel()
				g, engine, target := removalTokenStack(t, tc, vanillaCreature("Target", 2, 2))
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType,
					SetTypes: []types.Card{types.Enchantment}, Duration: game.DurationPermanent,
				})
				engine.resolveTopOfStack(g, &TurnLog{})
				if tc.entry.OracleID == authored.GenerousGift.OracleID {
					assertRemovalDestination(t, g, target, tc.destination)
					assertRemovalTokens(t, g, tc, game.Player2, 1, false)
				} else {
					if permanentByCardID(g, target.CardInstanceID) == nil {
						t.Fatal("target that lost the allowed types was affected")
					}
					assertRemovalTokens(t, g, tc, game.Player2, 0, false)
				}
			})
			t.Run("retargeted copy", func(t *testing.T) {
				t.Parallel()
				g, engine, originalTarget := removalTokenStack(t, tc, vanillaCreature("Original target", 2, 2))
				newTarget := addCombatPermanent(g, game.Player3, vanillaCreature("New target", 2, 2))
				original, _ := g.Stack.Peek()
				copySource := &game.StackObject{
					ID: g.IDGen.Next(), Controller: game.Player4,
					Targets: []game.Target{game.StackObjectTarget(original.ID)},
				}
				log := TurnLog{}
				policies := [game.NumPlayers]PlayerAgent{
					game.Player4: authoredChoicePolicy{choose: func(obs PlayerObservation, request game.ChoiceRequest) []int {
						if request.Kind != game.ChoiceTarget || request.Player != game.Player4 {
							t.Fatal("copy's retargeting did not use its controller's choice")
						}
						return chooseAuthoredTargets(t, request, game.PermanentTarget(newTarget.ObjectID))
					}},
				}
				resolver := newEffectResolver(engine, g, copySource, policies, &log)
				resolver.resolveInstruction(&game.Instruction{Primitive: game.CopyStackObject{
					Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
				}})
				if g.Stack.Size() != 2 || len(log.Choices) != 1 || log.Choices[0].UsedFallback {
					t.Fatal("real spell-copy retargeting was not executed and recorded")
				}
				branch := g.Clone()
				engine.resolveTopOfStack(branch, &TurnLog{})
				if permanentByCardID(g, newTarget.CardInstanceID) == nil {
					t.Fatal("resolving the clone mutated the original")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				assertRemovalDestination(t, g, newTarget, tc.destination)
				assertRemovalTokens(t, g, tc, game.Player3, 1, false)
				if permanentByCardID(g, originalTarget.CardInstanceID) == nil {
					t.Fatal("retargeting the copy changed the original spell's target")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				assertRemovalDestination(t, g, originalTarget, tc.destination)
				if namedTokenCount(g, game.Player2, tc.tokenName) != 1 ||
					namedTokenCount(g, game.Player3, tc.tokenName) != 1 ||
					namedTokenCount(branch, game.Player3, tc.tokenName) != 1 {
					t.Fatal("copy/original token recipients or clone resolution differed")
				}
			})
		})
	}
}

func TestAuthoredRemovalTokensUmbraArmorAndSourceDeparture(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g, engine, target := removalTokenStack(t, tc, vanillaCreature("Target", 2, 2))
			aura := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Protecting Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
				StaticAbilities: []game.StaticAbility{{
					KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}}},
				}, game.UmbraArmorStaticBody},
			}})
			if !attachPermanent(g, aura, target) {
				t.Fatal("fixture could not attach umbra armor")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if tc.destination == zone.Graveyard {
				if permanentByCardID(g, target.CardInstanceID) == nil ||
					!g.Players[game.Player2].Graveyard.Contains(aura.CardInstanceID) {
					t.Fatal("non-regeneration umbra armor replacement was bypassed")
				}
			} else {
				assertRemovalDestination(t, g, target, zone.Exile)
				if permanentByCardID(g, aura.CardInstanceID) == nil {
					t.Fatal("umbra armor was wrongly applied to exile")
				}
			}
			assertRemovalTokens(t, g, tc, game.Player2, 1, false)
		})
	}
	t.Run("Generous Gift destroys token-doubling source", func(t *testing.T) {
		t.Parallel()
		tc := removalTokenCases()[0]
		g, engine, target := removalTokenStack(t, tc, tokenDoublingReplacementCardDef())
		engine.resolveTopOfStack(g, &TurnLog{})
		assertRemovalDestination(t, g, target, zone.Graveyard)
		assertRemovalTokens(t, g, tc, game.Player2, 1, false)
	})
}

func TestAuthoredRemovalTokensControllerSnapshot(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g, engine, target := removalTokenStack(t, tc, vanillaCreature("Stolen target", 2, 2))
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerControl,
				NewController: opt.Val(game.Player3), Duration: game.DurationPermanent,
			})
			engine.resolveTopOfStack(g, &TurnLog{})
			assertRemovalDestination(t, g, target, tc.destination)
			assertRemovalTokens(t, g, tc, game.Player3, 1, false)
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) &&
				!g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
				t.Fatal("target went to its controller's zone instead of its owner's")
			}

		})
	}
}

func TestAuthoredRemovalTokensSurvivingTargetUsesCurrentController(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases()[:3] {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g, engine, target := removalTokenStack(t, tc, vanillaCreature("Stolen target", 2, 2))
			aura := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
				Name: "Control-changing umbra Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
				StaticAbilities: []game.StaticAbility{{
					KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}}},
				}, game.UmbraArmorStaticBody, {
					ContinuousEffects: []game.ContinuousEffect{{
						Layer: game.LayerControl, NewController: opt.Val(game.Player3),
						Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
					}},
				}},
			}})
			if !attachPermanent(g, aura, target) {
				t.Fatal("fixture could not attach Aura")
			}
			if effectiveController(g, target) != game.Player3 {
				t.Fatal("fixture did not steal target")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if permanentByCardID(g, target.CardInstanceID) == nil ||
				!g.Players[game.Player3].Graveyard.Contains(aura.CardInstanceID) ||
				effectiveController(g, target) != game.Player2 {
				t.Fatal("umbra did not end the control effect while preserving the target")
			}
			assertRemovalTokens(t, g, tc, game.Player2, 1, false)
		})
	}
}

func TestAuthoredRemovalTokensZoneAndTokenReplacements(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g, engine, target := removalTokenStack(t, tc, vanillaCreature("Target", 2, 2))
			destination := zone.Exile
			if tc.destination == zone.Exile {
				destination = zone.Hand
			}
			g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
				ID: g.IDGen.Next(), Description: "redirect removal", MatchEvent: game.EventZoneChanged,
				MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: tc.destination,
				ReplaceToZone: destination,
			})
			addReplacementPermanent(t, g, game.Player2, tokenDoublingReplacementCardDef())
			addReplacementPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Entry replacement", Types: []types.Card{types.Enchantment},
				ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedGroupReplacement(
					"Opponents' creatures enter tapped.", game.TriggerControllerOpponent, types.Creature,
				)},
			}})
			engine.resolveTopOfStack(g, &TurnLog{})
			assertRemovalDestination(t, g, target, destination)
			assertRemovalTokens(t, g, tc, game.Player2, 2, true)
		})
	}
}

func TestAuthoredRemovalTokensTargetedTokensAndCommanderSBA(t *testing.T) {
	t.Parallel()
	for _, tc := range removalTokenCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			t.Run("token", func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := removalTokenEngine(t, tc.entry)
				target, ok := createTokenPermanentWithChoicesInBatch(engine, g, game.Player2,
					vanillaCreature("Target Token", 2, 2), 0, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if !ok {
					t.Fatal("fixture could not create target token")
				}
				addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.PermanentTarget(target.ObjectID)})
				engine.resolveTopOfStack(g, &TurnLog{})
				assertRemovalTokens(t, g, tc, game.Player2, 1, false)
				engine.applyStateBasedActions(g)
				if namedTokenCount(g, game.Player2, "Target Token") != 0 ||
					g.Players[game.Player2].Graveyard.Contains(target.ObjectID) || g.Players[game.Player2].Exile.Contains(target.ObjectID) {
					t.Fatal("departed token did not cease to exist")
				}
			})
			t.Run("commander", func(t *testing.T) {
				t.Parallel()
				g, engine, target := removalTokenStack(t, tc, vanillaCreature("Commander", 2, 2))
				g.CommanderIDs[target.CardInstanceID] = true
				engine.resolveTopOfStack(g, &TurnLog{})
				assertRemovalDestination(t, g, target, tc.destination)
				assertRemovalTokens(t, g, tc, game.Player2, 1, false)
				engine.applyStateBasedActions(g)
				if !g.Players[game.Player2].CommandZone.Contains(target.CardInstanceID) {
					t.Fatal("commander did not move to command at SBA")
				}
				assertRemovalTokens(t, g, tc, game.Player2, 1, false)
			})
		})
	}
}

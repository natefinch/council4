package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Pinned identities/revisions and current rulings: testdata/card-authoring/production.json.
// CR effective 2026-09-25: 400.7, 601.2c, 608.2b, 614/616, 701.8/19, 702.12.
type productionDestructionCase struct {
	entry      cardset.Entry
	mana       []mana.Color
	types      []types.Card
	sorcery    bool
	noRegrowth bool
}

func productionDestructionCases() []productionDestructionCase {
	return []productionDestructionCase{
		{entry: authored.Bedevil, mana: []mana.Color{mana.B, mana.B, mana.R}, types: []types.Card{types.Artifact, types.Creature, types.Planeswalker}},
		{entry: authored.Terminate, mana: []mana.Color{mana.B, mana.R}, types: []types.Card{types.Creature}, noRegrowth: true},
		{entry: authored.Vindicate, mana: []mana.Color{mana.C, mana.W, mana.B}, types: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, sorcery: true},
		{entry: authored.Disenchant, mana: []mana.Color{mana.C, mana.W}, types: []types.Card{types.Artifact, types.Enchantment}},
		{entry: authored.Naturalize, mana: []mana.Color{mana.C, mana.G}, types: []types.Card{types.Artifact, types.Enchantment}},
	}
}

func productionArtifactCreature() *game.CardDef {
	def := vanillaCreature("Artifact creature", 2, 2)
	def.Types = []types.Card{types.Artifact, types.Creature}
	return def
}

func TestProductionDestructionCastAndTypes(t *testing.T) {
	t.Parallel()
	for _, tc := range productionDestructionCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, targetType := range []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle} {
				t.Run(string(targetType), func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					def := vanillaCreature("Subject", 2, 2)
					def.Types = []types.Card{targetType}
					target := addCombatPermanent(g, game.Player2, def)
					spell := addCardToHand(g, game.Player1, tc.entry.New())
					for _, c := range tc.mana {
						g.Players[game.Player1].ManaPool.Add(c, 1)
					}
					setMainPhasePriority(g, game.Player1)
					act := action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
					if cast := engine.applyAction(g, game.Player1, act); cast != slices.Contains(tc.types, targetType) {
						t.Fatalf("cast on %s = %v", targetType, cast)
					}
					if g.Stack.IsEmpty() {
						if !g.Players[game.Player1].Hand.Contains(spell) {
							t.Fatal("illegal target consumed spell")
						}
						return
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
						!g.Players[game.Player1].Graveyard.Contains(spell) {
						t.Fatal("real cast did not destroy target/finish spell")
					}
				})
			}
			t.Run("timing and insufficient payment", func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := removalTokenEngine(t, tc.entry)
				target := addCombatPermanent(g, game.Player2, productionArtifactCreature())
				spell := addCardToHand(g, game.Player1, tc.entry.New())
				setMainPhasePriority(g, game.Player1)
				act := action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
				if engine.applyAction(g, game.Player1, act) {
					t.Fatal("unpaid spell cast")
				}
				for _, c := range tc.mana {
					g.Players[game.Player1].ManaPool.Add(c, 1)
				}
				g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
				if cast := engine.applyAction(g, game.Player1, act); cast == tc.sorcery {
					t.Fatal("instant/sorcery timing incorrect")
				}
			})
		})
	}
}

func TestProductionDestructionProtectionAndReplacements(t *testing.T) {
	t.Parallel()
	for _, tc := range productionDestructionCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, protection := range []string{"indestructible", "shield", "regeneration", "exile", "umbra"} {
				t.Run(protection, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					def := productionArtifactCreature()
					if protection == "indestructible" {
						def.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
					}
					target := addCombatPermanent(g, game.Player2, def)
					if protection == "shield" {
						target.Counters.Add(counter.Shield, 1)
					}
					if protection == "regeneration" {
						target.RegenerationShields = 1
						target.MarkedDamage = 1
					}
					if protection == "exile" {
						g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
							ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
							MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Graveyard, AffectedObjectID: target.ObjectID,
							ReplaceToZone: zone.Exile, Duration: game.DurationPermanent,
						})
					}
					var aura *game.Permanent
					if protection == "umbra" {
						aura = addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
							Name: "Umbra", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
							StaticAbilities: []game.StaticAbility{
								{KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
									Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
								}}}},
								game.UmbraArmorStaticBody,
							},
						}})
						if !attachPermanent(g, aura, target) {
							t.Fatal("failed to attach umbra")
						}
					}
					addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.PermanentTarget(target.ObjectID)})
					engine.resolveTopOfStack(g, &TurnLog{})
					destroyed := protection == "exile" || protection == "regeneration" && tc.noRegrowth
					if (permanentByCardID(g, target.CardInstanceID) == nil) != destroyed {
						t.Fatal("protection/destination replacement failed")
					}
					if protection == "shield" && target.Counters.Get(counter.Shield) != 0 {
						t.Fatal("shield was not consumed")
					}
					if protection == "regeneration" && !tc.noRegrowth &&
						(target.RegenerationShields != 0 || !target.Tapped || target.MarkedDamage != 0) {
						t.Fatal("regeneration did not replace destruction")
					}
					if protection == "exile" && !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
						t.Fatal("zone replacement was bypassed")
					}
					if protection == "umbra" && !g.Players[game.Player2].Graveyard.Contains(aura.CardInstanceID) {
						t.Fatal("cannot regenerate wrongly bypassed umbra armor")
					}
				})
			}
		})
	}
}

func TestProductionDestructionIdentityAndCopies(t *testing.T) {
	t.Parallel()
	for _, tc := range productionDestructionCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, change := range []string{"blink", "hexproof", "illegal type", "still legal type"} {
				t.Run(change, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					target := addCombatPermanent(g, game.Player2, productionArtifactCreature())
					spell := addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.PermanentTarget(target.ObjectID)})
					switch change {
					case "blink":
						movePermanentToZone(g, target, zone.Exile)
						card, _ := g.GetCardInstance(target.CardInstanceID)
						if _, ok := createCardPermanent(g, card, game.Player2, zone.Exile); !ok {
							t.Fatal("fixture failed to return target")
						}
					case "hexproof":
						g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
							ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof}, Duration: game.DurationPermanent,
						})
					default:
						newType := types.Land
						if change == "still legal type" {
							newType = tc.types[len(tc.types)-1]
						}
						g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
							ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType,
							SetTypes: []types.Card{newType}, Duration: game.DurationPermanent,
						})
					}
					engine.resolveTopOfStack(g, &TurnLog{})
					shouldDestroy := change == "still legal type" || change == "illegal type" && tc.entry.OracleID == authored.Vindicate.OracleID
					if (permanentByCardID(g, target.CardInstanceID) == nil) != shouldDestroy ||
						!g.Players[game.Player1].Graveyard.Contains(spell) {
						t.Fatal("target revalidation followed physical card or lost a still-legal type union")
					}
				})
			}
			t.Run("retargeted copy and clone", func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := removalTokenEngine(t, tc.entry)
				first := addCombatPermanent(g, game.Player2, productionArtifactCreature())
				second := addCombatPermanent(g, game.Player3, productionArtifactCreature())
				addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.PermanentTarget(first.ObjectID)})
				original, _ := g.Stack.Peek()
				source := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player4, Targets: []game.Target{game.StackObjectTarget(original.ID)}}
				agents := [game.NumPlayers]PlayerAgent{game.Player4: authoredChoicePolicy{choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
					return chooseAuthoredTargets(t, request, game.PermanentTarget(second.ObjectID))
				}}}
				newEffectResolver(engine, g, source, agents, &TurnLog{}).resolveInstruction(&game.Instruction{Primitive: game.CopyStackObject{
					Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
				}})
				clone := g.Clone()
				engine.resolveTopOfStack(clone, &TurnLog{})
				if permanentByCardID(g, second.CardInstanceID) == nil {
					t.Fatal("clone changed original")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player3].Graveyard.Contains(second.CardInstanceID) ||
					permanentByCardID(g, first.CardInstanceID) == nil {
					t.Fatal("copy targets or owner destination changed")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player2].Graveyard.Contains(first.CardInstanceID) {
					t.Fatal("original target mutated during copy")
				}
			})
		})
	}
}

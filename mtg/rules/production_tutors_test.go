package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type productionTutorCase struct {
	entry   cardset.Entry
	mana    []mana.Color
	quality types.Card
	subtype types.Sub
	instant bool
}

func productionTutorCases() []productionTutorCase {
	return []productionTutorCase{
		{entry: authored.EladamrisCall, mana: []mana.Color{mana.G, mana.W}, quality: types.Creature, instant: true},
		{entry: authored.SteelshapersGift, mana: []mana.Color{mana.W}, quality: types.Artifact, subtype: types.Equipment},
		{entry: authored.Fabricate, mana: []mana.Color{mana.C, mana.C, mana.U}, quality: types.Artifact},
		{entry: authored.IdyllicTutor, mana: []mana.Color{mana.C, mana.C, mana.W}, quality: types.Enchantment},
		{entry: authored.SylvanScrying, mana: []mana.Color{mana.C, mana.G}, quality: types.Land},
	}
}

func productionTutorFind(tc productionTutorCase) *game.CardDef {
	def := vanillaCreature("Selected card", 2, 2)
	if tc.quality != "" {
		def.Types = append(def.Types, tc.quality)
	}
	if tc.subtype != "" {
		def.Subtypes = []types.Sub{tc.subtype}
	}
	return def
}

func TestProductionTutorsSearchChoicesAndVisibility(t *testing.T) {
	t.Parallel()
	for _, tc := range productionTutorCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, scenario := range []string{"find", "decline", "empty", "no match"} {
				t.Run(scenario, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					var wanted game.ObjectID
					if scenario != "empty" {
						if scenario == "no match" {
							wanted = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonmatching card", Types: []types.Card{types.Sorcery}}})
						} else {
							wanted = addCardToLibrary(g, game.Player1, productionTutorFind(tc))
						}
					}
					before := g.Players[game.Player1].Library.Size()
					epoch := g.Players[game.Player1].Library.ReorderVersion()
					agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
						if request.Kind != game.ChoiceSearch || request.MinChoices != 0 || request.MaxChoices != 1 ||
							len(obs.LibraryLookable(game.Player1)) != before ||
							len(NewObservation(g, game.Player2).LibraryLookable(game.Player1)) != 0 {
							t.Fatal("search cardinality/private visibility not engine-owned")
						}
						for _, option := range request.Options {
							if option.Label != "Selected card" {
								t.Fatal("qualified search included a nonmatch")
							}
						}
						if scenario == "decline" {
							return nil
						}
						return request.DefaultSelection
					}}
					spell := addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
					log := &TurnLog{}
					engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent}, log)
					moved := scenario == "find"
					if g.Players[game.Player1].Hand.Contains(wanted) != moved ||
						!g.Players[game.Player1].Graveyard.Contains(spell) ||
						g.Players[game.Player1].Library.ReorderVersion() != epoch+1 ||
						countEventsOfKind(g.Events, game.EventLibrarySearched) != 1 ||
						countEventsOfKind(g.Events, game.EventCardDrawn) != 0 {
						t.Fatal("find/mandatory find/fail-to-find/shuffle incorrectly implemented")
					}
					wantReveals := 0
					if moved {
						wantReveals = 1
					}
					if len(searchPartitionReveals(g)) != wantReveals ||
						len(NewObservation(g, game.Player2).HandOf(game.Player1)) != 0 ||
						len(g.SearchControlScopes) != 0 || len(NewObservation(g, game.Player1).LibraryLookable(game.Player1)) != 0 {
						t.Fatal("unrestricted tutor revealed its find or scoped library access persisted")
					}
					if scenario == "decline" && !moved && len(log.Choices) != 0 && log.Choices[0].UsedFallback {
						t.Fatal("qualified fail-to-find was rejected")
					}
				})
			}
		})
	}
}

func TestProductionTutorsCastTimingAndPayment(t *testing.T) {
	t.Parallel()
	for _, tc := range append(productionTutorCases(), productionPrivateTutorCases()...) {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, tc.entry)
			find := addCardToLibrary(g, game.Player1, productionTutorFind(tc))
			spell := addCardToHand(g, game.Player1, tc.entry.New())
			setMainPhasePriority(g, game.Player1)
			act := action.CastSpell(spell, nil, 0, nil)
			if engine.applyAction(g, game.Player1, act) {
				t.Fatal("cast without printed cost")
			}
			for _, c := range tc.mana {
				g.Players[game.Player1].ManaPool.Add(c, 1)
			}
			g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
			if cast := engine.applyAction(g, game.Player1, act); cast != tc.instant {
				t.Fatal("instant/sorcery timing incorrect")
			}
			if tc.instant {
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player1].Hand.Contains(find) {
					t.Fatal("instant tutor did not resolve during combat")
				}
				return
			}

			setMainPhasePriority(g, game.Player1)
			if engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) ||
				!engine.applyAction(g, game.Player1, act) {
				t.Fatal("untargeted tutor cardinality/casting failed")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(find) {
				t.Fatal("actual paid tutor did not execute")
			}
		})
	}
}

func TestProductionPrivateTutorPlacementBlocker(t *testing.T) {
	for _, name := range []string{"Demonic Tutor", "Diabolic Tutor"} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			found := addCardToLibrary(g, game.Player1, vanillaCreature("Private find", 2, 2))
			g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
				ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Library,
				MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
			})
			def := &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery},
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.Search{
					Player: game.ControllerReference(), Amount: game.Fixed(1),
					Spec: game.SearchSpec{SourceZone: zone.Library, Destination: zone.Hand, FailToFindPolicy: game.SearchMustFindIfAvailable},
				}}}}.Ability()),
			}}
			addImplementationSpellToStack(g, game.Player1, def, nil)
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Exile.Contains(found) || g.Players[game.Player1].Hand.Contains(found) || len(searchPartitionReveals(g)) != 0 {
				t.Fatal("private find must use the actual replacement destination without revealing")
			}
		})
	}
}

func TestProductionTutorsReplacementAndSearchControl(t *testing.T) {
	t.Parallel()
	for _, tc := range append(productionTutorCases(), productionPrivateTutorCases()...) {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, opposition := range []bool{false, true} {
				t.Run(map[bool]string{false: "hand destination replacement", true: "Opposition Agent"}[opposition], func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					find := addCardToLibrary(g, game.Player1, productionTutorFind(tc))
					if opposition {
						addCombatPermanent(g, game.Player4, oppositionAgentDef())
					} else {
						g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
							ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
							MatchFromZone: true, FromZone: zone.Library, MatchToZone: true, ToZone: zone.Hand,
							ReplaceToZone: zone.Exile,
						})
					}
					actor := game.Player1
					if opposition {
						actor = game.Player4
					}
					agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
						if obs.Player != game.Player1 || obs.DecisionMaker != actor ||
							len(obs.LibraryLookable(game.Player1)) != 1 ||
							len(NewObservation(g, game.Player3).LibraryLookable(game.Player1)) != 0 {
							t.Fatal("search-control decision or visibility incorrect")
						}
						return request.DefaultSelection
					}}
					policies := [game.NumPlayers]PlayerAgent{}
					policies[actor] = agent
					addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
					engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
					if !g.Players[game.Player1].Exile.Contains(find) || g.Players[game.Player1].Hand.Contains(find) ||
						len(g.SearchControlScopes) != 0 {
						t.Fatal("actual destination/search replacement was bypassed")
					}
					if opposition {
						permission, exists := playFromExilePermissionFor(g, find)
						if !exists || permission.Controller != game.Player4 {
							t.Fatal("Opposition Agent lost play permission")
						}
					}
				})
			}
		})
	}
}

func TestProductionTutorsCopyAndClone(t *testing.T) {
	t.Parallel()
	for _, tc := range append(productionTutorCases(), productionPrivateTutorCases()...) {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, tc.entry)
			first := addCardToLibrary(g, game.Player1, productionTutorFind(tc))
			second := addCardToLibrary(g, game.Player3, productionTutorFind(tc))
			addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
			subject, _ := g.Stack.Peek()
			source := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player3, Targets: []game.Target{game.StackObjectTarget(subject.ID)}}
			newEffectResolver(engine, g, source, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).resolveInstruction(&game.Instruction{Primitive: game.CopyStackObject{
				Object: game.TargetStackObjectReference(0),
			}})
			before := g.Clone()
			engine.resolveTopOfStack(before, &TurnLog{})
			if g.Players[game.Player3].Hand.Contains(second) || !before.Players[game.Player3].Hand.Contains(second) {
				t.Fatal("clone shared search placement")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player3].Hand.Contains(second) || !g.Players[game.Player1].Library.Contains(first) {
				t.Fatal("copy searched original caster's library")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(first) || !g.Stack.IsEmpty() {
				t.Fatal("original tutor changed with its copy")
			}
		})
	}
}

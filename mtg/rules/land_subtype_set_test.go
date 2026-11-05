package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review (Engine gap #3622, pinned corpus 6da9ced0…):
// Spreading Seas (oracle af32757c-9d3c-4f6c-a6fa-d5729a440382, {1}{U} Aura):
// "Enchant land / When this Aura enters, draw a card. / Enchanted land is an
// Island." Ruling 2009-10-01: the enchanted land loses its existing land types
// and any abilities printed on it, has the land type Island and can tap for
// {U}; its name and legendary/basic/snow status are unchanged.
// Song of the Dryads (oracle 7c3944fa-7c86-4979-85a9-86196aa94594): "Enchanted
// permanent is a colorless Forest land." Ruling 2014-11-07: it loses previous
// card types, subtypes and colors, keeps supertypes and name, gains "{T}: Add
// {G}" and loses all other abilities from its rules text but keeps abilities
// gained from other effects.
// Forest (oracle b34bb2dc-c1af-4d77-b0b3-a0fb342a5fc6) "({T}: Add {G}.)";
// Island (oracle b2c6aa39-2d2a-459c-a555-fb48ba993373) "({T}: Add {U}.)";
// Llanowar Elves (oracle 68954295-54e3-4303-a6bc-fc4547a4e3a3) {G} "{T}: Add
// {G}."; Urborg, Tomb of Yawgmoth (oracle db6174d7-211d-4817-b8e4-8384594c83f9)
// "Each land is a Swamp in addition to its other land types."
// Imprisoned in the Moon (oracle 339a7418-9a29-4eae-a3c2-ea590d175936,
// {2}{U} Aura): "Enchant creature, land, or planeswalker / Enchanted permanent
// is a colorless land with "{T}: Add {C}" and loses all other card types and
// abilities." Ruling 2016-07-13: a land keeps its land types but loses the
// intrinsic mana abilities associated with them (a Plains taps only for {C}).
// Comprehensive Rules read: MagicCompRules 20260925.txt (effective September
// 25, 2026, from magic.wizards.com/en/rules). 205.1a: setting subtypes replaces
// only existing subtypes from the appropriate set (land types). 305.6: a land
// with a basic land type has its intrinsic "{T}: Add [mana]". 305.7: setting a
// land's subtype to a basic land type removes its old land types and abilities
// from rules text, old land types and copiable effects, grants the new types'
// mana abilities, keeps abilities granted by other effects and card types and
// supertypes; gaining land types "in addition" keeps old types and abilities.
// 613.1d/613.1f: type changes apply in layer 4, ability grants and removal in
// layer 6. 613.8a: an effect that removes another type effect's source ability
// is applied first (dependency).

// landSetCard returns the validated default-registry card with name.
func landSetCard(t *testing.T, registry *cards.Registry, name string) *game.CardDef {
	t.Helper()
	def := lookupRegisteredCard(t, registry, name)
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("%s declaration: %+v", name, issues)
	}
	return def
}

// castAuraOnto casts the registered Aura from hand at target, resolves it and
// its enters trigger, and returns the Aura permanent.
func castAuraOnto(t *testing.T, g *game.Game, e *Engine, def *game.CardDef, target *game.Permanent, payment ...mana.Color) *game.Permanent {
	t.Helper()
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library filler", Types: []types.Card{types.Land}}})
	card := addCardToHand(g, game.Player1, def)
	for _, c := range payment {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	if !e.applyAction(g, game.Player1, action.CastSpell(card, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatalf("%s could not be cast at its target", def.Name)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	aura := permanentByCardID(g, card)
	if aura == nil || aura.AttachedTo != opt.Val(target.ObjectID) {
		t.Fatalf("%s did not resolve attached to its target", def.Name)
	}
	for e.putTriggeredAbilitiesOnStack(g) {
		e.resolveTopOfStack(g, &TurnLog{})
	}
	if g.Stack.Size() != 0 {
		t.Fatal("stack not empty after the Aura's triggers resolved")
	}
	return aura
}

func effectiveManaColors(g *game.Game, permanent *game.Permanent) []color.Color {
	_, colors := abilitiesManaProduction(permanentEffectiveAbilitiesView(g, permanent), nil, nil)
	return colors
}

func manaAbilityIndexes(g *game.Game, permanent *game.Permanent) []int {
	var indexes []int
	for i, ability := range permanentEffectiveAbilitiesView(g, permanent) {
		if _, ok := ability.(*game.ManaAbility); ok {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func TestSpreadingSeasForestLosesGreenManaForLlanowarElves(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	forest := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Forest"))
	otherForest := addCombatPermanent(g, game.Player2, landSetCard(t, registry, "Forest"))
	castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), forest, mana.C, mana.U)

	values := effectivePermanentValues(g, forest)
	if !slices.Equal(values.subtypes, []types.Sub{types.Island}) || values.name != "Forest" ||
		!slices.Contains(values.supertypes, types.Basic) || !slices.Equal(values.types, []types.Card{types.Land}) {
		t.Fatalf("enchanted Forest characteristics = %+v, want a basic land named Forest that is only an Island", values)
	}
	if got := effectiveManaColors(g, forest); !slices.Equal(got, []color.Color{color.Blue}) {
		t.Fatalf("enchanted Forest mana colors = %v, want only blue", got)
	}
	view := findPermanentView(t, NewObservation(g, game.Player1), forest.ObjectID)
	if !slices.Equal(view.ProducesColors, []color.Color{color.Blue}) {
		t.Fatalf("observed production = %v, want only blue", view.ProducesColors)
	}
	if got := effectiveManaColors(g, otherForest); !slices.Equal(got, []color.Color{color.Green}) {
		t.Fatalf("unenchanted Forest mana colors = %v, want green", got)
	}

	elves := addCardToHand(g, game.Player1, landSetCard(t, registry, "Llanowar Elves"))
	castElves := action.CastSpell(elves, nil, 0, nil)
	if containsAction(e.legalActions(g, game.Player1), castElves) {
		t.Fatal("Llanowar Elves offered with only an Island-enchanted Forest for mana")
	}
	if e.applyAction(g, game.Player1, castElves) || forest.Tapped {
		t.Fatal("the enchanted Forest paid {G} for Llanowar Elves")
	}

	blue := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Blue payment witness", Types: []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{cost.U}),
	}})
	if !e.applyAction(g, game.Player1, action.CastSpell(blue, nil, 0, nil)) || !forest.Tapped {
		t.Fatal("automatic payment did not use the enchanted Forest's Island mana ability")
	}
	e.resolveTopOfStack(g, &TurnLog{})

	forest.Tapped = false
	if n := len(permanentEffectiveAbilitiesView(g, forest)); n != 1 || !slices.Equal(manaAbilityIndexes(g, forest), []int{0}) {
		t.Fatalf("enchanted Forest has %d abilities, want only the Island mana ability", n)
	}
	if e.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 1, nil, 0)) {
		t.Fatal("activated a second (printed Forest) ability")
	}
	if !e.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("direct activation of the Island mana ability failed")
	}
	pool := g.Players[game.Player1].ManaPool
	if pool.Amount(mana.U) != 1 || pool.Amount(mana.G) != 0 {
		t.Fatalf("direct activation produced U=%d G=%d, want only one blue", pool.Amount(mana.U), pool.Amount(mana.G))
	}
}

func TestSpreadingSeasOnIslandKeepsSingleBlueAbility(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	island := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Island"))
	castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), island, mana.C, mana.U)
	if got := manaAbilityIndexes(g, island); len(got) != 1 {
		t.Fatalf("Island under Spreading Seas has mana abilities %v, want one", got)
	}
	if got := effectiveManaColors(g, island); !slices.Equal(got, []color.Color{color.Blue}) {
		t.Fatalf("Island under Spreading Seas mana colors = %v, want blue", got)
	}
}

// A land creature keeps its card types and creature types (CR 205.1a, 305.7);
// an ability granted by another effect survives, while a printed basic land type
// fixture without declared mana abilities pays only with its new type.
func TestSpreadingSeasKeepsNonLandSubtypesAndGrantedAbilities(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	arbor := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Arbor fixture", Colors: []color.Color{color.Green},
		Types: []types.Card{types.Land, types.Creature}, Subtypes: []types.Sub{types.Forest, types.Dryad},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody},
	}})
	red := game.TapManaAbility(mana.R)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 1, AffectedObjectID: arbor.ObjectID,
		Layer: game.LayerAbility, AddAbilities: []game.Ability{&red},
	})
	castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), arbor, mana.C, mana.U)

	values := effectivePermanentValues(g, arbor)
	if !slices.Contains(values.types, types.Creature) || !slices.Contains(values.subtypes, types.Dryad) ||
		!slices.Contains(values.subtypes, types.Island) || slices.Contains(values.subtypes, types.Forest) {
		t.Fatalf("land creature under Spreading Seas = types %v subtypes %v", values.types, values.subtypes)
	}
	if hasKeyword(g, arbor, game.Indestructible) {
		t.Fatal("printed ability survived the basic land type setting")
	}
	got := effectiveManaColors(g, arbor)
	if !slices.Contains(got, color.Blue) || !slices.Contains(got, color.Red) || slices.Contains(got, color.Green) {
		t.Fatalf("mana colors = %v, want blue and granted red without green", got)
	}

	arbor.SummoningSick = false
	green := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Green payment witness", Types: []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{cost.G}),
	}})
	if e.applyAction(g, game.Player1, action.CastSpell(green, nil, 0, nil)) || arbor.Tapped {
		t.Fatal("old Forest type paid {G} through the intrinsic basic-land fallback")
	}
}

// Gaining a land type in addition keeps old types and abilities (CR 305.7).
// With Urborg entering after Spreading Seas, the Forest is an Island and a
// Swamp; Spreading Seas on Urborg removes Urborg's ability, so its effect stops
// applying (CR 613.8a).
func TestSpreadingSeasWithUrborgAdditiveAndDependency(t *testing.T) {
	t.Run("enchanted Forest gains Urborg's Swamp", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		forest := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Forest"))
		castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), forest, mana.C, mana.U)
		addCombatPermanent(g, game.Player2, landSetCard(t, registry, "Urborg, Tomb of Yawgmoth"))
		values := effectivePermanentValues(g, forest)
		if !slices.Contains(values.subtypes, types.Island) || !slices.Contains(values.subtypes, types.Swamp) ||
			slices.Contains(values.subtypes, types.Forest) {
			t.Fatalf("subtypes = %v, want Island and Swamp", values.subtypes)
		}
		got := effectiveManaColors(g, forest)
		if !slices.Contains(got, color.Blue) || !slices.Contains(got, color.Black) || slices.Contains(got, color.Green) {
			t.Fatalf("mana colors = %v, want blue and black only", got)
		}
	})
	t.Run("Urborg made an Island loses its ability", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		urborg := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Urborg, Tomb of Yawgmoth"))
		forest := addCombatPermanent(g, game.Player2, landSetCard(t, registry, "Forest"))
		if !permanentHasSubtype(g, forest, types.Swamp) {
			t.Fatal("Urborg did not make the Forest a Swamp")
		}
		castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), urborg, mana.C, mana.U)
		if permanentHasSubtype(g, forest, types.Swamp) || permanentHasSubtype(g, urborg, types.Swamp) {
			t.Fatal("Urborg's removed ability still made lands Swamps")
		}
		if got := effectiveManaColors(g, urborg); !slices.Equal(got, []color.Color{color.Blue}) {
			t.Fatalf("Urborg mana colors = %v, want blue", got)
		}
		if got := effectiveManaColors(g, forest); !slices.Equal(got, []color.Color{color.Green}) {
			t.Fatalf("Forest mana colors = %v, want green", got)
		}
	})
}

// Song of the Dryads sets the card type too, so all subtypes are replaced and
// the permanent loses its rules-text abilities but keeps external grants.
func TestSongOfTheDryadsReplacesLandAbilities(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	island := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Island"))
	elves := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Llanowar Elves"))
	white := game.TapManaAbility(mana.W)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Timestamp: 1, AffectedObjectID: elves.ObjectID,
		Layer: game.LayerAbility, AddAbilities: []game.Ability{&white},
	})
	song := landSetCard(t, registry, "Song of the Dryads")
	castAuraOnto(t, g, e, song, island, mana.C, mana.C, mana.G)
	castAuraOnto(t, g, e, song, elves, mana.C, mana.C, mana.G)

	if got := effectiveManaColors(g, island); !slices.Equal(got, []color.Color{color.Green}) {
		t.Fatalf("Island under Song mana colors = %v, want green", got)
	}
	values := effectivePermanentValues(g, elves)
	if !slices.Equal(values.types, []types.Card{types.Land}) || !slices.Equal(values.subtypes, []types.Sub{types.Forest}) ||
		len(values.colors) != 0 || values.name != "Llanowar Elves" {
		t.Fatalf("Elves under Song = %+v, want a colorless Forest land", values)
	}
	got := effectiveManaColors(g, elves)
	if len(manaAbilityIndexes(g, elves)) != 2 || !slices.Contains(got, color.Green) || !slices.Contains(got, color.White) {
		t.Fatalf("Elves under Song mana colors = %v, want the Forest ability and the granted white ability only", got)
	}
}

// castManaWitness reports whether automatic payment casts a one-symbol
// enchantment, and resolves it when it does.
func castManaWitness(g *game.Game, e *Engine, symbol cost.Symbol) bool {
	witness := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Payment witness", Types: []types.Card{types.Enchantment},
		ManaCost: opt.Val(cost.Mana{symbol}),
	}})
	if !e.applyAction(g, game.Player1, action.CastSpell(witness, nil, 0, nil)) {
		return false
	}
	e.resolveTopOfStack(g, &TurnLog{})
	return true
}

// A land's basic land types grant one intrinsic mana ability each (CR 305.6),
// whether its face states them as one "{T}: Add {U} or {B}." reminder or not;
// automatic payment can use every one of them.
func TestBasicLandTypeIntrinsicAbilitiesPayAutomatically(t *testing.T) {
	for _, tc := range []struct {
		land    string
		symbols []cost.Symbol
	}{
		{land: "Underground Sea", symbols: []cost.Symbol{cost.U, cost.B}},
		{land: "Ketria Triome", symbols: []cost.Symbol{cost.G, cost.U, cost.R}},
	} {
		for _, symbol := range tc.symbols {
			t.Run(tc.land+" pays "+symbol.String(), func(t *testing.T) {
				g, e, registry := lifetimeGame(t)
				land := addCombatPermanent(g, game.Player1, landSetCard(t, registry, tc.land))
				if !castManaWitness(g, e, symbol) || !land.Tapped {
					t.Fatalf("untapped %s did not automatically pay %s", tc.land, symbol)
				}
			})
		}
	}
	t.Run("Underground Sea pays for Ponder", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		sea := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Underground Sea"))
		ponder := addCardToHand(g, game.Player1, landSetCard(t, registry, "Ponder"))
		if !e.applyAction(g, game.Player1, action.CastSpell(ponder, nil, 0, nil)) || !sea.Tapped {
			t.Fatal("untapped Underground Sea could not automatically pay {U} for Ponder")
		}
	})
	t.Run("undeclared Island Swamp pays black", func(t *testing.T) {
		g, e, _ := lifetimeGame(t)
		land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Island Swamp fixture", Types: []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp},
		}})
		if !castManaWitness(g, e, cost.B) || !land.Tapped {
			t.Fatal("a land without declared mana abilities did not pay with its second basic land type")
		}
	})
}

// Setting a dual land's subtype (CR 305.7) or removing its abilities removes
// the intrinsic abilities its old land types gave; automatic payment must not
// restore them.
func TestRemovedDualLandAbilitiesStayRemovedForPayment(t *testing.T) {
	t.Run("Spreading Seas", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		sea := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Underground Sea"))
		castAuraOnto(t, g, e, landSetCard(t, registry, "Spreading Seas"), sea, mana.C, mana.U)
		if sea.Tapped {
			t.Fatal("Spreading Seas was paid with the Underground Sea")
		}
		if castManaWitness(g, e, cost.B) || sea.Tapped {
			t.Fatal("Underground Sea made an Island paid {B}")
		}
		if !castManaWitness(g, e, cost.U) || !sea.Tapped {
			t.Fatal("Underground Sea made an Island did not pay {U}")
		}
	})
	t.Run("Song of the Dryads", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		sea := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Underground Sea"))
		castAuraOnto(t, g, e, landSetCard(t, registry, "Song of the Dryads"), sea, mana.C, mana.C, mana.G)
		if castManaWitness(g, e, cost.U) || castManaWitness(g, e, cost.B) || sea.Tapped {
			t.Fatal("Underground Sea made a Forest paid with an old land type")
		}
		if !castManaWitness(g, e, cost.G) || !sea.Tapped {
			t.Fatal("Underground Sea made a Forest did not pay {G}")
		}
	})
	t.Run("loses all abilities", func(t *testing.T) {
		g, e, registry := lifetimeGame(t)
		sea := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Underground Sea"))
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), Timestamp: 1, AffectedObjectID: sea.ObjectID,
			Layer: game.LayerAbility, RemoveAllAbilities: true,
		})
		if castManaWitness(g, e, cost.U) || castManaWitness(g, e, cost.B) || sea.Tapped {
			t.Fatal("Underground Sea without abilities paid mana")
		}
	})
}

// A basic land type's intrinsic mana ability (CR 305.6) is an ability like any
// other: an ability-removing effect (CR 613.1f) removes it while the land keeps
// its type, whether a face declares it, the land is unprinted fixture text, or
// it came from a type-setting effect. Automatic payment reads the same
// effective abilities, so none of these lands pays {G}.
func TestImprisonedInTheMoonRemovesIntrinsicLandManaAbilities(t *testing.T) {
	for _, tc := range []struct {
		name    string
		untyped bool
		setup   func(*testing.T, *game.Game, *Engine, *cards.Registry) *game.Permanent
	}{
		{name: "registered Forest", setup: func(t *testing.T, g *game.Game, _ *Engine, registry *cards.Registry) *game.Permanent {
			return addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Forest"))
		}},
		{name: "undeclared Forest fixture", setup: func(_ *testing.T, g *game.Game, _ *Engine, _ *cards.Registry) *game.Permanent {
			return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Forest fixture", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
			}})
		}},
		{name: "fixture named Forest", untyped: true, setup: func(_ *testing.T, g *game.Game, _ *Engine, _ *cards.Registry) *game.Permanent {
			return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Forest", Types: []types.Card{types.Land},
			}})
		}},
		{name: "Grizzly Bears made a Forest by Song of the Dryads", setup: func(t *testing.T, g *game.Game, e *Engine, registry *cards.Registry) *game.Permanent {
			bears := addCombatPermanent(g, game.Player1, landSetCard(t, registry, "Grizzly Bears"))
			castAuraOnto(t, g, e, landSetCard(t, registry, "Song of the Dryads"), bears, mana.C, mana.C, mana.G)
			return bears
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, e, registry := lifetimeGame(t)
			land := tc.setup(t, g, e, registry)
			if got := effectiveManaColors(g, land); !slices.Equal(got, []color.Color{color.Green}) {
				t.Fatalf("mana colors before Imprisoned in the Moon = %v, want green", got)
			}
			castAuraOnto(t, g, e, landSetCard(t, registry, "Imprisoned in the Moon"), land, mana.C, mana.C, mana.U)
			if land.Tapped {
				t.Fatal("the land paid for Imprisoned in the Moon")
			}
			if !tc.untyped && !permanentHasSubtype(g, land, types.Forest) {
				t.Fatal("Imprisoned in the Moon removed the Forest land type")
			}
			if got := effectiveManaColors(g, land); len(got) != 0 || len(manaAbilityIndexes(g, land)) != 1 {
				t.Fatalf("mana colors = %v with abilities %v, want only the colorless ability", got, manaAbilityIndexes(g, land))
			}
			view := findPermanentView(t, NewObservation(g, game.Player1), land.ObjectID)
			if len(view.ProducesColors) != 0 {
				t.Fatalf("observed production = %v, want colorless only", view.ProducesColors)
			}
			if castManaWitness(g, e, cost.G) || land.Tapped {
				t.Fatal("automatic payment used a removed intrinsic Forest ability for {G}")
			}
			if !castManaWitness(g, e, cost.O(1)) || !land.Tapped {
				t.Fatal("automatic payment did not use the colorless ability for {1}")
			}
		})
	}
}

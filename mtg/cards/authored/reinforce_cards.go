package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	burrentonBombardierOracleID = "13ec2390-e08a-4645-99e9-bac7e41917bb"
	breakTiesOracleID           = "1e44bb4e-aeea-4f43-9cbd-4c24a9b6b022"
	fowlStrikeOracleID          = "390320f7-11b8-4b61-8316-36c396470a09"
	swellOfCourageOracleID      = "60033841-71c7-4370-a2bc-e570a5c100ea"
	huntingTriadOracleID        = "66bf0c4b-0c36-4d63-a929-5a7427a07279"
	mosquitoGuardOracleID       = "6beaeec2-b4e9-4904-b817-c0118549d4a9"
	earthbrawnOracleID          = "91edf44a-0f46-4936-909a-934e2c4fe176"
	wrensRunHydraOracleID       = "a2196bc3-2fe6-49d4-8c79-d6beb3d6c5d8"
	rusticClachanOracleID       = "cde68428-0033-4ede-92f1-ab91de0a41fb"
	bannerhideKrushokOracleID   = "e3191093-1aaa-4002-a603-7c8d132b6a31"
)

var (
	BurrentonBombardier = cardset.Entry{OracleID: burrentonBombardierOracleID, Name: "Burrenton Bombardier", New: newBurrentonBombardier, Verification: "mtg/rules/reinforce_cards_test.go: TestBurrentonBombardierReinforceFromHand"}
	BreakTies           = cardset.Entry{OracleID: breakTiesOracleID, Name: "Break Ties", New: newBreakTies, Verification: "mtg/rules/reinforce_cards_test.go: TestBreakTiesReinforceFromHand; TestBreakTiesSpellModes"}
	FowlStrike          = cardset.Entry{OracleID: fowlStrikeOracleID, Name: "Fowl Strike", New: newFowlStrike, Verification: "mtg/rules/reinforce_cards_test.go: TestFowlStrikeReinforceFromHand; TestFowlStrikeDestroysFlyingCreature"}
	SwellOfCourage      = cardset.Entry{OracleID: swellOfCourageOracleID, Name: "Swell of Courage", New: newSwellOfCourage, Verification: "mtg/rules/reinforce_cards_test.go: TestSwellOfCourageReinforceFromHand; TestSwellOfCourageBuffsYourCreatures"}
	HuntingTriad        = cardset.Entry{OracleID: huntingTriadOracleID, Name: "Hunting Triad", New: newHuntingTriad, Verification: "mtg/rules/reinforce_cards_test.go: TestHuntingTriadReinforceFromHand; TestHuntingTriadCreatesElfWarriorTokens"}
	MosquitoGuard       = cardset.Entry{OracleID: mosquitoGuardOracleID, Name: "Mosquito Guard", New: newMosquitoGuard, Verification: "mtg/rules/reinforce_cards_test.go: TestMosquitoGuardReinforceFromHand"}
	Earthbrawn          = cardset.Entry{OracleID: earthbrawnOracleID, Name: "Earthbrawn", New: newEarthbrawn, Verification: "mtg/rules/reinforce_cards_test.go: TestEarthbrawnReinforceFromHand; TestEarthbrawnTargetsCreature"}
	WrensRunHydra       = cardset.Entry{OracleID: wrensRunHydraOracleID, Name: "Wren's Run Hydra", New: newWrensRunHydra, Verification: "mtg/rules/reinforce_cards_test.go: TestWrensRunHydraReinforceFromHand; TestWrensRunHydraEntersWithCastX"}
	RusticClachan       = cardset.Entry{OracleID: rusticClachanOracleID, Name: "Rustic Clachan", New: newRusticClachan, Verification: "mtg/rules/reinforce_cards_test.go: TestRusticClachanReinforceFromHand; TestRusticClachanRevealOrEntersTappedAndProducesWhite"}
	BannerhideKrushok   = cardset.Entry{OracleID: bannerhideKrushokOracleID, Name: "Bannerhide Krushok", New: newBannerhideKrushok, Verification: "mtg/rules/reinforce_cards_test.go: TestBannerhideKrushokReinforceFromHand; TestBannerhideKrushokScavengesFromGraveyard"}
)

func reinforceActivatedAbility(text string, manaCost cost.Mana, amount game.Quantity) game.ActivatedAbility {
	return game.ActivatedAbility{
		Text: text, ManaCost: opt.Val(append(cost.Mana(nil), manaCost...)), ZoneOfFunction: zone.Hand,
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard this card", Amount: 1, Source: zone.Hand}},
		Content: game.Mode{
			Text: text,
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.AddCounter{Object: game.TargetPermanentReference(0), Amount: amount, CounterKind: counter.PlusOnePlusOne}}},
		}.Ability(),
	}
}

func newBurrentonBombardier() *game.CardDef {
	const reinforceText = "Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two +1/+1 counters on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "Burrenton Bombardier", OracleText: "Flying\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
		Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Kithkin, types.Soldier}, Colors: []color.Color{color.White},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}), StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(2), cost.W}, game.Fixed(2))},
	}}
}

func newBreakTies() *game.CardDef {
	const reinforceText = "Reinforce 1—{W} ({W}, Discard this card: Put a +1/+1 counter on target creature.)"
	const artifactText = "Destroy target artifact."
	const enchantmentText = "Destroy target enchantment."
	const graveyardText = "Exile target card from a graveyard."
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "Break Ties", OracleText: "Choose one —\n• " + artifactText + "\n• " + enchantmentText + "\n• " + graveyardText + "\n" + reinforceText,
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}), Types: []types.Card{types.Instant}, Colors: []color.Color{color.White},
		SpellAbility: opt.Val(game.AbilityContent{Modes: []game.Mode{
			{Text: artifactText, Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target artifact", Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}})}}, Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}}},
			{Text: enchantmentText, Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target enchantment", Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Enchantment}})}}, Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}}},
			{Text: graveyardText, Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target card from a graveyard", Allow: game.TargetAllowCard, TargetZone: zone.Graveyard}}, Sequence: []game.Instruction{{Primitive: game.MoveCard{Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Exile}}}},
		}, MinModes: 1, MaxModes: 1}),
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.W}, game.Fixed(1))},
	}}
}

func newFowlStrike() *game.CardDef {
	const spellText = "Destroy target creature with flying."
	const reinforceText = "Reinforce 2—{2}{G} ({2}{G}, Discard this card: Put two +1/+1 counters on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Green), CardFace: game.CardFace{
		Name: "Fowl Strike", OracleText: spellText + "\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.Green},
		SpellAbility:       opt.Val(game.Mode{Text: spellText, Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target creature with flying", Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying})}}, Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}}}.Ability()),
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(2), cost.G}, game.Fixed(2))},
	}}
}

func newSwellOfCourage() *game.CardDef {
	const spellText = "Creatures you control get +2/+2 until end of turn."
	const reinforceText = "Reinforce X—{X}{W}{W} ({X}{W}{W}, Discard this card: Put X +1/+1 counters on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "Swell of Courage", OracleText: spellText + "\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.W, cost.W}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.White},
		SpellAbility: opt.Val(game.Mode{Text: spellText, Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{ContinuousEffects: []game.ContinuousEffect{{
			Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
			Layer: game.LayerPowerToughnessModify, PowerDelta: 2, ToughnessDelta: 2,
		}}, Duration: game.DurationUntilEndOfTurn}}}}.Ability()),
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.X, cost.W, cost.W}, game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}))},
	}}
}

func newHuntingTriad() *game.CardDef {
	const spellText = "Create three 1/1 green Elf Warrior creature tokens."
	const reinforceText = "Reinforce 3—{3}{G} ({3}{G}, Discard this card: Put three +1/+1 counters on target creature.)"
	token := &game.CardDef{CardFace: game.CardFace{Name: "Elf Warrior Token", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Elf, types.Warrior}, Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}}
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Green), CardFace: game.CardFace{
		Name: "Hunting Triad", OracleText: spellText + "\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.G}),
		Types: []types.Card{types.Kindred, types.Sorcery}, Subtypes: []types.Sub{types.Elf}, Colors: []color.Color{color.Green},
		SpellAbility:       opt.Val(game.Mode{Text: spellText, Sequence: []game.Instruction{{Primitive: game.CreateToken{Amount: game.Fixed(3), Source: game.TokenDef(token)}}}}.Ability()),
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(3), cost.G}, game.Fixed(3))},
	}}
}

func newMosquitoGuard() *game.CardDef {
	const reinforceText = "Reinforce 1—{1}{W} ({1}{W}, Discard this card: Put a +1/+1 counter on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "Mosquito Guard", OracleText: "First strike\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.W}),
		Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Kithkin, types.Soldier}, Colors: []color.Color{color.White},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}), StaticAbilities: []game.StaticAbility{game.FirstStrikeStaticBody},
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(1), cost.W}, game.Fixed(1))},
	}}
}

func newEarthbrawn() *game.CardDef {
	const spellText = "Target creature gets +3/+3 until end of turn."
	const reinforceText = "Reinforce 1—{1}{G} ({1}{G}, Discard this card: Put a +1/+1 counter on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Green), CardFace: game.CardFace{
		Name: "Earthbrawn", OracleText: spellText + "\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.Green},
		SpellAbility:       opt.Val(game.Mode{Text: spellText, Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}}, Sequence: []game.Instruction{{Primitive: game.ModifyPT{Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(3), ToughnessDelta: game.Fixed(3), Duration: game.DurationUntilEndOfTurn}}}}.Ability()),
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(1), cost.G}, game.Fixed(1))},
	}}
}

func newWrensRunHydra() *game.CardDef {
	const entersText = "This creature enters with X +1/+1 counters on it."
	const reinforceText = "Reinforce X—{X}{G}{G} ({X}{G}{G}, Discard this card: Put X +1/+1 counters on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Green), CardFace: game.CardFace{
		Name: "Wren's Run Hydra", OracleText: "Reach\n" + entersText + "\n" + reinforceText, ManaCost: opt.Val(cost.Mana{cost.X, cost.G}),
		Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Hydra}, Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 0}), Toughness: opt.Val(game.PT{Value: 0}),
		StaticAbilities: []game.StaticAbility{game.ReachStaticBody}, ReplacementAbilities: []game.ReplacementAbility{game.EntersWithCountersReplacement(entersText, game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true})},
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.X, cost.G, cost.G}, game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}))},
	}}
}

func newRusticClachan() *game.CardDef {
	const entersText = "As this land enters, you may reveal a Kithkin card from your hand. If you don't, this land enters tapped."
	const reinforceText = "Reinforce 1—{1}{W} ({1}{W}, Discard this card: Put a +1/+1 counter on target creature.)"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.White), CardFace: game.CardFace{
		Name: "Rustic Clachan", OracleText: entersText + "\n{T}: Add {W}.\n" + reinforceText, Types: []types.Card{types.Land},
		ManaAbilities:        []game.ManaAbility{game.TapManaAbility(mana.W)},
		ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedUnlessPaidReplacement(entersText, game.ResolutionPayment{Prompt: "Reveal a Kithkin card?", AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalReveal, Amount: 1, Source: zone.Hand, SubtypesAny: cost.SubtypeSet{types.Kithkin}}}})},
		ActivatedAbilities:   []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(1), cost.W}, game.Fixed(1))},
	}}
}

func newBannerhideKrushok() *game.CardDef {
	const reinforceText = "Reinforce 2—{1}{G} ({1}{G}, Discard this card: Put two +1/+1 counters on target creature.)"
	const scavengeText = "Scavenge {5}{G}{G}"
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Green), CardFace: game.CardFace{
		Name: "Bannerhide Krushok", OracleText: "Trample\n" + reinforceText + "\n" + scavengeText + " ({5}{G}{G}, Exile this card from your graveyard: Put a number of +1/+1 counters equal to this card's power on target creature. Scavenge only as a sorcery.)",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.G}), Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Beast}, Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities:    []game.StaticAbility{game.TrampleStaticBody},
		ActivatedAbilities: []game.ActivatedAbility{reinforceActivatedAbility(reinforceText, cost.Mana{cost.O(1), cost.G}, game.Fixed(2)), game.ScavengeActivatedAbility(cost.Mana{cost.O(5), cost.G, cost.G})},
	}}
}

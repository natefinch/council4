package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Keyword represents an evergreen or commonly-used keyword ability (CR 702).
type Keyword int

// Keyword values enumerate supported keyword abilities.
const (
	KeywordNone Keyword = iota
	Devoid
	Changeling
	Deathtouch
	Defender
	DoubleStrike
	FirstStrike
	Flash
	Flying
	Haste
	Hexproof
	Indestructible
	Lifelink
	Menace
	Protection
	Reach
	Shroud
	Trample
	Vigilance
	Ward
	SplitSecond
	Equip
	Enchant
	Cycling
	Flashback
	Kicker
	Madness
	Morph
	Disguise
	Convoke
	Delve
	Suspend
	Storm
	Cascade
	Prowess
	Mutate
	Companion
	Ninjutsu
	Escape
	Foretell
	Craft
	Discover
	Eternalize
	Affinity
	Improvise
	Emerge
	Undying
	Persist
	Wither
	Infect
	Toxic
	Annihilator
	Exalted
	ReadAhead
	Horsemanship
	CumulativeUpkeep
	Riot
	Embalm
	Fear
	Shadow
	Intimidate
	Skulk
	Evolve
	Unleash
	Fabricate
	Flanking
	Outlast
	Scavenge
	Dethrone
	Rampage
	LivingWeapon
	Soulshift
	Landwalk
	Dredge
	Unearth
	Training
	Saddle
	Rebound
	Retrace
	// Banding (CR 702.22) is a static combat keyword. It is modeled as a
	// recognized, grantable simple keyword so cards that have or grant banding
	// are representable. Banding's combat damage-assignment-control nuance is not
	// simulated by the deterministic combat engine (which has no per-player
	// Appended at the end of the enum so existing keyword ordinals are unchanged.
	Banding
	// Crew (CR 702.122) is the Vehicles activated keyword. It is modeled as a
	// grantable keyword identity carried by CrewKeyword inside the activated
	// ability built by CrewActivatedAbility. Appended at the end of the enum so
	// existing keyword ordinals are unchanged.
	Crew
	// Fuse (CR 702.102) is printed on both halves of a fuse split card: "You may
	// cast one or both halves of this card from your hand." It is modeled as a
	// recognized simple keyword carried on each split face so fuse split cards
	// are representable and the rules layer can detect the fuse permission via
	// HasKeyword(Fuse). Appended at the end of the enum so existing keyword
	// ordinals are unchanged.
	Fuse
	// JumpStart (CR 702.134) is printed on instants and sorceries: "Jump-start
	// (You may cast this card from your graveyard by discarding a card in
	// addition to paying its other costs. Then exile this card.)" It is modeled
	// as a recognized simple keyword carried on the card so HasKeyword(JumpStart)
	// reports true; the rules layer reads it on a card in its owner's graveyard
	// to offer the graveyard cast with a discard additional cost and exile the
	// card on resolution. Appended at the end of the enum so existing keyword
	// ordinals are unchanged.
	JumpStart
	// PartnerWith (CR 702.124e) is the "Partner with <name>" keyword. It is
	// modeled as a recognized simple keyword carried on the card so
	// HasKeyword(PartnerWith) reports true. The "partner commander"
	// deck-construction permission and the pair-fetch enters trigger are not
	// simulated by the deterministic playtester, so the keyword is inert; it is
	// modeled as a simple keyword purely so partner-with cards are
	// representable. Appended at the end of the enum so existing keyword ordinals
	// are unchanged.
	PartnerWith
	// ChooseABackground (CR 702.124f) is the "Choose a Background" keyword: "You
	// can have a Background as a second commander." It is a deck-construction
	// permission the deterministic playtester does not simulate, so the keyword
	// is inert; it is modeled as a recognized simple keyword carried on the card
	// so HasKeyword(ChooseABackground) reports true and choose-a-background cards
	// are representable. Appended at the end of the enum so existing keyword
	// ordinals are unchanged.
	ChooseABackground
	// Reconfigure (CR 702.151) is the Equipment-creature attach keyword:
	// "Reconfigure <cost>" is a sorcery-speed activated ability that attaches the
	// source to target creature you control (and may unattach it). It is modeled
	// as a recognized keyword identity carried by ReconfigureKeyword inside the
	// activated ability built by ReconfigureActivatedAbility; the rules layer
	// treats it like Equip for attachment activation and resolution. The
	// "or unattach" mode and the "while attached, this isn't a creature"
	// type-change are not yet simulated. Appended at the end of the enum so
	// existing keyword ordinals are unchanged.
	Reconfigure
	// Partner (CR 702.124a) is the "Partner" keyword and its "Partner—<quality>"
	// restricted variants (CR 702.124f). It grants the "partner commander"
	// deck-construction permission; the restricted variants only narrow which
	// other partner cards a card may pair with. Both the permission and the
	// pairing restrictions are deck-construction mechanics the deterministic
	// playtester does not simulate, so the keyword is inert; it is modeled as a
	// recognized simple keyword carried on the card so HasKeyword(Partner)
	// reports true and partner cards are representable. Appended at the end of the
	// enum so existing keyword ordinals are unchanged.
	Partner
	// Hideaway N (CR 702.75) is the land keyword printed on the "Hideaway lands":
	// "When this permanent enters, look at the top N cards of your library, exile
	// one face down, then put the rest on the bottom in a random order." A later
	// activated ability lets the controller play that exiled card without paying
	// its mana cost when a condition is met. It is modeled as a parameterized
	// keyword carried by HideawayKeyword inside the enters-the-battlefield
	// triggered ability built by HideawayTriggeredAbility, so HasKeyword(Hideaway)
	// reports true and the rules layer runs the look/exile body. Appended at the
	// end of the enum so existing keyword ordinals are unchanged.
	Hideaway
)

// Reusable StaticAbilityBody templates for non-parameterized keyword abilities.
// Use these in CardFace.StaticAbilities slices or initializer-function appends.
// Treat these values as immutable.
var (
	// DevoidStaticBody is the reusable StaticAbilityBody for devoid.
	DevoidStaticBody = simpleKeywordStaticBody("Devoid", Devoid)

	// ChangelingStaticBody is the reusable StaticAbilityBody for changeling.
	ChangelingStaticBody = simpleKeywordStaticBody("Changeling", Changeling)

	// DeathtouchStaticBody is the reusable StaticAbilityBody for deathtouch.
	DeathtouchStaticBody = simpleKeywordStaticBody("Deathtouch", Deathtouch)

	// DefenderStaticBody is the reusable StaticAbilityBody for defender.
	DefenderStaticBody = simpleKeywordStaticBody("Defender", Defender)

	// DoubleStrikeStaticBody is the reusable StaticAbilityBody for double strike.
	DoubleStrikeStaticBody = simpleKeywordStaticBody("Double strike", DoubleStrike)

	// FirstStrikeStaticBody is the reusable StaticAbilityBody for first strike.
	FirstStrikeStaticBody = simpleKeywordStaticBody("First strike", FirstStrike)

	// FlashStaticBody is the reusable StaticAbilityBody for flash.
	FlashStaticBody = simpleKeywordStaticBody("Flash", Flash)

	// FlyingStaticBody is the reusable StaticAbilityBody for flying.
	FlyingStaticBody = simpleKeywordStaticBody("Flying", Flying)

	// HasteStaticBody is the reusable StaticAbilityBody for haste.
	HasteStaticBody = simpleKeywordStaticBody("Haste", Haste)

	// HexproofStaticBody is the reusable StaticAbilityBody for hexproof.
	HexproofStaticBody = simpleKeywordStaticBody("Hexproof", Hexproof)

	// IndestructibleStaticBody is the reusable StaticAbilityBody for indestructible.
	IndestructibleStaticBody = simpleKeywordStaticBody("Indestructible", Indestructible)

	// LifelinkStaticBody is the reusable StaticAbilityBody for lifelink.
	LifelinkStaticBody = simpleKeywordStaticBody("Lifelink", Lifelink)

	// MenaceStaticBody is the reusable StaticAbilityBody for menace.
	MenaceStaticBody = simpleKeywordStaticBody("Menace", Menace)

	// ReachStaticBody is the reusable StaticAbilityBody for reach.
	ReachStaticBody = simpleKeywordStaticBody("Reach", Reach)

	// ShroudStaticBody is the reusable StaticAbilityBody for shroud.
	ShroudStaticBody = simpleKeywordStaticBody("Shroud", Shroud)

	// TrampleStaticBody is the reusable StaticAbilityBody for trample.
	TrampleStaticBody = simpleKeywordStaticBody("Trample", Trample)

	// VigilanceStaticBody is the reusable StaticAbilityBody for vigilance.
	VigilanceStaticBody = simpleKeywordStaticBody("Vigilance", Vigilance)

	// SplitSecondStaticBody is the reusable StaticAbilityBody for split second.
	SplitSecondStaticBody = simpleKeywordStaticBody("Split second", SplitSecond)

	// ConvokeStaticBody is the reusable StaticAbilityBody for convoke.
	ConvokeStaticBody = simpleKeywordStaticBody("Convoke", Convoke)

	// DelveStaticBody is the reusable StaticAbilityBody for delve.
	DelveStaticBody = simpleKeywordStaticBody("Delve", Delve)

	// StormStaticBody is the reusable StaticAbilityBody for storm.
	StormStaticBody = simpleKeywordStaticBody("Storm", Storm)

	// CascadeStaticBody is the reusable StaticAbilityBody for cascade.
	CascadeStaticBody = simpleKeywordStaticBody("Cascade", Cascade)

	// ProwessStaticBody is the reusable StaticAbilityBody for prowess.
	ProwessStaticBody = simpleKeywordStaticBody("Prowess", Prowess)

	// ImproviseStaticBody is the reusable StaticAbilityBody for improvise.
	ImproviseStaticBody = simpleKeywordStaticBody("Improvise", Improvise)

	// UndyingStaticBody is the reusable StaticAbilityBody for undying.
	UndyingStaticBody = simpleKeywordStaticBody("Undying", Undying)

	// PersistStaticBody is the reusable StaticAbilityBody for persist.
	PersistStaticBody = simpleKeywordStaticBody("Persist", Persist)

	// WitherStaticBody is the reusable StaticAbilityBody for wither.
	WitherStaticBody = simpleKeywordStaticBody("Wither", Wither)

	// InfectStaticBody is the reusable StaticAbilityBody for infect.
	InfectStaticBody = simpleKeywordStaticBody("Infect", Infect)

	// ExaltedStaticBody is the reusable StaticAbilityBody for exalted.
	ExaltedStaticBody = simpleKeywordStaticBody("Exalted", Exalted)

	// ReadAheadStaticBody is the reusable StaticAbilityBody for read ahead.
	ReadAheadStaticBody = simpleKeywordStaticBody("Read ahead", ReadAhead)

	// HorsemanshipStaticBody is the reusable StaticAbilityBody for horsemanship,
	// an evasion ability: a creature with horsemanship can't be blocked except by
	// creatures with horsemanship (CR 702.31, Portal Three Kingdoms).
	HorsemanshipStaticBody = simpleKeywordStaticBody("Horsemanship", Horsemanship)

	// ShadowStaticBody is the reusable StaticAbilityBody for shadow, an evasion
	// ability: a creature with shadow can block or be blocked by only creatures
	// with shadow (CR 702.28).
	ShadowStaticBody = simpleKeywordStaticBody("Shadow", Shadow)

	// RiotStaticBody is the reusable StaticAbilityBody for riot. Riot is an
	// enters-the-battlefield keyword (CR 702.137): as a permanent with riot
	// enters, its controller chooses for it to enter with a +1/+1 counter or to
	// gain haste. The runtime reads the riot keyword on an entering permanent and
	// applies that modal choice; the keyword itself carries no continuous effect.
	RiotStaticBody = simpleKeywordStaticBody("Riot", Riot)

	// FearStaticBody is the reusable StaticAbilityBody for fear, an evasion
	// ability: a creature with fear can't be blocked except by artifact creatures
	// and/or black creatures (CR 702.36c).
	FearStaticBody = simpleKeywordStaticBody("Fear", Fear)

	// SkulkStaticBody is the reusable StaticAbilityBody for skulk, an evasion
	// ability: a creature with skulk can't be blocked by creatures with greater
	// power than it (CR 702.72b).
	SkulkStaticBody = simpleKeywordStaticBody("Skulk", Skulk)

	// IntimidateStaticBody is the reusable StaticAbilityBody for intimidate, an
	// evasion ability: a creature with intimidate can't be blocked except by
	// artifact creatures and/or creatures that share a color with it (CR 702.13b).
	IntimidateStaticBody = simpleKeywordStaticBody("Intimidate", Intimidate)

	// EvolveStaticBody is the reusable StaticAbilityBody for evolve (CR 702.100):
	// "Whenever a creature you control enters, if that creature has greater power
	// or toughness than this creature, put a +1/+1 counter on this creature." The
	// triggered ability is realized at runtime from the evolve keyword; the
	// keyword itself carries no continuous effect.
	EvolveStaticBody = simpleKeywordStaticBody("Evolve", Evolve)

	// UnleashStaticBody is the reusable StaticAbilityBody for unleash (CR
	// 702.86): "You may have this creature enter with a +1/+1 counter on it. It
	// can't block as long as it has a +1/+1 counter on it." The runtime reads the
	// unleash keyword on an entering permanent to offer the optional +1/+1
	// counter, and prohibits blocking while such a permanent has a +1/+1 counter;
	// the keyword itself carries no continuous effect.
	UnleashStaticBody = simpleKeywordStaticBody("Unleash", Unleash)

	// ReboundStaticBody is the reusable StaticAbilityBody for rebound (CR
	// 702.88): "If this spell was cast from your hand, instead of putting it into
	// your graveyard as it resolves, exile it and, at the beginning of your next
	// upkeep, you may cast this card from exile without paying its mana cost."
	// The runtime reads the rebound keyword on a resolving spell; the keyword
	// itself carries no continuous effect.
	ReboundStaticBody = simpleKeywordStaticBody("Rebound", Rebound)

	// RetraceStaticBody is the reusable StaticAbilityBody for retrace (CR
	// 702.81): "You may cast this card from your graveyard by discarding a land
	// card in addition to paying its other costs." The runtime reads the retrace
	// keyword on a card in its owner's graveyard to offer the alternative cast;
	// the keyword itself carries no continuous effect.
	RetraceStaticBody = simpleKeywordStaticBody("Retrace", Retrace)

	// BandingStaticBody is the reusable StaticAbilityBody for banding (CR
	// 702.22). It carries the Banding keyword so HasKeyword(Banding) reports
	// true. Banding's combat damage-assignment-control rule is not simulated by
	// the deterministic combat engine, so the keyword is inert in combat; it is
	// modeled as a simple keyword purely so cards that have or grant banding are
	// representable.
	BandingStaticBody = simpleKeywordStaticBody("Banding", Banding)

	// CompanionStaticBody is the reusable StaticAbilityBody for companion (CR
	// 702.139): "Companion — <deckbuilding condition>." Companion is a static
	// ability that functions from outside the game: it lets a player begin with
	// the card in their sideboard as a designated companion and, once per game,
	// pay {3} to put it into their hand. Both halves are deck-construction and
	// sideboard mechanics the deterministic playtester does not simulate, so the
	// keyword carries no continuous in-game effect; it is modeled as a simple
	// keyword purely so companion cards are representable.
	CompanionStaticBody = simpleKeywordStaticBody("Companion", Companion)

	// PartnerWithStaticBody is the reusable StaticAbilityBody for the "Partner
	// with <name>" keyword (CR 702.124e). Partner with names a specific partner
	// card, grants the two cards the "partner commander" deck-construction
	// permission, and gives each an enters trigger that lets the chosen player
	// tutor the named partner into hand. Both halves are deck-construction and
	// pair-fetch mechanics the deterministic playtester does not simulate, so the
	// keyword carries no continuous in-game effect; it is modeled as a simple
	// keyword purely so partner-with cards are representable.
	PartnerWithStaticBody = simpleKeywordStaticBody("Partner with", PartnerWith)

	// ChooseABackgroundStaticBody is the reusable StaticAbilityBody for the
	// "Choose a Background" keyword (CR 702.124f): "You can have a Background as a
	// second commander." The permission is a deck-construction mechanic the
	// deterministic playtester does not simulate, so the keyword carries no
	// continuous in-game effect; it is modeled as a simple keyword purely so
	// choose-a-background cards are representable.
	ChooseABackgroundStaticBody = simpleKeywordStaticBody("Choose a Background", ChooseABackground)

	// PartnerStaticBody is the reusable StaticAbilityBody for the "Partner"
	// keyword (CR 702.124a) and its "Partner—<quality>" restricted variants (CR
	// 702.124f, e.g. "Partner—Survivors", "Partner—Character select"). Partner
	// grants the two cards the "partner commander" deck-construction permission;
	// the restricted variants only narrow which other partner cards a card may
	// pair with. Both the permission and the pairing restrictions are
	// deck-construction mechanics the deterministic playtester does not simulate,
	// so the keyword carries no continuous in-game effect; it is modeled as a
	// simple keyword purely so partner cards are representable.
	PartnerStaticBody = simpleKeywordStaticBody("Partner", Partner)

	// FuseStaticBody is the reusable StaticAbility for fuse (CR 702.102). It
	// carries the Fuse keyword so HasKeyword(Fuse) reports true on each half of a
	// fuse split card. The fused-casting permission itself is granted by the
	// rules layer when it detects this keyword on a split card's faces.
	FuseStaticBody = simpleKeywordStaticBody("Fuse", Fuse)

	// JumpStartStaticBody is the reusable StaticAbility for jump-start (CR
	// 702.134): "Jump-start (You may cast this card from your graveyard by
	// discarding a card in addition to paying its other costs. Then exile this
	// card.)" It carries the JumpStart keyword so HasKeyword(JumpStart) reports
	// true on the card. The graveyard cast permission, the discard additional
	// cost, and the exile-on-resolution are supplied by the rules layer when it
	// detects this keyword on a card in its owner's graveyard.
	JumpStartStaticBody = simpleKeywordStaticBody("Jump-start", JumpStart)
)

func simpleKeywordStaticBody(text string, keyword Keyword) StaticAbility {
	return StaticAbility{Text: text, KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: keyword}}}
}

// Reusable TriggeredAbility templates for keywords that expand to a canonical
// triggered ability. Treat these values as immutable.
var (
	// UndyingTriggeredBody is the canonical triggered ability for undying
	// (CR 702.92): "When this creature dies, if it had no +1/+1 counters on it,
	// return it to the battlefield under its owner's control with a +1/+1 counter
	// on it." The ability carries the Undying keyword so HasKeyword(Undying)
	// reports true.
	UndyingTriggeredBody = diesReturnWithCounterTriggeredBody("Undying", Undying, counter.PlusOnePlusOne)

	// PersistTriggeredBody is the canonical triggered ability for persist
	// (CR 702.78): "When this creature dies, if it had no -1/-1 counters on it,
	// return it to the battlefield under its owner's control with a -1/-1 counter
	// on it." The ability carries the Persist keyword so HasKeyword(Persist)
	// reports true.
	PersistTriggeredBody = diesReturnWithCounterTriggeredBody("Persist", Persist, counter.MinusOneMinusOne)

	// DethroneTriggeredBody is the canonical triggered ability for dethrone
	// (CR 702.103): "Whenever this creature attacks the player with the most life
	// or tied for most life, put a +1/+1 counter on this creature." The ability
	// carries the Dethrone keyword so HasKeyword(Dethrone) reports true.
	DethroneTriggeredBody = dethroneTriggeredBody()

	// FlankingTriggeredBody is the canonical triggered ability for flanking
	// (CR 702.25): "Whenever a creature without flanking blocks this creature,
	// the blocking creature gets -1/-1 until end of turn." The ability carries
	// the Flanking keyword so HasKeyword(Flanking) reports true, which lets
	// another flanker's "without flanking" blocker filter exclude this creature.
	// Each printed instance is its own triggered ability, so multiple instances
	// stack to -N/-N (CR 702.25c).
	FlankingTriggeredBody = flankingTriggeredBody()

	// TrainingTriggeredBody is the canonical triggered ability for training
	// (CR 702.150): "Whenever this creature attacks with another creature with
	// greater power, put a +1/+1 counter on this creature." The ability carries
	// the Training keyword so HasKeyword(Training) reports true.
	TrainingTriggeredBody = trainingTriggeredBody()

	// StartEnginesTriggeredBody is the canonical triggered ability for the
	// "Start your engines!" keyword (CR 702.179): "When this permanent enters,
	// you get your speed. (If you have no speed, it starts at 1. ...)" Every
	// printed instance is on a permanent, so it is modeled as an enters-the-
	// battlefield trigger that runs the StartEngines primitive for the
	// controller, which sets their speed to 1 if they have none. The recurring
	// once-per-turn increase on opponent life loss is a built-in rule keyed off
	// the player's speed, so the ability itself only seeds the starting speed.
	StartEnginesTriggeredBody = startEnginesTriggeredBody()
)

func startEnginesTriggeredBody() TriggeredAbility {
	return TriggeredAbility{
		Text: "Start your engines!",
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: StartEngines{
				Player: ControllerReference(),
			},
		}}}.Ability(),
	}
}

func trainingTriggeredBody() TriggeredAbility {
	return TriggeredAbility{
		Text:             "Training",
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Training}},
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:                           EventAttackerDeclared,
				Source:                          TriggerSourceSelf,
				AttacksWithGreaterPowerCreature: true,
			},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: AddCounter{
				Amount:      Fixed(1),
				Object:      SourcePermanentReference(),
				CounterKind: counter.PlusOnePlusOne,
			},
		}}}.Ability(),
	}
}

func dethroneTriggeredBody() TriggeredAbility {
	return TriggeredAbility{
		Text:             "Dethrone",
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Dethrone}},
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:                     EventAttackerDeclared,
				Source:                    TriggerSourceSelf,
				AttackRecipient:           AttackRecipientPlayer,
				AttackedPlayerHasMostLife: true,
			},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: AddCounter{
				Amount:      Fixed(1),
				Object:      SourcePermanentReference(),
				CounterKind: counter.PlusOnePlusOne,
			},
		}}}.Ability(),
	}
}

func flankingTriggeredBody() TriggeredAbility {
	return TriggeredAbility{
		Text:             "Flanking",
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Flanking}},
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:  EventAttackerBecameBlocked,
				Source: TriggerSourceSelf,
				RelatedSubjectSelection: Selection{
					RequiredTypes:   []types.Card{types.Creature},
					ExcludedKeyword: Flanking,
				},
			},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: ModifyPT{
				Object:         EventRelatedPermanentReference(),
				PowerDelta:     Fixed(-1),
				ToughnessDelta: Fixed(-1),
				Duration:       DurationUntilEndOfTurn,
			},
		}}}.Ability(),
	}
}

// diesReturnWithCounterTriggeredBody builds the canonical undying/persist
// triggered ability: a self dies-trigger gated on the dying creature having had
// no counters of the given kind, returning it to the battlefield under its
// owner's control with one such counter (CR 702.78, CR 702.92). The triggered
// ability carries the keyword itself so HasKeyword reports the printed keyword.
func diesReturnWithCounterTriggeredBody(text string, keyword Keyword, kind counter.Kind) TriggeredAbility {
	return TriggeredAbility{
		Text:             text,
		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: keyword}},
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:            EventPermanentDied,
				Source:           TriggerSourceSelf,
				SubjectSelection: Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			InterveningIfEventPermanentHadNoCounterKind: opt.Val(kind),
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: PutOnBattlefield{
				Source:        CardBattlefieldSource(CardReference{Kind: CardReferenceEvent}),
				EntryCounters: []CounterPlacement{{Kind: kind, Amount: 1}},
			},
		}}}.Ability(),
	}
}

// KeywordStaticBody returns the reusable StaticAbility granting the given
// non-parameterized keyword ability. It reports ok=false for parameterized or
// unsupported keywords (ward, protection, equip, and similar) that cannot be
// expressed as a bare keyword grant. Use it to attach keywords granted to a
// created copy token ("That token gains haste").
func KeywordStaticBody(keyword Keyword) (StaticAbility, bool) {
	switch keyword {
	case Deathtouch:
		return DeathtouchStaticBody, true
	case Defender:
		return DefenderStaticBody, true
	case DoubleStrike:
		return DoubleStrikeStaticBody, true
	case FirstStrike:
		return FirstStrikeStaticBody, true
	case Flash:
		return FlashStaticBody, true
	case Flying:
		return FlyingStaticBody, true
	case Haste:
		return HasteStaticBody, true
	case Hexproof:
		return HexproofStaticBody, true
	case Indestructible:
		return IndestructibleStaticBody, true
	case Lifelink:
		return LifelinkStaticBody, true
	case Menace:
		return MenaceStaticBody, true
	case Reach:
		return ReachStaticBody, true
	case Shroud:
		return ShroudStaticBody, true
	case Trample:
		return TrampleStaticBody, true
	case Vigilance:
		return VigilanceStaticBody, true
	case Fear:
		return FearStaticBody, true
	case Skulk:
		return SkulkStaticBody, true
	case Intimidate:
		return IntimidateStaticBody, true
	default:
		return StaticAbility{}, false
	}
}

// TriggerType classifies what kind of event triggers a triggered ability.
type TriggerType int

// Trigger type values identify supported trigger wordings.
const (
	TriggerWhen     TriggerType = iota // "When [event]" — fires once
	TriggerWhenever                    // "Whenever [event]" — fires each time
	TriggerAt                          // "At the beginning of [step]"
	TriggerState                       // State trigger checked whenever a player would get priority
)

// TriggerCondition describes when a triggered ability fires.
type TriggerCondition struct {
	// Type is whether this is a When, Whenever, or At trigger.
	Type TriggerType

	// Pattern is the structured event pattern this ability listens for.
	Pattern TriggerPattern

	// InterveningIf is the "if" condition that must be true both when the
	// event occurs and when the trigger resolves (CR 603.4). Empty if none.
	InterveningIf string

	// InterveningCondition is the structured form of InterveningIf. The rules
	// layer evaluates it with the trigger controller and triggering event bound.
	InterveningCondition opt.V[Condition]

	// InterveningIfEventPermanentHadCounters is true for intervening-if clauses
	// such as "if it had counters on it" on zone-change triggers. mtg/rules
	// checks the event permanent's current object or last-known information.
	InterveningIfEventPermanentHadCounters bool

	// InterveningIfEventPermanentHadNoCounterKind identifies a counter kind that
	// must be absent from the event permanent's current object or last-known
	// information.
	InterveningIfEventPermanentHadNoCounterKind opt.V[counter.Kind]

	// InterveningIfEventPermanentHadCounterKind identifies a counter kind that
	// must be present (at least one) on the event permanent's current object or
	// last-known information, e.g. "if it had a +1/+1 counter on it" on a dies
	// trigger. It is the affirmative counterpart of
	// InterveningIfEventPermanentHadNoCounterKind.
	InterveningIfEventPermanentHadCounterKind opt.V[counter.Kind]

	// InterveningIfEventPermanentWasKicked is true for "if it was kicked" on
	// enter triggers. The entering permanent event preserves the spell's kicker
	// choice for both trigger-time and resolution-time checks.
	InterveningIfEventPermanentWasKicked bool

	// InterveningIfEventPermanentWasCast is true for "if it was cast" on enter
	// triggers.
	InterveningIfEventPermanentWasCast bool
	// InterveningIfEventPermanentWasEvoked is true for the evoke sacrifice
	// trigger's intervening "if its evoke cost was paid" condition (CR 702.74).
	// The entering permanent event preserves whether the spell was cast for its
	// Evoke alternative cost for both trigger-time and resolution-time checks.
	InterveningIfEventPermanentWasEvoked bool
	// InterveningIfEventPermanentWasCastByController is true for "if you cast
	// it" and additionally requires the trigger controller to be the caster.
	InterveningIfEventPermanentWasCastByController bool

	// InterveningIfEventPermanentWasCastFromControllerHand is true for "if you
	// cast it from your hand" on enter triggers. It requires the entering
	// permanent to have resulted from a cast the trigger controller made from
	// their hand, so mtg/rules additionally checks the caster and the cast
	// source zone. The entering permanent event preserves the cast's caster and
	// source zone for both trigger-time and resolution-time checks (CR 603.4).
	InterveningIfEventPermanentWasCastFromControllerHand bool

	// InterveningIfEventPermanentEnteredOrCastFromGraveyard is true for the
	// enter-trigger intervening "if" that gates on the entering object(s) having
	// come from any graveyard, either by entering directly from a graveyard or by
	// being cast from a graveyard ("if they entered or were cast from a
	// graveyard").
	InterveningIfEventPermanentEnteredOrCastFromGraveyard bool

	// InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard is the
	// controller-scoped variant gating on the entering object(s) having come from
	// the trigger controller's own graveyard ("if it entered from your graveyard
	// or you cast it from your graveyard"). It additionally requires the entering
	// card's owner to be the trigger controller, and the cast branch requires the
	// controller to be the caster.
	InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard bool

	// State describes a state trigger. State triggers latch while true and only
	// trigger again after becoming false, then true again (CR 603.8).
	State opt.V[StateTriggerCondition]
}

// StateTriggerCondition describes a simple state trigger condition. Empty
// fields mean no state condition is active.
type StateTriggerCondition struct {
	MatchControllerLifeLessOrEqual bool
	ControllerLifeLessOrEqual      int
}

// TriggerControllerFilter constrains a trigger by the controller recorded on an event.
type TriggerControllerFilter int

// Trigger controller filters match events by controller.
const (
	TriggerControllerAny TriggerControllerFilter = iota
	TriggerControllerYou
	TriggerControllerOpponent
)

// TriggerSourceFilter constrains a trigger by the source of the event.
type TriggerSourceFilter int

// Trigger source filters match events by source.
const (
	TriggerSourceAny TriggerSourceFilter = iota
	TriggerSourceSelf
	TriggerSourceAttachedPermanent
)

// TriggerSubjectObject identifies which permanent on an event is the trigger
// subject for source/controller matching. Event-specific object fields that are
// not the subject, such as EventBlockerDeclared.PermanentID for the blocker,
// continue to feed general permanent filters.
type TriggerSubjectObject int

// Trigger subject object values identify event permanent roles.
const (
	TriggerSubjectDefault TriggerSubjectObject = iota
	TriggerSubjectPermanent
	TriggerSubjectBlockedAttacker
	TriggerSubjectDamageSource
)

// TriggerPlayerFilter constrains a trigger by the affected player recorded on an event.
type TriggerPlayerFilter int

// Trigger player filters match events by affected player.
const (
	TriggerPlayerAny TriggerPlayerFilter = iota
	TriggerPlayerYou
	TriggerPlayerOpponent
)

// AttackRecipientKind identifies what an attacker was declared against.
type AttackRecipientKind uint8

// Attack recipient values are flags so exact player-or-permanent unions remain
// representable without interpreting Oracle wording at runtime.
const (
	AttackRecipientAny    AttackRecipientKind = 0
	AttackRecipientPlayer AttackRecipientKind = 1 << (iota - 1)
	AttackRecipientPlaneswalker
	AttackRecipientBattle
)

// TriggerPattern matches a Event for triggered-ability detection.
// Zero-valued filters are wildcards except Event, which must be set.
type TriggerPattern struct {
	Event EventKind

	Controller TriggerControllerFilter
	// CauseController constrains the controller of the spell or ability that
	// caused the event, independently from Controller's event-subject relation.
	CauseController TriggerControllerFilter
	Source          TriggerSourceFilter
	ExcludeSelf     bool
	Player          TriggerPlayerFilter

	Subject TriggerSubjectObject

	RequirePermanentTypes []types.Card
	ExcludePermanentTypes []types.Card
	RequireNonToken       bool

	// SubjectSelection is the Selection-based form of the event subject
	// permanent filters (RequirePermanentTypes/ExcludePermanentTypes and
	// RequireNonToken). It is wildcard by default; the rules matcher adapts the
	// legacy fields when it is empty, and the two forms must not both be set.
	SubjectSelection Selection

	// SubjectSelectionOrSelf widens a SubjectSelection-filtered event subject to
	// also match the ability's own source, expressing "this permanent or another
	// <Selection> you control" zone-change triggers (CR 603.2). When set, the
	// trigger fires if the event subject matches SubjectSelection or is the
	// source itself. It is only valid with a non-empty SubjectSelection,
	// Source == TriggerSourceAny, and ExcludeSelf == false.
	SubjectSelectionOrSelf bool

	// RelatedSubjectSelection matches a secondary combat permanent, such as the
	// attacker a creature blocks or the blocker that caused an attacker to
	// become blocked.
	RelatedSubjectSelection Selection

	// RequireCardTypes and ExcludeCardTypes filter spell-cast events by the
	// spell's types as chosen/cast on the stack (CR 601.2, CR 603.2).
	RequireCardTypes []types.Card
	ExcludeCardTypes []types.Card

	// CardSelection is the Selection-based form of the cast-spell card filters
	// (RequireCardTypes/ExcludeCardTypes). It is wildcard by default; the rules
	// matcher adapts the legacy fields when it is empty, and the two forms must
	// not both be set.
	CardSelection Selection

	// MatchSpellCopy widens an EventSpellCast pattern to also fire on
	// EventSpellCopied events ("Whenever you cast or copy ...", magecraft). It is
	// only valid with Event == EventSpellCast and never affects cast counts.
	MatchSpellCopy bool

	// SelfWasCast restricts an EventSpellCast pattern to the casting of the
	// ability's own source spell ("When you cast this spell", CR 601.3i). The
	// trigger fires once as the source spell is put on the stack, detected from
	// the cast spell's own card definition rather than from a battlefield
	// permanent. It is only valid with Event == EventSpellCast.
	SelfWasCast bool

	MatchFromZone bool
	FromZone      zone.Type
	MatchToZone   bool
	ToZone        zone.Type
	ExcludeToZone bool

	// ExcludeFromZone matches a zone change only when its origin is NOT FromZone,
	// expressing "put into <zone> from anywhere other than the battlefield".
	ExcludeFromZone bool

	MatchFaceDown bool
	FaceDown      bool

	MatchStackObjectKind bool
	StackObjectKind      StackObjectKind

	DamageRecipient            DamageRecipientKind
	DamageRecipientCombatState CombatStateFilter
	DamageRecipientTypes       []types.Card

	// DamageRecipientSelection is the Selection-based form of
	// DamageRecipientTypes. The two forms must not both be set.
	DamageRecipientSelection Selection
	// DamageRecipientIsSource restricts damage to the ability's source.
	DamageRecipientIsSource bool
	// DamageSourceSelection restricts the permanent that dealt damage.
	DamageSourceSelection Selection

	// DamageSourceCaptured restricts an EventDamageDealt pattern to combat damage
	// dealt by a specific permanent captured when the trigger was created, rather
	// than by a static filter. It is only meaningful on an event-based delayed
	// trigger whose DelayedTriggerDef carries a DamageSourceObject reference; the
	// delayed-trigger matcher resolves that reference at schedule time and
	// enforces the captured object identity. It must not be combined with a
	// Source filter, Subject, or DamageSourceSelection.
	DamageSourceCaptured bool

	// AttackRecipient restricts attacker-declared events by what was attacked.
	AttackRecipient AttackRecipientKind
	// AttackRecipientSelection restricts attacked planeswalkers and battles.
	AttackRecipientSelection Selection

	// AttackAlone restricts an EventAttackerDeclared trigger to a creature that
	// attacks alone, i.e. the only attacking creature in the combat ("attacks
	// alone", CR 508). It is only valid with Event == EventAttackerDeclared.
	AttackAlone bool
	// AttackerCountAtLeast restricts a controller-scoped EventAttackerDeclared
	// trigger to combats where at least this many creatures are attacking
	// ("attack with two or more creatures"). Zero imposes no minimum; a positive
	// value is only valid with Event == EventAttackerDeclared and OneOrMore.
	AttackerCountAtLeast int

	// RequireCombatDamage restricts damage triggers to combat damage events.
	RequireCombatDamage bool
	// RequireNonCombatDamage restricts damage triggers to noncombat damage events.
	RequireNonCombatDamage bool

	SpellTargetsSource bool
	SpellTargetAllow   TargetAllow
	// SpellTargetPattern restricts a spell-cast trigger to spells whose targets
	// match this permanent/card Selection ("a spell that targets a creature you
	// don't control"). It is matched through the canonical Selection matcher.
	SpellTargetPattern opt.V[Selection]
	RequireKickerPaid  bool
	RequireHistoric    bool
	// ExcludeManaAbility is required for EventAbilityActivated patterns until
	// payment-time mana activations join the authoritative event stream.
	ExcludeManaAbility bool
	// PlayerEventOrdinalThisTurn restricts a player event to its occurrence
	// number during the current turn. Zero does not restrict the event.
	PlayerEventOrdinalThisTurn int

	// ExcludeFirstDrawInDrawStep skips an EventCardDrawn that is the drawing
	// player's first draw during their own draw step ("except the first one they
	// draw in each of their draw steps", Orcish Bowmasters, Xyris). It is only
	// valid with Event == EventCardDrawn.
	ExcludeFirstDrawInDrawStep bool

	// OneOrMore coalesces matching events that happened as one batch into one
	// trigger. The first matching event is retained as TriggerEvent.
	OneOrMore bool
	// OneOrMorePerAttackTarget coalesces attack events separately for each
	// attacked player, planeswalker, or battle.
	OneOrMorePerAttackTarget bool

	// MatchCounterKind restricts EventCountersAdded triggers to a specific
	// counter type. When false, any counter kind satisfies the pattern.
	MatchCounterKind bool
	CounterKind      counter.Kind

	// Step filters EventBeginningOfStep triggers such as "At the beginning of
	// your upkeep" (CR 603.6c).
	Step Step
	// StepPlayerSourceAttachedSelection matches a step whose active player
	// controls the permanent the ability source is attached to.
	StepPlayerSourceAttachedSelection Selection

	// RequireTappedForMana restricts an EventPermanentTapped trigger to taps that
	// paid a mana ability's cost ("is tapped for mana"), CR 106.11a / 605.
	RequireTappedForMana bool

	// RequireProducedManaColor restricts a RequireTappedForMana trigger to taps
	// whose produced mana included this type ("tap a permanent for {C}" requires
	// colorless). It is empty for the unrestricted "for mana" wording, which
	// matches a tap producing any type. It is checked against the triggering
	// event's ProducedManaColors.
	RequireProducedManaColor mana.Color

	// UnionEvent joins a second event kind to Event under the pattern's shared
	// subject and player filters, expressing "Whenever you create or sacrifice a
	// token" (CR 603.2). When set, the trigger fires if the event kind equals
	// Event or UnionEvent. It is EventUnknown for single-event patterns.
	UnionEvent EventKind

	// AttackedPlayerHasMostLife restricts an EventAttackerDeclared trigger to
	// attacks against a player who has the most life among non-eliminated
	// players, or is tied for most ("attacks the player with the most life or
	// tied for most life", dethrone CR 702.103). It is only valid with
	// Event == EventAttackerDeclared.
	AttackedPlayerHasMostLife bool

	// AttacksWithGreaterPowerCreature restricts an EventAttackerDeclared trigger
	// to combats where another creature with power greater than the ability's
	// source is also attacking ("attacks with another creature with greater
	// power", training CR 702.150). It is only valid with
	// Event == EventAttackerDeclared and Source == TriggerSourceSelf.
	AttacksWithGreaterPowerCreature bool

	// AttackWhileSaddled restricts an EventAttackerDeclared trigger to combats
	// where the ability's source is saddled ("attacks while saddled", saddle
	// CR 702.166). It is only valid with Event == EventAttackerDeclared.
	AttackWhileSaddled bool

	// CastDuringTurn restricts an EventSpellCast trigger by whose turn the spell
	// was cast on, relative to the ability's controller ("Whenever you cast a
	// spell during your turn" / "during an opponent's turn"). TriggerTurnAny
	// imposes no restriction. It is only valid with Event == EventSpellCast.
	CastDuringTurn TriggerTurnRelation

	// ClassBecameLevel restricts an EventClassLevelGained trigger to the level
	// the Class became ("When this Class becomes level N"). Zero imposes no
	// restriction; a positive value is only valid with
	// Event == EventClassLevelGained and Source == TriggerSourceSelf.
	ClassBecameLevel int

	// DyingDamagedBySource restricts an EventPermanentDied trigger to a permanent
	// that was dealt damage by the ability's own source earlier this turn
	// ("Whenever a creature dealt damage by this creature this turn dies",
	// CR 603.2). The rules layer scans the current turn's damage events for one
	// whose source is this ability's source and whose damaged permanent is the
	// dying permanent. It is only valid with Event == EventPermanentDied.
	DyingDamagedBySource bool
}

// TriggerTurnRelation restricts a trigger by whose turn the triggering event
// occurred on, relative to the ability's controller.
type TriggerTurnRelation int

const (
	// TriggerTurnAny imposes no turn restriction.
	TriggerTurnAny TriggerTurnRelation = iota
	// TriggerTurnYours restricts to the controller's own turn ("during your
	// turn").
	TriggerTurnYours
	// TriggerTurnNotYours restricts to a turn that is not the controller's
	// ("during an opponent's turn").
	TriggerTurnNotYours
)

// TimingRestriction constrains when an activated ability can be used.
type TimingRestriction int

const (
	// NoTimingRestriction means the ability can be activated at instant speed.
	NoTimingRestriction TimingRestriction = iota

	// SorceryOnly means "activate only as a sorcery" (CR 113.6e).
	SorceryOnly

	// OncePerTurn means "activate only once each turn.".
	OncePerTurn

	// SorceryOncePerTurn combines both restrictions.
	SorceryOncePerTurn

	// DuringCombat means "activate only during combat.".
	DuringCombat

	// DuringUpkeep means "activate only during your upkeep.".
	DuringUpkeep

	// DuringYourTurn means "activate only during your turn." (CR 113.6).
	// The ability may be activated at any time the player has priority during
	// any phase or step of their own turn.
	DuringYourTurn

	// DuringYourTurnBeforeAttackers means "activate only during your turn, before
	// attackers are declared." (the Portal precombat cycle, e.g. Stern Marshal).
	// The ability may be activated only while its controller is the active player
	// and the turn has not yet reached the declare-attackers step: during the
	// beginning phase, the precombat main phase, or the beginning-of-combat step.
	DuringYourTurnBeforeAttackers
)

// EffectResultAmountKind identifies which numeric result an effect records for
// later linked "that much" or X instructions.
type EffectResultAmountKind int

// Effect result amount values select stored numeric results.
const (
	EffectResultAmountDefault EffectResultAmountKind = iota
	EffectResultAmountExcessDamage
)

// CounterSourceKind identifies where an effect reads counters from.
type CounterSourceKind int

// Counter source values identify where counter-moving effects read counters.
const (
	CounterSourceNone CounterSourceKind = iota

	// CounterSourceTarget reads counters from another chosen target.
	CounterSourceTarget

	// CounterSourceEventPermanent reads counters from the event permanent that
	// caused a triggered ability to trigger. If that permanent has left the
	// battlefield, mtg/rules reads its last-known information.
	CounterSourceEventPermanent

	// CounterSourceSelf reads counters from the source permanent of the effect
	// itself ("Move a +1/+1 counter from this creature onto target creature.").
	CounterSourceSelf
)

// CounterSourceSpec describes the source object for counter-moving effects.
type CounterSourceSpec struct {
	Kind   CounterSourceKind
	Object ObjectReference
}

// TokenCopySource identifies what object/card supplies copiable values for a
// token-copy effect.
type TokenCopySource int

// Token copy source values identify what supplies copiable values.
const (
	TokenCopySourceNone TokenCopySource = iota
	TokenCopySourceObject
	TokenCopySourceSourceCard
	// TokenCopySourceEachInGroup copies each member of Group in turn, creating
	// one token per matched permanent ("For each token you control, create a
	// token that's a copy of that permanent." — Second Harvest). It is the only
	// source that reads Group; the others read Object or the source card.
	TokenCopySourceEachInGroup
	// TokenCopySourceChosenFromTriggerBatch copies one permanent chosen by the
	// controller from the set of permanents that triggered the resolving ability
	// ("create a token that's a copy of one of them." on a "Whenever one or more
	// ... enter" trigger, Twilight Diviner). The candidate set is the resolving
	// ability's triggering event batch filtered by its own trigger pattern; it
	// reads neither Object nor Group.
	TokenCopySourceChosenFromTriggerBatch
)

// TokenCopySpec describes a token that starts as a copy of another object/card,
// then applies explicit copy-modifying exceptions such as Eternalize's color,
// type, power/toughness, and mana-cost overrides.
type TokenCopySpec struct {
	Source TokenCopySource
	Object ObjectReference

	SetName       string
	SetColors     []color.Color
	SetTypes      []types.Card
	SetSubtypes   []types.Sub
	SetPower      opt.V[PT]
	SetToughness  opt.V[PT]
	NoManaCost    bool
	NoPrintedText bool

	// SetNotLegendary drops the Legendary supertype from the copy ("except the
	// token isn't legendary"), so a copy of a legendary permanent does not force
	// the legend rule on its original.
	SetNotLegendary bool
	// AddColors, AddTypes, and AddSubtypes grant additional colors, card types,
	// and subtypes on top of the copied characteristics, modeling the
	// characteristic-overriding copy exception "except it's a <N/N> <color>
	// <subtype> <type> in addition to its other [colors and] types" (Saheeli's
	// Artistry, Urza, Prince of Kroog, Ratadrabik of Urborg). They append to the
	// copied values rather than replacing them, in contrast to the Set* fields.
	AddColors   []color.Color
	AddTypes    []types.Card
	AddSubtypes []types.Sub
	// AddKeywords grants additional keyword abilities to the created token on top
	// of the copied characteristics ("That token gains haste").
	AddKeywords []Keyword
	// Group is the controlled battlefield group copied member-by-member when
	// Source is TokenCopySourceEachInGroup; one token is created per matched
	// permanent, copying that permanent. It is nil for every other source. It is
	// held by pointer so the embedded GroupReference does not inflate the heavily
	// value-passed TokenCopySpec past the by-value size budget.
	Group *GroupReference
}

// EternalizeActivatedBody builds the ActivatedAbilityBody for the Eternalize
// keyword. Use this in CardFace.ActivatedAbilities with categorized fields.
func EternalizeActivatedBody(manaCost cost.Mana, creatureSubtypes ...types.Sub) ActivatedAbility {
	tokenSubtypes := make([]types.Sub, 0, len(creatureSubtypes)+1)
	tokenSubtypes = append(tokenSubtypes, types.Zombie)
	tokenSubtypes = append(tokenSubtypes, creatureSubtypes...)
	return ActivatedAbility{
		Text:           "Eternalize " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: zone.Graveyard,
		Timing:         SorceryOnly,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalExileSource,
			Text: "Exile this card from your graveyard",
		}},
		Content: Mode{Sequence: []Instruction{{
			Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenCopyOf(TokenCopySpec{
					Source:       TokenCopySourceSourceCard,
					SetColors:    []color.Color{color.Black},
					SetSubtypes:  tokenSubtypes,
					SetPower:     opt.Val(PT{Value: 4}),
					SetToughness: opt.Val(PT{Value: 4}),
					NoManaCost:   true,
				}),
			},
		}}}.Ability(),

		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Eternalize}},
	}
}

// EmbalmActivatedBody builds the ActivatedAbilityBody for the Embalm keyword.
// Like Eternalize, it exiles the card from its owner's graveyard at sorcery
// speed to create a token copy, except the token is white, gains the Zombie
// creature type, and has no mana cost. Unlike Eternalize it keeps the card's
// printed power and toughness. Use this in CardFace.ActivatedAbilities with
// categorized fields.
func EmbalmActivatedBody(manaCost cost.Mana, creatureSubtypes ...types.Sub) ActivatedAbility {
	tokenSubtypes := make([]types.Sub, 0, len(creatureSubtypes)+1)
	tokenSubtypes = append(tokenSubtypes, types.Zombie)
	tokenSubtypes = append(tokenSubtypes, creatureSubtypes...)
	return ActivatedAbility{
		Text:           "Embalm " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: zone.Graveyard,
		Timing:         SorceryOnly,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalExileSource,
			Text: "Exile this card from your graveyard",
		}},
		Content: Mode{Sequence: []Instruction{{
			Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenCopyOf(TokenCopySpec{
					Source:      TokenCopySourceSourceCard,
					SetColors:   []color.Color{color.White},
					SetSubtypes: tokenSubtypes,
					NoManaCost:  true,
				}),
			},
		}}}.Ability(),

		KeywordAbilities: []KeywordAbility{SimpleKeyword{Kind: Embalm}},
	}
}

// ScavengeActivatedAbility builds the canonical graveyard-activated ability for
// the Scavenge keyword (CR 702.94): exile this card from your graveyard at
// sorcery speed to put a number of +1/+1 counters equal to this card's power on
// target creature.
func ScavengeActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	return ActivatedAbility{
		Text:           "Scavenge " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: zone.Graveyard,
		Timing:         SorceryOnly,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalExileSource,
			Text: "Exile this card from your graveyard",
		}},
		Content: Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      TargetAllowPermanent,
				Selection:  opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []Instruction{{
				Primitive: AddCounter{
					Amount: Dynamic(DynamicAmount{
						Kind: DynamicAmountSourceCardPower,
					}),
					Object:      TargetPermanentReference(0),
					CounterKind: counter.PlusOnePlusOne,
				},
			}},
		}.Ability(),
		KeywordAbilities: []KeywordAbility{ScavengeKeyword{Cost: append(cost.Mana(nil), manaCost...)}},
	}
}

// UnearthActivatedAbility builds the canonical graveyard-activated ability for
// the Unearth keyword (CR 702.83): while this card is in its owner's graveyard,
// "{cost}: Return this card from your graveyard to the battlefield. It gains
// haste. Exile it at the beginning of the next end step or if it would leave the
// battlefield. Unearth only as a sorcery." The ability returns the source card
// to the battlefield under its controller, grants it haste while it remains
// there, and schedules a delayed end-step trigger that exiles it.
func UnearthActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	return ActivatedAbility{
		Text:           "Unearth " + manaCost.String(),
		ManaCost:       opt.Val(append(cost.Mana(nil), manaCost...)),
		ZoneOfFunction: zone.Graveyard,
		Timing:         SorceryOnly,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: PutOnBattlefield{
					Source:    CardBattlefieldSource(CardReference{Kind: CardReferenceSource}),
					Recipient: opt.Val(ControllerReference()),
					ContinuousEffects: []ContinuousEffect{{
						Layer:          LayerAbility,
						AffectedSource: true,
						AddKeywords:    []Keyword{Haste},
					}},
				},
			},
			{
				Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
					Timing: DelayedAtBeginningOfNextEndStep,
					Content: Mode{Sequence: []Instruction{{
						Primitive: Exile{Object: SourceCardPermanentReference()},
					}}}.Ability(),
				}},
			},
		}}.Ability(),
		KeywordAbilities: []KeywordAbility{UnearthKeyword{Cost: append(cost.Mana(nil), manaCost...)}},
	}
}

// SearchSpec describes a deterministic library-search slice. The rules
// implementation supports library -> hand and library -> battlefield templates
// with common type, supertype, and subtype filters.
type SearchSpec struct {
	SourceZone  zone.Type
	Destination zone.Type
	// DestinationPosition is required when Destination is an ordered zone. The
	// runtime currently supports only putting one found card on top of its
	// owner's library after shuffling.
	DestinationPosition SearchPosition
	// FailToFindPolicy controls whether the searching player may choose no card
	// when matching cards exist. Qualified hidden-zone searches may fail to find;
	// an unrestricted exact-card search must find one when the library is nonempty.
	FailToFindPolicy SearchFailToFindPolicy

	// Filter is the canonical predicate every matched library card must satisfy.
	// It carries the search's card-type, permanent-card, supertype, subtype,
	// color, mana-value, power, and toughness riders as a single game.Selection,
	// matched against a card in its library zone. An empty Filter matches every
	// library card (a plain "search your library for a card" tutor). The
	// dedicated MaxManaValueFromX and Name fields below carry the riders no
	// fixed Selection field can express.
	Filter Selection

	// MaxManaValueFromX, when true, restricts matches to cards whose mana value
	// is less than or equal to the spell's chosen {X}, modeling the "with mana
	// value X or less" rider on an X-cost library-search tutor (Green Sun's
	// Zenith, Chord of Calling, Wargate). The bound is resolved from the
	// resolving stack object's X as the search runs, so it is mutually exclusive
	// with a fixed Filter.ManaValue bound.
	MaxManaValueFromX bool

	Reveal       bool
	EntersTapped bool

	// RevealOnly, when true, makes the search find and reveal a single matching
	// card but leave it in the library with no destination move, publishing the
	// found card under PublishLinked so a following ConditionalDestinationPlace
	// can route it. It backs the search half of "Search your library for a Plains
	// card and reveal it. ... you may put that card onto the battlefield ..."
	// (Scholar of New Horizons), where the placement and the closing shuffle are
	// separate instructions. RevealOnly requires Destination zone.None, Reveal
	// true, a single searching player, and no split destination, tapped entry, or
	// controller rider.
	RevealOnly bool

	// SplitDestination, when present, makes the search distribute the found
	// cards across two distinct single-card destination slots instead of sending
	// every found card to Destination. The primary slot is (Destination,
	// EntersTapped); SplitDestination is the secondary slot. At most two cards
	// may be found. When two are found, the searching player assigns one card to
	// each slot; when only one is found, the searching player chooses which slot
	// it fills (CR 701.19; Cultivate, Kodama's Reach). It is meaningful only when
	// both slots are a Hand or Battlefield destination.
	SplitDestination opt.V[SearchDestination]

	// SharedSubtype, when true, requires every card found by a multi-card search
	// to share at least one subtype with each other found card, modeling the
	// "that share a land type" correlation rider on Myriad Landscape's "up to two
	// basic land cards" search. The search-choice machinery enforces the
	// correlation while the cards are chosen, preventing an illegal pair rather
	// than finding two cards and silently dropping one (CR 701.19). Finding zero
	// or one card satisfies it vacuously. It is meaningful only when more than
	// one card may be found and the matched cards carry subtypes.
	SharedSubtype bool

	// Name, when non-empty, restricts matches to cards whose name equals it,
	// modeling a "card named <Name>" library search (Daru Cavalier, Trustworthy
	// Scout, Embermage Goblin). It composes with the other filters but in
	// practice stands alone on a plain "card named X" tutor.
	Name string

	// SlotFilters, when non-empty, makes the search find one card matching each
	// per-slot filter in source order instead of applying the single Filter to
	// every found card, modeling a heterogeneous multi-slot search ("a Forest
	// card and a Plains card", Krosan Verge). Each found card enters the single
	// shared Destination (with EntersTapped); the searching player makes one
	// optional dependent choice per slot, finding at most one card per slot and
	// never assigning a card to two slots. It is meaningful only when Filter is
	// empty, the destination is a single non-split Hand or Battlefield slot, and
	// Amount equals len(SlotFilters); it is mutually exclusive with
	// SplitDestination, SharedSubtype, RevealOnly, MaxManaValueFromX, and Name.
	SlotFilters []Selection

	// AlsoGraveyard, when true, extends the library search to also consider the
	// searching player's graveyard, modeling the "search your library and/or
	// graveyard for a card named X, reveal it, and put it into your hand. If you
	// search your library this way, shuffle." planeswalker-companion tutor
	// (Ashiok's Forerunner, Niambi, Faithful Healer). The searching player finds
	// a single matching card in either zone, reveals it, and puts it into their
	// hand; the library is shuffled afterward because the library is always among
	// the searched zones. It is meaningful only with SourceZone Library,
	// Destination Hand, Reveal true, a single searching player, and no split
	// destination, slot filters, RevealOnly, MaxManaValueFromX, shared subtype,
	// or tapped entry; the search may fail to find a card (CR 701.19e).
	AlsoGraveyard bool
}

// IsUnrestricted reports whether every library card matches the search filter.
func (s SearchSpec) IsUnrestricted() bool {
	return s.Filter.Empty() &&
		!s.MaxManaValueFromX &&
		!s.SharedSubtype &&
		!s.AlsoGraveyard &&
		len(s.SlotFilters) == 0 &&
		s.Name == ""
}

// SearchDestination is one single-card destination slot of a split-destination
// library search, naming the zone a found card enters and whether it enters the
// battlefield tapped.
type SearchDestination struct {
	Zone         zone.Type
	Position     SearchPosition
	EntersTapped bool
}

// SearchPosition identifies an ordered position within a search destination.
type SearchPosition uint8

// Supported ordered search destination positions.
const (
	SearchPositionUnspecified SearchPosition = iota
	SearchPositionTop
)

// SearchFailToFindPolicy controls whether a library search may return no card.
type SearchFailToFindPolicy uint8

// Supported library-search fail-to-find policies.
const (
	// SearchFailToFindDefault derives the rule from the typed search shape:
	// qualified or "up to" searches may fail, while a singular unrestricted
	// search must find a card when the library is nonempty.
	SearchFailToFindDefault SearchFailToFindPolicy = iota
	SearchMayFailToFind
	SearchMustFindIfAvailable
)

// EffectCondition describes a simple condition that must be true when an
// effect resolves. It is data only; mtg/rules owns evaluation.
type EffectCondition struct {
	// Text preserves the printed condition for logs, diagnostics, and review.
	Text string

	// Object identifies the object whose current characteristics are tested.
	Object ObjectReference

	PermanentType opt.V[types.Card]

	// Negate inverts the permanent-type match, e.g. "it isn't a creature".
	Negate bool

	// Condition is an additional shared condition evaluated with the resolving
	// stack object bound.
	Condition opt.V[Condition]
}

// TargetSpec describes the targeting requirements of an ability.
type TargetSpec struct {
	// MinTargets is the minimum number of targets (0 for "up to").
	MinTargets int

	// MaxTargets is the maximum number of targets.
	MaxTargets int

	// Constraint describes what can be targeted (e.g., "creature",
	// "creature or planeswalker", "player").
	Constraint string

	// Allow describes the broad categories this target may choose from.
	// Constraint remains for display and as a legacy fallback.
	Allow TargetAllow

	// Predicate carries the stack-object and spell-only qualifiers for
	// stack-object targets (kinds, controller, mana value, spell card
	// types/colors/supertypes, source types). Permanent, card, and player
	// characteristic predicates live on Selection. A combined "spell or
	// permanent" target sets both: Predicate gates its stack-object alternative
	// and Selection gates its permanent alternative.
	Predicate TargetPredicate

	// Selection is the canonical permanent/card/player characteristic predicate
	// for this target. It is the sole characteristic matcher; the runtime
	// permanent, card, and player legality tests read it directly.
	Selection opt.V[Selection]

	// TargetZone restricts card targets to one zone. It is meaningful only when
	// Allow includes TargetAllowCard.
	TargetZone zone.Type

	// Chooser identifies who chooses this target slot during announcement. The
	// default controller chooser preserves normal targeting. For non-controller
	// choosers, structured "you" predicates are evaluated relative to the
	// choosing player.
	Chooser TargetChooser

	// DistinctFromPriorTargets requires every object chosen for this spec to
	// differ from every object already chosen for the preceding target specs of
	// the same spell or ability ("... fights another target creature"). It is
	// meaningful only for specs after the first; the default false preserves the
	// ordinary rule that distinct target specs may otherwise overlap.
	DistinctFromPriorTargets bool

	// CountEqualsX requires the number of targets chosen for this spec to equal
	// the spell's chosen X ("Exile X target creatures"). The spec carries a 0..N
	// range so announcement enumerates every count, and casting is legal only at
	// the combination whose size matches X, binding the variable cost to the
	// number of targets. The default false leaves the count governed solely by
	// MinTargets/MaxTargets.
	CountEqualsX bool

	// SameGraveyard requires every card chosen for this spec to lie in one and
	// the same graveyard ("Exile up to three target cards from a single
	// graveyard"). A card in a graveyard is always in its owner's graveyard
	// (CR 404.2), so the constraint is satisfied exactly when every chosen card
	// target shares one owner. It is meaningful only for card targets and the
	// default false leaves the targets independently chosen.
	SameGraveyard bool
}

// TargetChooser identifies who chooses a target slot during announcement.
type TargetChooser int

// Target chooser values identify who chooses a target slot.
const (
	TargetChooserController TargetChooser = iota
	// TargetChooserOpponent means the ability controller chooses an opponent,
	// then that opponent chooses this target slot.
	TargetChooserOpponent
)

// Mode represents one mode of a modal spell or ability ("Choose one —",
// "Choose two —", etc.; CR 700.2).
type Mode struct {
	// Text is the oracle text of this mode.
	Text string

	// Targets are the targeting requirements of this mode.
	Targets []TargetSpec

	// Sequence is the typed instruction sequence this mode produces.
	Sequence []Instruction

	// Cost is the additional mana cost paid to choose this mode when casting a
	// Spree spell (CR 702.171). It is set only on Spree spell modes; choosing a
	// mode adds its Cost to the spell's total cost. An empty value means the mode
	// has no additional cost.
	Cost opt.V[cost.Mana]
}

// Ability creates ordinary non-modal ability content from this mode.
func (m Mode) Ability() AbilityContent {
	return AbilityContent{
		Modes:               []Mode{m},
		MinModes:            1,
		MaxModes:            1,
		AllowDuplicateModes: false,
	}
}

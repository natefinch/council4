// Package authored owns non-generated Card Implementations.
// Legacy corpus generation only writes letter packages, never this directory.
package authored

import "github.com/natefinch/council4/mtg/cards/cardset"

// Cards is extended by completed, behavior-verified authored card slices.
// The program infrastructure itself does not promote a synthetic parity fixture.
var Cards = verifiedEntries()

func verifiedEntries() []cardset.Entry {
	entries := []cardset.Entry{Terastodon, HiveMind, GiftsUngiven, GildedDrake, WorldgorgerDragon, NimDeathmantle, Abundance, BloodMoon, Muldrotha, Skullbriar, Mindslaver, GenerousGift, CoastalDiscovery, BoilingEarth, RoilSpout, RushOfIce, RisingMiasma, Pongify, RapidHybridization, Resculpt, Twincast, Reverberate, DualcasterMage, CacklingCounterpart, FatedInfatuation, Intuition, RealmsUncharted, ChainsOfMephistopheles, Humility, Opalescence, ForcedMarch, IncineratingBlast, Aethertow, Gigadrowse, LeapOfFlame, LoseFocus, ShatteringSpree, Pyromatics, SuperCombo, TrainOfThought, Vacuumelt, OneDozenEyes, BarbedLightning}
	entries = append(entries, RosethornAcolyte, CrystalDragon, PollenShieldHare,
		MurderousRider, GumdropPoisoner, QuestingDruid, QuirionDryad)
	entries = append(entries, Bedevil, Terminate, Vindicate, Disenchant, Naturalize, GleefulSabotage, BurnTrail, MemorySluice, GhastlyDiscovery, BarkshellBlessing)
	entries = append(entries, RallyTheGaladhrim)
	entries = append(entries, WildRide, MammothBellow)
	entries = append(entries, AnOfferYouCantRefuse, Counterspell, Negate, EssenceScatter, Disallow)
	entries = append(entries, OverwhelmingDenial)
	entries = append(entries, EladamrisCall, SteelshapersGift, Fabricate, IdyllicTutor, SylvanScrying, DemonicTutor, DiabolicTutor)
	entries = append(entries, DeathDenied, WildestDreams, RatsFeast, ShatteredCrypt, ReturnToTheRanks, ChampionOfStraySouls)
	entries = append(entries, RestlessDreams, TurbulentDreams, NostalgicDreams)
	entries = append(entries, WarpWorld, PanglacialWurm)
	entries = append(entries, GhostVacuum, NecromanticSelection, ToothCollector, MarchesaDealerOfDeath)
	entries = append(entries, MyrCustodian)
	entries = append(entries, MoggSalvage, RefreshingRain, Submerge, Massacre)
	entries = append(entries, DeepwoodLegate, KyrenLegate, ChoArrimLegate, SaprazzanLegate, RushwoodLegate)
	entries = append(entries, Reiterate, Allay, Capsize)
	entries = append(entries, JwarIsleAvenger, BoulderSalvo, FallOfTheTitans, ComparativeAnalysis,
		ContainmentMembrane)
	entries = append(entries, SearingTouch, SeethingAnger, MindPeel, ElvishFury)
	entries = append(entries, RecruitTheWorthy, MindGames, HazeOfRage)
	entries = append(entries, LabRats, MysticSpeculation, SproutSwarm)
	entries = append(entries, ShatteringPulse, WhispersOfTheMuse, DisturbedBurial)
	entries = append(entries, Mundungu, SlyInstigator, HalfElfMonk, MinamoSightbender)
	entries = append(entries, EagerConstruct, SeaGodsRevenge, DevoutDecree)
	entries = append(entries, BuildersTalent, LeadersTalent, PaladinClass, GossipsTalent,
		FortuneTellersTalent, RangerClass, WarlockClass, ArtistsTalent)
	entries = append(entries, Hammerheim, CrashLanding, KinscaerHarpoonist)
	entries = append(entries, BurningPalmEfreet, WindShear, LeeringGargoyle)
	entries = append(entries, FearOfFalling, ShortCircuit, CanopyDragon)
	entries = append(entries, AgonyWarp, HeatedDebate, SliceFromTheShadows)
	entries = append(entries, MizziumSkin, BountyOfMight, LongGoodbye)
	entries = append(entries, FearAura, LifelinkAura, VigilanceAura)
	entries = append(entries, PartyDude, SengirAutocrat, WhiteGloveGourmand)
	entries = append(entries, TouchOfTheVoid, RebelSalvo, KonstrariImproviserSoulTether,
		MilitantInquisitor, ManifestDread, DevilsPlay, OxiddaFinisher)
	entries = append(entries, BurrentonBombardier, BreakTies, FowlStrike, SwellOfCourage, HuntingTriad,
		MosquitoGuard, Earthbrawn, WrensRunHydra, RusticClachan, BannerhideKrushok)
	entries = append(entries, DefilerOfSouls, WoebringerDemon, WasitoraNekoruQueen)
	entries = append(entries, MirranBardiche, SylvokBattleChair, GoldwardensHelm, DragonwingGlider,
		VulshokSplitter, BarbedBatterfist, KembasBanner)
	entries = append(entries, AncestralVision, SearchForTomorrow, VisceridDeepwalker,
		KeldonHalberdier, CorpulentCorpse, ErrantEphemeron, MoxTantalite,
		InfiltratorIlKor, IthHighArcanist, RiftSower)
	entries = append(entries, CommonBond, SeedsOfStrength, MartialGlory)
	entries = append(entries, IgneousInspiration, EnthusiasticStudy, ArcaneSubtraction,
		GuidingVoice, CramSession, StudyBreak, PopQuiz)
	entries = append(entries, ExquisiteHuntmaster, SoulOfEternity, AngelOfIndemnity, BriarbladeAdept,
		ImpulsivePilferer, KinsbaileCourier, BroodmateTyrant, FinCladeFugitives, TroveTracker)
	entries = append(entries, SaibaCryptomancer, ChompingKavu, ConsumingAetherborn,
		GloomfangMauler, CragsmasherYeti, SerpentBladeAssailant, SigiledSentinel,
		DeathGreetersChampion, ScornBladeBerserker, VoldarenThrillseeker,
		EnduringBondwarden, HangarScrounger, RedcapHeelslasher,
		GoldenScaleAeronaut, FearlessSkald)
	entries = append(entries, BenalishFaithbonder, BalduvianBerserker, CoalitionWarbrute,
		YavimayaSteelcrusher, LinebreakerBaloth, ArgivianCavalier, BarkweaveCrusher,
		CoalitionSkyknight, HexbaneTortoise)
	entries = append(entries, GoringWarplow, CombatThresher, BlitzAutomaton, CradleClearcutter,
		RustGoliath, FrogmyrEnforcer, FallajiDragonEngine, DepthChargeColossus)
	entries = append(entries, SqueakingPieSneak, SadisticSkymarcher, FlamekinBladewhirl,
		SilvergillAdept, WrensRunVanquisher)
	entries = append(entries, Breezekeeper, NightClubber, RiveteersDecoy, RiveteersRequisitioner)
	entries = append(entries, PlasmaJockey)
	entries = append(entries, CaldaiaStrongarm)
	entries = append(entries, MayhemPatrol)
	entries = append(entries, Chaoslace)
	entries = append(entries, Purelace, Moonlace, Thoughtlace, Lifelace, Deathlace)
	entries = append(entries, MerfolkRaiders, TeferisDrake, TolarianDrake, SandbarCrocodile, TeferisIsle)
	entries = append(entries, GoblinGrappler, CrestedCraghorn, Brontotherium,
		DeftbladeElite, KrosanVorine, LowlandTracker)
	entries = append(entries, SkyshroudRidgeback, Cloudskate, SkyshroudBehemoth,
		RejuvenationChamber, Blastoderm)
	entries = append(entries, RainOfRust, UnboundedPotential)
	entries = append(entries, Controvert, GrimHarvest, Icefall, KrovikanRot, Resize, SunsBounty)
	entries = append(entries, RuinationGuide, BenthicInfiltrator, CullingDrone, SludgeCrawler, MistIntruder)
	for i := range entries {
		entries[i].ReplaceLegacy = replacesCompiledEntry(entries[i].Name)
	}
	return entries
}

func replacesCompiledEntry(name string) bool {
	// These exact authored identities intentionally supersede their newly
	// installed compiler definitions; the generated sources remain available.
	switch name {
	case "An Offer You Can't Refuse", "Bedevil", "Cackling Counterpart",
		"Champion of Stray Souls", "Counterspell", "Death Denied", "Demonic Tutor", "Diabolic Tutor", "Disallow", "Disenchant", "Dualcaster Mage", "Eladamri's Call",
		"Essence Scatter", "Fabricate", "Fated Infatuation", "Generous Gift",
		"Idyllic Tutor", "Naturalize", "Negate", "Pongify", "Rapid Hybridization",
		"Rats' Feast", "Resculpt", "Return to the Ranks", "Reverberate", "Shattered Crypt", "Steelshaper's Gift", "Sylvan Scrying", "Terminate",
		"Twincast", "Vindicate", "Wildest Dreams":
		return true
	default:
		return false
	}
}

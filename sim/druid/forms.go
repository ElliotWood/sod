package druid

import (
	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
	"github.com/wowsims/sod/sim/core/stats"
)

type DruidForm uint8

const (
	Humanoid DruidForm = 1 << iota
	Bear
	Cat
	Moonkin
	Tree
	Any = Humanoid | Bear | Cat | Moonkin | Tree
)

func (form DruidForm) Matches(other DruidForm) bool {
	return (form & other) != 0
}

func (druid *Druid) GetForm() DruidForm {
	return druid.form
}

func (druid *Druid) InForm(form DruidForm) bool {
	return druid.form.Matches(form)
}

func (druid *Druid) ClearForm(sim *core.Simulation) {
	if druid.InForm(Cat) {
		druid.CatFormAura.Deactivate(sim)
	} else if druid.InForm(Bear) {
		druid.BearFormAura.Deactivate(sim)
	} else if druid.InForm(Moonkin) {
		panic("cant clear moonkin form")
	}
	druid.form = Humanoid
	druid.SetCurrentPowerBar(core.ManaBar)
}

// TODO: don't hardcode numbers
func (druid *Druid) GetCatWeapon() core.Weapon {
	return core.Weapon{
		BaseDamageMin:        16.3866,
		BaseDamageMax:        24.5799,
		SwingSpeed:           1.0,
		NormalizedSwingSpeed: 1.0,
		CritMultiplier:       druid.MeleeCritMultiplier(1, 0),
		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
	}
}

// Func (druid *Druid) GetBearWeapon() core.Weapon {
// 	return core.Weapon{
// 		BaseDamageMin:        109,
// 		BaseDamageMax:        165,
// 		SwingSpeed:           2.5,
// 		NormalizedSwingSpeed: 2.5,
// 		CritMultiplier:       druid.MeleeCritMultiplier(Bear),
// 		AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
// 	}
// }

// TODO: Class bonus stats for both cat and bear.
func (druid *Druid) GetFormShiftStats() stats.Stats {
	s := stats.Stats{
		stats.AttackPower: float64(druid.Talents.PredatoryStrikes) * 0.5 * float64(druid.Level),
		stats.MeleeCrit:   float64(druid.Talents.SharpenedClaws) * 2 * core.CritRatingPerCritChance,
	}
	/*
		if weapon := druid.GetMHWeapon(); weapon != nil {
			dps := (weapon.WeaponDamageMax+weapon.WeaponDamageMin)/2.0/weapon.SwingSpeed + druid.PseudoStats.BonusMHDps
			weapAp := weapon.Stats[stats.AttackPower] + weapon.Enchant.Stats[stats.AttackPower]
			fap := math.Floor((dps - 54.8) * 14)

			s[stats.AttackPower] += fap
			s[stats.AttackPower] += (fap + weapAp) * ((0.2 / 3) * float64(druid.Talents.PredatoryStrikes))
		}
	*/

	return s
}

func (druid *Druid) GetDynamicPredStrikeStats() stats.Stats {
	// Accounts for ap bonus for 'dynamic' enchants
	// just scourgebane currently, this is a bit hacky but is needed as the bonus varies based on current target
	// so has to be 'cached' differently
	s := stats.Stats{}
	if weapon := druid.GetMHWeapon(); weapon != nil {
		bonusAp := 0.0
		if weapon.Enchant.EffectID == 3247 && druid.CurrentTarget.MobType == proto.MobType_MobTypeUndead {
			bonusAp += 140
		}
		s[stats.AttackPower] += bonusAp * ((0.2 / 3) * float64(druid.Talents.PredatoryStrikes))
	}
	return s
}

// TODO: Classic feral and bear
func (druid *Druid) registerCatFormSpell() {
	actionID := core.ActionID{SpellID: 768}

	srm := druid.getSavageRoarMultiplier()

	statBonus := druid.GetFormShiftStats().Add(stats.Stats{
		stats.AttackPower: float64(druid.Level) * 2,
	})

	agiApDep := druid.NewDynamicStatDependency(stats.Agility, stats.AttackPower, 1)
	feralApDep := druid.NewDynamicStatDependency(stats.FeralAttackPower, stats.AttackPower, 1)

	var hotwDep *stats.StatDependency
	if druid.Talents.HeartOfTheWild > 0 {
		hotwDep = druid.NewDynamicMultiplyStat(stats.AttackPower, 1.0+0.02*float64(druid.Talents.HeartOfTheWild))
	}

	clawWeapon := druid.GetCatWeapon()

	predBonus := stats.Stats{}

	druid.CatFormAura = druid.RegisterAura(core.Aura{
		Label:      "Cat Form",
		ActionID:   actionID,
		Duration:   core.NeverExpires,
		BuildPhase: core.Ternary(druid.StartingForm.Matches(Cat), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if !druid.Env.MeasuringStats && druid.form != Humanoid {
				druid.ClearForm(sim)
			}
			druid.form = Cat
			druid.SetCurrentPowerBar(core.EnergyBar)

			druid.AutoAttacks.SetMH(clawWeapon)

			druid.PseudoStats.ThreatMultiplier *= 0.71
			druid.PseudoStats.BaseDodge += 0.02 * float64(druid.Talents.FelineSwiftness)
			druid.PseudoStats.Shapeshifted = true

			predBonus = druid.GetDynamicPredStrikeStats()
			druid.AddStatsDynamic(sim, predBonus)
			druid.AddStatsDynamic(sim, statBonus)
			druid.EnableDynamicStatDep(sim, agiApDep)
			druid.EnableDynamicStatDep(sim, feralApDep)
			if hotwDep != nil {
				druid.EnableDynamicStatDep(sim, hotwDep)
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetReplaceMHSwing(nil)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()

				// These buffs stay up, but corresponding changes don't
				if druid.SavageRoarAura.IsActive() {
					druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= srm
				}

				if druid.PredatoryInstinctsAura != nil {
					druid.PredatoryInstinctsAura.Activate(sim)
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			druid.form = Humanoid

			druid.AutoAttacks.SetMH(druid.WeaponFromMainHand(druid.MeleeCritMultiplier(1, 0)))

			druid.PseudoStats.ThreatMultiplier /= 0.71
			druid.PseudoStats.BaseDodge -= 0.02 * float64(druid.Talents.FelineSwiftness)
			druid.PseudoStats.Shapeshifted = false

			druid.AddStatsDynamic(sim, predBonus.Invert())
			druid.AddStatsDynamic(sim, statBonus.Invert())
			druid.DisableDynamicStatDep(sim, agiApDep)
			druid.DisableDynamicStatDep(sim, feralApDep)
			if hotwDep != nil {
				druid.DisableDynamicStatDep(sim, hotwDep)
			}

			if !druid.Env.MeasuringStats {
				druid.AutoAttacks.SetReplaceMHSwing(nil)
				druid.AutoAttacks.EnableAutoSwing(sim)
				druid.manageCooldownsEnabled()
				druid.UpdateManaRegenRates()

				//druid.TigersFuryAura.Deactivate(sim)

				// These buffs stay up, but corresponding changes don't
				if druid.SavageRoarAura.IsActive() {
					druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= srm
				}

				if druid.PredatoryInstinctsAura != nil {
					druid.PredatoryInstinctsAura.Deactivate(sim)
				}
			}
		},
	})

	energyMetrics := druid.NewEnergyMetrics(actionID)

	druid.CatForm = druid.RegisterSpell(Any, core.SpellConfig{
		ActionID: actionID,
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

		ManaCost: core.ManaCostOptions{
			BaseCost:   0.55,
			Multiplier: 1.0 - 0.1*float64(druid.Talents.NaturalShapeshifter),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			maxShiftEnergy := core.TernaryFloat64(sim.RandomFloat("Furor") < 0.2*float64(druid.Talents.Furor), 40, 0)
			energyDelta := maxShiftEnergy - druid.CurrentEnergy()

			if energyDelta > 0 {
				druid.AddEnergy(sim, energyDelta, energyMetrics)
			} else {
				druid.SpendEnergy(sim, -energyDelta, energyMetrics)
			}

			druid.CatFormAura.Activate(sim)
		},
	})
}

// func (druid *Druid) registerBearFormSpell() {
// 	actionID := core.ActionID{SpellID: 9634}
// 	healthMetrics := druid.NewHealthMetrics(actionID)

// 	statBonus := druid.GetFormShiftStats().Add(stats.Stats{
// 		stats.AttackPower: 3 * float64(druid.Level),
// 	})

// 	stamDep := druid.NewDynamicMultiplyStat(stats.Stamina, 1.25)

// 	var potpDep *stats.StatDependency
// 	if druid.Talents.ProtectorOfThePack > 0 {
// 		potpDep = druid.NewDynamicMultiplyStat(stats.AttackPower, 1.0+0.02*float64(druid.Talents.ProtectorOfThePack))
// 	}

// 	var hotwDep *stats.StatDependency
// 	if druid.Talents.HeartOfTheWild > 0 {
// 		hotwDep = druid.NewDynamicMultiplyStat(stats.Stamina, 1.0+0.02*float64(druid.Talents.HeartOfTheWild))
// 	}

// 	potpdtm := 1 - 0.04*float64(druid.Talents.ProtectorOfThePack)

// 	clawWeapon := druid.GetBearWeapon()
// 	predBonus := stats.Stats{}

// 	druid.BearFormAura = druid.RegisterAura(core.Aura{
// 		Label:      "Bear Form",
// 		ActionID:   actionID,
// 		Duration:   core.NeverExpires,
// 		BuildPhase: core.Ternary(druid.StartingForm.Matches(Bear), core.CharacterBuildPhaseBase, core.CharacterBuildPhaseNone),
// 		OnGain: func(aura *core.Aura, sim *core.Simulation) {
// 			if !druid.Env.MeasuringStats && druid.form != Humanoid {
// 				druid.ClearForm(sim)
// 			}
// 			druid.form = Bear
// 			druid.SetCurrentPowerBar(core.RageBar)

// 			druid.AutoAttacks.SetMH(clawWeapon)

// 			druid.PseudoStats.ThreatMultiplier *= 2.1021
// 			druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.0 + 0.02*float64(druid.Talents.MasterShapeshifter)
// 			druid.PseudoStats.DamageTakenMultiplier *= potpdtm
// 			druid.PseudoStats.BaseDodge += 0.02 * float64(druid.Talents.FeralSwiftness+druid.Talents.NaturalReaction)

// 			predBonus = druid.GetDynamicPredStrikeStats()
// 			druid.AddStatsDynamic(sim, predBonus)
// 			druid.AddStatsDynamic(sim, statBonus)
// 			druid.ApplyDynamicEquipScaling(sim, stats.Armor, druid.BearArmorMultiplier())
// 			if potpDep != nil {
// 				druid.EnableDynamicStatDep(sim, potpDep)
// 			}

// 			// Preserve fraction of max health when shifting
// 			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
// 			druid.EnableDynamicStatDep(sim, stamDep)
// 			if hotwDep != nil {
// 				druid.EnableDynamicStatDep(sim, hotwDep)
// 			}
// 			druid.GainHealth(sim, healthFrac*druid.MaxHealth()-druid.CurrentHealth(), healthMetrics)

// 			if !druid.Env.MeasuringStats {
// 				druid.AutoAttacks.SetReplaceMHSwing(druid.ReplaceBearMHFunc)
// 				druid.AutoAttacks.EnableAutoSwing(sim)

// 				druid.manageCooldownsEnabled()
// 				druid.UpdateManaRegenRates()
// 			}
// 		},
// 		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
// 			druid.form = Humanoid
// 			druid.AutoAttacks.SetMH(druid.WeaponFromMainHand(druid.MeleeCritMultiplier(Humanoid)))

// 			druid.PseudoStats.ThreatMultiplier /= 2.1021
// 			druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= 1.0 + 0.02*float64(druid.Talents.MasterShapeshifter)
// 			druid.PseudoStats.DamageTakenMultiplier /= potpdtm
// 			druid.PseudoStats.BaseDodge -= 0.02 * float64(druid.Talents.FeralSwiftness+druid.Talents.NaturalReaction)

// 			druid.AddStatsDynamic(sim, predBonus.Invert())
// 			druid.AddStatsDynamic(sim, statBonus.Invert())
// 			druid.RemoveDynamicEquipScaling(sim, stats.Armor, druid.BearArmorMultiplier())
// 			if potpDep != nil {
// 				druid.DisableDynamicStatDep(sim, potpDep)
// 			}

// 			healthFrac := druid.CurrentHealth() / druid.MaxHealth()
// 			druid.DisableDynamicStatDep(sim, stamDep)
// 			if hotwDep != nil {
// 				druid.DisableDynamicStatDep(sim, hotwDep)
// 			}
// 			druid.RemoveHealth(sim, druid.CurrentHealth()-healthFrac*druid.MaxHealth())

// 			if !druid.Env.MeasuringStats {
// 				druid.AutoAttacks.SetReplaceMHSwing(nil)
// 				druid.AutoAttacks.EnableAutoSwing(sim)

// 				druid.manageCooldownsEnabled()
// 				druid.UpdateManaRegenRates()
// 				druid.EnrageAura.Deactivate(sim)
// 				druid.MaulQueueAura.Deactivate(sim)
// 			}
// 		},
// 	})

// 	rageMetrics := druid.NewRageMetrics(actionID)

// 	furorProcChance := []float64{0, 0.2, 0.4, 0.6, 0.8, 1}[druid.Talents.Furor]

// 	druid.BearForm = druid.RegisterSpell(Any, core.SpellConfig{
// 		ActionID: actionID,
// 		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,

// 		ManaCost: core.ManaCostOptions{
// 			BaseCost:   0.55,
// 			Multiplier: (1 - 0.2*float64(druid.Talents.KingOfTheJungle)) * (1 - 0.1*float64(druid.Talents.NaturalShapeshifter)),
// 		},
// 		Cast: core.CastConfig{
// 			DefaultCast: core.Cast{
// 				GCD: core.GCDDefault,
// 			},
// 			IgnoreHaste: true,
// 		},

// 		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
// 			rageDelta := 0 - druid.CurrentRage()
// 			if sim.Proc(furorProcChance, "Furor") {
// 				rageDelta += 10
// 			}
// 			if rageDelta > 0 {
// 				druid.AddRage(sim, rageDelta, rageMetrics)
// 			} else if rageDelta < 0 {
// 				druid.SpendRage(sim, -rageDelta, rageMetrics)
// 			}
// 			druid.BearFormAura.Activate(sim)
// 		},
// 	})
// }

func (druid *Druid) manageCooldownsEnabled() {
	// Disable cooldowns not usable in form and/or delay others
	if druid.StartingForm.Matches(Cat | Bear) {
		for _, mcd := range druid.disabledMCDs {
			mcd.Enable()
		}
		druid.disabledMCDs = nil

		if druid.InForm(Humanoid) {
			// Disable cooldown that incurs a gcd, so we dont get stuck out of form when we dont need to (Greater Drums)
			for _, mcd := range druid.GetMajorCooldowns() {
				if mcd.Spell.DefaultCast.GCD > 0 {
					mcd.Disable()
					druid.disabledMCDs = append(druid.disabledMCDs, mcd)
				}
			}
		}
	}
}

func (druid *Druid) applyMoonkinForm() {
	if !druid.InForm(Moonkin) || !druid.Talents.MoonkinForm {
		return
	}

	druid.RegisterAura(core.Aura{
		Label:    "Moonkin Form",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}

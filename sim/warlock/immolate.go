package warlock

import (
	"strconv"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (warlock *Warlock) registerImmolateSpell() {
	actionID := core.ActionID{SpellID: 47811}
	baseCost := 0.17 * warlock.BaseMana
	costReduction := 0.0
	if float64(warlock.Talents.Cataclysm) > 0 {
		costReduction += 0.01 + 0.03 * float64(warlock.Talents.Cataclysm)
	}

	effect := core.SpellEffect{
		BonusSpellCritRating: core.TernaryFloat64(warlock.Talents.Devastation, 0, 1) * 5 * core.CritRatingPerCritChance,
		DamageMultiplier: (1 + (0.1 * float64(warlock.Talents.ImprovedImmolate))) * (1 + 0.03 * float64(warlock.Talents.Emberstorm)),
		ThreatMultiplier: 1 - 0.1*float64(warlock.Talents.DestructiveReach),
		BaseDamage:       core.BaseDamageConfigMagic(460.0, 460.0, 0.2+0.04*float64(warlock.Talents.ShadowAndFlame)),
		OutcomeApplier:   warlock.OutcomeFuncMagicHitAndCrit(warlock.SpellCritMultiplier(1, float64(warlock.Talents.Ruin)/5)),
		OnSpellHitDealt:  applyDotOnLanded(&warlock.ImmolateDot),
		ProcMask:         core.ProcMaskSpellDamage,
	}

	warlock.Immolate = warlock.RegisterSpell(core.SpellConfig{
		ActionID:     actionID,
		SpellSchool:  core.SpellSchoolFire,
		ResourceType: stats.Mana,
		BaseCost:     baseCost,
		
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost:     baseCost * (1 - costReduction),
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond * (2000 - 100 * time.Duration(warlock.Talents.Bane) - 50 * time.Duration(warlock.Talents.Emberstorm)),
			},
		},
		ApplyEffects: core.ApplyEffectFuncDirectDamage(effect),
	})

	target := warlock.CurrentTarget

	// DOT: 615 dmg over 15s (123 every 3 sec, mod 0.13)
	warlock.ImmolateDot = core.NewDot(core.Dot{
		Spell: warlock.Immolate,
		Aura: target.RegisterAura(core.Aura{
			Label:    "immolate-" + strconv.Itoa(int(warlock.Index)),
			ActionID: actionID,
		}),
		NumberOfTicks: 5 + core.TernaryInt(warlock.HasSetBonus(ItemSetVoidheartRaiment, 4), 1, 0), // voidheart 4p gives 1 extra tick
		TickLength:    time.Second * 3,
		TickEffects: core.TickFuncSnapshot(target, core.SpellEffect{
			DamageMultiplier: (1 + 0.03 * float64(warlock.Talents.Aftermath)) * (1 + 0.03 * float64(warlock.Talents.Emberstorm)),
			ThreatMultiplier: 1 - 0.1*float64(warlock.Talents.DestructiveReach),
			BaseDamage:       core.BaseDamageConfigMagicNoRoll(785/5, 0.2),
			OutcomeApplier:   warlock.OutcomeFuncTick(),
			IsPeriodic:       true,
			ProcMask:         core.ProcMaskPeriodicDamage,
		}),
	})
}

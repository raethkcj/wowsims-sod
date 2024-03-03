package rogue

import (
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
)

func (rogue *Rogue) registerAmbushSpell() {
	flatDamageBonus := map[int32]float64{
		25: 28,
		40: 50,
		50: 92,
		60: 116,
	}[rogue.Level]

	spellID := map[int32]int32{
		25: 8676,
		40: 8725,
		50: 11268,
		60: 11269,
	}[rogue.Level]

	// waylay := rogue.HasRune(proto.RogueRune_RuneWaylay)

	rogue.Ambush = rogue.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		ProcMask:    core.ProcMaskMeleeMHSpecial,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagIncludeTargetBonusDamage | SpellFlagBuilder | SpellFlagColdBlooded | core.SpellFlagAPL,

		EnergyCost: core.EnergyCostOptions{
			Cost: rogue.costModifier(60 -
				core.TernaryFloat64(rogue.HasRune(proto.RogueRune_RuneSlaughterFromTheShadows), 20, 0)),
			Refund: 0.8,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand) && rogue.IsStealthed()
		},

		BonusCritRating: 15 * core.CritRatingPerCritChance * float64(rogue.Talents.ImprovedAmbush),
		// All of these use "Apply Aura: Modifies Damage/Healing Done", and stack additively.
		DamageMultiplier: 2.5 * (1 +
			0.04*float64(rogue.Talents.Opportunity)),
		CritMultiplier:   rogue.MeleeCritMultiplier(true),
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			rogue.BreakStealth(sim)
			baseDamage := flatDamageBonus +
				spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower()) +
				spell.BonusWeaponDamage()

			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if result.Landed() {
				rogue.AddComboPoints(sim, 1, spell.ComboPointMetrics())
				/** Currently does not apply to bosses due to being a slow
				if waylay {
					rogue.WaylayAuras.Get(target).Activate(sim)
				} */
			} else {
				spell.IssueRefund(sim)
			}
		},

		RelatedAuras: []core.AuraArray{rogue.WaylayAuras},
	})
}

# Baseline Investment ROI

**Investment in strong baseline vs time saved**

---

## ROI Formula

```
ROI = time_saved / baseline_investment

Where:
- time_saved = (standard_iterations - actual_iterations) × avg_iteration_time
- baseline_investment = (iteration_0_time - minimal_baseline_time)
```

---

## Examples

### Bootstrap-003 (High ROI)

```
Baseline investment: 120 min (vs 60 min minimal) = +60 min
Iterations saved: 6 - 3 = 3 iterations
Time per iteration: ~3 hours
Time saved: 3 × 3h = 9 hours = 540 min

ROI = 540 min / 60 min = 9x
```

### Bootstrap-002 (Low Investment)

```
Baseline investment: 60 min (minimal)
Result: 6 iterations (standard)
No time saved (baseline approach)
ROI = 0x (but no risk either)
```

---

## Investment Levels

| Investment | V_meta(s₀) | Iterations | Time Saved | ROI |
|------------|------------|------------|------------|-----|
| 8-10h | 0.70-0.80 | 3 | 15-20h | 2-3x |
| 6-8h | 0.50-0.70 | 3-4 | 12-18h | 2-3x |
| 4-6h | 0.40-0.50 | 4-5 | 8-12h | 2-2.5x |
| 2-4h | 0.20-0.40 | 5-7 | 0-4h | 0-1x |
| <2h | <0.20 | 7-10 | N/A | N/A |

---

**Recommendation**: Invest 4-6 hours for V_meta(s₀) = 0.40-0.50 (2-3x ROI).

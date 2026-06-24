# /// script
# requires-python = ">=3.11"
# dependencies = ["matplotlib>=3.8", "numpy>=1.26"]
# ///
"""
Generate dataset growth projection chart → site/dogfood/growth.png
Uses vaked dark aesthetic (#070b16 bg, #00d4ff cyan, #00e660 green).

Projection assumptions:
  - Current pace: ~500 records/day (dogfeed loop + 6-site telemetry)
  - 3x YoY growth as ecosystem expands (new sites, continuous loop)
  - Conservative lower bound: 1.5x YoY
  - Optimistic upper bound: 5x YoY

Usage:
  uv run scripts/plot_growth.py
  python3 scripts/plot_growth.py
"""

import numpy as np
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.ticker as mticker
from pathlib import Path

OUT = Path(__file__).parent.parent / "site" / "dogfood" / "growth.png"

# ── projection math ──────────────────────────────────────────────────────────
START_RECORDS = 250          # current total records
START_RATE    = 500          # records/day now
GROWTH_YR     = 3.0          # central: 3x per year
GROWTH_LOW    = 1.5          # conservative: 1.5x
GROWTH_HIGH   = 5.0          # optimistic: 5x

# generate daily points from Jun 2026 → Dec 2040 (5328 days)
days = np.arange(0, 365 * 14 + 185)          # ~14.5 years
years_elapsed = days / 365.0

def project(growth_per_year: float) -> np.ndarray:
    """Cumulative records: integral of rate(t) = START_RATE * growth^t dt"""
    if growth_per_year == 1.0:
        return START_RECORDS + START_RATE * days
    g = growth_per_year
    # rate(t) = START_RATE * g^t  →  integral = START_RATE * (g^t - 1) / ln(g)
    return START_RECORDS + START_RATE * (g ** years_elapsed - 1) / np.log(g)

central  = project(GROWTH_YR)
low      = project(GROWTH_LOW)
high     = project(GROWTH_HIGH)

# ── milestones ────────────────────────────────────────────────────────────────
def find_milestone(arr, target):
    idx = np.searchsorted(arr, target)
    if idx >= len(arr): return None
    return 2026 + idx / 365.0

milestones = {
    "1K":   1_000,
    "10K":  10_000,
    "100K": 100_000,
    "1M":   1_000_000,
    "10M":  10_000_000,
    "100M": 100_000_000,
}

# ── plot ──────────────────────────────────────────────────────────────────────
BG      = "#070b16"
SURFACE = "#0a0a14"
CARD    = "#14141f"
FG      = "#e0e8f5"
ACCENT  = "#00d4ff"
GREEN   = "#00e660"
DIM     = "#6878a0"
BORDER  = "#26304a"
WARN    = "#ffb020"

fig, ax = plt.subplots(figsize=(11, 5.5), facecolor=BG)
ax.set_facecolor(SURFACE)
for spine in ax.spines.values():
    spine.set_edgecolor(BORDER)

x_years = 2026 + days / 365.0

# confidence band
ax.fill_between(x_years, low, high, alpha=0.12, color=ACCENT, linewidth=0)

# central line
ax.semilogy(x_years, central, color=ACCENT, linewidth=2, label="projected (3× YoY)")
ax.semilogy(x_years, low,     color=DIM,    linewidth=1, linestyle="--", alpha=0.7, label="conservative (1.5× YoY)")
ax.semilogy(x_years, high,    color=GREEN,  linewidth=1, linestyle="--", alpha=0.7, label="optimistic (5× YoY)")

# today marker
ax.axvline(2026 + 175/365, color=WARN, linewidth=1, linestyle=":", alpha=0.7)
ax.text(2026 + 175/365 + 0.05, 200, "now", color=WARN, fontsize=8,
        fontfamily="monospace", va="bottom")

# milestone annotations
for label, val in milestones.items():
    yr = find_milestone(central, val)
    if yr and yr < 2041:
        ax.axhline(val, color=BORDER, linewidth=0.5, linestyle=":", alpha=0.5)
        ax.text(2026.1, val * 1.15, label,
                color=DIM, fontsize=7.5, fontfamily="monospace", va="bottom")

# axes
ax.set_xlim(2026, 2041)
ax.set_ylim(100, high[-1] * 3)
ax.set_xlabel("year", color=DIM, fontsize=9, fontfamily="monospace")
ax.set_ylabel("total records (log scale)", color=DIM, fontsize=9, fontfamily="monospace")
ax.tick_params(colors=DIM, labelsize=8)
for label in ax.get_xticklabels() + ax.get_yticklabels():
    label.set_fontfamily("monospace")
ax.xaxis.set_major_locator(mticker.MultipleLocator(2))
ax.xaxis.set_minor_locator(mticker.MultipleLocator(1))
ax.yaxis.set_major_formatter(mticker.FuncFormatter(
    lambda v, _: f"{v/1e6:.0f}M" if v >= 1e6 else
                  f"{v/1e3:.0f}K" if v >= 1e3 else f"{v:.0f}"
))
ax.grid(axis="y", color=BORDER, linewidth=0.4, alpha=0.5)
ax.grid(axis="x", color=BORDER, linewidth=0.3, alpha=0.3, which="minor")

# title + legend
ax.set_title("PeetPedro/ultrawhale-dogfood — projected growth",
             color=FG, fontsize=10, fontfamily="monospace", pad=12)
legend = ax.legend(fontsize=8, facecolor=CARD, edgecolor=BORDER,
                   labelcolor=FG, loc="upper left")
for text in legend.get_texts():
    text.set_fontfamily("monospace")

# current record count callout
ax.annotate(f"today: ~250 records\n~500 rec/day",
            xy=(2026 + 175/365, 250),
            xytext=(2027.2, 800),
            color=WARN, fontsize=7.5, fontfamily="monospace",
            arrowprops=dict(arrowstyle="->", color=WARN, lw=0.8),
            bbox=dict(boxstyle="round,pad=0.3", facecolor=CARD, edgecolor=WARN, alpha=0.8))

fig.tight_layout(pad=1.5)
fig.savefig(OUT, dpi=150, bbox_inches="tight", facecolor=BG)
print(f"saved → {OUT}  ({OUT.stat().st_size // 1024} KB)")

# print milestone table
print("\ncentral projection milestones:")
for label, val in milestones.items():
    yr = find_milestone(central, val)
    print(f"  {label:>6}  →  {yr:.1f}" if yr else f"  {label:>6}  →  after 2040")

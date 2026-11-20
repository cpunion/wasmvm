"""Format observed performance deltas without changing statistical controls."""
import math


def percent(ratio, *, higher_is_better=False):
    """Convert candidate/reference to signed speed change (positive improves)."""
    if not math.isfinite(ratio) or ratio <= 0:
        raise ValueError("comparison ratios must be finite and positive")
    return 100 * (ratio - 1 if higher_is_better else 1 / ratio - 1)


def signed(value):
    # Avoid a negative zero after rounding a negligible difference.
    return f"{0.0 if abs(value) < 0.005 else value:+.2f}%"


def observed(value, reference, *, higher_is_better=False):
    """Describe a ratio of displayed observations, without an acceptance claim."""
    if not math.isfinite(value) or value <= 0:
        raise ValueError("observations must be finite and positive")
    if not math.isfinite(reference) or reference <= 0:
        raise ValueError("reference observations must be finite and positive")
    return signed(percent(value / reference, higher_is_better=higher_is_better))


def comparison(stats, *, higher_is_better=False, speed=True):
    """Retain paired ratios/CIs and transform the same interval to percentages.

    Counts such as allocations use speed=False: report their quantity delta,
    without describing fewer objects as faster execution.
    """
    ratio = stats["ratio"]
    low, high = stats["ci95"]
    for value in [ratio, low, high]:
        percent(value)
    if not low <= ratio <= high:
        raise ValueError("the ratio must lie inside its ordered interval")
    text = (f"{ratio:.3f} [{low:.3f}, {high:.3f}]; "
            f"Δ {signed(100 * (ratio - 1))}")
    if speed:
        bounds = sorted(percent(x, higher_is_better=higher_is_better)
                        for x in [low, high])
        text += (f"; speed {signed(percent(ratio, higher_is_better=higher_is_better))} "
                 f"[{signed(bounds[0])}, {signed(bounds[1])}]")
    return text + ("; pass" if stats.get("controlsResolve") else "; unresolved")

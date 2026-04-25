package tools

// checkDCACondition evaluates the TriggerConfig condition for a DCA plan.
// Supported indicators mirror the calculate_indicators tool:
//
//	rsi   : oversold, overbought
//	sma   : price_above, price_below, cross_above, cross_below  (period vs period2)
//	ema   : price_above, price_below, cross_above, cross_below  (period vs period2)
//	macd  : histogram_positive, histogram_negative, macd_above_signal, macd_below_signal
//	bb    : touch_upper, touch_lower, outside_upper, outside_lower
//	atr   : above_threshold, below_threshold
//	stoch : oversold, overbought
//	vwap  : price_above, price_below

import (
	"context"
	"fmt"
	"strings"

	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/dca"
	"github.com/cryptoquantumwave/khunquant/pkg/providers/broker"
	"github.com/cryptoquantumwave/khunquant/pkg/ta"
)

// checkDCACondition returns (conditionMet, humanReason, error).
// humanReason explains why the condition was NOT met (used in skip messages).
func checkDCACondition(ctx context.Context, plan *dca.Plan, cfg *config.Config) (bool, string, error) {
	tc := plan.TriggerConfig
	if tc == nil {
		return true, "", nil
	}

	limit := tc.Limit
	if limit <= 0 {
		limit = 100
	}

	p, err := broker.CreateProviderForAccount(plan.Provider, plan.Account, cfg)
	if err != nil {
		return false, "", fmt.Errorf("create provider: %w", err)
	}
	md, ok := p.(broker.MarketDataProvider)
	if !ok {
		return false, "", fmt.Errorf("provider %q does not support market data", plan.Provider)
	}
	candles, err := md.FetchOHLCV(ctx, plan.Symbol, tc.Timeframe, nil, limit)
	if err != nil {
		return false, "", fmt.Errorf("FetchOHLCV: %w", err)
	}
	if len(candles) < 20 {
		return false, "", fmt.Errorf("not enough candles: got %d, need at least 20", len(candles))
	}

	closes := make([]float64, len(candles))
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	volumes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
		highs[i] = c.High
		lows[i] = c.Low
		volumes[i] = c.Volume
	}

	switch strings.ToLower(tc.Indicator) {
	case "rsi":
		return checkRSI(closes, tc)
	case "sma":
		return checkSMAEMA(closes, tc, false)
	case "ema":
		return checkSMAEMA(closes, tc, true)
	case "macd":
		return checkMACD(closes, tc)
	case "bb", "bollinger_bands":
		return checkBB(closes, tc)
	case "atr":
		return checkATR(highs, lows, closes, tc)
	case "stoch":
		return checkStoch(highs, lows, closes, tc)
	case "vwap":
		return checkVWAP(highs, lows, closes, volumes, tc)
	default:
		return false, "", fmt.Errorf("unsupported indicator %q", tc.Indicator)
	}
}

func checkRSI(closes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	period := defInt(tc.Period, 14)
	vals := ta.RSI(closes, period)
	if len(vals) == 0 {
		return false, "", fmt.Errorf("RSI: insufficient data for period %d", period)
	}
	last := vals[len(vals)-1]
	switch tc.Condition {
	case "oversold":
		threshold := defFloat(tc.Threshold, 30)
		if last < threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("RSI(%.2f) >= %.2f (not oversold)", last, threshold), nil
	case "overbought":
		threshold := defFloat(tc.Threshold, 70)
		if last > threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("RSI(%.2f) <= %.2f (not overbought)", last, threshold), nil
	default:
		return false, "", fmt.Errorf("RSI: unknown condition %q", tc.Condition)
	}
}

func checkSMAEMA(closes []float64, tc *dca.TriggerConfig, useEMA bool) (bool, string, error) {
	name := "SMA"
	if useEMA {
		name = "EMA"
	}
	period := defInt(tc.Period, 20)
	var vals []float64
	if useEMA {
		vals = ta.EMA(closes, period)
	} else {
		vals = ta.SMA(closes, period)
	}
	if len(vals) == 0 {
		return false, "", fmt.Errorf("%s: insufficient data for period %d", name, period)
	}
	last := vals[len(vals)-1]
	lastClose := closes[len(closes)-1]

	switch tc.Condition {
	case "price_above":
		if lastClose > last {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) <= %s(%d)=%.4f", lastClose, name, period, last), nil
	case "price_below":
		if lastClose < last {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) >= %s(%d)=%.4f", lastClose, name, period, last), nil
	case "cross_above", "cross_below":
		period2 := defInt(tc.Period2, 50)
		var vals2 []float64
		if useEMA {
			vals2 = ta.EMA(closes, period2)
		} else {
			vals2 = ta.SMA(closes, period2)
		}
		if len(vals2) < 2 || len(vals) < 2 {
			return false, "", fmt.Errorf("%s: insufficient data for cross detection", name)
		}
		prev1 := vals[len(vals)-2]
		prev2 := vals2[len(vals2)-2]
		cur1 := vals[len(vals)-1]
		cur2 := vals2[len(vals2)-1]
		if tc.Condition == "cross_above" {
			met := prev1 <= prev2 && cur1 > cur2
			if met {
				return true, "", nil
			}
			return false, fmt.Sprintf("%s(%d)=%.4f did not cross above %s(%d)=%.4f", name, period, cur1, name, period2, cur2), nil
		}
		met := prev1 >= prev2 && cur1 < cur2
		if met {
			return true, "", nil
		}
		return false, fmt.Sprintf("%s(%d)=%.4f did not cross below %s(%d)=%.4f", name, period, cur1, name, period2, cur2), nil
	default:
		return false, "", fmt.Errorf("%s: unknown condition %q", name, tc.Condition)
	}
}

func checkMACD(closes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	fast := defInt(tc.Period, 12)
	slow := defInt(tc.Period2, 26)
	signal := defInt(tc.Period3, 9)
	result := ta.MACD(closes, fast, slow, signal)
	if result == nil || len(result.Histogram) == 0 {
		return false, "", fmt.Errorf("MACD: insufficient data")
	}
	lastHist := result.Histogram[len(result.Histogram)-1]
	lastMACD := result.MACD[len(result.MACD)-1]
	lastSig := result.Signal[len(result.Signal)-1]
	switch tc.Condition {
	case "histogram_positive":
		if lastHist > 0 {
			return true, "", nil
		}
		return false, fmt.Sprintf("MACD histogram=%.4f <= 0", lastHist), nil
	case "histogram_negative":
		if lastHist < 0 {
			return true, "", nil
		}
		return false, fmt.Sprintf("MACD histogram=%.4f >= 0", lastHist), nil
	case "macd_above_signal":
		if lastMACD > lastSig {
			return true, "", nil
		}
		return false, fmt.Sprintf("MACD(%.4f) <= signal(%.4f)", lastMACD, lastSig), nil
	case "macd_below_signal":
		if lastMACD < lastSig {
			return true, "", nil
		}
		return false, fmt.Sprintf("MACD(%.4f) >= signal(%.4f)", lastMACD, lastSig), nil
	default:
		return false, "", fmt.Errorf("MACD: unknown condition %q", tc.Condition)
	}
}

func checkBB(closes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	period := defInt(tc.Period, 20)
	mult := defFloat(tc.Multiplier, 2.0)
	bb := ta.BollingerBands(closes, period, mult)
	if bb == nil || len(bb.Upper) == 0 {
		return false, "", fmt.Errorf("BB: insufficient data for period %d", period)
	}
	upper := bb.Upper[len(bb.Upper)-1]
	lower := bb.Lower[len(bb.Lower)-1]
	lastClose := closes[len(closes)-1]
	switch tc.Condition {
	case "touch_upper":
		if lastClose >= upper {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) < BB upper(%.4f)", lastClose, upper), nil
	case "touch_lower":
		if lastClose <= lower {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) > BB lower(%.4f)", lastClose, lower), nil
	case "outside_upper":
		if lastClose > upper {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) <= BB upper(%.4f)", lastClose, upper), nil
	case "outside_lower":
		if lastClose < lower {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) >= BB lower(%.4f)", lastClose, lower), nil
	default:
		return false, "", fmt.Errorf("BB: unknown condition %q", tc.Condition)
	}
}

func checkATR(highs, lows, closes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	period := defInt(tc.Period, 14)
	vals := ta.ATR(highs, lows, closes, period)
	if len(vals) == 0 {
		return false, "", fmt.Errorf("ATR: insufficient data for period %d", period)
	}
	last := vals[len(vals)-1]
	switch tc.Condition {
	case "above_threshold":
		if last > tc.Threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("ATR(%.4f) <= threshold(%.4f)", last, tc.Threshold), nil
	case "below_threshold":
		if last < tc.Threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("ATR(%.4f) >= threshold(%.4f)", last, tc.Threshold), nil
	default:
		return false, "", fmt.Errorf("ATR: unknown condition %q", tc.Condition)
	}
}

func checkStoch(highs, lows, closes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	kPeriod := defInt(tc.Period, 14)
	dPeriod := defInt(tc.Period2, 3)
	result := ta.Stochastic(highs, lows, closes, kPeriod, dPeriod)
	if result == nil || len(result.K) == 0 {
		return false, "", fmt.Errorf("Stochastic: insufficient data")
	}
	lastK := result.K[len(result.K)-1]
	switch tc.Condition {
	case "oversold":
		threshold := defFloat(tc.Threshold, 20)
		if lastK < threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("Stoch%%K(%.2f) >= %.2f (not oversold)", lastK, threshold), nil
	case "overbought":
		threshold := defFloat(tc.Threshold, 80)
		if lastK > threshold {
			return true, "", nil
		}
		return false, fmt.Sprintf("Stoch%%K(%.2f) <= %.2f (not overbought)", lastK, threshold), nil
	default:
		return false, "", fmt.Errorf("Stochastic: unknown condition %q", tc.Condition)
	}
}

func checkVWAP(highs, lows, closes, volumes []float64, tc *dca.TriggerConfig) (bool, string, error) {
	vals := ta.VWAP(highs, lows, closes, volumes)
	if len(vals) == 0 {
		return false, "", fmt.Errorf("VWAP: insufficient data")
	}
	lastVWAP := vals[len(vals)-1]
	lastClose := closes[len(closes)-1]
	switch tc.Condition {
	case "price_above":
		if lastClose > lastVWAP {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) <= VWAP(%.4f)", lastClose, lastVWAP), nil
	case "price_below":
		if lastClose < lastVWAP {
			return true, "", nil
		}
		return false, fmt.Sprintf("price(%.4f) >= VWAP(%.4f)", lastClose, lastVWAP), nil
	default:
		return false, "", fmt.Errorf("VWAP: unknown condition %q", tc.Condition)
	}
}

func defInt(v, d int) int {
	if v <= 0 {
		return d
	}
	return v
}

func defFloat(v, d float64) float64 {
	if v <= 0 {
		return d
	}
	return v
}

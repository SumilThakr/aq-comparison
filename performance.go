package main

import (
	"fmt"
	"math"
)

//Mean bias: 1/N Σ(Mi - Oi)
func meanBias(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	recip := 1 / float64(len(Data))
	var summa float64
	for _, xy := range Data {
		arg := (xy.x - xy.y)
		summa += arg
	}
	return summa * recip, nil
}

//Mean error: 1/N Σ|Mi - Oi|
func meanError(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	recip := 1 / float64(len(Data))
	var summa float64
	for _, xy := range Data {
		arg := math.Abs(xy.x - xy.y)
		summa += arg
	}
	return summa * recip, nil
}

//Root Mean Squared Error: √((Σ(Mi-Oi))²/N)
func rmse(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	for _, xy := range Data {
		arg := xy.x - xy.y
		summa += arg
	}
	return math.Sqrt(math.Pow(summa, 2) / float64(len(Data))), nil
}

// Fractional Bias: 100% x 2/NΣ((Mi-Oi)/(Mi+Oi))
func fracBias(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	for _, xy := range Data {
		arg := (xy.x - xy.y) / (xy.x + xy.y)
		summa += arg
	}
	return 100 * (2 / float64(len(Data))) * summa, nil
}

// Fractional Error: 100% x 2/NΣ(|Mi-Oi|/(Mi+Oi))
func fracError(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	for _, xy := range Data {
		arg := (math.Abs(xy.x-xy.y) / (xy.x + xy.y))
		summa += arg
	}
	return 100 * (2 / float64(len(Data))) * summa, nil
}

// Normalised Mean Bias: 100% x Σ(Mi-Oi)/ΣOi
func normMeanBias(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	var sumOi float64
	for _, xy := range Data {
		arg := (xy.x - xy.y)
		summa += arg
		sumOi += xy.y
	}
	return 100 * summa / sumOi, nil
}

// Normalised Mean Error: 100% x Σ|Mi-Oi|/ΣOi
func normMeanError(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	var sumOi float64
	for _, xy := range Data {
		arg := math.Abs(xy.x - xy.y)
		summa += arg
		sumOi += xy.y
	}
	return 100 * summa / sumOi, nil
}

// Mean Normalized Bias: 100% x 1/NΣ((Mi-Oi)/Oi)
func meanNormBias(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	for _, xy := range Data {
		if xy.y != 0 {
			arg := ((xy.x - xy.y) / xy.y)
			summa += arg
		}
	}
	return 100 * (1 / float64(len(Data))) * summa, nil
}

// Mean Normalized Error: 100% x 1/NΣ|((Mi-Oi)/Oi)|
func meanNormError(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa float64
	for _, xy := range Data {
		if xy.y != 0 {
			arg := math.Abs((xy.x - xy.y) / xy.y)
			summa += arg
		}
	}
	return 100 * (1 / float64(len(Data))) * summa, nil
}

// Unpaired Peak Accuracy: 100% x (Mpeak - Opeak)/Opeak
func unpairedPeakAcc(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var mPeak, oPeak float64
	for _, xy := range Data {
		if xy.x > mPeak {
			mPeak = xy.x
		}
		if xy.y > oPeak {
			oPeak = xy.y
		}
	}
	return 100 * (mPeak - oPeak) / oPeak, nil
}

// Index of Agreement: 1 - Σ(Mi-Oi)²/Σ(|Mi-O*|+|Oi-O*|)
func indexOfAgr(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var summa, sumDenom, oMean float64
	for _, xy := range Data {
		oMean += xy.y
	}
	oMean = oMean / float64(len(Data))
	for _, xy := range Data {
		arg := math.Pow((xy.x - xy.y), 2)
		summa += arg
		denom := math.Abs(xy.x-oMean) + math.Abs(xy.y-oMean)
		sumDenom += denom
	}
	return 1 - (summa / sumDenom), nil
}

// Coefficient of Determination: ((Σⁿ₁((Mi-M*)x(Oi-O*)))/√(Σⁿ₁(Mi-M*)²Σⁿ(Oi-O*)²))²
func coefDeterm(Data []xy) (float64, error) {
	if len(Data) == 0 {
		return 0, fmt.Errorf("The data input is nil")
	}
	var sumNum, sumDenomM, sumDenomO, oMean, mMean float64
	for _, xy := range Data {
		mMean += xy.x
		oMean += xy.y
	}
	mMean = mMean / float64(len(Data))
	oMean = oMean / float64(len(Data))
	for _, xy := range Data {
		numerator := (xy.x - mMean) * (xy.y - oMean)
		sumNum += numerator
		argM := math.Pow(xy.x-mMean, 2)
		sumDenomM += argM
		argO := math.Pow(xy.y-oMean, 2)
		sumDenomO += argO
	}
	return math.Pow((sumNum)/math.Sqrt(sumDenomM*sumDenomO), 2), nil
}

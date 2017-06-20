package assets

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	usdSizeSeparator             = "."
	usdStringFractionLength      = 16
	usdIntFractionLength         = 2
	usdFractionSignificantDigits = 2
)

var usdCoinMultiplier = int64(math.Pow10(usdIntFractionLength))

// Compare compare usd ascending
func (usd usdStruct) Compare(other USD) int {
	if usd.GetIntValue() < other.GetIntValue() {
		return -1
	} else if usd.GetIntValue() > other.GetIntValue() {
		return 1
	}
	return 0
}

func (usd usdStruct) GetFractionLength() int64 {
	return usdIntFractionLength
}

// NewUSDFromString create USD from string
func NewUSDFromString(usdString string) (USD, error) {
	usdString = standardizeUsdString(usdString)
	usdInt, err := convertUsdStringToInt(usdString)
	if err != nil {
		return nil, err
	}
	return usdStruct{stringValue: usdString, intValue: usdInt}, nil
}

// NewUSDFromInt create USD from int value
func NewUSDFromInt(usdInt int64) USD {
	usdString := convertUsdIntToString(usdInt, usdStringFractionLength)
	return usdStruct{stringValue: usdString, intValue: usdInt}
}

func (usd usdStruct) GetStringValue() string {
	return usd.stringValue
}

func (usd usdStruct) GetPrettyStringValue() string {
	return convertUsdIntToString(usd.intValue, usdIntFractionLength)
}

func (usd usdStruct) GetIntValue() int64 {
	return usd.intValue
}

func (usd usdStruct) Add(usdToAdd USD) USD {
	usdInt := usd.GetIntValue()
	usdInt += usdToAdd.GetIntValue()
	return NewUSDFromInt(usdInt)
}

func (usd usdStruct) Subtract(usdToSubtract USD) USD {
	usdInt := usd.GetIntValue()
	usdInt -= usdToSubtract.GetIntValue()
	return NewUSDFromInt(usdInt)
}

func (usd usdStruct) Multiply(value int64, fractionLength int64) USD {
	percentMultiplier := int64(math.Pow10(int(fractionLength)))
	usdValue := usd.GetIntValue() * value / percentMultiplier
	return NewUSDFromInt(usdValue)
}

func convertUsdStringToInt(usdString string) (int64, error) {
	pieces := strings.Split(usdString, usdSizeSeparator)
	cents := int64(0)
	if len(pieces) > 1 {
		var err error
		cents, err = convertUsdFractionStringToInt(pieces[1])
		if err != nil {
			return 0, ConversionError{message: "Error converting fraction usd [" + pieces[1] + "] for string [" + usdString + "] -- [" + err.Error() + "]"}
		}
	}
	dollars, err := convertWholeUsdStringToInt(pieces)
	if err != nil {
		return 0, ConversionError{message: "Error converting whole usd [" + pieces[0] + "] -- [" + err.Error() + "]"}
	}
	usdRepresentation := dollars + cents
	return int64(usdRepresentation), nil
}

func standardizeUsdString(usdString string) string {
	pieces := strings.Split(usdString, usdSizeSeparator)
	fractionString := "0"
	if len(pieces) > 1 {
		fractionString = pieces[1]
	}
	fractionLength := utf8.RuneCountInString(fractionString)
	var fractionBuffer bytes.Buffer
	fractionBuffer.WriteString(fractionString)
	for i := fractionLength; i < usdStringFractionLength; i++ {
		fractionBuffer.WriteString("0")
	}
	var standardizedBuffer bytes.Buffer
	standardizedBuffer.WriteString(pieces[0])
	standardizedBuffer.WriteString(usdSizeSeparator)
	standardizedBuffer.WriteString(fractionBuffer.String())
	return standardizedBuffer.String()
}

func convertUsdFractionStringToInt(fractionString string) (int64, error) {
	fraction, err := strconv.Atoi(fractionString)
	if err != nil {
		return 0, err
	}
	unrounded := int64(fraction)
	return roundUsdFromStringRepresentation(unrounded), nil
}

func convertWholeUsdStringToInt(pieces []string) (int64, error) {
	dollar, err := strconv.Atoi(pieces[0])
	if err != nil {
		return 0, err
	}
	dollar = dollar * int(usdCoinMultiplier)
	return int64(dollar), nil
}

func convertUsdIntToString(usdInt int64, fractionsToPrint int) string {
	negative := usdInt < 0
	if negative {
		usdInt = usdInt * -1
	}
	wholeUsd := usdInt / usdCoinMultiplier
	fractionUsd := usdInt % usdCoinMultiplier
	fractionString := strconv.FormatInt(fractionUsd, 10)
	if fractionUsd < 10 {
		fractionString = "0" + fractionString
	}
	fractionLength := utf8.RuneCountInString(fractionString)
	wholeUsdString := strconv.FormatInt(wholeUsd, 10)
	var buffer bytes.Buffer
	if negative {
		buffer.WriteString("-")
	}
	buffer.WriteString(wholeUsdString)
	buffer.WriteString(usdSizeSeparator)
	buffer.WriteString(fractionString)
	for i := fractionLength; i < fractionsToPrint; i++ {
		buffer.WriteString("0")
	}
	return buffer.String()
}

func roundUsdFromStringRepresentation(usdInt int64) int64 {
	roundingPlace := usdStringFractionLength - usdFractionSignificantDigits
	significantModulus := int64(10)
	roundingValueMultiplier := 1
	for i := 1; i <= roundingPlace; i++ {
		currentRoundingValue := usdInt % significantModulus
		if currentRoundingValue >= int64(roundingValueMultiplier*5) {
			valueToAdd := significantModulus - currentRoundingValue
			usdInt += valueToAdd
		} else {
			usdInt -= currentRoundingValue
		}
		significantModulus = significantModulus * 10
		roundingValueMultiplier = roundingValueMultiplier * 10
	}
	divisor := int64(math.Pow10(roundingPlace))
	return usdInt / divisor
}

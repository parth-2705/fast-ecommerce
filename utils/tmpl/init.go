package tmpl

import "strings"

func GetFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"round":                  Round,
		"int":                    RoundToInt,
		"trunc":                  Truncate,
		"getEnvVariable":         GetEnvVariable,
		"VolumetricWeight":       VolumetricWeight,
		"dict":                   MakeMap,
		"toString":               ToString,
		"substring":              Substring,
		"minus":                  Minus,
		"minusInt":               MinusInt,
		"addInt":                 AddInt,
		"add":                    Add,
		"timeLeft":               TimeLeft,
		"timeLeftInHours":        TimeLeftInHours,
		"calculateDiscount":      CalculateDiscount,
		"formatTimestamp":        FormatTimestamp,
		"toUpperCase":            ToUpperCase,
		"toTitleCase":            ToTitleCase,
		"taxableValue":           TaxableValue,
		"taxValue":               TaxValue,
		"randomizer":             GetRandomizerStringtoControlCaching,
		"isProd":                 IsProd,
		"getImageURL":            GetImageURLByEnvironment,
		"getImageURLNew":           GetImageURLByEnvironment2,
		"replaceAll":             strings.ReplaceAll,
		"randomInt":              RandomInt,
		"makeGenderChoiceString": MakeGenderChoiceString,
		"join":                   strings.Join,
	}
}

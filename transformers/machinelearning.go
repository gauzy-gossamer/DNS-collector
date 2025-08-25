package transformers

import (
	"math"
	"strings"
	"unicode"

	"github.com/dmachard/go-dnscollector/dnsutils"
	"github.com/dmachard/go-dnscollector/pkgconfig"
	"github.com/dmachard/go-logger"
)

func isConsonant(char rune) bool {
	if !unicode.IsLower(char) && !unicode.IsUpper(char) {
		return false
	}
	switch char {
	case 'a', 'A', 'e', 'E', 'i', 'I', 'o', 'O', 'u', 'U', 'y', 'Y':
		return false
	}
	return true
}

type MlTransform struct {
	GenericTransformer
}

func NewMachineLearningTransform(config *pkgconfig.ConfigTransformers, logger *logger.Logger, name string, instance int, nextWorkers []chan dnsutils.DNSMessage) *MlTransform {
	t := &MlTransform{GenericTransformer: NewTransformer(config, logger, "machinelearning", name, instance, nextWorkers)}
	return t
}

func (t *MlTransform) GetTransforms() ([]Subtransform, error) {
	subtransforms := []Subtransform{}
	if t.config.MachineLearning.Enable {
		subtransforms = append(subtransforms, Subtransform{name: "machinelearning:add-feature", processFunc: t.addFeatures})
	}
	return subtransforms, nil
}

func (t *MlTransform) addFeatures(dm *dnsutils.DNSMessage) (int, error) {

	if dm.MachineLearning == nil {
		dm.MachineLearning = &dnsutils.TransformML{}
	}

	// count global number of chars
	n := float64(len(dm.DNS.Qname))
	if n == 0 {
		n = 1
	}

	// count number of unique chars
	uniq := make(map[rune]int)
	for _, c := range dm.DNS.Qname {
		uniq[c]++
	}

	// calculate the probability of occurrence for each unique character.
	probs := make(map[rune]float64)
	for char, count := range uniq {
		probs[char] = float64(count) / n
	}

	// calculate the entropy
	var entropy float64
	for _, prob := range probs {
		if prob > 0 {
			entropy -= prob * math.Log2(prob)
		}
	}

	// count digit
	countDigits := 0
	for _, char := range dm.DNS.Qname {
		if unicode.IsDigit(char) {
			countDigits++
		}
	}

	// count lower
	countLowers := 0
	for _, char := range dm.DNS.Qname {
		if unicode.IsLower(char) {
			countLowers++
		}
	}

	// count upper
	countUppers := 0
	for _, char := range dm.DNS.Qname {
		if unicode.IsUpper(char) {
			countUppers++
		}
	}

	// count specials
	countSpecials := 0
	for _, char := range dm.DNS.Qname {
		switch char {
		case '.', '-', '_', '=':
			countSpecials++
		}
	}

	// count others
	countOthers := len(dm.DNS.Qname) - (countDigits + countLowers + countUppers + countSpecials)

	// count labels
	numLabels := strings.Count(dm.DNS.Qname, ".") + 1

	// count consecutive chars
	consecutiveCount := 0
	nameLower := strings.ToLower(dm.DNS.Qname)
	for i := 1; i < len(nameLower); i++ {
		if nameLower[i] == nameLower[i-1] {
			consecutiveCount += 1
		}
	}

	// count consecutive vowel
	consecutiveVowelCount := 0
	for i := 1; i < len(nameLower); i++ {
		switch nameLower[i] {
		case 'a', 'e', 'i', 'o', 'u', 'y':
			if nameLower[i] == nameLower[i-1] {
				consecutiveVowelCount += 1
			}
		}
	}

	// count consecutive digit
	consecutiveDigitCount := 0
	for i := 1; i < len(nameLower); i++ {
		if unicode.IsDigit(rune(nameLower[i])) && unicode.IsDigit(rune(nameLower[i-1])) {
			consecutiveDigitCount += 1
		}
	}

	// count consecutive consonant
	consecutiveConsonantCount := 0
	for i := 1; i < len(nameLower); i++ {
		if isConsonant(rune(nameLower[i])) && isConsonant(rune(nameLower[i-1])) {
			consecutiveConsonantCount += 1
		}
	}

	// size
	dm.MachineLearning.Size = dm.DNS.Length
	if dm.Reducer != nil {
		dm.MachineLearning.Size = dm.Reducer.CumulativeLength
	}

	// occurrences
	if dm.Reducer != nil {
		dm.MachineLearning.Occurrences = dm.Reducer.Occurrences
	}

	// qtypes
	switch dm.DNS.Qtype {
	case "A", "AAAA", "HTTPS", "SRV", "PTR", "SOA", "NS":
		dm.MachineLearning.UncommonQtypes = 0
	default:
		dm.MachineLearning.UncommonQtypes = 1
	}

	dm.MachineLearning.Entropy = entropy
	dm.MachineLearning.Length = len(dm.DNS.Qname)
	dm.MachineLearning.Digits = countDigits
	dm.MachineLearning.Lowers = countLowers
	dm.MachineLearning.Uppers = countUppers
	dm.MachineLearning.Specials = countSpecials
	dm.MachineLearning.Others = countOthers
	dm.MachineLearning.Labels = numLabels
	dm.MachineLearning.RatioDigits = float64(countDigits) / n
	dm.MachineLearning.RatioLetters = float64(countLowers+countUppers) / n
	dm.MachineLearning.RatioSpecials = float64(countSpecials) / n
	dm.MachineLearning.RatioOthers = float64(countOthers) / n
	dm.MachineLearning.ConsecutiveChars = consecutiveCount
	dm.MachineLearning.ConsecutiveVowels = consecutiveVowelCount
	dm.MachineLearning.ConsecutiveDigits = consecutiveDigitCount
	dm.MachineLearning.ConsecutiveConsonants = consecutiveConsonantCount

	return ReturnKeep, nil
}

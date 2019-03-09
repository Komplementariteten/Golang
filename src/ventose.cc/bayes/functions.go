package bayes

func Odds(p float64) float64 {
	if p >= 1 {
		return 0
	}
	return p / (1 - p)
}

/*
Computes the Propability given its odds
Example: o=2 means 2:1 odds in favor, or 2/3 propability
*/
func Propability(o float64) float64 {
	return o / (o + 1)
}

func Propability2(yes float64, no float64) float64 {
	return yes / (yes + no)
}

/*
Helper Functions
*/

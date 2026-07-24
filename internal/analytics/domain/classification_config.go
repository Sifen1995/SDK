package domain

type ClassificationConfig struct {
	LookbackDays  int
	MinDaysActive int
	MinConfidence float64
	MaxAgeDays    int
	IntentClasses []string
}

func DefaultConfig() ClassificationConfig {
	return ClassificationConfig{
		LookbackDays:  30,
		MinDaysActive: 5,
		MinConfidence: 0.70,
		MaxAgeDays:    7,
		IntentClasses: []string{
			"coffee_interest",
			"crypto_interest",
			"fashion_interest",
			"shopping_interest",
			"food_interest",
			"fitness_interest",
			"fintech_interest",
			"travel_intent",
		},
	}
}

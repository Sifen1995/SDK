# skykin-ml/data/app_category_map.py
# This is your ground truth mapping

APP_CATEGORIES = {
    # Fashion & Shopping
    "com.shein.android":           "fashion",
    "com.zara.android":            "fashion",
    "com.aliexpress.androidapp":   "shopping",
    "com.amazon.mShop.android":    "shopping",
    "com.jumia.android":           "shopping",

    # Crypto & Finance
    "com.binance.dev":             "crypto",
    "com.coinbase.android":        "crypto",
    "com.telebirr":                "fintech",
    "com.cbebirr":                 "fintech",
    "net.boa.mobile":              "banking",

    # Food & Coffee
    "com.yene.rider":              "food",
    "com.addis.eats":              "food",

    # News & Information
    "com.bbc.mobile.news.ww":      "news",
    "com.addisstandard":           "news",

    # Social Media
    "com.instagram.android":       "social",
    "com.facebook.katana":         "social",
    "com.twitter.android":         "social",

    # Travel
    "com.booking.android":         "travel",
    "et.ethiopianairlines.app":    "travel",

    # Fitness
    "com.nike.ntc":                "fitness",
    "com.adidas.app":              "fitness",

    # Default for unknown apps
    "default":                     "other",
}
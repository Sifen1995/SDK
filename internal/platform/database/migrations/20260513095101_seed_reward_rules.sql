-- migrate:up

-- Create the table if it doesn't exist (safety check)
CREATE TABLE IF NOT EXISTS reward_rules (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_name  VARCHAR(100)   NOT NULL UNIQUE,
    reward_type  VARCHAR(50)    NOT NULL,
    amount       NUMERIC(10, 2) NOT NULL,
    currency     VARCHAR(50)    NOT NULL,
    message      TEXT           NOT NULL,
    is_active    BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Seed the initial rules matching your ML intents
INSERT INTO reward_rules (intent_name, reward_type, amount, currency, message)
VALUES 
    ('coffee_interest', 'cashback', 20.00, 'ETB', 'You earned 20 ETB cashback for your coffee passion!'),
    ('crypto_interest', 'coins', 50.00, 'FLIP_COINS', 'Crypto enthusiast! You earned 50 Flip Coins!'),
    ('fashion_interest', 'cashback', 15.00, 'ETB', 'Stylish! Here is 15 ETB store credit for your next look.'),
    ('abandoned_cart', 'discount', 10.00, 'PERCENT', 'We noticed you left something behind! Here is a 10% discount.'),
    ('signup_intent', 'points', 100.00, 'POINTS', 'Welcome to Skykin! You earned 100 loyalty points.')
ON CONFLICT (intent_name) DO UPDATE SET
    reward_type = EXCLUDED.reward_type,
    amount = EXCLUDED.amount,
    currency = EXCLUDED.currency,
    message = EXCLUDED.message;
-- migrate:down


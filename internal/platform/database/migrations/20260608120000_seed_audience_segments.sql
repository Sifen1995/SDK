-- migrate:up
-- Catalog Audiencemart cohorts (definitions only — membership resolved at runtime from intents).

CREATE UNIQUE INDEX IF NOT EXISTS uq_audience_segments_name ON audience_segments(name);

INSERT INTO audience_segments (
    name, description, top_intent_signals, approximate_size, estimated_cpm, available_from, is_active
)
SELECT v.name, v.description, v.signals::jsonb, v.approx_size, v.estimated_cpm, NOW(), true
FROM (VALUES
    (
        'Fashion Enthusiasts',
        'Users showing strong fashion and lifestyle purchase intent',
        '["fashion_interest"]',
        12500,
        4.50
    ),
    (
        'Crypto & Fintech',
        'Users interested in crypto trading and fintech products',
        '["crypto_interest", "fintech_interest"]',
        8200,
        6.00
    ),
    (
        'Food & Dining',
        'Food delivery and restaurant discovery intent',
        '["food_interest"]',
        15000,
        3.25
    ),
    (
        'Mobile Gamers',
        'Gaming and in-app engagement intent',
        '["gaming_interest"]',
        21000,
        2.75
    ),
    (
        'Lifelong Learners',
        'Education and upskilling intent',
        '["education_interest"]',
        9800,
        3.80
    ),
    (
        'Broad Reach',
        'General engagement across mixed verticals',
        '["general_interest", "fashion_interest", "food_interest"]',
        45000,
        1.50
    )
) AS v(name, description, signals, approx_size, estimated_cpm)
WHERE NOT EXISTS (SELECT 1 FROM audience_segments s WHERE s.name = v.name);

-- migrate:down
DELETE FROM audience_segments WHERE name IN (
    'Fashion Enthusiasts',
    'Crypto & Fintech',
    'Food & Dining',
    'Mobile Gamers',
    'Lifelong Learners',
    'Broad Reach'
);

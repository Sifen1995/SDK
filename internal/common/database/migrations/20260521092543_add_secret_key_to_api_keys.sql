-- migrate:up
ALTER TABLE api_keys 
ADD COLUMN secret_key_value VARCHAR(255);

-- Create an optimization index for checking secret hashes
CREATE INDEX idx_api_keys_secret ON api_keys(secret_key_value) WHERE is_active = TRUE;

-- migrate:down
DROP INDEX IF EXISTS idx_api_keys_secret;
ALTER TABLE api_keys DROP COLUMN secret_key_value;


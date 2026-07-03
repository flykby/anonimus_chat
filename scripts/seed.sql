-- Dev seed: one test persona and sample photos.
-- Run: make seed

INSERT INTO personas (name, gender, prompt_version, system_prompt, active)
SELECT
    'Lena',
    'female',
    'v1',
    'You are Lena, a friendly and playful girl chatting in Telegram.',
    TRUE
WHERE NOT EXISTS (SELECT 1 FROM personas WHERE name = 'Lena');

WITH persona AS (
    SELECT id FROM personas WHERE name = 'Lena' ORDER BY id LIMIT 1
)
INSERT INTO photos (persona_id, tags, nsfw_level, telegram_file_id, unlock_price_stars)
SELECT persona.id, ARRAY['selfie', 'smile'], 'safe', 'AgACAgIAAxkBAAITestSafe1', 0
FROM persona
ON CONFLICT (telegram_file_id) DO NOTHING;

WITH persona AS (
    SELECT id FROM personas WHERE name = 'Lena' ORDER BY id LIMIT 1
)
INSERT INTO photos (persona_id, tags, nsfw_level, telegram_file_id, unlock_price_stars)
SELECT persona.id, ARRAY['selfie', 'outdoor'], 'safe', 'AgACAgIAAxkBAAITestSafe2', 0
FROM persona
ON CONFLICT (telegram_file_id) DO NOTHING;

WITH persona AS (
    SELECT id FROM personas WHERE name = 'Lena' ORDER BY id LIMIT 1
)
INSERT INTO photos (persona_id, tags, nsfw_level, telegram_file_id, unlock_price_stars)
SELECT persona.id, ARRAY['lingerie'], 'adult', 'AgACAgIAAxkBAAITestAdult1', 50
FROM persona
ON CONFLICT (telegram_file_id) DO NOTHING;

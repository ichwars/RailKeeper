ALTER TABLE accessory_products ADD COLUMN set_contents_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(set_contents_json) AND json_type(set_contents_json) = 'array');

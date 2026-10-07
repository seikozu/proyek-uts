CREATE INDEX IF NOT EXISTS comics_created_at_id_desc_idx
    ON comics (created_at DESC, id DESC);
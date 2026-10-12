-- "Official" libraries ship with the application and can always be re-obtained, so a backup
-- records them as metadata only. A library the user imported is NEVER official, even if its
-- license says it is redistributable: the user may be the only one who has the file.
ALTER TABLE datasets ADD COLUMN is_official INTEGER NOT NULL DEFAULT 0 CHECK (is_official IN (0, 1));

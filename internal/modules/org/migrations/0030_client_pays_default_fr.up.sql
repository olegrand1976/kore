-- Align client country with société: empty → FR, default FR.
UPDATE org.clients SET pays = 'FR' WHERE TRIM(pays) = '';

ALTER TABLE org.clients
    ALTER COLUMN pays SET DEFAULT 'FR';

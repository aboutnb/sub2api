-- Preserve the check-in mode on the existing balance history records so
-- clients can label lucky check-ins without exposing internal notes.
UPDATE redeem_codes AS history
SET notes = 'checkin_mode:' || records.mode
FROM checkin_records AS records
WHERE history.code = 'SYS-CHECKIN-' || records.id
  AND history.type = 'checkin'
  AND (history.notes IS NULL OR history.notes = '');

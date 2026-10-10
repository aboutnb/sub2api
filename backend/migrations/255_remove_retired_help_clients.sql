-- Remove retired tutorials without changing historical seed migrations.
DELETE FROM help_documents
WHERE category IN ('read-frog', 'roo-code')
   OR slug IN ('clients/read-frog', 'clients/roo-code');

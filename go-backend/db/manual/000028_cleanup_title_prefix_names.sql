-- Migration / Maintenance: 000028_cleanup_title_prefix_names.sql
-- Removes administrative prefix names (e.g. เด็กหญิง, นางสาว, etc.) from public.names_miracle.
-- Backs up all removed rows to data_governance.names_miracle_prefix_backup for audit and recovery.

\set ON_ERROR_STOP on

BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '1min';

CREATE SCHEMA IF NOT EXISTS data_governance;

-- 1. Create or populate backup table with affected rows
CREATE TABLE IF NOT EXISTS data_governance.names_miracle_prefix_backup AS
SELECT * 
FROM public.names_miracle 
WHERE thname LIKE 'เด็กหญิง%' 
   OR thname LIKE 'เด็กหณิง%'
   OR thname LIKE 'เด็กชาย%'
   OR thname LIKE 'นางสาว%'
   OR thname LIKE 'คุณหญิง%'
   OR (thname LIKE 'นาย%' AND (meaning = 'ชื่อบุคคล' OR thname = 'นายภูริพัฒน์'));

-- 2. Delete affected rows from names_miracle
DELETE FROM public.names_miracle 
WHERE thname LIKE 'เด็กหญิง%' 
   OR thname LIKE 'เด็กหณิง%'
   OR thname LIKE 'เด็กชาย%'
   OR thname LIKE 'นางสาว%'
   OR thname LIKE 'คุณหญิง%'
   OR (thname LIKE 'นาย%' AND (meaning = 'ชื่อบุคคล' OR thname = 'นายภูริพัฒน์'));

COMMIT;

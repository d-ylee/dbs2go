CREATE PRIVATE TEMPORARY TABLE {{.Owner}}.ORA$PPT_TEMP_FILE_LUMIS
(RUN_NUM INTEGER, LUMI_SECTION_NUM INTEGER, FILE_ID INTEGER, EVENT_COUNT INTEGER)
ON COMMIT DROP DEFINITION
SELECT MR.MIGRATION_REQUEST_ID, MR.MIGRATION_URL,
       MR.MIGRATION_INPUT, MR.MIGRATION_STATUS,
       MR.CREATE_BY, MR.CREATION_DATE,
       MR.LAST_MODIFIED_BY, MR.LAST_MODIFICATION_DATE, MR.RETRY_COUNT
FROM {{.Owner}}.MIGRATION_REQUESTS MR
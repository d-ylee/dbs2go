SELECT 
    R.RELEASE_VERSION,
    P.PSET_HASH,
    FS.LOGICAL_FILE_NAME,
    A.APP_NAME,
    O.OUTPUT_MODULE_LABEL, 
    O.GLOBAL_TAG
FROM {{.Owner}}.OUTPUT_MODULE_CONFIGS O
JOIN {{.Owner}}.RELEASE_VERSIONS R ON O.RELEASE_VERSION_ID=R.RELEASE_VERSION_ID
JOIN {{.Owner}}.APPLICATION_EXECUTABLES A ON O.APP_EXEC_ID=A.APP_EXEC_ID
JOIN {{.Owner}}.PARAMETER_SET_HASHES P ON O.PARAMETER_SET_HASH_ID=P.PARAMETER_SET_HASH_ID
JOIN {{.Owner}}.FILE_OUTPUT_MOD_CONFIGS FC ON FC.OUTPUT_MOD_CONFIG_ID=O.OUTPUT_MOD_CONFIG_ID
JOIN {{.Owner}}.FILES FS ON FS.FILE_ID=FC.FILE_ID
JOIN {{.Owner}}.BLOCKS B ON B.BLOCK_ID = FS.BLOCK_ID
WHERE B.BLOCK_NAME = :blk
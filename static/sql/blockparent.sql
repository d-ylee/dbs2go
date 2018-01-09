SELECT BC.BLOCK_NAME as THIS_BLOCK_NAME, BP.BLOCK_NAME as PARENT_BLOCK_NAME FROM {{.Owner}}.BLOCKS BP
                        JOIN {{.Owner}}.BLOCK_PARENTS BPRTS
                            ON BPRTS.PARENT_BLOCK_ID = BP.BLOCK_ID
                        JOIN {{.Owner}}.BLOCKS BC
                            ON BPRTS.THIS_BLOCK_ID = BC.BLOCK_ID

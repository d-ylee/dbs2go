SELECT
        D.DATASET
	FROM {{.Owner}}.DATASETS D
	JOIN {{.Owner}}.DATASET_ACCESS_TYPES DP on DP.DATASET_ACCESS_TYPE_ID= D.DATASET_ACCESS_TYPE_ID
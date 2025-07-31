package main

var query_all_connection = `
 select
      _airbyte_raw_id as connection_id,
      cast(json_extract (_airbyte_data, '$.name') as varchar) as name,
      cast(json_extract (_airbyte_data, '$.prefix') as varchar) as prefix,
      cast(json_extract (_airbyte_data, '$.status') as varchar) as status
    from
      airbyte_destination_v2.raw_airbyte_cluster_data_connection
`
var query_all_connection_with_prefix = `
 select
      _airbyte_raw_id as connection_id,
      cast(json_extract (_airbyte_data, '$.name') as varchar) as name,
      cast(json_extract (_airbyte_data, '$.prefix') as varchar) as prefix,
      cast(json_extract (_airbyte_data, '$.status') as varchar) as status
    from
      airbyte_destination_v2.raw_airbyte_cluster_data_connection
 where cast(json_extract (_airbyte_data, '$.prefix') as varchar) like '%s'
`

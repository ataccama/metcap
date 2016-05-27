module MetricsCapacitor
  module Model
    INDEX_TEMPLATE = {
      template: "metrics*",
      settings: {
        number_of_shards: 2
      },
      mappings: {
        '_default_' => {
          '_source' => { 'enabled' => true },
          'dynamic_templates' => [
            {
              'values' => {
                'mapping' => {
                  'index' => 'not_analyzed',
                  'type' => 'float'
                },
                'path_match' => '@values.*'
              }
            },
            {
              'tags' => {
                'mapping' => {
                  'index' => 'not_analyzed',
                  'type' => 'string',
                  'copy_to' => '@uniq'
                },
                'path_match' => '@tags.*',
                'path_unmatch' => '@tags._counter'
              }
            }
          ],
          'properties' => {
            '@uniq' => {
              'type' => 'string',
              'index' => 'not_analyzed'
            },
            '@name' => {
              'type' => 'string',
              'index' => 'not_analyzed',
              'copy_to' => '@uniq'
            },
            '@timestamp' => {
              'type' => 'date'
            },
            '@tags' => {
              'properties' => {
                '_counter' => {
                  'type' => 'boolean',
                  'index' => 'not_analyzed'
                }
              }
            }
          }
        }
      }
    }
  end
end

## v1.0.0 / 2017-03-27:

First release of this exporter

#### NOTE: Collector Information
    * `mysql` - this collectors are not still tested on context


#### Collectors included : 
    * `bash`  - run custom command into the shell process (| are not allowed, so each result of command are include as stdin for the next command)
    * `redis` - run custom query to redis and parse result as JSON
    * `mysql` - run custom SQL Query and expose last columns as value

# WIP Prometheus custom_exporter

## Intro

This project is aimed to retrieve specific metrics that can't be found in dedicated exporters.

## Concepts
This exporter work with a dedicated config’s file that’s contain all needed metrics.
 
Each dedicated metrics will use a collector to access to data and a custom commands list that's will be run into this collector.
The final command must expose a ready to be parsed result set to extract metric value and tags value. 
A mapping config allows to specify the tags included otherwise, only last data on each row or record will be extract.
  
A collector is define by his name and include the type (bash, mysql ...) and credential (data source name, user, password, uri ...). 

The dedicated metrics are exposed with a prefix "custom" and the name of this metrics extract from the config file.
 
## How it's work
The main process will load the config file and will register a dedicated collector into Prometheus client framework for each metrics. Each metric is composed of a collector helper who's include a specific type collector defined into the config file

On each call to the metrics path of the exporter (i.e. http://localhost:9213/metrics/), the main process will call each registered Prometheus collectors in multithreading and grab all results to expose them to the caller.

If a metrics is not available (errors on running command, result empty ...) a minimal result will be exposing. When this metric’s commands rise up, the result will appear. If the config of a metrics is not well defined, the metrics will be not registered into the main process. If no metrics are registered, the main process will exit with an error status.

## Build from source 
To build from source, use [`promu` tools](https://github.com/prometheus/promu): 
```bash
go get github.com/prometheus/promu
cd $GOPATH/src/github.com/prometheus/promu
make
```

and use this command: 
```bash
go get github.com/orange-cloudfoundry/custom_exporter
cd $GOPATH/src/github.com/orange-cloudfoundry/custom_exporter
$GOPATH/bin/promu build --prefix $GOPATH/bin
```
Note: [a bosh release for cloudfoundry](https://github.com/orange-cloudfoundry/custom_exporter-boshrelease) is available at github

## Configuration 

The configuration is split in 2 separate parts:
 * **credentials**: provide credentials an data type to the custom export.
 * **metrics**: provide commands that are to be run to retrieve metrics and key-value mapping

#### Credential
The credential section is composed at least as:

  * **name**: name of the credential 
  * **type**: collector type (one of existing collector : redis, mysql, bash, ...). If the type is not understand the metrics connected to this credential will be ignored
  
This other options depends of collectors:
| Option Name | Description | Collector |
| :---------: | :---------- | :-------: |
| dsn | the DSN (Data Source Name) is an URL like string usually use to connect to database | mysql, redis | 
| user | the user to run command in shell process | bash |

The DSN form example for each collector: 
    * mysql: driver://user:password@protocol(addr:port|[addr_ip_v6]:port|socket)/database
    * redis: protocol://<empty>:password@host:port/database
 
#### Metric
The metrics section is composed at least as:
  * **name**: name of the metrics
  * **commands**: list of command to run to retrieve the metrics tags and value
  * **credential**: the credential's name to use in this metrics (cannot be null : collector type is include in the credential)
  * **value_type**: the prometheus value type (COUNTER, GAUGE, UNTYPED)
  
This others options are optionals: 

| Option Name | Description | Collector |
| :---------: | :---------- | :-------: |
| mapping | the list of tags to be found in result set | all |
| separator | the separator used in some collector like bash | bash |
| value_name | the name of the metric value key who's be found in result of command | redis |

## Manifest & result examples
### First example
#### Manifest
```yaml
custom_exporter:
  credentials:
  - name: mysql_credential_tcp
    type: mysql 
    dsn: mysql://user:password@tcp(127.0.0.1:3306)/database_name
  - name: mysql_credential_socket
    type: mysql 
    dsn: mysql://user:password@unix(/var/lib/mysql/mysql.sock)/database_name
  - name: shell_credential
    type: bash
    user: root
  - name: redis_credential
    type: redis
    dsn: tcp://:password@127.0.0.1:1234/0
  metrics:
  - name: node_database_size_bytes
    commands:
    - find /var/vcap/store/mysql/ -type d -name cf* -exec du -sb {} ;
    - sed -ne s/^\([0-9]\+\)\t\(\/var\/vcap\/store\/mysql\/\)\(.*\)$/\3 \1/p
    credential: shell_credential
    mapping:
    - database
    separator: ' '
    value_type: UNTYPED
  - name: node_database_provisioning_bytes
    commands:
    - select db_name,max_storage_mb*1024*1024 FROM mysql_broker.service_instances;
    credential: mysql_credential
    mapping:
    - database
    value_type: UNTYPED
  - name: node_redis_info
    commands:
    - INFO REPLICATION
    credential: redis_credential
    mapping:
    - role
    value_name: value
    value_type: UNTYPED
```

#### Results returned in the custom exporter

```bash
[08:53:09] BOSH MySQL ~ # curl -s 10.234.250.202:9100/metrics | grep -i 'node_database'
# HELP node_database_provisioning_bytes Metric read from /var/vcap/jobs/node_exporter/config/database_provisioning.prom
# TYPE node_database_provisioning_bytes untyped
custom_node_database_provisioning_bytes{database="cf_74df5b8f_e7fe_4151_8ec3_741296d42fbc"} 1.048576e+09
custom_node_database_provisioning_bytes{database="cf_d7161ef3_e6fc_4a05_9631_834525f0f7ba"} 1.048576e+09
custom_node_database_provisioning_bytes{database="cf_fa61054d_5c08_4734_a31e_4f2e6065897b"} 1.048576e+08
# HELP node_database_size_bytes Metric read from /var/vcap/jobs/node_exporter/config/database_size.prom
# TYPE node_database_size_bytes untyped
custom_node_database_size_bytes{database="cf_74df5b8f_e7fe_4151_8ec3_741296d42fbc"} 4157
custom_node_database_size_bytes{database="cf_d7161ef3_e6fc_4a05_9631_834525f0f7ba"} 4157
custom_node_database_size_bytes{database="cf_fa61054d_5c08_4734_a31e_4f2e6065897b"} 4157
```

### Another example 
#### Manifest 
```yaml
custom_exporter:
  credentials:
  - name: mysql_connector
    type: mysql ##Possible types are for the moment shell mysql redis
    dsn: mysql://root:password@1.2.3.4:1234/mydb
  metrics:
  - name: custom_metric
    commands:
    - 1
    - 2
    - 3
    credential: mysql_connector
    mapping:
    - tag1
    - tag2
    value_type: UNTYPED
    separator: \t #useless for MySQL but can be usefull for shell
```

#### Result example (MySQL view)
```mysql
|  1 | chicken | 128 |
|  2 | beef | 256 |
|  3 | snails | 14 | 
```

#### Result example (Exporter view)
```bash
custom_metric{tag1="1",tag2="chicken",instance="ip:port",job="custom_exporter"}  128
custom_metric{tag1="2",tag2="beef",instance="ip:port",job="custom_exporter"}  256
custom_metric{tag1="3",tag2="snails",instance="ip:port",job="custom_exporter"}  14
```

## Port binding
According to https://github.com/prometheus/prometheus/wiki/Default-port-allocations we will use TCP/9209

## WIP : Working schema
![custom_exporter_working_schema](custom_exporter.png)

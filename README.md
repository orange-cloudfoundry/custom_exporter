# WIP Prometheus custom_exporter

## Intro

This project is aimed to retrieve specific metrics that can't be found in dedicated exporters. Metrics that will be given by this exporter will be configured in the deployment manifest to make it easy to use and reusable.

## Configuration 

The configuration is split in 2 separate parts :
 * credentials : provide credentials ans data type to the custom export.
 * metrics : provide commands that are to be run to retrieve metrics and key-value mapping

## Manifest & result examples
### First example
#### Manifest
```yaml
custom_exporter:
  credentials:
  - name: mysql_connector
    type: mysql 
    dsn: mysql://monitoring:m0nit0ring4zew1n@127.0.0.1:3306/mysql_broker
  - name: shell_root
    type: bash
    user: root
  metrics:
  - name: node_database_size_bytes
    commands:
    - find /var/vcap/store/mysql/ -type d -name cf* -exec du -sb {} \;| sed -ne 's/^\([0-9]\+\)\t\(\/var\/vcap\/store\/mysql\/\)\(.*\)$/\3 \1/p'
    credential: shell_root
    mapping:
    - database
    separator: ' '
    value_type: UNTYPED
  - name: node_database_provisioning_bytes
    commands:
    - select db_name,max_storage_mb*1024*1024 FROM mysql_broker.service_instances;
    credential: mysql_connector
    mapping:
    - database
    value_type: UNTYPED
```

#### Results returned in the custom exporter

```bash
[08:53:09] BOSH MySQL ~ # curl -s 10.234.250.202:9100/metrics | grep -i 'node_database'
# HELP node_database_provisioning_bytes Metric read from /var/vcap/jobs/node_exporter/config/database_provisioning.prom
# TYPE node_database_provisioning_bytes untyped
node_database_provisioning_bytes{database="cf_74df5b8f_e7fe_4151_8ec3_741296d42fbc"} 1.048576e+09
node_database_provisioning_bytes{database="cf_d7161ef3_e6fc_4a05_9631_834525f0f7ba"} 1.048576e+09
node_database_provisioning_bytes{database="cf_fa61054d_5c08_4734_a31e_4f2e6065897b"} 1.048576e+08
# HELP node_database_size_bytes Metric read from /var/vcap/jobs/node_exporter/config/database_size.prom
# TYPE node_database_size_bytes untyped
node_database_size_bytes{database="cf_74df5b8f_e7fe_4151_8ec3_741296d42fbc"} 4157
node_database_size_bytes{database="cf_d7161ef3_e6fc_4a05_9631_834525f0f7ba"} 4157
node_database_size_bytes{database="cf_fa61054d_5c08_4734_a31e_4f2e6065897b"} 4157
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

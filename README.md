# Logstashbeat

**This fork changed from the original repo in a way that this beat will discover all IPs behind a DNS round robin.**
**Then each IPs will get an API call. To be able to identify API answer, 'ip' and 'domain' fields have been added into the result send to ES.**

**This was made to allow beat to query each logstash instance hidden behind a round robin dns**


Welcome to Logstashbeat.

**Important Notes:** 
This plugin will only work with Logstash 5.0.0 later as the Logstash Monitoring API (listening on port 5600) is only [available since that version](https://www.elastic.co/guide/en/logstash/5.0/alpha1.html).

Ensure that this folder is at the following location:
`${GOPATH}/github.com/consulthys`

## Getting Started with Logstashbeat

### Requirements

* [Golang](https://golang.org/dl/) 1.7.1

### Init Project
To get running with Logstashbeat, run the following command:

```
make init
```

To commit the first version before you modify it, run:

```
make commit
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Logstashbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/consulthys/logstashbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Logstashbeat run the command below. This will generate a binary
in the same directory with the name logstashbeat.

```
make
```


### Run

To run Logstashbeat with debugging output enabled, run:

```
./logstashbeat -c logstashbeat.yml -e -d "*"
```


### Test

To test Logstashbeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`


### Package

To be able to package Logstashbeat the requirements are as follows:

 * [Docker Environment](https://docs.docker.com/engine/installation/) >= 1.10
 * $GOPATH/bin must be part of $PATH: `export PATH=${PATH}:${GOPATH}/bin`

To cross-compile and package Logstashbeat for all supported platforms, run the following commands:

```
cd dev-tools/packer
make deps
make images
make
```

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/logstashbeat.template.json and etc/logstashbeat.asciidoc

```
make update
```


### Cleanup

To clean  Logstashbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Logstashbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/consulthys
cd ${GOPATH}/github.com/consulthys
git clone https://github.com/consulthys/logstashbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

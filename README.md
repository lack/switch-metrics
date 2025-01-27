# switch-metrics

A simple restconf switch metrics gatherer for Prometheus

## Usage

Create your switch authentication file called `switches.yaml` in our
`$XDG_CONFIG_HOME` directory:

```yaml
switches:
  - ip: 10.0.0.1
    username: admin
    password: foo
  - ip: 10.0.0.2
    username: admin
    password: bar
```

Run switch-metrics as a service somewhere, and configure Prometheus to scrape
from the metrics endpoint

## Limitations

* This is in its infancy and will change drastically.
* Only Dell switches are supported so far
* Only some PTP metrics are scraped

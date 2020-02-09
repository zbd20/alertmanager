alerts1='[
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "instance": "example1"
     },
     "annotations": {
        "info": "The disk sda1 is running full",
        "summary": "please check the instance example1"
      }
  },
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "instance": "example1"
     },
     "annotations": {
        "info": "The disk sda2 is running full",
        "summary": "please check the instance example1",
        "runbook": "the following link http://test-url should be clickable"
      }
  },
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "instance": "example2"
     },
     "annotations": {
        "info": "The disk sda1 is running full",
        "summary": "please check the instance example2"
      }
  },
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "instance": "example2"
     },
     "annotations": {
        "info": "The disk sdb2 is running full",
        "summary": "please check the instance example2"
      }
  },
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "command_group": "versionManagerService",
       "severity": "2"
     }
  },
  {
    "labels": {
       "alertname": "DiskRunningFull",
       "group": "SGT",
       "command_group": "versionManagerService",
       "instance": "example3",
       "severity": "1"
     }
  }
]'
curl -XPOST -d"$alerts1" http://localhost:9093/api/v1/alerts
# curl -XPOST -d"$alerts1" http://localhost:9094/api/v1/alerts
# curl -XPOST -d"$alerts1" http://localhost:9095/api/v1/alerts

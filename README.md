# Concourse vRealize Automation Resource

This project is about a new Resource Type in Concourse CI for VMware's product named vRealize Automation (https://www.vmware.com/in/products/vrealize-automation.html)

With this Resource Type, any pipeline from Concourse will be able to trigger a pipeline in vRealize Automation, execute the required tasks, and bring the execution context/output back to Concourse.

More about Concourse Resource Types can be found at https://resource-types.concourse-ci.org/

## Installation

Add following resource type to your Concourse CI pipeline

```yaml
resource_types:
- name: vra
  type: docker-image
  source:
    repository: vmware/concourse-vra-resource
    tag: latest # For reproducible builds, use a specific version tag
```

## Source Configuration
Add following resource to your Concourse CI pipeline

```yaml
resources:
- name: vra-pipeline
  type: vra
  source:
    host: https://www.mgmt.cloud.vmware.com/codestream
    apiToken: ****** # vRealize Automation API/Refresh token
    pipeline: my-vra-pipeline
```

* `host`: *Required.* Code Stream URL of vRealize Automation. For Cloud, use https://www.mgmt.cloud.vmware.com/codestream and for on-prem, provide your instance URL.
* `apiToken`: *Required.* API/Refresh token generated for your account
* `pipeline`: *Required.* vRealize Automation Code Stream pipeline name

## Behavior

### `check`: NA

### `in`: NA

### `out`: Executes vRealize Automation pipeline

```yaml
jobs:
- name: deploy-using-vra
  public: true
  plan:
  - put: vra-pipeline
    params:
      wait: true
      input: # key-value pairs (map).
        key1: val1
        key2: val2
```

#### Parameters

* `wait`: *Required.* Set to true if Concourse pipeline has to wait until vRealize Automation pipeline execution completes. Otherwise set it to false.
* `input`: *Optional.* Input to vRealize Automation pipeline. This param takes key-value pairs and passes them to vRealize Automation pipeline as Input Parameters.


## Examples

```yaml
---
resource_types:
- name: vra
  type: docker-image
  source:
    repository: vmware/concourse-vra-resource
    tag: latest
resources:
- name: vra-pipeline
  type: vra
  source:
    host: https://www.mgmt.cloud.vmware.com/codestream
    apiToken: ****** # vRealize Automation API/Refresh token
    pipeline: my-vra-pipeline
jobs:
- name: deploy-using-vra
  public: true
  plan:
  - put: vra-pipeline
    params:
      wait: true
      input:
        changeset: 5d459e220d7810deb2f62df4a4hd698ce64cf5ff
        developer: vishweshwar

```

## Contributing

The Concourse vRealize Automation Resource project team welcomes contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

For more details about contributing, refer to the [contributing guidelines](https://github.com/vmware/concourse-vrealize-automation-resource/blob/master/LICENSE.txt)

## License

Apache License 2.0, see [LICENSE](https://github.com/vmware/concourse-vrealize-automation-resource/blob/master/LICENSE.txt).

---
layout: "openshift"
page_title: "Openshift: openshift_project"
subcategory: "Project"
description: |-
  Openshift supports multiple virtual clusters backed by the same physical cluster. These virtual clusters are called projects.
---

# openshift_project

Openshift supports multiple virtual clusters backed by the same physical cluster. These virtual clusters are called projects.
Read more about projects at [Openshift reference](https://docs.openshift.com/container-platform/3.11/dev_guide/projects.html)

## Example Usage

```hcl
resource "openshift_project" "example" {
  metadata {
    annotations = {
      "openshift.io/description" = "example-description"
      "openshift.io/display-name" = "example-display-name"
    }

    name = "terraform-example-project"
  }

  lifecycle {
    ignore_changes = [metadata[0].annotations]
  }
}
```

## Argument Reference

The following arguments are supported:

* `metadata` - (Required) Standard project's [metadata](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).

### Timeouts

`openshift_project` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `delete` - Default `5 minutes`

## Nested Blocks

### `metadata`

#### Arguments

* `annotations` - (Optional) An unstructured key value map stored with the project that may be used to store arbitrary metadata. 
**By default, the provider ignores any annotations whose key names end with *openshift.io*. This is necessary because such annotations can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such annotations in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).**
For more info see [Kubernetes reference](http://kubernetes.io/docs/user-guide/annotations)
* `generate_name` - (Optional) Prefix, used by the server, to generate a unique name ONLY IF the `name` field has not been provided. This value will also be combined with a unique suffix. Read more about [name idempotency](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#idempotency).
* `labels` - (Optional) Map of string keys and values that can be used to organize and categorize (scope and select) projects. May match selectors of replication controllers and services. 
**By default, the provider ignores any labels whose key names end with *openshift.io*. This is necessary because such labels can be mutated by server-side components and consequently cause a perpetual diff in the Terraform plan output. If you explicitly specify any such labels in the configuration template then Terraform will consider these as normal resource attributes and manage them as expected (while still avoiding the perpetual diff problem).**
For more info see [Kubernetes reference](http://kubernetes.io/docs/user-guide/labels)
* `name` - (Optional) Name of the project, must be unique. Cannot be updated. For more info see [Kubernetes reference](http://kubernetes.io/docs/user-guide/identifiers#names)

#### Attributes

* `generation` - A sequence number representing a specific generation of the desired state.
* `resource_version` - An opaque value that represents the internal version of this project that can be used by clients to determine when projects have changed. Read more about [concurrency control and consistency](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency).
* `self_link` - A URL representing this project.
* `uid` - The unique in time and space value for this project. For more info see [Kubernetes reference](http://kubernetes.io/docs/user-guide/identifiers#uids)

## Import

projects can be imported using their name, e.g.

```
$ terraform import openshift_project.example terraform-example-project
```

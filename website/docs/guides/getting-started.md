---
layout: "openshift"
page_title: "Openshift: Getting Started with Openshift provider"
subcategory: "Guide"
description: |-
  This guide focuses on scheduling Openshift resources like Deployment Configs,
  Build Configs, Image Streams etc. on top of a properly configured
  and running Openshift cluster.
---

# Getting Started with Openshift provider

## Openshift

[Openshift](https://docs.openshift.com/container-platform/3.11/welcome/index.html) is an open-source workload scheduler 
with focus on containerized applications.

There are at least 2 steps involved in scheduling your first container
on a Openshift cluster. You need the Openshift cluster with all its components
running _somewhere_ and then schedule the Openshift resources, Deployment Configs,
Build Configs, Image Streams etc.

This guide focuses mainly on the latter part and expects you to have
a properly configured & running Openshift cluster.

## Why Terraform?

While you could use `oc` or similar CLI-based tools mapped to API calls
to manage all Openshift resources described in YAML files,
orchestration with Terraform presents a few benefits.

 - Use the same [configuration language](/docs/configuration/syntax.html)
    to provision the Openshift infrastructure and to deploy applications into it.
 - drift detection - `terraform plan` will always present you the difference
    between reality at a given time and config you intend to apply.
 - full lifecycle management - Terraform doesn't just initially create resources,
    but offers a single command for creation, update, and deletion of tracked
    resources without needing to inspect the API to identify those resources.
 - synchronous feedback - While asynchronous behaviour is often useful,
    sometimes it's counter-productive as the job of identifying operation result
    (failures or details of created resource) is left to the user. e.g. you don't
    have IP/hostname of load balancer until it has finished provisioning,
    hence you can't create any DNS record pointing to it.
 - [graph of relationships](https://www.terraform.io/docs/internals/graph.html) -
    Terraform understands relationships between resources which may help
    in scheduling - e.g. if a Persistent Volume Claim claims space from
    a particular Persistent Volume Terraform won't even attempt to create
    the PVC if creation of the PV has failed.

## Provider Setup

The easiest way to configure the provider is by creating/generating a config
in a default location (`~/.kube/config`). That allows you to leave the
provider block completely empty.

```hcl
provider "openshift" {}
```

If you wish to configure the provider statically you can do so by providing TLS certificates:

```hcl
provider "openshift" {
  host = "https://104.196.242.174"

  client_certificate     = file("~/.kube/client-cert.pem")
  client_key             = file("~/.kube/client-key.pem")
  cluster_ca_certificate = file("~/.kube/cluster-ca-cert.pem")
}
```

or by providing username and password (HTTP Basic Authorization):

```hcl
provider "openshift" {
  host = "https://104.196.242.174"

  username = "ClusterMaster"
  password = "MindTheGap"
}
```

After specifying the provider we may now run the following command
to copy the Openshift provider's binary to Terraform's plugins folder.

```
$ cp terraform-provider-openshift ~/.terraform.d/plugins/
```

## Conclusion

Terraform offers you an effective way to manage both compute for
your Openshift cluster and Openshift resources. Check out
the extensive documentation of the Openshift provider linked
from the menu.

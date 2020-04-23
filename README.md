Terraform Provider for Openshift
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)
<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">


Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.12+

Building the provider
---------------------

Clone repository to: `$GOPATH/src/github.com/llomgui/terraform-provider-openshift`

```sh
$ mkdir -p $GOPATH/src/github.com/llomgui; cd $GOPATH/src/github.com/llomgui
$ git clone git@github.com:llomgui/terraform-provider-openshift
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/llomgui/terraform-provider-openshift
$ make build
```

Developing the provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org)
installed on your machine (version 1.14+ is *required*). You can use [goenv](https://github.com/syndbg/goenv)
to manage your Go version. You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH),
as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`.
This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-openshift
...
```

Using the provider
------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory,  run `terraform init` to initialize it.

Provider Documents
--------------
Currently the documents for this provider is not hosted by the official site [Terraform Providers](https://www.terraform.io/docs/providers/index.html). Please enter the provider directory and build the website locally.

```sh
$ make website
```

The commands above will start a docker-based web server powered by [Middleman](https://middlemanapp.com/), which hosts the documents in `website` directory. Simply open `http://localhost:4567/docs/providers/openshift` and enjoy them.
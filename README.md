# `uc` is an uber Kubernetes update client.

[![CircleCI](https://circleci.com/gh/chirino/uc.svg?style=svg)](https://circleci.com/gh/chirino/uc)

It automatically downloads and updates Kubernets CLI clients so your using the the version that
best works with cluster your connected to.  You'll never need to download or update your kubernets 
related CLI client tools again.

## Installing

Browse the [releases page](https://github.com/chirino/uc/releases), extract the appropriate executable
for your platform, and install it to your `PATH`.

# Usage

Invoke `uc` with 

    $ uc help
    uc is an uber client that automatically installs keeps updated Kubernetes and 
    OpenShift related command line tools at versions that are best suited to operate 
    against the cluster that you are connected to.
    
    Usage:
      uc [command]
    
    Examples:
    
      uc kubectl get pods
      uc oc new-project sandbox1
      uc kamel run examples/dns.js
    
    Available Commands:
      help        Help about any command
      kamel       Manage your Apache Camel K integrations
      kn          Manage your Knative building blocks
      kubectl     Controls the Kubernetes cluster manager
      oc          OpenShift Client
      odo         Developer-focused CLI for OpenShift
    
    Flags:
          --cache-expires string   Controls when the catalog and command caches expire. One of *duration*|never|now (default "24h")
      -h, --help                   help for uc
          --kubeconfig string      path to the Kubeconfig file (default "/Users/chirino/.kube/config")
          --master string          Master url
      -v, --verbosity string       Sets the verbosity level: One of none|info|debug (default "info")
    
    Use "uc [command] --help" for more information about a command.

## How it works

The uc command uses an online catalog to discover all the supported sub commands.  It then uses
information found in the catalog to download and verify the sub command executable has not been
tampered with by checking it against a gpg signature.  Both the catalog and sub command executables
are stored in `$HOME/.uc/cache` or `%USERPROFILE%\.uc\cache`.  This command does not delete data 
from the cache.  It is safe to delete this cache directory to reclaim disk space.

The catalog cache expire every 24 hours by default and will be fetched against once expired.  You can use
the `--cache-expires string` flag to control how often catalog updates are fetched.

The version of the sub command executable use can vary based on the cluster your connected to.  For example,
the version for `kubectl` selected will match match the Kubernetes server version your connected against.  For 
other sub commands which we do not have a good way of selecting the best version based on the cluster
state, we will used the latest version released.

## Building From Source

Requires [Go 1.12+](https://golang.org/dl/).  To fetch the latest sources and install into your system:

    go get -u github.com/chirino/uc


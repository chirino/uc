# `uc` is an uber Kubernetes update client.

It automatically downloads and updates Kubernets CLI clients so your using the the version that
best works with cluster your connected to.  You'll never need to download or update your kubernets 
related CLI client tools again.

## Installation

requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).

    go get -u github.com/chirino/uc

To find out where `uc` was installed you can run `go list -f {{.Target}} github.com/chirino/uc`. 
For `uc` to be used globally add that directory to the `$PATH` environment setting.

# Usage

Invoke `uc` with 

    $ uc help
    uber client runs sub commands using clients at the version that are compatible with the cluster your logged into.
    
    Usage:
      main [flags]
      main [command]
    
    Available Commands:
      help           Help about any command
      kamel          Manage your Apache Camel K integrations
      kn             Manage your Knative building blocks
      kubectl        Controls the Kubernetes cluster manager
      oc             OpenShift Client
      odo            Developer-focused CLI for OpenShift
      update-catalog Updates and GPG signs the local uc catalog (only available when built with --tags dev)
    
    Flags:
      -h, --help                help for main
          --kubeconfig string   path to the kubeconfig file (default "/Users/chirino/.kube/config")
          --master string       master url
    
    Use "main [command] --help" for more information about a command.

## How it works

Supported uc sub commands are looked up once a day from the catalog stored at: 
https://github.com/chirino/uc/tree/master/docs 

The catalog keeps track of all the versions released for the sub command, the URLs where 
release was done to, how to extract the client binary from the released and a digital signature
that we expect the client binary to match.  If the the client binary has not yet been downloaded
or if the digital signature does not match, we try to download it to the `~/.uc/cache/commands`
directory.  The binary is then executed passing along all sub command arguments to it.

The version of the client binary used will be the latest known in the catalog unless we can
determine a more appropriate versions to use.  For example, the version for `kubectl` selected
match the Kubernetes server version. 
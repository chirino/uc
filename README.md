# The `UC` Catalog Site

`https://chirino.github.io/uc/catalog/v1` is the primary catalog API URL used by the `uc`command.  This page is used to document it's API resources.  You can verify the sign digital signatures for resources signed by the API with the public key in listing #1:


**Listing #1: The UC GNU PGP Key**

```
-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - http://gpgtools.org
mQGiBEPspSsRBADdguKAxMQbA32vTQrCyONR6Zs/YGdvau2Zrr3SSSSR0Ge4FMjZ
4tzwpf6+32m4Bsf7YIwdLl0H5hI1CgT5gDl9kXvfaFUehFnwR+FDyiBRiyHjUpGF
4dgkQfWy9diYeWGtsvszsvWHXtED4SXb322StX4MfJj+YesA1iEdTiXK6wCg1QDa
RucfjC+kx4zPsJwkJOgYpyMEAMTiXtNwQcke6nIFb/lb5374NjwwVAuuMTrRWLyq
5HodugEIHaw3EitQWtnFfXNkXTJZzS6t2HAGv29UTfhiBzKdkydgCkOk2MLWISOV
fqcg0tNIp5ZJCmUHg3s+OFNSH4oUi65u+FyDseUid3OKtPI+ZhIk8N+DjOIg2Kvo
/UALA/9q+WfBd7re+W3iUtU7TutUcwbKsjP+jpaJeUHg2ChOBxVfQKt4YlPHVdrR
iCrfNi90Z8qbsZ0iAXuqexrfMq20pAPmpHRpe54mmP1CMT5m+Gq71eKIfkUrb3LC
/zv08dLG2vm9oghd242wbcifaX+t7AhNAIpe/WTvQsB0gpdO4LQmSGlyYW0gQ2hp
cmlubyA8aGlyYW1AaGlyYW1jaGlyaW5vLmNvbT6IWwQTEQIAGwUCQ+ylKwYLCQgH
AwIDFQIDAxYCAQIeAQIXgAAKCRCf8lmA9bp+T/G/AKDM1QDs7il/CJhTycgDvE3c
EOgUBwCfelsVK4sgBCooZptoaCCDgVtt71GIRgQQEQIABgUCRO3MrwAKCRDs3+o8
tEk7lPoGAJ4qoY6sQPRCmVAvygftCnkHzOsc/gCeLoG4wCuTSnH1EjJoPdMHya0e
udGIWwQTEQIAGwUCQ+ylKwYLCQgHAwIDFQIDAxYCAQIeAQIXgAAKCRCf8lmA9bp+
T/G/AKCM+2vI3pYagtmNxdamMgJZ/AWIeQCff0OpzpKQNf5P0Hn+wVCzW2YbRai5
AQ0EQ+ylLhAEAJD25AWgwcNgBFKYsvExQaGIojIGJyn4Cf/5U30cui/K7fIU7Jty
NhKcfZdCrh2hKx+x3H/dTF6e0SrRhzKV7Dx0j76yhHHB1Ak25kjRxoU4Jk+CG0m+
bRNTF9xz9k1ALSm3Y+A5RqNU10K6e/5KsPuXMGSGoQgJ1H6g/i80Wf8PAAMFA/9m
Ixu7lMaqE1OE7EeAsHgLslNbi0h9pjDUVNv8bc1Os2gBPaJD8B89EeheTHw6NMNI
e75HVOpKk4UA0gvOBrxJqCr18yFJBM5sIlaEmuJwZOW4dDGOR1oS5qgE9NzpmyKh
E+fu/S1wmy0coL667+1xZcnrPbUFD4i7/aD1r8qJhohGBBgRAgAGBQJD7KUuAAoJ
EJ/yWYD1un5Pth0An0QEUs5cxpl8zL5kZCj7c8MN8YZDAKDR9LTb6woveul50+uG
tUl2fIH1uA==
=7BPT
-----END PGP PUBLIC KEY BLOCK-----
```

## GET `/index.yaml`

This is the catalog index resource used discover all the sub commands `uc` can support.  

Each command int the yaml resource can define the following fields:

| Field                | Description                                           |
| -------------------- | ------------------------------------------------------|
| `short-description`  | A short description of the command.
| `long-description`   | A long description of the command.                    |
| `catalog-base-url`   | The URL base to use to load catalog data for this command. If not set, the primary catalog base URL will be used. |
| `catalog-public-key` | An armored GNU PGP public key that shold be used to validate digital signatures for the catalog and executable files assocaited with the command. If not set the default public key will be used. |


**Example [`/index.yaml`](https://chirino.github.io/uc/catalog/v1/index.yaml):**
```yaml
commands:
  kubectl:
    short-description: Controls the Kubernetes cluster manager
  oc:
    short-description: OpenShift Client  
.....
```

### GET `/index.yaml.sig`

The digital signature for the `/index.yaml` resource.

## GET `/{:sub-command}.yaml`

When a `uc`sub command is executed, it will load sub command index resource to find the versions available for the subcommand.

**Example [`/kubectl.yaml`](https://chirino.github.io/uc/catalog/v1/kubectl.yaml):**
```yaml
latest: v1.15.1
versions:
  - "v1.15.1"
  - "v1.14.0"
  - "v1.13.0"
....  
```

### GET `/{:sub-command}.yaml.sig`

The digital signature for the `/{:sub-command}.yaml` resource.

## GET `/{:sub-command}/{:verison}.yaml`

This resource yaml resource contains the sub command release information. Once `uc` selects a version of sub command to to use, it gets this resource find find out where to download, what to extract form the download and the digitial signature of the sub command. 

It should contain more or more `$GOOS-$GOARCH` platforms entries for each platform that the sub command was built against.  See the [go enviornment docs](https://golang.org/doc/install/source#environment) for a valid list of values for `$GOOS` and `$GOARCH`.  Eeach platform can define the following fields:

| Field         | Description                                                                                     |
| ------------- | ------------------------------------------------------------------------------------------------|
| `url`         | The URL from which to download the sub command                                                  |
| `size`        | The expected size of the sub command exectuable once downloaded and extracted.                  |
| `signature`   | The digital signature of the  sub command exectuable once downloaded and extracted.             |
| `extract-tgz` | If the URL is a tar gz file, set this field to the path in the archive of the executable.       |
| `extract-zip` | If the URL is a zip file, set this field to the path in the archive of the executable.          |
| `uncompress`  | If the URL is a gz compressed file, set this field to 'gz'                                      |

**Example [`/kubectl/v1.15.1.yaml`](https://chirino.github.io/uc/catalog/v1/kubectl/v1.15.1.yaml):**
```yaml
darwin-amd64:
  extract-tgz: kubernetes/client/bin/kubectl
  signature: iF0EABECAB0WIQTluCR6+KYZoo+Q/fyf8lmA9bp+TwUCXUZSiwAKCRCf8lmA9bp+TzJYAJ40aN4f7SubrkJeYNwXYbSeE0Z4AwCbB3CrXO1zhhH5ICLdKBGVPM3Chhg=
  size: 48591648
  url: https://dl.k8s.io/v1.15.1/kubernetes-client-darwin-amd64.tar.gz
linux-amd64:
  extract-tgz: kubernetes/client/bin/kubectl
  signature: iF0EABECAB0WIQTluCR6+KYZoo+Q/fyf8lmA9bp+TwUCXUZSjQAKCRCf8lmA9bp+T6RnAKCtyQhiBhnCsgxSl/ZlZm2Cc1XDxgCgglHSEGjBer/5t64Rhz5fmph394s=
  size: 42989600
  url: https://dl.k8s.io/v1.15.1/kubernetes-client-linux-amd64.tar.gz
windows-amd64:
  extract-tgz: kubernetes/client/bin/kubectl.exe
  signature: iF0EABECAB0WIQTluCR6+KYZoo+Q/fyf8lmA9bp+TwUCXUZSiQAKCRCf8lmA9bp+TybFAKC4/vhJWvP8yIJ+zfCnKCp/xRUHUQCghqV1H/9YhvFuHgHz1W6HsU4Qs90=
  size: 43471360
  url: https://dl.k8s.io/v1.15.1/kubernetes-client-windows-amd64.tar.gz
....  
```

### GET `/{:sub-command}/{:verison}.yaml.sig`

The digital signature for the `/{:sub-command}/{:verison}.yaml` resource.


<style>
h1 code, h2 code, h3 code, h4 code, h5 code, h6 code { font-size: inherit; }
table { width: 100%; }
</style>



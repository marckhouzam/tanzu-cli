# Test CLI-Svc

This mock allows to run an nginx docker image which serves the same endpoints as
the ones the real CLI-Svc serves.  It allows to test the CLI without using the
real CLI-Svc.

The endpoints being mocked use SSL and are:

- localhost:9443/cli/v1/install
- localhost:9443/cli/v1/plugin/discovery
- localhost:9443/cli/v1/binary

## Using the test CLI-Svc

From the root directory of the tanzu-cli repo, run `make start-test-cli-service`.
This will start the test CLI-Svc as a configured nginx docker image.

To access the endpoints manually, e.g.,:

```console
# NOTE: the trailing / is essential
curl https://localhost:9443/cli/v1/plugin/discovery/ --cacert hack/central-repo/certs/localhost.crt
# or
curl https://localhost:9443/cli/v1/plugin/discovery/ -k
```

## Testing plugin discovery

If testing plugin discovery (localhost:9443/cli/v1/plugin/discovery/), the
test CLI-Svc will randomly serve different discovery data which is configured in
`hack/service/cli-service.conf`.

To tell the CLI to use the test CLI-Svc we must execute:

```console
export TANZU_CLI_PLUGIN_DISCOVERY_HOST_FOR_TANZU_CONTEXT=http://localhost:9443
```

To allow testing using different central repositories the endpoint serves some
discoveries using both:

- projects.packages.broadcom.com/tanzu_cli/plugins/plugin-inventory:latest
- localhost:9876/tanzu-cli/plugins/central:small

Therefore, it is required to also start the test central repo by doing:

```console
export TANZU_CLI_PLUGIN_DISCOVERY_IMAGE_SIGNATURE_VERIFICATION_SKIP_LIST=localhost:9876/tanzu-cli/plugins/central:small
make start-test-central-repo

# If the certificates should be handled automatically by the CLI itself
# we should remove their configuration done by the make file so that the
# testing is more appropriate.  This would be done by doing:
tanzu config cert delete localhost:9876
```

# Passthrough for the plugin discovery

The test CLI-Svc can also serves as a passthrough to the OCI plugin discovery.
The idea is to have a well-known address for the CLI to access the plugin discovery
as well as allowing for end-user machines not to have direct access to the OCI
registry itself.

In this test environment we achieve this by running the test CLI-Svc
which is already configured to passthrough to the test central repo:


                 localhost:9443/tanzu-cli/plugins/plugin-inventory:latest
                               ^
/----------\     /---------\   |
| Test OCI |     |  Test   |   |  /-----\
| Registry | <-- | CLI-Svc | <--- | CLI |
\----------/     \  nginx  /      \-----/
                   -------

```console
export TANZU_CLI_PLUGIN_DISCOVERY_IMAGE_SIGNATURE_VERIFICATION_SKIP_LIST=localhost:9443/tanzu-cli/plugins/central:small
make start-test-central-repo
make start-test-cli-service

# Use the now well-known plugin source, no matter where the real OCI registry resides
tanzu plugin source update default -u localhost:9443/tanzu-cli/plugins/central:small
```

To achieve the above, the test CLI-Svc also serves `localhost:9443/v2/.*` which is the endpoint
needed to access the OCI registry.

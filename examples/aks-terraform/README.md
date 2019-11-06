# AKS via Terraform Bundle

Create an AKS cluster via Terraform.

## Prerequisites
To run this bundle, an Azure Service Principal will need to be created with `Contributer`-level access to the subscription being used.  For example:
```
az ad sp create-for-rbac -n "MySP" --role contributor --scopes "/subscriptions/<subscriptionId>"
```
Note the `appId`, `password` and `tenant` values presented in the output and add these, as well as other pertinent values to your environment:
```
export SUBSCRIPTION_ID=<subscriptionId>
export CLIENT_ID=<appId>
export CLIENT_SECRET=<password>
export TENANT_ID=<tenant>
export SSH_KEY="$(cat ${HOME}/.ssh/duffle/id_rsa.pub)"
```

## Credentials
The credentials in your environment need to be supplied to this bundle.  First, we will generate a credentials file:
```
duffle creds generate my_azure_creds -f bundle.json
```

This will create a file in `~/.duffle/credentials/my_azure_creds.yaml` showing the credentials needed, albeit with empty values.

We can either fill the values in directly, or change the `value: EMPTY` sections under `source:` to read `env: ENV_VAR_NAME` so that the values will be pulled from our environment.
Let's do the latter by editing the file:
```
$EDITOR ~/.duffle/credentials/my_azure_creds.yaml
```

It should then look something like:
```
name: my_azure_creds
credentials:
  - name: subscription_id
    source:
      env: SUBSCRIPTION_ID
    destination:
      env: TF_VAR_subscription_id
...
```

# Parameters

The default parameters set in this bundle are as follows.

## Remote Terraform Backend

Relevant to the Terraform backend, using an Azure Storage Container:
  * `backend_storage_account`
  * `backend_storage_container`
  * `backend_storage_resource_group`

These resources will be automatically provisioned if they do not already exist.

**Note:** Azure Storage Account names must be globally unique.  Therefore, the `backend_storage_account` value will need to be edited to differ from the default. The default can be changed in the `bundle.json` file before running `duffle install`, or be passed via an override on the command line, like so:
```
duffle install aks-tf -c my_azure_creds -f bundle.json -s backend_storage_account=myuniqstorageaccount
```
(Also to be noted: Azure Storage Account names must only consist of lowercase alphanumeric characters, limited to a length of 24.)

## AKS Cluster

Relevant to the AKS cluster being created:
  * `location`
  * `kubernetes_version`
  * `agent_count`
  * `dns_prefix`
  * `cluster_name`
  * `resource_group_name`

All of these have defaults, and shouldn't be necessary to override.

See `bundle.json` for what the default configuration looks like.

# AKS cluster life cycle via Duffle

To create the AKS cluster (via Terraform):

```
duffle install aks-tf -c my_azure_creds -f bundle.json -s backend_storage_account=myuniqstorageaccount
```

To upgrade the AKS cluster (via Terraform):

```
duffle upgrade aks-tf -c my_azure_creds
```

To see the status of the Duffle claim as well as the AKS cluster itself (via Terraform):

```
duffle status aks-tf -c my_azure_creds
```

To tear down the AKS cluster (via Terraform):

```
duffle uninstall aks-tf -c my_azure_creds
```

# Access the AKS cluster

Once the AKS cluster is created, assets (such as kubeconfig, etc.) will exist in the output.  These can be captured and saved locally to a file for subsequent `kubectl` commands.  Alternatively, one can get the credentials via `az`:

```
touch my_aks_kubeconfig
az aks get-credentials -n akstest -g azure-akstest -f my_aks_kubeconfig
KUBECONFIG=my_aks_kubeconfig cluster-info
```
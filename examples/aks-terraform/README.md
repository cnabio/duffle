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

## Build
First, build the bundle with `duffle`:

```console
duffle build
```

## Credentials
The credentials in your environment need to be supplied to this bundle.  First, we will generate a credentials file:

```
duffle creds generate my_azure_creds aks-terraform:0.1.0
```

Follow the prompts to enter credential sources using the environment variable names exported above:

```
? Choose a source for "client_id" environment variable
? Enter a value for "client_id" CLIENT_ID
? Choose a source for "client_secret" environment variable
? Enter a value for "client_secret" CLIENT_SECRET
? Choose a source for "ssh_authorized_key" environment variable
? Enter a value for "ssh_authorized_key" SSH_KEY
? Choose a source for "subscription_id" environment variable
? Enter a value for "subscription_id" SUBSCRIPTION_ID
? Choose a source for "tenant_id" environment variable
? Enter a value for "tenant_id" TENANT_ID
```

This file will be saved to `~/.duffle/credentials/my_azure_creds.yaml`.

# Parameters

The default parameters set in this bundle are as follows.

## Remote Terraform Backend

Relevant to the Terraform backend, using an Azure Storage Container:
  * `backend_storage_account`
  * `backend_storage_container`
  * `backend_storage_resource_group`

These resources will be automatically provisioned if they do not already exist.

**Note:** Azure Storage Account names must be globally unique.  Therefore, the `backend_storage_account` value will need to be edited to differ from the default. The default can be changed in the `duffle.json` file before running `duffle install`, or be passed via an override on the command line, like so:

```
duffle install aks-tf -c my_azure_creds aks-terraform:0.1.0 -s backend_storage_account=myuniqstorageaccount
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

See `duffle.json` for what the default configuration looks like.

# AKS cluster life cycle via Duffle

To create the AKS cluster (via Terraform):

```
duffle install aks-tf -c my_azure_creds aks-terraform:0.1.0 -s backend_storage_account=myuniqstorageaccount
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
KUBECONFIG=my_aks_kubeconfig kubectl cluster-info
```
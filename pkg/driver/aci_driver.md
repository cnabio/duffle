# Duffle - ACI Driver

The ACI Driver for Duffle enables the *installation* of CNAB Bundle using [Azure Container Instance](https://azure.microsoft.com/en-gb/services/container-instances/) as an installation driver, this enables installation of a CNAB bundle from environments where using the Docker driver is impossible (e.g. [Azure CloudShell](https://azure.microsoft.com/en-gb/features/cloud-shell/)). You must have an [Azure account](https://azure.microsoft.com/free/) to use this driver.

## Getting Started

The easiest way to get started is to use [Azure Cloud Shell](https://shell.azure.com). 

1. [Get the latest Duffle release for Linux](https://github.com/deislabs/duffle/releases).

```console
curl https://github.com/deislabs/duffle/releases/download/<latest-release>/duffle-linux-amd64 -L -o duffle
mv duffle $HOME/bin/duffle
chmod +x $HOME/bin/duffle
```

2. Run the command duffle init to setup duffle:
    
```console
$USER@Azure:~$ duffle init
==> The following new directories will be created:
/home/$USER/.duffle
/home/$USER/.duffle/bundles
/home/$USER/.duffle/logs
/home/$USER/.duffle/plugins
/home/$USER/.duffle/claims
/home/$USER/.duffle/credentials
==> The following new files will be created:
/home/$USER/.duffle/repositories.json
==> Generating a new secret keyring at /home/$USER/.duffle/secret.ring
==> Generating a new signing key with ID $USER <$USER@computer>
==> Generating a new public keyring at /home/$USER/.duffle/public.ring
```

3. Install a sample bundle:

A simple helloworld-aci bundle can be found [here](https://github.com/deislabs/duffle/tree/acidriver/examples/helloworld-aci). 

The sample bundle can be imported either with or without verification, to run import the bundle with verification first you need to import the public key that the bundle was signed with:

```console
curl https://raw.githubusercontent.com/deislabs/duffle/master/examples/helloworld-aci/helloworld-aci.gpg -L -o /tmp/helloworld-aci.gpg
duffle key import /tmp/helloworld-aci.gpg
```

Then the sample bundle can be imported:

```console
curl https://raw.githubusercontent.com/deislabs/duffle/master/examples/helloworld-aci/helloworld-aci.tgz -L -o /tmp/helloworld-aci.tgz
duffle import /tmp/helloworld-aci.tgz -d ~/.duffle/bundles
```

To import the sample bundle without verification run the import command with the -k option:

```console
duffle import /tmp/helloworld-aci.tgz -k -d ~/.duffle/bundles
```

4. Install the sample bundle using the aci-driver:

The aci-driver requires at a minimum one environment variable 'ACI_LOCATION' to be set to specify the Azure location where the ACI Instance will be created, in this configuration an Azure Resource Group will be automatically created and the Resource Group and the ACI Container Group will be deleted once the installation is complete. 

Install the bundle with verification:

```console
export ACI_LOCATION=westeurope
duffle install <installation name> helloworld-aci -d aci
```

`ACI_LOCATION` can be set to any region [where the ACI Service is available](https://azure.microsoft.com/en-us/global-infrastructure/services/?products=container-instances&regions=all)

By default in CloudShell the ACI Driver will use the credentials that you are logged into CloudShell with to create the ACI Container Group to run the invocation image, if you have more than one subscription it will pick the first subscription that is returned from listing all subscriptions available to the account, a specific subscription can be chosen by setting the environment Variable `AZURE_SUBSCRIPTION_ID`to the subscription ID to be used. More details on alternative authentication approaches are specified below. 

To install the bundle without verification add the -k flag to the commmand:

```console
duffle install <installation name> helloworld-aci -d aci -k
```

Run the bundle with a System Assigned MSI:

```console
export ACI_LOCATION=westeurope
export ACI_MSI_TYPE=user
duffle install <installation name> helloworld-aci -d aci -s aci_msi_type=${ACI_MSI_TYPE}
```

## Logging and tracing

To enable trace logs of the ACI Driver set the environment variable VERBOSE to `true`, to get logs of the HTTP requests sent to Azure set the environment variable `AZURE_GO_SDK_LOG_LEVEL` to `INFO`, to get the request and response bodies set the value to `DEBUG`

## Authentication to Azure

The ACI Driver can Authenticate to Azure using four different mechanisms and will evaluate them in this order:

1. Service Principal

Setting the environment variables `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET` and `AZURE_TENANT_ID` will cause the driver to attempt to login using those credentials as a service principal. More details on how to create a service prinicpal for authentication can be found [here](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest)

2. Device Code Flow

Setting the environment variables `AZURE_APP_ID` and `AZURE_TENANT_ID` will cause the driver to use the [Azure Device Code flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-device-code), you need to set the AZURE_APP_ID variable to the applicationId of a native application that has been registered with Azure AAD and has access to read user profiles from Azure Graph. To register  an application see the documentation [here](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app) and [here](https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-configure-app-access-web-apis)

3. MSI

If the driver is running in an environment where MSI is available (such as in a VM in Azure) , it will attempt to login using MSI, no configuration is necessary, the driver will detect the MSI endpoint and use it if it is availabel.

4. CloudShell 

If the driver is running in Azure CloudShell it will automatically login using the logged in users token if no environment variables are set.

## ACI Container Group Identity

By default the ACI Container Group that is created to run the invocation image has no identity, to perform authenticated actions against resources credentials need to be presented to the invocation image. It is possible to have the ACI Container Group that executes the invocation image use [Managed Service Identity(MSI)](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview) which allows the invocation image to be able to access the token for this identity and use it alongside any credentials passed to the bundle. The driver supports both System Assigned and User Assigned MSI, to use system assigned MSI set the environment variable `ACI_MSI_TYPE` to `system`. If no other environment variables are set the MSI will be assigned the Contributor role at the scope of the Resource Group that the ACI Container Group is created in, to override this behaviour the environment variable `ACI_SYSTEM_MSI_ROLE` can be set to the role required and `ACI_SYSTEM_MSI_SCOPE` can be set to set the scope fo the role. Note that as the MSI is created in parallel to the ACI Container Group to prevent a race condition between an code in the bundle that relies on permissions being allocated to the System Assigned MSI the Container Group is created initially using an alpine image to allow for the system assigned MSI to be created and permissions assigned, once this is done the invocation image is launched. To use User Assigned MSI `ACI_MSI_TYPE` should be set to `user` and environment variable `ACI_USER_MSI_RESOURCE_ID` should be set to the Resource Id of the User Assigned MSI. 

## Resource Group Usage

The ACI Driver will create a new resource group in the region specified by the environment variable 'ACI_LOCATION' if the environment variable `ACI_RESOURCE_GROUP` is not set or if the resource group specified in `ACI_RESOURCE_GROUP` does not exist, the name will be auto-generated. If `ACI_RESOURCE_GROUP` is not set. If the resource group specified in `ACI_RESOURCE_GROUP` already exists that will be used, this resource group will be used to create the ACI Container Group to run the invocation image, the invocation image is also free to create resources in this resource group if it can acquire the correct permissions.

## ACI Container Group and Container Instance Naming

The container group and the container instance use the value of the environment variable `` for their names, if this is not set a name is generated, it is not possible to use different names for the group and the container, they will always use the same name.

## Deleting Resources

By default the driver will delete the container group that it creates and also the resource group if it creates it (pre-existing resource groups are not deleted), this behaviour can be changed by setting the environment variable `ACI_DO_NOT_DELETE` to true. This can be useful for debugging or if you know that the invocation image is going to create resources in the same resource group.

## Environment Variables

|  Environment Variable 	| Description  	|
|---	|---	|
| VERBOSE  	| Verbose output - set to true to enable  	|
| AZURE_CLIENT_ID  	| AAD Client ID for Azure account authentication - used to authenticate to Azure using Service Principal for ACI creation  	|
| AZURE_CLIENT_SECRET  	|  AAD Client Secret for Azure account authentication - used to authenticate to Azure using Service Principal for ACI creation 	|
| AZURE_TENANT_ID  	|  Azure AAD Tenant Id Azure account authentication - used to authenticate to Azure using Service Principal or Device Code for ACI creation 	|
| AZURE_APP_ID  	|  Azure Application Id - this is the application to be used when authenticating to Azure using device flow	|
| AZURE_SUBSCRIPTION_ID    | Azure Subscription Id - this is the subscription to be used for ACI creation, if not specified the first (random) subscription is used    | 
| ACI_RESOURCE_GROUP  	|   The name of the existing Resource Group to create the ACI instance in, if not specified a Resource Group will be created, if specfied and it does not exist a new resource group with this name will be created	|
| ACI_LOCATION  	|   The location in which to create the ACI Container Group and Resource Group	|
| ACI_NAME  	|   The name of the ACI instance to create - if not specified a name will be generated	|
| ACI_DO_NOT_DELETE  	|  Set to true so as not to delete the RG and ACI container group created - useful for debugging - only deletes RG if it was created by the driver 	|
| ACI_MSI_TYPE  	|   If this is set to user or system the created ACI Container Group will be launched with MSI	|
| ACI_SYSTEM_MSI_ROLE  	|  The role to be asssigned to System MSI User - used if ACI_MSI_TYPE == system, if this is null or empty then the role defaults to Contributor 	|
| ACI_SYSTEM_MSI_SCOPE  	|  The scope to apply the role to System MSI User - will attempt to set scope to the  Resource Group that the ACI Instance is being created in if not set 	|
| ACI_USER_MSI_RESOURCE_ID  	|  The resource Id of the MSI User - required if ACI_MSI_TYPE == user 	|
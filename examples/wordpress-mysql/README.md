# Kubernetes Wordpress + Azure MySQL Bundle
### "Because you need a database and probably don't want to manage it yourself"

This bundle provisions an Azure MySQL instance, creates a database, installs Wordpress on a given AKS cluster and connects the wordpress application to the newly provisioned Azure MySQL instance.

## Prerequisites
- Running Kubernetes cluster (we used AKS))
- Log into azure-cli locally to put your Azure credentials in place

## Install this bundle
1. Build the bundle:
    ```console
    $ duffle build
    ```
1. Generate credentials via the interactive prompts:
    ```console
    $ duffle credentials generate wordpress-mysql-creds wordpress-mysql:0.2.0
    ? Choose a source for "azure_profile" file path
    ? Enter a value for "azure_profile" ${HOME}/.azure/azureProfile.json
    ? Choose a source for "azure_tokens" file path
    ? Enter a value for "azure_tokens" ${HOME}/.azure/accessTokens.json
    ? Choose a source for "mysql_password" specific value
    ? Enter a value for "mysql_password" secret
    ? Choose a source for "mysql_user" specific value
    ? Enter a value for "mysql_user" mysql-user
    ```
1. Use the `duffle install` command by passing in a name (`wordpress-mysql`), the bundle name (`wordpess-mysql:0.2.0`) and the credential set generated above (`-c wordpress-mysql-creds`):
    ```console
    $ duffle install wordpress-mysql wordpress-mysql:0.2.0 -c wordpress-mysql-creds
    ```

> Note: you can override values such as the AKS cluster_name, by adding `--set cluster_name=your-aks-name` to the duffle install command.

This can take a while to complete. Check for the Wordpress pod running in your cluster and view the logs. Wait for this pod to be running and ready.

## Log in to Wordpress

Use the commands below to find your Wordpress URL and credentials and test: 

```
export SERVICE_IP=$(kubectl get svc --namespace default wordpress-mysql --template "{{ range (index .status.loadBalancer.ingress 0) }}{{.}}{{ end }}")

echo "WordPress URL: http://$SERVICE_IP/"
echo "WordPress Admin URL: http://$SERVICE_IP/admin"

echo Username: user
echo Password: $(kubectl get secret --namespace default wordpress-mysql -o jsonpath="{.data.wordpress-password}" | base64 --decode)
```

### Considerations for future iterations of this bundle
- Connect to azure-cli with a service prinicpal

# Kubernetes Wordpress + Azure MySQL Bundle
### "Because you need a database and probably don't want to manage it yourself"

This bundle provisions an Azure MySQL instance, creates a database, installs Wordpress on a given AKS cluster and connects the wordpress application to the newly provisioned Azure MySQL instance.

## Prerequistes
- Running Kubernetes cluster (we used AKS) with Helm enabled (Tiller running)
- Log into azure-cli locally to put your Azure credentials in place

## Install this bundle
1. Clone this repo: `git clone git@github.com:deis/bundles.git`
2. `cd bundles`
3. You'll need to pass in credentials to connect to your Kubernetes cluster and MySQL database server.
```console
$ duffle creds add wordpress-mysql/example-wordpress-mysql-credential-set.yaml
$ duffle creds show example-wordpress-mysql-credential-set
```
4. Use the `duffle install` command by passing in a name (`wordpress-mysql`), a bundle metadata file (`-f <bundle.json>`) and a credential set(`-c example-wordpress-mysql-credential-set`):

> Note: the default credential set uses environment variables `MYSQL_USER` and `MYSQL_PASSWORD`. These should be set prior to running the install.

```console
$ export MYSQL_USER=username
$ export MYSQL_PASSWORD=SuperSecretPassword

$ duffle install wordpress-mysql -f wordpress-mysql/cnab/bundle.json -c example-wordpress-mysql-credential-set
```

> Note: you can override values such as the AKS cluster_name, by adding `--set cluster_name=your-aks-name` to the duffle install command.

This can take a while to complete. Check for the Wordpress pod running in your cluster and view the logs. Wait for this pod to be running and ready.

5. Login to Wordpress. Use the commands below to find your Wordpress URL and credentials and test: 

```
export SERVICE_IP=$(kubectl get svc --namespace default wordpress-mysql8-wordpress --template "{{ range (index .status.loadBalancer.ingress 0) }}{{.}}{{ end }}")

echo "WordPress URL: http://$SERVICE_IP/"
echo "WordPress Admin URL: http://$SERVICE_IP/admin"

echo Username: user
echo Password: $(kubectl get secret --namespace default wordpress-mysql8-wordpress -o jsonpath="{.data.wordpress-password}" | base64 --decode)
```


### Considerations for future iterations of this bundle
- Connect to azure-cli with a service prinicpal

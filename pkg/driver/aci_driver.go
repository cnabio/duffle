package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/containerinstance/mgmt/2018-10-01/containerinstance"
	"github.com/Azure/azure-sdk-for-go/services/msi/mgmt/2018-11-30/msi"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2015-11-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"

	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const userAgent string = "Duffle ACI Driver"

// ACIDriver runs Docker and OCI invocation images in ACI
type ACIDriver struct {
	config map[string]string
	// This property is set to true if Duffle is running in cloud shell
	inCloudShell bool
	// If true, this will not actually create ACI instance
	Simulate           bool
	deleteACIResources bool
	authorizer         autorest.Authorizer
	subscriptionID     string
}

// Config returns the ACI driver configuration options
func (d *ACIDriver) Config() map[string]string {
	return map[string]string{
		"VERBOSE":                  "Increase verbosity. true, false are supported values",
		"AZURE_CLIENT_ID":          "AAD Client ID for Azure account authentication - used to authenticate to Azure for ACI creation",
		"AZURE_CLIENT_SECRET":      "AAD Client Secret for Azure account authentication - used to authenticate to Azure for ACI creation",
		"AZURE_TENANT_ID":          "Azure AAD Tenant Id Azure account authentication - used to authenticate to Azure for ACI creation",
		"AZURE_SUBSCRIPTION_ID":    "Azure Subscription Id - this is the subscription to be used for ACI creation, if not specified the default subscription is used",
		"AZURE_APP_ID":             "Azure Applications Id - this is the application to be used to authenticate to Azure",
		"ACI_RESOURCE_GROUP":       "The name of the existing Resource Group to create the ACI instance in, if not specified a Resource Group will be created",
		"ACI_LOCATION":             "The location to create the ACI Instance in",
		"ACI_NAME":                 "The name of the ACI instance to create - if not specified a name will be generated",
		"ACI_DO_NOT_DELETE":        "Do not delete RG and ACI instance created - useful for debugging - only deletes RG if it was created by the driver",
		"ACI_MSI_TYPE":             "If this is set to User or System the created ACI Container Group will be launched with MSI",
		"ACI_SYSTEM_MSI_ROLE":      "The role to be asssigned to System MSI User - used if ACI_MSI_TYPE == System, if this is null or empty then the role defaults to contributor",
		"ACI_SYSTEM_MSI_SCOPE":     "The scope to apply the role to System MSI User - will attempt to set scope to the  Resource Group that the ACI Instance is being created in if not set",
		"ACI_USER_MSI_RESOURCE_ID": "The resource Id of the MSI User - required if ACI_MSI_TYPE == User",
	}
}

// SetConfig sets ACI driver configuration
func (d *ACIDriver) SetConfig(settings map[string]string) {
	d.config = settings
}

// Run executes the ACI driver
func (d *ACIDriver) Run(op *Operation) error {
	return d.exec(op)
}

// Handles indicates that the ACI driver supports "docker" and "oci"
func (d *ACIDriver) Handles(dt string) bool {
	return dt == ImageTypeDocker || dt == ImageTypeOCI
}

func (d *ACIDriver) exec(op *Operation) (reterr error) {
	d.deleteACIResources = true
	if len(d.config["ACI_DO_NOT_DELETE"]) > 0 && strings.ToLower(d.config["ACI_DO_NOT_DELETE"]) == "true" {
		d.deleteACIResources = false
	}

	d.log("Delete Resources:", d.deleteACIResources)
	d.inCloudShell = len(os.Getenv("ACC_CLOUD")) > 0
	d.log("In Cloud Shell:", d.inCloudShell)
	if d.Simulate {
		return nil
	}

	err := d.loginToAzure()
	if err != nil {
		return fmt.Errorf("cannot Login To Azure: %v", err)
	}

	err = d.setAzureSubscriptionID()
	if err != nil {
		return fmt.Errorf("cannot set Azure subscription: %v", err)
	}

	err = d.createACIInstance(op)
	if err != nil {
		return fmt.Errorf("creating ACI instance failed: %v", err)
	}

	return nil
}

func (d *ACIDriver) loginToAzure() error {
	clientID := d.config["AZURE_CLIENT_ID"]
	clientSecret := d.config["AZURE_CLIENT_SECRET"]
	tenantID := d.config["AZURE_TENANT_ID"]
	applicationID := d.config["AZURE_APP_ID"]
	d.log("clientID:", clientID)
	d.log("clientSecret:", clientSecret)
	d.log("tenantID:", tenantID)
	d.log("applicationID:", applicationID)
	// Attempt to login with Service Principal
	if len(clientID) != 0 && len(clientSecret) != 0 && len(tenantID) != 0 {
		d.log("Attempting to Login with Service Principal")
		clientCredentailsConfig := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID)
		authorizer, err := clientCredentailsConfig.Authorizer()
		if err != nil {
			return fmt.Errorf("Attempt to login to azure with Service Principal failed: %v", err)
		}

		d.authorizer = authorizer
		return nil
	}

	// Attempt to login with Device Code
	if len(applicationID) != 0 && len(tenantID) != 0 {
		d.log("Attempting to Login with Device Code")
		deviceConfig := auth.NewDeviceFlowConfig(applicationID, tenantID)
		authorizer, err := deviceConfig.Authorizer()
		if err != nil {
			return fmt.Errorf("Attempt to login to azure with Device Code failed: %v", err)
		}

		fmt.Println("Logged in with Device Code")
		d.authorizer = authorizer
		return nil
	}

	// Attempt to use token from CloudShell
	if d.inCloudShell {
		d.log("Attempting to Login with CloudShell")
		token, err := d.getCloudShellToken()
		if err != nil {
			return fmt.Errorf("Attempt to login to Azure with CloudShell failed: %v", err)
		}

		d.authorizer = autorest.NewBearerAuthorizer(token)
		return nil
	}

	// Attempt to login with MSI
	if checkForMSIEndpoint() {
		d.log("Attempting to Login with MSI")
		msiConfig := auth.NewMSIConfig()
		authorizer, err := msiConfig.Authorizer()
		if err != nil {
			return fmt.Errorf("Attempt to login to azure with MSI failed: %v", err)
		}

		d.authorizer = authorizer
		return nil
	}

	return errors.New("Cannot login to Azure - no valid credentials provided")
}

func (d *ACIDriver) setAzureSubscriptionID() error {
	subscription := d.config["AZURE_SUBSCRIPTION_ID"]
	d.log("Subscription:", subscription)
	if len(subscription) != 0 {
		d.log("Setting Subscription to ", subscription)
		d.subscriptionID = subscription
	} else {
		subscriptionClient := subscriptions.NewClient()
		subscriptionClient.Authorizer = d.authorizer
		subscriptionClient.AddToUserAgent(userAgent)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result, err := subscriptionClient.ListComplete(ctx)
		if err != nil {
			return fmt.Errorf("Attempt to List Subscriptions Failed: %v", err)
		}

		err = result.NextWithContext(ctx)
		if err != nil {
			return fmt.Errorf("Attempt to Get Subscription Failed: %v", err)
		}

		// Just choose the first subscription
		result.NextWithContext(ctx)
		if result.NotDone() {
			subscriptionID := *result.Value().SubscriptionID
			d.log("Setting Subscription to", subscriptionID)
			d.subscriptionID = subscriptionID
		} else {
			return errors.New("Cannot find a subscription")
		}

	}
	return nil
}

func (d *ACIDriver) createACIInstance(op *Operation) error {
	// GET ACI Config
	aciRG := d.config["ACI_RESOURCE_GROUP"]
	aciLocation := d.config["ACI_LOCATION"]
	d.log("Resource Group:", aciRG)
	d.log("Location:", aciLocation)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if len(aciLocation) > 0 {
		providersClient := d.getProvidersClient()
		provider, err := providersClient.Get(ctx, "Microsoft.ContainerInstance", "")
		if err != nil {
			return fmt.Errorf("Error getting provider details for ACI: %v", err)
		}

		for _, t := range *provider.ResourceTypes {
			if *t.ResourceType == "ContainerGroups" {
				if !locationIsAvailable(aciLocation, *t.Locations) {
					return fmt.Errorf("ACI Location is invalid: %v", aciLocation)
				}
			}
		}

	} else if len(aciRG) == 0 {
		return errors.New("ACI Driver requires ACI_LOCATION environment variable or an existing Resource Group in ACI_RESOURCE_GROUP")
	}

	groupsClient := d.getGroupsClient()
	if len(aciRG) == 0 {
		d.log("Creating Resource Group")
		aciRG = uuid.New().String()
		d.log("Resource Group:", aciRG)
		_, err := groupsClient.CreateOrUpdate(
			ctx,
			aciRG,
			resources.Group{
				Location: &aciLocation,
			})
		if err != nil {
			return fmt.Errorf("Failed to create resource group: %v", err)
		}

		defer func() {
			if d.deleteACIResources {
				d.log("Deleting Resource Group")
				future, err := groupsClient.Delete(ctx, aciRG)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to execute delete resource group %s error: %v\n", aciRG, err)
				}

				err = future.WaitForCompletionRef(ctx, groupsClient.Client)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to delete resource group %s error: %v\n", aciRG, err)
				} else {
					d.log("Deleted Resource Group ", aciRG)
				}

			}

		}()
	} else {
		_, err := groupsClient.Get(ctx, aciRG)
		if err != nil {
			return fmt.Errorf("Checking for existing resource group %s failed with error: %v", aciRG, err)
		}

	}

	aciName := d.config["ACI_NAME"]
	d.log("ACI Name:", aciName)
	if len(aciName) == 0 {
		aciName = fmt.Sprintf("duffle-%s", uuid.New().String())
		d.log("ACI Name:", aciName)
	}

	// TODO ACI does not support file copy
	// Does not support files because ACI does not support file copy yet
	if len(op.Files) > 0 {
		for k, v := range op.Files {
			d.log("File", k, "Value", v)
		}

		// Ignore Image Map if its empty
		if v, e := op.Files["/cnab/app/image-map.json"]; e && (len(v) > 0 && v != "{}") || !e || len(op.Files) > 1 {
			return errors.New("ACI Driver does not support files")
		}

	}

	// Create ACI Instance

	var env []containerinstance.EnvironmentVariable
	for k, v := range op.Environment {
		d.log("Environment Variable: Name:", k, "Value:", v)
		env = append(env, containerinstance.EnvironmentVariable{
			Name:        to.StringPtr(k),
			SecureValue: to.StringPtr(strings.Replace(v, "'", "''", -1)),
		})
	}
	identity, err := d.getContainerIdentity(ctx, aciRG)
	if err != nil {
		return fmt.Errorf("Failed to get container Identity:%v", err)
	}

	_, err = d.createInstance(aciName, aciLocation, aciRG, op.Image, env, *identity)
	if err != nil {
		return fmt.Errorf("Error creating ACI Instance:%v", err)
	}

	// TODO: Check if ACR under ACI supports MSI
	// TODO: Login to ACR if the registry is azurecr.io
	// TODO: Add support for private registry

	if d.deleteACIResources {
		defer func() {
			d.log("Deleting Container Instance")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			containerGroupsClient := d.getContainerGroupsClient()
			_, err := containerGroupsClient.Delete(ctx, aciRG, aciName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to delete container error: %v\n", err)
			}

			d.log("Deleted Container ", aciName)
		}()
	}

	// Check if the container is running
	state, err := d.getContainerState(aciRG, aciName)
	if err != nil {
		return fmt.Errorf("Error getting container state :%v", err)
	}

	// Get the logs if the container failed immediately
	if strings.Compare(state, "Failed") == 0 {
		_, err := d.getContainerLogs(ctx, aciRG, aciName, 0)
		if err != nil {
			return fmt.Errorf("Error getting container logs :%v", err)
		}

		return errors.New("Container execution failed")
	}

	containerRunning := true
	linesOutput := 0
	for containerRunning {
		d.log("Getting Container State")
		state, err := d.getContainerState(aciRG, aciName)
		if err != nil {
			return fmt.Errorf("Error getting container state :%v", err)
		}

		if strings.Compare(state, "Running") == 0 {
			linesOutput, err = d.getContainerLogs(ctx, aciRG, aciName, linesOutput)
			if err != nil {
				return fmt.Errorf("Error getting container logs :%v", err)
			}

			d.log("Sleeping")
			time.Sleep(5 * time.Second)
		} else {
			if strings.Compare(state, "Succeeded") != 0 {
				return fmt.Errorf("Unexpected Container Status:%s", state)
			}

			d.log("Container terminated successfully")
			containerRunning = false
		}

	}

	_, err = d.getContainerLogs(ctx, aciRG, aciName, linesOutput)
	if err != nil {
		return fmt.Errorf("Error getting container logs :%v", err)
	}

	d.log("Done")
	return nil
}

// This will only work if the logs dont get truncated because of size.
func (d *ACIDriver) getContainerLogs(ctx context.Context, aciRG string, aciName string, linesOutput int) (int, error) {
	d.log("Getting Logs")
	containerClient := d.getContainerClient()
	logs, err := containerClient.ListLogs(ctx, aciRG, aciName, aciName, nil)
	if err != nil {
		return 0, fmt.Errorf("Error getting container logs :%v", err)
	}

	lines := strings.Split(strings.TrimSuffix(*logs.Content, "\n"), "\n")
	noOfLines := len(lines)
	for currentLine := linesOutput; currentLine < noOfLines; currentLine++ {
		fmt.Println(lines[currentLine])
	}

	return noOfLines, nil
}

func (d *ACIDriver) getContainerIdentity(ctx context.Context, aciRG string) (*identityDetails, error) {
	useMSI := d.config["ACI_MSI_TYPE"]
	d.log("MSI Type:", useMSI)
	userMSIResourceID := d.config["ACI_USER_MSI_RESOURCE_ID"]
	d.log("User MSI Resource ID:", userMSIResourceID)
	if len(useMSI) == 0 {
		return &identityDetails{
			MSIType: "none",
		}, nil
	}

	// System MSI
	if strings.ToLower(useMSI) == "system" {
		scope := fmt.Sprintf("/subscriptions/%s/resourcegroups/%s", d.subscriptionID, aciRG)
		role := "Contributor"
		if len(d.config["ACI_SYSTEM_MSI_ROLE"]) > 0 {
			d.log("System MSI Role:", d.config["ACI_SYSTEM_MSI_ROLE"])
			role = d.config["ACI_SYSTEM_MSI_ROLE"]
		}
		d.log("MSI Role:", role)

		if len(d.config["ACI_SYSTEM_MSI_SCOPE"]) > 0 {
			d.log("System MSI Scope:", d.config["ACI_SYSTEM_MSI_SCOPE"])
			scope = d.config["ACI_SYSTEM_MSI_SCOPE"]
		}
		d.log("MSI Scope:", scope)

		return &identityDetails{
			MSIType: "system",
			Identity: &containerinstance.ContainerGroupIdentity{
				Type: containerinstance.SystemAssigned,
			},
			Scope: &scope,
			Role:  &role,
		}, nil
	}

	// User MSI
	if strings.ToLower(useMSI) == "user" {
		if len(userMSIResourceID) > 0 {
			resource, err := azure.ParseResourceID(userMSIResourceID)
			if err != nil {
				return nil, fmt.Errorf("ACI_USER_MSI_RESOURCE_ID environment variable parsing error: %v ", err)
			}

			userAssignedIdentitiesClient := d.getUserAssignedIdentitiesClient(resource.SubscriptionID)
			identity, err := userAssignedIdentitiesClient.Get(ctx, resource.ResourceGroup, resource.ResourceName)
			if err != nil {
				return nil, fmt.Errorf("Error getting User Assigned Identity:%s  Error: %v", userMSIResourceID, err)
			}
			return &identityDetails{
				MSIType: "user",
				Identity: &containerinstance.ContainerGroupIdentity{
					Type: containerinstance.UserAssigned,
					UserAssignedIdentities: map[string]*containerinstance.ContainerGroupIdentityUserAssignedIdentitiesValue{
						*identity.ID: {},
					},
				},
			}, nil
		}

		return nil, errors.New("ACI Driver requires ACI_USER_MSI_RESOURCE_ID environment variable when ACI_MSI_TYPE is set to user")
	}

	return nil, fmt.Errorf("ACI_MSI_TYPE environment variable unknown value: %s ", useMSI)
}

func (d *ACIDriver) getRoleDefinitionsClient(subscriptionID string) authorization.RoleDefinitionsClient {
	roleDefinitionsClient := authorization.NewRoleDefinitionsClient(subscriptionID)
	roleDefinitionsClient.Authorizer = d.authorizer
	roleDefinitionsClient.AddToUserAgent(userAgent)
	return roleDefinitionsClient
}

func (d *ACIDriver) getRoleAssignmentClient(subscriptionID string) authorization.RoleAssignmentsClient {
	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(subscriptionID)
	roleAssignmentsClient.Authorizer = d.authorizer
	roleAssignmentsClient.AddToUserAgent(userAgent)
	return roleAssignmentsClient
}

func (d *ACIDriver) getUserAssignedIdentitiesClient(subscriptionID string) msi.UserAssignedIdentitiesClient {
	userAssignedIdentitiesClient := msi.NewUserAssignedIdentitiesClient(subscriptionID)
	userAssignedIdentitiesClient.Authorizer = d.authorizer
	userAssignedIdentitiesClient.AddToUserAgent(userAgent)
	return userAssignedIdentitiesClient
}

func (d *ACIDriver) getContainerGroupsClient() containerinstance.ContainerGroupsClient {
	containerGroupsClient := containerinstance.NewContainerGroupsClient(d.subscriptionID)
	containerGroupsClient.Authorizer = d.authorizer
	containerGroupsClient.AddToUserAgent(userAgent)
	return containerGroupsClient
}

func (d *ACIDriver) getContainerClient() containerinstance.ContainerClient {
	containerClient := containerinstance.NewContainerClient(d.subscriptionID)
	containerClient.Authorizer = d.authorizer
	containerClient.AddToUserAgent(userAgent)
	return containerClient
}

func (d *ACIDriver) getGroupsClient() resources.GroupsClient {
	groupsClient := resources.NewGroupsClient(d.subscriptionID)
	groupsClient.Authorizer = d.authorizer
	groupsClient.AddToUserAgent(userAgent)
	return groupsClient
}

func (d *ACIDriver) getProvidersClient() resources.ProvidersClient {
	providersClient := resources.NewProvidersClient(d.subscriptionID)
	providersClient.Authorizer = d.authorizer
	providersClient.AddToUserAgent(userAgent)
	return providersClient
}

func (d *ACIDriver) getContainerState(aciRG string, aciName string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	containerGroupsClient := d.getContainerGroupsClient()
	resp, err := containerGroupsClient.Get(ctx, aciRG, aciName)
	if err != nil {
		return "", err
	}

	return *resp.InstanceView.State, nil
}

func locationIsAvailable(location string, locations []string) bool {
	location = strings.ToLower(strings.Replace(location, " ", "", -1))
	for _, l := range locations {
		l = strings.ToLower(strings.Replace(l, " ", "", -1))
		if l == location {
			return true
		}
	}

	return false
}
func checkForMSIEndpoint() bool {
	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	_, err := client.Head("http://169.254.169.254/metadata/identity/oauth2/token")
	return err == nil
}

func (d *ACIDriver) createContainerGroup(aciName string, aciRG string, containerGroup containerinstance.ContainerGroup) (containerinstance.ContainerGroup, error) {
	containerGroupsClient := d.getContainerGroupsClient()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	future, err := containerGroupsClient.CreateOrUpdate(ctx, aciRG, aciName, containerGroup)
	if err != nil {
		return containerinstance.ContainerGroup{}, fmt.Errorf("Error Creating Container Group: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, containerGroupsClient.Client)
	if err != nil {
		return containerinstance.ContainerGroup{}, fmt.Errorf("Error Waiting for Container Group creation: %v", err)
	}

	return future.Result(containerGroupsClient)
}
func (d *ACIDriver) createInstance(aciName string, aciLocation string, aciRG string, image string, env []containerinstance.EnvironmentVariable, identity identityDetails) (*containerinstance.ContainerGroup, error) {
	// if the MSI type is system assigned then need to create the ACI Instance first in order to create the identity and then assign permissions
	if identity.MSIType == "system" {
		d.log("Creating ACI to create System Identity")
		alpine := "alpine:latest"
		containerGroup, err := d.createContainerGroup(
			aciName,
			aciRG,
			containerinstance.ContainerGroup{
				Name:     &aciName,
				Location: &aciLocation,
				Identity: identity.Identity,
				ContainerGroupProperties: &containerinstance.ContainerGroupProperties{
					OsType:        containerinstance.Linux,
					RestartPolicy: containerinstance.Never,
					Containers: &[]containerinstance.Container{
						{
							Name: &aciName,
							ContainerProperties: &containerinstance.ContainerProperties{
								Image: &alpine,
								Resources: &containerinstance.ResourceRequirements{
									Requests: &containerinstance.ResourceRequests{
										MemoryInGB: to.Float64Ptr(1),
										CPU:        to.Float64Ptr(1.5),
									},
									Limits: &containerinstance.ResourceLimits{
										MemoryInGB: to.Float64Ptr(1),
										CPU:        to.Float64Ptr(1.5),
									},
								},
							},
						},
					},
				},
			})
		if err != nil {
			return nil, fmt.Errorf("Error Creating Container Group for System MSI creation: %v", err)
		}

		err = d.setUpSystemMSIRBAC(containerGroup.Identity.PrincipalID, *identity.Scope, *identity.Role)
		if err != nil {
			return nil, fmt.Errorf("Error setting up RBAC for System MSI : %v", err)
		}

	}

	d.log("Creating ACI for CNAB action")
	containerGroup, err := d.createContainerGroup(
		aciName,
		aciRG,
		containerinstance.ContainerGroup{
			Name:     &aciName,
			Location: &aciLocation,
			Identity: identity.Identity,
			ContainerGroupProperties: &containerinstance.ContainerGroupProperties{
				OsType:        containerinstance.Linux,
				RestartPolicy: containerinstance.Never,
				Containers: &[]containerinstance.Container{
					{
						Name: &aciName,
						ContainerProperties: &containerinstance.ContainerProperties{
							Image: &image,
							Resources: &containerinstance.ResourceRequirements{
								Requests: &containerinstance.ResourceRequests{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1.5),
								},
								Limits: &containerinstance.ResourceLimits{
									MemoryInGB: to.Float64Ptr(1),
									CPU:        to.Float64Ptr(1.5),
								},
							},
							EnvironmentVariables: &env,
						},
					},
				},
			},
		})
	if err != nil {
		return nil, fmt.Errorf("Error Creating Container Group: %v", err)
	}

	return &containerGroup, nil
}
func (d *ACIDriver) setUpSystemMSIRBAC(principalID *string, scope string, role string) error {
	d.log("Setting up System MSI Scope ", scope, "Role ", role)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	roleDefinitionsClient := d.getRoleDefinitionsClient(d.subscriptionID)
	roleDefinitionID := ""
	for roleDefinitions, err := roleDefinitionsClient.ListComplete(ctx, scope, ""); roleDefinitions.NotDone(); err = roleDefinitions.NextWithContext(ctx) {
		if err != nil {
			return fmt.Errorf("Error getting RoleDefinitions for Scope:%s Error: %v", scope, err)
		}

		if *roleDefinitions.Value().Properties.RoleName == role {
			roleDefinitionID = *roleDefinitions.Value().ID
			break
		}

	}

	if roleDefinitionID == "" {
		return fmt.Errorf("Role Definition for Role %s not found for Scope:%s", role, scope)
	}

	d.log("RoleDefinitionId", roleDefinitionID)
	// Wait for principal to be available
	attempts := 5
	var err error
	for i := 0; i < attempts; i++ {
		d.log("Creating Role Attempt", i)
		roleAssignmentsClient := d.getRoleAssignmentClient(d.subscriptionID)
		_, raerror := roleAssignmentsClient.Create(ctx, scope, uuid.New().String(), authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: &roleDefinitionID,
				PrincipalID:      principalID,
			},
		})
		if raerror != nil {
			err = fmt.Errorf("Error creating RoleDefinition Role:%s for Scope:%s Error: %v", role, scope, raerror)
			d.log("Creating Role Attempt:", i, "Error:", err)
			time.Sleep(20 * time.Second)
			continue
		}
		err = raerror
		break
	}

	return err
}

func (d *ACIDriver) log(message ...interface{}) {
	if len(d.config["VERBOSE"]) > 0 && strings.ToLower(d.config["VERBOSE"]) == "true" {
		fmt.Println(message...)
	}

}

type identityDetails struct {
	MSIType  string
	Identity *containerinstance.ContainerGroupIdentity
	Scope    *string
	Role     *string
}

func (d *ACIDriver) getCloudShellToken() (*adal.Token, error) {

	MSIEndpoint := os.Getenv("MSI_ENDPOINT")
	d.log("CloudShell MSI Endpoint", MSIEndpoint)
	if len(MSIEndpoint) == 0 {
		return nil, errors.New("MSI_ENDPOINT environment variable not set")
	}

	timeout := time.Duration(1 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", "http://localhost:50342/oauth2/token", nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP Request to CloudShell Token: %v", err)
	}

	req.Header.Set("Metadata", "true")
	query := req.URL.Query()
	query.Add("api-version", "2018-02-01")
	query.Add("resource", "https://management.azure.com/")
	req.URL.RawQuery = query.Encode()
	d.log("Token Query", query.Encode())
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting CloudShell Token: %v", err)
	}
	defer resp.Body.Close()
	rawResp, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return nil, fmt.Errorf("Error getting CloudShell Token. Status Code:'%d'. Failed reading response body error: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("Error getting Token. Status Code:'%d'. Response body: %s", resp.StatusCode, string(rawResp))
	}

	var token adal.Token
	err = json.Unmarshal(rawResp, &token)
	if err != nil {
		return nil, fmt.Errorf("Error deserialising CloudShell Token Status Code: '%d'. Token: %s", resp.StatusCode, string(rawResp))
	}
	return &token, nil
}

package atlas

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/utils/debug"

	"github.com/mongodb-forks/digest"
	"go.mongodb.org/atlas/mongodbatlas"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/config"
)

type Atlas struct {
	OrgID   string
	Public  string
	Private string
	Client  *mongodbatlas.Client
}

const (
	keyDescription = "created from the AO test"
)

func AClient() (Atlas, error) {
	var A Atlas
	A.OrgID = os.Getenv("MCLI_ORG_ID")
	A.Public = os.Getenv("MCLI_PUBLIC_API_KEY")
	A.Private = os.Getenv("MCLI_PRIVATE_API_KEY")
	t := digest.NewTransport(A.Public, A.Private)
	tc, err := t.Client()
	if err != nil {
		return A, err
	}
	A.Client = mongodbatlas.NewClient(tc)
	u, _ := url.Parse(config.AtlasHost)
	A.Client.BaseURL = u
	return A, nil
}

func (a *Atlas) AddKeyWithAccessList(projectID string, roles, access []string) (*mongodbatlas.APIKey, error) {
	createKeyRequest := &mongodbatlas.APIKeyInput{
		Desc:  keyDescription,
		Roles: roles,
	}
	newKey, _, err := a.Client.ProjectAPIKeys.Create(context.Background(), projectID, createKeyRequest)
	if err != nil {
		return nil, err
	}
	createAccessRequest := formAccessRequest(access)
	_, _, err = a.Client.WhitelistAPIKeys.Create(context.Background(), a.OrgID, newKey.ID, createAccessRequest)
	if err != nil {
		return nil, err
	}
	return newKey, nil
}

func formAccessRequest(access []string) []*mongodbatlas.WhitelistAPIKeysReq {
	createRequest := make([]*mongodbatlas.WhitelistAPIKeysReq, 0)
	var req *mongodbatlas.WhitelistAPIKeysReq
	for _, item := range access {
		if strings.Contains(item, "/") {
			req = &mongodbatlas.WhitelistAPIKeysReq{CidrBlock: item}
		} else {
			req = &mongodbatlas.WhitelistAPIKeysReq{IPAddress: item}
		}
		createRequest = append(createRequest, req)
	}
	return createRequest
}

func (a *Atlas) GetPrivateEndpoint(projectID, provider string) ([]mongodbatlas.PrivateEndpointConnection, error) {
	enpointsList, _, err := a.Client.PrivateEndpoints.List(context.Background(), projectID, provider, &mongodbatlas.ListOptions{})
	if err != nil {
		return nil, err
	}
	ginkgoPrettyPrintf(enpointsList, "listing private endpoints in project %s", projectID)
	return enpointsList, nil
}

func (a *Atlas) GetAdvancedDeployment(projectId, deploymentName string) (*mongodbatlas.AdvancedCluster, error) {
	advancedDeployment, _, err := a.Client.AdvancedClusters.Get(context.Background(), projectId, deploymentName)
	if err != nil {
		return nil, err
	}
	ginkgoPrettyPrintf(advancedDeployment, "getting advanced deployment %s in project %s", deploymentName, projectId)
	return advancedDeployment, nil
}

func (a *Atlas) GetServerlessInstance(projectId, deploymentName string) (*mongodbatlas.Cluster, error) {
	serverlessInstance, _, err := a.Client.ServerlessInstances.Get(context.Background(), projectId, deploymentName)
	if err != nil {
		return nil, err
	}
	ginkgoPrettyPrintf(serverlessInstance, "getting serverless instance %s in project %s", deploymentName, projectId)
	return serverlessInstance, nil
}

// ginkgoPrettyPrintf displays a message and a formatted json object through the Ginkgo Writer.
func ginkgoPrettyPrintf(obj interface{}, msg string, formatArgs ...interface{}) {
	ginkgo.GinkgoWriter.Println(fmt.Sprintf(msg, formatArgs...))
	ginkgo.GinkgoWriter.Println(debug.PrettyString(obj))
}

func (a *Atlas) GetIntegrationbyType(projectId, iType string) (*mongodbatlas.ThirdPartyIntegration, error) {
	integraion, _, err := a.Client.Integrations.Get(context.Background(), projectId, iType)
	if err != nil {
		return nil, err
	}
	return integraion, nil
}

func (a *Atlas) GetUserByName(database, projectID, username string) (*mongodbatlas.DatabaseUser, error) {
	dbUser, _, err := a.Client.DatabaseUsers.Get(context.Background(), database, projectID, username)
	if err != nil {
		return nil, err
	}
	return dbUser, nil
}

func (a *Atlas) DeleteGlobalKey(key mongodbatlas.APIKey) error {
	_, err := a.Client.APIKeys.Delete(context.Background(), a.OrgID, key.ID)
	if err != nil {
		return err
	}
	return nil
}

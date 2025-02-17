package e2e_test

import (
	"fmt"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/k8s"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/data"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/actions"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/actions/deploy"
	kubecli "github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/cli/kubecli"

	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/model"
	"github.com/mongodb/mongodb-atlas-kubernetes/test/e2e/utils"
)

var _ = Describe("Configuration namespaced. Deploy deployment", Label("deployment-ns"), func() {
	var testData *model.TestDataProvider

	BeforeEach(func() {
		Eventually(kubecli.GetVersionOutput()).Should(Say(K8sVersion))
	})
	AfterEach(func() {
		GinkgoWriter.Write([]byte("\n"))
		GinkgoWriter.Write([]byte("===============================================\n"))
		GinkgoWriter.Write([]byte("Operator namespace: " + testData.Resources.Namespace + "\n"))
		GinkgoWriter.Write([]byte("===============================================\n"))
		if CurrentSpecReport().Failed() {
			GinkgoWriter.Write([]byte("Test has been failed. Trying to save logs...\n"))
			utils.SaveToFile(
				fmt.Sprintf("output/%s/operatorDecribe.txt", testData.Resources.Namespace),
				[]byte(kubecli.DescribeOperatorPod(testData.Resources.Namespace)),
			)
			utils.SaveToFile(
				fmt.Sprintf("output/%s/operator-logs.txt", testData.Resources.Namespace),
				kubecli.GetManagerLogs(testData.Resources.Namespace),
			)
			actions.SaveTestAppLogs(testData.Resources)
			actions.SaveProjectsToFile(testData.Context, testData.K8SClient, testData.Resources.Namespace)
			actions.SaveK8sResources(
				[]string{"deploy", "atlasdeployments", "atlasdatabaseusers"},
				testData.Resources.Namespace,
			)
			actions.AfterEachFinalCleanup([]model.TestDataProvider{*testData})
			actions.DeleteTestDataDeployments(testData)
			actions.DeleteTestDataProject(testData)
		}
	})

	DescribeTable("Namespaced operators working only with its own namespace with different configuration",
		func(test *model.TestDataProvider) {
			testData = test
			mainCycle(test)
		},
		Entry("Trial - Simplest configuration with no backup and one Admin User", Label("ns-trial"),
			model.DataProvider(
				"operator-ns-trial",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess(),
				30000,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateBasicDeployment("basic-deployment")).
				WithUsers(data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-u1"), data.WithAdminRole())),
		),
		Entry("Almost Production - Backup and 2 DB users: one Admin and one read-only", Label("ns-backup2db", "long-run"),
			model.DataProvider(
				"operator-ns-prodlike",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess(),
				30001,
				[]func(*model.TestDataProvider){
					actions.UpdateSpecOfSelectedDeployment(data.NewDeploymentWithBackupSpec(), 0),
					actions.SuspendDeployment,
					actions.ReactivateDeployment,
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateDeploymentWithBackup("backup-deployment")).
				WithUsers(
					data.BasicUser("admin", "user1", data.WithSecretRef("dbuser-secret-u1"), data.WithAdminRole()),
					data.BasicUser("user2", "user2", data.WithSecretRef("dbuser-secret-u2"), data.WithCustomRole(string(model.RoleCustomReadWrite), "Ships", "readWrite")),
				)),
		Entry("Multiregion AWS, Backup and 2 DBUsers", Label("ns-multiregion-aws-2"),
			model.DataProvider(
				"operator-ns-multiregion-aws",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess(),
				30003,
				[]func(*model.TestDataProvider){
					actions.SuspendDeployment,
					actions.ReactivateDeployment,
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateDeploymentWithMultiregionAWS("multiregion-aws-deployment")).
				WithUsers(data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-u1"), data.WithAdminRole()),
					data.BasicUser("user2", "user2", data.WithSecretRef("dbuser-secret-u2"), data.WithAdminRole())),
		),
		Entry("Multiregion Azure, Backup and 1 DBUser", Label("ns-multiregion-azure-1"),
			model.DataProvider(
				"operator-multiregion-azure",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess().CreateAsGlobalLevelKey(),
				30012,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateDeploymentWithMultiregionAzure("multiregion-azure-deployment")).
				WithUsers(data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-azure"), data.WithAdminRole())),
		),
		Entry("Multiregion GCP, Backup and 1 DBUser", Label("ns-multiregion-gcp-1"),
			model.DataProvider(
				"operator-multiregion-gcp",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess().CreateAsGlobalLevelKey(),
				30013,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateDeploymentWithMultiregionGCP("multiregion-gcp-deployment")).
				WithUsers(data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-gcp"), data.WithAdminRole())),
		),
		Entry("Product Owner - Simplest configuration with ProjectOwner and update deployment to have backup", Label("ns-owner", "long-run"),
			model.DataProvider(
				"operator-ns-product-owner",
				model.NewEmptyAtlasKeyType().WithRoles([]model.AtlasRoles{model.GroupOwner}).WithWhiteList([]string{"0.0.0.1/1", "128.0.0.0/1"}),
				30010,
				[]func(*model.TestDataProvider){
					actions.UpdateSpecOfSelectedDeployment(data.NewDeploymentWithBackupSpec(), 0),
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateDeploymentWithBackup("backup-deployment")).
				WithUsers(
					data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-u1"), data.WithAdminRole()),
				)),
		Entry("Trial - Global connection", Label("ns-global-key"),
			model.DataProvider(
				"operator-ns-trial-global",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess().CreateAsGlobalLevelKey(),
				30011,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateBasicDeployment("trial")).
				WithUsers(
					data.BasicUser("user1", "user1", data.WithSecretRef("dbuser-secret-u1"), data.WithAdminRole()),
				),
		),
		Entry("Free - Users can use M0, default key", Label("ns-m0"),
			model.DataProvider(
				"operator-ns-free",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess(),
				30016,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateBasicFreeDeployment("basic-free-deployment")).
				WithUsers(data.BasicUser("user", "user1", data.WithSecretRef("dbuser-secret"), data.WithAdminRole())),
		),
		Entry("Free - Users can use M0, global", Label("ns-global-key-m0"),
			model.DataProvider(
				"operator-ns-free",
				model.NewEmptyAtlasKeyType().UseDefaulFullAccess().CreateAsGlobalLevelKey(),
				30017,
				[]func(*model.TestDataProvider){
					actions.DeleteFirstUser,
				},
			).WithProject(data.DefaultProject()).
				WithInitialDeployments(data.CreateBasicFreeDeployment("basic-free-deployment")).
				WithUsers(data.BasicUser("user", "user1", data.WithSecretRef("dbuser-secret"), data.WithAdminRole())),
		),
	)
})

func mainCycle(testData *model.TestDataProvider) {
	actions.PrepareUsersConfigurations(testData)
	deploy.NamespacedOperator(testData) // TODO: how to deploy operator by code?
	Expect(k8s.CreateNamespace(testData.Context, testData.K8SClient, testData.Resources.Namespace)).Should(Succeed())

	By("Deploy User Resouces", func() {
		k8s.CreateDefaultSecret(testData.Context, testData.K8SClient, config.DefaultOperatorGlobalKey, testData.Resources.Namespace)
		if !testData.Resources.AtlasKeyAccessType.GlobalLevelKey {
			actions.CreateConnectionAtlasKey(testData)
		}
		deploy.CreateProject(testData)
		deploy.CreateInitialDeployments(testData)
		deploy.CreateUsers(testData)
	})

	By("Additional check for the current data set", func() {
		for _, check := range testData.Actions {
			check(testData)
		}
	})
	By("Delete User Resources", func() {
		deploy.DeleteInitialDeployments(testData)
		deploy.DeleteProject(testData)
		deploy.DeleteUsers(testData)
	})
}

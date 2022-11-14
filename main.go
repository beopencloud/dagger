package main

import (
	"cno-ui-ci/cno/models"
	cno "cno-ui-ci/cno/services"
	"cno-ui-ci/config"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/spf13/viper"

	"k8s.io/client-go/util/homedir"
)

const (
	GIT_REPOSITORY_URL   = "git@github.com:beopencloud/cno-api.git"
	GIT_BRANCH           = "develop"
	PUBLISH_ADDRESS      = "beopenit/cno-api"
	IMAGE_TAG            = "v1.123"
	NAMESPACE            = "ali-ka"
	DEPLOYMENT_NAME      = "cno-api"
	DEPLOYMENT_FILE_NAME = "cno-api-deploy.yml"
	ENVIRONMENT_NAME     = "dev"
	PROJECT_NAME         = "develop"
	WORKLOAD_NAME        = "cno-api"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fmt.Println("Cannot connect to dagger client : ", err)
	}
	defer client.Close()

	codeSrc := client.Git(GIT_REPOSITORY_URL).Branch(GIT_BRANCH).Tree()

	goBaseImage := client.Container().From("golang:latest").WithMountedDirectory("/src", codeSrc).WithWorkdir("/src")
	resultTest := goBaseImage.Exec(dagger.ContainerExecOpts{
		Args: []string{"go", "test", "./..."},
	})
	fmt.Println(resultTest.Stdout().Contents(ctx))

	goApp := client.Container().Build(codeSrc)
	address, err := goApp.Publish(ctx, PUBLISH_ADDRESS+":"+IMAGE_TAG)
	if err != nil {
		fmt.Println("Could not publish container: ", err)
		os.Exit(1)
	}
	fmt.Println("Container pushed at ", address)

	err = setUpConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	token, err := cno.Login(models.Credentials{Username: config.CNO_USERNAME, Password: config.CNO_PASSWORD, OrgName: config.CNO_ORGANIZATION_NAME})
	if err != nil {
		fmt.Println("Could not login :", err)
		os.Exit(1)
	}

	workload, err := cno.GetWorkload(WORKLOAD_NAME, ENVIRONMENT_NAME, PROJECT_NAME)
	if err != nil {
		fmt.Println("Could not get workload")
		os.Exit(1)
	}

	workloadPatch := models.WorkloadPatchSpec{LiveContainers: []models.Container{{Image: PUBLISH_ADDRESS + ":" + IMAGE_TAG}}, AutoDeploy: true}
	err = cno.PatchWorkload(*workload, workloadPatch)
	if err != nil {
		fmt.Println("Could not patch workload:", err)
		os.Exit(1)
	}

	fmt.Println("Workload deployed")
}

func setUpConfig() error {
	home := homedir.HomeDir()
	viper.AddConfigPath(filepath.Join(home, ".cno"))
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	if _, err := os.Stat(filepath.Join(home, ".cno/config")); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(home, ".cno"), os.ModePerm)
		_, err := os.Create(filepath.Join(home, ".cno/config"))
		if err != nil {
			err = errors.New("error to create file config: " + err.Error())
			return err
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		err = errors.New("Using Config file error: " + err.Error())
		return err
	}
	return nil
}

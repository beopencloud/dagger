Project to run pipelines with dagger

Dagger allows us to run CI/CD pipeline as container.

In this project, we are running the CNO API CI pipeline where we test the app, package it in a container and push it to the docker registry.

The CD part is done by CNO.

To run the project, make sure you have docker installed and go.

Clone the project by running in a folder of choice:

1. With SSH : git@github.com:beopencloud/cno-api.git
2. With https: https://github.com/beopencloud/cno-api.git

Go to the root of the project :

cd ./dagger

Run the code with the command :  go run main.go

Enjoy !

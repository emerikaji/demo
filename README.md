# Container demo

This is an API designed to illustrate my abilities in working with Go, designing CI/CD pipelines, and deploying with Kubernetes.

It is available at [eaji.cc](https://eaji.cc), with all routes and their specifications listed in the open API doc at [eaji.cc/v1/openapi.yaml](https://eaji.cc/v1/swagger.yaml). That spec can be viewed directly at [eaji.cc/v1/swagger](https://eaji.cc/v1/swagger).

## Project design

### API: Pure-go

This API is a ticketing-type demonstration, with users, events and tickets to those events. The intention is to showcase a few important backend design features: authenticating securely using a password and tokens, tracking unique reservations with lifetimes, managing concurrent user requests, among other things.

It uses as few libraries as possible and reimplements all the API boilerplate onto a standard `http.ServeMux`. With this, and the very lightweight from-scratch Docker image, it should offer very satisfying performance, scalability, and startup time.

### Deployment: Kubernetes cluster

The API is deployed and replicated automatically on my k3s home lab consisting of one server node and one agent node. The setup is only exposed to the internet on ports 80 and 443, and communication between nodes happens on a separate network.

2 replicas of the API are set up, with the goal of splitting load between each node. In case the agent downs, the server is powerful enough to handle serving both replicas at once.

### CI/CD

The project is continuously tested upon a push to its dev or main branches, as well as upon pull request.
On push to main, if tests pass, the app is built into a container from its Dockerfile, pushed to Github repositories, and served to the aforementioned k8s cluster by applying the new image digest to the configuration.
For safety, the Kube configuration used in the GitHub Action is generated with a dedicated ServiceAccount and rotating keys.

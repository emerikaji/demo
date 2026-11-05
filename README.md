# Container demo

This is an API designed to illustrate my abilities in working with Go, designing CI/CD pipelines, and deploying with Kubernetes.

It is available at [eaji.cc](https://eaji.cc), with all routes and their specifications listed in the open API doc at [eaji.cc/v1/openapi.yaml](https://eaji.cc/v1/swagger.yaml). That spec can be viewed directly at [eaji.cc/v1/swagger](https://eaji.cc/v1/swagger).

## Project design

This API is a ticketing-type demonstration, with users, events and tickets to those events. The intention is to showcase a few important backend design features: authenticating securely using a password and tokens, tracking unique reservations with lifetimes, managing concurrent user requests, among other things.

The rest of the repo and project are designed around it. The API is deployed and replicated automatically on my k3s home lab consisting of two nodes, where it is set up behind Traefik reverse proxy. CI/CD is visible and handled in the workflows, with the two branches (dev/main) serving their separate role (develop and test continuously, then merge and deploy). For safety, the kubernetes configuration used in the GitHub Action is generated with a dedicated ServiceAccount and rotating keys.

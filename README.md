# Container demo

This is an API designed to illustrate my abilities in working with Go, designing CI/CD pipelines, deploying with Docker and hosting behind Traefik.

It is available at [eaji.cc](https://eaji.cc), with all routes and their specifications listed in the open API doc at [eaji.cc/v1/openapi.yaml](https://eaji.cc/v1/swagger.yaml). That spec can be viewed directly at [eaji.cc/v1/swagger](https://eaji.cc/v1/swagger).

You can also check out the traefik dashboard that I willingly made public for demonstration purposes at [traefik.eaji.cc](https://traefik.eaji.cc).

## Project design

This API is a ticketing-type demonstration, with users, events and tickets to those events. The intention is to showcase a few important backend design features: authenticating securely using a password and tokens, tracking unique reservations with lifetimes, managing concurrent user requests, among other things.

The rest of the repo and project are designed around it. The API is deployed automatically to a small server with a secure environment, where it is set up behind a Traefik reverse proxy in a Docker container. CI/CD is visible and handled in the workflows, with the two branches (dev/main) serving their separate role (develop and test continuously, then merge and deploy). A staging/test branch would be standard but two environments would be redundant for a small demonstration project.

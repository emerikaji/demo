# Containers demo

This is an API designed to illustrate my capabilities in working with Go, designing CI/CD pipelines, deploying with Docker and hosting behind Traefik.

It will be made available at `eaji.cc`, where you will also be able to navigate and discover the full deployment architecture.

If you're here before it's live, thank you for your patience! You are welcome to browse the codebase and check out what's already been achieved.

## Project design

This API is a ticketing-type demonstration, with users, events and tickets to those events. The intention is to showcase a few important backend design features: authenticating securely using a password and tokens, tracking unique reservations with lifetimes, managing concurrent user requests, among other things.

The rest of the repo and project is designed around it. The API is deployed automatically to a small server with a secure environment, where it is set up behind a Traefik reverse proxy in a Docker container. CI/CD is visible and handled in the workflows, with the two branches (dev/main) serving their separate role (develop and test continuously, then merge and deploy). A staging/test branch would be standard but two environments would be redundant for a small demonstration project.


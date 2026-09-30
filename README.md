# Containers demo

This is an API designed to illustrate my capabilities in working with Go, designing CI/CD pipelines, deploying with Docker and hosting behind Traefik.

It is available at [eaji.cc](https://eaji.cc), where you can already test the following API routes:
- `GET /v1/health`
- `POST /v1/auth/register`
  - JSON body: `{"email": "example@email.com", "password": "securepass", "role":"attendee"}` (`role` should either be "attendee" or "organizer").
- `POST /v1/auth/login`
  - JSON body: `{"email": "example@email.com", "password": "securepass"}`
  - JSON response: `"{"token": "a jwt token"}"`

A full API doc will be done at a later time, as well as a proper home page on the root domain name, which will feature more information on the project itself. You can also check out the traefik dashboard, that I willingly made public for demonstration purposes at [traefik.eaji.cc](https://traefik.eaji.cc).

## Project design

This API is a ticketing-type demonstration, with users, events and tickets to those events. The intention is to showcase a few important backend design features: authenticating securely using a password and tokens, tracking unique reservations with lifetimes, managing concurrent user requests, among other things.

The rest of the repo and project is designed around it. The API is deployed automatically to a small server with a secure environment, where it is set up behind a Traefik reverse proxy in a Docker container. CI/CD is visible and handled in the workflows, with the two branches (dev/main) serving their separate role (develop and test continuously, then merge and deploy). A staging/test branch would be standard but two environments would be redundant for a small demonstration project.

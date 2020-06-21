---
layout: default
title: Deployment - Highly-Available
parent: Deployment
nav_order: 2
---

# Highly-Available Deployment

**Authelia** can be deployed on bare metal or on Kubernetes with two
different kind of artifacts: the distributable version (binary and public_html)
or a Docker image.

**NOTE:** If not done already, we highly recommend you first follow the
[Getting Started] documentation.

## On Bare Metal

**Authelia** has been designed to be a proxy companion handling the 
authentication and authorization requests for your entire infrastructure.

As **Authelia** will be key to your architecture, it requires several
components to make it highly-available. Deploying it in production means having
an LDAP server for storing the information about the users, a Redis cache to
store the user sessions in a distributed manner, a SQL server like MariaDB to
persist user configurations and one or more nginx reverse proxies configured to
be used with Authelia. With such a setup **Authelia** can easily be scaled to
multiple instances to evenly handle the traffic.

Here are the available steps to deploy **Authelia** given 
the configuration file is **/path/to/your/configuration.yml**. Note that you can
create your own configuration file from [config.template.yml] located at
the root of the repo.

**NOTE**: Prefer using environment variables to set secrets in production otherwise
pay attention to the permissions of the configuration file. See
[secrets](../configuration/secrets.md) for more information.

### Deploy with the distributable version

    # Build it if not done already
    $ authelia-scripts build
    $ authelia --config /path/to/your/configuration.yml

### Deploy With Docker

    $ docker run -v /path/to/your/configuration.yml:/config/configuration.yml -e TZ=Europe/Paris authelia/authelia

## FAQ

### Why is this not automated?

Ansible would be a very good candidate to automate the installation of such
an infrastructure on bare metal. We would be more than happy to review any PR on that matter.



[config.template.yml]: https://github.com/authelia/authelia/blob/master/config.template.yml
[Getting Started]: ../getting-started.md
[Deployment for Devs]: ./deployment-dev.md
[Kubernetes]: https://kubernetes.io/
